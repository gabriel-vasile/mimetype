package magic

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"sync"

	mcsv "github.com/gabriel-vasile/mimetype/internal/csv"
	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// A bufio.Reader pool to alleviate problems with memory allocations.
var readerPool = sync.Pool{
	New: func() any {
		// Initiate with empty source reader.
		return bufio.NewReader(nil)
	},
}

func newReader(r io.Reader) *bufio.Reader {
	br := readerPool.Get().(*bufio.Reader)
	br.Reset(r)
	return br
}

// Csv matches a comma-separated values file.
func Csv(raw []byte, limit uint32) bool {
	r := mcsv.Reader{
		Comma:   ',',
		Comment: '#',
		S:       scan.Bytes(raw),
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

// Tsv matches a tab-separated values file.
func Tsv(raw []byte, limit uint32) bool {
	return sv(raw, '\t', limit)
}

func sv(in []byte, comma rune, limit uint32) bool {
	s := scan.Bytes(in)
	s.DropLastLine(limit)

	br := newReader(bytes.NewReader(s))
	defer readerPool.Put(br)
	r := csv.NewReader(br)
	r.Comma = comma
	r.ReuseRecord = true
	r.LazyQuotes = true
	r.Comment = '#'

	lines := 0
	for {
		_, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return false
		}
		lines++
	}

	return r.FieldsPerRecord > 1 && lines > 1
}
