package mimetype_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

// To check if some bytes/reader/file has a specific MIME type, first perform
// a detect on the input and then test against the MIME.
func Example_check() {
	mime, err := mimetype.DetectFile("testdata/zip.zip")
	// application/x-zip is an alias of application/zip,
	// therefore Is returns true both times.
	fmt.Println(mime.Is("application/zip"), mime.Is("application/x-zip"), err)

	// Output: true true <nil>
}

// To check if some bytes/reader/file has a base MIME type, first perform
// a detect on the input and then navigate the parents until the base MIME type
// is found.
func Example_parent() {
	// text/html is a subclass of text/plain.
	// mime is text/html.
	mime, err := mimetype.DetectFile("testdata/html.html")

	isText := false
	for ; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			isText = true
		}
	}

	fmt.Println(isText, err)

	// Output: true <nil>
}

func ExampleDetect() {
	data, err := ioutil.ReadFile("testdata/zip.zip")
	mime := mimetype.Detect(data)

	fmt.Println(mime.String(), err)

	// Output: application/zip <nil>
}

func ExampleDetectReader() {
	data, oerr := os.Open("testdata/zip.zip")
	mime, merr := mimetype.DetectReader(data)

	fmt.Println(mime.String(), oerr, merr)

	// Output: application/zip <nil> <nil>
}

func ExampleDetectFile() {
	mime, err := mimetype.DetectFile("testdata/zip.zip")

	fmt.Println(mime.String(), err)

	// Output: application/zip <nil>
}

func ExampleMIME_Is() {
	mime, err := mimetype.DetectFile("testdata/pdf.pdf")

	pdf := mime.Is("application/pdf")
	xpdf := mime.Is("application/x-pdf")
	txt := mime.Is("text/plain")
	fmt.Println(pdf, xpdf, txt, err)

	// Output: true true false <nil>
}
