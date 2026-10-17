package mimetype_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

func Example_detect() {
	testBytes := []byte("This random text has a MIME type of text/plain; charset=utf-8.")

	mtype := mimetype.Detect(testBytes)
	fmt.Println(mtype.Is("text/plain"), mtype.String(), mtype.Extension())

	mtype, err := mimetype.DetectReader(bytes.NewReader(testBytes))
	fmt.Println(mtype.Is("text/plain"), mtype.String(), mtype.Extension(), err)

	mtype, err = mimetype.DetectFile("a nonexistent file")
	fmt.Println(mtype.Is("application/octet-stream"), mtype.String(), os.IsNotExist(err))
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
	for mtype := detectedMIME; mtype != nil; mtype = mtype.Parent() {
		if mtype.Is("text/plain") {
			isBinary = false
		}
	}

	fmt.Println(isBinary, detectedMIME)
	// Output: false text/plain; charset=utf-8
}

func Example_whitelist() {
	testBytes := []byte("This random text has a MIME type of text/plain; charset=utf-8.")
	allowed := []string{"text/plain", "application/zip", "application/pdf"}
	mtype := mimetype.Detect(testBytes)

	if mimetype.EqualsAny(mtype.String(), allowed...) {
		fmt.Printf("%s is allowed\n", mtype)
	} else {
		fmt.Printf("%s is now allowed\n", mtype)
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

	mimetype.Lookup("text/plain").Extend(foobarDetector, "text/foobar", ".fb")
	mtype := mimetype.Detect([]byte("foobar file content"))

	fmt.Println(mtype.String(), mtype.Extension())
	// Output: text/foobar .fb
}

// MIME structs don't have marshal and unmarshal methods. To encode and
// decode a MIME, call String() and Lookup() manually to marshal and
// unmarshal it as a string value.
func Example_storeInConfiguration() {
	// Store the MIME type as a string in a JSON (or YAML) configuration.
	// Resolve it separately, so unsupported values can be reported explicitly.
	type config struct {
		MIME string `json:"mime"`
	}
	for _, input := range []string{
		`{"mime":"application/json"}`,
		`{"mime":"application/x-not-registered"}`,
	} {
		var cfg config
		if err := json.Unmarshal([]byte(input), &cfg); err != nil {
			fmt.Println("invalid configuration:", err)
			continue
		}
		m := mimetype.Lookup(cfg.MIME)
		if m == nil {
			fmt.Printf("unsupported MIME type: %q\n", cfg.MIME)
			continue
		}
		fmt.Println(m.String(), m.Extension())
		encoded, err := json.Marshal(cfg)
		if err != nil {
			fmt.Println("cannot encode configuration:", err)
			continue
		}
		fmt.Println(string(encoded))
	}
	// Output:
	// application/json .json
	// {"mime":"application/json"}
	// unsupported MIME type: "application/x-not-registered"
}
