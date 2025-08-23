package scan

import (
	"bufio"
	"strings"
	"testing"
)

func TestPeek(t *testing.T) {
	tcases := []struct {
		name   string
		in     string
		peeked byte
	}{{
		"empty", "", 0,
	}, {
		"123", "123", '1',
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			peeked := b.Peek()
			if string(b) != tc.in {
				t.Errorf("left: got: %s, want: %s", string(b), tc.in)
			}
			if peeked != tc.peeked {
				t.Errorf("peeked: got: %c, want: %c", peeked, tc.peeked)
			}
		})
	}
}

func TestPop(t *testing.T) {
	tcases := []struct {
		name   string
		in     string
		popped byte
		left   string
	}{{
		"empty", "", 0, "",
	}, {
		"123", "123", '1', "23",
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			popped := b.Pop()
			if string(b) != tc.left {
				t.Errorf("left: got: %s, want: %s", string(b), tc.left)
			}
			if popped != tc.popped {
				t.Errorf("popped: got: %c, want: %c", popped, tc.popped)
			}
		})
	}
}

func TestPopN(t *testing.T) {
	tcases := []struct {
		name   string
		in     string
		n      int
		popped string
		left   string
	}{{
		"empty", "", 0, "", "",
	}, {
		"1,0", "1", 0, "", "1",
	}, {
		"12,0", "12", 0, "", "12",
	}, {
		"1,1", "1", 1, "1", "",
	}, {
		"12,1", "12", 1, "1", "2",
	}, {
		"123,1", "123", 1, "1", "23",
	}, {
		"123,2", "123", 2, "12", "3",
	}, {
		"123,3", "123", 3, "123", "",
	}, {
		"123,4", "123", 4, "", "123",
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			popped := b.PopN(tc.n)
			if string(b) != tc.left {
				t.Errorf("left: got: %s, want: %s", string(b), tc.left)
			}
			if string(popped) != tc.popped {
				t.Errorf("popped: got: %s, want: %s", string(popped), tc.popped)
			}
		})
	}
}
func TestTrim(t *testing.T) {
	tcases := []struct {
		name  string
		in    string
		left  string
		right string
	}{{
		"empty", "", "", "",
	}, {
		"one space", " ", "", "",
	}, {
		"all spaces", " \r\n\t\x0c", "", "",
	}, {
		"one char and spaces", " \r\n\t\x0ca \r\n\t\x0c", "a \r\n\t\x0c", " \r\n\t\x0ca",
	}, {
		"one char", "a", "a", "a",
	}, {
		// Unicode Ogham space mark
		"unicode space ogham", " ", " ", " ",
	}, {
		// Unicode Em space mark
		"unicode em space", "\u2003", "\u2003", "\u2003",
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			b.TrimLWS()
			if string(b) != tc.left {
				t.Errorf("left: got: %s, want: %s", string(b), tc.left)
			}

			b = Bytes(tc.in)
			b.TrimRWS()
			if string(b) != tc.right {
				t.Errorf("right: got: %s, want: %s", string(b), tc.right)
			}
		})
	}
}

func TestAdvance(t *testing.T) {
	tcases := []struct {
		name     string
		in       string
		advance  int
		want     string
		shouldDo bool
	}{{
		"empty 0", "", 0, "", true,
	}, {
		"empty 1", "", 1, "", false,
	}, {
		"empty -1", "", -1, "", false,
	}, {
		"123 0", "123", 0, "123", true,
	}, {
		"123 -1", "123", -1, "123", false,
	}, {
		"123 1", "123", 1, "23", true,
	}, {
		"123 4", "123", 4, "123", false,
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			did := b.Advance(tc.advance)
			if did != tc.shouldDo {
				t.Errorf("got: %t, want: %t", did, tc.shouldDo)
			}
			if string(b) != tc.want {
				t.Errorf("got: %s, want: %s", string(b), tc.want)
			}
		})
	}
}

func TestLine(t *testing.T) {
	tcases := []struct {
		name     string
		in       string
		line     string
		leftover string
	}{{
		"empty", "", "", "",
	}, {
		"one line", "abc", "abc", "",
	}, {
		"just a \\n", "\n", "", "",
	}, {
		"just two \\n", "\n\n", "", "\n",
	}, {
		"one line with \\n", "abc\n", "abc", "",
	}, {
		"two lines", "abc\ndef", "abc", "def",
	}, {
		"two lines with \\n", "abc\ndef\n", "abc", "def\n",
	}, {
		"drops final cr", "abc\r", "abc", "",
	}, {
		"cr inside line", "abc\rdef", "abc\rdef", "",
	}, {
		"nl and cr", "\n\r", "", "\r",
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			line := b.Line()
			if string(line) != tc.line {
				t.Errorf("line: got: %s, want: %s", line, []byte(tc.line))
			}
			if string(b) != tc.leftover {
				t.Errorf("leftover: got: %s, want: %s", b, []byte(tc.leftover))
			}

			// Test if it behaves like bufio.Scanner as well.
			s := bufio.NewScanner(strings.NewReader(tc.in))
			s.Scan()
			if string(line) != s.Text() {
				t.Errorf("Bytes.Line not like bufio.Scanner")
			}
		})
	}
}

