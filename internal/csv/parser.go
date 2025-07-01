package csv

import (
	"bytes"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// Parser is a CSV reader that only counts fields.
// It avoids allocating/copying memory and to verify behaviour, it is tested
// and fuzzed against encoding/csv parser.
type Parser struct {
	comma   byte
	comment byte
	s       scan.Bytes
}

func NewParser(comma, comment byte, s scan.Bytes) *Parser {
	return &Parser{
		comma:   comma,
		comment: comment,
		s:       s,
	}
}

func (r *Parser) readLine() []byte {
	line := r.s.ReadSlice('\n')
	n := len(line)
	if n > 0 && line[n-1] == '\r' {
		return line[:n-1] // drop \r at end of line
	}

	// Normalize \r\n to \n on all input lines.
	if n := len(line); n >= 2 && line[n-2] == '\r' && line[n-1] == '\n' {
		line[n-2] = '\n'
		return line[:n-1]
	}
	return line
}

// CountFields reads one CSV line and counts how many records that line contained.
// hasMore reports whether there are more lines in the input.
func (r *Parser) CountFields(collectIndexes bool) (fields int, fieldPos []int, hasMore bool) {
	finished := false
	var line scan.Bytes
	for {
		line = r.readLine()
		if finished {
			return 0, nil, false
		}
		finished = len(r.s) == 0 && len(line) == 0
		if len(line) == lengthNL(line) {
			line = nil
			continue // Skip empty lines.
		}
		if len(line) > 0 && line[0] == r.comment {
			line = nil
			continue
		}
		break
	}

	indexes := []int{}
	originalLine := line
parseField:
	for {
		if len(line) == 0 || line[0] != '"' { // non-quoted string field
			fields++
			if collectIndexes {
				indexes = append(indexes, len(originalLine)-len(line))
			}
			i := bytes.IndexByte(line, r.comma)
			if i >= 0 {
				line.Advance(i + 1) // 1 to get over ending comma
				continue parseField
			}
			break parseField
		} else { // Quoted string field.
			if collectIndexes {
				indexes = append(indexes, len(originalLine)-len(line))
			}
			line.Advance(1) // get over starting quote
			for {
				i := bytes.IndexByte(line, '"')
				if i >= 0 {
					line.Advance(i + 1) // 1 for ending quote
					switch rn := line.Peek(); {
					case rn == '"':
						line.Advance(1)
					case rn == r.comma:
						line.Advance(1)
						fields++
						continue parseField
					case lengthNL(line) == len(line):
						fields++
						break parseField
					}
				} else if len(line) > 0 {
					line = r.readLine()
					originalLine = line
				} else {
					fields++
					break parseField
				}
			}
		}
	}

	return fields, indexes, fields != 0
}

// lengthNL reports the number of bytes for the trailing \n.
func lengthNL(b []byte) int {
	if len(b) > 0 && b[len(b)-1] == '\n' {
		return 1
	}
	return 0
}
