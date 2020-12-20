package ber

import (
	"bytes"
	"math"
	"testing"
	"time"
)

func TestEncoder(t *testing.T) {
	t.Run("string", testEncodeString)
	t.Run("bool", testEncodeBool)
	t.Run("null", testEncodeNull)
	t.Run("int", testEncodeInt)
	t.Run("real-base-2", testEncodeRealBase2)
	t.Run("real-base-10", testEncodeRealBase10)
	t.Run("time", testEncodeTime)
	t.Run("oid", testEncodeOID)
	t.Run("as", testEncodeAs)
	t.Run("array/int", testEncodeArrayInt)
	t.Run("array/string", testEncodeArrayString)
	t.Run("struct", testEncodeStruct)
}

func testEncodeAs(t *testing.T) {
	var (
		e    Encoder
		body = []byte{
			0x01, 0x01, 0xFF, 0x17, 0x11, 0x31, 0x39, 0x31, 0x32,
			0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b,
			0x30, 0x30, 0x30, 0x30, 0x0c, 0x06, 'f', 'o', 'o', 'b',
			'a', 'r', 0x02, 0x01, 0x80, 0x09, 0x03, 0x80, 0xfd, 0x05,
		}
		seq = append([]byte{0x30, 0x26}, body...) // 0011 0000
		set = append([]byte{0x31, 0x26}, body...)
	)
	e.EncodeBool(true)
	e.EncodeUniversalTime(time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC))
	e.EncodeStringUTF8("foobar")
	e.EncodeInt(-128)
	e.EncodeFloat2(0.625)

	if got, err := e.AsSequence(); err != nil {
		t.Errorf("sequence: fail to build sequence! %s", err)
	} else {
		if !bytes.Equal(got, seq) {
			t.Errorf("sequence: bytes mismatched! want %x, got %x", seq, got)
		}
	}
	if got, err := e.AsSet(); err != nil {
		t.Errorf("set: fail to build sequence! %s", err)
	} else {
		if !bytes.Equal(got, set) {
			t.Errorf("sequence: bytes mismatched! want %x, got %x", seq, got)
		}
	}
}

func testEncodeOID(t *testing.T) {
	data := []struct {
		Input string
		Id    Ident
		Want  []byte
	}{
		{
			Input: "1.2.840.113549.1.1.11",
			Id:    ObjectId,
			Want:  []byte{0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x01, 0x0b},
		},
		{
			Input: ".8571.3.2",
			Id:    RelObjectId,
			Want:  []byte{0x0d, 0x04, 0xc2, 0x7b, 0x03, 0x02},
		},
	}
	for _, d := range data {
		var e Encoder
		if err := e.EncodeOIDWithIdent(d.Input, d.Id); err != nil {
			t.Errorf("%s: fail to encode oid! %s", d.Input, err)
			continue
		}
		got := e.Bytes()
		if !bytes.Equal(d.Want, got) {
			t.Errorf("%s: bytes mismatched! want %x, got %x", d.Input, d.Want, got)
		}
	}
}

func testEncodeTime(t *testing.T) {
	data := []struct {
		Input time.Time
		Id    Ident
		Want  []byte
	}{
		{
			Input: time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC),
			Id:    UniversalTime,
			Want:  []byte{0x17, 0x11, 0x31, 0x39, 0x31, 0x32, 0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b, 0x30, 0x30, 0x30, 0x30},
		},
		{
			Input: time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC),
			Id:    GeneralizedTime,
			Want:  []byte{0x18, 0x13, 0x32, 0x30, 0x31, 0x39, 0x31, 0x32, 0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b, 0x30, 0x30, 0x30, 0x30},
		},
	}
	for _, d := range data {
		var e Encoder
		if err := e.EncodeTimeWithIdent(d.Input, d.Id); err != nil {
			t.Errorf("%s: fail to encode time! %s", d.Input, err)
			continue
		}
		got := e.Bytes()
		if !bytes.Equal(got, d.Want) {
			t.Errorf("%s: bytes mismatched! want: %x, got %x", d.Input, d.Want, got)
		}
	}
}

