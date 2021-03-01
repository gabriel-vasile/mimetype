package matchers

import (
	"bytes"
)

// Ttf matches a TrueType font file.
func Ttf(in []byte, _ uint32) bool {
	return bytes.HasPrefix(in, []byte{0x00, 0x01, 0x00, 0x00}) &&
		!MsAccessAce(in, 0) && !MsAccessMdb(in, 0)
}

// Woff matches a Web Open Font Format file.
func Woff(in []byte, _ uint32) bool {
	return bytes.HasPrefix(in, []byte("wOFF"))
}

// Woff2 matches a Web Open Font Format version 2 file.
func Woff2(in []byte, _ uint32) bool {
	return bytes.HasPrefix(in, []byte("wOF2"))
}

// Otf matches an OpenType font file.
func Otf(in []byte, _ uint32) bool {
	return bytes.HasPrefix(in, []byte{0x4F, 0x54, 0x54, 0x4F, 0x00})
}

// Eot matches an Embedded OpenType font file.
func Eot(in []byte, _ uint32) bool {
	return len(in) > 35 &&
		bytes.Equal(in[34:36], []byte{0x4C, 0x50}) &&
		(bytes.Equal(in[8:11], []byte{0x02, 0x00, 0x01}) ||
			bytes.Equal(in[8:11], []byte{0x01, 0x00, 0x00}) ||
			bytes.Equal(in[8:11], []byte{0x02, 0x00, 0x02}))
}
