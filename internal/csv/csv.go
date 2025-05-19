package csv

import (
	"github.com/gabriel-vasile/mimetype/internal/scan"
)

type Reader struct {
	Comma   byte
	Comment byte
	S       scan.Bytes
}

func (r *Reader) ReadLine() (fields int, hasMore bool) {
	// Step over whitespace and comments
	for {
		for scan.ByteIsWS(r.S.Peek()) {
			r.S.Advance(1)
		}
		if r.S.Peek() != r.Comment {
			break
		}
		r.S.PopUntil('\n')
	}

	nf := 0
	for {
		switch r.S.Pop() {
		case 0:
			return nf, false
		case '"':
			r.eatQuote()
		case r.Comma:
			nf++
		case '\n':
			return nf + 1, r.S.Peek() != 0
		default:
			continue
		}
	}
}

const (
	stateInField = iota
)

func (r *Reader) eatQuote() {
	quote := false
	for {
		c := r.S.Peek()
		if c == 0 {
			return
		}
		if c != '"' {
			if quote {
				return
			}
			r.S.Advance(1)
			continue
		}
		if quote {
			quote = false
			r.S.Advance(1)
			continue
		}
		quote = true
		r.S.Advance(1)
	}
}
