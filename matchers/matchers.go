package matchers

import "bytes"

func Dummy(_ []byte) bool {
	return true
}

func Pdf(in []byte) bool {
	return bytes.Equal(in[:4], []byte{0x25, 0x50, 0x44, 0x46})
}

func trimWS(in []byte) []byte {
	firstNonWS := 0
	for ; firstNonWS < len(in) && isWS(in[firstNonWS]); firstNonWS++ {
	}
	return in[firstNonWS:]
}

func isWS(b byte) bool {
	switch b {
	case '\t', '\n', '\x0c', '\r', ' ':
		return true
	}
	return false
}
