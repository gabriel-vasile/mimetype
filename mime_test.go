package mimetype

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabriel-vasile/mimetype/matchers"
)

const testDataDir = "testdata"

var files = map[string]*Node{
	// archives
	"a.pdf":  Pdf,
	"a.zip":  Zip,
	"a.tar":  Tar,
	"a.xls":  Xls,
	"a.xlsx": Xlsx,
	"a.doc":  Doc,
	"a.docx": Docx,
	"a.ppt":  Ppt,
	"a.pptx": Pptx,
	"a.epub": Epub,
	"a.7z":   SevenZ,
	"a.jar":  Jar,
	"a.gz":   Gzip,

	// images
	"a.png":  Png,
	"a.jpg":  Jpg,
	"a.psd":  Psd,
	"a.webp": Webp,
	"a.tif":  Tiff,
	"a.ico":  Ico,
	"a.bmp":  Bmp,

	// video
	"a.mp4":  Mp4,
	"b.mp4":  Mp4,
	"a.webm": WebM,
	"a.3gp":  ThreeGP,
	"a.3g2":  ThreeG2,
	"a.flv":  Flv,
	"a.avi":  Avi,
	"a.mov":  QuickTime,
	"a.mpeg": Mpeg,
	"a.mkv":  Mkv,

	// audio
	"a.mp3":  Mp3,
	"a.wav":  Wav,
	"a.flac": Flac,
	"a.midi": Midi,
	"a.ape":  Ape,
	"a.aiff": Aiff,
	"a.au":   Au,
	"a.ogg":  Ogg,
	"a.amr":  Amr,
	"a.mpc":  MusePack,
	"a.m4a":  M4a,
	"a.m4b":  AMp4,

	// source code
	"a.html": Html,
	"a.xml":  Xml,
	"a.svg":  Svg,
	"b.svg":  Svg,
	"a.txt":  Txt,
	"a.php":  Php,
	"a.ps":   Ps,
	"a.json": Json,
	"a.rtf":  Rtf,
	"a.js":   Js,
	"a.lua":  Lua,
	"a.pl":   Perl,
	"a.py":   Python,
	"a.tcl":  Tcl,

	// binary
	"a.class": Class,
	"a.swf":   Swf,
	"a.crx":   Crx,

	// fonts
	"a.woff":  Woff,
	"a.woff2": Woff2,
}

func TestMatching(t *testing.T) {
	errStr := "File: %s; Mime: %s != DetectedMime: %s; err: %v"
	for fName, node := range files {
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		if dMime, _ := Detect(data); dMime != node.Mime() {
			t.Errorf(errStr, fName, node.Mime(), dMime, nil)
		}

		if _, err := f.Seek(io.SeekStart, 0); err != nil {
			t.Errorf(errStr, fName, node.Mime(), Root.Mime(), err)
		}

		if dMime, _, err := DetectReader(f); dMime != node.Mime() {
			t.Errorf(errStr, fName, node.Mime(), dMime, err)
		}
		f.Close()

		if dMime, _, err := DetectFile(fileName); dMime != node.Mime() {
			t.Errorf(errStr, fName, node.Mime(), dMime, err)
		}
	}
}

func TestFaultyInput(t *testing.T) {
	inexistent := "inexistent.file"
	if _, _, err := DetectFile(inexistent); err == nil {
		t.Errorf("%s should not match successfully", inexistent)
	}

	f, _ := os.Open(inexistent)
	if _, _, err := DetectReader(f); err == nil {
		t.Errorf("%s reader should not match successfully", inexistent)
	}
}

func TestEmptyInput(t *testing.T) {
	if m, _ := Detect([]byte{}); m != "inode/x-empty" {
		t.Errorf("failed to detect empty file")
	}
}

// `foobar` func matches inputs starting with the string "foobar"
// `foobarNode` is the node holding the mimetype and extension to be returned
// when the `foobar` func returns true for an input
func TestAppend(t *testing.T) {
	foobar := func(input []byte) bool {
		return bytes.HasPrefix(input, []byte("foobar"))
	}
	foobarNode := NewNode("text/foobar", "fbExt", foobar)
	fbFile := filepath.Join(testDataDir, "foobar.fb")

	dMime, _, err := DetectFile(fbFile)
	if err != nil {
		t.Fatal(err)
	}
	// even though we tried detecting, at this point the function `foobar`
	// is not yet called because it is not appended in the tree
	if dMime == foobarNode.Mime() {
		t.Fatal("foobar should not get matched")
	}

	// our new node must be appended in the tree
	Txt.Append(foobarNode)

	// the next line calls our `foobar` func which returns true for our test file
	dMime, _, err = DetectFile(fbFile)
	if err != nil {
		t.Fatal(err)
	}
	if dMime != foobarNode.Mime() {
		t.Fatalf("foobar should get matched")
	}
}

func TestGenerateSupportedMimesFile(t *testing.T) {
	f, err := os.OpenFile("supported_mimes.md", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`## Supported MIME types

Extension | MIME type
--------- | --------
`); err != nil {
		t.Fatal(err)
	}
	for _, n := range Root.flatten() {
		ext := n.Extension()
		if ext == "" {
			ext = "n/a"
		}
		str := fmt.Sprintf("**%s** | %s\n", ext, n.Mime())
		if _, err := f.WriteString(str); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.WriteString("\nThis is file automatically generated when running tests.\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkMatchDetect(b *testing.B) {
	files := []string{"a.png", "a.jpg", "a.pdf", "a.zip", "a.docx", "a.doc"}
	data, fLen := [][matchers.ReadLimit]byte{}, len(files)
	for _, f := range files {
		d := [matchers.ReadLimit]byte{}

		file, err := os.Open(filepath.Join(testDataDir, f))
		if err != nil {
			b.Fatal(err)
		}

		io.ReadFull(file, d[:])
		data = append(data, d)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Detect(data[n%fLen][:])
	}
}
