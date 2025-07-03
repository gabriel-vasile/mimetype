package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

type line struct {
	fields int
	// indexes[i] says at which index in the line the i-th field starts at.
	indexes []int
	hasMore bool
}

var testcases = []struct {
	name    string
	csv     string
	comma   byte
	comment byte
}{{
	"empty", "", ',', '#',
}, {
	"simple",
	`foo,bar,baz
1,2,3
"1","a",b`,
	',', '#',
}, {
	"crlf line endings",
	"foo,bar,baz\r\n1,2,3\r\n",
	',', '#',
}, {
	"leading and trailing space",
	`1, abc ,3`,
	',', '#',
}, {
	"empty quote",
	`1,"",3`,
	',', '#',
}, {
	"quotes with comma",
	`1,",",3`,
	',', '#',
}, {
	"quotes with quote",
	`1,""",3`,
	',', '#',
}, {
	"fewer fields",
	`foo,bar,baz
1,2`,
	',', '#',
}, {
	"more fields",
	`1,2,3,4`,
	',', '#',
}, {
	"forgot quote",
	`1,"Forgot,3`,
	',', '#',
}, {
	"unescaped quote",
	`1,"abc"def",3`,
	',', '#',
}, {
	"unescaped quote",
	`1,"abc"def",3`,
	',', '#',
}, {
	"unescaped quote2",
	`1,abc"quote"def,3`,
	',', '#',
}, {
	"escaped quote",
	`1,abc""def,3`,
	',', '#',
}, {
	"new line",
	`1,abc
def,3`,
	',', '#',
}, {
	"new line quotes",
	`1,"abc
def",3`,
	',', '#',
}, {
	"quoted field at end",
	`1,"abc"`,
	',', '#',
}, {
	"not ended quoted field at end",
	`1,"abc`,
	',', '#',
}, {
	"empty field",
	`1,,3`,
	',', '#',
}, {
	"unicode fields",
	`💁,👌,🎍,😍`,
	',', '#',
}, {
	"comment",
	`#comment`,
	',', '#',
}, {
	"line with \\r at the end",
	"123\r\n456\r",
	',', '#',
}, {
	`from fuzz \"\"\r\n0`,
	"\"\"\r\n0",
	',', '\x11',
}}

// Test our parser against the one from encoding/csv.
func TestParser(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expected, recs, _ := stdlibLines(tc.csv, tc.comma, tc.comment)
			got := ourLines(tc.csv, tc.comma, tc.comment)
			if !reflect.DeepEqual(expected, got) {
				t.Errorf(`%s
expected: %v
     got: %v
 records: %v`, tc.csv, expected, got, recs)
			}
		})
	}
}

func ourLines(data string, comma, comment byte) []line {
	p := NewParser(comma, comment, scan.Bytes(data))
	lines := []line{}
	for {
		fields, indexes, hasMore := p.CountFields(true)
		if !hasMore {
			break
		}
		lines = append(lines, line{fields, indexes, hasMore})
	}
	return lines
}

// stdlibLines returns the []line records obtained using the stdlib CSV parser.
func stdlibLines(data string, comma, comment byte) ([]line, [][]string, error) {
	if comma > unicode.MaxASCII || comment > unicode.MaxASCII {
		return nil, nil, fmt.Errorf("comma or comment not ASCII")
	}

	if strings.IndexByte(data, 0) != -1 {
		return nil, nil, fmt.Errorf("CSV contains null byte 0x00")
	}
	r := csv.NewReader(strings.NewReader(data))
	r.Comma = rune(comma)
	r.ReuseRecord = true
	r.FieldsPerRecord = -1 // we don't care about lines having same number of fields
	r.LazyQuotes = true
	r.Comment = rune(comment)

	var err error
	lines := []line{}
	// To ease debugging, we keep records to print in tests.
	records := [][]string{}
	for {
		l, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		indexes := []int{}
		for i := 0; i < len(l); i++ {
			_, c := r.FieldPos(i)
			// FieldPos starts counting from 1, but our parser counts from 0.
			// Adjust -1 so tests match.
			indexes = append(indexes, c-1)
		}
		lines = append(lines, line{len(l), indexes, err != io.EOF})
		records = append(records, l)
	}

	return lines, records, err
}

var sample = `
1,2,3
"a", "b", "c"
a,b,c` + "\r\n1,2,3\r\na,b,c\r"

func BenchmarkCSVStdlibDecoder(b *testing.B) {
	b.ReportAllocs()
	// Reuse a single reader to prevent allocs inside the benchmark function.
	r := strings.NewReader(sample)
	for i := 0; i < b.N; i++ {
		_, err := r.Seek(0, 0)
		if err != nil {
			b.Fatalf("reader cannot seek: %s", err)
		}
		d := csv.NewReader(r)
		d.ReuseRecord = true
		d.FieldsPerRecord = -1 // we don't care about lines having same number of fields
		d.LazyQuotes = true
		for {
			_, err := d.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				b.Fatalf("error parsing CSV: %s", err)
			}
		}
	}
}
func BenchmarkCSVOurParser(b *testing.B) {
	b.ReportAllocs()
	// Reuse a single reader to prevent allocs inside the benchmark function.
	r := scan.Bytes(sample)
	p := NewParser(',', '#', r)
	for i := 0; i < b.N; i++ {
		p.s = r
		for {
			_, _, hasMore := p.CountFields(false)
			if !hasMore {
				break
			}
		}
	}
}

func FuzzParser(f *testing.F) {
	for _, p := range testcases {
		f.Add(p.csv, byte(','), byte('#'))
	}
	f.Fuzz(func(t *testing.T, data string, comma, comment byte) {
		expected, _, err := stdlibLines(data, comma, comment)
		// The sddlib CSV parser can accept UTF8 runes for comma and comment.
		// Our parser does not need that functionality, so it returns different
		// results for UTF8 inputs. Skip fuzzing when the generated data is UTF8.
		if err != nil {
			t.Skipf("not testable: %v", err)
		}
		got := ourLines(data, comma, comment)
		if !reflect.DeepEqual(got, expected) {
			t.Logf("input: %v, comma: %c, comment: %c", data, comma, comment)
			t.Errorf(`
expected: %v,
     got: %v`, expected, got)
		}
	})
}
