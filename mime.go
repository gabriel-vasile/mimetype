// Package mimetype uses magic number signatures to detect and
// to check the MIME type of a file.
package mimetype

import (
	"io"
	"mime"
	"os"

	"github.com/gabriel-vasile/mimetype/internal/matchers"
)

// Detect returns the MIME type and extension of the provided byte slice.
//
// mime is always a valid MIME type, with application/octet-stream as fallback.
// extension is empty string if detected file format does not have an extension.
func Detect(in []byte) (mime, extension string) {
	if len(in) == 0 {
		return "inode/x-empty", ""
	}
	n := root.match(in, root)
	return n.mime, n.extension
}

// DetectReader returns the MIME type and extension
// of the byte slice read from the provided reader.
//
// mime is always a valid MIME type, with application/octet-stream as fallback.
// extension is empty string if detection failed with an error or
// detected file format does not have an extension.
func DetectReader(r io.Reader) (mime, extension string, err error) {
	in := make([]byte, matchers.ReadLimit)
	n, err := r.Read(in)
	if err != nil && err != io.EOF {
		return root.mime, root.extension, err
	}
	in = in[:n]

	mime, extension = Detect(in)
	return mime, extension, nil
}

// DetectFile returns the MIME type and extension of the provided file.
//
// mime is always a valid MIME type, with application/octet-stream as fallback.
// extension is empty string if detection failed with an error or
// detected file format does not have an extension.
func DetectFile(file string) (mime, extension string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return root.mime, root.extension, err
	}
	defer f.Close()

	return DetectReader(f)
}

// Match returns whether the MIME type detected from the slice, or any of its
// aliases, is the same as any of the expected MIME types.
//
// MIME type equality test is done on the "type/subtype" sections, ignores any
// optional MIME parameters, ignores any leading and trailing whitespace,
// and is case insensitive.
// Any error returned is related to the parsing of the expected MIME type.
func Match(in []byte, expectedMimes ...string) (match bool, err error) {
	for i := 0; i < len(expectedMimes); i++ {
		expectedMimes[i], _, err = mime.ParseMediaType(expectedMimes[i])
		if err != nil {
			return false, err
		}
	}

	n := root.match(in, root)
	// This parsing is needed because some detected MIME types contain paramters.
	found, _, err := mime.ParseMediaType(n.mime)
	if err != nil {
		return false, err
	}

	for _, expected := range expectedMimes {
		if expected == found {
			return true, nil
		}
		for _, alias := range n.aliases {
			if alias == expected {
				return true, nil
			}
		}
	}

	return false, nil
}

// Match returns whether the MIME type detected from the reader, or any of its
// aliases, is the same as any of the expected MIME types.
//
// MIME type equality test is done on the "type/subtype" sections, ignores any
// optional MIME parameters, ignores any leading and trailing whitespace,
// and is case insensitive.
func MatchReader(r io.Reader, expectedMimes ...string) (match bool, err error) {
	in := make([]byte, matchers.ReadLimit)
	n, err := r.Read(in)
	if err != nil && err != io.EOF {
		return false, err
	}
	in = in[:n]

	return Match(in, expectedMimes...)
}

// Match returns whether the MIME type detected from the file, or any of its
// aliases, is the same as any of the expected MIME types.
//
// MIME type equality test is done on the "type/subtype" sections, ignores any
// optional MIME parameters, ignores any leading and trailing whitespace,
// and is case insensitive.
func MatchFile(file string, expectedMimes ...string) (match bool, err error) {
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()

	return MatchReader(f, expectedMimes...)
}
