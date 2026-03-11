package typx

import (
	"context"
	"fmt"
	"go/types"
	"reflect"

	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/typx/internal/typx"
)

func NewTType(t types.Type) Type {
	var (
		xt    types.Type
		alias *types.Alias
	)
	switch x := t.(type) {
	case *types.Union, *types.Tuple, *types.TypeParam:
		panic(fmt.Errorf("invalid NewTType by types.Type from `%T`", x))
	case *types.Alias:
		xt = x.Rhs()
		alias = x
	default:
		xt = x
	}
	return &ttype{
		methods: typx.InspectMethods(xt),
		t:       xt,
		u:       typx.NewLitType(xt),
		alias:   alias,
	}
}

type ttype struct {
	ctx     context.Context
	alias   *types.Alias
	t       types.Type
	u       *typx.LitType
	methods []*types.Func
}

func (t *ttype) Unwrap() any {
	return t.t
}

func (t *ttype) Kind() reflect.Kind {
	switch t.t.(type) {
	case *types.Basic:
		return t.u.Kind()
	case *types.Interface:
		return reflect.Interface
	case *types.Struct:
		return reflect.Struct
	case *types.Pointer:
		return reflect.Pointer
	case *types.Slice:
		return reflect.Slice
	case *types.Array:
		return reflect.Array
	case *types.Map:
		return reflect.Map
	case *types.Chan:
		return reflect.Chan
	case *types.Signature:
		return reflect.Func
	default:
		x, ok := t.t.(*types.Named)
		must.BeTrue(ok)
		return NewTType(x.Underlying()).Kind()
	}
}

func (t *ttype) PkgPath() string {
	return t.u.PkgPath()
}

func (t *ttype) Name() string {
	return t.u.Name()
}

func (t *ttype) String() string {
	return t.u.String()
}

func (t *ttype) Implements(u any) bool {
	switch x := u.(type) {
	case Type:
		return t.Implements(x.Unwrap())
	case types.Type:
		if underlying, ok := x.Underlying().(*types.Interface); ok {
			return types.Implements(t.t, underlying)
		}
		return false
	case reflect.Type:
		// if x.Kind() != reflect.Interface {
		// 	return false
		// }
		// return t.Implements(typx.NewTTByRT(x))
		return false
	default:
		return false
	}
}

func (t *ttype) AssignableTo(u any) bool {
	switch x := u.(type) {
	case Type:
		return t.AssignableTo(x.Unwrap())
	case reflect.Type:
		// return types.AssignableTo(t.t, typx.NewTTByRT(x))
		return false
	case types.Type:
		return types.AssignableTo(t.t, x)
	default:
		return false
	}
}

func (t *ttype) ConvertibleTo(u any) bool {
	switch x := u.(type) {
	case Type:
		return t.ConvertibleTo(x.Unwrap())
	case reflect.Type:
		// return types.ConvertibleTo(t.t, typx.NewTTByRT(x))
		return false
	case types.Type:
		return types.ConvertibleTo(t.t, x)
	default:
		return false
	}
}

func (t *ttype) Comparable() bool {
	return types.Comparable(typx.Underlying(t.t))
}

func (t *ttype) Key() Type {
	switch x := t.t.(type) {
	case interface{ Key() types.Type }:
		return NewTType(x.Key())
	case *types.Named:
		return NewTType(typx.Underlying(x)).Key()
	default:
		return nil
	}
}

func (t *ttype) Elem() Type {
	switch x := t.t.(type) {
	case interface{ Elem() types.Type }:
		return NewTType(x.Elem())
	case *types.Named:
		return NewTType(typx.Underlying(x)).Elem()
	default:
		return nil
	}
}

func (t *ttype) Len() int {
	switch x := t.t.(type) {
	case *types.Array:
		return int(x.Len())
	case *types.Named:
		return NewTType(typx.Underlying(x)).Len()
	default:
		return 0
	}
}

func (t *ttype) NumField() int {
	switch x := t.t.(type) {
	case *types.Struct:
		return x.NumFields()
	case *types.Named:
		return NewTType(x.Underlying()).NumField()
	default:
		return 0
	}
}

