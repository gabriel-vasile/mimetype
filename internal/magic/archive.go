package magic

import (
	"bytes"
	"encoding/binary"
)

// SevenZ matches a 7z archive.
func SevenZ(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C})
}

// Gzip matches gzip files based on http://www.zlib.org/rfc-gzip.html#header-trailer.
func Gzip(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x1f, 0x8b})
}

// Fits matches an Flexible Image Transport System file.
func Fits(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{
		0x53, 0x49, 0x4D, 0x50, 0x4C, 0x45, 0x20, 0x20, 0x3D, 0x20,
		0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
		0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x54,
	})
}

// Xar matches an eXtensible ARchive format file.
func Xar(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x78, 0x61, 0x72, 0x21})
}

// Bz2 matches a bzip2 file.
func Bz2(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x42, 0x5A, 0x68})
}

// Ar matches an ar (Unix) archive file.
func Ar(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x21, 0x3C, 0x61, 0x72, 0x63, 0x68, 0x3E})
}

// Deb matches a Debian package file.
func Deb(f *File) bool {
	return offset(f.Head, []byte{
		0x64, 0x65, 0x62, 0x69, 0x61, 0x6E, 0x2D,
		0x62, 0x69, 0x6E, 0x61, 0x72, 0x79,
	}, 8)
}

// Warc matches a Web ARChive file.
func Warc(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("WARC/1.0")) ||
		bytes.HasPrefix(f.Head, []byte("WARC/1.1"))
}

// Cab matches a Microsoft Cabinet archive file.
func Cab(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("MSCF\x00\x00\x00\x00"))
}

// Xz matches an xz compressed stream based on https://tukaani.org/xz/xz-file-format.txt.
func Xz(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00})
}

// Lzip matches an Lzip compressed file.
func Lzip(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x4c, 0x5a, 0x49, 0x50})
}

// RPM matches an RPM or Delta RPM package file.
func RPM(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0xed, 0xab, 0xee, 0xdb}) ||
		bytes.HasPrefix(f.Head, []byte("drpm"))
}

// RAR matches a RAR archive file.
func RAR(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("Rar!\x1A\x07\x00")) ||
		bytes.HasPrefix(f.Head, []byte("Rar!\x1A\x07\x01\x00"))
}

// InstallShieldCab matches an InstallShield Cabinet archive file.
func InstallShieldCab(f *File) bool {
	return len(f.Head) > 7 &&
		bytes.Equal(f.Head[0:4], []byte("ISc(")) &&
		f.Head[6] == 0 &&
		(f.Head[7] == 1 || f.Head[7] == 2 || f.Head[7] == 4)
}

// Zstd matches a Zstandard archive file.
// https://github.com/facebook/zstd/blob/dev/doc/zstd_compression_format.md
func Zstd(f *File) bool {
	if len(f.Head) < 4 {
		return false
	}
	sig := binary.LittleEndian.Uint32(f.Head)
	// Check for Zstandard frames and skippable frames.
	return (sig >= 0xFD2FB522 && sig <= 0xFD2FB528) ||
		(sig >= 0x184D2A50 && sig <= 0x184D2A5F)
}

// CRX matches a Chrome extension file: a zip archive prepended by a package header.
func CRX(f *File) bool {
	const minHeaderLen = 16
	if len(f.Head) < minHeaderLen || !bytes.HasPrefix(f.Head, []byte("Cr24")) {
		return false
	}
	pubkeyLen := binary.LittleEndian.Uint32(f.Head[8:12])
	sigLen := binary.LittleEndian.Uint32(f.Head[12:16])
	zipOffset := minHeaderLen + pubkeyLen + sigLen
	if uint32(len(f.Head)) < zipOffset {
		return false
	}
	return Zip(&File{Head: f.Head[zipOffset:], ReadLimit: f.ReadLimit})
}

// Cpio matches a cpio archive file.
func Cpio(f *File) bool {
	if len(f.Head) < 6 {
		return false
	}
	return binary.LittleEndian.Uint16(f.Head) == 070707 || // binary cpio
		bytes.HasPrefix(f.Head, []byte("070707")) || // portable ASCII cpios
		bytes.HasPrefix(f.Head, []byte("070701")) ||
		bytes.HasPrefix(f.Head, []byte("070702"))
}

// Tar matches a (t)ape (ar)chive file.
// Tar files are divided into 512 bytes records. First record contains a 257
// bytes header padded with NUL.
func Tar(f *File) bool {
	head := f.Head
	const sizeRecord = 512

	// The structure of a tar header:
	// type TarHeader struct {
	// 	Name     [100]byte
	// 	Mode     [8]byte
	// 	Uid      [8]byte
	// 	Gid      [8]byte
	// 	Size     [12]byte
	// 	Mtime    [12]byte
	// 	Chksum   [8]byte
	// 	Linkflag byte
	// 	Linkname [100]byte
	// 	Magic    [8]byte
	// 	Uname    [32]byte
	// 	Gname    [32]byte
	// 	Devmajor [8]byte
	// 	Devminor [8]byte
	// }

	if len(head) < sizeRecord {
		return false
	}
	head = head[:sizeRecord]

	// First 100 bytes of the header represent the file name.
	// Check if file looks like Gentoo GLEP binary package.
	if bytes.Contains(head[:100], []byte("/gpkg-1\x00")) {
		return false
	}

	// Get the checksum recorded into the file.
	recsum := tarParseOctal(head[148:156])
	if recsum == -1 {
		return false
	}
	sum1, sum2 := tarChksum(head)
	return recsum == sum1 || recsum == sum2
}

// tarParseOctal converts octal string to decimal int.
func tarParseOctal(b []byte) int64 {
	// Because unused fields are filled with NULs, we need to skip leading NULs.
	// Fields may also be padded with spaces or NULs.
	// So we remove leading and trailing NULs and spaces to be sure.
	b = bytes.Trim(b, " \x00")

	if len(b) == 0 {
		return -1
	}
	ret := int64(0)
	for _, b := range b {
		if b == 0 {
			break
		}
		if b < '0' || b > '7' {
			return -1
		}
		ret = (ret << 3) | int64(b-'0')
	}
	return ret
}

// tarChksum computes the checksum for the header block b.
// The actual checksum is written to same b block after it has been calculated.
// Before calculation the bytes from b reserved for checksum have placeholder
// value of ASCII space 0x20.
// POSIX specifies a sum of the unsigned byte values, but the Sun tar used
// signed byte values. We compute and return both.
func tarChksum(b []byte) (unsigned, signed int64) {
	for i, c := range b {
		if 148 <= i && i < 156 {
			c = ' ' // Treat the checksum field itself as all spaces.
		}
		unsigned += int64(c)
		signed += int64(int8(c))
	}
	return unsigned, signed
}
