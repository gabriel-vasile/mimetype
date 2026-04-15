package magic

import (
	"bytes"
	"encoding/binary"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

type mp3FrameHeader uint32

func newMP3FrameHeader(b scan.Bytes) mp3FrameHeader {
	if len(b) < 4 {
		return 0
	}
	return mp3FrameHeader(binary.BigEndian.Uint32(b))
}

// Sometimes mp3s come with leading bytes before the first frame.
// Those are still valid files, decoders are supposed to skip padding and find
// first valid frame. #310, #775
func mp3WithLeadingPadding(b scan.Bytes) bool {
	const (
		mp3MaxSearch       = 2048
		mp3FrameHeaderSize = 4
		// how many consecutive frames with no junk in-between for the file to pass as mp3
		mp3ConsecutiveFrames = 4
	)

	frames := 0
	search := mp3MaxSearch
	var firstHeader mp3FrameHeader
	for {
		// For the first frame we allow mp3MaxSearch bytes to find the frame sync.
		// For subsequent frames, we reduce that to 1 byte, meaning when jumping
		// from previous frame, we land exactly on the next frame.
		// We do this, because if we we're to do linear searches for all frames,
		// some binary files, mostly cpio, would give false-positives.
		if frames > 0 {
			search = 1
		}
		ff := bytes.IndexByte(b[:min(len(b), search)], 0xff)
		if ff == -1 {
			return false
		}

		b.Advance(ff)
		if h := newMP3FrameHeader(b); h.isValid() {
			if frames == 0 {
				firstHeader = h
			} else {
				if !h.equals(firstHeader) {
					return false
				}
			}
			frames++
			b.Advance(min(len(b), h.size()))
		} else {
			if !b.Advance(mp3FrameHeaderSize) {
				return false
			}
			frames = 0
			firstHeader = 0
		}
		if frames >= mp3ConsecutiveFrames {
			return true
		}
	}
	return false
}

func (h mp3FrameHeader) isValid() bool {
	return (h>>24)&0xFF == 0xFF &&
		(h>>16)&0xE0 == 0xE0 &&
		h.emphasis() != 2 &&
		h.layer() != 0 &&
		h.version() != 1 &&
		h.sampleRate() != -1 &&
		h.bitRate() != -1
}

func (h mp3FrameHeader) emphasis() byte {
	return byte(h & 0x03)
}

func (h mp3FrameHeader) layer() byte {
	return byte((h >> 17) & 0x03)
}

func (h mp3FrameHeader) version() byte {
	return byte((h >> 19) & 0x03)
}

func (h mp3FrameHeader) sampleRate() int {
	sri := (h >> 10) & 0x03
	if sri == 0x03 {
		return -1
	}
	return sampleRates[h.version()][sri]
}

func (h mp3FrameHeader) bitRate() int {
	bitrateIdx := (h >> 12) & 0x0F
	if bitrateIdx == 0x0F {
		return -1
	}
	br := bitrates[h.version()][h.layer()][bitrateIdx] * 1000
	if br == 0 {
		return -1
	}
	return br
}

func (h mp3FrameHeader) samples() int {
	return samplesPerFrame[h.version()][h.layer()]
}

func (h mp3FrameHeader) size() int {
	bps := h.samples() / 8
	fsize := (bps * h.bitRate()) / h.sampleRate()
	if h.pad() {
		fsize += slotSize[h.layer()]
	}
	return fsize
}

func (h mp3FrameHeader) pad() bool {
	return (h>>9)&0x01 == 0x01
}

func (h mp3FrameHeader) copyright() bool {
	return (h>>3)&0x01 == 0x01
}

func (h mp3FrameHeader) crc() bool {
	return (h>>16)&0x01 != 0x01
}

func (h mp3FrameHeader) mode() byte {
	return byte((h >> 6) & 0x03)
}

func (h mp3FrameHeader) original() bool {
	return (h>>2)&0x01 == 0x01
}

func (h mp3FrameHeader) freq() byte {
	return byte((h >> 10) & 0x03)
}

func (h mp3FrameHeader) equals(b mp3FrameHeader) bool {
	return h == b ||
		h.version() == b.version() &&
			h.layer() == b.layer() &&
			h.crc() == b.crc() &&
			h.freq() == b.freq() &&
			h.mode() == b.mode() &&
			h.copyright() == b.copyright() &&
			h.original() == b.original() &&
			h.emphasis() == b.emphasis()
}

var samplesPerFrame = [4][4]int{
	{0, 576, 1152, 384},  // MPEG25
	{0, 0, 0, 0},         // Reserved
	{0, 576, 1152, 384},  // MPEG2
	{0, 1152, 1152, 384}, // MPEG1
}

var sampleRates = [4][3]int{
	{11025, 12000, 8000},  // MPEG25
	{0, 0, 0},             // MPEGReserved
	{22050, 24000, 16000}, // MPEG2
	{44100, 48000, 32000}, // MPEG1
}

var bitrates = [4][4][15]int{
	{ // MPEG 2.5
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},                       // LayerReserved
		{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // Layer3
		{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // Layer2
		{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256}, // Layer1
	},
	{ // Reserved
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // LayerReserved
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Layer3
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Layer2
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Layer1
	},
	{ // MPEG 2
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},                       // LayerReserved
		{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // Layer3
		{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // Layer2
		{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256}, // Layer1
	},
	{ // MPEG 1
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},                          // LayerReserved
		{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320},     // Layer3
		{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384},    // Layer2
		{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448}, // Layer1
	},
}

var slotSize = [4]int{
	0, // LayerReserved
	1, // Layer3
	1, // Layer2
	4, // Layer1
}
