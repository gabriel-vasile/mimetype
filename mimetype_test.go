package mimetype_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gabriel-vasile/mimetype"
)

const testDataDir = "testdata"

// test files sorted by the file name in alphabetical order.
var files = map[string]string{
	"3g2.3g2":            "video/3gpp2",
	"3gp.3gp":            "video/3gpp",
	"3mf.3mf":            "application/vnd.ms-package.3dmanufacturing-3dmodel+xml",
	"7z.7z":              "application/x-7z-compressed",
	"a.a":                "application/x-archive",
	"aac.aac":            "audio/aac",
	"aaf.aaf":            "application/octet-stream",
	"accdb.accdb":        "application/x-msaccess",
	"aiff.aiff":          "audio/aiff",
	"amf.amf":            "application/x-amf",
	"amr.amr":            "audio/amr",
	"ape.ape":            "audio/ape",
	"asf.asf":            "video/x-ms-asf",
	"atom.atom":          "application/atom+xml",
	"au.au":              "audio/basic",
	"avi.avi":            "video/x-msvideo",
	"bmp.bmp":            "image/bmp",
	"bpg.bpg":            "image/bpg",
	"bz2.bz2":            "application/x-bzip2",
	"cab.cab":            "application/vnd.ms-cab-compressed",
	"class.class":        "application/x-java-applet",
	"crx.crx":            "application/x-chrome-extension",
	"csv.csv":            "text/csv",
	"cpio.cpio":          "application/x-cpio",
	"dae.dae":            "model/vnd.collada+xml",
	"dbf.dbf":            "application/x-dbf",
	"dcm.dcm":            "application/dicom",
	"deb.deb":            "application/vnd.debian.binary-package",
	"djvu.djvu":          "image/vnd.djvu",
	"doc.1.doc":          "application/msword",
	"doc.doc":            "application/msword",
	"docx.1.docx":        "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"docx.docx":          "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"drpm.rpm":           "application/x-rpm",
	"dwg.1.dwg":          "image/vnd.dwg",
	"dwg.dwg":            "image/vnd.dwg",
	"eot.eot":            "application/vnd.ms-fontobject",
	"epub.epub":          "application/epub+zip",
	"exe.exe":            "application/vnd.microsoft.portable-executable",
	"fdf.fdf":            "application/vnd.fdf",
	"fits.fits":          "application/fits",
	"flac.flac":          "audio/flac",
	"flv.flv":            "video/x-flv",
	"gbr.gbr":            "image/x-gimp-gbr",
	"geojson.1.geojson":  "application/geo+json",
	"geojson.geojson":    "application/geo+json",
	"gif.gif":            "image/gif",
	"glb.glb":            "model/gltf-binary",
	"gml.gml":            "application/gml+xml",
	"gpx.gpx":            "application/gpx+xml",
	"gz.gz":              "application/gzip",
	"hdr.hdr":            "image/vnd.radiance",
	"heic.single.heic":   "image/heic",
	"html.html":          "text/html; charset=utf-8",
	"html.iso88591.html": "text/html; charset=iso-8859-1",
	"html.usascii.html":  "text/html; charset=us-ascii",
	"html.utf8.html":     "text/html; charset=utf-8",
	"html.withbr.html":   "text/html; charset=utf-8",
	"ico.ico":            "image/x-icon",
	"ics.dos.ics":        "text/calendar",
	"ics.ics":            "text/calendar",
	"iso88591.txt":       "text/plain; charset=iso-8859-1",
	"jar.jar":            "application/jar",
	"jp2.jp2":            "image/jp2",
	"jpf.jpf":            "image/jpx",
	"jpg.jpg":            "image/jpeg",
	"jpm.jpm":            "image/jpm",
	"xpm.xpm":            "image/x-xpixmap",
	"js.js":              "application/javascript",
	"json.json":          "application/json",
	"kml.kml":            "application/vnd.google-earth.kml+xml",
	"lit.lit":            "application/x-ms-reader",
	"ln":                 "application/x-executable",
	"lua.lua":            "text/x-lua",
	"lz.lz":              "application/lzip",
	"m3u.m3u":            "application/vnd.apple.mpegurl",
	"m4a.m4a":            "audio/x-m4a",
	"audio.mp4":          "audio/mp4",
	"lnk.lnk":            "application/x-ms-shortcut",
	"macho.macho":        "application/x-mach-binary",
	"mdb.mdb":            "application/x-msaccess",
	"midi.midi":          "audio/midi",
	"mkv.mkv":            "video/x-matroska",
	"mobi.mobi":          "application/x-mobipocket-ebook",
	"mov.mov":            "video/quicktime",
	"mp3.mp3":            "audio/mpeg",
	"mp3.v1.notag.mp3":   "audio/mpeg",
	"mp3.v2.5.notag.mp3": "audio/mpeg",
	"mp3.v2.notag.mp3":   "audio/mpeg",
	"mp4.1.mp4":          "video/mp4",
	"mp4.mp4":            "video/mp4",
	"mpc.mpc":            "audio/musepack",
	"mpeg.mpeg":          "video/mpeg",
	"mqv.mqv":            "video/quicktime",
	"mrc.mrc":            "application/marc",
	"msg.msg":            "application/vnd.ms-outlook",
	"ndjson.ndjson":      "application/x-ndjson",
	"nes.nes":            "application/vnd.nintendo.snes.rom",
	"elfobject":          "application/x-object",
	"odf.odf":            "application/vnd.oasis.opendocument.formula",
	"sxc.sxc":            "application/vnd.sun.xml.calc",
	"odg.odg":            "application/vnd.oasis.opendocument.graphics",
	"odp.odp":            "application/vnd.oasis.opendocument.presentation",
	"ods.ods":            "application/vnd.oasis.opendocument.spreadsheet",
	"odt.odt":            "application/vnd.oasis.opendocument.text",
	"ogg.oga":            "audio/ogg",
	"ogg.ogv":            "video/ogg",
	"ogg.spx.oga":        "audio/ogg",
	"otf.otf":            "font/otf",
	"otg.otg":            "application/vnd.oasis.opendocument.graphics-template",
	"otp.otp":            "application/vnd.oasis.opendocument.presentation-template",
	"ots.ots":            "application/vnd.oasis.opendocument.spreadsheet-template",
	"ott.ott":            "application/vnd.oasis.opendocument.text-template",
	"odc.odc":            "application/vnd.oasis.opendocument.chart",
	"owl2.owl":           "application/owl+xml",
	"pat.pat":            "image/x-gimp-pat",
	"pdf.pdf":            "application/pdf",
	"php.php":            "text/x-php",
	"pl.pl":              "text/x-perl",
	"png.png":            "image/png",
	"ppt.ppt":            "application/vnd.ms-powerpoint",
	"pptx.pptx":          "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"ps.ps":              "application/postscript",
	"psd.psd":            "image/vnd.adobe.photoshop",
	"p7s_pem.p7s":        "application/pkcs7-signature",
	"p7s_der.p7s":        "application/pkcs7-signature",
	"pub.pub":            "application/vnd.ms-publisher",
	"py.py":              "application/x-python",
	"qcp.qcp":            "audio/qcelp",
	"rar.rar":            "application/x-rar-compressed",
	"rmvb.rmvb":          "application/vnd.rn-realmedia-vbr",
	"rpm.rpm":            "application/x-rpm",
	"rss.rss":            "application/rss+xml",
	"rtf.rtf":            "text/rtf",
	"sample32.macho":     "application/x-mach-binary",
	"sample64.macho":     "application/x-mach-binary",
	"shp.shp":            "application/octet-stream",
	"shx.shx":            "application/octet-stream",
	"so.so":              "application/x-sharedlib",
	"sqlite.sqlite":      "application/x-sqlite3",
	"svg.1.svg":          "image/svg+xml",
	"svg.svg":            "image/svg+xml",
	"swf.swf":            "application/x-shockwave-flash",
	"tar.tar":            "application/x-tar",
	"tcl.tcl":            "text/x-tcl",
	"tcx.tcx":            "application/vnd.garmin.tcx+xml",
	"tiff.tiff":          "image/tiff",
	"torrent.torrent":    "application/x-bittorrent",
	"tsv.tsv":            "text/tab-separated-values",
	"ttf.ttf":            "font/ttf",
	"tzfile":             "application/tzif",
	"utf16bebom.txt":     "text/plain; charset=utf-16be",
	"utf16lebom.txt":     "text/plain; charset=utf-16le",
	"utf32bebom.txt":     "text/plain; charset=utf-32be",
	"utf32lebom.txt":     "text/plain; charset=utf-32le",
	"utf8.txt":           "text/plain; charset=utf-8",
	"vcf.dos.vcf":        "text/vcard",
	"vcf.vcf":            "text/vcard",
	"voc.voc":            "audio/x-unknown",
	"warc.warc":          "application/warc",
	"wasm.wasm":          "application/wasm",
	"wav.wav":            "audio/wav",
	"webm.webm":          "video/webm",
	"webp.webp":          "image/webp",
	"woff.woff":          "font/woff",
	"woff2.woff2":        "font/woff2",
	"x3d.x3d":            "model/x3d+xml",
	"xar.xar":            "application/x-xar",
	"xcf.xcf":            "image/x-xcf",
	"xfdf.xfdf":          "application/vnd.adobe.xfdf",
	"xlf.xlf":            "application/x-xliff+xml",
	"xls.xls":            "application/vnd.ms-excel",
	"xlsx.1.xlsx":        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"xlsx.2.xlsx":        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"xlsx.xlsx":          "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"xml.xml":            "text/xml; charset=utf-8",
	"xml.withbr.xml":     "text/xml; charset=utf-8",
	"xz.xz":              "application/x-xz",
	"zip.zip":            "application/zip",
	"zst.zst":            "application/zstd",
}

