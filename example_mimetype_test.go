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
	fmt.Println(mime)

	// Detect the MIME type of a reader.
	reader, _ := os.Open(file) // ignoring error for brevity's sake
	mime, rerr := mimetype.DetectReader(reader)
	fmt.Println(mime, rerr)

	// Detect the MIME type of a file.
	mime, ferr := mimetype.DetectFile(file)
	fmt.Println(mime, ferr)

	// Output: application/pdf
	// application/pdf <nil>
	// application/pdf <nil>
}

// To check if some bytes/reader/file has a specific MIME type, first perform
// a detect on the input and then test against the MIME.
//
// Is method can also be called with MIME aliases.
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
//
// Considering JAR files are just ZIPs containing some metadata files,
// if, for example, you need to tell if the input can be unzipped, go up the
// hierarchy until zip is found or the root is reached.
func Example_parent() {
	detectedMIME, err := mimetype.DetectFile("testdata/jar.jar")

	zip := false
	for mime := detectedMIME; mime != nil; mime = mime.Parent() {
		if mime.Is("application/zip") {
			zip = true
		}
	}

	// zip is true, even if the detected MIME was application/jar.
	fmt.Println(zip, detectedMIME, err)

	// Output: true application/jar <nil>
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

	fmt.Println(mime, err)

	// Output: application/zip <nil>
}

func ExampleDetectReader() {
	data, oerr := os.Open("testdata/zip.zip")
	mime, merr := mimetype.DetectReader(data)

	fmt.Println(mime, oerr, merr)

	// Output: application/zip <nil> <nil>
}

func ExampleDetectFile() {
	mime, err := mimetype.DetectFile("testdata/zip.zip")

	fmt.Println(mime, err)

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
