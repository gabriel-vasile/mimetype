package matchers

import (
	"archive/zip"
	"bytes"
	"path/filepath"
)

func Xlsx(in []byte) bool {
	return checkMsOfficex(in, "xl")
}

func Docx(in []byte) bool {
	return checkMsOfficex(in, "word")
}

func Pptx(in []byte) bool {
	return checkMsOfficex(in, "ppt")
}

// TODO
func Doc(in []byte) bool {
	return false
}

func Ppt(in []byte) bool {
	return false
}

func Xls(in []byte) bool {
	return false
}

func checkMsOfficex(in []byte, folder string) bool {
	reader := bytes.NewReader(in)
	zipr, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return false
	}

	return zipHasFile(zipr, "[Content_Types].xml") && zipHasFolder(zipr, folder)
}

func zipHasFolder(r *zip.Reader, folder string) bool {
	for _, f := range r.File {
		if filepath.Dir(f.Name) == folder {
			return true
		}
	}

	return false
}

func zipHasFile(r *zip.Reader, file string) bool {
	for _, f := range r.File {
		if f.Name == file {
			return true
		}
	}

	return false
}
