package charset

import (
	"reflect"
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

var getAnAttributeTestCases = []struct {
	in      string
	name    string
	value   string
	hasMore bool
}{{
	"", "", "", false,
}, {
	"''", "", "", false,
}, {
	`""`, "", "", false,
}, {
	`"abc`, "", "", false,
}, {
	"1>", "1", "", false,
}, {
	"A>", "a", "", false,
}, {
	"a>", "a", "", false,
}, {
	"abc>", "abc", "", false,
}, {
	"'abc'", "", "", false,
}, {
	"'abc'>", "'abc'", "", false,
}, {
	// > as attribute ender
	"meta1=meta>", "meta1", "meta", false,
}, {
	"meta2=META>", "meta2", "META", false,
}, {
	`meta3="meta">`, "meta3", "meta", false,
}, {
	`meta4="'meta">`, "meta4", "'meta", false,
}, {
	" meta5 = meta >", "meta5", "meta", true,
}, {
	" meta6 =' meta '>", "meta6", " meta ", false,
}, {
	` meta7 =' "meta '>`, "meta7", ` "meta `, false,
	// / as attribute ender
}, {
	// when the value is unquoted / right after is a parse warning
	"meta1=meta/", "meta1", "", false,
}, {
	"meta2=META/", "meta2", "", false,
}, {
	"meta3=meta /", "meta3", "meta", true,
}, {
	"meta4=META /", "meta4", "META", true,
}, {
	`meta5="meta"/`, "meta5", "meta", true,
}, {
	`meta6="'meta"/`, "meta6", "'meta", true,
}, {
	" meta7 = meta /", "meta7", "meta", true,
}, {
	" meta8 =' meta '/", "meta8", " meta ", true,
}, {
	` meta9  =' "meta '/`, "meta9", ` "meta `, true,
}, {
	`  meta0 /`, "meta0", ``, true,
}, {
	"; charset=UTF-8", ";", "", true,
}, {
	` http-equiv="content-type" content="text/html; charset=iso-8859-15">`, "http-equiv", `content-type`, true,
}}

func TestGetAnAttribute(t *testing.T) {
	for _, tc := range getAnAttributeTestCases {
		t.Run(tc.in, func(t *testing.T) {
			s := scan.Bytes(tc.in)
			name, value, hasMore := getAnAttribute(&s)
			if string(name) != tc.name {
				t.Errorf("name: got: %s, want: %s", name, tc.name)
			}
			if string(value) != tc.value {
				t.Errorf("value: got: %s, want: %s", value, tc.value)
			}
			if hasMore != tc.hasMore {
				t.Errorf("hasMore: got: %t, want: %t", hasMore, tc.hasMore)
			}
		})
	}
}
func FuzzGetAnAttribute(f *testing.F) {
	for _, t := range getAnAttributeTestCases {
		f.Add([]byte(t.in))
	}

	f.Fuzz(func(t *testing.T, d []byte) {
		s := scan.Bytes(d)
		getAnAttribute(&s)
	})
}

var getAValueTestCases = []struct {
	in      string
	out     string
	hasMore bool
}{{
	"", "", false,
}, {
	"   ", "", false,
}, {
	"''", "", false,
}, {
	`""`, "", false,
}, {
	`"abc`, "", false,
}, {
	">", "", false,
}, {
	"1>", "1", false,
}, {
	"A>", "A", false,
}, {
	"a>", "a", false,
}, {
	"abc>", "abc", false,
}, {
	"ABCXYZ>", "ABCXYZ", false,
}, {
	"'abc'", "abc", false,
}, {
	"'abc'>", "abc", false,
}, {
	"abc def=ghi", "abc", true,
}, {
	"abc >", "abc", true,
}, {
	"'abc' >", "abc", true,
}, {
	"'ABCXYZ' >", "ABCXYZ", true,
}, {
	`"abc" >`, "abc", true,
}}

func TestGetAValue(t *testing.T) {
	for _, tc := range getAValueTestCases {
		t.Run(tc.in, func(t *testing.T) {
			s := scan.Bytes(tc.in)
			got, hasMore := getAValue(&s)
			if string(got) != tc.out {
				t.Errorf("got: %s, want: %s", got, tc.out)
			}
			if hasMore != tc.hasMore {
				t.Errorf("hasMore: got: %t, want: %t", hasMore, tc.hasMore)
			}
		})
	}
}
func FuzzGetAValue(f *testing.F) {
	for _, tc := range getAValueTestCases {
		f.Add([]byte(tc.in))
	}

	f.Fuzz(func(t *testing.T, d []byte) {
		s := scan.Bytes(d)
		getAValue(&s)
	})
}

var extractCharsetFromMetaTestCases = []struct {
	in  string
	out string
}{{
	"", "",
}, {
	"''", "",
}, {
	`""`, "",
}, {
	`charset`, "",
}, {
	`charset=`, "",
}, {
	`charset="`, "",
}, {
	`charset=""`, "",
}, {
	`charset="a"`, "a",
}, {
	`charset="'a'"`, "'a'",
}, {
	`charset = a`, "a",
}, {
	`charset = a;`, "a",
}}

func TestExtractCharsetFromMeta(t *testing.T) {
	for _, tc := range extractCharsetFromMetaTestCases {
		t.Run(tc.in, func(t *testing.T) {
			got := extractCharsetFromMeta(scan.Bytes(tc.in))
			if string(got) != tc.out {
				t.Errorf("got: %s, want: %s", got, tc.out)
			}
		})
	}
}
func FuzzExtractCharsetFromMeta(f *testing.F) {
	for _, tc := range extractCharsetFromMetaTestCases {
		f.Add([]byte(tc.in))
	}

	f.Fuzz(func(t *testing.T, d []byte) {
		extractCharsetFromMeta(d)
	})
}

var fromHTMLTestCases = []struct {
	in  string
	out string
}{{
	"", "",
}, {
	"<!--> ", "",
}, {
	"<not-meta", "",
}, {
	"<meta", "",
}, {
	"<meta=", "",
}, {
	"<meta ", "",
}, {
	`<meta content="text/html; charset=iso-8859-15">`, "",
}, {
	`<meta http-equiv="content-type">`, "",
}, {
	`<meta content="text/html; charset=iso-8859-15" http-equiv="content-type" >`, "iso-8859-15",
}, {
	`<meta http-equiv="content-type" content="a/b; charset=щ">`, "щ",
}, {
	`<f 1=2 /><meta b="b" charset="щ">`, "щ",
}}

func TestFromHTML(t *testing.T) {
	for _, tc := range fromHTMLTestCases {
		t.Run(tc.in, func(t *testing.T) {
			got := fromHTML([]byte(tc.in))
			if string(got) != tc.out {
				t.Errorf("got: %s, want: %s", got, tc.out)
			}
		})
	}
}
func FuzzFromHTML(f *testing.F) {
	for _, tc := range fromHTMLTestCases {
		f.Add([]byte(tc.in))
	}

	f.Fuzz(func(t *testing.T, d []byte) {
		fromHTML(d)
	})
}

var fromXMLTestCases = []struct {
	in  string
	out string
}{{
	"", "",
}, {
	"   not <?xml start ", "",
}, {
	"   not <?xml start ", "",
}, {
	`<?xml version="1.0" encoding=c ?>`, "c",
}, {
	`<?xml version="1.0" encoding="c"?>`, "c",
}, {
	`  <?xml   version  =  "1.0"  encoding  =  "c"?>`, "c",
}}

func TestFromXML(t *testing.T) {
	for _, tc := range fromXMLTestCases {
		t.Run(tc.in, func(t *testing.T) {
			got := fromXML([]byte(tc.in))
			if string(got) != tc.out {
				t.Errorf("got: %s, want: %s", got, tc.out)
			}
		})
	}
}
func FuzzFromXML(f *testing.F) {
	for _, s := range fromXMLTestCases {
		f.Add([]byte(s.in))
	}

	f.Fuzz(func(t *testing.T, d []byte) {
		if charset := FromXML(d); charset == "" {
			t.Skip()
		}
	})
}

func TestFromPlain(t *testing.T) {
	tcases := []struct {
		raw     []byte
		charset string
	}{
		{[]byte{0xe6, 0xf8, 0xe5, 0x85, 0x85}, "windows-1252"},
		{[]byte{0xe6, 0xf8, 0xe5}, "iso-8859-1"},
		{[]byte("æøå"), "utf-8"},
		{[]byte{}, ""},
	}
	for _, tc := range tcases {
		if cs := FromPlain(tc.raw); cs != tc.charset {
			t.Errorf("in: %v; expected: %s; got: %s", tc.raw, tc.charset, cs)
		}
	}
}

func FuzzFromPlain(f *testing.F) {
	samples := [][]byte{
		{0xe6, 0xf8, 0xe5, 0x85, 0x85},
		{0xe6, 0xf8, 0xe5},
		[]byte("æøå"),
	}

	for _, s := range samples {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, d []byte) {
		if charset := FromPlain(d); charset == "" {
			t.Skip()
		}
	})
}

