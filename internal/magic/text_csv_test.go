package magic

import (
	"io"
	"math/rand"
	"testing"
)

func BenchmarkCsv(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	data := make([]byte, 4096)
	if _, err := io.ReadFull(r, data); err != io.ErrUnexpectedEOF && err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Csv(data, 0)
	}
}
