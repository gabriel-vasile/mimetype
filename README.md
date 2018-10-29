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

## Extend
If, for example, you need to detect the **"text/foobar"** mime, for text files
containing the string "foobar" as their first line:
 - create the matching function
    ```go
	foobar := func(input []byte) bool {
		return bytes.HasPrefix(input, []byte("foobar\n"))
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
See [TestAppend](https://github.com/gabriel-vasile/mimetype/blob/master/mime_test.go) for a working example.
See [Contribute](https://github.com/gabriel-vasile/mimetype#contributing) if you consider the missing mime type should be included in the library by default.

## Supported mimes
##### Application
7Z, Zip, Pdf, Xlsx, Docx, Pptx, Epub, Jar, Apk, Doc, Ppt, Xls, Ps, Psd, Ogg,
JavaScript, Python, Lua, Perl
##### Image
Png, Jpg, Gif, Webp Tiff
##### Audio
Mp3, Flac, Midi, Ape, MusePack, Wav, Aiff, Au, Amr
##### Video
Mp4, WebM, Mpeg, Quicktime, ThreeGP, Avi, Flv, Mkv
##### Text
Txt, Html, Xml, Php, Json

## Structure
**mimetype** uses an hierarchical structure to keep the matching functions.
This reduces the number of calls needed for detecting the file type. The reason
behind this choice is that there are file formats used as containers for other
file formats. For example, Microsoft office files are just zip archives,
containing specific metadata files.
<div align="center">
  <img alt="structure" src="mimetype.gif" width="88%">
</div>

## Contribute
See [CONTRIBUTING.md](CONTRIBUTING.md)
