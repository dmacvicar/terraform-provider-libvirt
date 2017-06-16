package dataurl

import (
	"bytes"
	"fmt"
	"testing"
)

var tests = []struct {
	escaped   string
	unescaped []byte
}{
	{"A%20brief%20note%0A", []byte("A brief note\n")},
	{"%7B%5B%5Dbyte(%22A%2520brief%2520note%22)%2C%20%5B%5Dbyte(%22A%20brief%20note%22)%7D", []byte(`{[]byte("A%20brief%20note"), []byte("A brief note")}`)},
}

func TestEscape(t *testing.T) {
	for _, test := range tests {
		escaped := Escape(test.unescaped)
		if string(escaped) != test.escaped {
			t.Errorf("Expected %s, got %s", test.escaped, string(escaped))
		}
	}
}

func TestUnescape(t *testing.T) {
	for _, test := range tests {
		unescaped, err := Unescape(test.escaped)
		if err != nil {
			t.Error(err)
			continue
		}
		if !bytes.Equal(unescaped, test.unescaped) {
			t.Errorf("Expected %s, got %s", test.unescaped, unescaped)
		}
	}
}

func ExampleEscapeString() {
	fmt.Println(EscapeString("A brief note"))
	// Output: A%20brief%20note
}

func ExampleEscape() {
	fmt.Println(Escape([]byte("A brief note")))
	// Output: A%20brief%20note
}

func ExampleUnescape() {
	data, err := Unescape("A%20brief%20note")
	if err != nil {
		// can fail e.g if incorrect escaped sequence
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: A brief note
}

func ExampleUnescapeToString() {
	s, err := UnescapeToString("A%20brief%20note")
	if err != nil {
		// can fail e.g if incorrect escaped sequence
		fmt.Println(err)
		return
	}
	fmt.Println(s)
	// Output: A brief note
}
