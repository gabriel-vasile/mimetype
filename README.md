<h1 align="center">
  mimetype
</h1>

<h4 align="center">
  A library for detecting mime types and extensions based on magic numbers
</h4>

## Installation
```bash
go get github.com/gabriel-vasile/mimetype


## Usage
The library exposes three functions you can use in order to detect a file type.
```go
func Detect(in []byte) (mime, extension string) {...}
func DetectReader(r io.Reader) (mime, extension string, err error) {...}
func DetectFile(file string) (mime, extension string, err error) {...}
```
See [Godoc](https://godoc.org/github.com/gabriel-vasile/mimetype) for full reference.

## Structure
<b>mimetype</b> uses an hierarchical structure to keep the matching functions.
This reduces the number of calls needed for detecting the file type. The reason
behind this choice is that there are file formats used as containers for other
file formats. For example, Microsoft office files are just zip archives,
containing specific metadata files.
<div align="center">
  <img alt="Header" src="mimetype.gif" width="88%">
</div>
