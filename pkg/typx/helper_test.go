package typx_test

import (
	"context"
	"fmt"
	"iter"
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

type X[T any] struct{}

func TestTypeLit(t *testing.T) {
	ctx := context.Background()
	rt0 := reflect.TypeFor[int]()
	rt1 := reflect.TypeFor[X[iter.Seq[fmt.Stringer]]]()
	t0 := typx.LitType(rt0)
	t1 := typx.LitType(rt1)
	Expect(t, t0.String(), Equal(typx.TypeLit(ctx, rt0)))
	Expect(t, t1.String(), Equal(typx.TypeLit(ctx, rt1)))
}

func TestLitTypeByID(t *testing.T) {
	lt := typx.LitTypeByID("github.com/xoctopus/typx/internal/typx.literal")
	Expect(t, lt.PkgPath(), Equal("github.com/xoctopus/typx/internal/typx"))
	Expect(t, lt.Name(), Equal("literal"))

	lt = typx.LitTypeByID("iter.Seq2[int, iter.Seq[float32]]")
	Expect(t, lt.PkgPath(), Equal("iter"))
	Expect(t, lt.Typename(), Equal("Seq2"))
	Expect(t, lt.Name(), Equal("Seq2[int,iter.Seq[float32]]"))
	targs := lt.TypeArgs()
	Expect(t, len(targs), Equal(2))
	Expect(t, targs[0].String(), Equal("int"))
	Expect(t, targs[1].String(), Equal("iter.Seq[float32]"))
}
