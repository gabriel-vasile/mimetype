## Examples
 - [Detect MIME type](#detect-mime-type-go-playground)
 - [Test against a MIME type](#test-against-a-mime-type-go-playground)
 - [Parent](#parent-go-playground)
 - [Binary file vs text file](#binary-file-vs-text-file-go-playground)

### Detect MIME type [<kbd>Go Playground</kbd>](https://play.golang.org/p/Ti34KSv7HuR)
Get the MIME type from a path to a file.
```go
file := "testdata/pdf.pdf"
mime, err := mimetype.DetectFile(file)
fmt.Println(mime.String(), mime.Extension(), err)
// Output: application/pdf .pdf nil
```
Get the MIME type from a reader.
```go
reader, _ := os.Open(file) // ignoring error for brevity's sake
mime, err := mimetype.DetectReader(reader)
fmt.Println(mime.String(), mime.Extension(), err)
// Output: application/pdf .pdf nil
```

Get the MIME type from a byte slice.
```go
data, _ := ioutil.ReadFile(file) // ignoring error for brevity's sake
mime := mimetype.Detect(data)
fmt.Println(mime.String(), mime.Extension())
// Output: application/pdf .pdf
```

### Test against a MIME type [<kbd>Go Playground</kbd>](https://play.golang.org/p/luAl501AK1q)
Test if a file has a specific MIME type. Different from the string comparison,
e.g.: `mime.String() == "application/zip"`, `mime.Is("application/zip")` method
has the following advantages:
 - handles MIME aliases,
 - is case insensitive,
 - ignores optional MIME parameters,
 - ignores any leading and trailing whitespace.
```go
mime, err := mimetype.DetectFile("testdata/zip.zip")
// application/x-zip is an alias of application/zip,
// therefore Is returns true both times.
fmt.Println(mime.Is("application/zip"), mime.Is("application/x-zip"), err)
// Output: true true <nil>
```

### Parent [<kbd>Go Playground</kbd>](https://play.golang.org/p/F-HBv_Z1Bfj)
Upon detection, it may happen that the returned MIME type is more accurate than
needed.

Suppose we have a text file containing HTML code. Detection performed on this
file will retrieve the `text/html` MIME. If you are interested in telling if
the input can be used as a text file, you can walk up the MIME hierarchy until
`text/plain` is found.

Remember to always check for nil before using the result of the `Parent()` method.
```
           .Parent()              .Parent()
text/html ----------> text/plain ----------> application/octet-stream
```
```go
detectedMIME, err := mimetype.DetectFile("testdata/html.html")

isText := false
for mime := detectedMIME; mime != nil; mime = mime.Parent() {
    if mime.Is("text/plain") {
        isText = true
    }
}

// isText is true, even if the detected MIME was text/html.
fmt.Println(isText, detectedMIME, err)
// Output: true text/html <nil>
```

### Binary file vs text file [<kbd>Go Playground</kbd>](https://play.golang.org/p/CHEFnkn5LQp)
Considering the definition of a binary file as "a computer file that is not
a text file", they can be differentiated by searching for the `text/plain` MIME
in it's MIME hierarchy. This is a reiteration of the [Parent](#parent-go-playground) example.
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
