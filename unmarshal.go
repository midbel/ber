package ber

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

func unmarshal(r reader, v reflect.Value) error {
	if v.IsValid() && v.CanInterface() {
		if v, ok := v.Interface().(Unmarshaler); ok {
			t, _ := r.ReadByte()
			vs := make([]byte, readLength(r))
			if _, err := r.Read(vs); err != nil {
				return err
			}
			if err := v.UnmarshalBER(t, vs); err != nil && err != Skip {
				return err
			}
			return nil
		}
	}
	switch v.Kind() {
	case reflect.Interface:
		f := reflect.ValueOf(v.Interface()).Elem()
		return unmarshal(r, f)
	case reflect.Ptr:
		return unmarshal(r, v.Elem())
	}

	if _, err := r.ReadByte(); err != nil {
		return nil
	}
	z := readLength(r)
	if z == 0 {
		return nil
	}
	vs := make([]byte, z)
	if _, err := r.Read(vs); err != nil {
		return err
	}
	switch k := v.Kind(); k {
	default:
		return fmt.Errorf("unsupported data type: %s", k)
	case reflect.Bool:
		v.SetBool(vs[0] != 0x00)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if x, n := binary.Varint(vs); n < 0 {
			return fmt.Errorf("fail to decode %s", k)
		} else {
			v.SetInt(x)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if x, n := binary.Uvarint(vs); n < 0 {
			return fmt.Errorf("fail to decode %s", k)
		} else {
			v.SetUint(x)
		}
	case reflect.Float32:
		x, n := binary.Uvarint(vs)
		if n < 0 {
			return fmt.Errorf("fail to decode %s", k)
		}
		v.SetFloat(float64(math.Float32frombits(uint32(x))))
	case reflect.Float64:
		x, n := binary.Uvarint(vs)
		if n < 0 {
			return fmt.Errorf("fail to decode %s", k)
		}
		v.SetFloat(math.Float64frombits(x))
	case reflect.String:
		v.SetString(string(vs))
	case reflect.Slice, reflect.Array:
		if v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
		r := bytes.NewReader(vs)
		for r.Len() > 0 {
			f := reflect.New(v.Type().Elem()).Elem()
			if err := unmarshal(r, f); err != nil {
				return err
			}
			v.Set(reflect.Append(v, f))
		}
	case reflect.Map:
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
		r := bytes.NewReader(vs)
		for r.Len() > 0 {
			r.ReadByte()
			vs := make([]byte, readLength(r))
			if _, err := r.Read(vs); err != nil {
				return err
			}
			o := bytes.NewReader(vs)
			k := reflect.New(v.Type().Key()).Elem()
			if err := unmarshal(o, k); err != nil {
				return err
			}
			f := reflect.New(v.Type().Elem()).Elem()
			if err := unmarshal(o, f); err != nil {
				return err
			}
			v.SetMapIndex(k, f)
		}
	case reflect.Struct:
		r := bytes.NewReader(vs)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanSet() {
				continue
			}
			if t := t.Field(i).Tag; t.Get("ber") == "-" {
				continue
			}
			if err := unmarshal(r, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func readLength(r reader) int {
	c, err := r.ReadByte()
	if err != nil {
		return 0
	}
	if c>>7 == 0 {
		return int(c)
	}
	vs := make([]byte, int(c&0x7F))
	if _, err := r.Read(vs); err != nil {
		return 0
	}
	var i int
	for j, k := len(vs)-1, 0; j >= 0; j-- {
		i |= int(vs[j]) << uint(k*8)
		k++
	}
	return i
}
