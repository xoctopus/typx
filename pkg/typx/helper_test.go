package typx_test

import (
	"context"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	lit "github.com/xoctopus/typx/internal/typx"
	"github.com/xoctopus/typx/pkg/typx"
)

type (
	T1 struct {
		A string
	}
	T2 *int
	T3 *T1
)

func TestDeref(t *testing.T) {
	rt := typx.NewRType(reflect.TypeFor[T1]())
	dt := typx.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typx.NewRType(reflect.TypeFor[T2]())
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal(rt.String()))

	rt = typx.NewRType(reflect.TypeFor[*int]())
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typx.NewRType(reflect.TypeFor[*****int]())
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal("int"))

	rt = typx.NewRType(reflect.TypeFor[T3]())
	dt = typx.Deref(rt)
	Expect(t, dt.String(), Equal("github.com/xoctopus/typx/pkg/typx_test.T3"))
}

func TestPosOfStructField(t *testing.T) {
	tt := typx.NewTType(lit.NewTTByRT(reflect.TypeFor[T1]()))
	Expect(t, typx.PosOfStructField(tt.Field(0)), NotEqual(0))

	rt := typx.NewRType(reflect.TypeFor[T1]())
	Expect(t, typx.PosOfStructField(rt.Field(0)), Equal(0))
}

func TestTypeLit(t *testing.T) {
	Expect(t, typx.TypeLit(context.Background(), reflect.TypeFor[int]()), Equal("int"))
}
