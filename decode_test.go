package ber

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestDecoder(t *testing.T) {
	t.Run("null", testDecodeNull)
	t.Run("bool", testDecodeBool)
	t.Run("int", testDecodeInt)
	t.Run("string", testDecodeString)
	t.Run("time", testDecodeTime)
	t.Run("oid", testDecodeOID)
	t.Run("float", testDecodeFloat)
	t.Run("struct", testDecodeStruct)
	t.Run("map", testDecodeMap)
	t.Run("slice", testDecodeSlice)
}

func encodeValue(val interface{}) ([]byte, error) {
	var e Encoder
	if err := e.Encode(val); err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

func testDecodeSlice(t *testing.T) {
	want := []string{"foo", "bar", "ber", "der", "per", "xer"}
	input, err := encodeValue(want)
	if err != nil {
		t.Errorf("slice: fail to encode value %+v! %s", want, err)
		return
	}
	data := [][]string{
		nil,
		make([]string, 0),
		make([]string, 3),
	}
	for i := range data {
		d := NewDecoder(input)
		if err := d.Decode(&data[i]); err != nil {
			t.Errorf("slice: fail to decode! %s", err)
			return
		}
		if !reflect.DeepEqual(data[i], want) {
			t.Errorf("slices mismatched! want %+v, got %+v", want, data[i])
		}
	}
}

func testDecodeMap(t *testing.T) {
	want := map[string]int{
		"foo": 127,
		"bar": -128,
	}
	input, err := encodeValue(want)
	if err != nil {
		t.Errorf("map: fail to encode value %+v! %s", want, err)
		return
	}
	var got map[string]int
	d := NewDecoder(input)
	if err := d.Decode(&got); err != nil {
		t.Errorf("map: fail to decode! %s", err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("map mismatched! want %+v, got %+v", want, got)
	}
}

func testDecodeStruct(t *testing.T) {
	type Sample struct {
		Str   string
		Int   int64
		Float float64
		Bool  bool
		When  time.Time
	}
	var (
		want = Sample{
			Str:   "ber",
			Int:   127,
			Bool:  true,
			Float: 3.14,
			When:  time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC),
		}
		got Sample
	)
	input, err := encodeValue(want)
	if err != nil {
		t.Errorf("struct: fail to encode value %+v! %s", want, err)
		return
	}
	d := NewDecoder(input)
	if err := d.Decode(&got); err != nil {
		t.Errorf("struct: fail to decode! %s", err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("struct mismatched! want %+v, got %+v", want, got)
	}
}

func testDecodeFloat(t *testing.T) {
	data := []struct {
		Input []byte
		Want  float64
	}{
		{Input: []byte{0x09, 0x00}, Want: 0},
		{Input: []byte{0x09, 0x01, 0x43}, Want: math.Copysign(0, -1)},
		{Input: []byte{0x09, 0x01, 0x40}, Want: math.Inf(1)},
		{Input: []byte{0x09, 0x01, 0x41}, Want: math.Inf(-1)},
		{Input: []byte{0x09, 0x01, 0x42}, Want: math.NaN()},
		{Input: []byte{0x09, 0x03, 0x80, 0x00, 0x01}, Want: 1.0},
		{Input: []byte{0x09, 0x03, 0x80, 0x02, 0x19}, Want: 100.0},
		{Input: []byte{0x09, 0x03, 0xc0, 0x02, 0x19}, Want: -100.0},
		{Input: []byte{0x09, 0x03, 0x80, 0xfb, 0x05}, Want: 0.15625},
		{Input: []byte{0x09, 0x09, 0x80, 0xd6, 0x13, 0x4a, 0x45, 0x6d, 0x5c, 0xfa, 0xad}, Want: 1234.5678},
		{Input: []byte{0x09, 0x03, 0x80, 0xfd, 0x05}, Want: 0.625},
		{Input: []byte{0x09, 0x03, 0x80, 0x03, 0x01}, Want: 8},
		{Input: []byte{0x09, 0x02, 0x01, 0x31}, Want: 1.0},
		{Input: []byte{0x09, 0x04, 0x01, 0x31, 0x30, 0x30}, Want: 100.0},
		{Input: []byte{0x09, 0x05, 0x01, 0x2d, 0x31, 0x30, 0x30}, Want: -100.0},
		{Input: []byte{0x09, 0x08, 0x02, 0x30, 0x2e, 0x31, 0x35, 0x36, 0x32, 0x35}, Want: 0.15625},
		{Input: []byte{0x09, 0x0a, 0x02, 0x31, 0x32, 0x33, 0x34, 0x2e, 0x35, 0x36, 0x37, 0x38}, Want: 1234.5678},
		{Input: []byte{0x09, 0x06, 0x02, 0x30, 0x2e, 0x36, 0x32, 0x35}, Want: 0.625},
		{Input: []byte{0x09, 0x02, 0x01, 0x38}, Want: 8},
	}
	for _, d := range data {
		dec := NewDecoder(d.Input)
		got, err := dec.DecodeFloat()
		if err != nil {
			t.Errorf("float: fail to decode! %s", err)
			continue
		}
		if math.Float64bits(d.Want) != math.Float64bits(got) {
			t.Errorf("float mismatched! want: %f, got %f", d.Want, got)
		}
	}
}

func testDecodeOID(t *testing.T) {
	data := []struct {
		Input []byte
		Want  string
	}{
		{
			Input: []byte{0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x01, 0x0b},
			Want:  "1.2.840.113549.1.1.11",
		},
		{
			Input: []byte{0x0d, 0x04, 0xc2, 0x7b, 0x03, 0x02},
			Want:  ".8571.3.2",
		},
	}
	for _, d := range data {
		dec := NewDecoder(d.Input)
		got, err := dec.DecodeOID()
		if err != nil {
			t.Errorf("oid: fail to decode! %s", err)
			continue
		}
		if got != d.Want {
			t.Errorf("oid mismatched! want: %s, got %s", d.Want, got)
		}
	}
}

func testDecodeNull(t *testing.T) {
	var (
		null = []byte{0x05, 0x00}
		d    = NewDecoder(null)
	)
	if err := d.DecodeNull(); err != nil {
		t.Errorf("null: fail to decode null! %s", err)
	}
}

func testDecodeBool(t *testing.T) {
	data := []struct {
		Input []byte
		Want  bool
	}{
		{
			Input: []byte{0x01, 0x01, 0xFF},
			Want:  true,
		},
		{
			Input: []byte{0x01, 0x01, 0x00},
			Want:  false,
		},
	}
	for _, d := range data {
		dec := NewDecoder(d.Input)
		got, err := dec.DecodeBool()
		if err != nil {
			t.Errorf("bool: fail to decode! %s", err)
			continue
		}
		if got != d.Want {
			t.Errorf("bool mismatched! want: %t, got %t", d.Want, got)
		}
	}
}

func testDecodeInt(t *testing.T) {
	data := []struct {
		Input []byte
		Want  int64
	}{
		{Input: []byte{0x02, 0x01, 0x00}, Want: 0},
		{Input: []byte{0x02, 0x01, 0x7F}, Want: 127},
		{Input: []byte{0x02, 0x02, 0x00, 0x80}, Want: 128},
		{Input: []byte{0x02, 0x02, 0x01, 0x00}, Want: 256},
		{Input: []byte{0x02, 0x01, 0x80}, Want: -128},
		{Input: []byte{0x02, 0x02, 0xFF, 0x7F}, Want: -129},
		{Input: []byte{0x02, 0x01, 0x38}, Want: 56},
		{Input: []byte{0x02, 0x01, 0xc8}, Want: -56},
		{Input: []byte{0x02, 0x03, 0x07, 0xd1, 0xc8}, Want: 512456},
		{Input: []byte{0x02, 0x03, 0xf8, 0x2e, 0x38}, Want: -512456},
	}
	for _, d := range data {
		dec := NewDecoder(d.Input)
		got, err := dec.DecodeInt()
		if err != nil {
			t.Errorf("int: fail to decode! %s", err)
			continue
		}
		if got != d.Want {
			t.Errorf("int mismatched! want: %d, got %d", d.Want, got)
		}
	}
}

func testDecodeString(t *testing.T) {
	data := []struct {
		Input []byte
		Want  string
	}{
		{
			Input: []byte{0x0c, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
			Want:  "foobar",
		},
		{
			Input: []byte{0x16, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
			Want:  "foobar",
		},
		{
			Input: []byte{0x13, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
			Want:  "foobar",
		},
		{
			Input: []byte{0x2c, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
			Want:  "foobar",
		},
		{
			Input: []byte{0x36, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
			Want:  "foobar",
		},
		{
			Input: []byte{0x33, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'},
			Want:  "foobar",
		},
	}
	for _, d := range data {
		dec := NewDecoder(d.Input)
		got, err := dec.DecodeString()
		if err != nil {
			t.Errorf("string: fail to decode! %s", err)
			continue
		}
		if got != d.Want {
			t.Errorf("string mismatched! want: %s, got %s", d.Want, got)
		}
	}
}

func testDecodeTime(t *testing.T) {
	data := []struct {
		Input []byte
		Want  time.Time
	}{
		{
			Input: []byte{0x17, 0x11, 0x31, 0x39, 0x31, 0x32, 0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b, 0x30, 0x30, 0x30, 0x30},
			Want:  time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC),
		},
		{
			Input: []byte{0x18, 0x13, 0x32, 0x30, 0x31, 0x39, 0x31, 0x32, 0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b, 0x30, 0x30, 0x30, 0x30},
			Want:  time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC),
		},
	}
	for _, d := range data {
		dec := NewDecoder(d.Input)
		got, err := dec.DecodeTime()
		if err != nil {
			t.Errorf("time: fail to decode! %s", err)
			continue
		}
		if got != d.Want {
			t.Errorf("time mismatched! want: %s, got %s", d.Want, got)
		}
	}
}
