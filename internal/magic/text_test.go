package magic

import (
	"strings"
	"testing"
)

// Benchmark JSON inputs that can cause slow-downs.
func BenchmarkJSONPathological(b *testing.B) {
	const n = 1000
	hugeArray := []byte(
		strings.Repeat("[1,", n) +
			`2,3,"abc",true,false,null` +
			strings.Repeat("]", n))
	hugeObject := []byte(
		strings.Repeat(`{"a": 1, "b":`, n) +
			`{"c":[2,3,"abc",true,false,null]}` +
			strings.Repeat("}", n))

	b.ReportAllocs()
	for b.Loop() {
		if !JSON(hugeArray, 0) {
			b.Fatal("huge array should be JSON")
		}
		if !JSON(hugeObject, 0) {
			b.Fatal("huge object should be JSON")
		}
		GeoJSON(hugeArray, 0)
		GeoJSON(hugeObject, 0)
		HAR(hugeArray, 0)
		HAR(hugeObject, 0)
		GLTF(hugeArray, 0)
		GLTF(hugeObject, 0)
		NdJSON(hugeArray, 0)
		NdJSON(hugeObject, 0)
	}
}

func TestRFC822(t *testing.T) {
	testcases := []struct {
		name     string
		in       string
		expected bool
	}{{
		"empty", "", false,
	}, {
		"one hint", "Cc: cc@mail.com", false,
	}, {
		"two identical hints", "Cc: cc@mail.com\nCc: cc@mail.com", true,
	}, {
		"two different hints", "Cc: cc@mail.com\nTo: to@mail.com", true,
	}, {
		"junk at start", "junk\nCc: cc@mail.com\nTo: to@mail.com", false,
	}, {
		"junk later", "Cc: cc@mail.com\njunk To: to@mail.com", false,
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := RFC822([]byte(tc.in), 0)
			if tc.expected != got {
				t.Errorf("expected: %t, got: %t", tc.expected, got)
			}
		})
	}
}

func TestTextRTF(t *testing.T) {
	testcases := []struct {
		name     string
		in       []byte
		expected bool
	}{
		{
			name:     "valid rtf without trailing null",
			in:       []byte("{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Arial;}}Hello World}"),
			expected: true,
		},
		{
			name:     "valid rtf with single trailing null",
			in:       []byte("{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Arial;}}Hello World}\x00"),
			expected: true,
		},
		{
			name:     "valid rtf with multiple trailing nulls",
			in:       []byte("{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Arial;}}Hello World}\x00\x00\x00"),
			expected: true,
		},
		{
			name:     "valid rtf with trailing nulls and whitespace",
			in:       []byte("{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Arial;}}Hello World}\r\n\x00\x00 \t\r\n"),
			expected: true,
		},
		{
			name:     "valid rtf with trailing whitespace and null",
			in:       []byte("{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Arial;}}Hello World} \r\n\x00"),
			expected: true,
		},
		{
			name:     "corrupted rtf with null byte in middle",
			in:       []byte("{\\rtf1\x00\\ansi Hello World}"),
			expected: false,
		},
		{
			name:     "plain text with trailing null byte",
			in:       []byte("Hello World\x00"),
			expected: false,
		},
		{
			name:     "normal plain text",
			in:       []byte("Hello World"),
			expected: true,
		},
		{
			name:     "binary data",
			in:       []byte("\x00\x01\x02\x03"),
			expected: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := Text(tc.in, 0)
			if tc.expected != got {
				t.Errorf("expected: %t, got: %t", tc.expected, got)
			}
		})
	}
}

func TestRtf(t *testing.T) {
	testcases := []struct {
		name     string
		in       []byte
		expected bool
	}{
		{
			name:     "minimal rtf header",
			in:       []byte("{\\rtf"),
			expected: true,
		},
		{
			name:     "rtf version 1 header",
			in:       []byte("{\\rtf1"),
			expected: true,
		},
		{
			name:     "rtf with trailing null",
			in:       []byte("{\\rtf1\\ansi}\x00"),
			expected: true,
		},
		{
			name:     "missing opening brace",
			in:       []byte("\\rtf1"),
			expected: false,
		},
		{
			name:     "empty input",
			in:       []byte{},
			expected: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := Rtf(tc.in, 0)
			if tc.expected != got {
				t.Errorf("expected: %t, got: %t", tc.expected, got)
			}
		})
	}
}

