package matchers

import "bytes"

// OwlXml matches an owl/xml file.
func OwlXml(in []byte) bool {
	return bytes.Contains(in, []byte{0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x77, 0x77, 0x77, 0x2e, 0x77, 0x33,
		0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x32, 0x30, 0x30, 0x32, 0x2f, 0x30, 0x37, 0x2f, 0x6f, 0x77, 0x6c, 0x23})
}
