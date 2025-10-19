package magic

import (
	"bytes"
)

// Flv matches a Flash video file.
func Flv(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte("\x46\x4C\x56\x01"))
}

// Asf matches an Advanced Systems Format file.
func Asf(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{
		0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11,
		0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C,
	})
}

// Rmvb matches a RealMedia Variable Bitrate file.
func Rmvb(f *File) bool {
	return bytes.HasPrefix(f.Head, []byte{0x2E, 0x52, 0x4D, 0x46})
}

// WebM matches a WebM file.
func WebM(f *File) bool {
	return isMatroskaFileTypeMatched(f.Head, "webm")
}

// Mkv matches a mkv file.
func Mkv(f *File) bool {
	return isMatroskaFileTypeMatched(f.Head, "matroska")
}

// isMatroskaFileTypeMatched is used for webm and mkv file matching.
// It checks for .Eß£ sequence. If the sequence is found,
// then it means it is Matroska media container, including WebM.
// Then it verifies which of the file type it is representing by matching the
// file specific string.
func isMatroskaFileTypeMatched(in []byte, flType string) bool {
	if bytes.HasPrefix(in, []byte("\x1A\x45\xDF\xA3")) {
		return isFileTypeNamePresent(in, flType)
	}
	return false
}

// isFileTypeNamePresent accepts the matroska input data stream and searches
// for the given file type in the stream. Return whether a match is found.
// The logic of search is: find first instance of \x42\x82 and then
// search for given string after n bytes of above instance.
func isFileTypeNamePresent(in []byte, flType string) bool {
	ind, maxInd, lenIn := 0, 4096, len(in)
	if lenIn < maxInd { // restricting length to 4096
		maxInd = lenIn
	}
	ind = bytes.Index(in[:maxInd], []byte("\x42\x82"))
	if ind > 0 && lenIn > ind+2 {
		ind += 2

		// filetype name will be present exactly
		// n bytes after the match of the two bytes "\x42\x82"
		n := vintWidth(int(in[ind]))
		if lenIn > ind+n {
			return bytes.HasPrefix(in[ind+n:], []byte(flType))
		}
	}
	return false
}

// vintWidth parses the variable-integer width in matroska containers
func vintWidth(v int) int {
	mask, max, num := 128, 8, 1
	for num < max && v&mask == 0 {
		mask = mask >> 1
		num++
	}
	return num
}

// Mpeg matches a Moving Picture Experts Group file.
func Mpeg(f *File) bool {
	return len(f.Head) > 3 && bytes.HasPrefix(f.Head, []byte{0x00, 0x00, 0x01}) &&
		f.Head[3] >= 0xB0 && f.Head[3] <= 0xBF
}

// Avi matches an Audio Video Interleaved file.
func Avi(f *File) bool {
	return len(f.Head) > 16 &&
		bytes.Equal(f.Head[:4], []byte("RIFF")) &&
		bytes.Equal(f.Head[8:16], []byte("AVI LIST"))
}
