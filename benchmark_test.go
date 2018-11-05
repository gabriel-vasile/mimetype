package mimetype

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var testresult string

func BenchmarkDetectFile(b *testing.B) {
	type benchArgs struct {
		name string
		mime string
	}
	var benchTests []benchArgs
	for fName, node := range files {
		benchTests = append(benchTests,
			benchArgs{
				name: filepath.Join(testDataDir, fName),
				mime: node.Mime(),
			})
	}
	for _, bb := range benchTests {
		b.Run(bb.name, func(b *testing.B) {
			var dMime string
			for n := 0; n < b.N; n++ {
				dMime, _, _ = DetectFile(bb.name)
			}
			testresult = dMime
		})
		if bb.mime != testresult {
			fmt.Println("wanted : " + bb.mime + " got : " + testresult)
		}
	}
}

func BenchmarkDetect(b *testing.B) {
	type benchArgs struct {
		name string
		mime string
		data []byte
	}
	var benchTests []benchArgs
	for fName, node := range files {
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			continue
		}
		fdata, err := ioutil.ReadAll(f)
		if err != nil {
			continue
		}

		benchTests = append(benchTests,
			benchArgs{
				name: fileName,
				mime: node.Mime(),
				data: fdata,
			})
	}
	for _, bb := range benchTests {
		b.Run(bb.name, func(b *testing.B) {
			var dMime string
			for n := 0; n < b.N; n++ {
				dMime, _ = Detect(bb.data)
			}
			testresult = dMime
		})
		if bb.mime != testresult {
			fmt.Println("wanted : " + bb.mime + " got : " + testresult)
		}
	}
}

func BenchmarkDetectReader(b *testing.B) {
	type benchArgs struct {
		name string
		mime string
		fl   *os.File
	}
	var benchTests []benchArgs
	for fName, node := range files {
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			continue
		}

		benchTests = append(benchTests,
			benchArgs{
				name: fileName,
				mime: node.Mime(),
				fl:   f,
			})
	}
	for _, bb := range benchTests {
		b.Run(bb.name, func(b *testing.B) {
			var dMime string
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				bb.fl.Seek(0, 0)
				b.StartTimer()
				dMime, _, _ = DetectReader(bb.fl)
			}
			testresult = dMime
		})
		bb.fl.Close()
		if bb.mime != testresult {
			fmt.Println("wanted : " + bb.mime + " got : " + testresult)
		}
	}
}
