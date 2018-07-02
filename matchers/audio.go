package matchers

import "bytes"

func Mp3(in []byte) bool {
	return bytes.HasPrefix(in, []byte("\x49\x44\x33"))
}

func Flac(in []byte) bool {
	return false
}

func Midi(in []byte) bool {
	return false
}

func Ape(in []byte) bool {
	return false
}

func MusePack(in []byte) bool {
	return false
}

func Wav(in []byte) bool {
	return false
}

func Aiff(in []byte) bool {
	return false
}
