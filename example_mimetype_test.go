package mimetype_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

func Example_detect() {
	testBytes := []byte("This random text has a MIME type of text/plain; charset=utf-8.")

	mime := mimetype.Detect(testBytes)
	fmt.Println(mime.Is("text/plain"), mime.String(), mime.Extension())

	mime, err := mimetype.DetectReader(bytes.NewReader(testBytes))
	fmt.Println(mime.Is("text/plain"), mime.String(), mime.Extension(), err)

	mime, err = mimetype.DetectFile("a nonexistent file")
	fmt.Println(mime.Is("application/octet-stream"), mime.String(), os.IsNotExist(err))
	// Output: true text/plain; charset=utf-8 .txt
	// true text/plain; charset=utf-8 .txt <nil>
	// true application/octet-stream true
}

// Considering the definition of a binary file as "a computer file that is not
// a text file", they can differentiated by searching for the text/plain MIME
// in their MIME hierarchy.
func Example_textVsBinary() {
	testBytes := []byte("This random text has a MIME type of text/plain; charset=utf-8.")
	detectedMIME := mimetype.Detect(testBytes)

	isBinary := true
	for mime := detectedMIME; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			isBinary = false
		}
	}

	fmt.Println(isBinary, detectedMIME)
	// Output: false text/plain; charset=utf-8
}

func Example_whitelist() {
	testBytes := []byte("This random text has a MIME type of text/plain; charset=utf-8.")
	allowed := []string{"text/plain", "application/zip", "application/pdf"}
	mime := mimetype.Detect(testBytes)

	if mimetype.EqualsAny(mime.String(), allowed...) {
		fmt.Printf("%s is allowed\n", mime)
	} else {
		fmt.Printf("%s is now allowed\n", mime)
	}
	// Output: text/plain; charset=utf-8 is allowed
}

// Use Extend to add support for a file format which is not detected by mimetype.
//
// https://www.garykessler.net/library/file_sigs.html and
// https://github.com/file/file/tree/master/magic/Magdir
// have signatures for a multitude of file formats.
func Example_extend() {
	foobarDetector := func(raw []byte, limit uint32) bool {
		return bytes.HasPrefix(raw, []byte("foobar"))
	}

	mimetype.Extend(foobarDetector, "text/foobar", ".fb")
	mime := mimetype.Detect([]byte("foobar file content"))

	fmt.Println(mime.String(), mime.Extension())
	// Output: text/foobar .fb
}
