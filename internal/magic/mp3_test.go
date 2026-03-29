package magic

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math"
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// buildFrame constructs a 4-byte MP3 frame header with the given parameters.
// version: 0=MPEG2.5, 2=MPEG2, 3=MPEG1
// layer:   1=Layer3, 2=Layer2, 3=Layer1
// bitrateIdx: index into bitrates table (1-14 for valid)
// sampleRateIdx: 0-2 for valid
func buildFrame(version, layer, bitrateIdx, sampleRateIdx byte) []byte {
	b := make([]byte, 4)
	b[0] = 0xFF
	b[1] = 0xE0 | (version << 3) | (layer << 1) | 0x01 // crc bit set
	b[2] = (bitrateIdx << 4) | (sampleRateIdx << 2)
	b[3] = 0x00 // no emphasis, stereo, not copyrighted, not original
	return b
}

// buildValidMPEG1Layer3Frame builds a known-good MPEG1 Layer3 frame header.
// MPEG1=3, Layer3=1, bitrate index 9 (128kbps), sample rate index 0 (44100Hz)
// A real MPEG1 Layer3 frame at 128kbps/44100Hz is 417 bytes.
// We use a smaller size for test brevity; the function only checks
// the header of each frame, not the audio payload.
func buildValidMPEG1Layer3Frame() []byte {
	return buildFrame(0, 1, 1, 2)
	return buildFrame(3, 1, 9, 0)
}

func buildSmallestValidFrame() []byte {
	return buildFrame(0, 1, 1, 2)
}

// 00000000000000001111001100110000
func TestAsd(t *testing.T) {
	return
	f := newMP3FrameHeader(buildFrame(3, 1, 9, 0))
	asd := mp3FrameHeader(math.MaxUint32)
	asd = mp3FrameHeader(0)
	for i := uint32(0); i < math.MaxUint32; i++ {
		h := mp3FrameHeader(i)
		if !h.isValid() {
			continue
		}
		if f.equals(h) {
			// fmt.Printf("e bine %0.32b %0.32b\n", f, h)
			asd |= f ^ h
		}
	}
	fmt.Printf("aaaaa %0.32b\n", asd)
	fmt.Println(asd)
}

// repeatFrames builds a byte slice with `count` consecutive valid frames,
// each padded to `frameSize` bytes total. If rng, the padding values will be
// random instead of 0x00.
func repeatFrames(header []byte, frameSize, count int, rng bool) []byte {
	var out []byte
	for i := 0; i < count; i++ {
		frame := make([]byte, frameSize)
		if rng {
			rand.Read(frame)
		}
		copy(frame[:4], header)
		if rng {
			prepend := make([]byte, frameSize)
			rand.Read(prepend)
			out = append(out, prepend...)
		}
		out = append(out, frame...)
	}
	return out
}

