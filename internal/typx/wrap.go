package typx

import (
	"fmt"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/stringsx"
	"github.com/xoctopus/x/syncx"
)

var (
	// gWrappedIDs mapping origin typeid to wrapped typeid
	gWrappedIDs = syncx.NewXmap[string, string]()
	// gWrappedRTs mapping reflect.Type to wrapped typeid
	gWrappedRTs = syncx.NewXmap[reflect.Type, string]()
	// gWrappedTTs mapping types.Type to wrapped typeid
	gWrappedTTs = syncx.NewXmap[types.Type, string]()
	// gRBasicKinds mapping basic kind string to reflect.Kind
	gRBasicKinds = syncx.NewXmap[string, reflect.Kind]()
	// gTBasicKinds mapping basic kind string to reflect.Kind
	gTBasicKinds = syncx.NewXmap[string, types.Type]()
	// gRLiterals mapping reflect.Type to *LitType
	gRLiterals = syncx.NewXmap[reflect.Type, *LitType]()
	// gTLiterals mapping types.Type to *LitType
	gTLiterals = syncx.NewXmap[types.Type, *LitType]()

	// tError error types.Type
	tError = types.Universe.Lookup("error").Type()
)

func init() {
	for _, k := range []reflect.Kind{
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String,
	} {
		gRBasicKinds.Store(k.String(), k)
	}
	for _, t := range types.Typ {
		gTBasicKinds.Store(t.String(), t)
	}
	gTBasicKinds.Store("error", tError)
}

func Wrap(t any) (wrapped string) {
	switch x := t.(type) {
	case reflect.Type:
		return wrapRT(x)
	case types.Type:
		return wrapTT(x)
	default:
		panic(fmt.Errorf("expect Wrap from reflect.Type or types.Type, but got `%T`", x))
	}
}

func wrapID(id string) (w string) {
	if w, ok := gWrappedIDs.Load(id); ok {
		return w
	}

	defer func(id string) {
		gWrappedIDs.Store(id, w)
	}(id)

	// slice: []elem / array: [len]elem
	if strings.HasPrefix(id, "[") {
		idx, _, r := Bracketed(id, '[')
		ele := id[r+1:]
		return fmt.Sprintf("[%s]%s", idx, wrapID(ele))
	}

	// map: map[key]elem
	if strings.HasPrefix(id, "map[") {
		key, _, r := Bracketed(id, '[')
		ele := id[r+1:]
		return fmt.Sprintf("map[%s]%s", wrapID(key), wrapID(ele))
	}

	// chan elem
	if strings.HasPrefix(id, "chan ") {
		return fmt.Sprintf("chan %s", wrapID(id[5:]))
	}

	// chan<- elem
	if strings.HasPrefix(id, "chan<- ") {
		return fmt.Sprintf("chan<- %s", wrapID(id[7:]))
	}

	// <-chan elem
	if strings.HasPrefix(id, "<-chan ") {
		return fmt.Sprintf("<-chan %s", wrapID(id[7:]))
	}

	// struct: struct { fields... }
	if strings.HasPrefix(id, "struct {") {
		fields, _, _ := Bracketed(id, '{')
		if len(fields) == 0 {
			return "struct {}"
		}

		b := strings.Builder{}
		b.WriteString("struct { ")
		for i, f := range Separate(fields, ';') {
			if i > 0 {
				b.WriteString("; ")
			}
			name, typ, tag := FieldInfo(f)
			if len(name) > 0 {
				b.WriteString(name)
				b.WriteString(" ")
			}
			b.WriteString(wrapID(typ))
			if len(tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(tag))
			}
		}
		b.WriteString(" }")
		return b.String()
	}

	// interface: interface { methods... }
	if strings.HasPrefix(id, "interface {") {
		methods, _, _ := Bracketed(id, '{')
		if len(methods) == 0 {
			return "interface {}"
		}

		b := strings.Builder{}
		b.WriteString("interface { ")
		for i, m := range Separate(methods, ';') {
			if i > 0 {
				b.WriteString("; ")
			}
			idx := strings.Index(m, "(")
			name := m[0:idx]
			typ := wrapID("func" + m[idx:])
			b.WriteString(name + typ[4:])
		}
		b.WriteString(" }")
		return b.String()
	}

	// func: func(params...) results
	if strings.HasPrefix(id, "func(") {
		b := strings.Builder{}
		b.WriteString("func(")

		params, _, pr := Bracketed(id, '(')
		for i, p := range Separate(params, ',') {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(wrapID(p))
		}
		b.WriteString(")")

		if pr == len(id)-1 {
			return b.String()
		}

		b.WriteString(" ")
		id = strings.TrimSpace(id[pr+1:])
		if id[0] != '(' {
			b.WriteString(wrapID(id))
			return b.String()
		}

		b.WriteString("(")
		results, _, _ := Bracketed(id, '(')
		for i, r := range Separate(results, ',') {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(wrapID(r))
		}
		b.WriteString(")")
		return b.String()
	}

	// pointer: *elem
	if strings.HasPrefix(id, "*") {
		return "*" + wrapID(id[1:])
	}

	// variadic: ...elem
	if strings.HasPrefix(id, "...") {
		return "..." + wrapID(id[3:])
	}

	// ident only
	if stringsx.ValidIdentifier(id) {
		return id
	}

	// named: package_path.typename[type arguments...]
	path, name, targs := "", "", ""
	if v, l, r := Bracketed(id, '['); l > 0 && r > 0 {
		targs = v
		id = id[0:l]
	}
	dot := strings.LastIndex(id, ".")
	must.BeTrue(dot != -1)
	path, name = id[0:dot], id[dot+1:]
	b := strings.Builder{}
	b.WriteString(EncodePath(path))
	b.WriteString(".")
	b.WriteString(name)
	if len(targs) > 0 {
		b.WriteString("[")
		for i, targ := range Separate(targs, ',') {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(wrapID(targ))
		}
		b.WriteString("]")
	}
	return b.String()
}

