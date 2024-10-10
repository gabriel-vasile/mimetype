package magic

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"testing"
)

func createZip(files []string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	w := zip.NewWriter(buf)

	for _, f := range files {
		_, err := w.Create(f)
		if err != nil {
			return nil, err
		}
	}

	return buf, w.Close()
}

func createZipUncompressed(content *bytes.Buffer) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	w := zip.NewWriter(buf)

	for i := 0; i < 5; i++ {
		file, err := w.CreateHeader(&zip.FileHeader{
			Name:   fmt.Sprintf("file%d", i),
			Method: zip.Store, // Store means 0 compression.
		})
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(file, content); err != nil {
			return nil, err
		}
	}

	return buf, w.Close()
}

func TestZeroZip(t *testing.T) {
	tcases := []struct {
		name  string
		files []string
		xlsx  bool
		docx  bool
		pptx  bool
		jar   bool
	}{{
		name:  "empty zip",
		files: nil,
	}, {
		name:  "no customXml",
		files: []string{"foo", "word/"},
	}, {
		name:  "customXml, but no word/",
		files: []string{"customXml"},
	}, {
		name:  "customXml, and other files, but no word/",
		files: []string{"customXml", "1", "2", "3"},
	}, {
		name:  "customXml, and other files, but word/ is the 7th file", // we only check until 6th file
		files: []string{"customXml", "1", "2", "3", "4", "5", "word/"},
	}, {
		name:  "customXml, word/ xl/ pptx/ after 5 files",
		files: []string{"1", "2", "3", "4", "5", "customXml", "word/", "xl/", "ppt/"},
	}, {
		name:  "customXml, word/",
		files: []string{"customXml", "word/"},
		docx:  true,
	}, {
		name:  "customXml, word/with_suffix",
		files: []string{"customXml", "word/with_suffix"},
		docx:  true,
	}, {
		name:  "customXml, word/",
		files: []string{"customXml", "word/media"},
		docx:  true,
	}, {
		name:  "customXml, xl/",
		files: []string{"customXml", "xl/media"},
		xlsx:  true,
	}, {
		name:  "customXml, ppt/",
		files: []string{"customXml", "ppt/media"},
		pptx:  true,
	}, {
		name:  "META-INF",
		files: []string{"META-INF/MANIFEST.MF"},
		jar:   true,
	}, {
		name:  "1 2 3 4 5 6 META-INF", // we only check first 6 files
		files: []string{"1", "2", "3", "4", "5", "6", "META-INF/MANIFEST.MF"},
		jar:   false,
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			buf, err := createZip(tc.files)
			if err != nil {
				t.Fatal(err)
			}

			docx := Docx(buf.Bytes(), 0)
			xlsx := Xlsx(buf.Bytes(), 0)
			pptx := Pptx(buf.Bytes(), 0)
			jar := Jar(buf.Bytes(), 0)

			if tc.docx != docx || tc.xlsx != xlsx || tc.pptx != pptx || tc.jar != jar {
				t.Errorf(`expected %t %t %t %t;
                got %t %t %t %t`, tc.docx, tc.xlsx, tc.pptx, tc.jar, docx, xlsx, pptx, jar)
			}

			// #400 - xlsx, docx, pptx put as is (compression lvl 0) inside a zip
			// It should continue to get detected as regular zip, not xlsx or docx or pptx.
			uncompressedZip, err := createZipUncompressed(buf)
			if err != nil {
				t.Fatal(err)
			}

			docx = Docx(uncompressedZip.Bytes(), 0)
			xlsx = Xlsx(uncompressedZip.Bytes(), 0)
			pptx = Pptx(uncompressedZip.Bytes(), 0)
			jar = Jar(uncompressedZip.Bytes(), 0)

			if docx || xlsx || pptx || jar {
				t.Errorf(`uncompressedZip: expected false, false, false;
                got %t %t %t %t`, docx, xlsx, pptx, jar)
			}
		})
	}
}