func TestMP3WithLeadingPadding(t *testing.T) {
	validHeader := buildValidMPEG1Layer3Frame()

	const frameSize = 72
	const mp3MaxSearch = 2048

	tests := []struct {
		name string
		data []byte
		want bool
	}{{
		name: "four consecutive valid frames, no padding",
		data: repeatFrames(validHeader, frameSize, 4, false),
		want: true,
	}, {
		name: "four valid frames with leading 0x00 padding",
		data: append(make([]byte, 100), repeatFrames(validHeader, frameSize, 4, false)...),
		want: true,
	}, {
		name: "four valid frames with leading 0xFF padding",
		data: append(bytes.Repeat([]byte{0xFF}, 100), repeatFrames(validHeader, frameSize, 4, false)...),
		want: true,
	}, {
		name: "four valid frames with leading 2047 0x00 bytes padding",
		data: append(bytes.Repeat([]byte{0x00}, 2047), repeatFrames(validHeader, frameSize, 4, false)...),
		want: true,
	}, {
		name: "four valid frames with leading 2048 0x00 bytes padding",
		data: append(bytes.Repeat([]byte{0x00}, 2048), repeatFrames(validHeader, frameSize, 4, false)...),
		want: false,
	}, {
		name: "only three valid frames (below threshold)",
		data: repeatFrames(validHeader, frameSize, 3, false),
		want: false,
	}, {
		name: "empty input",
		data: []byte{},
		want: false,
	}, {
		name: "all zeros, no sync byte",
		data: make([]byte, 1024),
		want: false,
	}, {
		name: "0xFF bytes but no valid frame follows",
		data: []byte{0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00},
		want: false,
	}, {
		name: "leading padding exceeds mp3MaxSearch before first frame",
		data: append(make([]byte, mp3MaxSearch+1), repeatFrames(validHeader, frameSize, 4, false)...),
		want: false,
	}, {
		name: "leading padding exactly at mp3MaxSearch boundary",
		// mp3MaxSearch-4 bytes of zeros, then a valid frame starting with 0xFF
		data: append(make([]byte, mp3MaxSearch-4), repeatFrames(validHeader, frameSize, 4, false)...),
		want: true,
	}, {
		name: "mismatched second frame version",
		// First frame: MPEG1 (version=3), second frame: MPEG2 (version=2)
		data: func() []byte {
			f1 := buildFrame(3, 1, 9, 0)
			f2 := buildFrame(2, 1, 9, 0)
			var out []byte
			frame1 := make([]byte, frameSize)
			copy(frame1, f1)
			out = append(out, frame1...)
			frame2 := make([]byte, frameSize)
			copy(frame2, f2)
			out = append(out, frame2...)
			return out
		}(),
		want: false,
	}, {
		name: "invalid emphasis byte (reserved value 2)",
		data: func() []byte {
			h := make([]byte, 4)
			copy(h, validHeader)
			h[3] = 0x02 // emphasis = 2 (reserved)
			return repeatFrames(h, frameSize, 4, false)
		}(),
		want: false,
	}, {
		name: "invalid layer (reserved value 0)",
		data: func() []byte {
			h := make([]byte, 4)
			copy(h, validHeader)
			h[1] = (h[1] & 0xF9) // layer bits = 00 (reserved)
			return repeatFrames(h, frameSize, 4, false)
		}(),
		want: false,
	}, {
		name: "invalid version (reserved value 1)",
		data: func() []byte {
			h := make([]byte, 4)
			copy(h, validHeader)
			h[1] = (h[1] & 0xE7) | (1 << 3) // version bits = 01 (reserved)
			return repeatFrames(h, frameSize, 4, false)
		}(),
		want: false,
	}, {
		name: "invalid sample rate index (0x03 = reserved)",
		data: func() []byte {
			h := make([]byte, 4)
			copy(h, validHeader)
			h[2] = (h[2] & 0xF0) | 0x0C // sampleRateIdx = 3
			return repeatFrames(h, frameSize, 4, false)
		}(),
		want: false,
	}, {
		name: "invalid bitrate index (0x0F = bad)",
		data: func() []byte {
			h := make([]byte, 4)
			copy(h, validHeader)
			h[2] = (h[2] & 0x0F) | 0xF0 // bitrateIdx = 15
			return repeatFrames(h, frameSize, 4, false)
		}(),
		want: false,
	}, {
		name: "input shorter than one frame header",
		data: []byte{0xFF, 0xFB},
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := scan.Bytes(tt.data)
			got := mp3WithLeadingPadding(b)
			if got != tt.want {
				t.Errorf("mp3WithLeadingPadding() = %v, want %v", got, tt.want)
			}
		})
	}
}

// This could be a fuzzing test, but there are only 256*256*256 possible values
// for header so we can test most of them in under 1s.
func TestRandomJunkBeforeFrames(t *testing.T) {
	h := [4]byte{0xFF, 0x00, 0x00, 0x00}
	// a must start from 0xE0 to be valid.
	// Start from 0xD0 in this test to check for a bad value.
	for a := byte(0xD0); a < 0xFF; a++ {
		for b := byte(0x00); b < 0xFF; b++ {
			// Last byte in header doesn't unfluence much, so test just four values.
			for _, c := range []byte{0x00, 0b01010101, 0b10101010, 0xFF} {
				h[1], h[2], h[3] = a, b, c

				fh := newMP3FrameHeader(scan.Bytes(h[:]))
				if !fh.isValid() {
					continue
				}
				if fh.size() < 24 {
					t.Errorf("size should never go below 24")
				}
				mp3 := repeatFrames(h[:], fh.size(), 4, true)
				mp3WithLeadingPadding(scan.Bytes(mp3))
			}
		}
	}
}

// TODO: sober.mp3 debug cu minimp3.h ca sa aflu cum face cu free_format_bytes
func TestAsd1(t *testing.T) {
	d, _ := ioutil.ReadFile("/home/gabriel/Downloads/sober.mp3")

	fmt.Println(len(d))
	fmt.Println(mp3WithLeadingPadding(scan.Bytes(d)))
	return
	for i := 0; i < len(d); i++ {
		if mp3WithLeadingPadding(scan.Bytes(d[:i])) {
			fmt.Println(i)
			break
		} else {
			fmt.Println("no")
		}
	}
}
