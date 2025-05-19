package csv

import (
	"bytes"
	"fmt"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

type Reader struct {
	Comma   byte
	Comment byte
	S       scan.Bytes
}

func (r *Reader) ReadLine() (fields int, hasMore bool) {
	finished := false
	var line scan.Bytes
	for {
		line = r.S.Line()
		if finished {
			break
		}
		if len(r.S) == 0 {
			finished = true
		}
		if len(line) == 0 {
			continue
		}
		if line[0] == r.Comment {
			continue
		}
		break
	}

	run := 0
parseField:
	for {
		run++
		fmt.Println("run", run, fields, string(line))
		for scan.ByteIsWS(line.Peek()) {
			line.Advance(1)
		}
		if len(line) == 0 {
			return fields, !finished
		}
		if len(line) == 0 || line[0] != '"' { // Non-quoted string field
			i := bytes.IndexByte(line, r.Comma)
			if i >= 0 {
				line.Advance(i)
				line.Advance(1) // get over ending comma
				fields++
				continue parseField
			} else {
				line.Advance(len(line) - lengthNL(line))
			}
			fields++
			break parseField
		} else { // Quoted string field.
			line.Advance(1) // get over starting comma
			for {
				i := bytes.IndexByte(line, '"')
				if i >= 0 {
					line.Advance(i + 1) // 1 for comma
					switch rn := line.Peek(); {
					case rn == '"':
						line.Advance(1)
					case rn == r.Comma:
						line.Advance(1)
						fields++
						continue parseField
					case lengthNL(line) == len(line):
						fields++
						break parseField
					default:
						break parseField
					}
				} else if len(line) > 0 {
					line = r.S.Line()
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
