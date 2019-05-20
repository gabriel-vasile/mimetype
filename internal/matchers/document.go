package matchers

import "bytes"

// Pdf matches a Portable Document Format file.
func Pdf(in []byte) bool {
	return len(in) > 4 && bytes.Equal(in[:4], []byte{0x25, 0x50, 0x44, 0x46})
}

// DjVu matches a DjVu file
func DjVu(in []byte) bool {
	if !bytes.HasPrefix(in, []byte{0x41, 0x54, 0x26, 0x54, 0x46, 0x4F, 0x52, 0x4D}) {
		return false
	}
	if len(in) < 15 {
		return false
	}
	return bytes.HasPrefix(in[12:], []byte("DJVM")) || bytes.HasPrefix(in[12:], []byte("DJVU")) || bytes.HasPrefix(in[12:], []byte("DJVI")) || bytes.HasPrefix(in[12:], []byte("THUM"))
}
