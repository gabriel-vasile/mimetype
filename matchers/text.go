package matchers

import (
	"bytes"
	"encoding/json"
)

type (
	markupSig  []byte
	ciSig      []byte // case insensitive signature
	shebangSig []byte
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
	svgSigs = []sig{
		markupSig("<SVG "),
	}
	x3dSigs = []sig{
		markupSig("<X3D "),
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

func Html(in []byte) bool {
	return detect(in, htmlSigs)
}

func Xml(in []byte) bool {
	return detect(in, xmlSigs)
}

func Svg(in []byte) bool {
	return detect(in, svgSigs)
}

func X3d(in []byte) bool {
	return detect(in, x3dSigs)
}

func Php(in []byte) bool {
	return detect(in, phpSigs)
}

func Json(in []byte) bool {
	return json.Valid(in)
}
func Js(in []byte) bool {
	return detect(in, jsSigs)
}

func Lua(in []byte) bool {
	return detect(in, luaSigs)
}

func Perl(in []byte) bool {
	return detect(in, perlSigs)
}

func Python(in []byte) bool {
	return detect(in, pythonSigs)
}
func (hSig markupSig) detect(in []byte) bool {
	if len(in) < len(hSig)+1 {
		return false
	}

	match := true

	for i, b := range hSig {
		db := in[i]
		if 'A' <= b && b <= 'Z' {
			db &= 0xDF
		}
		if b != db {
			match = false
			break
		}
	}

	if match == false {
		indice := getIndexBreakLine(in)
		for i, b := range hSig {
			if indice+i >= len(in) {
				return false
			}
			db := in[indice+i]
			if 'A' <= b && b <= 'Z' {
				db &= 0xDF
			}
			if b != db {
				return false
			}
		}
	}
	// Next byte must be space or right angle bracket.
	if db := in[len(hSig)]; db != ' ' && db != '>' {
		return false
	}

	return true
}

func (tSig ciSig) detect(in []byte) bool {
	if len(in) < len(tSig)+1 {
		return false
	}

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

func getIndexBreakLine(in []byte) int {
	for i, b := range in {
		if b == '\n' {
			return i + 1
		}
	}
	return 0
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

func Rtf(in []byte) bool {
	return bytes.Equal(in[:6], []byte("\x7b\x5c\x72\x74\x66\x31"))
}