func wrapRT(t reflect.Type) (id string) {
	if id, ok := gWrappedRTs.Load(t); ok {
		return id
	}

	defer func(t reflect.Type) {
		gWrappedRTs.Store(t, id)
	}(t)

	if t.Name() != "" {
		if t.PkgPath() == "" {
			return wrapID(t.Name())
		}
		return wrapID(t.PkgPath() + "." + t.Name())
	}

	switch t.Kind() {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), wrapRT(t.Elem()))
	case reflect.Chan:
		return fmt.Sprintf("%s%s", ChanDir(t.ChanDir()), wrapRT(t.Elem()))
	case reflect.Func:
		b := strings.Builder{}
		b.WriteString("func(")
		for i := range t.NumIn() {
			if i > 0 {
				b.WriteString(", ")
			}
			if i == t.NumIn()-1 && t.IsVariadic() {
				b.WriteString("...")
				b.WriteString(wrapRT(t.In(i).Elem()))
				break
			}
			b.WriteString(wrapRT(t.In(i)))
		}
		b.WriteString(")")
		if t.NumOut() == 0 {
			return b.String()
		}
		b.WriteString(" ")
		if t.NumOut() == 1 {
			b.WriteString(wrapRT(t.Out(0)))
			return b.String()
		}
		b.WriteString("(")
		for i := range t.NumOut() {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(wrapRT(t.Out(i)))
		}
		b.WriteString(")")
		return b.String()
	case reflect.Interface:
		if t.NumMethod() == 0 {
			return "interface {}"
		}
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i := range t.NumMethod() {
			if i > 0 {
				b.WriteString("; ")
			}
			m := t.Method(i)
			f := wrapRT(m.Type)
			b.WriteString(m.Name)
			b.WriteString(f[4:])
		}
		b.WriteString(" }")
		return b.String()
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", wrapRT(t.Key()), wrapRT(t.Elem()))
	case reflect.Pointer:
		return "*" + wrapRT(t.Elem())
	case reflect.Slice:
		return "[]" + wrapRT(t.Elem())
	default:
		must.BeTrueF(t.Kind() == reflect.Struct, "unexpected kind %s", t.Kind())
		if t.NumField() == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i := range t.NumField() {
			if i > 0 {
				b.WriteString("; ")
			}
			f := t.Field(i)
			if !f.Anonymous {
				b.WriteString(f.Name)
				b.WriteString(" ")
			}
			b.WriteString(wrapRT(f.Type))
			if len(f.Tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(string(f.Tag)))
			}
		}
		b.WriteString(" }")
		return b.String()
	}
}