func TestGetAllAttributes(t *testing.T) {
	tcases := []struct {
		in       string
		expected [][2]string
	}{{
		"", [][2]string{},
	}, {
		// doesn't have ending >
		"a", [][2]string{},
	}, {
		// doesn't have ending >
		"abc", [][2]string{},
	}, {
		"a b c", [][2]string{{"a", ""}, {"b", ""}},
	}, {
		"abc abc abc", [][2]string{{"abc", ""}, {"abc", ""}},
	}, {
		"a=1 b=2 c=3", [][2]string{{"a", "1"}, {"b", "2"}, {"c", ""}},
	}, {
		"a=1 b c=3", [][2]string{{"a", "1"}, {"b", ""}, {"c", ""}},
	}, {
		"a b=2 c", [][2]string{{"a", ""}, {"b", "2"}},
	}, {
		">", [][2]string{},
	}, {
		"a>", [][2]string{{"a", ""}},
	}, {
		"abc>", [][2]string{{"abc", ""}},
	}, {
		"a b c>", [][2]string{{"a", ""}, {"b", ""}, {"c", ""}},
	}, {
		"a b/ c>", [][2]string{{"a", ""}, {"b", ""}, {"c", ""}},
	}, {
		"/a b/ c>", [][2]string{{"a", ""}, {"b", ""}, {"c", ""}},
	}, {
		"a b abc/>", [][2]string{{"a", ""}, {"b", ""}, {"abc", ""}},
	}}

	getAll := func(in string) [][2]string {
		s := scan.Bytes(in)
		ret := [][2]string{}
		for {
			name, value, _ := getAnAttribute(&s)
			if name == "" {
				return ret
			}
			ret = append(ret, [2]string{name, value})
		}
	}

	for _, tc := range tcases {
		t.Run(tc.in, func(t *testing.T) {
			got := getAll(tc.in)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("got: %v, want: %v", got, tc.expected)
			}
		})
	}
}

