package magic

import (
	"bytes"
	"encoding/binary"

	"github.com/gabriel-vasile/mimetype/internal/scan"
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
	return mp3PRONOM(raw)
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

	return mp3PRONOM(raw)
}

// mp3PRONOM implements the BOF signatures from
// https://www.nationalarchives.gov.uk/PRONOM/Format/proFormatSearch.aspx?status=detailReport&id=687&strPageToDisplay=signatures
func mp3PRONOM(b scan.Bytes) bool {
	for i := 0; i < 31; i++ {
		// 1439 is the maximum spacing between frames.
		ff := bytes.IndexByte(b[:min(len(b), 1439)], 0xff)
		if ff == -1 {
			return false
		}

		b.Advance(ff + 1)
		c := b.PopN(2)
		if len(c) != 2 {
			return false
		}
		if c[0] != 0xfb && c[0] != 0xf3 && c[0] != 0xfa && c[0] != 0xf2 && c[0] != 0xe3 {
			return false
		}
		if c[1] < 0x10 || c[1] > 0xeb {
			return false
		}
	}
	return true
}

func mp3v2(b scan.Bytes, ff []byte) bool {
	for i := 0; i < 3; i++ {
		ff := bytes.Index(b, ff)
		if ff == -1 {
			return false
		}

		b.Advance(ff + 2)
		c := b.Pop()
		if c < 0x10 || c > 0xeb {
			return false
		}
		b.Advance(min(46, len(b)))
	}
	return true
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
