package typx_test

import (
	"fmt"
	"go/types"
	"io"
	"iter"
	"net"
	"reflect"
	"testing"
	"unsafe"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/typx/internal/typx"
	"github.com/xoctopus/typx/testdata"
)

// packages
var (
	stdErrors  = typx.Load("errors")
	stdFmt     = typx.Load("fmt")
	stdNet     = typx.Load("net")
	stdIter    = typx.Load("iter")
	stdIO      = typx.Load("io")
	stdStrconv = typx.Load("strconv")
	testPkg    = typx.Load("github.com/xoctopus/typx/testdata")
)

// generic type
var _tTypedArray = typx.Lookup[*types.Named](testPkg, "TypedArray")

// reflect.Type and types.Type for testing
var (
	rBool           = reflect.TypeFor[bool]()
	tBool           = types.Typ[types.Bool]
	rInt            = reflect.TypeFor[int]()
	tInt            = types.Typ[types.Int]
	rString         = reflect.TypeFor[string]()
	tString         = types.Typ[types.String]
	rRune           = reflect.TypeFor[rune]()
	tRune           = typx.Lookup[*types.Signature](stdStrconv, "QuoteRune").Params().At(0).Type()
	rByte           = reflect.TypeFor[byte]()
	tByte           = types.Typ[types.Byte]
	rUnsafePointer  = reflect.TypeFor[unsafe.Pointer]()
	tUnsafePointer  = types.Typ[types.UnsafePointer]
	rEmptyInterface = reflect.TypeFor[any]()
	tEmptyInterface = types.NewInterfaceType(nil, nil)
	rEmptyStruct    = reflect.TypeFor[struct{}]()
	tEmptyStruct    = types.NewStruct(nil, nil)
	rError          = reflect.TypeFor[error]()
	tError          = typx.Lookup[*types.Signature](stdErrors, "New").Results().At(0).Type()
	rFmtStringer    = reflect.TypeFor[fmt.Stringer]()
	tFmtStringer    = typx.Lookup[*types.Named](stdFmt, "Stringer")
	rNetAddr        = reflect.TypeFor[net.Addr]()
	tNetAddr        = typx.Lookup[*types.Named](stdNet, "Addr")
	rIoReadCloser   = reflect.TypeFor[io.ReadCloser]()
	tIoReadCloser   = typx.Lookup[*types.Named](stdIO, "ReadCloser")
	rIoWriteCloser  = reflect.TypeFor[io.WriteCloser]()
	tIoWriteCloser  = typx.Lookup[*types.Named](stdIO, "WriteCloser")
	rIoReadWriter   = reflect.TypeFor[io.ReadWriter]()
	tIoReadWriter   = typx.Lookup[*types.Named](stdIO, "ReadWriter")
	rTagged         = reflect.TypeFor[testdata.Tagged]()
	tTagged         = typx.Lookup[*types.Named](testPkg, "Tagged")
	rStringSlice    = reflect.TypeFor[[]string]()
	tStringSlice    = types.NewSlice(tString)
	rMapRuneString  = reflect.TypeFor[map[rune]string]()
	tMapRuneString  = types.NewMap(tRune, tString)

	rFunc = reflect.TypeFor[func()]()
	tFunc = types.NewSignatureType(nil, nil, nil, nil, nil, false)

	rFuncVariadic = reflect.TypeFor[func(fmt.Stringer, ...any)]()
	tFuncVariadic = types.NewSignatureType(
		nil, nil, nil,
		types.NewTuple(
			types.NewParam(0, nil, "", tFmtStringer),
			types.NewParam(0, nil, "", types.NewSlice(tEmptyInterface)),
		),
		types.NewTuple(),
		true,
	)

	rFuncWithMultiReturn = reflect.TypeFor[func(int, ...any) (bool, error)]()
	tFuncWithMultiReturn = types.NewSignatureType(
		nil, nil, nil,
		types.NewTuple(
			types.NewParam(0, nil, "", tInt),
			types.NewParam(0, nil, "", types.NewSlice(tEmptyInterface)),
		),
		types.NewTuple(
			types.NewParam(0, nil, "", tBool),
			types.NewParam(0, nil, "", tError),
		),
		true,
	)

	rFuncWithOneReturn = reflect.TypeFor[func(fmt.Stringer, ...any) net.Addr]()
	tFuncWithOneReturn = types.NewSignatureType(
		nil, nil, nil,
		types.NewTuple(
			types.NewParam(0, nil, "", tFmtStringer),
			types.NewParam(0, nil, "", types.NewSlice(tEmptyInterface)),
		),
		types.NewTuple(types.NewParam(0, nil, "", tNetAddr)),
		true,
	)

	rIterSeqArray = reflect.TypeFor[[3]iter.Seq[int]]()
	tIterSeqArray = types.NewArray(typx.Instantiate(typx.Lookup[*types.Named](stdIter, "Seq"), tInt), 3)

	rIterSeq2StringEmptyInterface = reflect.TypeFor[iter.Seq2[string, any]]()
	tIterSeq2StringEmptyInterface = typx.Instantiate(typx.Lookup[*types.Named](stdIter, "Seq2"), tString, tEmptyInterface)

	rChanError = reflect.TypeFor[chan error]()
	tChanError = types.NewChan(types.SendRecv, tError)

	rSendChanTestdataTagged = reflect.TypeFor[chan<- testdata.Tagged]()
	tSendChanTestdataTagged = types.NewChan(types.SendOnly, tTagged)

	rRecvChanTestdataTaggedPointer = reflect.TypeFor[<-chan *testdata.Tagged]()
	tRecvChanTestdataTaggedPointer = types.NewChan(types.RecvOnly, types.NewPointer(tTagged))

	rTypedArrayFmtString = reflect.TypeFor[testdata.TypedArray[fmt.Stringer]]()
	tTypedArrayFmtString = typx.Instantiate(_tTypedArray, tFmtStringer)

	rUnnamedInterfaceComposer = reflect.TypeFor[interface {
		fmt.Stringer
		io.ReadCloser
		io.WriteCloser
		io.ReadWriter
	}]()
	tUnnamedInterfaceComposer = types.NewInterfaceType(
		nil, []types.Type{
			tFmtStringer,
			tIoReadCloser,
			tIoWriteCloser,
			tIoReadWriter,
		},
	)
	wUnnamedInterfaceComposer = `interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }`
	oUnnamedInterfaceComposer = `interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }`
	dUnnamedInterfaceComposer = `interface { Close() error; Read([]uint8) (int, error); String() string; Write([]uint8) (int, error) }`

	rTypedSliceAliasNetAddr = reflect.TypeFor[testdata.TypedSliceAliasNetAddr]()
	tTypedSliceAliasNetAddr = typx.Lookup[*types.Alias](testPkg, "TypedSliceAliasNetAddr")

	rMap = reflect.TypeFor[testdata.Map]()
	tMap = typx.Lookup[*types.Named](testPkg, "Map")

	rUnnamedStruct = reflect.TypeFor[struct {
		A            string
		B            int
		testdata.Map `json:"esc''{}[]\""`
		testdata.TypedArray[net.Addr]
		C struct {
			testdata.TypedArray[struct {
				testdata.TypedArray[fmt.Stringer]
			}]
		}
		D interface {
			fmt.Stringer
			io.ReadCloser
			io.WriteCloser
			io.ReadWriter
		}
	}]()
	tUnnamedStruct = types.NewStruct(
		[]*types.Var{
			types.NewField(0, nil, "A", tString, false),
			types.NewField(0, nil, "B", tInt, false),
			types.NewField(0, nil, "Map", tMap, true),
			types.NewField(0, nil, "TypedArray", typx.Instantiate(_tTypedArray, tNetAddr), true),
			types.NewField(0, nil, "C", types.NewStruct(
				[]*types.Var{
					types.NewField(
						0, nil, "",
						typx.Instantiate(_tTypedArray, types.NewStruct(
							[]*types.Var{types.NewField(0, nil, "", tTypedArrayFmtString, true)},
							[]string{""},
						)),
						true,
					),
				},
				[]string{""},
			), false),
			types.NewField(0, nil, "D", tUnnamedInterfaceComposer, false),
		}, []string{"", "", `json:"esc''{}[]\""`, "", "", ""},
	)
	wUnnamedStruct = `struct { ` +
		`A string; ` +
		`B int; ` +
		`github_com_xoctopus_typx_testdata.Map "json:\"esc''{}[]\\\"\""; ` +
		`github_com_xoctopus_typx_testdata.TypedArray[net.Addr]; ` +
		`C struct { ` +
		`github_com_xoctopus_typx_testdata.TypedArray[struct { github_com_xoctopus_typx_testdata.TypedArray[fmt.Stringer] }] ` +
		`}; ` +
		`D interface { ` +
		`Close() error; ` +
		`Read([]uint8) (int, error); ` +
		`String() string; ` +
		`Write([]uint8) (int, error) ` +
		`} ` +
		`}`
	oUnnamedStruct = `struct { ` +
		`A string; ` +
		`B int; ` +
		`github.com/xoctopus/typx/testdata.Map "json:\"esc''{}[]\\\"\""; ` +
		`github.com/xoctopus/typx/testdata.TypedArray[net.Addr]; ` +
		`C struct { ` +
		`github.com/xoctopus/typx/testdata.TypedArray[struct { github.com/xoctopus/typx/testdata.TypedArray[fmt.Stringer] }] ` +
		`}; ` +
		`D interface { ` +
		`Close() error; ` +
		`Read([]uint8) (int, error); ` +
		`String() string; ` +
		`Write([]uint8) (int, error) ` +
		`} ` +
		`}`
	dUnnamedStruct = `struct { ` +
		`A string; ` +
		`B int; ` +
		`testdata.Map "json:\"esc''{}[]\\\"\""; ` +
		`testdata.TypedArray[net.Addr]; ` +
		`C struct { ` +
		`testdata.TypedArray[struct { testdata.TypedArray[fmt.Stringer] }] ` +
		`}; ` +
		`D interface { ` +
		`Close() error; ` +
		`Read([]uint8) (int, error); ` +
		`String() string; ` +
		`Write([]uint8) (int, error) ` +
		`} ` +
		`}`

	rTypedArrayStringSlice = reflect.TypeFor[testdata.TypedArray[[]string]]()
	tTypedArrayStringSlice = typx.Instantiate(_tTypedArray, types.NewSlice(tString))

	rTypedArrayStringArray = reflect.TypeFor[testdata.TypedArray[[2]string]]()
	tTypedArrayStringArray = typx.Instantiate(_tTypedArray, types.NewArray(tString, 2))

	rTypedArrayMapIntString = reflect.TypeFor[testdata.TypedArray[map[int]string]]()
	tTypedArrayMapIntString = typx.Instantiate(_tTypedArray, types.NewMap(tInt, tString))

	rTypedArrayChanError = reflect.TypeFor[testdata.TypedArray[chan error]]()
	tTypedArrayChanError = typx.Instantiate(_tTypedArray, tChanError)

	rTypedArrayChanTagged = reflect.TypeFor[testdata.TypedArray[chan<- testdata.Tagged]]()
	tTypedArrayChanTagged = typx.Instantiate(_tTypedArray, types.NewChan(types.SendOnly, tTagged))

	rTypedArrayChanTaggedPointer = reflect.TypeFor[testdata.TypedArray[<-chan *testdata.Tagged]]()
	tTypedArrayChanTaggedPointer = typx.Instantiate(_tTypedArray, types.NewChan(types.RecvOnly, types.NewPointer(tTagged)))

	rTypedArrayEmptyStruct = reflect.TypeFor[testdata.TypedArray[struct{}]]()
	tTypedArrayEmptyStruct = typx.Instantiate(_tTypedArray, tEmptyStruct)

	rTypedArrayUnnamedStruct = reflect.TypeFor[testdata.TypedArray[struct {
		A            string
		B            int
		testdata.Map `json:"esc''{}[]\""`
		testdata.TypedArray[net.Addr]
		C struct {
			testdata.TypedArray[struct {
				testdata.TypedArray[fmt.Stringer]
			}]
		}
		D interface {
			fmt.Stringer
			io.ReadCloser
			io.WriteCloser
			io.ReadWriter
		}
	}]]()
	tTypedArrayUnnamedStruct = typx.Instantiate(_tTypedArray, tUnnamedStruct)
	wTypedArrayUnnamedStruct = `github_com_xoctopus_typx_testdata.TypedArray[` + wUnnamedStruct + `]`
	oTypedArrayUnnamedStruct = `github.com/xoctopus/typx/testdata.TypedArray[` + oUnnamedStruct + `]`

	rTypedArrayEmptyInterface = reflect.TypeFor[testdata.TypedArray[any]]()
	tTypedArrayEmptyInterface = typx.Instantiate(_tTypedArray, tEmptyInterface)

	rTypedArrayUnnamedInterface = reflect.TypeFor[testdata.TypedArray[interface {
		fmt.Stringer
		io.ReadCloser
		io.WriteCloser
		io.ReadWriter
	}]]()
	tTypedArrayUnnamedInterface = typx.Instantiate(_tTypedArray, tUnnamedInterfaceComposer)

	rTypedArrayFunc = reflect.TypeFor[testdata.TypedArray[func()]]()
	tTypedArrayFunc = typx.Instantiate(_tTypedArray, tFunc)

	rTypedArrayFuncVariadic = reflect.TypeFor[testdata.TypedArray[func(fmt.Stringer, ...any)]]()
	tTypedArrayFuncVariadic = typx.Instantiate(_tTypedArray, tFuncVariadic)

	rTypedArrayFuncWithMultiReturn = reflect.TypeFor[testdata.TypedArray[func(int, ...any) (bool, error)]]()
	tTypedArrayFuncWithMultiReturn = typx.Instantiate(_tTypedArray, tFuncWithMultiReturn)
)

