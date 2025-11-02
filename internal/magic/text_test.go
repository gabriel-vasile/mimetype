package magic

import (
	"strings"
	"testing"
)

// Benchmark JSON inputs that can cause slow-downs.
func BenchmarkJSONPathological(b *testing.B) {
	const n = 1000
	hugeArray := []byte(
		strings.Repeat("[1,", n) +
			`2,3,"abc",true,false,null` +
			strings.Repeat("]", n))
	hugeObject := []byte(
		strings.Repeat(`{"a": 1, "b":`, n) +
			`{"c":[2,3,"abc",true,false,null]}` +
			strings.Repeat("}", n))

	b.ReportAllocs()
	for b.Loop() {
		if !JSON(hugeArray, 0) {
			b.Fatal("huge array should be JSON")
		}
		if !JSON(hugeObject, 0) {
			b.Fatal("huge object should be JSON")
		}
		GeoJSON(hugeArray, 0)
		GeoJSON(hugeObject, 0)
		HAR(hugeArray, 0)
		HAR(hugeObject, 0)
		GLTF(hugeArray, 0)
		GLTF(hugeObject, 0)
		NdJSON(hugeArray, 0)
		NdJSON(hugeObject, 0)
	}
}
