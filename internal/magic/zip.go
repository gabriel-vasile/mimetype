package magic

import (
	"bytes"
	"encoding/binary"

	"github.com/gabriel-vasile/mimetype/internal/scan"
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

// Zip matches a zip archive.
func Zip(raw []byte, limit uint32) bool {
	return len(raw) > 3 &&
		raw[0] == 0x50 && raw[1] == 0x4B &&
		(raw[2] == 0x3 || raw[2] == 0x5 || raw[2] == 0x7) &&
		(raw[3] == 0x4 || raw[3] == 0x6 || raw[3] == 0x8)
}

// Jar matches a Java archive file.
func Jar(raw []byte, limit uint32) bool {
	return zipContains(raw, []byte("META-INF/MANIFEST.MF"), false)
}

func zipContains(raw, sig []byte, msoCheck bool) bool {
	b := scan.Bytes(raw)
	pk := []byte("PK\003\004")
	if len(b) < 0x1E {
		return false
	}

	if !b.Advance(0x1E) {
		return false
	}
	if bytes.HasPrefix(b, sig) {
		return true
	}

	if msoCheck {
		skipFiles := [][]byte{
			[]byte("[Content_Types].xml"),
			[]byte("_rels/.rels"),
			[]byte("docProps"),
			[]byte("customXml"),
			[]byte("[trash]"),
		}

		hasSkipFile := false
		for _, sf := range skipFiles {
			if bytes.HasPrefix(b, sf) {
				hasSkipFile = true
				break
			}
		}
		if !hasSkipFile {
			return false
		}
	}

	searchOffset := binary.LittleEndian.Uint32(raw[18:]) + 49
	if !b.Advance(int(searchOffset)) {
		return false
	}

	nextHeader := bytes.Index(raw[searchOffset:], pk)
	if !b.Advance(nextHeader) {
		return false
	}
	if bytes.HasPrefix(b, sig) {
		return true
	}

	// Previously i was 4 at max, but #679 reported zip files where signatures
	// occur later than 4. Because mimetype only looks at the file header, this
	// for loop might as well be unbounded, ie: until the input bytes are all
	// consumed. But users can call SetLimit(0) to make mimetype analyze whole
	// files. So keep max 100 just in case. The reason I initially made it 4
	// was because FILE(1) had this limit.
	for i := 0; i < 100; i++ {
		if !b.Advance(0x1A) {
			return false
		}
		nextHeader = bytes.Index(b, pk)
		if nextHeader == -1 {
			return false
		}
		if !b.Advance(nextHeader + 0x1E) {
			return false
		}
		if bytes.HasPrefix(b, sig) {
			return true
		}
	}
	return false
}

// APK matches an Android Package Archive.
// The source of signatures is https://github.com/file/file/blob/1778642b8ba3d947a779a36fcd81f8e807220a19/magic/Magdir/archive#L1820-L1887
func APK(raw []byte, _ uint32) bool {
	apkSignatures := [][]byte{
		[]byte("AndroidManifest.xml"),
		[]byte("META-INF/com/android/build/gradle/app-metadata.properties"),
		[]byte("classes.dex"),
		[]byte("resources.arsc"),
		[]byte("res/drawable"),
	}
	for _, sig := range apkSignatures {
		if zipContains(raw, sig, false) {
			return true
		}
	}

	return false
}
