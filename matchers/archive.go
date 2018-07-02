// Package matchers holds the matching functions used to find mime types
package matchers

import (
	"archive/zip"
	"bytes"
)

func Zip(in []byte) bool {
	return len(in) > 3 &&
		in[0] == 0x50 && in[1] == 0x4B &&
		(in[2] == 0x3 || in[2] == 0x5 || in[2] == 0x7) &&
		(in[3] == 0x4 || in[3] == 0x6 || in[3] == 0x8)
}

func SevenZ(in []byte) bool {
	return bytes.Equal(in[:6], []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C})
}

func Epub(in []byte) bool {
	if len(in) < 58 {
		return false
	}
	in = in[30:58]

	return bytes.Equal(in, []byte("mimetypeapplication/epub+zip"))
}

func Jar(in []byte) bool {
	reader := bytes.NewReader(in)
	zipr, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return false
	}

	return zipHasFile(zipr, "META-INF/MANIFEST.MF")
}

func Apk(in []byte) bool {
	reader := bytes.NewReader(in)
	zipr, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return false
	}

	return zipHasFile(zipr, "AndroidManifest.xml") &&
		zipHasFile(zipr, "classes.dex") &&
		zipHasFile(zipr, "resources.arsc")
}
