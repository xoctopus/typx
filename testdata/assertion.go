package testdata

import (
	"fmt"
	"go/types"
	"reflect"
	"testing"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"
	. "github.com/xoctopus/x/testx"

	lit "github.com/xoctopus/typx/internal/typx"
	"github.com/xoctopus/typx/pkg/typx"
)

var (
	Compares = []*CompareBundles{
		{
			name: "Stringer",
			rtyp: reflect.TypeFor[fmt.Stringer](),
		},
		{
			name: "Bytes",
			rtyp: reflect.TypeFor[interface{ Bytes() []byte }](),
		},
		{
			name: "EmptyInterface",
			rtyp: reflect.TypeFor[any](),
		},
		{
			name: "Error",
			rtyp: reflect.TypeFor[error](),
		},
		{
			name: "EmptyStruct",
			rtyp: reflect.TypeFor[struct{}](),
		},
		{
			name: "Struct",
			rtyp: reflect.TypeFor[struct{ some any }](),
		},
	}

	Cases = BundleCases{
		&Bundle{
			name: "Basics",
			v:    Basics{},
		},
		&Bundle{
			name: "Composites",
			v:    Composites{},
		},
		&Bundle{
			name: "Functions",
			v:    Functions{},
		},
		&Bundle{
			name: "Structures",
			v:    Structures{},
		},
	}
)

func init() {
	for _, b := range Compares {
		b.Init()
	}

	for _, c := range Cases {
		c.Init()
	}
}

type CompareBundle struct {
	name string
	typ  any
}

func (b *CompareBundle) Name() string {
	return b.name
}

func (b *CompareBundle) T() any {
	return b.typ
}

// CompareBundles helps to check Implements/AssignableTo/ConvertibleTo
type CompareBundles struct {
	name  string
	rtyp  reflect.Type
	types []typx.Type
}

func (b *CompareBundles) Name() string {
	return b.name
}

func (b *CompareBundles) Init() {
	rt := b.rtyp
	tt := lit.NewTTByRT(b.rtyp)
	b.types = []typx.Type{typx.NewRType(rt), typx.NewTType(tt)}
}

func (b *CompareBundles) AssignableTo(t *testing.T, c *Case) {
	t.Run(b.Name(), func(t *testing.T) {
		for _, v := range b.types {
			switch x := v.Unwrap().(type) {
			case reflect.Type:
				expect := c.r.AssignableTo(x)
				Expect(t, c.rt.AssignableTo(x), Equal(expect))
				Expect(t, c.rt.AssignableTo(v), Equal(expect))
				Expect(t, c.tt.AssignableTo(x), BeFalse())
				Expect(t, c.tt.AssignableTo(v), BeFalse())
			case types.Type:
				expect := types.AssignableTo(c.t, x)
				Expect(t, c.rt.AssignableTo(x), BeFalse())
				Expect(t, c.rt.AssignableTo(v), BeFalse())
				Expect(t, c.tt.AssignableTo(x), Equal(expect))
				Expect(t, c.tt.AssignableTo(v), Equal(expect))
			}
		}
	})
}

func (b *CompareBundles) ConvertibleTo(t *testing.T, c *Case) {
	t.Run(b.Name(), func(t *testing.T) {
		for _, v := range b.types {
			switch x := v.Unwrap().(type) {
			case reflect.Type:
				expect := c.r.ConvertibleTo(x)
				Expect(t, c.rt.ConvertibleTo(x), Equal(expect))
				Expect(t, c.rt.ConvertibleTo(v), Equal(expect))
				Expect(t, c.tt.ConvertibleTo(x), BeFalse())
				Expect(t, c.tt.ConvertibleTo(v), BeFalse())
			case types.Type:
				expect := types.ConvertibleTo(c.t, x)
				Expect(t, c.rt.ConvertibleTo(x), BeFalse())
				Expect(t, c.rt.ConvertibleTo(v), BeFalse())
				Expect(t, c.tt.ConvertibleTo(x), Equal(expect))
				Expect(t, c.tt.ConvertibleTo(v), Equal(expect))
			}
		}
	})
}

func (b *CompareBundles) Implements(t *testing.T, c *Case) {
	t.Run(b.Name(), func(t *testing.T) {
		for _, v := range b.types {
			if b.rtyp.Kind() != reflect.Interface {
				Expect(t, c.rt.Implements(v), BeFalse())
				Expect(t, c.tt.Implements(v), BeFalse())
				continue
			}
			switch x := v.Unwrap().(type) {
			case reflect.Type:
				expect := c.r.Implements(x)
				Expect(t, c.rt.Implements(x), Equal(expect))
				Expect(t, c.rt.Implements(v), Equal(expect))
				Expect(t, c.tt.Implements(x), BeFalse())
				Expect(t, c.tt.Implements(v), BeFalse())
			case types.Type:
				expect := types.Implements(c.t, x.Underlying().(*types.Interface))
				Expect(t, c.rt.Implements(x), BeFalse())
				Expect(t, c.rt.Implements(v), BeFalse())
				Expect(t, c.tt.Implements(x), Equal(expect))
				Expect(t, c.tt.Implements(v), Equal(expect))
			}
		}
	})
}

