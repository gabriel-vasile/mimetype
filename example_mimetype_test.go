package mimetype_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

func Example_detect() {
	testBytes := []byte("This random text should have a MIME type of text/plain; charset=utf-8.")

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
	testBytes := []byte("This random text should have a MIME type of text/plain; charset=utf-8.")
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
	testBytes := []byte("This random text should have a MIME type of text/plain; charset=utf-8.")
	allowed := []string{"text/plain", "application/zip", "application/pdf"}
	mime := mimetype.Detect(testBytes)

	if mimetype.EqualsAny(mime.String(), allowed...) {
		fmt.Printf("%s is allowed\n", mime)
	} else {
		fmt.Printf("%s is now allowed\n", mime)
	}
	// Output: text/plain; charset=utf-8 is allowed
}

// When detecting from an io.Reader, mimetype will read the header of the input.
// This means the reader cannot just be reused (to save the file, for example)
// because the header is now missing from the reader.
//
// If the input is a pure io.Reader, use io.TeeReader, io.MultiReader and bytes.Buffer
// to create a new reader containing the whole unaltered data.
//
// If the input is an io.ReadSeeker, call reader.Seek(0, io.SeekStart) to rewind it.
func Example_reusableReader() {
	// Set header size to 10 bytes for this example.
	mimetype.SetLimit(10)

	testBytes := []byte("This random text should have a MIME type of text/plain; charset=utf-8.")
	inputReader := bytes.NewReader(testBytes)

	// buf will store the 10 bytes mimetype used for detection.
	header := bytes.NewBuffer(nil)

	// After DetectReader, the first 10 bytes are stored in buf.
	mime, err := mimetype.DetectReader(io.TeeReader(inputReader, header))

	// Concatenate back the first 10 bytes.
	// reusableReader now contains the complete, original data.
	reusableReader := io.MultiReader(header, inputReader)
	text, _ := ioutil.ReadAll(reusableReader)
	fmt.Println(mime, bytes.Equal(testBytes, text), err)
	// Output: text/plain; charset=utf-8 true <nil>
}
