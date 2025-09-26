package magic

import (
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

func TestShebangCheck(t *testing.T) {
	tests := []struct {
		name     string
		sig      []byte
		input    string
		expected bool
	}{
		// Valid shebangs
		{
			name:     "valid bash shebang",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bash",
			expected: true,
		},
		{
			name:     "valid bash shebang with spaces",
			sig:      []byte("/bin/bash"),
			input:    "#! /bin/bash",
			expected: true,
		},
		{
			name:     "valid bash shebang with multiple spaces",
			sig:      []byte("/bin/bash"),
			input:    "#!   /bin/bash",
			expected: true,
		},
		{
			name:     "valid bash shebang with tabs",
			sig:      []byte("/bin/bash"),
			input:    "#!\t/bin/bash",
			expected: true,
		},
		{
			name:     "valid bash shebang with mixed whitespace",
			sig:      []byte("/bin/bash"),
			input:    "#! \t /bin/bash",
			expected: true,
		},
		{
			name:     "valid bash shebang with trailing whitespace",
			sig:      []byte("/bin/bash"),
			input:    "#! /bin/bash \t ",
			expected: true,
		},
		{
			name:     "valid bash shebang with arguments",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bash -exu",
			expected: true,
		},
		{
			name:     "valid env/python shebang",
			sig:      []byte("/usr/bin/env python"),
			input:    "#!/usr/bin/env python",
			expected: true,
		},
		{
			name:     "valid env/python shebang with spaces",
			sig:      []byte("/usr/bin/env python"),
			input:    "#! /usr/bin/env python",
			expected: true,
		},
		{
			name:     "valid env -S/python shebang with arguments",
			sig:      []byte("/usr/bin/env -S python"),
			input:    "#!/usr/bin/env -S python -u",
			expected: true,
		},

		// Invalid shebangs
		{
			name:     "missing shebang prefix",
			sig:      []byte("/bin/bash"),
			input:    "/bin/bash",
			expected: false,
		},
		{
			name:     "wrong shebang prefix",
			sig:      []byte("/bin/bash"),
			input:    "##!/bin/bash",
			expected: false,
		},
		{
			name:     "wrong shebang prefix 2",
			sig:      []byte("/bin/bash"),
			input:    "!#/bin/bash",
			expected: false,
		},
		{
			name:     "wrong interpreter path",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/sh",
			expected: false,
		},
		{
			name:     "partial interpreter path",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bas",
			expected: false,
		},
		{
			name:     "extra characters before interpreter",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bashx",
			expected: false,
		},
		{
			name:     "extra characters after interpreter",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bashx",
			expected: false,
		},

		// Edge cases
		{
			name:     "empty input",
			sig:      []byte("/bin/bash"),
			input:    "",
			expected: false,
		},
		{
			name:     "too short input",
			sig:      []byte("/bin/bash"),
			input:    "#!",
			expected: false,
		},
		{
			name:     "just shebang prefix",
			sig:      []byte("/bin/bash"),
			input:    "#!",
			expected: false,
		},
		{
			name:     "shebang with only spaces",
			sig:      []byte("/bin/bash"),
			input:    "#!   ",
			expected: false,
		},
		{
			name:     "shebang with only tabs",
			sig:      []byte("/bin/bash"),
			input:    "#!\t\t",
			expected: false,
		},
		{
			name:     "empty signature",
			sig:      []byte(""),
			input:    "#!",
			expected: true,
		},
		{
			name:     "empty signature with spaces",
			sig:      []byte(""),
			input:    "#!   ",
			expected: true,
		},
		{
			name:     "signature longer than input",
			sig:      []byte("/very/long/path/to/interpreter"),
			input:    "#!/bin/bash",
			expected: false,
		},
		{
			name:     "case sensitivity test",
			sig:      []byte("/bin/bash"),
			input:    "#!/BIN/BASH",
			expected: false,
		},
		{
			name:     "case sensitivity test 2",
			sig:      []byte("/BIN/BASH"),
			input:    "#!/bin/bash",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := scan.Bytes([]byte(tt.input))
			line := raw.Line()
			result := shebangCheck(tt.sig, line)
			if result != tt.expected {
				t.Errorf("shebangCheck(%q, %q) = %v, want %v", tt.sig, tt.input, result, tt.expected)
			}
		})
	}
}
