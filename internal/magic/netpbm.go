package magic

import "bytes"

// NetPBM matches a Netpbm Portable BitMap ASCII/Binary file.
//
// See: https://en.wikipedia.org/wiki/Netpbm
func NetPBM(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("P1")) ||
		bytes.HasPrefix(raw, []byte("P4"))
}

// NetPGM matches a Netpbm Portable GrayMap ASCII/Binary file.
//
// See: https://en.wikipedia.org/wiki/Netpbm
func NetPGM(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("P2")) ||
		bytes.HasPrefix(raw, []byte("P5"))
}

// NetPPM matches a Netpbm Portable PixMap ASCII/Binary file.
//
// See: https://en.wikipedia.org/wiki/Netpbm
func NetPPM(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("P3")) ||
		bytes.HasPrefix(raw, []byte("P6"))
}

// NetPFM matches a Netpbm Portable FloatMap ASCII/Binary file.
//
// See: https://en.wikipedia.org/wiki/Netpbm
func NetPFM(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("PF")) ||
		bytes.HasPrefix(raw, []byte("Pf"))
}

// NetPAM matches a Netpbm Portable Arbitrary Map file.
//
// See: https://en.wikipedia.org/wiki/Netpbm
func NetPAM(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("P7"))
}
