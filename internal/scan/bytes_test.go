package scan

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"testing"

	"math/rand"
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

func TestFirstNonWS(t *testing.T) {
	tcases := []struct {
		name string
		in   string
		c    byte
	}{{
		"empty", "", 0x00,
	}, {
		"all ws", "   ", 0x00,
	}, {
		"first char", "a", 'a',
	}, {
		"second char", " a", 'a',
	}, {
		"space then nil", " \x00", 0x00,
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.in)
			c := b.FirstNonWS()
			if c != tc.c {
				t.Errorf("got: %x, want: %x", c, tc.c)
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
	name      string
	haystack  string
	needle    string
	flags     Flags
	expectIdx int
	expectLen int
}{{
	"empty", "", "", 0, 0, 0,
}, {
	"empty cws", "", "", CompactWS, 0, 0,
}, {
	"empty ic", "", "", IgnoreCase, 0, 0,
}, {
	"just haystack", "abc", "", 0, 0, 0,
}, {
	"just haystack cws", "abc", "", CompactWS, 0, 0,
}, {
	"just haystack ic", "abc", "", IgnoreCase, 0, 0,
}, {
	"just needle", "", "abc", 0, -1, 0,
}, {
	"just needle cws", "", "abc", CompactWS, -1, 0,
}, {
	"just needle ic", "", "abc", IgnoreCase, -1, 0,
}, {
	"simple", "abc", "abc", 0, 0, 3,
}, {
	"not found", "abc", "def", 0, -1, 0,
}, {
	"simple cws", "abc", "abc", CompactWS, 0, 3,
}, {
	"simple ic", "abc", "abc", IgnoreCase, 0, 3,
}, {
	"ic 1 upper", "aBc", "ABC", IgnoreCase, 0, 3,
}, {
	"ic prefixed", "aaBcß", "ABC", IgnoreCase, 1, 3,
}, {
	"ic prefixed utf8", "ßaBcß", "ABC", IgnoreCase, 2, 3, // 2 because ß is 2 bytes long
}, {
	"simple cws|ic", "  a", " A", CompactWS | IgnoreCase, 0, 3,
}, {
	"simple cws|ic with suffix and prefix", "a  ab", " A", CompactWS | IgnoreCase, 1, 3,
}, {
	"trailing space in input", "a  a ", " A", CompactWS | IgnoreCase, 1, 3,
}, {
	"empty haystack with needle cws|ic", "", "abc", CompactWS | IgnoreCase, -1, 0,
}, {
	"empty haystack with needle cws", "", "abc", CompactWS, -1, 0,
}}

func TestSearch(t *testing.T) {
	for _, tc := range searchTestcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.haystack)
			i, l := b.Search([]byte(tc.needle), tc.flags)
			if i != tc.expectIdx || l != tc.expectLen {
				t.Errorf("want: %d,%d got: %d,%d", tc.expectIdx, tc.expectLen, i, l)
			}
		})
	}
}

func FuzzSearch(f *testing.F) {
	for _, tc := range searchTestcases {
		f.Add([]byte(tc.haystack), []byte(tc.needle), int(tc.flags))
	}
	f.Fuzz(func(t *testing.T, haystack, needle []byte, flags int) {
		b := Bytes(haystack)
		b.Search(needle, Flags(flags)%CompactWS|IgnoreCase|FullWord)
	})
}

var matchTestcases = []struct {
	name      string
	b         string
	p         string
	flags     Flags
	expectLen int
}{{
	"empty", "", "", 0, 0,
}, {
	"empty compact ws", "", "", CompactWS, 0,
}, {
	"empty ic", "", "", IgnoreCase, 0,
}, {
	"empty cws|ic", "", "", CompactWS | IgnoreCase, 0,
}, {
	"simple", "abc", "abc", 0, 3,
}, {
	"simple cws|ic", "abc", "abc", CompactWS | IgnoreCase, 3,
}, {
	"not found", "abc", "def", 0, -1,
}, {
	"simple cws", "abc", "abc", CompactWS, 3,
}, {
	"simple ic", "abc", "abc", IgnoreCase, 3,
}, {
	"ic 1 upper", "aBc", "ABC", IgnoreCase, 3,
}, {
	"ic prefixed", "aaBcß", "ABC", IgnoreCase, -1,
}, {
	"ic prefixed utf8", "ßaBcß", "ABC", IgnoreCase, -1,
}, {
	"simple cws|ic with space", "  a", " A", CompactWS | IgnoreCase, 3,
}, {
	"trailing space in input", "a  a ", " A", CompactWS | IgnoreCase, -1,
}, {
	"empty b with p", "", "/bin/bash", CompactWS, -1,
}, {
	"failing", "asd", "asdf", IgnoreCase, -1,
}, {
	"exact fw", "abc", "abc", FullWord, 3,
}, {
	"success fw", "abc ", "abc", FullWord, 3,
}, {
	"fail fw", "abcd", "abc", FullWord, -1,
}, { // #762
	"fw+ic", "abc ", "ABC", FullWord | IgnoreCase, 3,
}, {
	"fw+cws", "a  bc d", "a bc", FullWord | CompactWS, 5,
}, {
	"fw+ic+cws", "a  bc d", "A BC", FullWord | IgnoreCase | CompactWS, 5,
}}

func TestMatch(t *testing.T) {
	for _, tc := range matchTestcases {
		t.Run(tc.name, func(t *testing.T) {
			b := Bytes(tc.b)
			l := b.Match([]byte(tc.p), tc.flags)
			if l != tc.expectLen {
				t.Errorf("want: %d got: %d", tc.expectLen, l)
			}
		})
	}
}

func FuzzMatch(f *testing.F) {
	for _, tc := range matchTestcases {
		f.Add([]byte(tc.b), []byte(tc.p), int(tc.flags))
	}
	f.Fuzz(func(t *testing.T, b, p []byte, flags int) {
		Bytes(b).Match(p, Flags(flags)%CompactWS|IgnoreCase|FullWord)
	})
}

func BenchmarkMatch(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	randData := make([]byte, 1024)
	if _, err := io.ReadFull(r, randData); err != io.ErrUnexpectedEOF && err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for _, f := range []Flags{
		0,
		CompactWS,
		IgnoreCase,
		FullWord,
	} {
		b.Run(fmt.Sprintf("%d", f), func(b *testing.B) {
			for b.Loop() {
				Bytes(randData).Match(randData, f)
			}
		})
	}
}
