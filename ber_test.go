// +build ignore

package ber

import (
	"bytes"
	"encoding/hex"
	"math"
	"testing"
	"time"
)

func TestEncodePrimitive(t *testing.T) {
	t.Run("bool", testEncodeBool)
	t.Run("bytes", testEncodeBytes)
	t.Run("null", testEncodeNull)
	t.Run("int", testEncodeInt)
	t.Run("string", testEncodeString)
	t.Run("time", testEncodeTime)
	t.Run("real", testEncodeReal)
	t.Run("oid", testEncodeOID)
}

func testEncodeOID(t *testing.T) {
	var (
		oid   = "1.2.840.113549.1.1.11"
		rid   = ".8571.3.2"
		want1 = []byte{0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x01, 0x0b}
		want2 = []byte{0x0d, 0x04, 0xc2, 0x7b, 0x03, 0x02}
		got   []byte
		err   error
	)
	if got, err = EncodeOID(oid); err != nil {
		t.Errorf("%s: encoding failed! %s", oid, err)
		return
	}
	compareBytes(t, oid, want1, got)

	if got, err = EncodeROID(rid); err != nil {
		t.Errorf("%s: encoding failed! %s", rid, err)
		return
	}
	compareBytes(t, rid, want2, got)
}

func testEncodeReal(t *testing.T) {
	data := []struct {
		Float float64
		Want  []byte
		Base  int
	}{
		{Float: 0, Want: []byte{0x09, 0x00}},
		{Float: math.Copysign(0, -1), Want: []byte{0x09, 0x01, 0x43}},
		{Float: math.Inf(1), Want: []byte{0x09, 0x01, 0x40}},
		{Float: math.Inf(-1), Want: []byte{0x09, 0x01, 0x41}},
		{Float: math.NaN(), Want: []byte{0x09, 0x01, 0x42}},
		{Float: 1.0, Base: 2, Want: []byte{0x09, 0x03, 0x80, 0x00, 0x01}},
		{Float: 100.0, Base: 2, Want: []byte{0x09, 0x03, 0x80, 0x02, 0x19}},
		{Float: -100.0, Base: 2, Want: []byte{0x09, 0x03, 0xc0, 0x02, 0x19}},
		{Float: 0.15625, Base: 2, Want: []byte{0x09, 0x03, 0x80, 0xfb, 0x05}},
		{Float: 1234.5678, Base: 2, Want: []byte{0x09, 0x09, 0x80, 0xd6, 0x13, 0x4a, 0x45, 0x6d, 0x5c, 0xfa, 0xad}},
		{Float: 0.625, Base: 2, Want: []byte{0x09, 0x03, 0x80, 0xfd, 0x05}},
		{Float: 8, Base: 2, Want: []byte{0x09, 0x03, 0x80, 0x03, 0x01}},
	}
	for _, d := range data {
		got, err := EncodeFloat(d.Float, d.Base)
		if err != nil {
			t.Errorf("%f(base=%d): encoding failed! %s", d.Float, d.Base, err)
			continue
		}
		compareBytes(t, d.Float, d.Want, got)
	}
}

func testEncodeTime(t *testing.T) {
	var (
		now = time.Date(2019, 12, 15, 19, 2, 10, 0, time.UTC)
		gen = []byte{0x18, 0x13, 0x32, 0x30, 0x31, 0x39, 0x31, 0x32, 0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b, 0x30, 0x30, 0x30, 0x30}
		utc = []byte{0x17, 0x11, 0x31, 0x39, 0x31, 0x32, 0x31, 0x35, 0x31, 0x39, 0x30, 0x32, 0x31, 0x30, 0x2b, 0x30, 0x30, 0x30, 0x30}
	)

	compareBytes(t, now, gen, EncodeTime(now))
	compareBytes(t, now, utc, EncodeUTC(now))
}

