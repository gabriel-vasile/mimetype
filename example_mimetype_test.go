package mimetype_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

// To find the MIME type of some input, perform a detect.
// In addition to the basic Detect,
//  mimetype.Detect([]byte) *MIME
// there are shortcuts for detecting from a reader:
//  mimetype.DetectReader(io.Reader) (*MIME, error)
// or from a file:
//  mimetype.DetectFile(string) (*MIME, error)
func Example_detect() {
	file := "testdata/pdf.pdf"

	// Detect the MIME type of a file stored as a byte slice.
	data, _ := ioutil.ReadFile(file) // ignoring error for brevity's sake
	mime := mimetype.Detect(data)
	fmt.Println(mime.String(), mime.Extension())

	// Detect the MIME type of a reader.
	reader, _ := os.Open(file) // ignoring error for brevity's sake
	mime, rerr := mimetype.DetectReader(reader)
	fmt.Println(mime.String(), mime.Extension(), rerr)

	// Detect the MIME type of a file.
	mime, ferr := mimetype.DetectFile(file)
	fmt.Println(mime.String(), mime.Extension(), ferr)

	// Output: application/pdf .pdf
	// application/pdf .pdf <nil>
	// application/pdf .pdf <nil>
}

// To check if some bytes/reader/file has a specific MIME type, first perform
// a detect on the input and then test against the MIME.
//
// Different from the string comparison,
// e.g.: mime.String() == "application/zip", mime.Is("application/zip") method
// has the following advantages: it handles MIME aliases, is case insensitive,
// ignores optional MIME parameters, and ignores any leading and trailing
// whitespace.
func Example_check() {
	mime, err := mimetype.DetectFile("testdata/zip.zip")
	// application/x-zip is an alias of application/zip,
	// therefore Is returns true both times.
	fmt.Println(mime.Is("application/zip"), mime.Is("application/x-zip"), err)

	// Output: true true <nil>
}

// It may happen that the returned MIME type is more accurate than needed.
//
// Suppose we have a text file containing HTML code. Detection performed on this
// file will retrieve the `text/html` MIME. If you are interested in telling if
// the input can be used as a text file, you can walk up the MIME hierarchy
// until `text/plain` is found.
//
// Remember to always check for nil before using the result of the Parent() method.
//            .Parent()              .Parent()
//  text/html ----------> text/plain ----------> application/octet-stream
func Example_parent() {
	detectedMIME, err := mimetype.DetectFile("testdata/html.html")

	isText := false
	for mime := detectedMIME; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			isText = true
		}
	}

	// isText is true, even if the detected MIME was text/html.
	fmt.Println(isText, detectedMIME, err)

	// Output: true text/html; charset=utf-8 <nil>
}

// Considering the definition of a binary file as "a computer file that is not
// a text file", they can differentiated by searching for the text/plain MIME
// in it's MIME hierarchy.
func Example_textVsBinary() {
	detectedMIME, err := mimetype.DetectFile("testdata/xml.xml")

	isBinary := true
	for mime := detectedMIME; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			isBinary = false
		}
	}

	fmt.Println(isBinary, detectedMIME, err)

	// Output: false text/xml; charset=utf-8 <nil>
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
