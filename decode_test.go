package ber

import (
	"testing"
)

func TestDecoder(t *testing.T) {
	t.Run("null", testDecodeNull)
	t.Run("bool", testDecodeBool)
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
			t.Errorf("boo: fail to decode bool! %s", err)
			continue
		}
		if got != d.Want {
			t.Errorf("bool mismatched! want: %t, got %t", d.Want, got)
		}
	}
}
