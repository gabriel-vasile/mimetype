package matchers

import "bytes"

// Aaf matches an Advanced Authoring Format file.
// See: https://www.digipres.org/formats/sources/fdd/formats/#fdd000004
// See: https://tech.ebu.ch/docs/techreview/trev_291-gilmer.pdf
func Aaf(in []byte) bool {
	return bytes.HasPrefix(in, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1, 0x41, 0x41, 0x46, 0x42, 0x0D, 0x00, 0x4F, 0x4D})
}