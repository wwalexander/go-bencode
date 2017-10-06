package bencode

import (
	"bytes"
	"testing"
)

func TestEncode_string(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode([]byte("foo")); err != nil {
		t.Fatal(err)
	} else if buf.String() != "3:foo" {
		t.Error("encoded wrong value for string")
	}
}

func TestEncode_string_empty(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode([]byte("")); err != nil {
		t.Fatal(err)
	} else if buf.String() != "0:" {
		t.Error("encoded wrong value for string")
	}
}

func TestEncode_integer_zero(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(0); err != nil {
		t.Fatal(err)
	} else if buf.String() != "i0e" {
		t.Error("encoded wrong value for integer")
	}
}

func TestEncode_integer_one(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(1); err != nil {
		t.Fatal(err)
	} else if buf.String() != "i1e" {
		t.Error("encoded wrong value for integer")
	}
}

func TestEncode_integer_negone(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(-1); err != nil {
		t.Fatal(err)
	} else if buf.String() != "i-1e" {
		t.Error("encoded wrong value for integer")
	}
}

func TestEncode_integer_ten(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(10); err != nil {
		t.Fatal(err)
	} else if buf.String() != "i10e" {
		t.Error("encoded wrong value for integer")
	}
}

func TestEncode_integer_negten(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(-10); err != nil {
		t.Fatal(err)
	} else if buf.String() != "i-10e" {
		t.Error("encoded wrong value for integer")
	}
}

func TestEncode_list(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode([]int{-10, -1, 0, 1, 10}); err != nil {
		t.Fatal(err)
	} else if buf.String() != "li-10ei-1ei0ei1ei10ee" {
		t.Error("encoded wrong value for list")
	}
}

func TestEncode_list_empty(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode([]int{}); err != nil {
		t.Fatal(err)
	} else if buf.String() != "le" {
		t.Error("encoded wrong value for list")
	}
}

func TestEncode_struct(t *testing.T) {
	var v = struct {
		Bizz int
		Fizz int `bencode:"fizz"`
		Buzz int `bencode:"buzz"`
	}{
		Fizz: 3,
		Buzz: 5,
	}
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(v); err != nil {
		t.Fatal(err)
	} else if buf.String() != "d4:buzzi5e4:fizzi3ee" {
		t.Error("encoded wrong value for struct")
	}
}
