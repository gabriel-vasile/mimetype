package matchers

import "testing"

func TestText(t *testing.T) {
	in := []byte("<!DOCTYPE HTML ")

	if !detect(in, htmlSigs) {
		t.Errorf("html should be matched")
	}
}
