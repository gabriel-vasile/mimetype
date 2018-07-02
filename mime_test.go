package mimetype

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabriel-vasile/mimetype/matchers"
)

const testDataDir = "testdata"

var files = map[string]Node{
	"a.pdf":  Pdf,
	"a.zip":  Zip,
	"a.xlsx": Xlsx,
	"a.docx": Docx,
	"a.pptx": Pptx,
	"a.epub": Epub,
	"a.7z":   SevenZ,
	"a.jar":  Jar,
	"a.apk":  Apk,

	"a.png": Png,
	"a.psd": Psd,

	"a.mp4":  Mp4,
	"a.webm": WebM,
	"a.3gp":  ThreeGP,
	"a.flv":  Flv,

	"a.html": Html,
	"a.xml":  Xml,
	"a.txt":  Txt,
	"a.ps":   Ps,
}

func TestMatching(t *testing.T) {
	errStr := "Mime: %s; DetectedMime: %s\nExt: %s; DetectedExt"
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

		if dMime, dExt := Detect(data); dMime != node.Mime() || dExt != node.Extension() {
			t.Errorf(errStr, dMime, node.Mime(), dExt, node.Extension())
		}

		f.Seek(0, 0)
		if dMime, dExt, err := DetectReader(f); dMime != node.Mime() || dExt != node.Extension() {
			t.Errorf(errStr, dMime, node.Mime(), dExt, node.Extension())
		} else if err != nil {
			t.Fatal(err)
		}
		f.Close()

		if dMime, dExt, err := DetectFile(fileName); dMime != node.Mime() || dExt != node.Extension() {
			t.Errorf(errStr, dMime, node.Mime(), dExt, node.Extension())
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func TestAppend(t *testing.T) {
	fooMime := NewNode("foo/foo", "foo", matchers.Dummy)
	barMime := NewNode("bar/bar", "bar", matchers.Dummy)
	foobarFile := []byte("\x11\x12foobar")

	if dMime, _ := Detect(foobarFile); dMime == fooMime.Mime() || dMime == barMime.Mime() {
		t.Fatal(fmt.Errorf("Foo and bar matchers not yet registered"))
	}

	fooMime.Append(barMime)
	Root.Append(fooMime)

	if dMime, _ := Detect(foobarFile); dMime != barMime.Mime() {
		t.Fatal(fmt.Errorf("Bar matcher should trigger successfully"))
	}
}

func TestTreePrint1(_ *testing.T) {
	fmt.Println(Root.Tree())
}
