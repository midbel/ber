package ber

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrPrimitive   = errors.New("encoding shall be primitive")
	ErrConstructed = errors.New("encoding shall be constructed")
)

type Encoder struct {
	err error
	buf []byte
}

func (e *Encoder) AsSequence() ([]byte, error) {
	return e.encodeConstructed(Sequence)
}

func (e *Encoder) AsSet() ([]byte, error) {
	return e.encodeConstructed(Set)
}

func (e *Encoder) As(id Ident) ([]byte, error) {
	return e.encodeConstructed(id)
}

func (e *Encoder) Bytes() []byte {
	buf := make([]byte, len(e.buf))
	copy(buf, e.buf)
	return buf
}

func (e *Encoder) Encode(val interface{}) error {
	return e.EncodeWithIdent(val, 0)
}

func (e *Encoder) EncodeWithIdent(val interface{}, tag Ident) error {
	if e.err != nil {
		return e.err
	}
	if m, ok := val.(Marshaler); ok {
		buf, err := m.Marshal()
		if err != nil {
			e.err = err
			return e.err
		}
		e.buf = append(e.buf, buf...)
		return e.err
	}
	if val == nil {
		return e.EncodeNullWithIdent(tag)
	}
	switch val := val.(type) {
	default:
		e.err = e.encodeValue(reflect.ValueOf(val), tag)
	case []byte:
		e.err = e.EncodeBytesWithIdent(val, tag)
	case string:
		if g := tag.Tag(); g == ObjectId.Tag() || g == RelObjectId.Tag() {
			e.err = e.EncodeOIDWithIdent(val, tag)
		} else {
			e.err = e.EncodeStringWithIdent(val, tag)
		}
	case bool:
		e.err = e.EncodeBoolWithIdent(val, tag)
	case float32:
		e.err = e.EncodeFloat2WithIdent(float64(val), tag)
	case float64:
		e.err = e.EncodeFloat2WithIdent(val, tag)
	case int8:
		e.err = e.EncodeIntWithIdent(int64(val), tag)
	case int16:
		e.err = e.EncodeIntWithIdent(int64(val), tag)
	case int32:
		e.err = e.EncodeIntWithIdent(int64(val), tag)
	case int64:
		e.err = e.EncodeIntWithIdent(val, tag)
	case int:
		e.err = e.EncodeIntWithIdent(int64(val), tag)
	case uint8:
		e.err = e.EncodeUintWithIdent(uint64(val), tag)
	case uint16:
		e.err = e.EncodeUintWithIdent(uint64(val), tag)
	case uint32:
		e.err = e.EncodeUintWithIdent(uint64(val), tag)
	case uint64:
		e.err = e.EncodeUintWithIdent(val, tag)
	case uint:
		e.err = e.EncodeUintWithIdent(uint64(val), tag)
	case time.Time:
		e.err = e.EncodeTimeWithIdent(val, tag)
	}
	return e.err
}

func (e *Encoder) EncodeChild(fn func(*Encoder) error) error {
	return e.EncodeChildWithIdent(0, fn)
}

func (e *Encoder) EncodeChildWithIdent(id Ident, fn func(*Encoder) error) error {
	var ex Encoder
	if err := fn(&ex); err != nil {
		return err
	}
	return e.merge(&ex, id)
}

func (e *Encoder) EncodeNull() error {
	return e.EncodeNullWithIdent(Null)
}

func (e *Encoder) EncodeNullWithIdent(tag Ident) error {
	if tag.isZero() {
		tag = Null
	}
	if tag.Type() != Primitive {
		return fmt.Errorf("null: %w", ErrPrimitive)
	}
	return e.encodeBytes(nil, tag)
}

func (e *Encoder) EncodeBool(val bool) error {
	return e.EncodeBoolWithIdent(val, Bool)
}

func (e *Encoder) EncodeBoolWithIdent(val bool, tag Ident) error {
	if tag.isZero() {
		tag = Bool
	}
	if tag.Type() != Primitive {
		return fmt.Errorf("bool: %w", ErrPrimitive)
	}
	var b byte
	if val {
		b = 0xFF
	}
	return e.encodeBytes([]byte{b}, tag)
}

func (e *Encoder) EncodeEnumerated(val int64) error {
	return e.EncodeEnumeratedWithIdent(val, Enumerated)
}

func (e *Encoder) EncodeEnumeratedWithIdent(val int64, tag Ident) error {
	if tag.isZero() {
		tag = Enumerated
	}
	if tag.Type() != Primitive {
		return fmt.Errorf("enumerated: %w", ErrPrimitive)
	}
	b := encodeInt(val)
	return e.encodeBytes(b, tag)
}

