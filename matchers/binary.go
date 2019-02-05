package matchers

import "bytes"

// Class matches an java class file.
func Class(in []byte) bool {
	return bytes.Equal(in[:4], []byte{0xCA, 0xFE, 0xBA, 0xBE})
}

// Swf matches an Adobe Flash swf file.
func Swf(in []byte) bool {
	return bytes.Equal(in[:3], []byte("CWS")) ||
		bytes.Equal(in[:3], []byte("FWS")) ||
		bytes.Equal(in[:3], []byte("ZWS"))
}
