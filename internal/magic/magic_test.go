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
		flags    scan.Flags
		expected bool
	}{
		// Valid shebangs
		{
			name:     "valid bash shebang",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bash",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid bash shebang with spaces",
			sig:      []byte("/bin/bash"),
			input:    "#! /bin/bash",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid bash shebang with multiple spaces",
			sig:      []byte("/bin/bash"),
			input:    "#!   /bin/bash",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid bash shebang with tabs",
			sig:      []byte("/bin/bash"),
			input:    "#!\t/bin/bash",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid bash shebang with mixed whitespace",
			sig:      []byte("/bin/bash"),
			input:    "#! \t /bin/bash",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid bash shebang with trailing whitespace",
			sig:      []byte("/bin/bash"),
			input:    "#! /bin/bash \t ",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid bash shebang with arguments",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bash -exu",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid env/python shebang",
			sig:      []byte("/usr/bin/env python"),
			input:    "#!/usr/bin/env python",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid env/python shebang with spaces",
			sig:      []byte("/usr/bin/env python"),
			input:    "#! /usr/bin/env python",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid env/python shebang with arguments",
			sig:      []byte("/usr/bin/env python"),
			input:    "#!/usr/bin/env python -u",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "valid env/python shebang with arguments and trailing ws",
			sig:      []byte("/usr/bin/env python"),
			input:    "#!/usr/bin/env python -u \n",
			flags:    scan.CompactWS,
			expected: true,
		},

		// Invalid shebangs
		{
			name:     "missing shebang prefix",
			sig:      []byte("/bin/bash"),
			input:    "/bin/bash",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "wrong shebang prefix",
			sig:      []byte("/bin/bash"),
			input:    "##!/bin/bash",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "wrong shebang prefix 2",
			sig:      []byte("/bin/bash"),
			input:    "!#/bin/bash",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "wrong interpreter path",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/sh",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "partial interpreter path",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bas",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "extra characters after interpreter",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bashx",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "extra characters after interpreter but FullWord",
			sig:      []byte("/bin/bash"),
			input:    "#!/bin/bashx",
			flags:    scan.CompactWS | scan.FullWord,
			expected: false,
		},
		{
			name:     "extra characters after env interpreter",
			sig:      []byte("/usr/bin/env bash"),
			input:    "#!/usr/bin/env bash123",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "extra characters after env interpreter but FullWord",
			sig:      []byte("/usr/bin/env bash"),
			input:    "#!/usr/bin/env bash123",
			flags:    scan.CompactWS | scan.FullWord,
			expected: false,
		},

		// Edge cases
		{
			name:     "empty input",
			sig:      []byte("/bin/bash"),
			input:    "",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "too short input",
			sig:      []byte("/bin/bash"),
			input:    "#!",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "just shebang prefix",
			sig:      []byte("/bin/bash"),
			input:    "#!",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "shebang with only spaces",
			sig:      []byte("/bin/bash"),
			input:    "#!   ",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "shebang with only tabs",
			sig:      []byte("/bin/bash"),
			input:    "#!\t\t",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "empty signature",
			sig:      []byte(""),
			input:    "#!",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "empty signature with spaces",
			sig:      []byte(""),
			input:    "#!   ",
			flags:    scan.CompactWS,
			expected: true,
		},
		{
			name:     "signature longer than input",
			sig:      []byte("/very/long/path/to/interpreter"),
			input:    "#!/bin/bash",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "case sensitivity test",
			sig:      []byte("/bin/bash"),
			input:    "#!/BIN/BASH",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "case sensitivity test 2",
			sig:      []byte("/BIN/BASH"),
			input:    "#!/bin/bash",
			flags:    scan.CompactWS,
			expected: false,
		},
		{
			name:     "case sensitivity test 2",
			sig:      []byte("/BIN/BASH"),
			input:    "#!/bin/bash",
			flags:    scan.CompactWS,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := shebang(tt.flags, tt.sig)
			result := d([]byte(tt.input), 0)
			if result != tt.expected {
				t.Errorf("shebang(%q, %q) = %v, want %v", tt.sig, tt.input, result, tt.expected)
			}
		})
	}
}