func TestDetect(t *testing.T) {
	errStr := "File: %s; Expected: %s != Detected: %s; err: %v"
	extStr := "File: %s; ExpectedExt: %s != DetectedExt: %s"
	for fName, expected := range files {
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		if mtype := mimetype.Detect(data); mtype.String() != expected {
			t.Errorf(errStr, fName, expected, mtype.String(), nil)
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		if mtype, err := mimetype.DetectReader(f); mtype.String() != expected {
			t.Errorf(errStr, fName, expected, mtype.String(), err)
		}
		f.Close()

		if mtype, err := mimetype.DetectFile(fileName); mtype.String() != expected {
			t.Errorf(errStr, fName, expected, mtype.String(), err)
		} else if mtype.Extension() != filepath.Ext(fName) {
			t.Errorf(extStr, fName, filepath.Ext(fName), mtype.Extension())
		}
	}
}

func TestEqualsAny(t *testing.T) {
	type ss []string
	testCases := []struct {
		m1  string
		m2  ss
		res bool
	}{
		{"foo/bar", ss{"foo/bar"}, true},
		{"  foo/bar", ss{"foo/bar	"}, true}, // whitespace
		{"  foo/bar", ss{"foo/BAR	"}, true}, // case
		{"  foo/bar", ss{"foo/baz"}, false},
		{";charset=utf-8", ss{""}, true},
		{"", ss{"", "foo/bar"}, true},
		{"foo/bar", ss{""}, false},
		{"foo/bar", nil, false},
	}
	for _, tc := range testCases {
		if mimetype.EqualsAny(tc.m1, tc.m2...) != tc.res {
			t.Errorf("Equality test failed for %+v", tc)
		}
	}
}

func TestDetectReader(t *testing.T) {
	errStr := "File: %s; Mime: %s != DetectedMime: %s; err: %v"
	for fName, expected := range files {
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		r := breakReader{
			r:         f,
			breakSize: 3,
		}
		if mtype, err := mimetype.DetectReader(&r); mtype.String() != expected {
			t.Errorf(errStr, fName, expected, mtype.String(), err)
		}
		f.Close()
	}
}

// breakReader breaks the string every breakSize characters.
// It is like:
//   <html><h
//   ead><tit
//   le>html<
//   ...
type breakReader struct {
	r         io.Reader
	breakSize int
}

func (b *breakReader) Read(p []byte) (int, error) {
	if len(p) > b.breakSize {
		p = p[:b.breakSize]
	}
	n, err := io.ReadFull(b.r, p)
	if err == io.ErrUnexpectedEOF {
		return n, io.EOF
	}
	return n, err
}

func TestFaultyInput(t *testing.T) {
	inexistent := "inexistent.file"
	if _, err := mimetype.DetectFile(inexistent); err == nil {
		t.Errorf("%s should not match successfully", inexistent)
	}

	f, _ := os.Open(inexistent)
	if _, err := mimetype.DetectReader(f); err == nil {
		t.Errorf("%s reader should not match successfully", inexistent)
	}
}

func TestZeroLimit(t *testing.T) {
	mimetype.SetLimit(0)
	mtype, err := mimetype.DetectFile("testdata/utf8.txt")
	if err != nil {
		t.Fatal(err)
	}
	if mtype.String() != "text/plain; charset=utf-8" {
		t.Fatal("utf8.txt should have text/plain MIME")
	}
}

func TestHierarchy(t *testing.T) {
	detectedMIME, err := mimetype.DetectFile("testdata/html.html")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"text/html",
		"text/plain",
		"application/octet-stream",
	}

	i := 0
	for mtype := detectedMIME; mtype != nil; mtype = mtype.Parent() {
		if len(expected)-1 < i {
			t.Fatalf("hierarchy len error; expected: %d, got: %d", len(expected), i)
		}
		if !mtype.Is(expected[i]) {
			t.Fatalf("hierarchy error; expected: %s, got: %s", expected[i], mtype)
		}
		i++
	}
	if len(expected) != i {
		t.Fatalf("hierarchy len error; expected: %d, got: %d", len(expected), i)
	}
}