const xmlDoc = `<?xml version="1.0" encoding="UTF-8"?>
<note>
  <to>Tove</to>
  <from>Jani</from>
  <heading>Reminder</heading>
  <body>Don't forget me this weekend!</body>
</note>`
const htmlDoc = `<!DOCTYPE html>
<html>
  <head><!--[if lt IE 9]><script language="javascript" type="text/javascript" src="//html5shim.googlecode.com/svn/trunk/html5.js"></script><![endif]-->
    <meta charset="UTF-8"><style>/*
     </style>
    <link rel="stylesheet" href="css/animation.css"><!--[if IE 7]><link rel="stylesheet" href="css/" + font.fontname + "-ie7.css"><![endif]-->
    <script>
    </script>
  </head>
  <body>
    <div class="container footer">さ</div>
  </body>
</html>`

func BenchmarkFromHTML(b *testing.B) {
	b.ReportAllocs()
	doc := []byte(htmlDoc)
	for i := 0; i < b.N; i++ {
		FromHTML(doc)
	}
}
func BenchmarkFromXML(b *testing.B) {
	b.ReportAllocs()
	doc := []byte(xmlDoc)
	for i := 0; i < b.N; i++ {
		FromXML(doc)
	}
}
func BenchmarkFromPlain(b *testing.B) {
	b.ReportAllocs()
	doc := []byte(xmlDoc)
	for i := 0; i < b.N; i++ {
		FromPlain(doc)
	}
}
