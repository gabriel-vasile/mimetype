package magic

import (
	"bufio"
	"bytes"
)

// Stl matches a StereoLithography file.
// See more: https://docs.fileformat.com/cad/stl/
//           https://www.iana.org/assignments/media-types/model/stl
// STL is available in ASCII as well as Binary representations for compact file format.
func Stl(raw []byte, limit uint32) bool {
	// ASCII check.
	if bytes.HasPrefix(raw, []byte("solid")) {
		// If the full file content was provided, check end of file last line.
		if len(raw) < int(limit) {
			return bytes.Contains(lastNonWSLine(raw), []byte("endsolid"))
		}
		return true
	}

	// Binary check.
	return bytes.HasPrefix(raw, bytes.Repeat([]byte{0x20}, 80))
}

// Ply matches a Polygon File Format or the Stanford Triangle Format file.
// See more: https://www.loc.gov/preservation/digital/formats/fdd/fdd000501.shtml
// Ply is available in ASCII as well as Binary representations for compact file format.
func Ply(raw []byte, limit uint32) bool {
	s := bufio.NewScanner(bytes.NewReader(raw))

	// First line must be "ply".
	if !s.Scan() {
		return false
	}
	if s.Text() != "ply" {
		return false
	}

	// Second line declares the subtype.
	if !s.Scan() {
		return false
	}
	return s.Text() == "format ascii 1.0" ||
		s.Text() == "format binary_little_endian 1.0" ||
		s.Text() == "format binary_big_endian 1.0"
}
