package ber

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

type Bind struct {
	Version uint8
	Name    string
	Passwd  string `ber:"context,tag:0"`
}

type Message struct {
	Id uint32
	Op interface{}
}

type Result struct {
	Code    int8
	Node    string
	Message string
}

func TestDebug(t *testing.T) {
	const s = "302d02010160280201030417636e3d61646d696e2c64633d6275736f632c64633d6265800a68656c6c6f776f726c64"
	bs, err := hex.DecodeString(s)
	if err != nil {
		t.Error(err)
		return
	}
	Debug(bytes.NewReader(bs), ioutil.Discard)
}

func TestUnmarshalMessage(t *testing.T) {
	m := Message{Id: 1, Op: &Bind{3, "cn=admin,dc=busoc,dc=be", "helloworld"}}
	s := Message{Id: 1, Op: &Result{0, "", ""}}

	data := []struct {
		Value string
		Want  Message
		Op    interface{}
	}{
		{Value: "0x302d02010160280201030417636e3d61646d696e2c64633d6275736f632c64633d6265800a68656c6c6f776f726c64", Want: m, Op: &Bind{}},
		{Value: "0x300c02010161070a010004000400", Want: s, Op: &Result{}},
	}
	for i, test := range data {
		bs, err := hex.DecodeString(test.Value[2:])
		if err != nil {
			continue
		}
		v := &Message{Op: test.Op}
		if err := Unmarshal(bs, v); err != nil {
			t.Errorf("%d) unmarshal of %+v failed: %s", i+1, test.Want, err)
			continue
		}
		if !reflect.DeepEqual(*v, test.Want) {
			t.Errorf("%d) unmarshal failed: %#v != %#v", i+1, *v, test.Want)
		}
	}
}

func TestUnmarshalBool(t *testing.T) {
	data := []struct {
		Value string
		Want  bool
	}{
		{Value: "0x0101ff", Want: true},
		{Value: "0x010118", Want: true},
		{Value: "0x010100", Want: false},
	}
	for i, test := range data {
		bs, err := hex.DecodeString(test.Value[2:])
		if err != nil {
			continue
		}
		var v bool
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("%d) unmarshal of %s failed: %s", i+1, test.Want, err)
			continue
		}
		if v != test.Want {
			t.Errorf("%d) expected %t => got: %t", test.Want, v)
		}
	}
}

func TestUnmarshalNumber(t *testing.T) {
	data := []struct {
		Value string
		Want  int
	}{
		{Value: "0x020103", Want: int(3)},
	}
	for i, d := range data {
		bs, err := hex.DecodeString(d.Value[2:])
		if err != nil {
			continue
		}
		v := d.Want
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("%d) unmarshal of %v failed: %s", i+1, d.Want, err)
			continue
		}
		if fmt.Sprint(v) != fmt.Sprint(d.Want) {
			t.Errorf("%d) unmarshal failed: %v != %v", i+1, v, d.Want)
		}
	}
}

func TestMarshal(t *testing.T) {
	b := Bind{3, "cn=admin,dc=busoc,dc=be", "helloworld"}
	m1 := struct {
		Id uint32
		Op interface{} `ber:"application,tag:0"`
	}{1, b}
	m2 := struct {
		Id uint32
		Op interface{} `ber:"application,simple,tag:2"`
	}{3, struct{}{}}

	data := []struct {
		Value interface{}
		Want  string
	}{
		{Value: true, Want: "0101ff"},
		{Value: false, Want: "010100"},
		{Value: int(3), Want: "020103"},
		{Value: int8(3), Want: "020103"},
		{Value: uint8(3), Want: "020103"},
		{Value: "helloworld", Want: "040a68656c6c6f776f726c64"},
		{Value: b, Want: ""},
		{Value: m1, Want: "302d02010160280201030417636e3d61646d696e2c64633d6275736f632c64633d6265800a68656c6c6f776f726c64"},
		{Value: m2, Want: "30050201034200"},
	}

	for i, test := range data {
		bs, err := Marshal(test.Value)
		if err != nil {
			t.Errorf("%d) marshal failed for %v: %s", i+1, test.Value, err)
			continue
		}
		hs, _ := hex.DecodeString(test.Want)
		if !bytes.Equal(hs, bs) {
			t.Errorf("%d) bytes mismatched: got %x, want %x", i+1, bs, hs)
		}
	}
}
