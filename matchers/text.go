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
		ciSig("#!/bin/node"),
		ciSig("#!/usr/bin/node"),
		ciSig("#!/bin/nodejs"),
		ciSig("#!/usr/bin/nodejs"),
		ciSig("#!/usr/bin/env node"),
		ciSig("#!/usr/bin/env nodejs"),
	}
	luaSigs = []sig{
		ciSig("#!/usr/bin/lua"),
		ciSig("#!/usr/local/bin/lua"),
		ciSig("#!/usr/bin/env lua"),
	}
	perlSigs = []sig{
		ciSig("# !/usr/bin/perl"),
		ciSig("# !/usr/bin/env perl"),
		ciSig("# !/usr/bin/env perl"),
	}
	pythonSigs = []sig{
		ciSig("#!/usr/bin/python"),
		ciSig("#!/usr/local/bin/python"),
		ciSig("#!/usr/bin/env python"),
		ciSig("#!/usr/bin/env python"),
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
