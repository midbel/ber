package ber

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Marshaler interface {
	Marshal() ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal([]byte) error
}

const (
	Universal uint8 = iota
	Application
	Context
	Private
)

const (
	Primitive uint8 = iota
	Constructed
)

var (
	Bool            Ident = NewPrimitive(0x01)
	Int                   = NewPrimitive(0x02)
	BitString             = NewPrimitive(0x03)
	OctetString           = NewPrimitive(0x04)
	Null                  = NewPrimitive(0x05)
	ObjectId              = NewPrimitive(0x06)
	RelObjectId           = NewPrimitive(0x0d)
	Real                  = NewPrimitive(0x09)
	Enumerated            = NewPrimitive(0x0a)
	UTF8String            = NewPrimitive(0x0c)
	Sequence              = NewConstructed(0x10)
	Set                   = NewConstructed(0x11)
	PrintableString       = NewPrimitive(0x13)
	IA5String             = NewPrimitive(0x16)
	UniversalTime         = NewPrimitive(0x17)
	GeneralizedTime       = NewPrimitive(0x18)
)

const (
	patGeneralTime = "20060102150405-0700"
	patUniversTime = "060102150405-0700"
)

type Raw []byte

func (r *Raw) Peek() (Ident, error) {
	id, _, err := decodeIdentifier([]byte(*r))
	return id, err
}

type Ident uint64

func NewPrimitive(code uint64) Ident {
	return Ident(code)
}

func NewConstructed(code uint64) Ident {
	i := Ident(code)
	return i.Constructed()
}

func (i Ident) Class() uint8 {
	val := uint64(i) >> 33
	return uint8(val) & 0x3
}

func (i Ident) Type() uint8 {
	val := uint64(i) >> 32
	return uint8(val) & 0x1
}

func (i Ident) Tag() uint32 {
	val := uint64(i) & 0xFFFFFFFF
	return uint32(val)
}

func (i Ident) Primitive() Ident {
	v := uint64(i) | (uint64(Primitive) << 32)
	return Ident(v)
}

func (i Ident) Constructed() Ident {
	v := uint64(i) | (uint64(Constructed) << 32)
	return Ident(v)
}

func (i Ident) Universal() Ident {
	v := uint64(i) | (uint64(Universal) << 33)
	return Ident(v)
}

func (i Ident) Context() Ident {
	v := uint64(i) | (uint64(Context) << 33)
	return Ident(v)
}

func (i Ident) Application() Ident {
	v := uint64(i) | (uint64(Application) << 33)
	return Ident(v)
}

func (i Ident) Private() Ident {
	v := uint64(i) | (uint64(Private) << 33)
	return Ident(v)
}

func (i Ident) isZero() bool {
	return i == 0
}

func (i Ident) setTag(tag uint32) Ident {
	v := uint64(i.clearTag()) | uint64(tag)
	return Ident(v)
}

func (i Ident) clearTag() Ident {
	v := uint64(i) &^ ((1 << 5) - 1)
	return Ident(v)
}

func parseTag(str string, i Ident) (Ident, bool, error) {
	var omit bool
	for _, str := range strings.Split(str, ",") {
		switch {
		case strings.HasPrefix(str, "tag:"):
			str = strings.TrimSpace(strings.TrimPrefix(str, "tag:"))
			x, err := strconv.ParseUint(str, 0, 32)
			if err != nil {
				return i, omit, err
			}
			i = i.setTag(uint32(x))
		case strings.HasPrefix(str, "class:"):
			str = strings.TrimSpace(strings.TrimPrefix(str, "class:"))
			x, err := strconv.ParseUint(str, 0, 8)
			if err != nil {
				return i, omit, err
			}
			y := uint8(x)
			if y == Universal {
				i = i.Universal()
			} else if y == Context {
				i = i.Context()
			} else if y == Application {
				i = i.Application()
			} else if y == Private {
				i = i.Private()
			} else {
				return i, omit, fmt.Errorf("%x: invalid class", x)
			}
		case strings.HasPrefix(str, "type:"):
			str = strings.TrimSpace(strings.TrimPrefix(str, "type:"))
			x, err := strconv.ParseUint(str, 0, 8)
			if err != nil {
				return i, omit, err
			}
			y := uint8(x)
			if y == Primitive {
				i = i.Primitive()
			} else if y == Constructed {
				i = i.Constructed()
			} else {
				return i, omit, fmt.Errorf("%x: invalid type", x)
			}
		case str == "omitempty":
			omit = true
		case str == "enumerated":
			i = Enumerated
		case str == "sequence":
			i = Sequence
		case str == "set":
			i = Set
		case str == "utc":
			i = UniversalTime
		case str == "generalized":
			i = GeneralizedTime
		case str == "ia5":
			i = IA5String
		case str == "printable":
			i = PrintableString
		case str == "utf8":
			i = UTF8String
		case str == "octetstr":
			i = OctetString
		case str == "oid":
			i = ObjectId
		case str == "roid":
			i = RelObjectId
		default:
		}
	}
	return i, omit, nil
}

func ValidPrintableString(str string) bool {
	isLetter := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
	}
	isDigit := func(r rune) bool {
		return r >= '0' && r <= '9'
	}
	isPunct := func(r rune) bool {
		switch r {
		case ' ', '\'', '(', ')', '+', ',', '-', '.', '/', ':', '=', '?':
			return true
		default:
			return false
		}
	}
	accept := func(r rune) bool {
		return isLetter(r) || isDigit(r) || isPunct(r)
	}
	return validString(str, accept)
}

func ValidIA5String(str string) bool {
	return validString(str, func(r rune) bool { return r >= 0 && r <= 127 })
}

func validString(str string, accept func(rune) bool) bool {
	var i int
	for i < len(str) {
		c, z := utf8.DecodeRuneInString(str[i:])
		if !accept(c) {
			return false
		}
		i += z
	}
	return true
}

func validTimeUTC(t time.Time) bool {
	return t.Year() >= 1950 && t.Year() <= 2050
}

func validTimeGeneralized(t time.Time) bool {
	return t.Year() >= 0 && t.Year() <= 9999
}
