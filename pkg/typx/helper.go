package typx

import (
	"context"
	"reflect"

	"github.com/xoctopus/typx/internal/dumper"
	"github.com/xoctopus/typx/internal/typx"
)

var CtxPkgNamer = dumper.CtxPkgNamer

func Deref(t Type) Type {
	for t.Kind() == reflect.Pointer && t.Name() == "" {
		t = t.Elem()
	}
	return t
}

func PosOfStructField(f StructField) int {
	if x, ok := f.(interface{ Pos() int }); ok {
		return x.Pos()
	}
	return 0
}

func TypeLit(ctx context.Context, x any) string {
	return typx.NewLitType(x).Dump(ctx)
}

func LitType(x any) *typx.LitType {
	return typx.NewLitType(x)
}