type Bundle struct {
	v     any
	name  string
	cases []*Case
}

func (bc *Bundle) Init() {
	bundle := reflect.TypeOf(bc.v)
	must.BeTrue(bundle.Kind() == reflect.Struct)

	for i := range bundle.NumField() {
		f := bundle.Field(i)
		rt := f.Type
		tt := lit.NewTTByRT(rt)
		name := f.Name
		for range 2 {
			bc.cases = append(bc.cases, &Case{
				name: name,
				r:    rt,
				t:    tt,
				rt:   typx.NewRType(rt),
				tt:   typx.NewTType(tt),
			})
			rt = reflect.PointerTo(rt)
			tt = types.NewPointer(tt)
			name = name + "Ptr"
		}
	}
}

func (bc *Bundle) Name() string {
	return bc.name
}

func (bc *Bundle) Run(t *testing.T) {
	for _, c := range bc.cases {
		t.Run(c.name, c.Run)
	}
}

type BundleCases []*Bundle

func NewFieldResult(f typx.StructField, exists bool) *FieldResult {
	return &FieldResult{f, exists}
}

type FieldResult struct {
	f      typx.StructField
	exists bool
}

type FieldAssertion struct {
	f     reflect.StructField
	exist bool
	typ   string
	xfs   []*FieldResult
}

type FieldAssertions []*FieldAssertion

func (fas FieldAssertions) Run(t *testing.T) {
	for _, fa := range fas {
		for _, r := range fa.xfs {
			if fa.exist {
				Expect(t, r.exists, BeTrue())
				Expect(t, r.f, NotBeNil[typx.StructField]())
				Expect(t, r.f.Name(), Equal(fa.f.Name))
				Expect(t, r.f.Tag(), Equal(fa.f.Tag))
				Expect(t, r.f.PkgPath(), Equal(fa.f.PkgPath))
				Expect(t, r.f.Anonymous(), Equal(fa.f.Anonymous))
				Expect(t, r.f.Type().String(), Equal(fa.typ))
			} else {
				Expect(t, r.exists, BeFalse())
				Expect(t, r.f, BeNil[typx.StructField]())
			}
		}
	}
}

func NewMethodResult(m typx.Method, exists bool) *MethodResult {
	return &MethodResult{m, exists}
}

type MethodResult struct {
	m      typx.Method
	exists bool
}

type MethodAssertion struct {
	m      reflect.Method
	exists bool
	typ    string
	xms    []*MethodResult
}

type MethodAssertions []*MethodAssertion

func (mas MethodAssertions) Run(t *testing.T) {
	for _, ma := range mas {
		for _, m := range ma.xms {
			if ma.exists {
				Expect(t, m.exists, BeTrue())
				Expect(t, m.m, NotBeNil[typx.Method]())
				Expect(t, m.m.Name(), Equal(ma.m.Name))
				Expect(t, m.m.PkgPath(), Equal(ma.m.PkgPath))
				Expect(t, m.m.Type().String(), Equal(ma.typ))
			} else {
				Expect(t, m.exists, BeFalse())
				Expect(t, m.m, BeNil[typx.Method]())
			}
		}
	}
}

type Case struct {
	name string
	r    reflect.Type
	t    types.Type
	rt   typx.Type
	tt   typx.Type
}

func (c *Case) Name() string {
	return c.name
}

