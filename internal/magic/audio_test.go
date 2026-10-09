package magic

import (
	"testing"
)

func synchsafeBytes(size uint32) [4]byte {
	return [4]byte{
		byte((size >> 21) & 0x7F),
		byte((size >> 14) & 0x7F),
		byte((size >> 7) & 0x7F),
		byte(size & 0x7F),
	}
}

func makeID3v2Header(version byte, revision byte, flags byte, size uint32) []byte {
	sizeBytes := synchsafeBytes(size)
	return []byte{
		'I', 'D', '3',
		version,
		revision,
		flags,
		sizeBytes[0],
		sizeBytes[1],
		sizeBytes[2],
		sizeBytes[3],
	}
}

func TestID3v2(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		expected bool
	}{
		{
			name:     "valid v2.2 small tag",
			header:   makeID3v2Header(2, 0, 0, 1024),
			expected: true,
		},
		{
			name:     "valid v2.3 small tag",
			header:   makeID3v2Header(3, 0, 0, 1024),
			expected: true,
		},
		{
			name:     "valid v2.4 small tag",
			header:   makeID3v2Header(4, 0, 0, 1024),
			expected: true,
		},
		{
			name:     "valid v2.4 minimal tag size 1",
			header:   makeID3v2Header(4, 0, 0, 1),
			expected: true,
		},
		{
			name:     "valid v2.3 tag below 10MB boundary",
			header:   makeID3v2Header(3, 0, 0, 10*1024*1024-1),
			expected: true,
		},
		{
			name:     "valid v2.3 tag exactly 10MB",
			header:   makeID3v2Header(3, 0, 0, 10*1024*1024),
			expected: true,
		},
		{
			name:     "valid v2.3 tag exceeding 10MB by one byte",
			header:   makeID3v2Header(3, 0, 0, 10*1024*1024+1),
			expected: true,
		},
		{
			name:     "valid v2.3 tag ~13MB (issue 839 repro)",
			header:   makeID3v2Header(3, 0, 0, 13*1024*1024),
			expected: true,
		},
		{
			name:     "valid v2.4 tag 50MB",
			header:   makeID3v2Header(4, 0, 0, 50*1024*1024),
			expected: true,
		},
		{
			name:     "valid v2.4 tag 100MB",
			header:   makeID3v2Header(4, 0, 0, 100*1024*1024),
			expected: true,
		},
		{
			name:     "valid v2.4 maximum 28-bit synchsafe tag size",
			header:   makeID3v2Header(4, 0, 0, 0x0FFFFFFF),
			expected: true,
		},
		{
			name:     "valid v2.3 with allowed flags (upper 4 bits)",
			header:   makeID3v2Header(3, 0, 0b11100000, 1024),
			expected: true,
		},
		{
			name:     "empty header slice",
			header:   []byte{},
			expected: false,
		},
		{
			name:     "header shorter than 10 bytes",
			header:   []byte{'I', 'D', '3', 3, 0},
			expected: false,
		},
		{
			name:     "wrong magic prefix",
			header:   []byte{'I', 'D', '4', 3, 0, 0, 0, 0, 1, 0},
			expected: false,
		},
		{
			name:     "unsupported version v2.1",
			header:   makeID3v2Header(1, 0, 0, 1024),
			expected: false,
		},
		{
			name:     "unsupported version v2.5",
			header:   makeID3v2Header(5, 0, 0, 1024),
			expected: false,
		},
		{
			name:     "non-zero revision",
			header:   makeID3v2Header(3, 1, 0, 1024),
			expected: false,
		},
		{
			name:     "invalid flags lower 4 bits non-zero",
			header:   makeID3v2Header(3, 0, 0b00000001, 1024),
			expected: false,
		},
		{
			name:     "tag size zero",
			header:   makeID3v2Header(3, 0, 0, 0),
			expected: false,
		},
		{
			name: "non-synchsafe size byte 0 (bit 7 set)",
			header: []byte{
				'I', 'D', '3', 3, 0, 0,
				0x80, 0x00, 0x00, 0x01,
			},
			expected: false,
		},
		{
			name: "non-synchsafe size byte 1 (bit 7 set)",
			header: []byte{
				'I', 'D', '3', 3, 0, 0,
				0x00, 0x80, 0x00, 0x01,
			},
			expected: false,
		},
		{
			name: "non-synchsafe size byte 2 (bit 7 set)",
			header: []byte{
				'I', 'D', '3', 3, 0, 0,
				0x00, 0x00, 0x80, 0x01,
			},
			expected: false,
		},
		{
			name: "non-synchsafe size byte 3 (bit 7 set)",
			header: []byte{
				'I', 'D', '3', 3, 0, 0,
				0x00, 0x00, 0x00, 0x80,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := id3v2(tt.header)
			if got != tt.expected {
				t.Errorf("id3v2(%v) = %v, want %v", tt.header, got, tt.expected)
			}
		})
	}
}

func TestMP3WithID3v2(t *testing.T) {
	// A 13MB tag header followed by payload data.
	header := makeID3v2Header(3, 0, 0, 13*1024*1024)
	payload := append(header, make([]byte, 1024)...)

	if !MP3(payload, 0) {
		t.Fatal("MP3 with ID3v2 tag > 10MB should be recognized as MP3")
	}
}
