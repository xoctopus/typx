package typx

import (
	"context"
	"reflect"
	_ "unsafe"

	"github.com/xoctopus/typx/internal/dumper"
	"github.com/xoctopus/typx/internal/typx"
)

type Literal = typx.LitType

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

func LitType(x any) *Literal {
	return typx.NewLitType(x)
}

//go:linkname wrapID github.com/xoctopus/typx/internal/typx.wrapID
func wrapID(string) string

func LitTypeByID(id string) *Literal {
	id = wrapID(id)
	return typx.NewLitTypeByID(id)
}
