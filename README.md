<h1 align="center">
  mimetype
</h1>

<h4 align="center">
  A package for detecting mime types and extensions based on magic numbers
</h4>
<h6 align="center">
  No bindings, all written in pure go
</h6>

<p align="center">
  <a href="https://travis-ci.org/gabriel-vasile/mimetype">
    <img alt="Build Status" src="https://travis-ci.org/gabriel-vasile/mimetype.svg?branch=master">
  </a>
  <a href="https://godoc.org/github.com/gabriel-vasile/mimetype">
    <img alt="Documentation" src="https://godoc.org/github.com/gabriel-vasile/mimetype?status.svg">
  </a>
  <a href="https://goreportcard.com/report/github.com/gabriel-vasile/mimetype">
    <img alt="Go report card" src="https://goreportcard.com/badge/github.com/gabriel-vasile/mimetype">
  </a>
  <a href="https://coveralls.io/github/gabriel-vasile/mimetype?branch=master">
    <img alt="Go report card" src="https://coveralls.io/repos/github/gabriel-vasile/mimetype/badge.svg?branch=master">
  </a>
  <a href="LICENSE">
    <img alt="License" src="https://img.shields.io/badge/License-MIT-green.svg">
  </a>
</p>

## Install
```bash
go get github.com/gabriel-vasile/mimetype
```

## Use
The library exposes three functions you can use in order to detect a file type.
See [Godoc](https://godoc.org/github.com/gabriel-vasile/mimetype) for full reference.
```go
func Detect(in []byte) (mime, extension string) {...}
func DetectReader(r io.Reader) (mime, extension string, err error) {...}
func DetectFile(file string) (mime, extension string, err error) {...}
```
When detecting from a `ReadSeeker` interface, such as `os.File`, make sure
to reset the offset of the reader to the beginning if needed:
```go
_, err = file.Seek(io.SeekStart, 0)
```

## Extend
If, for example, you need to detect the **"text/foobar"** mime, for text files
containing the string "foobar" at the start of their first line:
 - create the matching function
    ```go
	foobar := func(input []byte) bool {
		return bytes.HasPrefix(input, []byte("foobar"))
	}
    ```
 - create the mime type node
    ```go
    foobarNode := mimetype.NewNode("text/foobar", "fbExt", foobar)
    ````
 - append the new node in the tree
    ```go
    mimetype.Txt.Append(foobarNode)
    ```
 - detect
    ```go
	mime, extension := mimetype.Detect([]byte("foobar\nfoo foo bar"))
    ```
See [TestAppend](mime_test.go) for a working example.
See [Contributing](CONTRIBUTING.md) if you consider the missing mime type should be included in the library by default.

## Supported mimes
##### Application
Pdf, Xlsx, Docx, Pptx, Epub, Doc, Ppt, Xls, Ps, Psd, Ogg,
JavaScript, Python, Lua, Perl, Tcl
##### Archive
7Z, Zip, Jar, Apk, Tar
##### Image
Png, Jpg, Gif, Webp Tiff
##### Audio
Mp3, Flac, Midi, Ape, MusePack, Wav, Aiff, Au, Amr, M4a, Mp4
##### Video
Mp4, WebM, Mpeg, Quicktime, 3gp, 3g2, Avi, Flv, Mkv
##### Text
Txt, Html, Xml, Php, Json
##### Binary
Class, Swf
##### Font
Woff, Woff2

## Structure
**mimetype** uses an hierarchical structure to keep the matching functions.
This reduces the number of calls needed for detecting the file type. The reason
behind this choice is that there are file formats used as containers for other
file formats. For example, Microsoft office files are just zip archives,
containing specific metadata files.
<div align="center">
  <img alt="structure" src="mimetype.gif" width="88%">
</div>

## Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md).
