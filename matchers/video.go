package matchers

import (
	"bytes"
	"encoding/binary"
)

func Mp4(in []byte) bool {
	if len(in) < 12 {
		return false
	}

	mp4ftype := []byte("ftyp")
	mp4 := []byte("mp4")
	boxSize := int(binary.BigEndian.Uint32(in[:4]))
	if boxSize%4 != 0 || len(in) < boxSize {
		return false
	}
	if !bytes.Equal(in[4:8], mp4ftype) {
		return false
	}
	for st := 8; st < boxSize; st += 4 {
		if st == 12 {
			// minor version number
			continue
		}
		if bytes.Equal(in[st:st+3], mp4) {
			return true
		}
	}

	return false
}

func WebM(in []byte) bool {
	return bytes.HasPrefix(in, []byte("\x1A\x45\xDF\xA3"))
}

func ThreeGP(in []byte) bool {
	return len(in) > 11 &&
		bytes.HasPrefix(in[4:], []byte("\x66\x74\x79\x70\x33\x67\x70"))
}

func Flv(in []byte) bool {
	return bytes.HasPrefix(in, []byte("\x46\x4C\x56\x01"))
}

// TODO
func Mpeg(in []byte) bool {
	return false
}

func Quicktime(in []byte) bool {
	return false
}

func Avi(in []byte) bool {
	return false
}