func (e *Encoder) EncodeInt(val int64) error {
	return e.EncodeIntWithIdent(val, Int)
}

func (e *Encoder) EncodeIntWithIdent(val int64, tag Ident) error {
	if tag.isZero() {
		tag = Int
	}
	if tag.Type() != Primitive {
		return fmt.Errorf("int: %w", ErrPrimitive)
	}
	b := encodeInt(val)
	return e.encodeBytes(b, tag)
}

func (e *Encoder) EncodeUint(val uint64) error {
	return e.EncodeUintWithIdent(val, Int)
}

func (e *Encoder) EncodeUintWithIdent(val uint64, tag Ident) error {
	if tag.isZero() {
		tag = Int
	}
	if tag.Type() != Primitive {
		return fmt.Errorf("uint: %w", ErrPrimitive)
	}
	b := encodeUint(val, false)
	return e.encodeBytes(b, tag)
}

func (e *Encoder) EncodeFloat2(val float64) error {
	return e.EncodeFloat2WithIdent(val, Real)
}

func (e *Encoder) EncodeFloat2WithIdent(val float64, tag Ident) error {
	return e.encodeFloat(val, 2, tag)
}

func (e *Encoder) EncodeFloat10(val float64) error {
	return e.EncodeFloat10WithIdent(val, Real)
}

func (e *Encoder) EncodeFloat10WithIdent(val float64, tag Ident) error {
	return e.encodeFloat(val, 10, tag)
}

func (e *Encoder) EncodeStringUTF8(val string) error {
	return e.EncodeStringWithIdent(val, UTF8String)
}

func (e *Encoder) EncodeStringPrintable(val string) error {
	return e.EncodeStringWithIdent(val, PrintableString)
}

func (e *Encoder) EncodeStringIA5(val string) error {
	return e.EncodeStringWithIdent(val, IA5String)
}

func (e *Encoder) EncodeStringWithIdent(val string, tag Ident) error {
	if tag.isZero() {
		tag = OctetString
	}
	if tag.Tag() == UTF8String.Tag() && !utf8.ValidString(val) {
		return fmt.Errorf("%s: invalid utf8 string", val)
	} else if tag.Tag() == PrintableString.Tag() && !ValidPrintableString(val) {
		return fmt.Errorf("%s: invalid printable string", val)
	} else if tag.Tag() == IA5String.Tag() && !ValidIA5String(val) {
		return fmt.Errorf("%s: invalid IA5 string", val)
	} else {

	}
	return e.encodeBytes([]byte(val), tag)
}

func (e *Encoder) EncodeBytes(val []byte) error {
	return e.EncodeBytesWithIdent(val, OctetString)
}

func (e *Encoder) EncodeBytesWithIdent(val []byte, tag Ident) error {
	if tag.isZero() {
		tag = OctetString
	}
	return e.encodeBytes(val, tag)
}

func (e *Encoder) EncodeUniversalTime(val time.Time) error {
	return e.EncodeTimeWithIdent(val, UniversalTime)
}

func (e *Encoder) EncodeGeneralizedTime(val time.Time) error {
	return e.EncodeTimeWithIdent(val, GeneralizedTime)
}

func (e *Encoder) EncodeTimeWithIdent(val time.Time, tag Ident) error {
	var pattern string
	if tag.isZero() {
		tag = GeneralizedTime
	}
	switch tag.Tag() {
	case Int.Tag():
		return e.EncodeInt(val.Unix())
	case UniversalTime.Tag():
		if !validTimeUTC(val) {
			return fmt.Errorf("%s: date outside utc range", val)
		}
		pattern = patUniversTime
	case GeneralizedTime.Tag():
		if !validTimeGeneralized(val) {
			return fmt.Errorf("%s: date outside generalized range", val)
		}
		pattern = patGeneralTime
	default:
		return fmt.Errorf("invalid tag for time encoding")
	}
	str := val.Format(pattern)
	return e.encodeBytes([]byte(str), tag)
}

func (e *Encoder) EncodeOID(val string, tag Ident) error {
	return e.EncodeOIDWithIdent(val, ObjectId)
}

func (e *Encoder) EncodeROID(val string) error {
	return e.EncodeOIDWithIdent(val, RelObjectId)
}

func (e *Encoder) EncodeOIDWithIdent(val string, tag Ident) error {
	if tag.isZero() {
		tag = ObjectId
	}
	if tag.Type() != Primitive {
		return fmt.Errorf("oid: %w", ErrPrimitive)
	}
	var minlen int
	if tag.Tag() == ObjectId.Tag() {
		minlen = 2
	} else {
		val = strings.TrimLeft(val, ".")
	}
	return e.encodeOID(val, minlen, tag)
}