func (c *Case) Run(t *testing.T) {
	t.Run("Kind", func(t *testing.T) {
		Expect(t, c.rt.Kind(), Equal(c.r.Kind()))
		Expect(t, c.tt.Kind(), Equal(c.r.Kind()))
	})
	t.Run("PkgPath", func(t *testing.T) {
		Expect(t, c.rt.PkgPath(), Equal(c.r.PkgPath()))
		Expect(t, c.tt.PkgPath(), Equal(c.r.PkgPath()))
	})
	t.Run("Name", func(t *testing.T) {
		Expect(t, c.rt.Name(), Equal(c.r.Name()))
		Expect(t, c.tt.Name(), Equal(c.r.Name()))
	})

	t.Run("Implements", func(t *testing.T) {
		for _, b := range Compares {
			b.Implements(t, c)
		}
		t.Run("Nil", func(t *testing.T) {
			Expect(t, c.rt.Implements(nil), BeFalse())
			Expect(t, c.tt.Implements(nil), BeFalse())
		})
	})

	t.Run("AssignableTo", func(t *testing.T) {
		for _, b := range Compares {
			b.AssignableTo(t, c)
		}
		t.Run("Nil", func(t *testing.T) {
			Expect(t, c.rt.ConvertibleTo(nil), BeFalse())
			Expect(t, c.tt.ConvertibleTo(nil), BeFalse())
		})
	})

	t.Run("ConvertibleTo", func(t *testing.T) {
		for _, b := range Compares {
			b.ConvertibleTo(t, c)
		}
		t.Run("Nil", func(t *testing.T) {
			Expect(t, c.rt.AssignableTo(nil), BeFalse())
			Expect(t, c.tt.AssignableTo(nil), BeFalse())
		})
	})

	t.Run("Comparable", func(t *testing.T) {
		expect := c.r.Comparable()
		Expect(t, c.rt.Comparable(), Equal(expect))
		Expect(t, c.tt.Comparable(), Equal(expect))
	})

	t.Run("Key", func(t *testing.T) {
		if c.r.Kind() == reflect.Map {
			Expect(t, c.rt.Key().String(), Equal(c.tt.Key().String()))
		} else {
			Expect(t, c.rt.Key(), BeNil[typx.Type]())
			Expect(t, c.tt.Key(), BeNil[typx.Type]())
		}
	})

	t.Run("Elem", func(t *testing.T) {
		if reflectx.CanElem(c.r) {
			Expect(t, c.rt.Elem().String(), Equal(c.tt.Elem().String()))
		} else {
			Expect(t, c.rt.Elem(), BeNil[typx.Type]())
			Expect(t, c.tt.Elem(), BeNil[typx.Type]())
		}
	})

	t.Run("Len", func(t *testing.T) {
		if c.r.Kind() == reflect.Array {
			Expect(t, c.rt.Len(), Equal(c.r.Len()))
			Expect(t, c.tt.Len(), Equal(c.r.Len()))
		} else {
			Expect(t, c.rt.Len(), Equal(0))
			Expect(t, c.tt.Len(), Equal(0))
		}
	})

	fields := 0
	if c.r.Kind() == reflect.Struct {
		fields = c.r.NumField()
	}

	t.Run("NumField", func(t *testing.T) {
		Expect(t, c.rt.NumField(), Equal(fields))
		Expect(t, c.tt.NumField(), Equal(fields))
	})

	t.Run("Field", func(t *testing.T) {
		fas := c.FieldAssertions(fields, "Name", "name", "str", "_", "unexported", "")
		fas.Run(t)
		t.Run("OutOfRange", func(t *testing.T) {
			Expect(t, c.rt.Field(fields+1), BeNil[typx.StructField]())
			Expect(t, c.tt.Field(fields+1), BeNil[typx.StructField]())
		})
		t.Run("MatchAlwaysTrue", func(t *testing.T) {
			match := func(v string) bool { return true }
			rf, ok1 := c.rt.FieldByNameFunc(match)
			tf, ok2 := c.tt.FieldByNameFunc(match)
			if fields != 1 {
				Expect(t, rf, BeNil[typx.StructField]())
				Expect(t, ok1, BeFalse())
				Expect(t, tf, BeNil[typx.StructField]())
				Expect(t, ok2, BeFalse())
			} else {
				rf0 := c.rt.Field(0)
				Expect(t, ok1, BeTrue())
				Expect(t, rf.Type().String(), Equal(rf0.Type().String()))
				tf0 := c.tt.Field(0)
				Expect(t, ok2, BeTrue())
				Expect(t, tf.Type().String(), Equal(tf0.Type().String()))
			}
		})
	})

	methods := c.r.NumMethod()
	t.Run("NumMethod", func(t *testing.T) {
		Expect(t, c.rt.NumMethod(), Equal(methods))
		Expect(t, c.tt.NumMethod(), Equal(methods))
	})

	t.Run("Method", func(t *testing.T) {
		mas := c.MethodAssertions(methods, "Name", "name", "str", "_", "unexported", "")
		mas.Run(t)
		t.Run("OutOfRange", func(t *testing.T) {
			Expect(t, c.rt.Method(methods+1), BeNil[typx.Method]())
			Expect(t, c.tt.Method(methods+1), BeNil[typx.Method]())
		})
	})

	t.Run("IsVariadic", func(t *testing.T) {
		expect := false
		if c.r.Kind() == reflect.Func {
			expect = c.r.IsVariadic()
		}
		Expect(t, c.rt.IsVariadic(), Equal(expect))
		Expect(t, c.tt.IsVariadic(), Equal(expect))
	})

	t.Run("Ins", func(t *testing.T) {
		if c.r.Kind() == reflect.Func {
			Expect(t, c.rt.NumIn(), Equal(c.r.NumIn()))
			Expect(t, c.tt.NumIn(), Equal(c.r.NumIn()))
			for i := range c.r.NumIn() {
				Expect(t, c.rt.In(i).Name(), Equal(c.tt.In(i).Name()))
				Expect(t, c.rt.In(i).String(), Equal(c.tt.In(i).String()))
				Expect(t, c.rt.In(i).PkgPath(), Equal(c.tt.In(i).PkgPath()))
			}
			t.Run("OutOfRange", func(t *testing.T) {
				Expect(t, c.rt.In(c.r.NumIn()), BeNil[typx.Type]())
				Expect(t, c.tt.In(c.r.NumIn()), BeNil[typx.Type]())
			})
		} else {
			Expect(t, c.rt.NumIn(), Equal(0))
			Expect(t, c.tt.NumIn(), Equal(0))
			Expect(t, c.rt.In(0), BeNil[typx.Type]())
			Expect(t, c.tt.In(0), BeNil[typx.Type]())
		}
	})

	t.Run("Outs", func(t *testing.T) {
		if c.r.Kind() == reflect.Func {
			Expect(t, c.rt.NumOut(), Equal(c.r.NumOut()))
			Expect(t, c.tt.NumOut(), Equal(c.r.NumOut()))
			for i := range c.r.NumOut() {
				Expect(t, c.rt.Out(i).Name(), Equal(c.tt.Out(i).Name()))
				Expect(t, c.rt.Out(i).String(), Equal(c.tt.Out(i).String()))
				Expect(t, c.rt.Out(i).PkgPath(), Equal(c.tt.Out(i).PkgPath()))
			}
			t.Run("OutOfRange", func(t *testing.T) {
				Expect(t, c.rt.Out(c.r.NumOut()), BeNil[typx.Type]())
				Expect(t, c.tt.Out(c.r.NumOut()), BeNil[typx.Type]())
			})
		} else {
			Expect(t, c.rt.NumOut(), Equal(0))
			Expect(t, c.tt.NumOut(), Equal(0))
			Expect(t, c.rt.Out(0), BeNil[typx.Type]())
			Expect(t, c.tt.Out(0), BeNil[typx.Type]())
		}
	})
}

