package matchers

import "bytes"

type (
	markupSig  []byte
	ciSig      []byte // case insensitive signature
	shebangSig []byte // matches !# followed by the signature
	ftypSig    []byte // matches audio/video files. www.ftyps.com
	sig        interface {
		detect([]byte) bool
	}
)

// Implement sig interface.
func (hSig markupSig) detect(in []byte) bool {
	if len(in) < len(hSig)+1 {
		return false
	}

	// perform case insensitive check
	for i, b := range hSig {
		db := in[i]
		if 'A' <= b && b <= 'Z' {
			db &= 0xDF
		}
		if b != db {
			return false
		}
	}
	// Next byte must be space or right angle bracket.
	if db := in[len(hSig)]; db != ' ' && db != '>' {
		return false
	}

	return true
}

// Implement sig interface.
func (tSig ciSig) detect(in []byte) bool {
	if len(in) < len(tSig)+1 {
		return false
	}

	// perform case insensitive check
	for i, b := range tSig {
		db := in[i]
		if 'A' <= b && b <= 'Z' {
			db &= 0xDF
		}
		if b != db {
			return false
		}
	}

	return true
}

// a valid shebang starts with the "#!" characters
// followed by any number of spaces
// followed by the path to the interpreter and optionally, the args for the interpreter
func (sSig shebangSig) detect(in []byte) bool {
	in = firstLine(in)

	if len(in) < len(sSig)+2 {
		return false
	}
	if in[0] != '#' || in[1] != '!' {
		return false
	}

	in = trimLWS(trimRWS(in[2:]))

	return bytes.Equal(in, sSig)
}

// Implement sig interface.
func (fSig ftypSig) detect(in []byte) bool {
	return len(in) > 12 &&
		bytes.Equal(in[4:8], []byte("ftyp")) &&
		bytes.Equal(in[8:12], fSig)
}

func detect(in []byte, sigs []sig) bool {
	for _, sig := range sigs {
		if sig.detect(in) {
			return true
		}
	}

	return false
}
