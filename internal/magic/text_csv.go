package magic

import (
	"bytes"
	"io"
)

const (
	svLineLimit = 10
	quote       = '"'
	comment     = '#'
)

// Csv matches a comma-separated values file.
func Csv(raw []byte, limit uint32) bool {
	return sv(raw, ',', limit)
}

// Tsv matches a tab-separated values file.
func Tsv(raw []byte, limit uint32) bool {
	return sv(raw, '\t', limit)
}

func sv(raw []byte, delimiter byte, limit uint32) bool {
	reader := prepSvReader(raw, limit)

	isWithinQuote := false
	isWithinComment := false
	lineIdx := 0
	recordFields := make(map[int]int)

	buf := make([]byte, 1024)
	n, err := reader.Read(buf)

	var prev, cur, next byte
loop:
	for err == nil {
		for i := 0; i < n; i++ {
			cur = buf[i]

			if i > 0 {
				prev = buf[i-1]
			} else {
				prev = byte(0)
			}

			if i < n-1 {
				next = buf[i+1]
			} else {
				next = byte(0)
			}

			isNewline := cur == '\n' && prev != '\r' && next != byte(0) && next != '\n' || cur == '\r'

			switch {
			case cur == quote:
				if (!isWithinQuote || next != quote) && !isWithinComment {
					isWithinQuote = !isWithinQuote
				} else {
					i++
				}

			case isNewline && !isWithinQuote:
				if lineIdx >= svLineLimit {
					break loop
				}
				_, ok := recordFields[lineIdx]
				if !isWithinComment && !ok {
					// this should have been a csv line, but we saw content without a delimiter that was not in a comment
					return false
				}
				lineIdx++
				isWithinComment = false

			case !isWithinQuote && !isWithinComment:
				switch cur {
				case comment:
					isWithinComment = true

				case delimiter:
					if recordFields[lineIdx] == 0 {
						recordFields[lineIdx] = 1
					}
					recordFields[lineIdx]++
				}
			}

		}

		n, err = reader.Read(buf)
	}

	var fieldCount int
	for _, fields := range recordFields {
		if fields > 0 {
			fieldCount = fields
			break
		}
	}

	var badFieldCount bool
	for _, fields := range recordFields {
		if fields != fieldCount {
			badFieldCount = true
			break
		}
	}

	return !badFieldCount && fieldCount > 1 && lineIdx > 0
}

func prepSvReader(in []byte, limit uint32) io.Reader {
	var reader io.Reader = bytes.NewReader(dropLastLine(in, limit))
	if limit > 0 {
		reader = io.LimitReader(reader, int64(limit))
	}

	return reader
}

// dropLastLine drops the last incomplete line from b.
//
// mimetype limits itself to ReadLimit bytes when performing a detection.
// This means, for file formats like CSV for NDJSON, the last line of the input
// can be an incomplete line.
func dropLastLine(b []byte, readLimit uint32) []byte {
	if readLimit == 0 || uint32(len(b)) < readLimit {
		return b
	}
	for i := len(b) - 1; i > 0; i-- {
		if b[i] == '\n' {
			return b[:i]
		}
	}
	return b
}
