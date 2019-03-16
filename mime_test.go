package mimetype

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const testDataDir = "testdata"

var files = map[string]*Node{
	// archives
	"a.pdf":  Pdf,
	"a.zip":  Zip,
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
	if dMime == foobarNode.Mime() {
		t.Fatal("foobar should not get matched")
	}

	Txt.Append(foobarNode)

	dMime, _, err = DetectFile(fbFile)
	if err != nil {
		t.Fatal(err)
	}
	if dMime != foobarNode.Mime() {
		t.Fatalf("foobar should get matched")
	}
}

func TestTreePrint(t *testing.T) {
	t.Logf("\n%s", Root.Tree())
}
