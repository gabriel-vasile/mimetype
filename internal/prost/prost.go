package prost

import "bytes"

var asd = prefixer(nil)

func Shebang(raw []byte) bool {
	return shebangCheck([]byte("/usr/bin/lua"), firstLine(raw))
}

func shebangCheck(sig, raw []byte) bool {
	if len(raw) < len(sig)+2 {
		return false
	}
	if raw[0] != '#' || raw[1] != '!' {
		return false
	}

	return false
}

func firstLine(in []byte) []byte {
	lineEnd := 0
	for ; lineEnd < len(in) && in[lineEnd] != '\n'; lineEnd++ {
	}

	return in[:lineEnd]
}

// prefix creates a Detector which returns true if any of the provided signatures
// is the prefix of the raw input.
func prefixer(sigs ...[]byte) func(raw []byte, limit uint32) bool {
	return func(raw []byte, limit uint32) bool {
		for _, s := range sigs {
			if bytes.HasPrefix(raw, s) {
				return true
			}
		}
		return false
	}
}
