package ber

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

func Debug(r io.Reader, w io.Writer) error {
	return debug(bufio.NewReader(r), w, " ")
}

func debug(r *bufio.Reader, w io.Writer, s string) error {
	t, err := r.ReadByte()
	if err != nil {
		return nil
	}
	z := readLength(r)
	if z == 0 {
		return fmt.Errorf("done")
	}
	vs := make([]byte, z)
	if _, err := r.Read(vs); err != nil {
		return err
	}
	if (t & (1 << 5)) == TypeCompound<<5 {
		fmt.Fprintf(w, "%s[%s (%d)]\n", s[:len(s)-1], "compound", z)
		s += s

		o := bufio.NewReader(bytes.NewReader(vs))
		if err := debug(o, w, s); err != nil {
			return err
		}
	} else {
		var n string
		switch t := t & (1<<5 - 1); t {
		default:
			n = "other"
		case TagString:
			n = "string"
		case TagBoolean:
			n = "bool"
		case TagReal:
			n = "float"
		case TagInteger:
			n = "integer"
		}
		fmt.Fprintf(w, "%s > %-8s(%02x) - % 3d - %x\n", s, n, t, z, vs)
	}
	return debug(r, w, s)
}
