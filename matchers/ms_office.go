package matchers

import (
	"archive/zip"
	"bytes"
	"fmt"
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

func Doc(in []byte) bool {
	if len(in) < 515 {
		return false
	}
	head := fmt.Sprintf("%x", in[:8])
	offset512 := fmt.Sprintf("%x", in[512:516])
	if head == "d0cf11e0a1b11ae1" && offset512 == "eca5c100" {
		return true
	}
	return false
}

func Ppt(in []byte) bool {
	if len(in) < 519 {
		return false
	}
	head := fmt.Sprintf("%x", in[:8])
	if head == "d0cf11e0a1b11ae1" {
		offset512 := fmt.Sprintf("%x", in[512:516])
		if offset512 == "a0461df0" || offset512 == "006e1ef0" || offset512 == "0f00e803" {
			return true
		}
		if offset512 == "fdffffff" && fmt.Sprintf("%x", in[518:520]) == "0000" {
			return true
		}
	}
	return false
}

func Xls(in []byte) bool {
	if len(in) < 519 {
		return false
	}
	head := fmt.Sprintf("%x", in[:8])
	offset512 := fmt.Sprintf("%x", in[512:520])
	if head == "d0cf11e0a1b11ae1" && offset512 == "0908100000060500" {
		return true
	}
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
