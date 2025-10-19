// Package magic holds the matching functions used to find MIME types.
package magic

import (
	"bytes"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// File represents the input file and all the details we gathered about it.
// Facts about the input are shared between detection functions through the File
// struct.
type File struct {
	// Head is the first bytes from the file (up to readLimit length)
	Head scan.Bytes

	ReadLimit      uint32
	geo, har, gltf bool
}

// Detector functions exist for each file format. They report whether the file
// looks like a specific format. Additionally, they collect file facts and
// cache them in f for other Detectors to use.
type Detector func(f *File) bool
type xmlSig struct {
	// the local name of the root tag
	localName []byte
	// the namespace of the XML document
	xmlns []byte
}

// offset returns true if the provided signature can be
// found at offset in the raw input.
func offset(raw []byte, sig []byte, offset int) bool {
	return len(raw) > offset && bytes.HasPrefix(raw[offset:], sig)
}

// ciPrefix is like prefix but the check is case insensitive.
func ciPrefix(raw []byte, sigs ...[]byte) bool {
	for _, s := range sigs {
		if ciCheck(s, raw) {
			return true
		}
	}
	return false
}
func ciCheck(sig, raw []byte) bool {
	if len(raw) < len(sig)+1 {
		return false
	}
	// perform case insensitive check
	for i, b := range sig {
		db := raw[i]
		if 'A' <= b && b <= 'Z' {
			db &= 0xDF
		}
		if b != db {
			return false
		}
	}

	return true
}

// xml returns true if any of the provided XML signatures matches the raw input.
func xml(b scan.Bytes, sigs ...xmlSig) bool {
	b.TrimLWS()
	if len(b) == 0 {
		return false
	}
	for _, s := range sigs {
		if xmlCheck(s, b) {
			return true
		}
	}
	return false
}
func xmlCheck(sig xmlSig, raw []byte) bool {
	raw = raw[:min(len(raw), 512)]

	if len(sig.localName) == 0 {
		return bytes.Index(raw, sig.xmlns) > 0
	}
	if len(sig.xmlns) == 0 {
		return bytes.Index(raw, sig.localName) > 0
	}

	localNameIndex := bytes.Index(raw, sig.localName)
	return localNameIndex != -1 && localNameIndex < bytes.Index(raw, sig.xmlns)
}

// markup returns true is any of the HTML signatures matches the raw input.
func markup(b scan.Bytes, sigs ...[]byte) bool {
	if bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}) {
		// We skip the UTF-8 BOM if present to ensure we correctly
		// process any leading whitespace. The presence of the BOM
		// is taken into account during charset detection in charset.go.
		b.Advance(3)
	}
	b.TrimLWS()
	if len(b) == 0 {
		return false
	}
	for _, s := range sigs {
		if markupCheck(s, b) {
			return true
		}
	}
	return false
}
func markupCheck(sig, raw []byte) bool {
	if len(raw) < len(sig)+1 {
		return false
	}

	// perform case insensitive check
	for i, b := range sig {
		db := raw[i]
		if 'A' <= b && b <= 'Z' {
			db &= 0xDF
		}
		if b != db {
			return false
		}
	}
	// Next byte must be space or right angle bracket.
	if db := raw[len(sig)]; !scan.ByteIsWS(db) && db != '>' {
		return false
	}

	return true
}

// ftyp returns true if any of the FTYP signatures matches the raw input.
func ftyp(raw []byte, sigs ...[]byte) bool {
	if len(raw) < 12 {
		return false
	}
	for _, s := range sigs {
		if bytes.Equal(raw[8:12], s) {
			return true
		}
	}
	return false
}

// A valid shebang starts with the "#!" characters,
// followed by any number of spaces,
// followed by the path to the interpreter,
// and, optionally, followed by the arguments for the interpreter.
//
// Ex:
//
//	#! /usr/bin/env php
//
// /usr/bin/env is the interpreter, php is the first and only argument.
func shebang(b scan.Bytes, matchFlags scan.Flags, sigs ...[]byte) bool {
	line := b.Line()
	if len(line) < 2 || line[0] != '#' || line[1] != '!' {
		return false
	}
	line = line[2:]
	line.TrimLWS()
	for _, s := range sigs {
		if line.Match(s, matchFlags) != -1 {
			return true
		}
	}
	return false
}
