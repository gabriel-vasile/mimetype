module github.com/gabriel-vasile/mimetype

go 1.21

// v1.4.14 had an int overflow causing slice index panic on 32bit arch. #829
retract v1.4.14

// v1.4.4 had a test file detected as malicious by antivirus software. #575
retract v1.4.4
