// Package mimetype uses magic number signatures to detect the mime type and
// extension of a file.
package mimetype

import (
	"io"
	"os"

	"github.com/gabriel-vasile/mimetype/matchers"
)

// Detect returns the mime type and extension of the provided byte slice.
func Detect(in []byte) (mime, extension string) {
	n := Root.match(in, Root)
	return n.Mime(), n.Extension()
}

// DetectReader returns the mime type and extension of the byte slice read
// from the provided reader.
func DetectReader(r io.Reader) (mime, extension string, err error) {
	in := make([]byte, matchers.ReadLimit)
	n, err := r.Read(in)
	if err != nil && err != io.EOF {
		return Root.Mime(), Root.Extension(), err
	}
	in = in[:n]

	mime, ext := Detect(in)
	return mime, ext, nil
}

// DetectFile returns the mime type and extension of the provided file.
func DetectFile(file string) (mime, extension string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return Root.Mime(), Root.Extension(), err
	}
	defer f.Close()

	return DetectReader(f)
}
