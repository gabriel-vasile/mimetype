package matchers

import (
	"bytes"

	"github.com/gabriel-vasile/mimetype/matchers/json"
)

type (
	markupSig  []byte
	ciSig      []byte // case insensitive signature
	shebangSig []byte // matches !# followed by the signature
	sig        interface {
		detect([]byte) bool
	}
)

var (
	htmlSigs = []sig{
		markupSig("<!DOCTYPE HTML"),
		markupSig("<HTML"),
		markupSig("<HEAD"),
		markupSig("<SCRIPT"),
		markupSig("<IFRAME"),
		markupSig("<H1"),
		markupSig("<DIV"),
		markupSig("<FONT"),
		markupSig("<TABLE"),
		markupSig("<A"),
		markupSig("<STYLE"),
		markupSig("<TITLE"),
		markupSig("<B"),
		markupSig("<BODY"),
		markupSig("<BR"),
		markupSig("<P"),
		markupSig("<!--"),
	}
	xmlSigs = []sig{
		markupSig("<?XML"),
	}
	phpSigs = []sig{
		ciSig("<?PHP"),
		ciSig("<?\n"),
		ciSig("<?\r"),
		ciSig("<? "),
		shebangSig("/usr/local/bin/php"),
		shebangSig("/usr/bin/php"),
		shebangSig("/usr/bin/env php"),
	}
	jsSigs = []sig{
		shebangSig("/bin/node"),
		shebangSig("/usr/bin/node"),
		shebangSig("/bin/nodejs"),
		shebangSig("/usr/bin/nodejs"),
		shebangSig("/usr/bin/env node"),
		shebangSig("/usr/bin/env nodejs"),
	}
	luaSigs = []sig{
		shebangSig("/usr/bin/lua"),
		shebangSig("/usr/local/bin/lua"),
		shebangSig("/usr/bin/env lua"),
	}
	perlSigs = []sig{
		shebangSig("/usr/bin/perl"),
		shebangSig("/usr/bin/env perl"),
		shebangSig("/usr/bin/env perl"),
	}
	pythonSigs = []sig{
		shebangSig("/usr/bin/python"),
		shebangSig("/usr/local/bin/python"),
		shebangSig("/usr/bin/env python"),
		shebangSig("/usr/bin/env python"),
	}
)

// Txt matches a text file.
func Txt(in []byte) bool {
	in = trimLWS(in)
	for _, b := range in {
		if b <= 0x08 ||
			b == 0x0B ||
			0x0E <= b && b <= 0x1A ||
			0x1C <= b && b <= 0x1F {
			return false
		}
	}

	return true
}

func detect(in []byte, sigs []sig) bool {
	for _, sig := range sigs {
		if sig.detect(in) {
			return true
		}
	}

	return false
}

// Html matches a Hypertext Markup Language file.
func Html(in []byte) bool {
	return detect(in, htmlSigs)
}

// Xml matches an Extensible Markup Language file.
func Xml(in []byte) bool {
	return detect(in, xmlSigs)
}

// Php matches a PHP: Hypertext Preprocessor file.
func Php(in []byte) bool {
	return detect(in, phpSigs)
}

// Json matches a JavaScript Object Notation file.
func Json(in []byte) bool {
	parsed, err := json.Scan(in)
	if len(in) < ReadLimit {
		return err == nil
	}

	return parsed == len(in)
}

// Js matches a Javascript file.
func Js(in []byte) bool {
	return detect(in, jsSigs)
}

// Lua matches a Lua programming language file.
func Lua(in []byte) bool {
	return detect(in, luaSigs)
}

// Perl matches a Perl programming language file.
func Perl(in []byte) bool {
	return detect(in, perlSigs)
}

// Python matches a Python programming language file.
func Python(in []byte) bool {
	return detect(in, pythonSigs)
}

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

// Rtf matches a Rich Text Format file.
func Rtf(in []byte) bool {
	return bytes.Equal(in[:6], []byte("\x7b\x5c\x72\x74\x66\x31"))
}
