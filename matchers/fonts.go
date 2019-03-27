package matchers

import "bytes"

func Woff(in []byte) bool {
	return bytes.Equal(in[:4], []byte("wOFF"))
}

func Woff2(in []byte) bool {
	return bytes.Equal(in[:4], []byte("wOF2"))
}
