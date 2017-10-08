package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// A Decoder reads and decodes bencoded values from an input stream.
type Decoder struct {
	b *bufio.Reader
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may read data from r beyond the
// bencoded values requested.
func NewDecoder(r io.Reader) *Decoder {
	b := bufio.NewReader(r)
	return &Decoder{b}
}

func (dec *Decoder) decodeInteger(delim byte) (int, error) {
	s, err := dec.b.ReadBytes(delim)
	if err != nil {
		return 0, err
	}
	s = s[:len(s)-1]
	n, err := strconv.Atoi(string(s))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (dec *Decoder) decodeString() ([]byte, error) {
	length, err := dec.decodeInteger(':')
	if err != nil {
		return nil, err
	}
	s := make([]byte, length)
	length = 0
	for length < len(s) {
		n, err := dec.b.Read(s[length:])
		if err != nil {
			return nil, err
		}
		length += n
	}
	return s, nil
}

func (dec *Decoder) next(c byte) (bool, error) {
	buf, err := dec.b.Peek(1)
	if err != nil {
		return false, err
	}
	if buf[0] != c {
		return false, nil
	}
	if _, err := dec.b.ReadByte(); err != nil {
		return true, err
	}
	return true, nil
}

// Decode reads the next bencoded value from its input and stores it in the
// value pointed to by v.
//
// See the documentation for Unmarshal for details about the conversion of
// a bencoded value into a Go value.
func (dec *Decoder) Decode(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("non-pointer type")
	}
	val = val.Elem()
	typ := val.Type()
	kind := typ.Kind()
	switch {
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		s, err := dec.decodeString()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(s))
	case kind == reflect.String:
		s, err := dec.decodeString()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(string(s)))
	case kind == reflect.Int:
		if ok, err := dec.next('i'); err != nil {
			return err
		} else if !ok {
			return errors.New("cannot unmarshal into Go value of type int")
		}
		n, err := dec.decodeInteger('e')
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(n))
	case kind == reflect.Slice:
		if ok, err := dec.next('l'); err != nil {
			return err
		} else if !ok {
			return errors.New("cannot unmarshal into Go slice")
		}
		if val.Len() > 0 {
			val.Set(val.Slice(0, 0))
		}
		etyp := typ.Elem()
		for {
			if done, err := dec.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			elem := reflect.New(etyp)
			if err := dec.Decode(elem.Interface()); err != nil {
				return err
			}
			val.Set(reflect.Append(val, elem.Elem()))
		}
	case kind == reflect.Struct:
		if ok, err := dec.next('d'); err != nil {
			return err
		} else if !ok {
			return errors.New("cannot unmarshal into Go struct")
		}
		fields := make(map[string]reflect.Value)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name, ok := field.Tag.Lookup("bencode")
			if ok {
				if name == "-" {
					continue
				}
				strs := strings.Split(name, ",")
				name = strs[0]
			} else {
				name = field.Name
			}
			fields[name] = val.Field(i)
		}
		for {
			if done, err := dec.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			key, err := dec.decodeString()
			if err != nil {
				return err
			}
			field, ok := fields[string(key)]
			if !ok {
				if err := dec.discard(); err != nil {
					return err
				}
				continue
			}
			if err := dec.Decode(field.Addr().Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported type")
	}
	return nil
}

func (dec *Decoder) discard() error {
	buf, err := dec.b.Peek(1)
	if err != nil {
		return err
	}
	c := buf[0]
	switch {
	case c >= '0' && c <= '9':
		s, err := dec.b.ReadBytes(':')
		if err != nil {
			return err
		}
		s = s[:len(s)-1]
		n, err := strconv.Atoi(string(s))
		if err != nil {
			return err
		}
		if _, err := dec.b.Discard(n); err != nil {
			return err
		}
	case c == 'i':
		if _, err := dec.b.ReadBytes('e'); err != nil {
			return err
		}
	case c == 'l':
		if _, err := dec.b.ReadByte(); err != nil {
			return err
		}
		for {
			if done, err := dec.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			if err := dec.discard(); err != nil {
				return err
			}
		}
	case c == 'd':
		if _, err := dec.b.ReadByte(); err != nil {
			return err
		}
		for {
			if done, err := dec.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			if err := dec.discard(); err != nil {
				return err
			}
			if err := dec.discard(); err != nil {
				return err
			}
		}
	default:
		return errors.New("invalid character looking for beginning of value")
	}
	return nil
}

// Unmarshal parses the bencoded data and stores the result in the value pointed
// to by v.
//
// Unmarshal uses the inverse of the encodings that Marshal uses, allocating
// slices and pointers as necessary, with the following additional rules:
//
// To unmarshal bencode into a struct, Unmarshal matches incoming dictionary
// keys to the key used by Marshal (either the struct field name or its tag).
// Unmarshal will only set exported fields of the struct.
//
// To unmarshal a bencoded list into a slice, Unmarshal resets the slice length
// to zero and then appends each element to the slice.
func Unmarshal(data []byte, v interface{}) error {
	r := bytes.NewReader(data)
	return NewDecoder(r).Decode(v)
}
