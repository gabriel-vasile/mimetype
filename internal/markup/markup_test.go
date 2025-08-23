package markup

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
			name, value, hasMore := GetAnAttribute(&s)
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
		GetAnAttribute(&s)
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
			name, value, _ := GetAnAttribute(&s)
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

func TestSkipAComment(t *testing.T) {
	tcases := []struct {
		in      string
		out     string
		skipped bool
	}{{
		"", "", false,
	}, {
		"abc", "abc", false,
	}, {
		"<!--", "<!--", false, // not ending comment
	}, {
		"<!-- abc -->", "", true, // regular comment
	}, {
		"<!-->", "", true, // the beginning and ending -- are the same chars
	}}
	for _, tc := range tcases {
		t.Run(tc.in, func(t *testing.T) {
			s := scan.Bytes(tc.in)
			skipped := SkipAComment(&s)
			if tc.skipped != skipped {
				t.Errorf("skipped got: %v, want: %v", skipped, tc.skipped)
			}
			if string(s) != tc.out {
				t.Errorf("got: %v, want: %v", string(s), tc.out)
			}
		})
	}
}
