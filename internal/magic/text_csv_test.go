package magic

import (
	"strings"
	"testing"
)

func TestCSV(t *testing.T) {
	testcases := []struct {
		name     string
		in       string
		limit    uint32
		expected bool
	}{
		// Minimum valid CSV: 2 rows and 2 columns.
		{"two rows two cols", "a,b\nc,d", 0, true},
		{"two rows three cols", "a,b,c\n1,2,3", 0, true},

		// Too few rows.
		{"empty", "", 0, false},
		{"one row two cols", "a,b", 0, false},
		{"one row only newline", "\n", 0, false},

		// Too few columns.
		{"two rows one col each", "a\nb", 0, false},
		{"one col with newlines", "a\nb\nc", 0, false},

		// Inconsistent column counts.
		{"mismatched col count", "a,b\nc,d,e", 0, false},
		{"mismatched col count reversed", "a,b,c\nd,e", 0, false},
		{"first row ok third mismatches", "a,b\nc,d\ne", 0, false},

		// Quoted fields.
		{"quoted fields", `"a","b"` + "\n" + `"c","d"`, 0, true},
		{"quoted field with comma inside", `"a,x","b"` + "\n" + `"c,y","d"`, 0, true},
		{"quoted field with newline inside", "\"a\nb\",c\nd,e", 0, true},

		// CRLF line endings.
		{"crlf endings", "a,b\r\nc,d", 0, true},
		{"crlf one row", "a,b\r\n", 0, false},

		// Comment lines (# is the comment character) are skipped.
		{"comment then two rows", "# skip\na,b\nc,d", 0, true},
		{"comment then one row", "# skip\na,b", 0, false},
		{"all comments", "# a\n# b\n", 0, false},

		// Empty lines are skipped.
		{"empty lines between rows", "a,b\n\nc,d", 0, true},
		{"leading empty line", "\na,b\nc,d", 0, true},

		// Early return after 10 rows.
		{"ten rows returns true", strings.Repeat("a,b\n", 10), 0, true},

		// Early return after 10 rows ignores subsequent invalid input.
		{"ten rows returns true", strings.Repeat("a,b\n", 10) + "bad", 0, true},

		// Truncation: limit=0 means whole input should be parsed and correct.
		{"limit zero no drop", "a,b\nc,d\nbad", 0, false},

		// Truncation: last incomplete line is allowed to have different number of fields.
		{"truncated last line dropped", "a,b\nc,d\nbad", uint32(len("a,b\nc,d\nbad")), true},

		// Truncation: last line is dropped when len(data)==limit.
		{"gets truncated", "a,b\nc,d", uint32(len("a,b\nc,d")), true},

		// Truncated but the remaining data is still invalid.
		{"truncated still one row", "a,b\nbad", uint32(len("a,b\nbad")), false},

		// Limit set but data is smaller than limit (complete file): no drop.
		{"limit larger than data", "a,b\nc,d\nbad", uint32(len("a,b\nc,d\nbad")) + 1, false},

		// TSV via the CSV function should not match tab-separated data.
		{"tab separated not csv", "a\tb\nc\td", 0, false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := CSV([]byte(tc.in), tc.limit)
			if got != tc.expected {
				t.Errorf("CSV(%q, %d) = %v, want %v", tc.in, tc.limit, got, tc.expected)
			}
		})
	}
}
