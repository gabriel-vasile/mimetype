package magic

import (
	"bytes"
	"encoding/binary"
)

// Woff matches a Web Open Font Format file.
func Woff(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("wOFF"))
}

// Woff2 matches a Web Open Font Format version 2 file.
func Woff2(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("wOF2"))
}

// Otf matches an OpenType font file.
func Otf(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("OTTO")) && hasSFNTTable(raw)
}

// Ttf matches a TrueType font file.
func Ttf(raw []byte, limit uint32) bool {
	if !bytes.HasPrefix(raw, []byte{0x00, 0x01, 0x00, 0x00}) {
		return false
	}
	return hasSFNTTable(raw)
}

func hasSFNTTable(raw []byte) bool {
	// 49 possible tables as explained below
	if len(raw) < 16 || binary.BigEndian.Uint16(raw[4:]) >= 49 {
		return false
	}

	// libmagic says there are 47 table names in specification, but it seems
	// they reached 49 in the meantime.
	// https://github.com/file/file/blob/5184ca2471c0e801c156ee120a90e669fe27b31d/magic/Magdir/fonts#L279
	// At the same time, the TrueType docs seem misleading:
	// 1. https://developer.apple.com/fonts/TrueType-Reference-Manual/index.html
	// 2. https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6.html
	// Page 1. has 48 tables. Page 2. has 49 tables. The diff is the gcid table.
	// Take a permissive approach,
	possibleTables := []string{
		"acnt", "ankr", "avar", "bdat", "bhed", "bloc", "bsln", "cmap", "cvar",
		"cvt ", "EBSC", "fdsc", "feat", "fmtx", "fond", "fpgm", "fvar", "gasp",
		"gcid", "glyf", "gvar", "hdmx", "head", "hhea", "hmtx", "hvgl", "hvpm",
		"just", "kern", "kerx", "lcar", "loca", "ltag", "maxp", "meta", "mort",
		"morx", "name", "opbd", "OS/2", "post", "prep", "prop", "sbix", "trak",
		"vhea", "vmtx", "xref", "Zapf",
	}
	// TODO: benchmark these strings comparisons. They are 4 bytes, so another
	// option is to compare them as ints. Probably less readable that way.
	for _, t := range possibleTables {
		if string(raw[12:16]) == t {
			return true
		}
	}
	return false
}

// Eot matches an Embedded OpenType font file.
func Eot(raw []byte, limit uint32) bool {
	return len(raw) > 35 &&
		bytes.Equal(raw[34:36], []byte{0x4C, 0x50}) &&
		(bytes.Equal(raw[8:11], []byte{0x02, 0x00, 0x01}) ||
			bytes.Equal(raw[8:11], []byte{0x01, 0x00, 0x00}) ||
			bytes.Equal(raw[8:11], []byte{0x02, 0x00, 0x02}))
}

// Ttc matches a TrueType Collection font file.
func Ttc(raw []byte, limit uint32) bool {
	return len(raw) > 7 &&
		bytes.HasPrefix(raw, []byte("ttcf")) &&
		(bytes.Equal(raw[4:8], []byte{0x00, 0x01, 0x00, 0x00}) ||
			bytes.Equal(raw[4:8], []byte{0x00, 0x02, 0x00, 0x00}))
}
