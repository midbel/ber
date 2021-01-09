package ber

import (
	"fmt"
	"math"
	"reflect"
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

func (d *Decoder) Peek() (Ident, error) {
	id, _, err := decodeIdentifier(d.buf[d.offset:])
	return id, err
}

func (d *Decoder) Need() int {
	offset := d.offset
	_, n, err := decodeIdentifier(d.buf[offset:])
	if err != nil {
		return n
	}
	offset += n
	size, x, err := decodeLength(d.buf[offset:])
	if err != nil {
		return 0
	}
	return size + x + n
}

func (d *Decoder) Reset(buf []byte) {
	d.offset = 0
	d.err = nil
	d.buf = append(d.buf[:0], buf...)
}

func (d *Decoder) Bytes() []byte {
	if d.Empty() {
		return nil
	}
	return append([]byte{}, d.buf[d.offset:]...)
}

func (d *Decoder) Append(buf []byte) {
	if d.offset > 0 {
		d.buf = d.buf[d.offset:]
		d.offset = 0
	}
	d.buf = append(d.buf, buf...)
}

func (d *Decoder) Empty() bool {
	return d.offset >= len(d.buf)
}

func (d *Decoder) Len() int {
	return len(d.buf) - d.offset
}

func (d *Decoder) Size() int {
	return len(d.buf)
}

func (d *Decoder) Can() bool {
	if d.Empty() {
		return false
	}
	offset := d.offset
	_, n, err := decodeIdentifier(d.buf[offset:])
	if err != nil {
		return false
	}
	offset += n
	if offset >= len(d.buf) {
		return false
	}
	size, n, err := decodeLength(d.buf[offset:])
	if err != nil {
		return false
	}
	offset += n + size
	return offset <= len(d.buf)
}

func (d *Decoder) Skip() error {
	_, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return err
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err == nil {
		d.offset += n + size
	}
	return err
}

func (d *Decoder) Decode(value interface{}) error {
	if u, ok := value.(Unmarshaler); ok {
		return d.decodeUnmarshaler(u)
	}
	var err error
	switch val := value.(type) {
	case *string:
		*val, err = d.DecodeString()
	case *bool:
		*val, err = d.DecodeBool()
	case *int:
		x, e := d.DecodeInt()
		*val, err = int(x), e
	case *int8:
		x, e := d.DecodeInt()
		*val, err = int8(x), e
	case *int16:
		x, e := d.DecodeInt()
		*val, err = int16(x), e
	case *int32:
		x, e := d.DecodeInt()
		*val, err = int32(x), e
	case *int64:
		*val, err = d.DecodeInt()
	case *uint:
		x, e := d.DecodeUint()
		*val, err = uint(x), e
	case *uint8:
		x, e := d.DecodeUint()
		*val, err = uint8(x), e
	case *uint16:
		x, e := d.DecodeUint()
		*val, err = uint16(x), e
	case *uint32:
		x, e := d.DecodeUint()
		*val, err = uint32(x), e
	case *uint64:
		*val, err = d.DecodeUint()
	case *time.Time:
		*val, err = d.DecodeTime()
	default:
		err = d.decodeValue(reflect.ValueOf(value).Elem())
	}
	return err
}

func (d *Decoder) DecodeTagged() (Ident, int, error) {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return id, 0, err
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err == nil {
		d.offset += n
	}
	return id, size, nil
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
	case Int.Tag():
		i, err := d.DecodeInt()
		if err != nil {
			return t, err
		}
		return time.Unix(i, 0), nil
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

var unmarshaltype = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

func (d *Decoder) decodeUnmarshaler(u Unmarshaler) error {
	_, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return err
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return err
	}
	d.offset += n + size
	return u.Unmarshal(d.buf[d.offset-size:d.offset])
}

func (d *Decoder) decodeValue(val reflect.Value) error {
	if val.CanInterface() && val.Type().Implements(unmarshaltype) {
		return d.decodeUnmarshaler(val.Interface().(Unmarshaler))
	}
	if val.CanAddr() {
		pv := val.Addr()
		if pv.CanInterface() && pv.Type().Implements(unmarshaltype) {
			return d.decodeUnmarshaler(pv.Interface().(Unmarshaler))
		}
	}
	switch k := val.Kind(); k {
	case reflect.Struct:
		if timetype == val.Type() {
			t, err := d.DecodeTime()
			if err == nil {
				val.Set(reflect.ValueOf(t))
			}
			return err
		}
		return d.decodeStruct(val)
	case reflect.Array:
		return d.decodeArray(val)
	case reflect.Slice:
		return d.decodeSlice(val)
	case reflect.Map:
		return d.decodeMap(val)
	case reflect.Ptr:
		return d.decodeValue(val.Elem())
	case reflect.Interface:
	case reflect.String:
		v, err := d.DecodeString()
		if err != nil {
			return err
		}
		val.SetString(v)
	case reflect.Bool:
		v, err := d.DecodeBool()
		if err != nil {
			return err
		}
		val.SetBool(v)
	case reflect.Float32, reflect.Float64:
		v, err := d.DecodeFloat()
		if err != nil {
			return err
		}
		val.SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := d.DecodeInt()
		if err != nil {
			return err
		}
		val.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := d.DecodeUint()
		if err != nil {
			return err
		}
		val.SetUint(v)
	default:
		return fmt.Errorf("element can not be decoded into %s", k)
	}
	return nil
}

var identtype = reflect.TypeOf(Ident(0))

func (d *Decoder) decodeStruct(val reflect.Value) error {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return err
	}
	if id.Type() != Constructed {
		return fmt.Errorf("struct: %w", ErrConstructed)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return err
	}
	d.offset += n
	if size == 0 {
		return nil
	}
	limit := d.offset + size
	for i := 0; i < val.NumField() && d.offset < limit; i++ {
		f := val.Field(i)
		if !f.CanSet() {
			continue
		}
		if tf := val.Type().Field(i); f.Type() == identtype {
			if tf.Name == "Id" || tf.Tag.Get("ber") == "id" {
				f.Set(reflect.ValueOf(id))
				continue
			}
		}
		if err := d.decodeValue(f); err != nil {
			return err
		}
		if d.offset > limit {
			return fmt.Errorf("struct: too many bytes consumed to decode value")
		}
	}
	return nil
}