func wrapTT(t types.Type) (id string) {
	for x, id := range gWrappedTTs.Range {
		if x.String() == t.String() {
			return id
		}
	}

	defer func(t types.Type) {
		if id != "" {
			gWrappedTTs.Store(t, id)
		}
	}(t)

	switch x := t.(type) {
	case *types.Alias:
		return wrapTT(types.Unalias(x))
	case *types.Basic:
		switch id := x.String(); id {
		case "byte":
			return "uint8"
		case "rune":
			return "int32"
		default:
			return id
		}
	case *types.Array:
		return fmt.Sprintf("[%d]%s", x.Len(), wrapTT(x.Elem()))
	case *types.Chan:
		return fmt.Sprintf("%s%s", ChanDir(x.Dir()), wrapTT(x.Elem()))
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", wrapTT(x.Key()), wrapTT(x.Elem()))
	case *types.Interface:
		if x.NumMethods() == 0 {
			return "interface {}"
		}
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i := range x.NumMethods() {
			if i > 0 {
				b.WriteString("; ")
			}
			m := x.Method(i)
			f := wrapTT(m.Signature())
			b.WriteString(m.Name() + f[4:])
		}
		b.WriteString(" }")
		return b.String()
	case *types.Pointer:
		return fmt.Sprintf("*%s", wrapTT(x.Elem()))
	case *types.Signature:
		b := strings.Builder{}
		b.WriteString("func(")
		for i := range x.Params().Len() {
			if i > 0 {
				b.WriteString(", ")
			}
			p := x.Params().At(i)
			if x.Variadic() && i == x.Params().Len()-1 {
				b.WriteString("...")
				b.WriteString(wrapTT(p.Type().(*types.Slice).Elem()))
				break
			}
			b.WriteString(wrapTT(p.Type()))
		}
		b.WriteString(")")
		if x.Results().Len() == 0 {
			return b.String()
		}
		b.WriteString(" ")
		if x.Results().Len() == 1 {
			b.WriteString(wrapTT(x.Results().At(0).Type()))
			return b.String()
		}
		b.WriteString("(")
		for i := range x.Results().Len() {
			if i > 0 {
				b.WriteString(", ")
			}
			r := x.Results().At(i)
			b.WriteString(wrapTT(r.Type()))
		}
		b.WriteString(")")

		return b.String()
	case *types.Slice:
		return fmt.Sprintf("[]%s", wrapTT(x.Elem()))
	case *types.Struct:
		if x.NumFields() == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i := range x.NumFields() {
			if i > 0 {
				b.WriteString("; ")
			}
			f := x.Field(i)
			if !f.Anonymous() {
				b.WriteString(f.Name())
				b.WriteString(" ")
			}
			b.WriteString(wrapTT(f.Type()))
			if tag := x.Tag(i); len(tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(tag))
			}
		}
		b.WriteString(" }")
		return b.String()
	default:
		n, ok := t.(*types.Named)
		must.BeTrueF(ok, "invalid wrapTT type: %T", x)
		ok = n.TypeArgs().Len() == n.TypeParams().Len()
		must.BeTrueF(ok, "uninstantiated generic type: %s", x.String())

		b := strings.Builder{}
		path := ""
		if p := n.Obj().Pkg(); p != nil {
			path = p.Path()
		}
		b.WriteString(path)
		if path != "" {
			b.WriteString(".")
		}
		b.WriteString(n.Obj().Name())
		if n.TypeArgs().Len() > 0 {
			b.WriteString("[")
			for i := range n.TypeArgs().Len() {
				if i > 0 {
					b.WriteString(",")
				}
				targ := n.TypeArgs().At(i)
				b.WriteString(wrapTT(targ))
			}
			b.WriteString("]")
		}
		return wrapID(b.String())
	}
}
