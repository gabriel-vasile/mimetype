package magic

import (
	"bytes"
	"encoding/binary"
	"strings"
)

var (
	// Odt matches an OpenDocument Text file.
	Odt = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.text"), 30)
	// Ott matches an OpenDocument Text Template file.
	Ott = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.text-template"), 30)
	// Ods matches an OpenDocument Spreadsheet file.
	Ods = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.spreadsheet"), 30)
	// Ots matches an OpenDocument Spreadsheet Template file.
	Ots = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.spreadsheet-template"), 30)
	// Odp matches an OpenDocument Presentation file.
	Odp = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.presentation"), 30)
	// Otp matches an OpenDocument Presentation Template file.
	Otp = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.presentation-template"), 30)
	// Odg matches an OpenDocument Drawing file.
	Odg = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.graphics"), 30)
	// Otg matches an OpenDocument Drawing Template file.
	Otg = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.graphics-template"), 30)
	// Odf matches an OpenDocument Formula file.
	Odf = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.formula"), 30)
	// Odc matches an OpenDocument Chart file.
	Odc = offset([]byte("mimetypeapplication/vnd.oasis.opendocument.chart"), 30)
	// Epub matches an EPUB file.
	Epub = offset([]byte("mimetypeapplication/epub+zip"), 30)
	// Sxc matches an OpenOffice Spreadsheet file.
	Sxc = offset([]byte("mimetypeapplication/vnd.sun.xml.calc"), 30)
)

var (
	zipHeaderSig         = []byte("PK\u0003\u0004")
	zipFileDescriptorSig = []byte("PK\u0007\u0008")
	zipDirectorySig      = []byte("PK\u0001\u0002")
)

// Zip matches a zip archive.
func Zip(raw []byte, limit uint32) bool {
	return len(raw) > 3 &&
		raw[0] == 0x50 && raw[1] == 0x4B &&
		(raw[2] == 0x3 || raw[2] == 0x5 || raw[2] == 0x7) &&
		(raw[3] == 0x4 || raw[3] == 0x6 || raw[3] == 0x8)
}

// Jar matches a Java archive file.
func Jar(raw []byte, limit uint32) bool {
	return zipContains(raw, "META-INF/MANIFEST.MF")
}

// zipTokenizer holds the source zip file and scanned index.
type zipTokenizer struct {
	in []byte
}

// read checks if in has a size of at least n and then returns
// a slice of in[:n] while setting the head of in to head + n.
//
// If length of in is smaller than n, nil is returned.
func (t *zipTokenizer) read(n int) []byte {
	if n == 0 || len(t.in) < n {
		return nil
	}

	buf := t.in[:n]
	t.in = t.in[n:]

	return buf
}

// next returns the next file name from the zip headers.
// https://web.archive.org/web/20191129114319/https://users.cs.jmu.edu/buchhofp/forensics/formats/pkzip.html
func (t *zipTokenizer) next() string {
	// When the rest length is smaller than the minimum header size, exit.
	if len(t.in) < 30 {
		return ""
	}

	buf := t.in[:4]

	// If central directory signature is found, exit.
	if bytes.Equal(buf, zipDirectorySig) {
		return ""
	}

	// Looking for the file header signature. If it is not at the start
	// of buf, then look for it inside in.
	if !bytes.Equal(buf, zipHeaderSig) {
		i := bytes.Index(t.in, zipHeaderSig)
		if i < 0 {
			return ""
		}
		t.in = t.in[i:]
	}

	// skip header + version
	t.read(4 + 2)

	// read general purpose bit field
	buf = t.read(2)
	if buf == nil {
		return ""
	}
	flags := binary.LittleEndian.Uint16(buf)
	fdFlag := int(flags)&0x08 != 0

	// skip compression method, last modified time and date and crc32
	t.read(10)

	buf = t.read(4)
	if buf == nil {
		return ""
	}
	compressedSize := binary.LittleEndian.Uint32(buf)

	// skip uncompressed size
	t.read(4)

	buf = t.read(2)
	if buf == nil {
		return ""
	}
	fileNameLength := binary.LittleEndian.Uint16(buf)

	buf = t.read(2)
	if buf == nil {
		return ""
	}
	extraFieldsLength := binary.LittleEndian.Uint16(buf)

	buf = t.read(int(fileNameLength))
	if buf == nil {
		return ""
	}

	// skip extra fields and compressed data
	t.read(int(extraFieldsLength) + int(compressedSize))

	// If the file descriptor flag is set, search for the next occurrence
	// of the file descriptor signature. If found, skip the field.
	// Otherwise, look for the next occurrence of the file header by calling
	// next recursively.
	if fdFlag {
		i := bytes.Index(t.in, zipFileDescriptorSig)
		if i < 0 {
			return t.next()
		}

		t.in = t.in[i:]
		t.read(16)
	}

	return string(buf)
}

// zipContains returns true if the zip file headers from in contain any of the paths.
func zipContains(in []byte, paths ...string) bool {
	t := zipTokenizer{in: in}
	for i, tok := 0, t.next(); tok != ""; i, tok = i+1, t.next() {
		for p := range paths {
			if strings.HasPrefix(tok, paths[p]) {
				return true
			}
		}
	}

	return false
}