func (d *Decoder) decodeMap(val reflect.Value) error {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return err
	}
	if id.Type() != Constructed {
		return fmt.Errorf("map: %w", ErrConstructed)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return err
	}
	d.offset += n
	if size == 0 {
		return nil
	}
	var (
		limit = d.offset + size
		mp    = reflect.MakeMap(val.Type())
		typ   = mp.Type()
	)
	for d.offset < limit {
		// TODO: element of map should be decoded as sequence type
		k, v := reflect.New(typ.Key()).Elem(), reflect.New(typ.Elem()).Elem()
		if err := d.decodeValue(k); err != nil {
			return err
		}
		if err := d.decodeValue(v); err != nil {
			return err
		}
		if d.offset > limit {
			return fmt.Errorf("map: too many bytes consumed to decode value")
		}
		mp.SetMapIndex(k, v)
	}
	val.Set(mp)
	return nil
}

func (d *Decoder) decodeSlice(val reflect.Value) error {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return err
	}
	if id.Type() != Constructed {
		return fmt.Errorf("slice: %w", ErrConstructed)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return err
	}
	d.offset += n
	if size == 0 {
		return nil
	}
	var (
		limit = d.offset + size
		typ   = val.Type()
		slice = reflect.MakeSlice(typ, 0, val.Len())
	)
	for d.offset < limit {
		e := reflect.New(typ.Elem()).Elem()
		if err := d.decodeValue(e); err != nil {
			return err
		}
		if d.offset > limit {
			return fmt.Errorf("slice: too many bytes consumed to decode value")
		}
		slice = reflect.Append(slice, e)
	}
	if n := reflect.Copy(val, slice); n < slice.Len() {
		slice = reflect.AppendSlice(val, slice.Slice(n, slice.Len()))
		val.Set(slice)
	}
	return nil
}

func (d *Decoder) decodeArray(val reflect.Value) error {
	id, n, err := decodeIdentifier(d.buf[d.offset:])
	if err != nil {
		return err
	}
	if id.Type() != Constructed {
		return fmt.Errorf("array: %w", ErrConstructed)
	}
	d.offset += n
	size, n, err := decodeLength(d.buf[d.offset:])
	if err != nil {
		return err
	}
	d.offset += n
	if size == 0 {
		return nil
	}
	limit := d.offset + size
	for i := 0; i < val.Len(); i++ {
		if d.offset >= limit {
			break
		}
		if err := d.decodeValue(val.Index(i)); err != nil {
			return err
		}
	}
	if d.offset < limit {
		return fmt.Errorf("array: undecoded values remained! array too short")
	}
	return nil
}

func decodeIdentifier(b []byte) (Ident, int, error) {
	if len(b) == 0 {
		return 0, 0, fmt.Errorf("identifier should have at least 1 byte")
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
