package mimetype_test

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/gabriel-vasile/mimetype"
)

func TestMatch(t *testing.T) {
	errStr := "File: %s; Expected: %s; Supported: %v; Matched: %v; err: %v"
	for fName, expected := range files {
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		supported := mimetype.IsSupported(expected)
		if !supported {
			t.Errorf(errStr, fName, expected, supported, false, nil)
		}

		if matched := mimetype.Match(data, expected); !matched {
			t.Errorf(errStr, fName, expected, supported, matched, nil)
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		if matched, err := mimetype.MatchReader(f, expected); !matched {
			t.Errorf(errStr, fName, expected, supported, matched, err)
		}
		f.Close()

		if matched, err := mimetype.MatchFile(fileName, expected); !matched {
			t.Errorf(errStr, fName, expected, supported, matched, err)
		}
	}
}

func TestMatchExtension(t *testing.T) {
	errStr := "File: %s; Expected: %s; Supported: %v; Matched: %v; err: %v"
	for fName := range files {
		expected := path.Ext(fName)
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		supported := mimetype.IsSupportedExtension(expected)
		if !supported {
			t.Errorf(errStr, fName, expected, supported, false, nil)
		}

		if matched := mimetype.MatchExtension(data, expected); !matched {
			t.Errorf(errStr, fName, expected, supported, matched, nil)
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		if matched, err := mimetype.MatchReaderExtension(f, expected); !matched {
			t.Errorf(errStr, fName, expected, supported, matched, err)
		}
		f.Close()

		if matched, err := mimetype.MatchFileExtension(fileName, expected); !matched {
			t.Errorf(errStr, fName, expected, supported, matched, err)
		}
	}
}

func TestMatchFaultyInput(t *testing.T) {
	inexistent := "inexistent.file"
	inexistentType := "inexistent/file-type"
	if _, err := mimetype.MatchFile(inexistent, inexistentType); err == nil {
		t.Errorf("%s should not match %s successfully", inexistent, inexistentType)
	}

	f, _ := os.Open(inexistent)
	if _, err := mimetype.MatchReader(f, inexistentType); err == nil {
		t.Errorf("%s reader should not match %s successfully", inexistent, inexistentType)
	}

	data := []byte{'f', 'i', 'l', 'e'}
	if mimetype.Match(data, inexistentType) {
		t.Errorf("%s data should not match %s successfully", inexistent, inexistentType)
	}
	if mimetype.Match(data, "image/jpeg") {
		t.Errorf("%s data should not match image/jpeg successfully", inexistent)
	}
}

func TestMatchExtensionFaultyInput(t *testing.T) {
	inexistent := "inexistent.file"
	inexistentExt := path.Ext(inexistent)
	if _, err := mimetype.MatchFileExtension(inexistent, inexistentExt); err == nil {
		t.Errorf("%s should not match %s successfully", inexistent, inexistentExt)
	}

	f, _ := os.Open(inexistent)
	if _, err := mimetype.MatchReaderExtension(f, inexistentExt); err == nil {
		t.Errorf("%s reader should not match %s successfully", inexistent, inexistentExt)
	}

	data := []byte{'f', 'i', 'l', 'e'}
	if mimetype.MatchExtension(data, inexistentExt) {
		t.Errorf("%s data should not match %s successfully", inexistent, inexistentExt)
	}
	if mimetype.MatchExtension(data, ".jpg") {
		t.Errorf("%s data should not match image/jpeg successfully", inexistent)
	}
}

func BenchmarkMatchSliceTar(b *testing.B) {
	tar, err := ioutil.ReadFile("testdata/tar.tar")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.MatchExtension(tar, ".tar")
	}
}

func BenchmarkMatchSliceZip(b *testing.B) {
	zip, err := ioutil.ReadFile("testdata/zip.zip")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.MatchExtension(zip, ".zip")
	}
}

func BenchmarkMatchSliceJpeg(b *testing.B) {
	jpeg, err := ioutil.ReadFile("testdata/jpg.jpg")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Match(jpeg, ".jpeg")
	}
}

func BenchmarkMatchSliceGif(b *testing.B) {
	gif, err := ioutil.ReadFile("testdata/gif.gif")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Match(gif, ".gif")
	}
}

func BenchmarkMatchSlicePng(b *testing.B) {
	png, err := ioutil.ReadFile("testdata/png.png")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Match(png, ".png")
	}
}
