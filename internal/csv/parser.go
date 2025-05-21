package csv

import (
	"bytes"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

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

func (r *Parser) ReadLine() (fields int, hasMore bool) {
	finished := false
	var line scan.Bytes
	for {
		line = r.s.Line()
		if finished {
			break
		}
		finished = len(r.s) == 0
		if len(line) == 0 {
			continue
		}
		if line[0] == r.comment {
			continue
		}
		break
	}

parseField:
	for {
		for scan.ByteIsWS(line.Peek()) {
			line.Advance(1)
		}
		if len(line) == 0 {
			return fields, !finished
		}
		if len(line) == 0 || line[0] != '"' { // non-quoted string field
			i := bytes.IndexByte(line, r.comma)
			fields++
			if i >= 0 {
				line.Advance(i)
				line.Advance(1) // get over ending comma
				continue parseField
			}
			break parseField
		} else { // Quoted string field.
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
					line = r.s.Line()
					finished = len(r.s) == 0
				} else {
					fields++
					break parseField
				}
			}
		}
	}

	return fields, !finished
}

// lengthNL reports the number of bytes for the trailing \n.
func lengthNL(b []byte) int {
	if len(b) > 0 && b[len(b)-1] == '\n' {
		return 1
	}
	return 0
}