func (t *ttype) Field(i int) StructField {
	switch x := t.t.(type) {
	case *types.Struct:
		if i >= 0 && i < x.NumFields() {
			return &TStructField{ctx: t.ctx, v: x.Field(i), tag: x.Tag(i)}
		}
		return nil
	case *types.Named:
		return NewTType(typx.Underlying(x)).Field(i)
	default:
		return nil
	}
}

func (t *ttype) FieldByName(name string) (StructField, bool) {
	f := typx.FieldByName(t.t, name)
	if f != nil {
		return &TStructField{ctx: t.ctx, v: f.Var(), tag: f.Tag()}, true
	}
	return nil, false
}

func (t *ttype) FieldByNameFunc(match func(string) bool) (StructField, bool) {
	f := typx.FieldByNameFunc(t.t, match)
	if f != nil {
		return &TStructField{ctx: t.ctx, v: f.Var(), tag: f.Tag()}, true
	}
	return nil, false
}

func (t *ttype) NumMethod() int {
	return len(t.methods)
}

func (t *ttype) Method(i int) Method {
	if i >= 0 && i < len(t.methods) {
		return &TMethod{ctx: t.ctx, r: t.t, f: t.methods[i]}
	}
	return nil
}

func (t *ttype) MethodByName(name string) (Method, bool) {
	for _, m := range t.methods {
		if m.Name() == name {
			return &TMethod{ctx: t.ctx, r: t.t, f: m}, true
		}
	}
	return nil, false
}

func (t *ttype) IsVariadic() bool {
	switch x := t.t.(type) {
	case *types.Signature:
		return x.Variadic()
	case *types.Named:
		return NewTType(x.Underlying()).IsVariadic()
	default:
		return false
	}
}

func (t *ttype) NumIn() int {
	switch x := t.t.(type) {
	case *types.Signature:
		return x.Params().Len()
	case *types.Named:
		return NewTType(x.Underlying()).NumIn()
	default:
		return 0
	}
}

func (t *ttype) In(i int) Type {
	switch x := t.t.(type) {
	case *types.Signature:
		if i >= 0 && i < x.Params().Len() {
			return NewTType(x.Params().At(i).Type())
		}
		return nil
	case *types.Named:
		return NewTType(x.Underlying()).In(i)
	default:
		return nil
	}
}

func (t *ttype) NumOut() int {
	switch x := t.t.(type) {
	case *types.Signature:
		return x.Results().Len()
	case *types.Named:
		return NewTType(x.Underlying()).NumOut()
	default:
		return 0
	}
}

func (t *ttype) Out(i int) Type {
	switch x := t.t.(type) {
	case *types.Signature:
		if i >= 0 && i < x.Results().Len() {
			return NewTType(x.Results().At(i).Type())
		}
		return nil
	case *types.Named:
		return NewTType(x.Underlying()).Out(i)
	default:
		return nil
	}
}

type TStructField struct {
	ctx context.Context
	v   *types.Var
	tag string
}

func (f *TStructField) Pos() int {
	return int(f.v.Pos())
}

func (f *TStructField) PkgPath() string {
	if pkg := f.v.Pkg(); pkg != nil && !f.v.Exported() {
		return pkg.Path()
	}
	return ""
}

func (f *TStructField) Name() string {
	return f.v.Name()
}

func (f *TStructField) Type() Type {
	return NewTType(f.v.Type())
}

func (f *TStructField) Tag() reflect.StructTag {
	return reflect.StructTag(f.tag)
}

func (f *TStructField) Anonymous() bool {
	return f.v.Anonymous()
}

type TMethod struct {
	ctx context.Context
	r   types.Type
	f   *types.Func
}

func (m *TMethod) PkgPath() string {
	// unexported methods were hidden in static analysis
	return ""
}

func (m *TMethod) Name() string {
	return m.f.Name()
}

func (m *TMethod) Type() Type {
	s := m.f.Signature()
	params := make([]*types.Var, 0, s.Params().Len()+1)
	if _, ok := m.r.Underlying().(*types.Interface); !ok {
		params = append(params, types.NewParam(0, nil, "", m.r))
	}
	for v := range s.Params().Variables() {
		params = append(params, v)
	}
	return NewTType(
		types.NewSignatureType(
			nil, nil, nil,
			types.NewTuple(params...),
			s.Results(),
			s.Variadic(),
		),
	)
}
