package typx

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/stringsx"

	"github.com/xoctopus/typx/internal/dumper"
)

func NewLitType(t any) (x *LitType) {
	switch u := t.(type) {
	case reflect.Type:
		if l, ok := gRLiterals.Load(u); ok {
			return l
		}
	case types.Type:
		if l, ok := gTLiterals.Load(u); ok {
			return l
		}
	case *LitType:
		return u
	default:
		panic(fmt.Errorf("unexpect input for new a LitType from %T", u))
	}

	defer func() {
		if x != nil {
			x.underlying = t
			switch u := t.(type) {
			case reflect.Type:
				gRLiterals.Store(u, x)
			case types.Type:
				gTLiterals.Store(u, x)
			}
		}
	}()

	id := Wrap(t)
	return NewLitTypeByID(id)
}

func NewLitTypeByID(id string) (x *LitType) {
	ident := func(code string, x ast.Node) string {
		return code[x.Pos()-1 : x.End()-1]
	}
	expr := must.NoErrorV(parser.ParseExpr(id))

	switch e := expr.(type) {
	case *ast.ArrayType:
		if e.Len != nil {
			return &LitType{
				kind: reflect.Array,
				ele:  NewLitTypeByID(ident(id, e.Elt)),
				len:  must.NoErrorV(stringsx.Atoi(ident(id, e.Len))),
			}
		}
		return &LitType{
			kind: reflect.Slice,
			ele:  NewLitTypeByID(ident(id, e.Elt)),
		}
	case *ast.ChanType:
		return &LitType{
			kind: reflect.Chan,
			dir:  e.Dir,
			ele:  NewLitTypeByID(ident(id, e.Value)),
		}
	case *ast.FuncType:
		u := &LitType{kind: reflect.Func}
		if e.Params != nil && len(e.Params.List) > 0 {
			u.ins = make([]*LitType, len(e.Params.List))
			for i, p := range e.Params.List {
				param := ident(id, p.Type)
				if i == len(e.Params.List)-1 && strings.HasPrefix(param, "...") {
					u.variadic = true
					u.ins[i] = NewLitTypeByID("[]" + param[3:])
					break
				}
				u.ins[i] = NewLitTypeByID(param)
			}
		}
		if e.Results != nil && len(e.Results.List) > 0 {
			u.outs = make([]*LitType, len(e.Results.List))
			for i, r := range e.Results.List {
				u.outs[i] = NewLitTypeByID(ident(id, r.Type))
			}
		}
		return u
	case *ast.InterfaceType:
		u := &LitType{
			kind:    reflect.Interface,
			methods: make([]*LitType, len(e.Methods.List)),
		}
		for i, m := range e.Methods.List {
			mi := NewLitTypeByID("func" + ident(id, m.Type))
			mi.name = m.Names[0].Name
			u.methods[i] = mi
		}
		return u
	case *ast.MapType:
		return &LitType{
			kind: reflect.Map,
			key:  NewLitTypeByID(ident(id, e.Key)),
			ele:  NewLitTypeByID(ident(id, e.Value)),
		}
	case *ast.StarExpr:
		return &LitType{
			kind: reflect.Pointer,
			ele:  NewLitTypeByID(ident(id, e.X)),
		}
	case *ast.StructType:
		u := &LitType{kind: reflect.Struct}
		if e.Fields != nil {
			u.fields = make([]*LitType, 0, len(e.Fields.List))

			for _, f := range e.Fields.List {
				ft := NewLitTypeByID(ident(id, f.Type))
				if f.Tag != nil {
					ft.tag = must.NoErrorV(strconv.Unquote(f.Tag.Value))
				}
				if len(f.Names) == 0 {
					ft.name = ft.Name()
					if idx := strings.Index(ft.name, "["); idx != -1 {
						ft.name = ft.name[:idx]
					}
					ft.embedded = true
					u.fields = append(u.fields, ft)
				} else {
					for _, n := range f.Names {
						_ft := *ft
						_ft.name = n.Name
						u.fields = append(u.fields, &_ft)
					}
				}
			}
		}
		return u
	case *ast.SelectorExpr:
		u := &LitType{
			pkg:      ident(id, e.X),
			typename: ident(id, e.Sel),
		}
		if u.pkg == "unsafe" && u.typename == "Pointer" {
			u.kind = reflect.UnsafePointer
		}
		return u
	case *ast.IndexExpr:
		u := NewLitTypeByID(ident(id, e.X))
		u.targs = []*LitType{NewLitTypeByID(ident(id, e.Index))}
		return u
	case *ast.IndexListExpr:
		u := NewLitTypeByID(ident(id, e.X))
		u.targs = make([]*LitType, len(e.Indices))
		for i, index := range e.Indices {
			u.targs[i] = NewLitTypeByID(ident(id, index))
		}
		return u
	default:
		code := ident(id, e)
		ex, ok := e.(*ast.Ident)
		must.BeTrueF(ok, "expect an ast.Ident but caught %T: %s", ex, code)
		u := &LitType{typename: ex.Name}
		if k, ok := gRBasicKinds.Load(ex.Name); ok {
			u.kind = k
		}
		return u
	}
}

