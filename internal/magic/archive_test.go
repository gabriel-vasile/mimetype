package magic

import "testing"

func TestTarParseOctal(t *testing.T) {
	tests := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"0000000\x00", 0, true},
		{" \x0000000\x00", 0, true},
		{" \x0000003\x00", 3, true},
		{"00000000227\x00", 0227, true},
		{"032033\x00 ", 032033, true},
		{"320330\x00 ", 0320330, true},
		{"0000660\x00 ", 0660, true},
		{"\x00 0000660\x00 ", 0660, true},
		{"0123456789abcdef", 0, false},
		{"0123456789\x00abcdef", 0, false},
		{"01234567\x0089abcdef", 342391, true},
		{"0123\x7e\x5f\x264123", 0, false},
	}

	for _, tt := range tests {
		got, err := tarParseOctal([]byte(tt.in))
		ok := err == nil
		if ok != tt.ok {
			if tt.ok {
				t.Errorf("parseOctal(%q): got parsing failure, want success", tt.in)
			} else {
				t.Errorf("parseOctal(%q): got parsing success, want failure", tt.in)
			}
		}
		if got != tt.want {
			t.Errorf("parseOctal(%q): got %d, want %d", tt.in, got, tt.want)
		}
	}
}
