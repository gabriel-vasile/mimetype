package magic

import (
	"bytes"
	"debug/macho"
	"encoding/binary"
)

// Lnk matches Microsoft lnk binary format.
func Lnk(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x4C, 0x00, 0x00, 0x00, 0x01, 0x14, 0x02, 0x00})
}

// Wasm matches a web assembly File Format file.
func Wasm(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x00, 0x61, 0x73, 0x6D})
}

// Exe matches a Windows/DOS executable file.
func Exe(f *File) bool {
	return len(f.Head) > 1 && f.Head[0] == 0x4D && f.Head[1] == 0x5A
}

// Elf matches an Executable and Linkable Format file.
func Elf(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x7F, 0x45, 0x4C, 0x46})
}

// Nes matches a Nintendo Entertainment system ROM file.
func Nes(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x4E, 0x45, 0x53, 0x1A})
}

// SWF matches an Adobe Flash swf file.
func SWF(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("CWS")) ||
		bytes.HasPrefix(f.Head, []byte("FWS")) ||
		bytes.HasPrefix(f.Head, []byte("ZWS"))
}

// Torrent has bencoded text in the beginning.
func Torrent(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("d8:announce"))
}

// PAR1 matches a parquet file.
func Par1(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x50, 0x41, 0x52, 0x31})
}

// CBOR matches a Concise Binary Object Representation https://cbor.io/
func CBOR(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0xD9, 0xD9, 0xF7})
}

// Java bytecode and Mach-O binaries share the same magic number.
// More info here https://github.com/threatstack/libmagic/blob/master/magic/Magdir/cafebabe
func classOrMachOFat(in []byte) bool {
	// There should be at least 8 bytes for both of them because the only way to
	// quickly distinguish them is by comparing byte at position 7
	if len(in) < 8 {
		return false
	}

	return binary.BigEndian.Uint32(in) == macho.MagicFat
}

// Class matches a java class file.
func Class(f *File) bool {
	return classOrMachOFat(f.Head) && f.Head[7] > 30
}

// MachO matches Mach-O binaries format.
func MachO(f *File) bool {
	if classOrMachOFat(f.Head) && f.Head[7] < 0x14 {
		return true
	}

	if len(f.Head) < 4 {
		return false
	}

	be := binary.BigEndian.Uint32(f.Head)
	le := binary.LittleEndian.Uint32(f.Head)

	return be == macho.Magic32 ||
		le == macho.Magic32 ||
		be == macho.Magic64 ||
		le == macho.Magic64
}

// Dbf matches a dBase file.
// https://www.dbase.com/Knowledgebase/INT/db7_file_fmt.htm
func Dbf(f *File) bool {
	if len(f.Head) < 68 {
		return false
	}

	// 3rd and 4th bytes contain the last update month and day of month.
	if f.Head[2] == 0 || f.Head[2] > 12 || f.Head[3] == 0 || f.Head[3] > 31 {
		return false
	}

	// 12, 13, 30, 31 are reserved bytes and always filled with 0x00.
	if f.Head[12] != 0x00 || f.Head[13] != 0x00 || f.Head[30] != 0x00 || f.Head[31] != 0x00 {
		return false
	}
	// Production MDX flag;
	// 0x01 if a production .MDX file exists for this table;
	// 0x00 if no .MDX file exists.
	if f.Head[28] > 0x01 {
		return false
	}

	// dbf type is dictated by the first byte.
	dbfTypes := []byte{
		0x02, 0x03, 0x04, 0x05, 0x30, 0x31, 0x32, 0x42, 0x62, 0x7B, 0x82,
		0x83, 0x87, 0x8A, 0x8B, 0x8E, 0xB3, 0xCB, 0xE5, 0xF5, 0xF4, 0xFB,
	}
	for _, b := range dbfTypes {
		if f.Head[0] == b {
			return true
		}
	}

	return false
}

