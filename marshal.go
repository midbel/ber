package ber

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

func marshal(b writer, v reflect.Value, t string) error {
	if v, ok := v.Interface().(Marshaler); ok {
		bs, err := v.MarshalBER()
		if err != nil {
			return err
		}
		if len(bs) > 0 {
			b.Write(bs)
		}
		return nil
	}
	var (
		vs []byte
		id byte
	)
	switch k := v.Kind(); k {
	default:
		return fmt.Errorf("ber: unsupported data type %s", k)
	case reflect.Interface:
		return marshal(b, reflect.ValueOf(v.Interface()), t)
	case reflect.Bool:
		id = ClassUnivers<<6 | TypeSimple<<5 | TagBoolean
		vs = []byte{0x00}
		if v.Bool() {
			vs = []byte{0xFF}
		}
	case reflect.String:
		id = ClassUnivers<<6 | TypeSimple<<5 | TagString
		vs = []byte(v.String())
	case reflect.Float32:
		id = ClassUnivers<<6 | TypeSimple<<5 | TagReal

		u := math.Float32bits(float32(v.Float()))
		vs = make([]byte, binary.MaxVarintLen64)
		if n := binary.PutUvarint(vs, uint64(u)); n > 0 {
			vs = vs[:n]
		}
	case reflect.Float64:
		id = ClassUnivers<<6 | TypeSimple<<5 | TagReal
		vs = make([]byte, binary.MaxVarintLen64)
		if n := binary.PutUvarint(vs, math.Float64bits(v.Float())); n > 0 {
			vs = vs[:n]
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		id = ClassUnivers<<6 | TypeSimple<<5 | TagInteger
		vs = make([]byte, binary.MaxVarintLen64)
		if n := binary.PutVarint(vs, v.Int()); n > 0 {
			vs = vs[:n]
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		id = ClassUnivers<<6 | TypeSimple<<5 | TagInteger
		vs = make([]byte, binary.MaxVarintLen64)
		if n := binary.PutUvarint(vs, v.Uint()); n > 0 {
			vs = vs[:n]
		}
	case reflect.Slice, reflect.Array:
		id = ClassUnivers<<6 | TypeCompound<<5 | TagSequence

		buf := new(bytes.Buffer)
		for i := 0; i < v.Len(); i++ {
			if err := marshal(buf, v.Index(i), ""); err != nil {
				return err
			}
		}
		vs = buf.Bytes()
	case reflect.Map:
		id = ClassUnivers<<6 | TypeCompound<<5 | TagSequence

		buf := new(bytes.Buffer)
		for _, k := range v.MapKeys() {
			if err := marshal(buf, k, ""); err != nil {
				return err
			}
			if err := marshal(buf, v.MapIndex(k), ""); err != nil {
				return err
			}
		}
		vs = buf.Bytes()
	case reflect.Struct:
		id = ClassUnivers<<6 | TypeCompound<<5 | TagSequence

		buf := new(bytes.Buffer)
		for i, t := 0, v.Type(); i < v.NumField(); i++ {
			f := t.Field(i)
			if len(f.PkgPath) > 0 {
				continue
			}
			tag := f.Tag.Get("ber")
			if strings.Index(tag, "optional") >= 0 && isEmptyValue(v.Field(i)) {
				continue
			}
			if err := marshal(buf, v.Field(i), tag); err != nil {
				return err
			}
		}
		vs = buf.Bytes()
	}
	if i, ok := parseTag(t); ok {
		id = i
	}
	b.WriteByte(id)
	b.Write(sizeOf(len(vs)))
	if len(vs) > 0 {
		b.Write(vs)
	}

	return nil
}

func parseTag(t string) (byte, bool) {
	if t == "" || t == "-" {
		return 0x00, false
	}
	var id byte
	for _, s := range strings.Split(t, ",") {
		switch {
		case s == "simple":
			id = (id &^ (1 << 5)) | TypeSimple<<5
		case s == "application":
			id = (id &^ 3 << 6) | ClassApplication<<6 | TypeCompound<<5
		case s == "context":
			id = (id &^ 3 << 6) | ClassContext<<6
		case s == "set":
			id = ClassUnivers<<6 | TypeCompound<<5 | TagSet
		case s == "enum":
			id = ClassUnivers<<6 | TypeSimple | TagEnum
		case strings.HasPrefix(s, "tag"):
			ix := strings.Index(s, ":") + 1
			t, err := strconv.Atoi(s[ix:])
			if err != nil {
				return 0x00, false
			}
			id = (id &^ (1<<5 - 1)) | byte(t)
		}
	}
	return id, true
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func sizeOf(z int) []byte {
	if z <= 1<<7-1 {
		return []byte{byte(z)}
	}
	vs := make([]byte, 16)
	i := len(vs) - 1
	for z > 0 {
		vs[i], z = byte(z&0xff), z>>8
		i--
	}
	vs[i] = 1<<7 | byte(len(vs)-i-1)
	return vs[i:]
}