func testEncodeRealBase10(t *testing.T) {
	data := []struct {
		Input float64
		Want  []byte
	}{
		{Input: 0, Want: []byte{0x09, 0x00}},
		{Input: math.Copysign(0, -1), Want: []byte{0x09, 0x01, 0x43}},
		{Input: math.Inf(1), Want: []byte{0x09, 0x01, 0x40}},
		{Input: math.Inf(-1), Want: []byte{0x09, 0x01, 0x41}},
		{Input: math.NaN(), Want: []byte{0x09, 0x01, 0x42}},
		{Input: 1.0, Want: []byte{0x09, 0x02, 0x01, 0x31}},
		{Input: 100.0, Want: []byte{0x09, 0x04, 0x01, 0x31, 0x30, 0x30}},
		{Input: -100.0, Want: []byte{0x09, 0x05, 0x01, 0x2d, 0x31, 0x30, 0x30}},
		{Input: 0.15625, Want: []byte{0x09, 0x08, 0x02, 0x30, 0x2e, 0x31, 0x35, 0x36, 0x32, 0x35}},
		{Input: 1234.5678, Want: []byte{0x09, 0x0a, 0x02, 0x31, 0x32, 0x33, 0x34, 0x2e, 0x35, 0x36, 0x37, 0x38}},
		{Input: 0.625, Want: []byte{0x09, 0x06, 0x02, 0x30, 0x2e, 0x36, 0x32, 0x35}},
		{Input: 8, Want: []byte{0x09, 0x02, 0x01, 0x38}},
	}
	for _, d := range data {
		var e Encoder
		if err := e.EncodeFloat10(d.Input); err != nil {
			t.Errorf("%f: fail to encode real! %s", d.Input, err)
			continue
		}
		got := e.Bytes()
		if !bytes.Equal(got, d.Want) {
			t.Errorf("%f: bytes mismatched! want: %x, got %x", d.Input, d.Want, got)
		}
	}
}

func testEncodeRealBase2(t *testing.T) {
	data := []struct {
		Input float64
		Want  []byte
	}{
		{Input: 0, Want: []byte{0x09, 0x00}},
		{Input: math.Copysign(0, -1), Want: []byte{0x09, 0x01, 0x43}},
		{Input: math.Inf(1), Want: []byte{0x09, 0x01, 0x40}},
		{Input: math.Inf(-1), Want: []byte{0x09, 0x01, 0x41}},
		{Input: math.NaN(), Want: []byte{0x09, 0x01, 0x42}},
		{Input: 1.0, Want: []byte{0x09, 0x03, 0x80, 0x00, 0x01}},
		{Input: 100.0, Want: []byte{0x09, 0x03, 0x80, 0x02, 0x19}},
		{Input: -100.0, Want: []byte{0x09, 0x03, 0xc0, 0x02, 0x19}},
		{Input: 0.15625, Want: []byte{0x09, 0x03, 0x80, 0xfb, 0x05}},
		{Input: 1234.5678, Want: []byte{0x09, 0x09, 0x80, 0xd6, 0x13, 0x4a, 0x45, 0x6d, 0x5c, 0xfa, 0xad}},
		{Input: 0.625, Want: []byte{0x09, 0x03, 0x80, 0xfd, 0x05}},
		{Input: 8, Want: []byte{0x09, 0x03, 0x80, 0x03, 0x01}},
	}
	for _, d := range data {
		var e Encoder
		if err := e.EncodeFloat2(d.Input); err != nil {
			t.Errorf("%f: fail to encode real! %s", d.Input, err)
			continue
		}
		got := e.Bytes()
		if !bytes.Equal(got, d.Want) {
			t.Errorf("%f: bytes mismatched! want: %x, got %x", d.Input, d.Want, got)
		}
	}
}

func testEncodeInt(t *testing.T) {
	data := []struct {
		Input int64
		Want  []byte
	}{
		{Input: 0, Want: []byte{0x02, 0x01, 0x00}},
		{Input: 127, Want: []byte{0x02, 0x01, 0x7F}},
		{Input: 128, Want: []byte{0x02, 0x02, 0x00, 0x80}},
		{Input: 256, Want: []byte{0x02, 0x02, 0x01, 0x00}},
		{Input: -128, Want: []byte{0x02, 0x01, 0x80}},
		{Input: -129, Want: []byte{0x02, 0x02, 0xFF, 0x7F}},
		{Input: 56, Want: []byte{0x02, 0x01, 0x38}},
		{Input: -56, Want: []byte{0x02, 0x01, 0xc8}},
		{Input: 512456, Want: []byte{0x02, 0x03, 0x07, 0xd1, 0xc8}},
		{Input: -512456, Want: []byte{0x02, 0x03, 0xf8, 0x2e, 0x38}},
	}
	for _, d := range data {
		var e Encoder
		if err := e.EncodeInt(d.Input); err != nil {
			t.Errorf("%d: fail to encode int! %s", d.Input, err)
			continue
		}
		got := e.Bytes()
		if !bytes.Equal(got, d.Want) {
			t.Errorf("%d: bytes mismatched! want: %x, got %x", d.Input, d.Want, got)
		}
	}
}