// ElfObj matches an object file.
func ElfObj(f *File) bool {
	return len(f.Head) > 17 && ((f.Head[16] == 0x01 && f.Head[17] == 0x00) ||
		(f.Head[16] == 0x00 && f.Head[17] == 0x01))
}

// ElfExe matches an executable file.
func ElfExe(f *File) bool {
	return len(f.Head) > 17 && ((f.Head[16] == 0x02 && f.Head[17] == 0x00) ||
		(f.Head[16] == 0x00 && f.Head[17] == 0x02))
}

// ElfLib matches a shared library file.
func ElfLib(f *File) bool {
	return len(f.Head) > 17 && ((f.Head[16] == 0x03 && f.Head[17] == 0x00) ||
		(f.Head[16] == 0x00 && f.Head[17] == 0x03))
}

// ElfDump matches a core dump file.
func ElfDump(f *File) bool {
	return len(f.Head) > 17 && ((f.Head[16] == 0x04 && f.Head[17] == 0x00) ||
		(f.Head[16] == 0x00 && f.Head[17] == 0x04))
}

// Dcm matches a DICOM medical format file.
func Dcm(f *File) bool {
	return len(f.Head) > 131 &&
		bytes.Equal(f.Head[128:132], []byte{0x44, 0x49, 0x43, 0x4D})
}

// Marc matches a MARC21 (MAchine-Readable Cataloging) file.
func Marc(f *File) bool {
	// File is at least 24 bytes ("leader" field size).
	if len(f.Head) < 24 {
		return false
	}

	// Fixed bytes at offset 20.
	if !bytes.Equal(f.Head[20:24], []byte("4500")) {
		return false
	}

	// First 5 bytes are ASCII digits.
	for i := 0; i < 5; i++ {
		if f.Head[i] < '0' || f.Head[i] > '9' {
			return false
		}
	}

	// Field terminator is present in first 2048 bytes.
	return bytes.Contains(f.Head[:min(2048, len(f.Head))], []byte{0x1E})
}

// GLB matches a glTF model format file.
// GLB is the binary file format representation of 3D models saved in
// the GL transmission Format (glTF).
// GLB uses little endian and its header structure is as follows:
//
//	<-- 12-byte header                             -->
//	| magic            | version          | length   |
//	| (uint32)         | (uint32)         | (uint32) |
//	| \x67\x6C\x54\x46 | \x01\x00\x00\x00 | ...      |
//	| g   l   T   F    | 1                | ...      |
//
// Visit [glTF specification] and [IANA glTF entry] for more details.
//
// [glTF specification]: https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.html
// [IANA glTF entry]: https://www.iana.org/assignments/media-types/model/gltf-binary
func GLB(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("\x67\x6C\x54\x46\x02\x00\x00\x00")) ||
		bytes.HasPrefix(f.Head, []byte("\x67\x6C\x54\x46\x01\x00\x00\x00"))
}

// TzIf matches a Time Zone Information Format (TZif) file.
// See more: https://tools.ietf.org/id/draft-murchison-tzdist-tzif-00.html#rfc.section.3
// Its header structure is shown below:
//
//	+---------------+---+
//	|  magic    (4) | <-+-- version (1)
//	+---------------+---+---------------------------------------+
//	|           [unused - reserved for future use] (15)         |
//	+---------------+---------------+---------------+-----------+
//	|  isutccnt (4) |  isstdcnt (4) |  leapcnt  (4) |
//	+---------------+---------------+---------------+
//	|  timecnt  (4) |  typecnt  (4) |  charcnt  (4) |
func TzIf(f *File) bool {
	// File is at least 44 bytes (header size).
	if len(f.Head) < 44 {
		return false
	}

	if !bytes.HasPrefix(f.Head, []byte("TZif")) {
		return false
	}

	// Field "typecnt" MUST not be zero.
	if binary.BigEndian.Uint32(f.Head[36:40]) == 0 {
		return false
	}

	// Version has to be NUL (0x00), '2' (0x32) or '3' (0x33).
	return f.Head[4] == 0x00 || f.Head[4] == 0x32 || f.Head[4] == 0x33
}
