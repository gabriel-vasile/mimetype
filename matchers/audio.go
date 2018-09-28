package matchers

import (
	"bytes"
)

func Mp3(in []byte) bool {
	return bytes.HasPrefix(in, []byte("\x49\x44\x33"))
}

func Flac(in []byte) bool {
	return bytes.HasPrefix(in, []byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
}

func Midi(in []byte) bool {
	return bytes.HasPrefix(in, []byte("\x4D\x54\x68\x64"))
}

func Ape(in []byte) bool {
	return bytes.HasPrefix(in, []byte("\x4D\x41\x43\x20\x96\x0F\x00\x00\x34\x00\x00\x00\x18\x00\x00\x00\x90\xE3"))
}

// TODO
func MusePack(in []byte) bool {
	return false
}

func Wav(in []byte) bool {
	return bytes.Equal(in[:4], []byte("\x52\x49\x46\x46")) && bytes.Equal(in[8:12], []byte("\x57\x41\x56\x45"))
}

func Aiff(in []byte) bool {
	return bytes.Equal(in[:4], []byte("\x46\x4F\x52\x4D")) && bytes.Equal(in[8:12], []byte("\x41\x49\x46\x46"))
}

func Ogg(in []byte) bool {
	return bytes.Equal(in[:5], []byte("\x4F\x67\x67\x53\x00"))
}
