package charset

import (
	"bytes"
	"encoding/xml"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

const (
	f = 0 /* character never appears in text */
	t = 1 /* character appears in plain ASCII text */
	i = 2 /* character appears in ISO-8859 text */
	x = 3 /* character appears in non-ISO extended ASCII (Mac, IBM PC) */
)

var (
	boms = []struct {
		bom []byte
		enc string
	}{
		{[]byte{0xEF, 0xBB, 0xBF}, "utf-8"},
		{[]byte{0x00, 0x00, 0xFE, 0xFF}, "utf-32be"},
		{[]byte{0xFF, 0xFE, 0x00, 0x00}, "utf-32le"},
		{[]byte{0xFE, 0xFF}, "utf-16be"},
		{[]byte{0xFF, 0xFE}, "utf-16le"},
	}

	// https://github.com/file/file/blob/fa93fb9f7d21935f1c7644c47d2975d31f12b812/src/encoding.c#L241
	textChars = [256]byte{
		/*                  BEL BS HT LF VT FF CR    */
		f, f, f, f, f, f, f, t, t, t, t, t, t, t, f, f, /* 0x0X */
		/*                              ESC          */
		f, f, f, f, f, f, f, f, f, f, f, t, f, f, f, f, /* 0x1X */
		t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, /* 0x2X */
		t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, /* 0x3X */
		t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, /* 0x4X */
		t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, /* 0x5X */
		t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, /* 0x6X */
		t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, f, /* 0x7X */
		/*            NEL                            */
		x, x, x, x, x, t, x, x, x, x, x, x, x, x, x, x, /* 0x8X */
		x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, x, /* 0x9X */
		i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, /* 0xaX */
		i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, /* 0xbX */
		i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, /* 0xcX */
		i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, /* 0xdX */
		i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, /* 0xeX */
		i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, /* 0xfX */
	}
)

func FromBOM(content []byte) string {
	for _, b := range boms {
		if bytes.HasPrefix(content, b.bom) {
			return b.enc
		}
	}
	return ""
}

func FromPlain(content []byte) string {
	if len(content) == 0 {
		return ""
	}
	if cset := FromBOM(content); cset != "" {
		return cset
	}
	origContent := content
	// Try to detect UTF-8.
	// First eliminate any partial rune at the end.
	for i := len(content) - 1; i >= 0 && i > len(content)-4; i-- {
		b := content[i]
		if b < 0x80 {
			break
		}
		if utf8.RuneStart(b) {
			content = content[:i]
			break
		}
	}
	hasHighBit := false
	for _, c := range content {
		if c >= 0x80 {
			hasHighBit = true
			break
		}
	}
	if hasHighBit && utf8.Valid(content) {
		return "utf-8"
	}

	// ASCII is a subset of UTF8. Follow W3C recommendation and replace with UTF8.
	if ascii(origContent) {
		return "utf-8"
	}

	return latin(origContent)
}

func latin(content []byte) string {
	hasControlBytes := false
	for _, b := range content {
		tt := textChars[b]
		if tt != t && tt != i {
			return ""
		}
		if b >= 0x80 && b <= 0x9F {
			hasControlBytes = true
		}
	}
	// Code range 0x80 to 0x9F is reserved for control characters in ISO-8859-1
	// (so-called C1 Controls). Windows 1252, however, has printable punctuation
	// characters in this range.
	if hasControlBytes {
		return "windows-1252"
	}
	return "iso-8859-1"
}

func ascii(content []byte) bool {
	for _, b := range content {
		if textChars[b] != T {
			return false
		}
	}
	return true
}

func FromXML(content []byte) string {
	if cset := fromXML(content); cset != "" {
		return cset
	}
	return FromPlain(content)
}
func fromXML(content []byte) string {
	content = trimLWS(content)
	dec := xml.NewDecoder(bytes.NewReader(content))
	rawT, err := dec.RawToken()
	if err != nil {
		return ""
	}

	t, ok := rawT.(xml.ProcInst)
	if !ok {
		return ""
	}

	return strings.ToLower(xmlEncoding(string(t.Inst)))
}

func FromHTML(content []byte) string {
	if cset := fromHTML(content); cset != "" {
		return cset
	}
	return FromPlain(content)
}

func fromHTML(content []byte) string {
	z := html.NewTokenizer(bytes.NewReader(content))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return ""

		case html.StartTagToken, html.SelfClosingTagToken:
			tagName, hasAttr := z.TagName()
			if !bytes.Equal(tagName, []byte("meta")) {
				continue
			}
			attrList := make(map[string]bool)
			gotPragma := false

			const (
				dontKnow = iota
				doNeedPragma
				doNotNeedPragma
			)
			needPragma := dontKnow

			name := ""
			for hasAttr {
				var key, val []byte
				key, val, hasAttr = z.TagAttr()
				ks := string(key)
				if attrList[ks] {
					continue
				}
				attrList[ks] = true
				for i, c := range val {
					if 'A' <= c && c <= 'Z' {
						val[i] = c + 0x20
					}
				}

				switch ks {
				case "http-equiv":
					if bytes.Equal(val, []byte("content-type")) {
						gotPragma = true
					}

				case "content":
					name = fromMetaElement(string(val))
					if name != "" {
						needPragma = doNeedPragma
					}

				case "charset":
					name = string(val)
					needPragma = doNotNeedPragma
				}
			}

			if needPragma == dontKnow || needPragma == doNeedPragma && !gotPragma {
				continue
			}

			if strings.HasPrefix(name, "utf-16") {
				name = "utf-8"
			}

			return name
		}
	}
}

func fromMetaElement(s string) string {
	for s != "" {
		csLoc := strings.Index(s, "charset")
		if csLoc == -1 {
			return ""
		}
		s = s[csLoc+len("charset"):]
		s = strings.TrimLeft(s, " \t\n\f\r")
		if !strings.HasPrefix(s, "=") {
			continue
		}
		s = s[1:]
		s = strings.TrimLeft(s, " \t\n\f\r")
		if s == "" {
			return ""
		}
		if q := s[0]; q == '"' || q == '\'' {
			s = s[1:]
			closeQuote := strings.IndexRune(s, rune(q))
			if closeQuote == -1 {
				return ""
			}
			return s[:closeQuote]
		}

		end := strings.IndexAny(s, "; \t\n\f\r")
		if end == -1 {
			end = len(s)
		}
		return s[:end]
	}
	return ""
}

func xmlEncoding(s string) string {
	param := "encoding="
	idx := strings.Index(s, param)
	if idx == -1 {
		return ""
	}
	v := s[idx+len(param):]
	if v == "" {
		return ""
	}
	if v[0] != '\'' && v[0] != '"' {
		return ""
	}
	idx = strings.IndexRune(v[1:], rune(v[0]))
	if idx == -1 {
		return ""
	}
	return v[1 : idx+1]
}

// trimLWS trims whitespace from beginning of the input.
// TODO: find a way to call trimLWS once per detection instead of once in each
// detector which needs the trimmed input.
func trimLWS(in []byte) []byte {
	firstNonWS := 0
	for ; firstNonWS < len(in) && isWS(in[firstNonWS]); firstNonWS++ {
	}

	return in[firstNonWS:]
}

func isWS(b byte) bool {
	return b == '\t' || b == '\n' || b == '\x0c' || b == '\r' || b == ' '
}
