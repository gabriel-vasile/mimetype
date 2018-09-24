package matchers

import "bytes"

func Png(in []byte) bool {
	return bytes.Equal(in[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
}

func Jpg(in []byte) bool {
	return bytes.Equal(in[:3], []byte{0xFF, 0xD8, 0xFF})
}

func Gif(in []byte) bool {
	return bytes.HasPrefix(in, []byte("GIF87a")) ||
		bytes.HasPrefix(in, []byte("GIF89a"))
}

func Webp(in []byte) bool {
	return len(in) > 11 &&
		bytes.Equal(in[0:4], []byte{0x52, 0x49, 0x46, 0x46}) &&
		bytes.Equal(in[8:12], []byte{0x57, 0x45, 0x42, 0x50})
}

func Bmp(in []byte) bool {
	return len(in) > 1 &&
		in[0] == 0x42 &&
		in[1] == 0x4D
}

func Ps(in []byte) bool {
	return bytes.HasPrefix(in, []byte("%!PS-Adobe-"))
}

func Psd(in []byte) bool {
	return bytes.HasPrefix(in, []byte("8BPS"))
}

func Ico(in []byte) bool {
	return len(in) > 3 &&
		in[0] == 0x00 && in[1] == 0x00 &&
		in[2] == 0x01 && in[3] == 0x00
}

func Tiff(in []byte) bool {
	return bytes.Equal(in[:4], []byte{0x49, 0x49, 0x2A, 0x00}) ||
		bytes.Equal(in[:4], []byte{0x4D, 0x4D, 0x00, 0x2A})
}
