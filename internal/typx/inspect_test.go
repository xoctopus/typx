package typx_test

import (
	"fmt"
	"go/types"
	"reflect"
	"testing"

	"github.com/xoctopus/x/ptrx"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/typx"
	"github.com/xoctopus/typx/testdata"
)

func runner(t *testing.T, rt reflect.Type, tt types.Type) {
	methods := typx.InspectMethods(tt)
	Expect(t, rt.NumMethod(), Equal(len(methods)))

	for mi, m := range methods {
		Expect(t, m.Name(), Equal(rt.Method(mi).Name))
	}
}

func TestInspectMethods(t *testing.T) {
	rtyp := reflect.TypeFor[testdata.Structures]()
	for f := range rtyp.Fields() {
		rt := f.Type
		tt := typx.NewTTByRT(rt)
		name := f.Name

		for range 2 {
			t.Run(name, func(t *testing.T) {
				runner(t, rt, tt)
			})
			tt = types.NewPointer(tt)
			rt = reflect.PointerTo(rt)
			name = name + "Ptr"
		}
	}

	t.Run("MultiLevelPointer", func(t *testing.T) {
		tt := typx.NewTTByRT(reflect.TypeFor[**testdata.UnambiguousL1AndL2x2]())
		Expect(t, len(typx.InspectMethods(tt)), Equal(0))
		tt = typx.NewTTByRT(reflect.TypeFor[*error]())
		Expect(t, len(typx.InspectMethods(tt)), Equal(0))
	})
}

func TestInspectField(t *testing.T) {
	rtyp := reflect.TypeFor[testdata.Structures]()
	for fi := range rtyp.Fields() {
		rti := fi.Type
		if rti.Kind() != reflect.Struct {
			continue
		}
		t.Run(fi.Name, func(t *testing.T) {
			for fj := range rti.Fields() {
				tt := typx.NewTTByRT(rti)

				tf := typx.FieldByName(tt, fj.Name)
				Expect(t, tf.Var().Name(), Equal(fj.Name))
				Expect(t, tf.Tag(), Equal(string(fj.Tag)))

				tf = typx.FieldByNameFunc(tt, func(s string) bool { return true })
				if rti.NumField() > 1 {
					Expect(t, tf, BeNil[*typx.Field]())
				}

				tf = typx.FieldByNameFunc(tt, func(v string) bool { return v == fj.Name })
				Expect(t, tf.Var().Name(), Equal(fj.Name))
				Expect(t, tf.Tag(), Equal(string(fj.Tag)))

				tf = typx.FieldByName(types.NewPointer(tt), "any")
				Expect(t, tf, BeNil[*typx.Field]())
			}
		})
	}
}

func Example() {
	for _, v := range []any{
		testdata.AmbiguousL1x2{
			StringerL1: testdata.StringerL1("StringerL1"),
			Stringer:   ptrx.Ptr(testdata.StringerL1("fmt.Stringer")),
		},
		testdata.AmbiguousL1AndField{
			StringerL1: testdata.StringerL1("v.StringerL1"),
			String:     "StringField",
		},
		testdata.AmbiguousL2x2{
			StringerL2:       testdata.StringerL2{Stringer: ptrx.Ptr(testdata.StringerL1("StringerL2"))},
			StringerL2WrapL1: &testdata.StringerL2WrapL1{StringerL1: ptrx.Ptr(testdata.StringerL1("StringerL2WrapL1"))},
		},
		testdata.UnambiguousL1AndL2x2{
			StringerL2:       &testdata.StringerL2{Stringer: ptrx.Ptr(testdata.StringerL1("StringerL2"))},
			StringerL2WrapL1: testdata.StringerL2WrapL1{StringerL1: ptrx.Ptr(testdata.StringerL1("StringerL2WrapL1"))},
			StringerL1:       ptrx.Ptr(testdata.StringerL1("StringerL1")),
		},
		testdata.AmbiguousL1x2AndL2{
			StringerL1: testdata.StringerL1("StringerL1"),
			Stringer:   ptrx.Ptr(testdata.StringerL1("fmt.Stringer")),
			StringerL2: testdata.StringerL2{Stringer: ptrx.Ptr(testdata.StringerL1("StringerL2"))},
		},
		testdata.UnambiguousL2AndL3x2{
			StringerL2:       testdata.StringerL2{Stringer: ptrx.Ptr(testdata.StringerL1("StringerL2"))},
			StringerL3:       testdata.StringerL3{StringerL2: testdata.StringerL2{Stringer: ptrx.Ptr(testdata.StringerL1("StringerL3"))}},
			StringerL3WrapL2: testdata.StringerL3WrapL2{StringerL2WrapL1: testdata.StringerL2WrapL1{StringerL1: ptrx.Ptr(testdata.StringerL1("StringerL3WrapL2"))}},
		},
		testdata.AmbiguousL2AndL3x2AndField{
			StringerL2:       testdata.StringerL2{Stringer: ptrx.Ptr(testdata.StringerL1("StringerL2"))},
			StringerL3:       testdata.StringerL3{StringerL2: testdata.StringerL2{Stringer: ptrx.Ptr(testdata.StringerL1("StringerL3"))}},
			StringerL3WrapL2: testdata.StringerL3WrapL2{StringerL2WrapL1: testdata.StringerL2WrapL1{StringerL1: ptrx.Ptr(testdata.StringerL1("StringerL3WrapL2"))}},
			StringField:      testdata.StringField{String: "any"},
		},
	} {
		rt := reflect.TypeOf(v)
		fmt.Println(rt, rt.NumMethod())
		for i := range rt.NumMethod() {
			fmt.Println(rt.Method(i).Name, reflect.ValueOf(v).Method(i).Call(nil)[0].Interface())
		}
		fmt.Println()
	}

	// Output:
	// testdata.AmbiguousL1x2 0
	//
	// testdata.AmbiguousL1AndField 0
	//
	// testdata.AmbiguousL2x2 0
	//
	// testdata.UnambiguousL1AndL2x2 1
	// String StringerL1
	//
	// testdata.AmbiguousL1x2AndL2 0
	//
	// testdata.UnambiguousL2AndL3x2 1
	// String StringerL2
	//
	// testdata.AmbiguousL2AndL3x2AndField 0
}
