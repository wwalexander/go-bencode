package bencode

import (
	"bufio"
	"errors"
	"io"
	"reflect"
	"strconv"
)

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{bufio.NewReader(r)}
}

func (d *Decoder) decodeInteger(delim byte) (int, error) {
	s, err := d.r.ReadBytes(delim)
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

func (d *Decoder) decodeString() ([]byte, error) {
	length, err := d.decodeInteger(':')
	if err != nil {
		return nil, err
	}
	s := make([]byte, length)
	length = 0
	for length < len(s) {
		n, err := d.r.Read(s[length:])
		if err != nil {
			return nil, err
		}
		length += n
	}
	return s, nil
}

func (d *Decoder) next(c byte) (bool, error) {
	buf, err := d.r.Peek(1)
	if err != nil {
		return false, err
	}
	if buf[0] != c {
		return false, nil
	}
	if _, err := d.r.ReadByte(); err != nil {
		return true, err
	}
	return true, nil
}

func (d *Decoder) Decode(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("non-pointer type")
	}
	val = val.Elem()
	typ := val.Type()
	kind := typ.Kind()
	switch {
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		s, err := d.decodeString()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(s))
	case kind == reflect.Int:
		if ok, err := d.next('i'); err != nil {
			return err
		} else if !ok {
			return errors.New("invalid prefix for int")
		}
		n, err := d.decodeInteger('e')
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(n))
	case kind == reflect.Slice:
		if ok, err := d.next('l'); err != nil {
			return err
		} else if !ok {
			return errors.New("invalid prefix for list")
		}
		etyp := typ.Elem()
		for {
			if done, err := d.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			elem := reflect.New(etyp)
			if err := d.Decode(elem.Interface()); err != nil {
				return err
			}
			val.Set(reflect.Append(val, elem.Elem()))
		}
	case kind == reflect.Struct:
		if ok, err := d.next('d'); err != nil {
			return err
		} else if !ok {
			return errors.New("invalid prefix for dictionary")
		}
		fields := make(map[string]reflect.Value)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := field.Tag.Get("bencode")
			if name == "" {
				continue
			}
			fields[name] = val.Field(i)
		}
		for {
			if done, err := d.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			key, err := d.decodeString()
			if err != nil {
				return err
			}
			field, ok := fields[string(key)]
			if !ok {
				if err := d.discard(); err != nil {
					return err
				}
				continue
			}
			if err := d.Decode(field.Addr().Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Decoder) discard() error {
	buf, err := d.r.Peek(1)
	if err != nil {
		return err
	}
	c := buf[0]
	switch {
	case c >= '0' && c <= '9':
		s, err := d.r.ReadBytes(':')
		if err != nil {
			return err
		}
		s = s[:len(s)-1]
		n, err := strconv.Atoi(string(s))
		if err != nil {
			return err
		}
		if _, err := d.r.Discard(n); err != nil {
			return err
		}
	case c == 'i':
		if _, err := d.r.ReadBytes('e'); err != nil {
			return err
		}
	case c == 'l':
		if _, err := d.r.ReadByte(); err != nil {
			return err
		}
		for {
			if done, err := d.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			if err := d.discard(); err != nil {
				return err
			}
		}
	case c == 'd':
		if _, err := d.r.ReadByte(); err != nil {
			return err
		}
		for {
			if done, err := d.next('e'); err != nil {
				return err
			} else if done {
				break
			}
			if err := d.discard(); err != nil {
				return err
			}
			if err := d.discard(); err != nil {
				return err
			}
		}
	}
	return nil
}
