package mimetype

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const testDataDir = "testdata"

var files = map[string]*Node{
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
	"a.apk":  Apk,

	"a.png":  Png,
	"a.psd":  Psd,
	"a.webp": Webp,
	"a.tif":  Tiff,

	"a.mp4":  Mp4,
	"a.webm": WebM,
	"a.3gp":  ThreeGP,
	"a.flv":  Flv,
	"a.avi":  Avi,
	"a.mov":  Quicktime,
	"a.mpeg": Mpeg,

	"a.mp3":  Mp3,
	"a.wav":  Wav,
	"a.flac": Flac,
	"a.midi": Midi,
	"a.ape":  Ape,
	"a.aiff": Aiff,
	"a.ogg":  Ogg,

	"a.html": Html,
	"a.xml":  Xml,
	"a.txt":  Txt,
	"a.php":  Php,
	"a.ps":   Ps,
}

func TestMatching(t *testing.T) {
	errStr := "Mime: %s != DetectedMime: %s"
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
			t.Errorf(errStr, node.Mime(), dMime)
		}

		f.Seek(0, 0)
		if dMime, _, err := DetectReader(f); dMime != node.Mime() {
			t.Errorf(errStr, node.Mime(), dMime)
		} else if err != nil {
			t.Fatal(err)
		}
		f.Close()

		if dMime, _, err := DetectFile(fileName); dMime != node.Mime() {
			t.Errorf(errStr, node.Mime(), dMime)
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func TestAppend(t *testing.T) {
	Foobar := func(input []byte) bool {
		return bytes.HasPrefix(input, []byte("foobar\n"))
	}
	foobarNode := NewNode("text/foobar", "fbExt", Foobar)
	fbFile := filepath.Join(testDataDir, "foobar.fb")

	dMime, _, err := DetectFile(fbFile)
	if err != nil {
		t.Fatal(err)
	}
	if dMime == foobarNode.Mime() {
		t.Fatal("Foobar should not get matched")
	}

	Txt.Append(foobarNode)

	dMime, _, err = DetectFile(fbFile)
	if err != nil {
		t.Fatal(err)
	}
	if dMime != foobarNode.Mime() {
		t.Fatalf("Foobar should get matched")
	}
}

func TestTreePrint(_ *testing.T) {
	fmt.Println(Root.Tree())
}
