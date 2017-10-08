package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

func (enc *Encoder) encodeString(v []byte) error {
	length := strconv.Itoa(len(v))
	if _, err := enc.b.Write([]byte(length)); err != nil {
		return err
	}
	if err := enc.b.WriteByte(':'); err != nil {
		return err
	}
	if _, err := enc.b.Write(v); err != nil {
		return err
	}
	return nil
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
	case kind == reflect.String:
		data := []byte(v.(string))
		if err := enc.encodeString(data); err != nil {
			return err
		}
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		if err := enc.encodeString(v.([]byte)); err != nil {
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
			field := typ.Field(i)
			name, ok := field.Tag.Lookup("bencode")
			if ok {
				if name == "-" {
					continue
				}
				strs := strings.Split(name, ",")
				name = strs[0]
				if len(strs) > 1 {
					opts := make(map[string]bool)
					for _, s := range strs[1:] {
						opts[s] = true
					}
					if _, ok := opts["omitempty"]; ok {
						field := val.Field(i)
						x := field.Interface()
						y := reflect.Zero(field.Type()).Interface()
						if reflect.DeepEqual(x, y) {
							continue
						}
					}
				}
			} else {
				name = field.Name
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
// String values encode as bencoded strings.
//
// Slice values encode as bencoded lists, except that []byte encodes as a
// bencoded string.
//
// Struct values encode as bencoded dictionaries. Each exported struct field
// becomes a member of the object, using the field name as the dictionary key,
// unless the field is omitted for one of the reasons given below.
//
// The encoding of each struct field can be customized by the format string
// stored under the "bencode" key in the struct field's tag. The format string
// gives the name of the field, possibly followed by a comma-separated list of
// options. The name may be empty in order to specify options without overriding
// the default field name.
//
// The "omitempty" option specifies that the field should be omitted from the
// encoding if the field has an empty value, defined as 0 and any empty slice or
// string.
//
// As a special case, if the field tag is "-", the field is always omitted. Note
// that a field with name "-" can still be generated using the tag "-,".
//
// Examples of struct field tags and their meanings:
//
//   // Field appears in bencoded as key "myName".
//   Field int `bencode:"myName"`
//
//   // Field appears bencoded as key "myName" and
//   // the field is omitted from the dictionary if its value is empty,
//   // as defined above.
//   Field int `bencode:"myName,omitempty"`
//
//   // Field appears bencoded as key "Field" (the default), but
//   // the field is skipped if empty.
//   // Note the leading comma.
//   Field int `bencode:",omitempty"`
//
//   // Field is ignored by this package.
//   Field int `bencode:"-"`
//
//   // Field appears bencoded as key "-".
//   Field int `bencode:"-,"`
func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
