package magic

import "testing"

func TestTarParseOctal(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"0000000\x00", 0},
		{" \x0000000\x00", 0},
		{" \x0000003\x00", 3},
		{"00000000227\x00", 0227},
		{"032033\x00 ", 032033},
		{"320330\x00 ", 0320330},
		{"0000660\x00 ", 0660},
		{"\x00 0000660\x00 ", 0660},
		{"0123456789abcdef", -1},
		{"0123456789\x00abcdef", -1},
		{"01234567\x0089abcdef", 01234567},
		{"0123\x7e\x5f\x264123", -1},
	}

	for _, tt := range tests {
		got := tarParseOctal([]byte(tt.in))
		if got != tt.want {
			t.Errorf("parseOctal(%q): got %d, want %d", tt.in, got, tt.want)
		}
	}
}
