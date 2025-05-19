// Package scan has functions for scanning byte slices.
package scan

import (
	"bytes"
	"unicode/utf8"
)

// Bytes is a byte slice with helper methods for easier scanning.
type Bytes []byte

func (b *Bytes) Advance(n int) bool {
	if n < 0 || len(*b) < n {
		return false
	}
	*b = (*b)[n:]
	return true
}

// TrimLWS trims whitespace from beginning of the bytes.
func (b *Bytes) TrimLWS() {
	firstNonWS := 0
	for ; firstNonWS < len(*b) && ByteIsWS((*b)[firstNonWS]); firstNonWS++ {
	}

	*b = (*b)[firstNonWS:]
}

// TrimRWS trims whitespace from the end of the bytes.
func (b *Bytes) TrimRWS() {
	lb := len(*b)
	for lb > 0 && ByteIsWS((*b)[lb-1]) {
		*b = (*b)[:lb-1]
		lb--
	}
}

func (b *Bytes) Peek() byte {
	if len(*b) > 0 {
		return (*b)[0]
	}
	return 0
}
func (b *Bytes) Pop() byte {
	if len(*b) > 0 {
		ret := (*b)[0]
		*b = (*b)[1:]
		return ret
	}
	return 0
}

func (b *Bytes) PeekRune() rune {
	r, _ := utf8.DecodeRune(*b)
	return r
}

// PopUntil will advance b until, but not including, the first occurence of stopAt
// character. If no occurence is found, then it will advance until the end of b.
// The returned Bytes is a slice of all the bytes that we're advanced over.
func (b *Bytes) PopUntil(stopAt ...byte) Bytes {
	if len(*b) == 0 {
		return Bytes{}
	}
	i := bytes.IndexAny(*b, string(stopAt))
	if i == -1 {
		i = len(*b)
	}

	prefix := (*b)[:i]
	*b = (*b)[i:]
	return Bytes(prefix)
}

// Is will return true if all bytes in b are one of the allowed bytes.
func (b *Bytes) Is(allowed []byte) bool {
	for _, c := range *b {
		if bytes.IndexByte(allowed, c) == -1 {
			return false
		}
	}
	return true
}

// Line returns the first line from b and advances b with the length of the
// line. One new line character is trimmed after the line if it exists.
func (b *Bytes) Line() Bytes {
	line := b.PopUntil('\n')
	lline := len(line)
	if lline > 0 && line[lline-1] == '\r' {
		line = line[:lline-1]
	}
	b.Advance(1)
	return line
}

// DropLastLine drops the last incomplete line from b.
//
// mimetype limits itself to ReadLimit bytes when performing a detection.
// This means, for file formats like CSV for NDJSON, the last line of the input
// can be an incomplete line.
// If b length is less than readLimit, it means we received an incomplete file
// and proceed with dropping the last line.
func (b *Bytes) DropLastLine(readLimit uint32) {
	if readLimit == 0 || uint32(len(*b)) < readLimit {
		return
	}

	for i := len(*b) - 1; i > 0; i-- {
		if (*b)[i] == '\n' {
			*b = (*b)[:i]
			return
		}
	}
}

func ByteIsWS(b byte) bool {
	return b == '\t' || b == '\n' || b == '\x0c' || b == '\r' || b == ' '
}

var (
	ASCIISpaces = []byte{' ', '\r', '\n', '\x0c', '\t'}
	ASCIIDigits = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
)
