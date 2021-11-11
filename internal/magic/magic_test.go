package magic

import "testing"

var magicTests = []struct {
	raw      string
	limit    uint32
	res      bool
	detector Detector
}{
	{`["an incomplete JSON array`, 0, false, JSON},
}

func TestMagic(t *testing.T) {
	for i, tt := range magicTests {
		if got := tt.detector([]byte(tt.raw), tt.limit); got != tt.res {
			t.Errorf("Detector %d error: expected: %t; got: %t", i, tt.res, got)
		}
	}
}
