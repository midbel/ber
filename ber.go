package ber

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
)

const (
	ClassUnivers byte = iota
	ClassApplication
	ClassContext
	ClassPrivate
)

const (
	TypeSimple byte = iota
	TypeCompound
)

const (
	TagBoolean  byte = 0x01
	TagInteger  byte = 0x02
	TagBytes    byte = 0x03
	TagString   byte = 0x04
	TagNull     byte = 0x05
	TagReal     byte = 0x09
	TagEnum     byte = 0x0a
	TagSequence byte = 0x10
	TagSet      byte = 0x11
	TagTime     byte = 0x17
)

var Skip = errors.New("skip")

type Unmarshaler interface {
	UnmarshalBER(byte, []byte) error
}

type Marshaler interface {
	MarshalBER() ([]byte, error)
}

type reader interface {
	ReadByte() (byte, error)
	io.Reader
}

type writer interface {
	WriteByte(byte) error
	io.Writer
}

type Decoder struct {
	r reader
}

func NewDecoder(r io.Reader) *Decoder {
	d := new(Decoder)
	if rd, ok := r.(reader); ok {
		d.r = rd
	} else {
		d.r = bufio.NewReaderSize(r, 8192)
	}
	return d
}

func (d *Decoder) Decode(v interface{}) error {
	return unmarshal(d.r, reflect.ValueOf(v).Elem())
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (e *Encoder) Encode(v interface{}) error {
	b := bufio.NewWriter(e.w)
	if err := marshal(b, reflect.ValueOf(v), ""); err != nil {
		return err
	}
	return b.Flush()
}

func Marshal(v interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	if err := marshal(b, reflect.ValueOf(v), ""); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Unmarshal(bs []byte, v interface{}) error {
	r := bytes.NewReader(bs)
	return unmarshal(r, reflect.ValueOf(v).Elem())
}

func Length(bs []byte) []byte {
	return sizeOf(len(bs))
}