type LitType struct {
	underlying any
	pkg        string
	typename   string
	targs      []*LitType
	kind       reflect.Kind
	name       string
	key        *LitType
	ele        *LitType
	len        int
	dir        any // ChanDir
	ins        []*LitType
	outs       []*LitType
	variadic   bool
	fields     []*LitType
	methods    []*LitType
	tag        string
	embedded   bool
}

// PkgPath returns type's full package path
func (t *LitType) PkgPath() string {
	return DecodePath(t.pkg)
}

// Name returns type's typename with instantiated type arguments.
func (t *LitType) Name() string {
	if t.typename == "" || len(t.targs) == 0 {
		return t.typename
	}

	b := strings.Builder{}
	b.WriteString(t.typename)
	b.WriteRune('[')
	for i, targ := range t.targs {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(targ.String())
	}
	b.WriteRune(']')
	return b.String()
}

func (t *LitType) Typename() string {
	return t.typename
}

func (t *LitType) TypeArgs() []*LitType {
	return t.targs
}

func (t *LitType) Key() *LitType {
	return t.key
}

func (t *LitType) Elem() *LitType {
	return t.ele
}

func (t *LitType) Ins() []*LitType {
	return t.ins
}

func (t *LitType) Outs() []*LitType {
	return t.outs
}

func (t *LitType) Methods() []*LitType {
	return t.methods
}

func (t *LitType) Fields() []*LitType {
	return t.fields
}

// Kind return literal type kind. it can be seen only when type is unnamed or basic.
// If type is named type. use pkg/typx.Type instead
func (t *LitType) Kind() reflect.Kind {
	return t.kind
}

func (t *LitType) literal(ctx context.Context) string {
	if t.typename != "" {
		b := strings.Builder{}
		if path := t.pkg; path != "" {
			if namer, ok := dumper.CtxPkgNamer.From(ctx); ok {
				path = namer.PackageName(DecodePath(path))
			} else {
				w, _ := dumper.CtxWrapID.From(ctx)
				if !w {
					path = DecodePath(path) // origin
				} else {
					path = EncodePath(path) // wrapped
				}
			}
			if path != "" {
				b.WriteString(path)
				b.WriteString(".")
			}
		}
		b.WriteString(t.typename)
		if len(t.targs) > 0 {
			b.WriteRune('[')
			for i, targ := range t.targs {
				if i > 0 {
					b.WriteString(",")
				}
				b.WriteString(targ.literal(ctx))
			}
			b.WriteRune(']')
		}
		return b.String()
	}

	switch t.kind {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.len, t.ele.literal(ctx))
	case reflect.Chan:
		return fmt.Sprintf("%s%s", ChanDir(t.dir), t.ele.literal(ctx))
	case reflect.Func:
		b := strings.Builder{}

		name := "func"
		if t.name != "" {
			name = t.name
		}
		b.WriteString(name + "(")
		for i := range t.ins {
			if i > 0 {
				b.WriteString(", ")
			}
			if i == len(t.ins)-1 && t.variadic {
				b.WriteString("..." + t.ins[i].literal(ctx)[2:])
				break
			}
			b.WriteString(t.ins[i].literal(ctx))
		}
		b.WriteString(")")

		if len(t.outs) == 0 {
			return b.String()
		}
		b.WriteString(" ")
		if len(t.outs) > 1 {
			b.WriteString("(")
		}
		for i, v := range t.outs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(v.literal(ctx))
		}
		if len(t.outs) > 1 {
			b.WriteString(")")
		}
		return b.String()
	case reflect.Interface:
		if len(t.methods) == 0 {
			return "interface {}"
		}
		b := strings.Builder{}
		b.WriteString("interface { ")
		for i, m := range t.methods {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(m.literal(ctx))
		}
		b.WriteString(" }")
		return b.String()
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", t.key.literal(ctx), t.ele.literal(ctx))
	case reflect.Pointer:
		return "*" + t.ele.literal(ctx)
	case reflect.Slice:
		return "[]" + t.ele.literal(ctx)
	default:
		must.BeTrueF(t.kind == reflect.Struct, "got unexpected kind %s", t.kind)
		if len(t.fields) == 0 {
			return "struct {}"
		}
		b := strings.Builder{}
		b.WriteString("struct { ")
		for i, f := range t.fields {
			if i > 0 {
				b.WriteString("; ")
			}
			if !f.embedded {
				b.WriteString(f.name)
				b.WriteString(" ")
			}
			b.WriteString(f.literal(ctx))
			if len(f.tag) > 0 {
				b.WriteString(" ")
				b.WriteString(strconv.Quote(f.tag))
			}
		}
		b.WriteString(" }")
		return b.String()
	}
}

// Dump returns wrapped type string, this will treat all package path as an identifier.
func (t *LitType) Dump(ctx context.Context) string {
	return t.literal(ctx)
}

// String return type's string with full package path everywhere
func (t *LitType) String() string {
	return t.literal(dumper.CtxWrapID.With(context.Background(), false))
}
