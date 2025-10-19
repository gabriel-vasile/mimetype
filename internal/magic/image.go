package magic

import "bytes"

// Png matches a Portable Network Graphics file.
// https://www.w3.org/TR/PNG/
func Png(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
}

// Apng matches an Animated Portable Network Graphics file.
// https://wiki.mozilla.org/APNG_Specification
func Apng(f *File) bool {
	return offset(f.Head, []byte("acTL"), 37)
}

// Jpg matches a Joint Photographic Experts Group file.
func Jpg(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0xFF, 0xD8, 0xFF})
}

// Jp2 matches a JPEG 2000 Image file (ISO 15444-1).
func Jp2(f *File) bool {
	return jpeg2k(f.Head, []byte{0x6a, 0x70, 0x32, 0x20})
}

// Jpx matches a JPEG 2000 Image file (ISO 15444-2).
func Jpx(f *File) bool {
	return jpeg2k(f.Head, []byte{0x6a, 0x70, 0x78, 0x20})
}

// Jpm matches a JPEG 2000 Image file (ISO 15444-6).
func Jpm(f *File) bool {
	return jpeg2k(f.Head, []byte{0x6a, 0x70, 0x6D, 0x20})
}

// Gif matches a Graphics Interchange Format file.
func Gif(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("GIF87a")) ||
		bytes.HasPrefix(f.Head, []byte("GIF89a"))
}

// Bmp matches a bitmap image file.
func Bmp(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x42, 0x4D})
}

// Ps matches a PostScript file.
func Ps(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("%!PS-Adobe-"))
}

// Psd matches a Photoshop Document file.
func Psd(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("8BPS"))
}

// Ico matches an ICO file.
func Ico(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x00, 0x00, 0x01, 0x00}) ||
		bytes.HasPrefix(f.Head, []byte{0x00, 0x00, 0x02, 0x00})
}

// Icns matches an ICNS (Apple Icon Image format) file.
func Icns(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("icns"))
}

// Tiff matches a Tagged Image File Format file.
func Tiff(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x49, 0x49, 0x2A, 0x00}) ||
		bytes.HasPrefix(f.Head, []byte{0x4D, 0x4D, 0x00, 0x2A})
}

// Bpg matches a Better Portable Graphics file.
func Bpg(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x42, 0x50, 0x47, 0xFB})
}

// Xcf matches GIMP image data.
func Xcf(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("gimp xcf"))
}

// Pat matches GIMP pattern data.
func Pat(f *File) bool {
	return offset(f.Head, []byte("GPAT"), 20)
}

// Gbr matches GIMP brush data.
func Gbr(f *File) bool {
	return offset(f.Head, []byte("GIMP"), 20)
}

// Hdr matches Radiance HDR image.
// https://web.archive.org/web/20060913152809/http://local.wasp.uwa.edu.au/~pbourke/dataformats/pic/
func Hdr(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("#?RADIANCE\n"))
}

// Xpm matches X PixMap image data.
func Xpm(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x2F, 0x2A, 0x20, 0x58, 0x50, 0x4D, 0x20, 0x2A, 0x2F})
}

// Jxs matches a JPEG XS coded image file (ISO/IEC 21122-3).
func Jxs(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x00, 0x00, 0x00, 0x0C, 0x4A, 0x58, 0x53, 0x20, 0x0D, 0x0A, 0x87, 0x0A})
}

// Jxr matches Microsoft HD JXR photo file.
func Jxr(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x49, 0x49, 0xBC, 0x01})
}

func jpeg2k(head []byte, sig []byte) bool {
	if len(head) < 24 {
		return false
	}

	if !bytes.Equal(head[4:8], []byte{0x6A, 0x50, 0x20, 0x20}) &&
		!bytes.Equal(head[4:8], []byte{0x6A, 0x50, 0x32, 0x20}) {
		return false
	}
	return bytes.Equal(head[20:24], sig)
}

// Webp matches a WebP file.
func Webp(f *File) bool {
	return len(f.Head) > 12 &&
		bytes.Equal(f.Head[0:4], []byte("RIFF")) &&
		bytes.Equal(f.Head[8:12], []byte{0x57, 0x45, 0x42, 0x50})
}

// Dwg matches a CAD df.Heading file.
func Dwg(f *File) bool {
	if len(f.Head) < 6 || f.Head[0] != 0x41 || f.Head[1] != 0x43 {
		return false
	}
	dwgVersions := [][]byte{
		{0x31, 0x2E, 0x34, 0x30},
		{0x31, 0x2E, 0x35, 0x30},
		{0x32, 0x2E, 0x31, 0x30},
		{0x31, 0x30, 0x30, 0x32},
		{0x31, 0x30, 0x30, 0x33},
		{0x31, 0x30, 0x30, 0x34},
		{0x31, 0x30, 0x30, 0x36},
		{0x31, 0x30, 0x30, 0x39},
		{0x31, 0x30, 0x31, 0x32},
		{0x31, 0x30, 0x31, 0x34},
		{0x31, 0x30, 0x31, 0x35},
		{0x31, 0x30, 0x31, 0x38},
		{0x31, 0x30, 0x32, 0x31},
		{0x31, 0x30, 0x32, 0x34},
		{0x31, 0x30, 0x33, 0x32},
	}

	for _, d := range dwgVersions {
		if bytes.Equal(f.Head[2:6], d) {
			return true
		}
	}

	return false
}

// Jxl matches JPEG XL image file.
func Jxl(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0xFF, 0x0A}) ||
		bytes.HasPrefix(f.Head, []byte("\x00\x00\x00\x0cJXL\x20\x0d\x0a\x87\x0a"))
}

// DXF matches Df.Heading Exchange Format AutoCAD file.
// There does not seem to be a clear specification and the files in the wild
// differ wildly.
// https://images.autodesk.com/adsk/files/autocad_2012_pdf_dxf-reference_enu.pdf
//
// I collected these signatures by downloading a few dozen files from
// http://cd.textfiles.com/amigaenv/DXF/OBJEKTE/ and
// https://sembiance.com/fileFormatSamples/poly/dxf/ and then
// xxd -l 16 {} | sort | uniq.
// These signatures are only for the ASCII version of DXF. There is a binary version too.
func DXF(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("  0\x0ASECTION\x0A")) ||
		bytes.HasPrefix(f.Head, []byte("  0\x0D\x0ASECTION\x0D\x0A")) ||
		bytes.HasPrefix(f.Head, []byte("0\x0ASECTION\x0A")) ||
		bytes.HasPrefix(f.Head, []byte("0\x0D\x0ASECTION\x0D\x0A"))
}