func TestExtend(t *testing.T) {
	foobarDet := func(raw []byte, limit uint32) bool {
		return bytes.HasPrefix(raw, []byte("foobar"))
	}

	mimetype.Extend(foobarDet, "text/foobar", ".fb")

	mtype := mimetype.Detect([]byte("foobar file content"))
	if !mtype.Is("text/foobar") {
		t.Fatalf("extend error; expected text/foobar, got: %s", mtype)
	}
	if !mtype.Parent().Is("application/octet-stream") {
		t.Fatalf("extend parent error; expected application/octet-stream, got: %s", mtype.Parent())
	}
}

func TestConcurrent(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for i := 0; i < 1000; i++ {
			mimetype.Detect([]byte("text content"))
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			mimetype.SetLimit(5000 + uint32(i))
		}
		wg.Done()
	}()

	wg.Wait()
	// Reset to original limit for benchmarks.
	mimetype.SetLimit(3072)
}

// For #162.
func TestEmptyInput(t *testing.T) {
	mtype, err := mimetype.DetectReader(bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("empty reader err; expected: nil, got: %s", err)
	}
	plain := "text/plain"
	if !mtype.Is(plain) {
		t.Fatalf("empty reader detection; expected: %s, got: %s", plain, mtype)
	}
	mtype = mimetype.Detect(nil)
	if !mtype.Is(plain) {
		t.Fatalf("empty bytes slice detection; expected: %s, got: %s", plain, mtype)
	}
}

