package mimetype

import (
	"io"
	"os"

	"github.com/gabriel-vasile/mimetype/internal/matchers"
)

var matchMIME = map[string][]func([]byte) bool{}

func registerMIME(mime string, matchFunc func([]byte) bool) {
	matchMIME[mime] = append(matchMIME[mime], matchFunc)
}

// IsSupported reports whether the provided MIME type is supported.
func IsSupported(mime string) bool {
	_, ok := matchMIME[mime]
	return ok
}

// Match reports whether the provided MIME type matches the provided byte slice.
func Match(in []byte, mime string) bool {
	matchFuncs, ok := matchMIME[mime]
	if !ok {
		return false
	}
	for _, matchFunc := range matchFuncs {
		if matchFunc(in) {
			return true
		}
	}
	return false
}

// MatchReader reports whetehr the provided MIME type matches the provided reader.
//
// MatchReader assumes the reader offset is at the start. If the input is a
// ReadSeeker you read from before, it should be rewinded before detection:
// reader.Seek(0, io.SeekStart)
//
// To prevent loading entire files into memory, MatchReader reads at most
// matchers.ReadLimit bytes from the reader.
func MatchReader(r io.Reader, mime string) (bool, error) {
	in := make([]byte, matchers.ReadLimit)
	n, err := io.ReadFull(r, in)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, err
	}
	in = in[:n]

	return Match(in, mime), nil
}

// MatchFile reports whether the provided MIME type matches the provided file.
//
// To prevent loading entire files into memory, MatchFile reads at most
// matchers.ReadLimit bytes from the input file.
func MatchFile(file string, mime string) (bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()

	return MatchReader(f, mime)
}

var matchExtension = map[string][]func([]byte) bool{}

func registerExtension(extension string, matchFunc func([]byte) bool) {
	matchExtension[extension] = append(matchExtension[extension], matchFunc)
}

// IsSupportedExtension reports whether the provided MIME type matches the
// provided byte slice.
func IsSupportedExtension(ext string) bool {
	_, ok := matchExtension[ext]
	return ok
}

// MatchExtension reports whether the provided MIME type matches the provided
// byte slice.
func MatchExtension(in []byte, extension string) bool {
	matchFuncs, ok := matchExtension[extension]
	if !ok {
		return false
	}
	for _, matchFunc := range matchFuncs {
		if matchFunc(in) {
			return true
		}
	}
	return false
}

// MatchReaderExtension reports whetehr the provided extension matches the
// provided reader.
//
// MatchReaderExtension assumes the reader offset is at the start. If the input
// is a ReadSeeker you read from before, it should be rewinded before detection:
// reader.Seek(0, io.SeekStart)
//
// To prevent loading entire files into memory, MatchReaderExtension reads at
// most matchers.ReadLimit bytes from the reader.
func MatchReaderExtension(r io.Reader, extension string) (bool, error) {
	in := make([]byte, matchers.ReadLimit)
	n, err := io.ReadFull(r, in)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, err
	}
	in = in[:n]

	return MatchExtension(in, extension), nil
}

// MatchFileExtension reports whether the provided MIME type matches the
// provided file.
//
// To prevent loading entire files into memory, MatchFileExtension reads at most
// matchers.ReadLimit bytes from the input file.
func MatchFileExtension(file string, extension string) (bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()

	return MatchReaderExtension(f, extension)
}
