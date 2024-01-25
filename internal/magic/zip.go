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

var zipHeader = []byte("PK\u0003\u0004")

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
	i  int // current index
}

// next returns the next file name from the zip headers.
// https://web.archive.org/web/20191129114319/https://users.cs.jmu.edu/buchhofp/forensics/formats/pkzip.html
func (t *zipTokenizer) next() (fileName string) {
	if len(t.in)-t.i < 30 {
		return ""
	}

	in := t.in[t.i:]

	offset := 0

	buf := in[offset : offset+4]
	offset += 4

	if !bytes.Equal(buf, zipHeader) {
		i := bytes.IndexByte(buf, zipHeader[0])
		t.i += offset
		if i > 0 {
			t.i += len(zipHeader) - i
		}
		return t.next()
	}

	offset += 14

	buf = in[offset : offset+4]
	offset += 4
	compressedSize := binary.LittleEndian.Uint32(buf)

	offset += 4
	buf = in[offset : offset+2]
	offset += 2
	fileNameLength := binary.LittleEndian.Uint16(buf)

	buf = in[offset : offset+2]
	offset += 2
	extraFieldsLength := binary.LittleEndian.Uint16(buf)

	buf = in[offset : offset+int(fileNameLength)]
	offset += int(fileNameLength)

	offset += int(extraFieldsLength) + int(compressedSize)

	t.i += offset

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