// Benchmarking a random slice of bytes is as close as possible to the real
// world usage. A random byte slice is almost guaranteed to fail being detected.
//
// When performing a detection on a file it is very likely there will be
// multiple rules failing before finding the one that matches, ex: a jpg file
// might be tested for zip, gzip, etc., before it is identified.
func BenchmarkSliceRand(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	data := make([]byte, 3072)
	if _, err := io.ReadFull(r, data); err != io.ErrUnexpectedEOF && err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mimetype.Detect(data)
		}
	})
}

func BenchmarkSliceTar(b *testing.B) {
	tar, err := ioutil.ReadFile("testdata/tar.tar")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Detect(tar)
	}
}

func BenchmarkSliceZip(b *testing.B) {
	zip, err := ioutil.ReadFile("testdata/zip.zip")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Detect(zip)
	}
}

func BenchmarkSliceJpeg(b *testing.B) {
	jpeg, err := ioutil.ReadFile("testdata/jpg.jpg")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Detect(jpeg)
	}
}

func BenchmarkSliceGif(b *testing.B) {
	gif, err := ioutil.ReadFile("testdata/gif.gif")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Detect(gif)
	}
}

func BenchmarkSlicePng(b *testing.B) {
	png, err := ioutil.ReadFile("testdata/png.png")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mimetype.Detect(png)
	}
}
