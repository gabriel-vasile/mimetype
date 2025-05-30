package charset

import (
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

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
