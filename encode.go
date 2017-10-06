package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
	"sort"
	"strconv"
)

// An Encoder writes bencoded values to an output stream.
type Encoder struct {
	b *bufio.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	b := bufio.NewWriter(w)
	return &Encoder{b}
}

func (enc *Encoder) encode(v interface{}) (err error) {
	defer func() {
		if err != nil {
			return
		}
		err = enc.b.Flush()
	}()
	val := reflect.ValueOf(v)
	typ := val.Type()
	kind := typ.Kind()
	switch {
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		data := v.([]byte)
		length := strconv.Itoa(len(data))
		if _, err := enc.b.Write([]byte(length)); err != nil {
			return err
		}
		if err := enc.b.WriteByte(':'); err != nil {
			return err
		}
		if _, err := enc.b.Write(data); err != nil {
			return err
		}
	case kind == reflect.Int:
		if err := enc.b.WriteByte('i'); err != nil {
			return err
		}
		n := strconv.Itoa(v.(int))
		if _, err := enc.b.Write([]byte(n)); err != nil {
			return err
		}
		if err := enc.b.WriteByte('e'); err != nil {
			return err
		}
	case kind == reflect.Slice:
		if err := enc.b.WriteByte('l'); err != nil {
			return err
		}
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			if err := enc.Encode(elem); err != nil {
				return err
			}
		}
		if err := enc.b.WriteByte('e'); err != nil {
			return err
		}
	case kind == reflect.Struct:
		if err := enc.b.WriteByte('d'); err != nil {
			return err
		}
		names := make([]string, 0, val.NumField())
		fields := make(map[string]int)
		for i := 0; i < val.NumField(); i++ {
			name, ok := typ.Field(i).Tag.Lookup("bencode")
			if !ok {
				continue
			}
			names = append(names, name)
			fields[name] = i
		}
		sort.Sort(sort.StringSlice(names))
		for _, name := range names {
			if err := enc.Encode([]byte(name)); err != nil {
				return err
			}
			i := fields[name]
			field := val.Field(i).Interface()
			if err := enc.Encode(field); err != nil {
				return err
			}
		}
		if err := enc.b.WriteByte('e'); err != nil {
			return err
		}
	default:
		return errors.New("unsupported type")
	}
	return nil
}

// Encode writes the bencoding of v to the stream.
//
// See the documentation for Marshal for details about the conversion of Go
// values to bencoded values.
func (enc *Encoder) Encode(v interface{}) error {
	return enc.encode(v)
}

// Marshal returns the bencoding of v.
//
// Marshal traverses the value v recursively.
//
// Marshal uses the following type-dependent encodings:
//
// int values encode as bencoded integers.
//
// Slice values encode as bencoded lists, except that []byte encodes as a
// bencoded string.
//
// Struct values encode as bencoded dictionaries. Each exported struct field
// with a "bencode" key in its tag becomes an entry in the dictionary, using the
// key as the dictionary key. Dictionaries are encoded with their entries sorted
// by their keys as raw strings.
func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
