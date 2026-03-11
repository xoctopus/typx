package typx_test

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/typx"
)

var ExtractCases = []struct {
	id      string
	bracket rune
	sep     rune
	parts   []string
}{
	{
		`path/to/pkg.TypeName[TypeArg1,TypeArg2[TArg2_1,TArg2_2]]`,
		'[', ',',
		[]string{"TypeArg1", "TypeArg2[TArg2_1,TArg2_2]"},
	},
	{
		`struct { A string "json:\"\\{}[]()\\\"\""; B int }`,
		'{', ';',
		[]string{
			`A string "json:\"\\{}[]()\\\"\""`,
			"B int",
		},
	},
	{
		`func(` +
			`int, ` +
			`struct { F func(); T TypeName[Ta1,Ta2] "json:\"t,omitempty\"" }, ` +
			`path/to/pkg.TypeName[Ta1,Ta2[struct { T T[Ta1,Ta2] "json:\"t\""; F func() }]], ` +
			`func() bool, ` +
			`...any` +
			`) (any, bool)`,
		'(', ',',
		[]string{
			"int",
			`struct { F func(); T TypeName[Ta1,Ta2] "json:\"t,omitempty\"" }`,
			`path/to/pkg.TypeName[Ta1,Ta2[struct { T T[Ta1,Ta2] "json:\"t\""; F func() }]]`,
			"func() bool",
			"...any",
		},
	},
	{
		"",
		'(', ',',
		nil,
	},
}

func TestBracketedAndSeparate(t *testing.T) {
	for _, c := range ExtractCases {
		sub, l, r := typx.Bracketed(c.id, c.bracket)
		parts := typx.Separate(sub, c.sep)
		Expect(t, parts, Equal(c.parts))
		if sub != "" {
			Expect(t, c.id[l+1:r], Equal(sub))
		} else {
			Expect(t, l, Equal(-1))
			Expect(t, r, Equal(-1))
		}
	}
}

type TT[T any] struct{}

func Example_structInTypeArguments() {
	t1 := reflect.TypeFor[TT[struct {
		string             // unexported field
		TT[int] `json:"x"` // embedded generic type field and has tag
	}]]()
	fmt.Println(t1.String())

	targs0, _, _ := typx.Bracketed(t1.String(), '[')
	fields, _, _ := typx.Bracketed(targs0, '{')

	for i, f := range typx.Separate(fields, ';') {
		name, typ, tag := typx.FieldInfo(f)
		fmt.Printf("field%d: name=%s;type=%s;tag=%s\n", i, name, typ, tag)
	}

	// Output:
	// typx_test.TT[struct { github.com/xoctopus/typx/internal/typx_test.string = string; TT = github.com/xoctopus/typx/internal/typx_test.TT[int] "json:\"x\"" }]
	// field0: name=;type=string;tag=
	// field1: name=;type=github.com/xoctopus/typx/internal/typx_test.TT[int];tag=json:"x"
}

func TestFieldInfo(t *testing.T) {
	t.Run("StructInTypeArg", func(t *testing.T) {
		tt := reflect.TypeFor[TT[struct {
			string
			A       int "json:\"a,\\\"'{}()[]//\\\\\""
			TT[int] "json:\"tt\""
			TT2     TT[struct{ A int }]
			Reader  io.Reader
			a       struct{ A string }
			AA      struct{ B int }
		}]]()

		targs, _, _ := typx.Bracketed(tt.String(), '[')
		fields, _, _ := typx.Bracketed(targs, '{')

		expects := [][3]string{
			{"", "string", ""},
			{"A", "int", `json:"a,\"'{}()[]//\\"`},
			{"", "github.com/xoctopus/typx/internal/typx_test.TT[int]", `json:"tt"`},
			{"TT2", "github.com/xoctopus/typx/internal/typx_test.TT[struct { A int }]", ""},
			{"Reader", "io.Reader", ""},
			{"a", "struct { A string }", ""},
			{"AA", "struct { B int }", ""},
		}

		for i, f := range typx.Separate(fields, ';') {
			name, typ, tag := typx.FieldInfo(f)
			Expect(t, name, Equal(expects[i][0]))
			Expect(t, typ, Equal(expects[i][1]))
			Expect(t, tag, Equal(expects[i][2]))
		}
	})
	t.Run("Struct", func(t *testing.T) {
		tt := reflect.TypeFor[struct {
			string
			A       int "json:\"a,\\\"'{}()[]//\\\\\""
			TT[int] "json:\"tt\""
			TT2     TT[struct{ A int }]
			Reader  io.Reader
			a       struct{ A string }
			AA      struct{ B int }
		}]()
		fields, _, _ := typx.Bracketed(tt.String(), '{')

		expects := [][3]string{
			{"", "string", ""},
			{"A", "int", `json:"a,\"'{}()[]//\\"`},
			{"", "typx_test.TT[int]", `json:"tt"`},
			{"TT2", "typx_test.TT[struct { A int }]", ""},
			{"Reader", "io.Reader", ""},
			{"a", "struct { A string }", ""},
			{"AA", "struct { B int }", ""},
		}

		for i, f := range typx.Separate(fields, ';') {
			name, typ, tag := typx.FieldInfo(f)
			Expect(t, name, Equal(expects[i][0]))
			Expect(t, typ, Equal(expects[i][1]))
			Expect(t, tag, Equal(expects[i][2]))
		}
	})
}
