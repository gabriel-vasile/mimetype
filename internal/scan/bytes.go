// Package scan has functions for scanning byte slices.
package scan

// Bytes is a byte slice with helper methods for easier scanning.
type Bytes []byte

func (b *Bytes) Advance(n int) bool {
	if n < 0 || len(*b) < n {
		return false
	}
	*b = (*b)[n:]
	return true
}

// TrimLWS trims whitespace from beginning of the bytes.
func (b *Bytes) TrimLWS() {
	firstNonWS := 0
	for ; firstNonWS < len(*b) && isWS((*b)[firstNonWS]); firstNonWS++ {
	}

	*b = (*b)[firstNonWS:]
}

// TrimRWS trims whitespace from the end of the bytes.
func (b *Bytes) TrimRWS() {
	lb := len(*b)
	for lb > 0 && isWS((*b)[lb-1]) {
		*b = (*b)[:lb-1]
		lb--
	}
}

func (b *Bytes) Peek() byte {
	if len(*b) > 0 {
		return (*b)[0]
	}
	return 0
}
func (b *Bytes) Pop() byte {
	if len(*b) > 0 {
		ret := (*b)[0]
		*b = (*b)[1:]
		return ret
	}
	return 0
}

func (b *Bytes) PopUntil(anyChar ...byte) []byte {
	i := 0
	for ; i < len(*b); i++ {
		if equalsAny((*b)[i], anyChar...) {
			break
		}
	}

	prefix := (*b)[:i]
	*b = (*b)[i:]
	return prefix
}

func equalsAny(needle byte, haystack ...byte) bool {
	for _, c := range haystack {
		if needle == c {
			return true
		}
	}
	return false
}

// First line returns the first line from b and advances b with the length of the
// line. One new line character is trimmed after the line if it exists.
func (b *Bytes) FirstLine() []byte {
	lineEnd := 0
	for ; lineEnd < len(*b) && (*b)[lineEnd] != '\n'; lineEnd++ {
	}

	line := (*b)[:lineEnd]
	*b = (*b)[lineEnd:]
	// Strip leading \n from leftover bytes.
	if len(*b) > 0 && (*b)[0] == '\n' {
		*b = (*b)[1:]
	}
	return line
}

func isWS(b byte) bool {
	return b == '\t' || b == '\n' || b == '\x0c' || b == '\r' || b == ' '
}
