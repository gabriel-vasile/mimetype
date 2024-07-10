package magic

import (
	"io"
	"reflect"
	"testing"
)

func TestCsv(t *testing.T) {
	tests := []struct {
		name  string
		input string
		limit uint32
		want  bool
	}{

		{
			name:  "csv multiple lines",
			input: "a,b,c\n1,2,3",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Csv([]byte(tt.input), tt.limit); got != tt.want {
				t.Errorf("Csv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTsv(t *testing.T) {
	tests := []struct {
		name  string
		input string
		limit uint32
		want  bool
	}{

		{
			name:  "tsv multiple lines",
			input: "a\tb\tc\n1\t2\t3",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Tsv([]byte(tt.input), tt.limit); got != tt.want {
				t.Errorf("Csv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSv(t *testing.T) {
	tests := []struct {
		name      string
		delimiter byte
		input     string
		limit     uint32
		want      bool
	}{
		{
			name:      "empty",
			delimiter: ',',
			input:     "",
			want:      false,
		},
		{
			name:      "csv single line",
			delimiter: ',',
			input:     "a,b,c",
			want:      false,
		},
		{
			name:      "csv multiple lines",
			delimiter: ',',
			input:     "a,b,c\n1,2,3",
			want:      true,
		},
		{
			name:      "csv with spaces",
			delimiter: ',',
			input:     "  a ,\t\tb,   c\n1, 2 , 3  ",
			want:      true,
		},
		{
			name:      "csv multiple lines under limit",
			delimiter: ',',
			input:     "a,b,c\n1,2,3\n4,5,6",
			limit:     10,
			want:      true,
		},
		{
			name:      "csv multiple lines over limit",
			delimiter: ',',
			input:     "a,b,c\n1,2,3\n4,5,6",
			limit:     1,
			want:      false,
		},
		{
			name:      "csv 2 line with incomplete last line",
			delimiter: ',',
			input:     "a,b,c\n1,2",
			want:      false,
		},
		{
			name:      "csv 3 line with incomplete last line",
			delimiter: ',',
			input:     "a,b,c\na,b,c\n1,2",
			limit:     10,
			want:      true,
		},
		{
			name:      "within quotes",
			delimiter: ',',
			input:     "\"a,b,c\n1,2,3\n4,5,6\"",
			want:      false,
		},
		{
			name:      "partial quotes",
			delimiter: ',',
			input:     "\"a,b,c\n1,2,3\n4,5,6",
			want:      false,
		},
		{
			name:      "has quotes",
			delimiter: ',',
			input:     "\"a\",\"b\",\"c\"\n1,\",\"2,3\n\"4\",5,6",
			want:      true,
		},
		{
			name:      "comma within quotes",
			delimiter: ',',
			input:     "\"a,b\",\"c\"\n1,2,3\n\"4\",5,6",
			want:      false,
		},
		{
			name:      "ignore comments",
			delimiter: ',',
			input:     "#a,b,c\n#1,2,3",
			want:      false,
		},
		{
			name:      "multiple comments at the end of line",
			delimiter: ',',
			input:     "a,b#,c\n1,2#,3",
			want:      true,
		},
		{
			name:      "a non csv line within a csv file",
			delimiter: ',',
			input:     "#comment\nsomething else\na,b,c\n1,2,3",
			want:      false,
		},
		{
			name:      "mixing comments and csv lines",
			delimiter: ',',
			input:     "#comment\na,b,c\n#something else\n1,2,3",
			want:      true,
		},
		{
			name:      "ignore empty lines",
			delimiter: ',',
			input:     "#comment\na,b,c\n\n\n#something else\n1,2,3",
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sv([]byte(tt.input), tt.delimiter, tt.limit); got != tt.want {
				t.Errorf("Csv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepSvReader(t *testing.T) {

	tests := []struct {
		name  string
		input string
		limit uint32
		want  string
	}{
		{
			name:  "multiple lines",
			input: "a,b,c\n1,2,3",
			limit: 0,
			want:  "a,b,c\n1,2,3",
		},
		{
			name:  "limit",
			input: "a,b,c\n1,2,3",
			limit: 5,
			want:  "a,b,c",
		},
		{
			name:  "drop last line",
			input: "a,b,c\na,b,c\na,b,c\n1,2",
			limit: 20,
			want:  "a,b,c\na,b,c\na,b,c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := prepSvReader([]byte(tt.input), tt.limit)
			by, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("prepSvReader() error = %v", err)
			}
			if !reflect.DeepEqual(string(by), tt.want) {
				t.Errorf("prepSvReader() = '%v', want '%v'", string(by), tt.want)
			}
		})
	}
}
