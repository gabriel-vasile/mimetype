module github.com/gabriel-vasile/mimetype

go 1.20

require golang.org/x/net v0.31.0

// v1.4.4 had a test file detected as malicious by antivirus software. #575
retract v1.4.4
