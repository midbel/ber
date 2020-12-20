package ber

import (
	"testing"
	"time"
)

func TestDecoder(t *testing.T) {
	t.Run("null", testDecodeNull)
	t.Run("bool", testDecodeBool)
	t.Run("int", testDecodeInt)
	t.Run("string", testDecodeString)
	t.Run("time", testDecodeTime)
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