var LitTypeCases = []struct {
	name    string
	rt      reflect.Type
	tt      types.Type
	wrapped string
	origin  string
	PkgPath string
	Name    string
	Dump    string
}{
	{
		name:    "Bool",
		rt:      rBool,
		tt:      tBool,
		wrapped: "bool",
		origin:  "bool",
		PkgPath: "",
		Name:    "bool",
		Dump:    "bool",
	},
	{
		name:    "Int",
		rt:      rInt,
		tt:      tInt,
		wrapped: "int",
		origin:  "int",
		PkgPath: "",
		Name:    "int",
		Dump:    "int",
	},
	{
		name:    "String",
		rt:      rString,
		tt:      tString,
		wrapped: "string",
		origin:  "string",
		PkgPath: "",
		Name:    "string",
		Dump:    "string",
	},
	{
		name:    "Rune",
		rt:      rRune,
		tt:      tRune,
		wrapped: "int32",
		origin:  "int32",
		PkgPath: "",
		Name:    "int32",
		Dump:    "int32",
	},
	{
		name:    "Byte",
		rt:      rByte,
		tt:      tByte,
		wrapped: "uint8",
		origin:  "uint8",
		PkgPath: "",
		Name:    "uint8",
		Dump:    "uint8",
	},
	{
		name:    "UnsafePoint",
		rt:      rUnsafePointer,
		tt:      tUnsafePointer,
		wrapped: "unsafe.Pointer",
		origin:  "unsafe.Pointer",
		PkgPath: "unsafe",
		Name:    "Pointer",
		Dump:    "unsafe.Pointer",
	},
	{
		name:    "EmptyInterface",
		rt:      rEmptyInterface,
		tt:      tEmptyInterface,
		wrapped: "interface {}",
		origin:  "interface {}",
		Name:    "",
		PkgPath: "",
		Dump:    "interface {}",
	},
	{
		name:    "EmptyStruct",
		rt:      rEmptyStruct,
		tt:      tEmptyStruct,
		wrapped: "struct {}",
		origin:  "struct {}",
		Name:    "",
		PkgPath: "",
		Dump:    "struct {}",
	},
	{
		name:    "Error",
		rt:      rError,
		tt:      tError,
		wrapped: "error",
		origin:  "error",
		PkgPath: "",
		Name:    "error",
		Dump:    "error",
	},
	{
		name:    "FmtStringer",
		rt:      rFmtStringer,
		tt:      tFmtStringer,
		wrapped: "fmt.Stringer",
		origin:  "fmt.Stringer",
		PkgPath: "fmt",
		Name:    "Stringer",
		Dump:    "fmt.Stringer",
	},
	{
		name:    "NetAddr",
		rt:      rNetAddr,
		tt:      tNetAddr,
		wrapped: "net.Addr",
		origin:  "net.Addr",
		PkgPath: "net",
		Name:    "Addr",
		Dump:    "net.Addr",
	},
	{
		name:    "IoReadCloser",
		rt:      rIoReadCloser,
		tt:      tIoReadCloser,
		wrapped: "io.ReadCloser",
		origin:  "io.ReadCloser",
		PkgPath: "io",
		Name:    "ReadCloser",
		Dump:    "io.ReadCloser",
	},
	{
		name:    "IoWriteCloser",
		rt:      rIoWriteCloser,
		tt:      tIoWriteCloser,
		wrapped: "io.WriteCloser",
		origin:  "io.WriteCloser",
		PkgPath: "io",
		Name:    "WriteCloser",
		Dump:    "io.WriteCloser",
	},
	{
		name:    "IoReadWriter",
		rt:      rIoReadWriter,
		tt:      tIoReadWriter,
		wrapped: "io.ReadWriter",
		origin:  "io.ReadWriter",
		PkgPath: "io",
		Name:    "ReadWriter",
		Dump:    "io.ReadWriter",
	},
	{
		name:    "TestdataTagged",
		rt:      rTagged,
		tt:      tTagged,
		wrapped: "github_com_xoctopus_typx_testdata.Tagged",
		origin:  "github.com/xoctopus/typx/testdata.Tagged",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "Tagged",
		Dump:    "testdata.Tagged",
	},
	{
		name:    "StringSlice",
		rt:      rStringSlice,
		tt:      tStringSlice,
		wrapped: "[]string",
		origin:  "[]string",
		PkgPath: "",
		Name:    "",
		Dump:    "[]string",
	},
	{
		name:    "MapRuneString",
		rt:      rMapRuneString,
		tt:      tMapRuneString,
		wrapped: "map[int32]string",
		origin:  "map[int32]string",
		PkgPath: "",
		Name:    "",
		Dump:    "map[int32]string",
	},
	{
		name:    "Func",
		rt:      rFunc,
		tt:      tFunc,
		wrapped: "func()",
		origin:  "func()",
		PkgPath: "",
		Name:    "",
		Dump:    "func()",
	},
	{
		name:    "FuncVariadic",
		rt:      rFuncVariadic,
		tt:      tFuncVariadic,
		wrapped: "func(fmt.Stringer, ...interface {})",
		origin:  "func(fmt.Stringer, ...interface {})",
		PkgPath: "",
		Name:    "",
		Dump:    "func()",
	},
	{
		name:    "FuncWithMultiReturn",
		rt:      rFuncWithMultiReturn,
		tt:      tFuncWithMultiReturn,
		wrapped: "func(int, ...interface {}) (bool, error)",
		origin:  "func(int, ...interface {}) (bool, error)",
		PkgPath: "",
		Name:    "",
		Dump:    "func(int, ...interface {}) (bool, error)",
	},
	{
		name:    "FuncWithOneReturn",
		rt:      rFuncWithOneReturn,
		tt:      tFuncWithOneReturn,
		wrapped: "func(fmt.Stringer, ...interface {}) net.Addr",
		origin:  "func(fmt.Stringer, ...interface {}) net.Addr",
		PkgPath: "",
		Name:    "",
		Dump:    "func(fmt.Stringer, ...interface {}) net.Addr",
	},
	{
		name:    "IterSeqArray",
		rt:      rIterSeqArray,
		tt:      tIterSeqArray,
		wrapped: "[3]iter.Seq[int]",
		origin:  "[3]iter.Seq[int]",
		PkgPath: "",
		Name:    "",
		Dump:    "[3]iter.Seq[int]",
	},
	{
		name:    "IterSeq2StringEmptyInterface",
		rt:      rIterSeq2StringEmptyInterface,
		tt:      tIterSeq2StringEmptyInterface,
		wrapped: "iter.Seq2[string,interface {}]",
		origin:  "iter.Seq2[string,interface {}]",
		PkgPath: "iter",
		Name:    "Seq2[string,interface {}]",
		Dump:    "iter.Seq2[string,interface {}]",
	},
	{
		name:    "ChanError",
		rt:      rChanError,
		tt:      tChanError,
		wrapped: "chan error",
		origin:  "chan error",
		PkgPath: "",
		Name:    "",
		Dump:    "chan error",
	},
	{
		name:    "SendChanTestdataTagged",
		rt:      rSendChanTestdataTagged,
		tt:      tSendChanTestdataTagged,
		wrapped: "chan<- github_com_xoctopus_typx_testdata.Tagged",
		origin:  "chan<- github.com/xoctopus/typx/testdata.Tagged",
		PkgPath: "",
		Name:    "",
		Dump:    "chan<- testdata.Tagged",
	},
	{
		name:    "RecvChanTestdataTagged",
		rt:      rRecvChanTestdataTaggedPointer,
		tt:      tRecvChanTestdataTaggedPointer,
		wrapped: "<-chan *github_com_xoctopus_typx_testdata.Tagged",
		origin:  "<-chan *github.com/xoctopus/typx/testdata.Tagged",
		PkgPath: "",
		Name:    "",
		Dump:    "<-chan *testdata.Tagged",
	},
	{
		name:    "UnnamedInterfaceComposer",
		rt:      rUnnamedInterfaceComposer,
		tt:      tUnnamedInterfaceComposer,
		wrapped: wUnnamedInterfaceComposer,
		origin:  oUnnamedInterfaceComposer,
		PkgPath: "",
		Name:    "",
		Dump:    dUnnamedInterfaceComposer,
	},
	{
		name:    "TestdataTypedSliceAliasNetAddr",
		rt:      rTypedSliceAliasNetAddr,
		tt:      tTypedSliceAliasNetAddr,
		wrapped: "github_com_xoctopus_typx_testdata.TypedSlice[net.Addr]",
		origin:  "github.com/xoctopus/typx/testdata.TypedSlice[net.Addr]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedSlice[net.Addr]",
		Dump:    "testdata.TypedSlice[net.Addr]",
	},
	{
		name:    "TestdataMap",
		rt:      rMap,
		tt:      tMap,
		wrapped: "github_com_xoctopus_typx_testdata.Map",
		origin:  "github.com/xoctopus/typx/testdata.Map",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "Map",
		Dump:    "testdata.Map",
	},
	{
		name:    "UnnamedStruct",
		rt:      rUnnamedStruct,
		tt:      tUnnamedStruct,
		wrapped: wUnnamedStruct,
		origin:  oUnnamedStruct,
		PkgPath: "",
		Name:    "",
		Dump:    dUnnamedStruct,
	},
	{
		name:    "TypedArrayFmtString",
		rt:      rTypedArrayFmtString,
		tt:      tTypedArrayFmtString,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[fmt.Stringer]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[fmt.Stringer]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[fmt.Stringer]",
		Dump:    "testdata.TypedArray[fmt.Stringer]",
	},
	{
		name:    "TypedArrayStringSlice",
		rt:      rTypedArrayStringSlice,
		tt:      tTypedArrayStringSlice,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[[]string]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[[]string]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[[]string]",
		Dump:    "testdata.TypedArray[[]string]",
	},
	{
		name:    "TypedArrayStringArray",
		rt:      rTypedArrayStringArray,
		tt:      tTypedArrayStringArray,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[[2]string]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[[2]string]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[[2]string]",
		Dump:    "testdata.TypedArray[[2]string]",
	},
	{
		name:    "TypedArrayMapIntString",
		rt:      rTypedArrayMapIntString,
		tt:      tTypedArrayMapIntString,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[map[int]string]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[map[int]string]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[map[int]string]",
		Dump:    "testdata.TypedArray[map[int]string]",
	},
	{
		name:    "TypedArrayChanError",
		rt:      rTypedArrayChanError,
		tt:      tTypedArrayChanError,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[chan error]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[chan error]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[chan error]",
		Dump:    "testdata.TypedArray[chan error]",
	},
	{
		name:    "TypedArrayChanTagged",
		rt:      rTypedArrayChanTagged,
		tt:      tTypedArrayChanTagged,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[chan<- github_com_xoctopus_typx_testdata.Tagged]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[chan<- github.com/xoctopus/typx/testdata.Tagged]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[chan<- github.com/xoctopus/typx/testdata.Tagged]",
		Dump:    "testdata.TypedArray[chan<- testdata.Tagged]",
	},
	{
		name:    "TypedArrayChanTaggedPointer",
		rt:      rTypedArrayChanTaggedPointer,
		tt:      tTypedArrayChanTaggedPointer,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[<-chan *github_com_xoctopus_typx_testdata.Tagged]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[<-chan *github.com/xoctopus/typx/testdata.Tagged]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[<-chan *github.com/xoctopus/typx/testdata.Tagged]",
		Dump:    "testdata.TypedArray[<-chan *testdata.Tagged]",
	},
	{
		name:    "TypedArrayEmptyStruct",
		rt:      rTypedArrayEmptyStruct,
		tt:      tTypedArrayEmptyStruct,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[struct {}]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[struct {}]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[struct {}]",
		Dump:    "testdata.TypedArray[struct {}]",
	},
	{
		name:    "TypedArrayUnnamedStruct",
		rt:      rTypedArrayUnnamedStruct,
		tt:      tTypedArrayUnnamedStruct,
		wrapped: wTypedArrayUnnamedStruct,
		origin:  oTypedArrayUnnamedStruct,
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[" + oUnnamedStruct + "]",
		Dump:    "testdata.TypedArray[" + dUnnamedStruct + "]",
	},
	{
		name:    "TypedArrayEmptyInterface",
		rt:      rTypedArrayEmptyInterface,
		tt:      tTypedArrayEmptyInterface,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[interface {}]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[interface {}]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[interface {}]",
		Dump:    "testdata.TypedArray[interface {}]",
	},
	{
		name:    "TypedArrayUnnamedInterface",
		rt:      rTypedArrayUnnamedInterface,
		tt:      tTypedArrayUnnamedInterface,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[" + wUnnamedInterfaceComposer + "]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[" + oUnnamedInterfaceComposer + "]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[" + oUnnamedInterfaceComposer + "]",
		Dump:    "testdata.TypedArray[" + dUnnamedInterfaceComposer + "]",
	},
	{
		name:    "TypedArrayFunc",
		rt:      rTypedArrayFunc,
		tt:      tTypedArrayFunc,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[func()]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[func()]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[func()]",
		Dump:    "testdata.TypedArray[func()]",
	},
	{
		name:    "TypedArrayFuncVariadic",
		rt:      rTypedArrayFuncVariadic,
		tt:      tTypedArrayFuncVariadic,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[func(fmt.Stringer, ...interface {})]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[func(fmt.Stringer, ...interface {})]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[func(fmt.Stringer, ...interface {})]",
		Dump:    "testdata.TypedArray[func(fmt.Stringer, ...interface {})]",
	},
	{
		name:    "TypedArrayFuncWithMultiReturn",
		rt:      rTypedArrayFuncWithMultiReturn,
		tt:      tTypedArrayFuncWithMultiReturn,
		wrapped: "github_com_xoctopus_typx_testdata.TypedArray[func(int, ...interface {}) (bool, error)]",
		origin:  "github.com/xoctopus/typx/testdata.TypedArray[func(int, ...interface {}) (bool, error)]",
		PkgPath: "github.com/xoctopus/typx/testdata",
		Name:    "TypedArray[func(int, ...interface {}) (bool, error)]",
		Dump:    "testdata.TypedArray[func(int, ...interface {}) (bool, error)]",
	},
}

func TestWrap(t *testing.T) {
	for _, c := range LitTypeCases {
		t.Run(c.name, func(t *testing.T) {
			rt := typx.Wrap(c.rt)
			tt := typx.Wrap(c.tt)

			// t.Log(c.expect)
			// t.Log(rt)
			// t.Log(tt)

			Expect(t, rt, Equal(c.wrapped))
			Expect(t, tt, Equal(c.wrapped))
		})
	}
	ExpectPanic[error](t, func() { typx.Wrap(1) })
}
