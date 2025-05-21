package csv

import (
	"encoding/csv"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

type line struct {
	fields  int
	hasMore bool
}

var testcases = []struct {
	name string
	csv  string
}{{
	"empty", "",
}, {
	"simple",
	`foo,bar,baz
1,2,3
"1","a",b`,
}, {
	"crlf line endings",
	"foo,bar,baz\r\n1,2,3\r\n",
}, {
	"leading and trailing space",
	`1, abc ,3`,
}, {
	"empty quote",
	`1,"",3`,
}, {
	"quotes with comma",
	`1,",",3`,
}, {
	"quotes with quote",
	`1,""",3`,
}, {
	"fewer fields",
	`foo,bar,baz
1,2`,
}, {
	"more fields",
	`1,2,3,4`,
}, {
	"forgot quote",
	`1,"Forgot,3`,
}, {
	"unescaped quote",
	`1,"abc"def",3`,
}, {
	"unescaped quote",
	`1,"abc"def",3`,
}, {
	"unescaped quote2",
	`1,abc"quote"def,3`,
}, {
	"escaped quote",
	`1,abc""def,3`,
}, {
	"new line",
	`1,abc
def,3`,
}, {
	"new line quotes",
	`1,"abc
def",3`,
}, {
	"quoted field at end",
	`1,"abc"`,
}, {
	"not ended quoted field at end",
	`1,"abc`,
}, {
	"empty field",
	`1,,3`,
}, {
	"unicode fields",
	`💁,👌,🎍,😍`,
}, {
	"comment",
	`#comment`,
}}

func TestParser(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expected := stdlibLines(tc.csv)
			got := ourLines(tc.csv)
			if !reflect.DeepEqual(expected, got) {
				t.Errorf("\n%s\n expected: %v got: %v", tc.csv, expected, got)
			}
		})
	}
}

func ourLines(data string) []line {
	p := NewParser(',', '#', scan.Bytes(data))
	lines := []line{}
	for {
		fields, hasMore := p.ReadLine()
		lines = append(lines, line{fields, hasMore})
		if !hasMore {
			break
		}
	}
	return lines
}

func stdlibLines(data string) []line {
	r := csv.NewReader(strings.NewReader(data))
	r.Comma = ','
	r.ReuseRecord = true
	r.FieldsPerRecord = -1 // we don't care about lines having same number of fields
	r.LazyQuotes = true
	r.Comment = '#'

	lines := []line{}
	for {
		l, err := r.Read()
		if err == io.EOF {
			// Adjust for a difference between our parser and the stdlib one.
			// stdlib: returns EOF at an extra call to Read after there is no more input
			// our parser: returns hasMore=false when there is no more input
			if len(lines) > 0 {
				lines[len(lines)-1].hasMore = false
			}
			break
		}
		lines = append(lines, line{len(l), err != io.EOF})
	}
	if len(lines) == 0 {
		return []line{{}}
	}
	return lines
}

var sample = `
1,2,3
"a", "b", "c"
a,b,c`

func BenchmarkCSVStdlibDecoder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d := csv.NewReader(strings.NewReader(sample))
		for {
			_, err := d.Read()
			if err == io.EOF {
				break
			}
		}
	}
}
func BenchmarkCSVOurParser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := NewParser(',', '#', scan.Bytes(sample))
		for {
			_, hasMore := p.ReadLine()
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
		p := NewParser(comma, comment, scan.Bytes(data))
		for {
			_, hasMore := p.ReadLine()
			if !hasMore {
				break
			}
		}
	})
}
