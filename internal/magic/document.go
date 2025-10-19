package magic

import (
	"bytes"
	"encoding/binary"
)

// Fdf matches a Forms Data Format file.
func Fdf(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("%FDF"))
}

// Mobi matches a Mobi file.
func Mobi(f *File) bool {
	return offset(f.Head, []byte("BOOKMOBI"), 60)
}

// Lit matches a Microsoft Lit file.
func Lit(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("ITOLITLS"))
}

// PDF matches a Portable Document Format file.
// The %PDF- header should be the first thing inside the file but many
// implementations don't follow the rule. The PDF spec at Appendix H says the
// signature can be prepended by anything.
// https://bugs.astron.com/view.php?id=446
func PDF(f *File) bool {
	return bytes.Contains(f.Head[:min(len(f.Head), 1024)], []byte("%PDF-"))
}

// DjVu matches a DjVu file.
func DjVu(f *File) bool {
	if len(f.Head) < 12 {
		return false
	}
	if !bytes.HasPrefix(f.Head, []byte{0x41, 0x54, 0x26, 0x54, 0x46, 0x4F, 0x52, 0x4D}) {
		return false
	}
	return bytes.HasPrefix(f.Head[12:], []byte("DJVM")) ||
		bytes.HasPrefix(f.Head[12:], []byte("DJVU")) ||
		bytes.HasPrefix(f.Head[12:], []byte("DJVI")) ||
		bytes.HasPrefix(f.Head[12:], []byte("THUM"))
}

// P7s matches an .p7s signature File (PEM, Base64).
func P7s(f *File) bool {
	// Check for PEM Encoding.
	if bytes.HasPrefix(f.Head, []byte("-----BEGIN PKCS7")) {
		return true
	}
	// Check if DER Encoding is long enough.
	if len(f.Head) < 20 {
		return false
	}
	// Magic Bytes for the signedData ASN.1 encoding.
	startHeader := [][]byte{{0x30, 0x80}, {0x30, 0x81}, {0x30, 0x82}, {0x30, 0x83}, {0x30, 0x84}}
	signedDataMatch := []byte{0x06, 0x09, 0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x07}
	// Check if Header is correct. There are multiple valid headers.
	for i, match := range startHeader {
		// If first bytes match, then check for ASN.1 Object Type.
		if bytes.HasPrefix(f.Head, match) {
			if bytes.HasPrefix(f.Head[i+2:], signedDataMatch) {
				return true
			}
		}
	}

	return false
}

// Lotus123 matches a Lotus 1-2-3 spreadsheet document.
func Lotus123(f *File) bool {
	if len(f.Head) <= 20 {
		return false
	}
	version := binary.BigEndian.Uint32(f.Head)
	if version == 0x00000200 {
		return f.Head[6] != 0 && f.Head[7] == 0
	}

	return version == 0x00001a00 && f.Head[20] > 0 && f.Head[20] < 32
}

// CHM matches a Microsoft Compiled HTML Help file.
func CHM(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("ITSF\003\000\000\000\x60\000\000\000"))
}
