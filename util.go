package mimetype

import (
	"fmt"
	"io"
)

// IsBinary determines if the input data is binary by first seeing if it has
// text/plain in its type tree. This logic might change slightly if we find a
// better mechanism to do this in the future. NOTE: This differs from the Detect
// function signature symmetry that it wraps because it includes the possible
// error value in its return signature.
func IsBinary(in []byte) (bool, error) {
	mtype := Detect(in)
	b, err := isBinary(mtype)
	if err != nil {
		// FIXME: it's weird to me that Detect can't error.
		return false, err
	}
	return b, nil
}

// IsBinaryReader determines if the input data is binary by first seeing if it
// has text/plain in its type tree. This logic might change slightly if we find
// a better mechanism to do this in the future. This wraps the DetectReader
// method.
func IsBinaryReader(r io.Reader) (bool, error) {
	mtype, err := DetectReader(r)
	if err != nil {
		return false, err
	}
	return isBinary(mtype)
}

// IsBinaryFile determines if the input data is binary by first seeing if it has
// text/plain in its type tree. This logic might change slightly if we find a
// better mechanism to do this in the future. This wraps the DetectFile method.
func IsBinaryFile(path string) (bool, error) {
	mtype, err := DetectFile(path)
	if err != nil {
		return false, err
	}
	return isBinary(mtype)
}

// isBinary determines if a file is binary by whether or not it has text/plain
// in the tree. This logic might change slightly if we find a better mechanism
// to do this in the future. This logic was taken from the example
// documentation.
func isBinary(detectedMIME *MIME) (bool, error) {
	if detectedMIME == nil {
		return false, fmt.Errorf("got nil input")
	}
	for mtype := detectedMIME; mtype != nil; mtype = mtype.Parent() {
		if mtype.Is("text/plain") {
			return false, nil
		}
	}
	return true, nil
}
