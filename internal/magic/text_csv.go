package magic

import (
	"github.com/gabriel-vasile/mimetype/internal/csv"
	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// CSV matches a comma-separated values file.
func CSV(f *File) bool {
	return sv(f.Head, ',', f.ReadLimit)
}

// TSV matches a tab-separated values file.
func TSV(f *File) bool {
	return sv(f.Head, '\t', f.ReadLimit)
}

func sv(in []byte, comma byte, limit uint32) bool {
	s := scan.Bytes(in)
	s.DropLastLine(limit)
	r := csv.NewParser(comma, '#', s)

	headerFields, _, hasMore := r.CountFields(false)
	if headerFields < 2 || !hasMore {
		return false
	}
	csvLines := 1 // 1 for header
	for {
		fields, _, hasMore := r.CountFields(false)
		if !hasMore && fields == 0 {
			break
		}
		csvLines++
		if fields != headerFields {
			return false
		}
		if csvLines >= 10 {
			return true
		}
	}

	return csvLines >= 2
}
