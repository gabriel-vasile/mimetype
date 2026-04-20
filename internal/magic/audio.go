package magic

import (
	"bytes"
	"encoding/binary"
)

// Flac matches a Free Lossless Audio Codec file.
func Flac(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
}

// Midi matches a Musical Instrument Digital Interface file.
func Midi(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("\x4D\x54\x68\x64"))
}

// Ape matches a Monkey's Audio file.
func Ape(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("\x4D\x41\x43\x20\x96\x0F\x00\x00\x34\x00\x00\x00\x18\x00\x00\x00\x90\xE3"))
}

// MusePack matches a Musepack file.
func MusePack(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("MPCK"))
}

// Au matches a Sun Microsystems au file.
func Au(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("\x2E\x73\x6E\x64"))
}

// Amr matches an Adaptive Multi-Rate file.
func Amr(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("\x23\x21\x41\x4D\x52"))
}

// Voc matches a Creative Voice file.
func Voc(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("Creative Voice File"))
}

// M3U matches a Playlist file.
func M3U(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("#EXTM3U\n")) ||
		bytes.HasPrefix(raw, []byte("#EXTM3U\r\n"))
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

	// Any ID3v2 is reported as MP3. Not entirely correct, but the mimesniff
	// standard says so. https://mimesniff.spec.whatwg.org/#matching-an-audio-or-video-type-pattern
	// Despite the standard only checking for "ID3", we do more validations to
	// avoid false positives.
	if id3v2(raw) {
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

// Based on https://id3.org/Developer%20Information.
func id3v2(raw []byte) bool {
	if len(raw) < 10 || !bytes.HasPrefix(raw, []byte("ID3")) {
		return false
	}
	if raw[3] < 2 || raw[3] > 4 { // Version: ID3v2.2 - ID3v2.4.
		return false
	}
	if raw[4] != 0 { // Revision is 0 for all versions.
		return false
	}

	// v2.2 uses 2 bits, v2.3 uses 3 bits and v2.4 uses 4.
	// For all versions least significant 4 bits should be 0
	if raw[5]&0b1111 != 0 {
		return false
	}

	// Size bytes are synchsafe: most significant bit always 0.
	if raw[6]&0x80 != 0 || raw[7]&0x80 != 0 || raw[8]&0x80 != 0 || raw[9]&0x80 != 0 {
		return false
	}

	size := uint32(raw[6])<<21 | uint32(raw[7])<<14 | uint32(raw[8])<<7 | uint32(raw[9])
	// Disallow too big frames, let's say 10MB.
	return size > 0 && size < 10*1024*1024
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

// EightSVX matches an 8-bit Sampled Voice file.
func EightSVX(raw []byte, _ uint32) bool {
	return len(raw) > 12 &&
		bytes.Equal(raw[:4], []byte("FORM")) &&
		bytes.Equal(raw[8:12], []byte("8SVX"))
}

// Sid matches a Commodore 64 SID music file.
func Sid(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("PSID")) ||
		bytes.HasPrefix(raw, []byte("RSID"))
}

// XM matches a FastTracker II Extended Module file.
func XM(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("Extended Module: "))
}

// Mod matches a ProTracker Module file.
func Mod(raw []byte, _ uint32) bool {
	if len(raw) < 1084 {
		return false
	}

	// ProTracker and compatible modules have a signature at offset 1080.
	sig := raw[1080:1084]
	return bytes.Equal(sig, []byte("M.K.")) || // 4 channels
		bytes.Equal(sig, []byte("M!K!")) || // 4 channels
		bytes.Equal(sig, []byte("FLT4")) || // 4 channels
		bytes.Equal(sig, []byte("FLT8")) || // 8 channels
		bytes.Equal(sig, []byte("4CHN")) || // 4 channels
		bytes.Equal(sig, []byte("6CHN")) || // 6 channels
		bytes.Equal(sig, []byte("8CHN")) || // 8 channels
		bytes.Equal(sig, []byte("16CH")) || // 16 channels
		bytes.Equal(sig, []byte("32CH")) // 32 channels
}

// S3M matches a ScreamTracker 3 Module file.
func S3M(raw []byte, _ uint32) bool {
	return len(raw) > 48 && bytes.Equal(raw[44:48], []byte("SCRM"))
}

// IT matches an Impulse Tracker Module file.
func IT(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("IMPM"))
}

// Med matches an OctaMED tracker module.
func Med(raw []byte, _ uint32) bool {
	return len(raw) > 3 && bytes.HasPrefix(raw, []byte("MMD")) &&
		(raw[3] >= '0' && raw[3] <= '3')
}

// Ahx matches an AHX (Abyss' Highest Experience) tracker module.
func Ahx(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("THX\x00"))
}