func TestPopUntil(t *testing.T) {
	tcases := []struct {
		name     string
		in       string
		untilAny string
		popped   string
		leftover string
	}{{
		"empty", "", "", "", "",
	}, {
		"empty with until", "", "123", "", "",
	}, {
		"until empty", "123", "", "123", "",
	}, {
		"until 1", "123", "1", "", "123",
	}, {
		"until 2", "123", "2", "1", "23",
	}, {
		"until 3", "123", "3", "12", "3",
	}, {
		"until 4", "123", "4", "123", "",
	}, {
		"multiple untilAny", "123", "32", "1", "23",
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			popped := b.PopUntil([]byte(tc.untilAny)...)
			if string(popped) != tc.popped {
				t.Errorf("popped: got: %s, want: %s", popped, []byte(tc.popped))
			}
			if string(b) != tc.leftover {
				t.Errorf("leftover: got: %s, want: %s", b, []byte(tc.leftover))
			}
		})
	}
}

func TestReadSlice(t *testing.T) {
	tcases := []struct {
		name     string
		in       string
		stopAt   byte
		popped   string
		leftover string
	}{{
		"both empty", "", 0, "", "",
	}, {
		"stop at not found", "abc", 'd', "abc", "",
	}, {
		"stop at the end", "abc", 'c', "abc", "",
	}, {
		"stop at in the middle", "abcdef", 'c', "abc", "def",
	}, {
		"stop at the beginning", "abcdef", 'a', "a", "bcdef",
	}, {
		"just one char", "a", 'a', "a", "",
	}, {
		"same char twice", "aa", 'a', "a", "a",
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			got := b.ReadSlice(tc.stopAt)
			if tc.popped != string(got) {
				t.Errorf("popped got: %s, want: %s", got, tc.popped)
			}
			if tc.leftover != string(b) {
				t.Errorf("leftover got: %s, want: %s", string(b), tc.leftover)
			}
		})
	}
}

func TestUint16(t *testing.T) {
	tcases := []struct {
		name string
		in   []byte
		res  uint16
		ok   bool
	}{{
		"empty", nil, 0, false,
	}, {
		"too short", []byte{0}, 0, false,
	}, {
		"just enough", []byte{1, 0}, 1, true,
	}, {
		"longer", []byte{1, 0, 2}, 1, true,
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			res, ok := b.Uint16()
			if res != tc.res {
				t.Errorf("got: %d, want: %d", res, tc.res)
			}
			if ok != tc.ok {
				t.Errorf("ok: got: %t, want: %t", ok, tc.ok)
			}
		})
	}
}

var searchTestcases = []struct {
	name     string
	haystack string
	needle   string
	flags    int
	expect   int
}{{
	"empty", "", "", 0, 0,
}, {
	"empty compact ws", "", "", CompactWS, 0,
}, {
	"empty ignore case", "", "", IgnoreCase, 0,
}, {
	"simple", "abc", "abc", 0, 0,
}, {
	"simple compact ws", "abc", "abc", CompactWS, 0,
}, {
	"simple ignore case", "abc", "abc", IgnoreCase, 0,
}, {
	"ignore case 1 upper", "aBc", "ABC", IgnoreCase, 0,
}, {
	"ignore case prefixed", "aaBcß", "ABC", IgnoreCase, 1,
}, {
	"ignore case prefixed utf8", "ßaBcß", "ABC", IgnoreCase, 2, // 2 because ß is 2 bytes long
}, {
	"simple compact ws and ignore case", "  a", " A", CompactWS | IgnoreCase, 0,
}, {
	"simple compact ws and ignore prefix", "a  a", " A", CompactWS | IgnoreCase, 1,
}}

func TestSearch(t *testing.T) {
	for _, tc := range searchTestcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.haystack)
			i := b.Search([]byte(tc.needle), tc.flags)
			if i != tc.expect {
				t.Errorf("got: %d, want: %d", i, tc.expect)
			}
		})
	}
}

func FuzzSearch(f *testing.F) {
	for _, tc := range searchTestcases {
		f.Add([]byte(tc.haystack), []byte(tc.needle), tc.flags)
	}
	f.Fuzz(func(t *testing.T, haystack, needle []byte, flags int) {
		b := Bytes(haystack)
		b.Search(needle, flags%CompactWS|IgnoreCase)
	})
}
