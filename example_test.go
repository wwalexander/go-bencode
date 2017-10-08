package bencode_test

import (
	"fmt"
	"github.com/wwalexander/go-bencode"
	"io"
	"log"
	"os"
	"strings"
)

func ExampleMarshal() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}
	b, err := bencode.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
	// Output:
	// d6:Colorsl7:Crimson3:Red4:Ruby6:Maroone2:IDi1e4:Name4:Redse
}

func ExampleUnmarshal() {
	var bencodeBlob = []byte("ld4:Name8:Platypus5:Order11:Monotremataed4:Name5:Quoll5:Order14:Dasyuromorphiaee")
	type Animal struct {
		Name  string
		Order string
	}
	var animals []Animal
	err := bencode.Unmarshal(bencodeBlob, &animals)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", animals)
	// Output:
	// [{Name:Platypus Order:Monotremata} {Name:Quoll Order:Dasyuromorphia}]
}

// This example uses a Decoder to decode a stream of distinct bencoded values.
func ExampleDecoder() {
	const bencodeStream = `` +
		`d4:Name2:Ed4:Text12:Knock knock.e` +
		`d4:Name3:Sam4:Text12:Who's there?e` +
		`d4:Name2:Ed4:Text7:Go fmt.e` +
		`d4:Name3:Sam4:Text11:Go fmt who?e` +
		`d4:Name2:Ed4:Text16:Go fmt yourself!e`
	type Message struct {
		Name, Text string
	}
	dec := bencode.NewDecoder(strings.NewReader(bencodeStream))
	for {
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", m.Name, m.Text)
	}
	// Output:
	// Ed: Knock knock.
	// Sam: Who's there?
	// Ed: Go fmt.
	// Sam: Go fmt who?
	// Ed: Go fmt yourself!
}
