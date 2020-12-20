package ber

import (
	"fmt"
	"time"
)

type Decoder struct {
	buf    []byte
	offset int
	err    error
}

func NewDecoder(buf []byte) *Decoder {
	return &Decoder{
		buf: append([]byte{}, buf...),
	}
}

func (d *Decoder) Append(buf []byte) {
	d.buf = append(d.buf, buf...)
}

func (d *Decoder) Empty() bool {
	return d.offset >= len(d.buf)
}

func (d *Decoder) Skip(n int) error {
	return nil
}

func (d *Decoder) Decode(val interface{}) error {
	return nil
}

func (d *Decoder) DecodeTagged() (Ident, int, error) {
	return 0, 0, nil
}

func (d *Decoder) DecodeNull() error {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return err
	}
	if id.Type() != Primitive {
		return fmt.Errorf("null: %w", ErrPrimitive)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return err
	}
	if size != 0 {
		return fmt.Errorf("null: value should have length 0 (got, %d)", size)
	}
	d.offset += n
	return nil
}

func (d *Decoder) DecodeBool() (bool, error) {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return false, err
	}
	if id.Type() != Primitive {
		return false, fmt.Errorf("bool: %w", ErrPrimitive)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return false, err
	}
	if size != 1 {
		return false, fmt.Errorf("bool: value should have length 1 (got, %d)", size)
	}
	d.offset += n
	d.offset += size
	return d.buf[d.offset-size] > 0x00, nil
}

func (d *Decoder) DecodeInt() (int64, error) {
	return 0, nil
}

func (d *Decoder) DecodeUint() (uint64, error) {
	return 0, nil
}

func (d *Decoder) DecodeFloat() (float64, error) {
	return 0, nil
}

func (d *Decoder) DecodeString() (string, error) {
	return "", nil
}

func (d *Decoder) DecodeOID() (string, error) {
	return "", nil
}

func (d *Decoder) DecodeTime() (time.Time, error) {
	return time.Now(), nil
}

func decodeIdentifier(b []byte) (Ident, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}
	n := 1
	class, kind, tag := uint64(b[0]>>6), uint64(b[0]>>5)&0x01, uint64(b[0]&0x1F)
	if tag == 0x1F {
		g, x := decode128(b[1:])
		tag, n = uint64(g), n+x
	}
	return Ident(class<<33 | kind<<32 | tag), n, nil
}

func decodeLength(b []byte) (int, int, error) {
	if len(b) == 0 {
		return 0, 0, fmt.Errorf("length should have at least 1 byte")
	}
	if b[0]>>7 == 0 {
		return int(b[0] & 0x7F), 1, nil
	}
	var (
		i int64
		n int
		c = int(b[0] & 0x7F)
	)
	n++
	for j := 0; j < c && j < len(b); j++ {
		i = (i << 8) | int64(b[j+1])
		n++
	}
	return int(i), n, nil
}

func decode128(b []byte) (uint32, int) {
	if len(b) == 0 {
		return 0, 0
	}
	var (
		i uint32
		n int
	)
	for j := 0; j < len(b); j++ {
		i = (i << 7) | uint32(b[j]&0x7F)
		n++
		if b[j]>>7 == 0 {
			break
		}
	}
	return i, n
}
