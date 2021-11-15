package magic

import (
	"io"
	"testing"
)

var magicTests = []struct {
	raw      string
	limit    uint32
	res      bool
	detector Detector
}{
	{`["an incomplete JSON array`, 0, false, JSON},
}

func TestMagic(t *testing.T) {
	for i, tt := range magicTests {
		if got := tt.detector([]byte(tt.raw), tt.limit); got != tt.res {
			t.Errorf("Detector %d error: expected: %t; got: %t", i, tt.res, got)
		}
	}
}

var dropTests = []struct {
	raw   string
	cutAt uint32
	res   string
}{
	{"", 0, ""},
	{"", 1, ""},
	{"å", 2, "å"},
	{"\n", 0, "\n"},
	{"\n", 1, "\n"},
	{"\n\n", 1, "\n"},
	{"\n\n", 3, "\n\n"},
	{"a\n\n", 3, "a\n"},
	{"\na\n", 3, "\na"},
	{"å\n\n", 5, "å\n\n"},
	{"\nå\n", 5, "\nå\n"},
}

func TestDropLastLine(t *testing.T) {
	for i, tt := range dropTests {
		gotR := dropLastLine([]byte(tt.raw), tt.cutAt)
		got, _ := io.ReadAll(gotR)
		if got := string(got); got != tt.res {
			t.Errorf("dropLastLine %d error: expected %q; got %q", i, tt.res, got)
		}
	}
}
