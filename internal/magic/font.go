package magic

import (
	"bytes"
)

// Woff matches a Web Open Font Format file.
func Woff(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("wOFF"))
}

// Woff2 matches a Web Open Font Format version 2 file.
func Woff2(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("wOF2"))
}

// Otf matches an OpenType font file.
func Otf(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x4F, 0x54, 0x54, 0x4F, 0x00})
}

// Ttf matches a TrueType font file.
func Ttf(f *File) bool {
	if !bytes.HasPrefix(f.Head, []byte{0x00, 0x01, 0x00, 0x00}) {
		return false
	}
	return !MsAccessAce(f) && !MsAccessMdb(f)
}

// Eot matches an Embedded OpenType font file.
func Eot(f *File) bool {
	return len(f.Head) > 35 &&
		bytes.Equal(f.Head[34:36], []byte{0x4C, 0x50}) &&
		(bytes.Equal(f.Head[8:11], []byte{0x02, 0x00, 0x01}) ||
			bytes.Equal(f.Head[8:11], []byte{0x01, 0x00, 0x00}) ||
			bytes.Equal(f.Head[8:11], []byte{0x02, 0x00, 0x02}))
}

// Ttc matches a TrueType Collection font file.
func Ttc(f *File) bool {
	return len(f.Head) > 7 &&
		bytes.HasPrefix(f.Head, []byte("ttcf")) &&
		(bytes.Equal(f.Head[4:8], []byte{0x00, 0x01, 0x00, 0x00}) ||
			bytes.Equal(f.Head[4:8], []byte{0x00, 0x02, 0x00, 0x00}))
}
