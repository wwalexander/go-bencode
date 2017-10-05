package bencode

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecode_string(t *testing.T) {
	r := strings.NewReader("3:foo")
	var v []byte
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if bytes.Compare(v, []byte("foo")) != 0 {
		t.Error("decoded wrong value for string")
	}
}

func TestDecode_string_empty(t *testing.T) {
	r := strings.NewReader("0:")
	var v []byte
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if bytes.Compare(v, []byte{}) != 0 {
		t.Error("decoded wrong value for string")
	}
}

func TestDecode_integer_zero(t *testing.T) {
	r := strings.NewReader("i0e")
	var v int
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if v != 0 {
		t.Error("decoded wrong value for integer")
	}
}

func TestDecode_integer_one(t *testing.T) {
	r := strings.NewReader("i1e")
	var v int
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Error("decoded wrong value for integer")
	}
}

func TestDecode_integer_negone(t *testing.T) {
	r := strings.NewReader("i-1e")
	var v int
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if v != -1 {
		t.Error("decoded wrong value for integer")
	}
}

func TestDecode_integer_ten(t *testing.T) {
	r := strings.NewReader("i10e")
	var v int
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if v != 10 {
		t.Error("decoded wrong value for integer")
	}
}

func TestDecode_integer_negten(t *testing.T) {
	r := strings.NewReader("i-10e")
	var v int
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if v != -10 {
		t.Error("decoded wrong value for integer")
	}
}

func TestDecode_list(t *testing.T) {
	r := strings.NewReader("li-10ei-1ei0ei1ei10ee")
	var v []int
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if len(v) != 5 {
		t.Fatal("wrong length of list")
	}
	for i, n := range []int{-10, -1, 0, 1, 10} {
		if v[i] != n {
			t.Error("wrong value in list")
		}
	}
}

func TestDecode_list_empty(t *testing.T) {
	r := strings.NewReader("le")
	var v []int
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal("wrong length of list")
	}
}

func TestDecode_struct(t *testing.T) {
	r := strings.NewReader("d4:fizzi3e4:fuzzi4e4:buzzi5ee")
	var v struct {
		Bizz int
		Fizz int `bencode:"fizz"`
		Buzz int `bencode:"buzz"`
	}
	if err := NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v.Bizz != 0 {
		t.Error("decoded struct field without bencode tag")
	} else if v.Fizz != 3 || v.Buzz != 5 {
		t.Error("wrong value(s) in struct")
	}
}
