## Examples
 - [Detect MIME type](#detect)
 - [Check against MIME type](#check)
 - [Check base MIME type](#check-parent)
 - [Binary file vs text file](#binary-file-vs-text-file)

### Detect
Get the MIME type from a slice of bytes, from a reader and from a file.
```go
file := "testdata/pdf.pdf"
reader, _ := os.Open(file)       // ignoring error for brevity's sake
data, _ := ioutil.ReadFile(file) // ignoring error for brevity's sake

dmime := mimetype.Detect(data)
rmime, rerr := mimetype.DetectReader(reader)
fmime, ferr := mimetype.DetectFile(file)

fmt.Println(dmime, rmime, fmime)
fmt.Println(rerr, ferr)

// Output: application/pdf application/pdf application/pdf
// <nil> <nil>
```

### Check
Test if a file has a specific MIME type. Also accepts MIME type aliases.
```go
mime, err := mimetype.DetectFile("testdata/zip.zip")
// application/x-zip is an alias of application/zip,
// therefore Is returns true both times.
fmt.Println(mime.Is("application/zip"), mime.Is("application/x-zip"), err)

// Output: true true <nil>
```

### Check parent
Test if a file has a specific base MIME type. First perform a detect on the
input and then navigate the parents until the base MIME type is found.

Considering JAR files are just ZIPs containing some metadata files,
if, for example, you need to tell if the input can be unzipped, go up the
MIME hierarchy until zip is found or the root is reached.
```go
detectedMIME, err := mimetype.DetectFile("testdata/jar.jar")

zip := false
for mime := detectedMIME; mime != nil; mime = mime.Parent() {
    if mime.Is("application/zip") {
        zip = true
    }
}

// zip is true, even if the detected MIME was application/jar.
fmt.Println(zip, detectedMIME, err)

// Output: true application/jar <nil>
```

### Binary file vs text file
Considering the definition of a binary file as "a computer file that is not
a text file", they can be differentiated by searching for the text/plain MIME
in it's MIME hierarchy.
```go
detectedMIME, err := mimetype.DetectFile("testdata/xml.xml")

isBinary := true
for mime := detectedMIME; mime != nil; mime = mime.Parent() {
    if mime.Is("text/plain") {
        isBinary = false
    }
}

fmt.Println(isBinary, detectedMIME, err)

// Output: false text/xml; charset=utf-8 <nil>
```
