// Package matchers holds the matching functions used to find mime types
package matchers

func True(_ []byte) bool {
	return true
}
func False(_ []byte) bool {
	return false
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
