package matchers

import "bytes"

type htmlSignature []byte
type xmlSignature []byte

var htmlSignatures = []htmlSignature{
	htmlSignature("<!DOCTYPE HTML"),
	htmlSignature("<HTML"),
	htmlSignature("<HEAD"),
	htmlSignature("<SCRIPT"),
	htmlSignature("<IFRAME"),
	htmlSignature("<H1"),
	htmlSignature("<DIV"),
	htmlSignature("<FONT"),
	htmlSignature("<TABLE"),
	htmlSignature("<A"),
	htmlSignature("<STYLE"),
	htmlSignature("<TITLE"),
	htmlSignature("<B"),
	htmlSignature("<BODY"),
	htmlSignature("<BR"),
	htmlSignature("<P"),
	htmlSignature("<!--"),
}

func Txt(in []byte) bool {
	in = trimWS(in)
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

func Html(in []byte) bool {
	in = trimWS(in)

	for _, hSig := range htmlSignatures {
		if detectHtml(in, hSig) {
			return true
		}
	}

	return false
}

func Xml(in []byte) bool {
	in = trimWS(in)
	return bytes.HasPrefix(in, []byte("<?xml"))
}

func detectHtml(in []byte, hSig htmlSignature) bool {
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
