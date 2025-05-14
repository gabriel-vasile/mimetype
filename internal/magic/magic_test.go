package magic

import (
	"strings"
	"testing"
)

func TestMagic(t *testing.T) {
	tCases := []struct {
		name     string
		detector Detector
		raw      string
		limit    uint32
		res      bool
	}{{
		name:     "incomplete JSON, limit 0",
		detector: JSON,
		raw:      `["an incomplete JSON array`,
		limit:    0,
		res:      false,
	}, {
		name:     "incomplete JSON, limit 10",
		detector: JSON,
		raw:      `["an incomplete JSON array`,
		limit:    10,
		res:      true,
	}, {
		name:     "basic JSON data type null",
		detector: JSON,
		raw:      `null`,
		limit:    10,
		res:      false,
	}, {
		name:     "basic JSON data type string",
		detector: JSON,
		raw:      `"abc"`,
		limit:    10,
		res:      false,
	}, {
		name:     "basic JSON data type integer",
		detector: JSON,
		raw:      `120`,
		limit:    10,
		res:      false,
	}, {
		name:     "basic JSON data type float",
		detector: JSON,
		raw:      `.120`,
		limit:    10,
		res:      false,
	}, {
		name:     "NdJSON with basic data types",
		detector: NdJSON,
		raw:      "1\nnull\n\"foo\"\n0.1",
		limit:    10,
		res:      false,
	}, {
		name:     "NdJSON with basic data types and empty object",
		detector: NdJSON,
		raw:      "1\n2\n3\n{}",
		limit:    10,
		res:      true,
	}, {
		name:     "NdJSON with empty objects types",
		detector: NdJSON,
		raw:      "{}\n{}\n{}",
		limit:    10,
		res:      true,
	}, {
		name:     "MachO class or Fat but last byte > \\x14",
		detector: MachO,
		raw:      "\xCA\xFE\xBA\xBE   \x15",
		res:      false,
	}, {
		name:     "MachO class or Fat and last byte < \\x14",
		detector: MachO,
		raw:      "\xCA\xFE\xBA\xBE   \x13",
		res:      true,
	}, {
		name:     "MachO BE Magic32",
		detector: MachO,
		raw:      "\xFE\xED\xFA\xCE",
		res:      true,
	}, {
		name:     "MachO LE Magic32",
		detector: MachO,
		raw:      "\xCE\xFA\xED\xFE",
		res:      true,
	}, {
		name:     "MachO BE Magic64",
		detector: MachO,
		raw:      "\xFE\xED\xFA\xCF",
		res:      true,
	}, {
		name:     "MachO LE Magic64",
		detector: MachO,
		raw:      "\xCF\xFA\xED\xFE",
		res:      true,
	}}
	for _, tt := range tCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.detector([]byte(tt.raw), tt.limit); got != tt.res {
				t.Errorf("expected: %t; got: %t", tt.res, got)
			}
			// Empty inputs should not pass as anything.
			if got := tt.detector(nil, 0); got != false {
				t.Errorf("empty input: expected: %t; got: %t", false, got)
			}
		})
	}
}

func BenchmarkSrt(b *testing.B) {
	const subtitle = `1
00:02:16,612 --> 00:02:19,376
Senator, we're making
our final approach into Coruscant.

`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Srt([]byte(subtitle), 0)
	}
}

func BenchmarkJSON(b *testing.B) {
	var sample = []byte("{" +
		// It's no problem to repeat the same keys. The parser does not mind.
		strings.Repeat(`"fruit": {"apple": [{"red": 1}]}, "sizes": ["Large", 10, {"size": "small"}], "color": "Red",`, 1000) +
		`"fruit": "Apple", "size": "Large", "color": "Red"}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !JSON(sample, 0) {
			b.Error("should always be true")
		}
	}
}
