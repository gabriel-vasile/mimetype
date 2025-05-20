package magic

import (
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
	i := 0
	csvLines := 1 // 1 for header
	for {
		i++
		fields, hasMore := r.ReadLine()
		if !hasMore && fields == 0 {
			break
		}
		csvLines++
		if fields != headerFields {
			return false
		}
	}

	return csvLines >= 2
}

/*
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/empty.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints_comments.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints_cubed.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints_doesnt_end_in_newline.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints_join.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints_newline_sep.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints_skipline.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/fake_data/ints_squared.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/mimesis_data/persons.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/real_data/2015_StateDepartment.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/real_data/aug15_sample.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/real_data/GDPC1.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/real_data/noaa_storm_events/StormEvents_locations-ftp_v1.0_d2014_c20170718.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/real_data/noaa_storm_events/StormEvents_locations-ftp_v1.0_d2015_c20170718.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/real_data/noaa_storm_events/StormEvents_locations-ftp_v1.0_d2016_c20170816.csv"
	file --mime "/home/gabriel/Downloads/csvs/csv-data-master/real_data/noaa_storm_events/StormEvents_locations-ftp_v1.0_d2017_c20170816.csv"

	file --mime "/home/gabriel/tmp/csv-test-data/csv/all-empty.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/bad-header-less-fields.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/bad-header-more-fields.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/bad-header-no-header.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/bad-header-wrong-header.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/bad-missing-quote.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/bad-quotes-with-unescaped-quote.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/bad-unescaped-quote.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/empty-field.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/empty-one-column.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/header-no-rows.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/header-simple.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/leading-space.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/one-column.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/quotes-empty.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/quotes-with-comma.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/quotes-with-escaped-quote.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/quotes-with-newline.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/quotes-with-space.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/simple-crlf.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/simple-lf.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/trailing-newline.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/trailing-newline-one-field.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/trailing-space.csv"
	file --mime "/home/gabriel/tmp/csv-test-data/csv/utf8.csv"
*/
