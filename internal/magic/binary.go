package magic

import (
	"bytes"
	"debug/macho"
	"encoding/binary"
)

var (
	// Lnk matches Microsoft lnk binary format.
	Lnk = prefix([]byte{0x4C, 0x00, 0x00, 0x00, 0x01, 0x14, 0x02, 0x00})
	// Wasm matches a web assembly File Format file.
	Wasm = prefix([]byte{0x00, 0x61, 0x73, 0x6D})
	// Exe matches a Windows/DOS executable file.
	Exe = prefix([]byte{0x4D, 0x5A})
	// Elf matches an Executable and Linkable Format file.
	Elf = prefix([]byte{0x7F, 0x45, 0x4C, 0x46})
	// Nes matches a Nintendo Entertainment system ROM file.
	Nes = prefix([]byte{0x4E, 0x45, 0x53, 0x1A})
	// TzIf matches a Time Zone Information Format (TZif) file.
	TzIf = prefix([]byte("TZif"))
)

// Java bytecode and Mach-O binaries share the same magic number.
// More info here https://github.com/threatstack/libmagic/blob/master/magic/Magdir/cafebabe
func classOrMachOFat(in []byte) bool {
	// There should be at least 8 bytes for both of them because the only way to
	// quickly distinguish them is by comparing byte at position 7
	if len(in) < 8 {
		return false
	}

	return bytes.HasPrefix(in, []byte{0xCA, 0xFE, 0xBA, 0xBE})
}

// Class matches a java class file.
func Class(raw []byte, limit uint32) bool {
	return classOrMachOFat(raw) && raw[7] > 30
}

// MachO matches Mach-O binaries format.
func MachO(raw []byte, limit uint32) bool {
	if classOrMachOFat(raw) && raw[7] < 20 {
		return true
	}

	if len(raw) < 4 {
		return false
	}

	be := binary.BigEndian.Uint32(raw)
	le := binary.LittleEndian.Uint32(raw)

	return be == macho.Magic32 ||
		le == macho.Magic32 ||
		be == macho.Magic64 ||
		le == macho.Magic64
}

// Swf matches an Adobe Flash swf file.
func Swf(raw []byte, limit uint32) bool {
	return bytes.HasPrefix(raw, []byte("CWS")) ||
		bytes.HasPrefix(raw, []byte("FWS")) ||
		bytes.HasPrefix(raw, []byte("ZWS"))
}

// Dbf matches a dBase file.
// https://www.dbase.com/Knowledgebase/INT/db7_file_fmt.htm
func Dbf(raw []byte, limit uint32) bool {
	if len(raw) < 4 {
		return false
	}

	// 3rd and 4th bytes contain the last update month and day of month
	if !(0 < raw[2] && raw[2] < 13 && 0 < raw[3] && raw[3] < 32) {
		return false
	}

	// dbf type is dictated by the first byte
	dbfTypes := []byte{
		0x02, 0x03, 0x04, 0x05, 0x30, 0x31, 0x32, 0x42, 0x62, 0x7B, 0x82,
		0x83, 0x87, 0x8A, 0x8B, 0x8E, 0xB3, 0xCB, 0xE5, 0xF5, 0xF4, 0xFB,
	}
	for _, b := range dbfTypes {
		if raw[0] == b {
			return true
		}
	}

	return false
}

// ElfObj matches an object file.
func ElfObj(raw []byte, limit uint32) bool {
	return len(raw) > 17 && ((raw[16] == 0x01 && raw[17] == 0x00) ||
		(raw[16] == 0x00 && raw[17] == 0x01))
}

// ElfExe matches an executable file.
func ElfExe(raw []byte, limit uint32) bool {
	return len(raw) > 17 && ((raw[16] == 0x02 && raw[17] == 0x00) ||
		(raw[16] == 0x00 && raw[17] == 0x02))
}

// ElfLib matches a shared library file.
func ElfLib(raw []byte, limit uint32) bool {
	return len(raw) > 17 && ((raw[16] == 0x03 && raw[17] == 0x00) ||
		(raw[16] == 0x00 && raw[17] == 0x03))
}

