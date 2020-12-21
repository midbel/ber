package ber

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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
	if u, ok := val.(Unmarshaler); ok {
		return u.Unmarshal(d)
	}
	var err error
	switch val := val.(type) {
	case *string:
		*val, err = d.DecodeString()
	case *bool:
		*val, err = d.DecodeBool()
	case *int:
	case *int8:
	case *int16:
	case *int32:
	case *int64:
	case *uint:
	case *uint8:
	case *uint16:
	case *uint32:
	case *uint64:
	case *time.Time:
	default:
	}
	return err
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

func (d *Decoder) DecodeEnumerated() (int64, error) {
	return d.DecodeInt()
}

func (d *Decoder) DecodeInt() (int64, error) {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return 0, err
	}
	if id.Type() != Primitive {
		return 0, fmt.Errorf("int: %w", ErrPrimitive)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return 0, err
	}
	d.offset += size + n

	return decodeInt(d.buf[d.offset-size:d.offset], true), nil
}

func (d *Decoder) DecodeUint() (uint64, error) {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return 0, err
	}
	if id.Type() != Primitive {
		return 0, fmt.Errorf("uint: %w", ErrPrimitive)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return 0, err
	}
	d.offset += size + n

	j := decodeInt(d.buf[d.offset-size:d.offset], false)
	return uint64(j), nil
}

func (d *Decoder) DecodeFloat() (float64, error) {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return 0, err
	}
	if id.Type() != Primitive {
		return 0, fmt.Errorf("float: %w", ErrPrimitive)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return 0, err
	}
	if size == 0 {
		return 0, nil
	}
	d.offset += size + n
	var (
		str  = d.buf[d.offset-size+1 : d.offset]
		info = d.buf[d.offset-size]
	)
	switch {
	case info>>6 == 0: // decimal encoding
		return decodeDecimalFloat(str)
	case info>>6 == 1: // special float
		return decodeSpecialFloat(info)
	case info>>7 == 1: // binary encoding
		return decodeBinaryFloat(info, str)
	default:
		return 0, fmt.Errorf("invalid float encoding: %x (%x)", info, str)
	}
	return 0, nil
}

func (d *Decoder) DecodeString() (string, error) {
	_, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return "", err
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return "", err
	}
	d.offset += size + n
	str := d.buf[d.offset-size : d.offset]
	return string(str), nil
}

func (d *Decoder) DecodeOID() (string, error) {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return "", err
	}
	if id.Type() != Primitive {
		return "", fmt.Errorf("oid: %w", ErrPrimitive)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return "", err
	}
	d.offset += size + n
	var (
		str = d.buf[d.offset-size : d.offset]
		ids []string
		pos int
	)
	for pos < len(str) {
		i, n := decode128(str[pos:])
		if pos == 0 && id.Tag() == ObjectId.Tag() {
			div, mod := i/40, i%40
			ids = append(ids, strconv.Itoa(int(div)))
			ids = append(ids, strconv.Itoa(int(mod)))
		} else {
			ids = append(ids, strconv.Itoa(int(i)))
		}
		pos += n
	}
	oid := strings.Join(ids, ".")
	if id.Tag() == RelObjectId.Tag() {
		oid = "." + oid
	}
	return oid, nil
}

func (d *Decoder) DecodeTime() (time.Time, error) {
	var t time.Time
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return t, err
	}
	if id.Type() != Primitive {
		return t, fmt.Errorf("time: %w", ErrPrimitive)
	}
	var pattern string
	switch id.Tag() {
	case UniversalTime.Tag():
		pattern = patUniversTime
	case GeneralizedTime.Tag():
		pattern = patGeneralTime
	default:
		return t, fmt.Errorf("unsupported tag for time")
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return t, err
	}
	d.offset += size + n
	str := d.buf[d.offset-size : d.offset]
	t, err = time.Parse(pattern, string(str))
	if err == nil {
		t = t.UTC()
	}
	return t, err
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

func decodeInt(b []byte, sign bool) int64 {
	var j int64
	for _, i := range b {
		j <<= 8
		j |= int64(i)
	}
	if sign {
		size := 64 - (len(b) * 8)
		j <<= size
		j >>= size
	}
	return j
}

func decodeSpecialFloat(info byte) (float64, error) {
	var real float64
	if r := info & 0x3F; r == 0 {
		real = math.Inf(1)
	} else if r == 1 {
		real = math.Inf(-1)
	} else if r == 2 {
		real = math.NaN()
	} else if r == 3 {
		real = math.Copysign(0, -1)
	} else {
		return 0, fmt.Errorf("invalid special float")
	}
	return real, nil
}

func decodeBinaryFloat(info byte, str []byte) (float64, error) {
	sign := 1.0
	if info&0x40 == 0x40 {
		sign = -sign
	}
	if base := info & 0x30; base != 0 {
		return 0, fmt.Errorf("unsupported base")
	}
	var size int
	if e := info & 0x03; e == 0 {
		size = 1
	} else if e == 1 {
		size = 2
	} else if e == 2 {
		size = 3
	} else {
		size = 4
	}
	m := float64(decodeInt(str[size:], true))
	e := float64(decodeInt(str[:size], true))
	return sign * m * math.Pow(2, e), nil
}

func decodeDecimalFloat(str []byte) (float64, error) {
	return strconv.ParseFloat(string(str), 64)
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
