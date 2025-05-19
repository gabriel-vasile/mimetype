package csv

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/csv/originalcsv"
	"github.com/gabriel-vasile/mimetype/internal/scan"
)

func TestCSV(t *testing.T) {
	a := ` 1,"HH
a"
a,b
`
	// # 1,2,3
	// # "a","1","1",a
	// # "aaa
	// #     asd",2
	// # "asd''",2,2
	// # a "w,ord",1"2,a","a""sd"
	// #
	// # comment
	//  # comment
	//  	# comment
	// `
	// 	a = `
	// "a","1","1"
	// `
	r := Reader{
		Comma:   ',',
		Comment: '#',
		S:       scan.Bytes(a),
	}
	for {
		l, hasMore := r.ReadLine()
		fmt.Println((l), hasMore)
		fmt.Println()
		if !hasMore {
			break
		}
	}

	re := originalcsv.NewReader(strings.NewReader(a))
	re.ReuseRecord = true
	re.LazyQuotes = true
	re.Comment = '#'
	fmt.Println(re.Read())

}
