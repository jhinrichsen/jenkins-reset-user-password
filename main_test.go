package main

import (
	"encoding/xml"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestIntersect(t *testing.T) {
	s1 := []string{"a", "b", "c"}
	s2 := []string{"b"}
	want := []string{"b"}
	got := intersect(s1, s2)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %+v but got %+v\n", want, got)
	}
}

func TestMinus(t *testing.T) {
	s1 := []string{"a", "b", "c"}
	s2 := []string{"b"}
	// map may change order
	want1 := []string{"a", "c"}
	want2 := []string{"c", "a"}
	got := minus(s1, s2)
	if !reflect.DeepEqual(want1, got) && !reflect.DeepEqual(want2, got) {
		t.Fatalf("want %+v or %+v but got %+v\n", want1, want2, got)
	}
}

func TestXmlStreamingCopy(t *testing.T) {
	fromfile, err := os.Open("testdata/config.xml")
	die(err)
	defer fromfile.Close()
	dec := xml.NewDecoder(fromfile)

	intofile, err := os.Create("testdata/copy.xml")
	die(err)
	defer intofile.Close()
	enc := xml.NewEncoder(intofile)
	enc.Indent("", indentation)

	for {
		// Token() returns nil, io.EOF at end of input stream
		tok, err := dec.Token()
		// end of stream?
		if tok == nil && err == io.EOF {
			break
		}
		die(err)
		err = enc.EncodeToken(tok)
		die(err)
	}
	enc.Flush()
}

func TestXmlStreamingModify(t *testing.T) {
	resetPassword("testdata/config.xml", "testdata/update.xml")
}