func testEncodeString(t *testing.T) {
	data := []struct {
		Str   string
		UTF   []byte
		IA5   []byte
		Print []byte
	}{
		{
			Str:   "foo",
			UTF:   []byte{0x0c, 0x03, 'f', 'o', 'o'},
			IA5:   []byte{0x16, 0x03, 'f', 'o', 'o'},
			Print: []byte{0x13, 0x03, 'f', 'o', 'o'},
		},
		{
			Str:   "bar",
			UTF:   []byte{0x0c, 0x03, 'b', 'a', 'r'},
			IA5:   []byte{0x16, 0x03, 'b', 'a', 'r'},
			Print: []byte{0x13, 0x03, 'b', 'a', 'r'},
		},
	}
	for _, d := range data {
		got, err := EncodeString(d.Str)
		if err != nil {
			t.Errorf("%s(utf8): encoding failed! %s", d.Str, err)
			continue
		}
		compareBytes(t, d.Str, d.UTF, got)

		got, err = EncodeIA5String(d.Str)
		if err != nil {
			t.Errorf("%s(ia5): encoding failed! %s", d.Str, err)
			continue
		}
		compareBytes(t, d.Str, d.IA5, got)

		got, err = EncodePrintableString(d.Str)
		if err != nil {
			t.Errorf("%s(print): encoding failed! %s", d.Str, err)
			continue
		}
		compareBytes(t, d.Str, d.Print, got)
	}
}

func testEncodeNull(t *testing.T) {
	want := []byte{0x05, 0x00}
	compareBytes(t, nil, want, EncodeNull())
}

func testEncodeInt(t *testing.T) {
	data := []struct {
		Int  int64
		Want []byte
	}{
		{Int: 0, Want: []byte{0x02, 0x01, 0x00}},
		{Int: 127, Want: []byte{0x02, 0x01, 0x7F}},
		{Int: 128, Want: []byte{0x02, 0x02, 0x00, 0x80}},
		{Int: 256, Want: []byte{0x02, 0x02, 0x01, 0x00}},
		{Int: -128, Want: []byte{0x02, 0x01, 0x80}},
		{Int: -129, Want: []byte{0x02, 0x02, 0xFF, 0x7F}},
		{Int: 56, Want: []byte{0x02, 0x01, 0x38}},
		{Int: -56, Want: []byte{0x02, 0x01, 0xc8}},
		{Int: 512456, Want: []byte{0x02, 0x03, 0x07, 0xd1, 0xc8}},
		{Int: -512456, Want: []byte{0x02, 0x03, 0xf8, 0x2e, 0x38}},
	}
	for _, d := range data {
		got, err := EncodeInt(d.Int)
		if err != nil {
			t.Errorf("%d: encoding failed! %s", d.Int, err)
			continue
		}
		compareBytes(t, d.Int, d.Want, got)
	}
}

func testEncodeBytes(t *testing.T) {
	data := []struct {
		Bytes []byte
		Want  []byte
	}{
		{
			Bytes: []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Want:  []byte{04, 0x08, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		},
	}
	for _, d := range data {
		compareBytes(t, hex.EncodeToString(d.Bytes), d.Want, EncodeBytes(d.Bytes))
	}
}

func testEncodeBool(t *testing.T) {
	data := []struct {
		Bool bool
		Want []byte
	}{
		{Bool: true, Want: []byte{0x01, 0x01, 0xFF}},
		{Bool: false, Want: []byte{0x01, 0x01, 0x00}},
	}
	for _, d := range data {
		compareBytes(t, d.Bool, d.Want, EncodeBool(d.Bool))
	}
}

func TestEncodeLength(t *testing.T) {
	data := []struct {
		Length int
		Want   []byte
	}{
		{Length: 34, Want: []byte{0x22}},
		{Length: 201, Want: []byte{0x81, 0xc9}},
		{Length: 2201, Want: []byte{0x82, 0x08, 0x99}},
	}
	for _, d := range data {
		got, err := encodeLength(d.Length)
		if err != nil {
			t.Errorf("%d: encoding failed! %s", d.Length, err)
			continue
		}
		compareBytes(t, d.Length, d.Want, got)
	}
}

func TestEncodeIdentifier(t *testing.T) {
	data := []struct {
		Klass uint8
		Type  uint8
		Tag   uint32
		Want  []byte
	}{
		{
			Klass: Universal,
			Type:  Primitive,
			Tag:   Bool,
			Want:  []byte{0x01},
		},
		{
			Klass: Universal,
			Type:  Primitive,
			Tag:   261,
			Want:  []byte{0x1f, 0x82, 0x05},
		},
	}
	for _, d := range data {
		got, err := encodeIdentifier(d.Klass, d.Type, d.Tag)
		if err != nil {
			t.Errorf("encoding failed! %s", err)
			continue
		}
		compareBytes(t, d.Tag, d.Want, got)
	}
}

func compareBytes(t *testing.T, input interface{}, want, got []byte) {
	t.Helper()
	if bytes.Equal(want, got) {
		return
	}
	t.Errorf("%v: encoding failed!", input)
	t.Logf("\twant: %x", want)
	t.Logf("\tgot:  %x", got)
}
