module github.com/gabriel-vasile/mimetype

go 1.20

require golang.org/x/net v0.30.0

// v1.4.4 had a test file that was detected as malicious by antivirus software.
retract v1.4.4
