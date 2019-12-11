## Examples
 - [Detect MIME type](#detect)
 - [Check against MIME type](#check)
 - [Check base MIME type](#check-parent)

### Detect
Get the MIME type from a slice of bytes, from a reader and from a file.
```go
file := "testdata/pdf.pdf"
reader, _ := os.Open(file)
data, _ := ioutil.ReadFile(file)

// Detecting the same data three times, for the sake of example.
dmime := mimetype.Detect(data)
rmime, rerr := mimetype.DetectReader(reader)
fmime, ferr := mimetype.DetectFile(file)

fmt.Println(dmime, rmime, fmime)
fmt.Println(rerr, ferr)

// Output: application/pdf application/pdf application/pdf
// <nil> <nil>
```

### Check
Test if some file has a specific MIME type. Also accepts MIME type aliases.
```go
mime, err := mimetype.DetectFile("testdata/zip.zip")
// application/x-zip is an alias of application/zip,
// therefore Is returns true both times.
fmt.Println(mime.Is("application/zip"), mime.Is("application/x-zip"), err)

// Output: true true <nil>
```

### Check parent
Test if some file has a specific base MIME type. First perform a detect on the input, then navigate parents until the base MIME type if found.
```go
// Ex: if you are interested in text/plain and all of its subtypes:
// text/html, text/xml, text/csv, etc.
mime, err := mimetype.DetectFile("testdata/html.html")

isText := false
for ; mime != nil; mime = mime.Parent() {
	if mime.Is("text/plain") {
		isText = true
	}
}

// isText is true, even if the detected MIME was text/html.
fmt.Println(isText, err)

// Output: true <nil>
```
