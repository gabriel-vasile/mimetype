package matchers

import (
	"bufio"
	"bytes"
	"strings"
)

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {

	pos := bytes.IndexAny(data, "{}:;,")
	if pos == -1 {
		return 0, nil, nil
	}
	part := data[:pos]
	start := 0
	end := len(part) - 1

	for ; start < len(part) && (bytes.IndexByte([]byte("\n\r\t "), part[start]) != -1); start++ {
	}
	for ; end > 0 && (bytes.IndexByte([]byte("\n\r\t @"), part[end]) != -1); end-- {
	}

	if start >= end {
		return 0, nil, nil
	}
	return len(part) + 1, part[start : end+1], nil
}

// Css matches a cascading style sheet
func Css(in []byte) bool {
	reader := bytes.NewReader(in)
	scanner := bufio.NewScanner(reader)
	scanner.Split(split)

	cssCommonTerms := map[string]bool{
		"a":           true,
		"background":  true,
		"black":       true,
		"body":        true,
		"border":      true,
		"color":       true,
		"div":         true,
		"font-face":   true,
		"font-family": true,
		"font-size":   true,
		"height":      true,
		"html":        true,
		"img":         true,
		"margin":      true,
		"p":           true,
		"padding":     true,
		"sans-serif":  true,
		"screen":      true,
		"serif":       true,
		"span":        true,
		"table":       true,
		"white":       true,
		"width":       true,
	}

	foundCount := 0
	for scanner.Scan() {
		token := strings.ToLower(scanner.Text())
		if cssCommonTerms[token] {
			foundCount++
		}
	}
	return foundCount > 1
}