func testEncodeStruct(t *testing.T) {
	var (
		data = struct {
			Int  int8
			Uint uint8          `ber:"class:0x3"`
			Str  string         `ber:"printable,type:0x1"`
			Oid  string         `ber:"oid"`
			When time.Time      `ber:"generalized"`
			Set  map[string]int `ber:"set"`
			Arr  []interface{}  `ber:"set,omitempty"`
			Bool bool
			Ifi  interface{}
			omit string
		}{
			Int:  -128,
			Uint: 127,
			Str:  "foobar",
			Oid:  "1.2.840.113549.1.1.11",
			When: time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC),
			Set: map[string]int{
				"foo": 128,
				"bar": -128,
			},
			Bool: false,
			Ifi:  nil,
			omit: "unexported",
		}
		want = []byte{
			0x30, 0x46,
			0x02, 0x01, 0x80, // int
			0xc2, 0x01, 0x7F, // uint
			0x33, 0x06, 'f', 'o', 'o', 'b', 'a', 'r', // string
			0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x01, 0x0b, // oid: invalid encoded via stringwithident instead of oidwithident
			0x18, 0x13, 0x32, 0x30, 0x31, 0x39, 0x31, 0x32, 0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b, 0x30, 0x30, 0x30, 0x30, // time
			0x31, 0x11, 0x0c, 0x03, 'f', 'o', 'o', 0x02, 0x02, 0x00, 0x80, 0x0c, 0x03, 'b', 'a', 'r', 0x02, 0x01, 0x80, // map[string]int
			0x01, 0x01, 0x00, // bool
			0x05, 0x00, // nil
		}
		e Encoder
	)
	if err := e.EncodeWithIdent(data, Sequence); err != nil {
		t.Errorf("struct: encoding failed! %s", err)
		return
	}
	if got := e.Bytes(); !bytes.Equal(got, want) {
		t.Errorf("bytes mismatched!")
		t.Logf("want: %x", want)
		t.Logf("got:  %x", got)
	}
}

func testEncodeArrayInt(t *testing.T) {
	var (
		arr  = []int64{0, 127, -128, 56, -512456}
		want = []byte{
			0x30, 0x11,
			0x02, 0x01, 0x00,
			0x02, 0x01, 0x7F,
			0x02, 0x01, 0x80,
			0x02, 0x01, 0x38,
			0x02, 0x03, 0xf8, 0x2e, 0x38,
		}
		e Encoder
	)
	if err := e.Encode(arr); err != nil {
		t.Errorf("array: fail to encode! %s", err)
		return
	}
	if got := e.Bytes(); !bytes.Equal(got, want) {
		t.Errorf("bytes mismatched!")
		t.Logf("want: %x", want)
		t.Logf("got:  %x", got)
	}
}

func testEncodeArrayString(t *testing.T) {
	var (
		arr  = []string{"foo", "bar"}
		want = []byte{
			0x30, 0x0a,
			0x0c, 0x03, 'f', 'o', 'o',
			0x0c, 0x03, 'b', 'a', 'r',
		}
		e Encoder
	)
	if err := e.Encode(arr); err != nil {
		t.Errorf("array: fail to encode! %s", err)
		return
	}
	if got := e.Bytes(); !bytes.Equal(got, want) {
		t.Errorf("bytes mismatched!")
		t.Logf("want: %x", want)
		t.Logf("got:  %x", got)
	}
}

func testEncodeNull(t *testing.T) {
	var (
		e    Encoder
		want = []byte{0x05, 0x00}
	)
	if err := e.EncodeNull(); err != nil {
		t.Errorf("null: fail to encode null! %s", err)
		return
	}
	got := e.Bytes()
	if !bytes.Equal(got, want) {
		t.Errorf("null: bytes mismatched! want: %x, got %x", want, got)
	}
}

func testEncodeBool(t *testing.T) {
	data := []struct {
		Input bool
		Id    Ident
		Want  []byte
	}{
		{
			Input: true,
			Id:    Bool,
			Want:  []byte{0x01, 0x01, 0xFF},
		},
		{
			Input: false,
			Id:    Bool,
			Want:  []byte{0x01, 0x01, 0x00},
		},
	}

	for _, d := range data {
		var e Encoder
		if err := e.EncodeBoolWithIdent(d.Input, d.Id); err != nil {
			t.Errorf("%t: fail to encode bool! %s", d.Input, err)
			continue
		}
		got := e.Bytes()
		if !bytes.Equal(got, d.Want) {
			t.Errorf("%t: bytes mismatched! want: %x, got %x", d.Input, d.Want, got)
		}
	}
}

func testEncodeString(t *testing.T) {
	data := []struct {
		Input string
		Id    Ident
		Want  []byte
	}{
		{
			Input: "foobar",
			Id:    UTF8String,
			Want:  []byte{0x0c, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
		},
		{
			Input: "foobar",
			Id:    IA5String,
			Want:  []byte{0x16, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
		},
		{
			Input: "foobar",
			Id:    PrintableString,
			Want:  []byte{0x13, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
		},
		{
			Input: "foobar",
			Id:    UTF8String.Constructed(),
			Want:  []byte{0x2c, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
		},
		{
			Input: "foobar",
			Id:    IA5String.Constructed(),
			Want:  []byte{0x36, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
		},
		{
			Input: "foobar",
			Id:    PrintableString.Constructed(),
			Want:  []byte{0x33, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
		},
	}

	for _, d := range data {
		var e Encoder
		if err := e.EncodeStringWithIdent(d.Input, d.Id); err != nil {
			t.Errorf("%s: fail to encode string! %s", d.Input, err)
			continue
		}
		got := e.Bytes()
		if !bytes.Equal(got, d.Want) {
			t.Errorf("%s: bytes mismatched! want: %x, got %x", d.Input, d.Want, got)
		}
	}
}
