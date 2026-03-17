package typx_test

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/dumper"
	"github.com/xoctopus/typx/internal/typx"
	"github.com/xoctopus/typx/testdata"
)

func TestLitType(t *testing.T) {
	for _, c := range LitTypeCases {
		t.Run(c.name, func(t *testing.T) {
			rt := typx.NewLitType(c.rt)
			Expect(t, rt.String(), Equal(c.origin))
			Expect(t, rt.PkgPath(), Equal(c.PkgPath))
			Expect(t, rt.Name(), Equal(c.Name))
			Expect(t, rt.Dump(context.Background()), Equal(c.origin))
			Expect(t, rt.Dump(dumper.CtxWrapID.With(context.Background(), true)), Equal(c.wrapped))
			Expect(t, rt.Dump(dumper.CtxWrapID.With(context.Background(), false)), Equal(c.origin))
			Expect(t, typx.NewLitType(rt), Equal(rt))

			tt := typx.NewLitType(c.tt)
			Expect(t, tt.String(), Equal(c.origin))
			Expect(t, tt.PkgPath(), Equal(c.PkgPath))
			Expect(t, tt.Name(), Equal(c.Name))
			Expect(t, tt.Dump(context.Background()), Equal(c.origin))
			Expect(t, tt.Dump(dumper.CtxWrapID.With(context.Background(), true)), Equal(c.wrapped))
			Expect(t, tt.Dump(dumper.CtxWrapID.With(context.Background(), false)), Equal(c.origin))
			Expect(t, typx.NewLitType(tt), Equal(tt))
		})
	}
	t.Run("LitTypeMeta", func(t *testing.T) {
		rt := typx.NewLitType(reflect.TypeFor[any]())
		Expect(t, rt.Typename(), Equal(""))
		Expect(t, rt.TypeArgs(), HaveLen[[]*typx.LitType](0))

		rt = typx.NewLitType(reflect.TypeFor[struct{}]())
		Expect(t, rt.Typename(), Equal(""))
		Expect(t, rt.TypeArgs(), HaveLen[[]*typx.LitType](0))

		rt = typx.NewLitType(reflect.TypeFor[error]())
		Expect(t, rt.Typename(), Equal("error"))
		Expect(t, rt.TypeArgs(), HaveLen[[]*typx.LitType](0))

		rt = typx.NewLitType(reflect.TypeFor[fmt.Stringer]())
		Expect(t, rt.Typename(), Equal("Stringer"))
		Expect(t, rt.TypeArgs(), HaveLen[[]*typx.LitType](0))

		rt = typx.NewLitType(reflect.TypeFor[[]string]())
		Expect(t, rt.Typename(), Equal(""))
		Expect(t, rt.TypeArgs(), HaveLen[[]*typx.LitType](0))

		rt = typx.NewLitType(reflect.TypeFor[func()]())
		Expect(t, rt.Typename(), Equal(""))
		Expect(t, rt.TypeArgs(), HaveLen[[]*typx.LitType](0))

		rt = typx.NewLitType(reflect.TypeFor[iter.Seq[int]]())
		Expect(t, rt.Typename(), Equal("Seq"))
		targs := rt.TypeArgs()
		Expect(t, len(targs), Equal(1))
		Expect(t, targs[0].Typename(), Equal("int"))

		rt = typx.NewLitType(reflect.TypeFor[testdata.TypedSliceAliasNetAddr]())
		Expect(t, rt.Typename(), Equal("TypedSlice"))
		targs = rt.TypeArgs()
		Expect(t, len(targs), Equal(1))
		Expect(t, targs[0].PkgPath(), Equal("net"))
		Expect(t, targs[0].Typename(), Equal("Addr"))

		rt = typx.NewLitType(reflect.TypeFor[iter.Seq2[int, iter.Seq[fmt.Stringer]]]())
		Expect(t, rt.PkgPath(), Equal("iter"))
		Expect(t, rt.Typename(), Equal("Seq2"))
		targs = rt.TypeArgs()
		Expect(t, len(targs), Equal(2))
		Expect(t, targs[0].PkgPath(), Equal(""))
		Expect(t, targs[0].Typename(), Equal("int"))
		Expect(t, targs[1].PkgPath(), Equal("iter"))
		Expect(t, targs[1].Typename(), Equal("Seq"))
		targs = targs[1].TypeArgs()
		Expect(t, len(targs), Equal(1))
		Expect(t, targs[0].PkgPath(), Equal("fmt"))
		Expect(t, targs[0].Typename(), Equal("Stringer"))
	})
	t.Run("HitCache", func(t *testing.T) {
		for _, c := range LitTypeCases {
			t.Run(c.name, func(t *testing.T) {
				rt := typx.NewLitType(c.rt)
				tt := typx.NewLitType(c.tt)

				Expect(t, rt.String(), Equal(c.origin))
				Expect(t, rt.PkgPath(), Equal(c.PkgPath))
				Expect(t, rt.Name(), Equal(c.Name))

				Expect(t, tt.String(), Equal(c.origin))
				Expect(t, tt.PkgPath(), Equal(c.PkgPath))
				Expect(t, tt.Name(), Equal(c.Name))

				if c.Name == "int" {
					Expect(t, rt.Kind(), Equal(reflect.Int))
					Expect(t, tt.Kind(), Equal(reflect.Int))
				}
			})
		}
		ExpectPanic[error](t, func() { typx.NewLitType(nil) })
	})
	t.Run("WithPkgNamer", func(t *testing.T) {
		dUnnamedStruct = `struct { ` +
			`A string; ` +
			`B int; ` +
			`renamed.Map "json:\"esc''{}[]\\\"\""; ` +
			`renamed.TypedArray[net.Addr]; ` +
			`C struct { ` +
			`renamed.TypedArray[struct { renamed.TypedArray[fmt.Stringer] }] ` +
			`}; ` +
			`D interface { ` +
			`Close() error; ` +
			`Read([]uint8) (int, error); ` +
			`String() string; ` +
			`Write([]uint8) (int, error) ` +
			`} ` +
			`}`

		for _, c := range LitTypeCases {
			if c.name == "TypedArrayUnnamedStruct" {
				expect := "renamed.TypedArray[" + dUnnamedStruct + "]"

				rt := typx.NewLitType(c.rt)
				Expect(t, rt.Dump(dumper.CtxPkgNamer.With(context.Background(), &PkgNamer{})), Equal(expect))

				tt := typx.NewLitType(c.tt)
				Expect(t, tt.Dump(dumper.CtxPkgNamer.With(context.Background(), &PkgNamer{})), Equal(expect))

				break
			}
		}
	})
}

type PkgNamer struct{}

func (p PkgNamer) PackageName(path string) string {
	if path == "github.com/xoctopus/typx/testdata" {
		return "renamed"
	}
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		return path[idx:]
	}
	return path
}
