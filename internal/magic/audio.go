package magic

import (
	"bytes"
	"encoding/binary"
)

// Flac matches a Free Lossless Audio Codec file.
func Flac(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
}

// Midi matches a Musical Instrument Digital Interface file.
func Midi(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("\x4D\x54\x68\x64"))
}

// Ape matches a Monkey's Audio file.
func Ape(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("\x4D\x41\x43\x20\x96\x0F\x00\x00\x34\x00\x00\x00\x18\x00\x00\x00\x90\xE3"))
}

// MusePack matches a Musepack file.
func MusePack(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("MPCK"))
}

// Au matches a Sun Microsystems au file.
func Au(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("\x2E\x73\x6E\x64"))
}

// Amr matches an Adaptive Multi-Rate file.
func Amr(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("\x23\x21\x41\x4D\x52"))
}

// Voc matches a Creative Voice file.
func Voc(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("Creative Voice File"))
}

// M3u matches a Playlist file.
func M3u(raw []byte, _ uint32) bool {
	return prefix(raw, []byte("#EXTM3U"))
}

// AAC matches an Advanced Audio Coding file.
func AAC(raw []byte, _ uint32) bool {
	return len(raw) > 1 && ((raw[0] == 0xFF && raw[1] == 0xF1) || (raw[0] == 0xFF && raw[1] == 0xF9))
}

// Mp3 matches an mp3 file.
func Mp3(raw []byte, limit uint32) bool {
	if len(raw) < 3 {
		return false
	}

	if bytes.HasPrefix(raw, []byte("ID3")) {
		// MP3s with an ID3v2 tag will start with "ID3"
		// ID3v1 tags, however appear at the end of the file.
		return true
	}

	// Match MP3 files without tags
	switch binary.BigEndian.Uint16(raw[:2]) & 0xFFFE {
	case 0xFFFA:
		// MPEG ADTS, layer III, v1
		return true
	case 0xFFF2:
		// MPEG ADTS, layer III, v2
		return true
	case 0xFFE2:
		// MPEG ADTS, layer III, v2.5
		return true
	}

	return false
}

// Wav matches a Waveform Audio File Format file.
func Wav(raw []byte, limit uint32) bool {
	return len(raw) > 12 &&
		bytes.Equal(raw[:4], []byte("RIFF")) &&
		bytes.Equal(raw[8:12], []byte{0x57, 0x41, 0x56, 0x45})
}

// Aiff matches Audio Interchange File Format file.
func Aiff(raw []byte, limit uint32) bool {
	return len(raw) > 12 &&
		bytes.Equal(raw[:4], []byte{0x46, 0x4F, 0x52, 0x4D}) &&
		bytes.Equal(raw[8:12], []byte{0x41, 0x49, 0x46, 0x46})
}

// Qcp matches a Qualcomm Pure Voice file.
func Qcp(raw []byte, limit uint32) bool {
	return len(raw) > 12 &&
		bytes.Equal(raw[:4], []byte("RIFF")) &&
		bytes.Equal(raw[8:12], []byte("QLCM"))
}