func (e *Encoder) encodeOID(str string, min int, tag Ident) error {
	ids, err := splitOID(str, min)
	if err != nil {
		return err
	}
	pdu := make([]byte, 0, len(ids))
	var begin int
	if tag == ObjectId {
		pdu = append(pdu, encode128((ids[0]*40)+ids[1])...)
		begin = 2
	}
	for i := begin; i < len(ids); i++ {
		pdu = append(pdu, encode128(ids[i])...)
	}
	return e.encodeBytes(pdu, tag)
}

func (e *Encoder) encodeFloat(f float64, base int, tag Ident) error {
	if tag.isZero() {
		tag = Real
	}
	if tag.Type() != Primitive {
		return fmt.Errorf("real: %w", ErrPrimitive)
	}
	if b := encodeSpecialFloat(f); len(b) > 0 {
		if b[0] == 0 {
			b = b[:0]
		}
		return e.encodeBytes(b, tag)
	}
	var b []byte
	switch base {
	case 10:
		b, e.err = encodeDecimalFloat(f)
	case 2:
		b, e.err = encodeBinaryFloat(f, base)
	default:
		return fmt.Errorf("unsupported base %d", base)
	}
	if e.err == nil {
		e.err = e.encodeBytes(b, tag)
	}
	return e.err
}

func (e *Encoder) encodeBytes(b []byte, i Ident) error {
	if e.err != nil {
		return e.err
	}
	var (
		id, errk = encodeIdentifier(i.Class(), i.Type(), i.Tag())
		sz, errz = encodeLength(len(b))
	)
	switch {
	case errk != nil:
		e.err = errk
	case errz != nil:
		e.err = errz
	default:
		e.buf = append(e.buf, id...)
		e.buf = append(e.buf, sz...)
		if len(b) > 0 {
			e.buf = append(e.buf, b...)
		}
	}
	return e.err
}

func (e *Encoder) encodeConstructed(i Ident) ([]byte, error) {
	// if i.Type() != Constructed {
	// 	return nil, fmt.Errorf("constructed bit not set")
	// }
	if e.err != nil {
		return nil, e.err
	}
	var (
		id, errk = encodeIdentifier(i.Class(), i.Type(), i.Tag())
		sz, errz = encodeLength(len(e.buf))
	)
	switch {
	case errk != nil:
		return nil, errk
	case errz != nil:
		return nil, errz
	default:
		buf := make([]byte, 0, len(id)+len(sz)+len(e.buf))
		buf = append(buf, id...)
		buf = append(buf, sz...)
		return append(buf, e.buf...), nil
	}
}

func (e *Encoder) merge(other *Encoder, tag Ident) error {
	var (
		buf []byte
		err error
	)
	if tag.isZero() {
		buf, err = other.AsSequence()
	} else {
		buf, err = other.encodeConstructed(tag)
	}
	if err != nil {
		e.err = err
		return e.err
	}
	e.buf = append(e.buf, buf...)
	return e.err
}

var (
	timetype  = reflect.TypeOf(time.Now())
	bytestype = reflect.TypeOf([]byte{})
)

func (e *Encoder) encodeValue(val reflect.Value, tag Ident) error {
	switch val.Kind() {
	case reflect.Struct:
		if val.Type() == timetype {
			e.err = e.EncodeWithIdent(val.Interface(), tag)
			break
		}
		e.err = e.encodeStruct(val, tag)
	case reflect.Slice, reflect.Array:
		if val.Type() == bytestype {
			e.err = e.EncodeBytesWithIdent(val.Bytes(), tag)
			break
		}
		e.err = e.encodeArray(val, tag)
	case reflect.Map:
		e.err = e.encodeMap(val, tag)
	case reflect.Ptr:
		if val.IsNil() {
			e.err = e.EncodeNullWithIdent(tag)
			break
		}
		e.err = e.encodeValue(val.Elem(), tag)
	case reflect.Interface:
		if !val.CanInterface() {
			break
		}
		e.err = e.EncodeWithIdent(val.Interface(), tag)
	case reflect.String:
		if g := tag.Tag(); g == ObjectId.Tag() || g == RelObjectId.Tag() {
			e.err = e.EncodeOIDWithIdent(val.String(), tag)
		} else {
			e.err = e.EncodeStringWithIdent(val.String(), tag)
		}
	case reflect.Bool:
		e.err = e.EncodeBoolWithIdent(val.Bool(), tag)
	case reflect.Float32, reflect.Float64:
		e.err = e.EncodeFloat2WithIdent(val.Float(), tag)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.err = e.EncodeIntWithIdent(val.Int(), tag)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.err = e.EncodeUintWithIdent(val.Uint(), tag)
	default:
	}
	return e.err
}