func (c *Case) FieldAssertions(n int, options ...string) FieldAssertions {
	var fas FieldAssertions
	for i := range n {
		f := c.r.Field(i)
		match := func(v string) bool { return v == f.Name }

		rfi := c.rt.Field(i)
		tfi := c.tt.Field(i)

		fa := &FieldAssertion{
			f, true, rfi.Type().String(),
			[]*FieldResult{
				NewFieldResult(rfi, rfi != nil),
				NewFieldResult(tfi, tfi != nil),
				NewFieldResult(c.rt.FieldByName(f.Name)),
				NewFieldResult(c.tt.FieldByName(f.Name)),
				NewFieldResult(c.rt.FieldByNameFunc(match)),
				NewFieldResult(c.tt.FieldByNameFunc(match)),
			},
		}

		fas = append(fas, fa)
	}
	for _, name := range options {
		match := func(v string) bool { return v == name }

		var (
			f      reflect.StructField
			exists bool
		)
		if c.r.Kind() == reflect.Struct {
			f, exists = c.r.FieldByName(name)
		}
		fa := &FieldAssertion{
			f, exists, "",
			[]*FieldResult{
				NewFieldResult(c.rt.FieldByName(name)),
				NewFieldResult(c.tt.FieldByName(name)),
				NewFieldResult(c.rt.FieldByNameFunc(match)),
				NewFieldResult(c.tt.FieldByNameFunc(match)),
			},
		}
		if exists {
			fa.typ = fa.xfs[0].f.Type().String()
		}
		fas = append(fas, fa)
	}
	return fas
}

func (c *Case) MethodAssertions(n int, options ...string) MethodAssertions {
	var mas MethodAssertions
	for i := range n {
		m := c.r.Method(i)

		rmi := c.rt.Method(i)
		tmi := c.tt.Method(i)

		mas = append(mas, &MethodAssertion{
			m: m, exists: true, typ: rmi.Type().String(),
			xms: []*MethodResult{
				NewMethodResult(rmi, rmi != nil),
				NewMethodResult(tmi, tmi != nil),
				NewMethodResult(c.rt.MethodByName(m.Name)),
				NewMethodResult(c.tt.MethodByName(m.Name)),
			},
		})
	}

	for _, name := range options {
		m, exists := c.r.MethodByName(name)
		ma := &MethodAssertion{
			m: m, exists: exists, typ: "",
			xms: []*MethodResult{
				NewMethodResult(c.rt.MethodByName(name)),
				NewMethodResult(c.tt.MethodByName(name)),
			},
		}
		if xm := ma.xms[0]; xm.m != nil {
			ma.typ = xm.m.Type().String()
		}
		mas = append(mas, ma)
	}
	return mas
}