// ElfDump matches a core dump file.
func ElfDump(raw []byte, limit uint32) bool {
	return len(raw) > 17 && ((raw[16] == 0x04 && raw[17] == 0x00) ||
		(raw[16] == 0x00 && raw[17] == 0x04))
}

// Dcm matches a DICOM medical format file.
func Dcm(raw []byte, limit uint32) bool {
	return len(raw) > 131 &&
		bytes.Equal(raw[128:132], []byte{0x44, 0x49, 0x43, 0x4D})
}

// Marc matches a MARC21 (MAchine-Readable Cataloging) file.
func Marc(raw []byte, limit uint32) bool {
	// File is at least 24 bytes ("leader" field size).
	if len(raw) < 24 {
		return false
	}

	// Fixed bytes at offset 20.
	if !bytes.Equal(raw[20:24], []byte("4500")) {
		return false
	}

	// First 5 bytes are ASCII digits.
	for i := 0; i < 5; i++ {
		if raw[i] < '0' || raw[i] > '9' {
			return false
		}
	}

	// Field terminator is present.
	return bytes.Contains(raw, []byte{0x1E})
}

// CborSeq matches CBOR sequences
func CborSeq(raw []byte, limit uint32) bool {
	if len(raw) == 0 {
		return false
	}
	offset, i := 0, 0
	ok, oldok := true, true
	for ; ok && offset != len(raw); i++ {
		oldok = ok
		offset, ok = cborHelper(raw, offset)
	}
	if limit == uint32(len(raw)) {
		ok = oldok
	}
	return ok && i > 1
}

func cborHelper(raw []byte, offset int) (int, bool) {
	raw_len := len(raw) - offset
	if raw_len == 0 {
		return 0, false
	}

	mt := uint8(raw[offset] & 0xe0)
	ai := raw[offset] & 0x1f
	val := int(ai)
	offset++

	BgEn := binary.BigEndian
	switch ai {
	case 24:
		if raw_len < 2 {
			return 0, false
		}
		val = int(raw[offset])
		offset++
		if mt == 0xe0 && uint64(raw[offset]) < 32 {
			return 0, false
		}
	case 25:
		if raw_len < 3 {
			return 0, false
		}
		val = int(BgEn.Uint16(raw[offset : offset+2]))
		offset += 2
	case 26:
		if raw_len < 5 {
			return 0, false
		}
		val = int(BgEn.Uint32(raw[offset : offset+4]))
		offset += 4
	case 27:
		if raw_len < 9 {
			return 0, false
		}
		val = int(BgEn.Uint64(raw[offset : offset+8]))
		offset += 8
	case 31:
		switch mt {
		case 0x00, 0x20, 0xc0:
			return 0, false
		case 0xe0:
			return 0, false
		}
	default:
		if ai > 24 { // ie. case 28: case 29: case 30
			return 0, false
		}
	}

	switch mt {
	case 0x40, 0x60:
		if ai == 31 {
			return cborIndefinite(raw, mt, offset)
		}
		if val < 0 || len(raw)-offset < val {
			return 0, false
		}
		offset += val
	case 0x80, 0xa0:
		if ai == 31 {
			return cborIndefinite(raw, mt, offset)
		}
		if val < 0 {
			return 0, false
		}
		count := 1
		if mt == 0xa0 {
			count = 2
		}
		for i := 0; i < val*count; i++ {
			var ok bool
			offset, ok = cborHelper(raw, offset)
			if !ok {
				return 0, false
			}
		}
	case 0xc0:
		return cborHelper(raw, offset)
	default:
		return 0, false
	}
	return offset, true
}

func cborIndefinite(raw []byte, mt uint8, offset int) (int, bool) {
	var ok bool
	i := 0
	for {
		if len(raw) == offset {
			return 0, false
		}
		if raw[offset] == 0xff {
			offset++
			break
		}
		offset, ok = cborHelper(raw, offset)
		if !ok {
			return 0, false
		}
		i++
	}
	if mt == 0xa0 && i%2 == 1 {
		return 0, false
	}
	return offset, true
}