var identForKind = map[reflect.Kind]Ident{
	reflect.Uint:    Int,
	reflect.Uint8:   Int,
	reflect.Uint16:  Int,
	reflect.Uint32:  Int,
	reflect.Uint64:  Int,
	reflect.Int:     Int,
	reflect.Int8:    Int,
	reflect.Int16:   Int,
	reflect.Int32:   Int,
	reflect.Int64:   Int,
	reflect.String:  UTF8String,
	reflect.Bool:    Bool,
	reflect.Float32: Real,
	reflect.Float64: Real,
	// reflect.Array:   Sequence,
	// reflect.Slice:   Sequence,
	// reflect.Map:     Sequence,
}

func (e *Encoder) encodeStruct(val reflect.Value, tag Ident) error {
	if tag.isZero() {
		tag = Sequence
	}
	// if tag.Type() != Constructed {
	// 	return fmt.Errorf("struct: %w", ErrConstructed)
	// }
	var (
		ex Encoder
		tp = val.Type()
	)
	for i := 0; i < val.NumField(); i++ {
		var (
			f    = val.Field(i)
			sf   = tp.Field(i)
			id   = identForKind[f.Kind()]
			omit bool
			err  error
		)
		if sf.PkgPath != "" {
			continue
		}
		if val.Type() == identtype {
			if sf.Name == "Id" || sf.Tag.Get("ber") == "id" {
				tag = val.Interface().(Ident)
				continue
			}
		}
		if tag := sf.Tag.Get("ber"); tag != "" {
			if tag == "-" {
				continue
			}
			id, omit, err = parseTag(tag, id)
			if err != nil {
				return err
			}
		}
		switch k := f.Kind(); {
		default:
			omit = false
		case omit && (k == reflect.Ptr || k == reflect.Interface):
			omit = f.IsNil()
		case omit && k == reflect.Struct:
			omit = f.NumField() == 0
		case omit && (k == reflect.Slice || k == reflect.Array || k == reflect.String || k == reflect.Map):
			omit = f.Len() == 0
		}
		if omit {
			continue
		}
		if err = ex.encodeValue(f, id); err != nil {
			e.err = err
			return e.err
		}
	}
	return e.merge(&ex, tag)
}

func (e *Encoder) encodeArray(val reflect.Value, tag Ident) error {
	if tag.isZero() {
		tag = Sequence
	}
	// if tag.Type() != Constructed {
	// 	return fmt.Errorf("array: %w", ErrConstructed)
	// }
	var ex Encoder
	for i := 0; i < val.Len(); i++ {
		if err := ex.encodeValue(val.Index(i), 0); err != nil {
			e.err = err
			return err
		}
	}
	return e.merge(&ex, tag)
}

func (e *Encoder) encodeMap(val reflect.Value, tag Ident) error {
	if tag.isZero() {
		tag = Sequence
	}
	// if tag.Type() != Constructed {
	// 	return fmt.Errorf("map: %w", ErrConstructed)
	// }
	var ex Encoder
	for _, k := range val.MapKeys() {
		if err := ex.encodeValue(k, 0); err != nil {
			e.err = err
			return err
		}
		if err := ex.encodeValue(val.MapIndex(k), 0); err != nil {
			e.err = err
			return err
		}
	}
	return e.merge(&ex, tag)
}

func encodeIdentifier(klass, kind uint8, tag uint32) ([]byte, error) {
	if klass > Private {
		return nil, fmt.Errorf("invalid class(%02x) given", klass)
	}
	if kind > Constructed {
		return nil, fmt.Errorf("invalid type(%02x) given", kind)
	}
	b := klass<<6 | kind<<5
	if tag < 31 {
		b |= uint8(tag & 0xFF)
		return []byte{byte(b)}, nil
	}
	id := encode128(tag)
	return append([]byte{byte(b | 0x1f)}, id...), nil
}

func encodeLength(e int) ([]byte, error) {
	if e < 0 {
		return nil, fmt.Errorf("length: negative length")
	}
	if e <= 127 {
		b := byte(e)
		return []byte{b}, nil
	}
	var (
		c = encode256(uint64(e))
		n = 0x80 | len(c)
	)
	if n >= 0xFF {
		return nil, fmt.Errorf("length: too long (%d)", len(c))
	}
	vs := append([]byte{}, byte(n))
	return append(vs, c...), nil
}

