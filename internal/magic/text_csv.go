package magic

import (
	"github.com/gabriel-vasile/mimetype/internal/csv"
	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// CSV matches a comma-separated values file.
func CSV(raw []byte, limit uint32) bool {
	return sv(raw, ',', limit)
}

// TSV matches a tab-separated values file.
func TSV(raw []byte, limit uint32) bool {
	return sv(raw, '\t', limit)
}

func sv(in []byte, comma byte, limit uint32) bool {
	s := scan.Bytes(in)
	s.DropLastLine(limit)
	r := csv.NewParser(comma, '#', s)

	headerFields, hasMore := r.CountFields()
	if headerFields < 2 || !hasMore {
		return false
	}
	csvLines := 1 // 1 for header
	for {
		fields, hasMore := r.CountFields()
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
