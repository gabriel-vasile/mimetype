module github.com/gabriel-vasile/mimetype

// Seems go version will get updated automatically in x repos.
// For long time I tried to keep it as low as possible for compatibility reasons,
// but that is not possible anymore. Maybe that's more reason to drop dependency
// on x/net. Another reason is we're using a small portion of the code from x/net.
// https://github.com/golang/go/issues/69095
go 1.23.0

toolchain go1.23.1

require golang.org/x/net v0.39.0

// v1.4.4 had a test file detected as malicious by antivirus software. #575
retract v1.4.4