func encode256(i uint64) []byte {
	var (
		b = make([]byte, 0, 8)
		z = bits.Len64(i)
	)
	if mod := z % 8; mod > 0 {
		z += 8 - mod
	}
	for z -= 8; z >= 0; z -= 8 {
		y := (i >> z) & 0xFF
		b = append(b, byte(y))
	}
	return b
}

func encode128(i uint32) []byte {
	var (
		b = make([]byte, 0, 32)
		z = bits.Len32(i)
	)
	if mod := z % 7; mod > 0 {
		z -= mod
	} else {
		z -= 7
	}
	for z >= 0 {
		y := (i >> z) & 0x7F
		if z -= 7; z >= 0 {
			y |= 0x80
		}
		b = append(b, byte(y))
	}
	return b
}

func encodeBinaryFloat(f float64, base int) ([]byte, error) {
	if base != 2 {
		return nil, fmt.Errorf("%d: unsupported base", base)
	}

	frexp := func() (int64, int64) {
		const (
			bias  = (1 << 10) - 1
			shift = 64 - 11 - 1
			fmask = (1 << shift) - 1 // fraction mask
			emask = 0x7ff            // exponent mask
		)
		var (
			i = int64(math.Float64bits(f))
			e = (i >> shift) & emask
			m = (i & fmask) | (1 << shift)
			z = bits.TrailingZeros64(uint64(m))
		)
		return e - bias - (shift - int64(z)), m >> z
	}
	var (
		e, m      = frexp()
		es        = encodeInt(e)
		ms        = encodeInt(m)
		info byte = 0x80
	)

	if f < 0 {
		info |= 0x40
	}
	switch len(es) {
	case 1:
	case 2:
		info |= 0x01
	case 3:
		info |= 0x02
	default:
		if len(es) > 0xFF {
			return nil, fmt.Errorf("exponent too longth: %d", len(es))
		}
		info |= 0x03
	}
	b := make([]byte, 0, 64)
	b = append(b, info)
	if n := len(es); n > 3 {
		b = append(b, byte(n))
	}
	b = append(b, es...)
	b = append(b, ms...)
	return b, nil
}

func encodeDecimalFloat(f float64) ([]byte, error) {
	var (
		b = make([]byte, 0, 64)
		i byte
	)
	b = strconv.AppendFloat(b, f, 'G', -1, 64)
	switch {
	case bytes.IndexByte(b, 'E') > 0:
		i = 0x03
	case bytes.IndexByte(b, '.') > 0:
		i = 0x02
	default:
		i = 0x01
	}
	b = append(b[:1], b...)
	b[0] = i
	return b, nil
}

func encodeSpecialFloat(f float64) []byte {
	b := byte(0x40)
	switch {
	case f == 0 && !math.Signbit(f):
		b = 0x00
	case f == 0 && math.Signbit(f):
		b |= 0x03
	case math.IsInf(f, 1):
	case math.IsInf(f, -1):
		b |= 0x01
	case math.IsNaN(f):
		b |= 0x02
	default:
		return nil
	}
	return []byte{b}
}

func encodeUint(i uint64, less bool) []byte {
	c := encode256(i)
	if len(c) < 1 {
		return []byte{0x00}
	}
	if less && c[0]>>7 == 0 {
		c = append(c[:1], c[0:]...)
		c[0] = 0xFF
	} else if !less && c[0]>>7 == 1 {
		c = append(c[:1], c[0:]...)
		c[0] = 0x00
	}
	return c
}

func encodeInt(i int64) []byte {
	var x uint64
	switch {
	case i > 0:
		x = uint64(i)
	case i < 0:
		var (
			zero = bits.LeadingZeros64(uint64(^i))
			mask = uint64(1<<zero - 1)
		)
		if mod := zero % 8; mod > 0 {
			zero -= mod
		}
		x = uint64(i) ^ (mask << (64 - zero))
	default:
		return []byte{0x00}
	}
	return encodeUint(x, i < 0)
}

func splitOID(str string, min int) ([]uint32, error) {
	var (
		parts = strings.Split(str, ".")
		ids   = make([]uint32, 0, len(parts))
	)
	if min > 0 && len(parts) < min {
		return nil, fmt.Errorf("%s: short OID", str)
	}
	for i := range parts {
		x, err := strconv.ParseUint(parts[i], 10, 32)
		if err != nil {
			return nil, err
		}
		ids = append(ids, uint32(x))
	}
	return ids, nil
}
