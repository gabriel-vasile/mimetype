package magic

import "bytes"

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

// Zip matches a zip archive.
func Zip(raw []byte, limit uint32) bool {
	return len(raw) > 3 &&
		raw[0] == 0x50 && raw[1] == 0x4B &&
		(raw[2] == 0x3 || raw[2] == 0x5 || raw[2] == 0x7) &&
		(raw[3] == 0x4 || raw[3] == 0x6 || raw[3] == 0x8)
}

// Jar matches a Java archive file.
func Jar(raw []byte, limit uint32) bool {
	t := zipTokenizer{in: raw}
	for i, tok := 0, t.next(); i < 10 && tok != ""; i, tok = i+1, t.next() {
		if tok == "META-INF/MANIFEST.MF" {
			return true
		}
	}

	return false
}

func Epub2(raw []byte, limit uint32) bool {
	t := zipTokenizer{in: raw}
	i := 0
	for tok, data := t.nextFile(); i < 10 && tok != ""; tok, data = t.nextFile() {
		i += 1
		if tok == "mimetype" && string(bytes.TrimSpace(data)) == "application/epub+zip" {
			return true
		}
	}

	return false
}
