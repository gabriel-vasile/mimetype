package magic

import (
	"fmt"
	"os"
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/csv"
	"github.com/gabriel-vasile/mimetype/internal/scan"
)

var data = []byte(`
1,2,3
"a","1","1",a
"aaa
    asd",2
"asd''",2,2
a "w,ord",1"2,a","a""sd"

# comment
 # comment
 	# comment
`)

func TestCsv(t *testing.T) {
	files := []string{
		// "/home/gabriel/tmp/csv-test-data/csv/all-empty.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/bad-header-less-fields.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/bad-header-more-fields.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/bad-header-no-header.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/bad-header-wrong-header.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/bad-missing-quote.csv",
		"/home/gabriel/tmp/csv-test-data/csv/bad-quotes-with-unescaped-quote.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/bad-unescaped-quote.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/empty-field.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/empty-one-column.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/header-no-rows.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/header-simple.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/leading-space.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/one-column.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/quotes-empty.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/quotes-with-comma.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/quotes-with-escaped-quote.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/quotes-with-newline.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/quotes-with-space.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/simple-crlf.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/simple-lf.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/trailing-newline.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/trailing-newline-one-field.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/trailing-space.csv",
		// "/home/gabriel/tmp/csv-test-data/csv/utf8.csv",
	}

	for _, f := range files {
		data, _ := os.ReadFile(f)
		csv := Csv(data, 0)
		csv1 := Csv1(data)
		if csv1 != csv {
			fmt.Printf("%t\t%t\t%s\n", csv, csv1, f)
			fmt.Println(string(data))
			fmt.Println("------------------------")
			fmt.Println("------------------------")
			fmt.Println("------------------------")
			fmt.Println("------------------------")
		}
	}

}

func BenchmarkCsv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Csv(data, 0)
	}
}

func BenchmarkCsv1(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Csv1(data)
	}
}

func Csv1(data []byte) bool {
	r := csv.Reader{
		Comma:   ',',
		Comment: '#',
		S:       scan.Bytes(data),
	}

	headerFields, hasMore := r.ReadLine()
	if headerFields < 2 || !hasMore {
		return false
	}
	csvLines := 1 // 1 for header
	fmt.Println(string(r.S))
	for {
		fields, hasMore := r.ReadLine()
		fmt.Println("headers", headerFields, fields)
		if fields != 0 {
			csvLines++
		}
		if fields != headerFields {
			return false
		}
		if !hasMore {
			break
		}
	}

	fmt.Println("csvLines", csvLines)
	return csvLines > 2
}
