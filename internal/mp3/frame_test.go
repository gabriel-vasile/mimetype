package mp3

import (
	"bytes"
	"errors"
	"math/rand/v2"
	"testing"
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

func bytesToHeader(b []byte) header {
	return header{b[0], b[1], b[2], b[3]}
}

// buildValidMP3Header builds a known-good MPEG2.5 Layer3 frame header.
// MPEG2.5 (version=0), Layer3, bitrate index 1 (8 kbps), sample rate index 2
// (8000 Hz). That yields 72-byte frames, which is small enough to keep tests
// fast while exercising real frame-size arithmetic.
func buildValidMP3Header() []byte {
	return buildFrame(0, 1, 1, 2)
}

// repeatFrames builds a byte slice with `count` consecutive valid frames,
// each padded to `frameSize` bytes total. If rng, the padding values will be
// random instead of 0x00.
func repeatFrames(header []byte, frameSize, count int, rng bool) []byte {
	var out []byte

	r := rand.NewChaCha8([32]byte{})
	for i := 0; i < count; i++ {
		frame := make([]byte, frameSize)
		if rng {
			_, _ = r.Read(frame)
		}
		copy(frame[:4], header)
		if rng {
			prepend := make([]byte, frameSize)
			_, _ = r.Read(prepend)
			out = append(out, prepend...)
		}
		out = append(out, frame...)
	}
	return out
}

func TestExtractFrame(t *testing.T) {
	validHeader := buildValidMP3Header()
	frameSize := bytesToHeader(validHeader).frameBytes()

	// Matches how many bytes the implementation searches for.
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
		name: "only two valid frames (below threshold)",
		data: repeatFrames(validHeader, frameSize, 2, false),
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
		name: "4 headers with padding",
		data: func() []byte {
			padded := append([]byte{}, validHeader...)
			padded[2] |= 0x2
			return repeatFrames(padded, frameSize+1, 4, false)
		}(),
		want: true,
	}, {
		name: "11 headers to have more than maxFrameSyncMatches",
		data: repeatFrames(validHeader, frameSize, 11, false),
		want: true,
	}, {
		name: "second frame header is not valid",
		data: func() []byte {
			f1 := buildFrame(3, 1, 9, 0)
			f2 := buildFrame(0, 0, 0, 0)
			frame1 := make([]byte, bytesToHeader(f1).frameBytes())
			frame2 := make([]byte, frameSize)
			copy(frame1, f1)
			copy(frame2, f2)
			out := make([]byte, 0, len(frame1)+len(frame2))
			out = append(out, frame1...)
			out = append(out, frame2...)
			return out
		}(),
		want: false,
	}, {
		name: "mismatched second frame version",
		// First frame: MPEG1 (version=3), second frame: MPEG2 (version=2)
		data: func() []byte {
			f1 := buildFrame(3, 1, 9, 0)
			f2 := buildFrame(2, 1, 9, 0)
			frame1 := make([]byte, bytesToHeader(f1).frameBytes())
			frame2 := make([]byte, bytesToHeader(f2).frameBytes())
			copy(frame1, f1)
			copy(frame2, f2)
			out := make([]byte, 0, len(frame1)+len(frame2))
			out = append(out, frame1...)
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
			_, e := ExtractFrame(tt.data)
			if got := e != 0; got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkExtractFrame(b *testing.B) {
	validHeader := buildValidMP3Header()
	frameSize := bytesToHeader(validHeader).frameBytes()

	b.Run("repeatFrames", func(b *testing.B) {
		b.ReportAllocs()
		data := repeatFrames(validHeader, frameSize, 4, false)
		data = append(data, repeatFrames(validHeader, frameSize, 4, true)...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ExtractFrame(data)
		}
	})

	b.Run("random data", func(b *testing.B) {
		b.ReportAllocs()
		largeData := make([]byte, 4096)
		r := rand.NewChaCha8([32]byte{})
		if _, err := r.Read(largeData); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ExtractFrame(largeData)
		}
	})
}

func FuzzExtractFrame(f *testing.F) {
	overlapHeadOnFrame := func(head, frame []byte) {
		for i := 0; i < 4; i++ {
			frame[i] = head[i]
		}
	}
	// Only fuzz the header of the frame, because the body is treated as arbitrary
	// data that does not influence anything.
	// We're not seeding the first byte of the header because it is always 0xFF.
	h := buildValidMP3Header()
	f.Add(h[1], h[2], h[3], h[1], h[2], h[3], h[1], h[2], h[3])

	frame1, frame2, frame3 := make([]byte, 1441), make([]byte, 1441), make([]byte, 1441)
	r := rand.NewChaCha8([32]byte{})
	_, err1 := r.Read(frame1)
	_, err2 := r.Read(frame2)
	_, err3 := r.Read(frame3)
	if err := errors.Join(err1, err2, err3); err != nil {
		f.Fatal(err)
	}

	// Allocate data only once.
	data := make([]byte, 0, 3*1441)
	f.Fuzz(func(t *testing.T, h11, h12, h13, h21, h22, h23, h31, h32, h33 byte) {
		head1 := bytesToHeader([]byte{0xFF, h11, h12, h13})
		if !head1.valid() {
			return
		}
		head2 := bytesToHeader([]byte{0xFF, h21, h22, h23})
		if !head2.valid() {
			head2 = head1
		}
		head3 := bytesToHeader([]byte{0xFF, h31, h32, h33})
		if !head3.valid() {
			head3 = head2
		}
		overlapHeadOnFrame(head1[:], frame1)
		overlapHeadOnFrame(head2[:], frame2)
		overlapHeadOnFrame(head3[:], frame3)

		data = data[:0]
		// Sometimes put trash bytes before and after the frames.
		if h11%2 == 0 {
			data = append(data, h11, h12, h13)
		}
		data = append(data, frame1[:head1.frameBytes()+head1.padding()]...)
		data = append(data, frame2[:head2.frameBytes()+head2.padding()]...)
		data = append(data, frame3[:head3.frameBytes()+head3.padding()]...)
		if h21%2 == 0 {
			data = append(data, h21, h22, h23)
		}
		ExtractFrame(data)
	})
}
