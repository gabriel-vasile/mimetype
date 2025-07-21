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
		docx:  true,
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
		name:  "manifest file first",
		files: []string{"META-INF/MANIFEST.MF"},
		jar:   true,
	}, {
		name:  "manifest dir first",
		files: []string{"META-INF/"},
		jar:   true,
	}, {
		name:  "manifest second file",
		files: []string{"1", "META-INF/MANIFEST.MF"},
		jar:   false,
	}, {
		name: "ppt/ after 15 files",
		files: []string{
			"[Content_Types].xml",
			"_rels/.rels",
			"customXml/_rels/item1.xml",
			"customXml/_rels/item2.xml.rels",
			"customXml/_rels/item3.xml.rels",
			"customXml/_rels/item4.xml.rels",
			"customXml/item1.xml",
			"customXml/item2.xml",
			"customXml/item3.xml",
			"customXml/itemProps1.xml",
			"customXml/itemProps2.xml",
			"customXml/itemProps3.xml",
			"docProps/app.xml",
			"docProps/core.xml",
			"docProps/custom.xml",
			"ppt/_rels/presentation.xml.rel",
		},
		pptx: true,
	}}

	for i, tc := range tcases {
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
				t.Errorf(`
         docx	xlsx	pptx	jar %d
expected %t	%t	%t	%t;
     got %t	%t	%t	%t`, i, tc.docx, tc.xlsx, tc.pptx, tc.jar, docx, xlsx, pptx, jar)
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
				t.Errorf(`
uncompressedZip: docx	xlsx	pptx	jar %d
        expected false	false	false	false
             got %t	%t	%t	%t`, i, docx, xlsx, pptx, jar)
			}
		})
	}
}
