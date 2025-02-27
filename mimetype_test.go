package mimetype

import (
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/magic"
	"github.com/gabriel-vasile/mimetype/internal/prost"
)

func BenchmarkLuaProst(b *testing.B) {
	a := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		prost.Shebang(a)
	}
}
func BenchmarkLua(b *testing.B) {
	a := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		magic.Shebang(a)
	}
}
