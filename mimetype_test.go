package mimetype

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const testDataDir = "testdata"

type testcase struct {
	file         string
	mime         *MIME
	expectedMIME string
	// If bench is true, then this entry will be used in benchmarks.
	bench bool
}

var testcases = []testcase{
	{"3g2.3g2", threeG2, "video/3gpp2", true},
	{"3gp.3gp", threeGP, "video/3gpp", true},
	{"3mf.3mf", threemf, "application/vnd.ms-package.3dmanufacturing-3dmodel+xml", true},
	{"7z.7z", sevenZ, "application/x-7z-compressed", true},
	{"a.a", ar, "application/x-archive", true},
	{"aac.aac", aac, "audio/aac", true},
	{"aaf.aaf", aaf, "application/octet-stream", true},
	{"accdb.accdb", accdb, "application/x-msaccess", true},
	{"aiff.aiff", aiff, "audio/aiff", true},
	{"amf.amf", amf, "application/x-amf", true},
	{"amr.amr", amr, "audio/amr", true},
	{"ape.ape", ape, "audio/ape", true},
	{"apng.png", apng, "image/vnd.mozilla.apng", true},
	{"asf.asf", asf, "video/x-ms-asf", true},
	{"atom.atom", atom, "application/atom+xml", true},
	{"au.au", au, "audio/basic", true},
	{"avi.avi", avi, "video/x-msvideo", true},
	{"avif.avif", avif, "image/avif", true},
	{"avifsequence.avif", avif, "image/avif", false},
	{"bmp.bmp", bmp, "image/bmp", true},
	{"bpg.bpg", bpg, "image/bpg", true},
	{"bz2.bz2", bz2, "application/x-bzip2", true},
	{"cab.cab", cab, "application/vnd.ms-cab-compressed", true},
	{"cab.is.cab", cabIS, "application/x-installshield", true},
	{"class.class", class, "application/x-java-applet", true},
	{"crx.crx", crx, "application/x-chrome-extension", true},
	{"csv.csv", csv, "text/csv", true},
	{"cpio.cpio", cpio, "application/x-cpio", true},
	{"dae.dae", collada, "model/vnd.collada+xml", true},
	{"dbf.dbf", dbf, "application/x-dbf", true},
	{"dcm.dcm", dcm, "application/dicom", true},
	{"deb.deb", deb, "application/vnd.debian.binary-package", true},
	{"djvu.djvu", djvu, "image/vnd.djvu", true},
	{"doc.doc", doc, "application/msword", true},
	{"docx.docx", docx, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true},
	{"drpm.rpm", rpm, "application/x-rpm", true},
	{"dwg.1.dwg", dwg, "image/vnd.dwg", false},
	{"dwg.dwg", dwg, "image/vnd.dwg", true},
	{"eot.eot", eot, "application/vnd.ms-fontobject", true},
	{"epub.epub", epub, "application/epub+zip", true},
	{"fdf.fdf", fdf, "application/vnd.fdf", true},
	{"fits.fits", fits, "application/fits", true},
	{"flac.flac", flac, "audio/flac", true},
	{"flv.flv", flv, "video/x-flv", true},
	{"gbr.gbr", gbr, "image/x-gimp-gbr", true},
	{"geojson.1.geojson", geoJSON, "application/geo+json", false},
	{"geojson.geojson", geoJSON, "application/geo+json", true},
	{"gif.gif", gif, "image/gif", true},
	{"glb.glb", glb, "model/gltf-binary", true},
	{"gml.gml", gml, "application/gml+xml", true},
	{"gpx.gpx", gpx, "application/gpx+xml", true},
	{"gz.gz", gzip, "application/gzip", true},
	{"har.har", har, "application/json", true},
	{"hdr.hdr", hdr, "image/vnd.radiance", true},
	{"heic.single.heic", heic, "image/heic", true},
	{"heif.heif", heif, "image/heif", true},
	{"html.html", html, "text/html; charset=utf-8", true},
	{"html.iso88591.html", html, "text/html; charset=iso-8859-1", false},
	{"html.svg.html", html, "text/html; charset=utf-8", false},
	{"html.usascii.html", html, "text/html; charset=us-ascii", false},
	{"html.utf8.html", html, "text/html; charset=utf-8", false},
	{"html.withbr.html", html, "text/html; charset=utf-8", false},
	{"ico.ico", ico, "image/x-icon", true},
	{"ics.dos.ics", iCalendar, "text/calendar", true},
	{"ics.ics", iCalendar, "text/calendar", false},
	{"iso88591.txt", text, "text/plain; charset=iso-8859-1", false},
	{"jar.jar", jar, "application/jar", true},
	{"jp2.jp2", jp2, "image/jp2", true},
	{"jpf.jpf", jpx, "image/jpx", true},
	{"jpg.jpg", jpg, "image/jpeg", true},
	{"jpm.jpm", jpm, "image/jpm", true},
	{"jxl.jxl", jxl, "image/jxl", true},
	{"jxr.jxr", jxr, "image/jxr", true},
	{"xpm.xpm", xpm, "image/x-xpixmap", true},
	{"js.js", js, "application/javascript", true},
	{"json.json", json, "application/json", true},
	{"json.lowascii.json", json, "application/json", false},
	// json.{int,float,string}.txt contain a single JSON value. They are valid JSON
	// documents, but they should not be detected as application/json. This mimics
	// the behaviour of the file utility and seems the correct thing to do.
	{"json.int.txt", text, "text/plain; charset=utf-8", false},
	{"json.float.txt", text, "text/plain; charset=utf-8", false},
	{"json.string.txt", text, "text/plain; charset=utf-8", false},
	{"kml.kml", kml, "application/vnd.google-earth.kml+xml", true},
	{"lit.lit", lit, "application/x-ms-reader", true},
	{"lua.lua", lua, "text/x-lua", true},
	{"lz.lz", lzip, "application/lzip", true},
	{"m3u.m3u", m3u, "application/vnd.apple.mpegurl", true},
	{"m4a.m4a", m4a, "audio/x-m4a", true},
	{"audio.mp4", aMp4, "audio/mp4", true},
	{"lnk.lnk", lnk, "application/x-ms-shortcut", true},
	{"macho.macho", macho, "application/x-mach-binary", true},
	{"mdb.mdb", mdb, "application/x-msaccess", true},
	{"midi.midi", midi, "audio/midi", true},
	{"mkv.mkv", mkv, "video/x-matroska", true},
	{"mobi.mobi", mobi, "application/x-mobipocket-ebook", true},
	{"mov.mov", quickTime, "video/quicktime", true},
	{"mp3.mp3", mp3, "audio/mpeg", true},
	{"mp3.v1.notag.mp3", mp3, "audio/mpeg", false},
	{"mp3.v2.5.notag.mp3", mp3, "audio/mpeg", false},
	{"mp3.v2.notag.mp3", mp3, "audio/mpeg", false},
	{"mp4.1.mp4", mp4, "video/mp4", false},
	{"mp4.mp4", mp4, "video/mp4", true},
	{"mpc.mpc", musePack, "audio/musepack", true},
	{"mpeg.mpeg", mpeg, "video/mpeg", true},
	{"mqv.mqv", mqv, "video/quicktime", true},
	{"mrc.mrc", mrc, "application/marc", true},
	{"msi.msi", msi, "application/x-ms-installer", true},
	{"msg.msg", msg, "application/vnd.ms-outlook", true},
	{"ndjson.xl.ndjson", ndJSON, "application/x-ndjson", false},
	{"ndjson.ndjson", ndJSON, "application/x-ndjson", true},
	{"nes.nes", nes, "application/vnd.nintendo.snes.rom", true},
	{"elfobject", elfObj, "application/x-object", true},
	{"odf.odf", odf, "application/vnd.oasis.opendocument.formula", true},
	{"sxc.sxc", sxc, "application/vnd.sun.xml.calc", true},
	{"odg.odg", odg, "application/vnd.oasis.opendocument.graphics", true},
	{"odp.odp", odp, "application/vnd.oasis.opendocument.presentation", true},
	{"ods.ods", ods, "application/vnd.oasis.opendocument.spreadsheet", true},
	{"odt.odt", odt, "application/vnd.oasis.opendocument.text", true},
	{"ogg.oga", ogg, "audio/ogg", true},
	{"ogg.ogv", ogg, "video/ogg", true},
	{"ogg.spx.oga", ogg, "audio/ogg", true},
	{"otf.otf", otf, "font/otf", true},
	{"otg.otg", otg, "application/vnd.oasis.opendocument.graphics-template", true},
	{"otp.otp", otp, "application/vnd.oasis.opendocument.presentation-template", true},
	{"ots.ots", ots, "application/vnd.oasis.opendocument.spreadsheet-template", true},
	{"ott.ott", ott, "application/vnd.oasis.opendocument.text-template", true},
	{"odc.odc", odc, "application/vnd.oasis.opendocument.chart", true},
	{"owl2.owl", owl2, "application/owl+xml", true},
	{"pat.pat", pat, "image/x-gimp-pat", true},
	{"pdf.pdf", pdf, "application/pdf", true},
	{"php.php", php, "text/x-php", true},
	{"pl.pl", perl, "text/x-perl", true},
	{"png.png", png, "image/png", true},
	{"ppt.ppt", ppt, "application/vnd.ms-powerpoint", true},
	{"pptx.pptx", pptx, "application/vnd.openxmlformats-officedocument.presentationml.presentation", true},
	{"ps.ps", ps, "application/postscript", true},
	{"psd.psd", psd, "image/vnd.adobe.photoshop", true},
	{"p7s_pem.p7s", p7s, "application/pkcs7-signature", true},
	{"p7s_der.p7s", p7s, "application/pkcs7-signature", true},
	{"pub.pub", pub, "application/vnd.ms-publisher", true},
	{"py.py", python, "text/x-python", true},
	{"qcp.qcp", qcp, "audio/qcelp", true},
	{"rar.rar", rar, "application/x-rar-compressed", true},
	{"rmvb.rmvb", rmvb, "application/vnd.rn-realmedia-vbr", true},
	{"rpm.rpm", rpm, "application/x-rpm", true},
	{"rss.rss", rss, "application/rss+xml", true},
	{"rtf.rtf", rtf, "text/rtf", true},
	{"sample32.macho", macho, "application/x-mach-binary", false},
	{"sample64.macho", macho, "application/x-mach-binary", false},
	{"shp.shp", shp, "application/vnd.shp", true},
	{"shx.shx", shx, "application/vnd.shx", true},
	{"so.so", elfLib, "application/x-sharedlib", true},
	{"sqlite.sqlite", sqlite3, "application/vnd.sqlite3", true},
	{"srt.srt", srt, "application/x-subrip", true},
	{"svg.1.svg", svg, "image/svg+xml", false},
	{"svg.svg", svg, "image/svg+xml", true},
	{"swf.swf", swf, "application/x-shockwave-flash", true},
	{"tar.tar", tar, "application/x-tar", true},
	{"tar.gnu.tar", tar, "application/x-tar", false},
	{"tar.oldgnu.tar", tar, "application/x-tar", false},
	{"tar.posix.tar", tar, "application/x-tar", false},
	// tar.star.tar was generated with star 1.6.
	{"tar.star.tar", tar, "application/x-tar", false},
	{"tar.ustar.tar", tar, "application/x-tar", false},
	{"tar.v7.tar", tar, "application/x-tar", false},
	// tar.v7-gnu.tar is a v7 tar archive generated with GNU tar 1.29.
	{"tar.v7-gnu.tar", tar, "application/x-tar", false},
	{"tcl.tcl", tcl, "text/x-tcl", true},
	{"tcx.tcx", tcx, "application/vnd.garmin.tcx+xml", true},
	{"tiff.tiff", tiff, "image/tiff", true},
	{"torrent.torrent", torrent, "application/x-bittorrent", true},
	{"tsv.tsv", tsv, "text/tab-separated-values", true},
	{"ttc.ttc", ttc, "font/collection", true},
	{"ttf.ttf", ttf, "font/ttf", true},
	{"tzfile", tzif, "application/tzif", true},
	{"utf16bebom.txt", text, "text/plain; charset=utf-16be", false},
	{"utf16lebom.txt", text, "text/plain; charset=utf-16le", false},
	{"utf32bebom.txt", text, "text/plain; charset=utf-32be", false},
	{"utf32lebom.txt", text, "text/plain; charset=utf-32le", false},
	{"utf8.txt", text, "text/plain; charset=utf-8", true},
	{"utf8ctrlchars", root, "application/octet-stream", false},
	{"vcf.vcf", vCard, "text/vcard", true},
	{"vcf.dos.vcf", vCard, "text/vcard", false},
	{"voc.voc", voc, "audio/x-unknown", true},
	{"vtt.vtt", vtt, "text/vtt", true},
	{"vtt.space.vtt", vtt, "text/vtt", false},
	{"vtt.tab.vtt", vtt, "text/vtt", false},
	{"vtt.eof.vtt", vtt, "text/vtt", false},
	{"warc.warc", warc, "application/warc", true},
	{"wasm.wasm", wasm, "application/wasm", true},
	{"wav.wav", wav, "audio/wav", true},
	{"webm.webm", webM, "video/webm", true},
	{"webp.webp", webp, "image/webp", true},
	{"woff.woff", woff, "font/woff", true},
	{"woff2.woff2", woff2, "font/woff2", true},
	{"x3d.x3d", x3d, "model/x3d+xml", true},
	{"xar.xar", xar, "application/x-xar", true},
	{"xcf.xcf", xcf, "image/x-xcf", true},
	{"xfdf.xfdf", xfdf, "application/vnd.adobe.xfdf", true},
	{"xlf.xlf", xliff, "application/x-xliff+xml", true},
	{"xls.xls", xls, "application/vnd.ms-excel", true},
	{"xlsx.xlsx", xlsx, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true},
	{"xml.xml", xml, "text/xml; charset=utf-8", true},
	{"xml.withbr.xml", xml, "text/xml; charset=utf-8", false},
	{"xz.xz", xz, "application/x-xz", true},
	{"zip.zip", zip, "application/zip", true},
	{"zst.zst", zstd, "application/zstd", true},
}

func TestDetect(t *testing.T) {
	errStr := "File: %s; Expected: %s != Detected: %s; err: %v"
	extStr := "File: %s; ExpectedExt: %s != DetectedExt: %s"
	for _, tc := range testcases {
		fileName := filepath.Join(testDataDir, tc.file)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		if mtype := Detect(data); mtype.String() != tc.expectedMIME {
			t.Errorf(errStr, tc.file, tc.expectedMIME, mtype.String(), nil)
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		if mtype, err := DetectReader(f); mtype.String() != tc.expectedMIME {
			t.Errorf(errStr, tc.file, tc.expectedMIME, mtype.String(), err)
		}
		f.Close()

		if mtype, err := DetectFile(fileName); mtype.String() != tc.expectedMIME {
			t.Errorf(errStr, tc.file, tc.expectedMIME, mtype.String(), err)
		} else if mtype.Extension() != filepath.Ext(tc.file) {
			t.Errorf(extStr, tc.file, filepath.Ext(tc.file), mtype.Extension())
		}
	}
}

// This test generates the doc file containing the table with the supported MIMEs.
func TestGenerateSupportedFormats(t *testing.T) {
	f, err := os.OpenFile("supported_mimes.md", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	nodes := root.flatten()
	header := fmt.Sprintf(`## %d Supported MIME types
This file is automatically generated when running tests. Do not edit manually.

Extension | MIME type | Aliases
--------- | --------- | -------
`, len(nodes))

	if _, err := f.WriteString(header); err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes {
		ext := n.extension
		if ext == "" {
			ext = "n/a"
		}

		aliases := strings.Join(n.aliases, ", ")
		if aliases == "" {
			aliases = "-"
		}
		str := fmt.Sprintf("**%s** | %s | %s\n", ext, n.mime, aliases)
		if _, err := f.WriteString(str); err != nil {
			t.Fatal(err)
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
		if EqualsAny(tc.m1, tc.m2...) != tc.res {
			t.Errorf("Equality test failed for %+v", tc)
		}
	}
}

func TestDetectReader(t *testing.T) {
	errStr := "File: %s; Mime: %s != DetectedMime: %s; err: %v"
	for _, tc := range testcases {
		fileName := filepath.Join(testDataDir, tc.file)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		r := breakReader{
			r:         f,
			breakSize: 3,
		}
		if mtype, err := DetectReader(&r); mtype.String() != tc.expectedMIME {
			t.Errorf(errStr, tc.file, tc.expectedMIME, mtype.String(), err)
		}
		f.Close()
	}
}

// breakReader breaks the string every breakSize characters.
// It is like:
//
//	<html><h
//	ead><tit
//	le>html<
//	...
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
	if mtype, err := DetectFile(inexistent); err == nil {
		t.Errorf("%s should not match successfully", inexistent)
	} else if mtype.String() != "application/octet-stream" {
		t.Errorf("inexistent.file expected application/octet-stream, got %s", mtype)
	}

	f, _ := os.Open(inexistent)
	if mtype, err := DetectReader(f); err == nil {
		t.Errorf("%s reader should not match successfully", inexistent)
	} else if mtype.String() != "application/octet-stream" {
		t.Errorf("inexistent.file reader expected application/octet-stream, got %s", mtype)
	}
}

func TestHierarchy(t *testing.T) {
	detectedMIME, err := DetectFile("testdata/html.html")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"text/html; charset=utf-8",
		"text/plain",
		"application/octet-stream",
	}

	got := []string{}
	for mtype := detectedMIME; mtype != nil; mtype = mtype.Parent() {
		got = append(got, mtype.String())
	}
	if le, lg := len(expected), len(got); le != lg {
		t.Fatalf("hierarchy len error; expected: %d, got: %d", le, lg)
	}

	for i := range expected {
		if expected[i] != got[i] {
			t.Fatalf("hierarchy error; expected: %s, got: %s", expected, got)
		}
	}
}

func TestConcurrent(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(4)

	Extend(func([]byte, uint32) bool { return false }, "e", ".e")
	go func() {
		for i := 0; i < 1000; i++ {
			Detect([]byte("text content"))
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			SetLimit(5000 + uint32(i))
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			Lookup("text/plain")
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			Lookup("e").Extend(func([]byte, uint32) bool { return false }, "e", ".e")
		}
		wg.Done()
	}()

	wg.Wait()
	// Reset to original limit for benchmarks.
	SetLimit(defaultLimit)
}

// For #162.
func TestEmptyInput(t *testing.T) {
	mtype, err := DetectReader(bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("empty reader err; expected: nil, got: %s", err)
	}
	plain := "text/plain"
	if !mtype.Is(plain) {
		t.Fatalf("empty reader detection; expected: %s, got: %s", plain, mtype)
	}
	mtype = Detect(nil)
	if !mtype.Is(plain) {
		t.Fatalf("empty bytes slice detection; expected: %s, got: %s", plain, mtype)
	}
	SetLimit(0)
	mtype, err = DetectReader(bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("0 limіt, empty reader err; expected: nil, got: %s", err)
	}
	if !mtype.Is(plain) {
		t.Fatalf("0 limit, empty reader detection; expected: %s, got: %s", plain, mtype)
	}
	SetLimit(defaultLimit)
}

func BenchmarkAll(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	// randData is used for the negative case benchmark.
	randData := make([]byte, defaultLimit)
	if _, err := io.ReadFull(r, randData); err != io.ErrUnexpectedEOF && err != nil {
		b.Fatal(err)
	}

	for _, tc := range testcases {
		if !tc.bench {
			continue
		}
		// data is used for the positive case benchmark.
		data, err := os.ReadFile(filepath.Join(testDataDir, tc.file))
		if err != nil {
			b.Fatal(err)
		}
		b.Run(tc.file, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				if !tc.mime.detector(data, defaultLimit) {
					b.Fatalf("positive detection should never fail; file=%s", tc.file)
				}
				if tc.mime.detector(randData, defaultLimit) {
					b.Fatalf("negative detection should always fail; file=%s", tc.file)
				}
			}
		})
	}
}

// Check there are no panics for nil inputs.
func TestIndexOutOfRangePanic(t *testing.T) {
	for _, n := range root.flatten() {
		n.detector(nil, 1<<10)
	}
}

// MIME type equality ignores any optional MIME parameters, so, in order to not
// parse each alias when testing for equality, we must ensure they are
// registered with no parameters.
func TestMIMEFormat(t *testing.T) {
	for _, n := range root.flatten() {
		// All extensions must be dot prefixed so they are compatible
		// with the stdlib mime package.
		if n.Extension() != "" && !strings.HasPrefix(n.Extension(), ".") {
			t.Fatalf("extension %s should be dot prefixed", n.Extension())
		}
		// All MIMEs must be correctly formatted.
		_, _, err := mime.ParseMediaType(n.String())
		if err != nil {
			t.Fatalf("error parsing node MIME: %s", err)
		}
		// Aliases must have no optional MIME parameters.
		for _, a := range n.aliases {
			parsed, params, err := mime.ParseMediaType(a)
			if err != nil {
				t.Fatalf("error parsing node alias MIME: %s", err)
			}
			if parsed != a || len(params) > 0 {
				t.Fatalf("node alias MIME should have no optional params; alias: %s, params: %v", a, params)
			}
		}
	}
}

func TestLookup(t *testing.T) {
	data := []struct {
		mime string
		m    *MIME
	}{
		{root.mime, root},
		{zip.mime, zip},
		{zip.aliases[0], zip},
		{xlsx.mime, xlsx},
	}

	for _, tt := range data {
		t.Run(fmt.Sprintf("lookup %s", tt.mime), func(t *testing.T) {
			if m := Lookup(tt.mime); m != tt.m {
				t.Fatalf("failed to lookup: %s", tt.mime)
			}
		})
	}
}

func TestExtend(t *testing.T) {
	data := []struct {
		mime   string
		ext    string
		parent *MIME
	}{
		{"foo", ".foo", nil},
		{"bar", ".bar", root},
		{"baz", ".baz", zip},
	}

	for _, tt := range data {
		t.Run(fmt.Sprintf("extending to %s", tt.mime), func(t *testing.T) {
			extend := Extend
			if tt.parent != nil {
				extend = tt.parent.Extend
			} else {
				tt.parent = root
			}

			extend(func(raw []byte, limit uint32) bool { return false }, tt.mime, tt.ext)
			m := Lookup(tt.mime)
			if m == nil {
				t.Fatalf("mime %s not found", tt.mime)
			}
			if m.parent != tt.parent {
				t.Fatalf("mime %s has wrong parent: want %s, got %s", tt.mime, tt.parent.mime, m.parent.mime)
			}
		})
	}
}

// Because of the random nature of fuzzing I don't think there is a way to test
// the correctness of the Detect results. Still there is value in fuzzing in
// search for panics.
func FuzzMimetype(f *testing.F) {
	// Some of the more interesting file formats. Most formats are detected by
	// checking some magic numbers in headers, but these have more complicated
	// detection algorithms.
	corpus := []string{
		"testdata/mkv.mkv",
		"testdata/webm.webm",
		"testdata/docx.docx",
		"testdata/pptx.pptx",
		"testdata/xlsx.xlsx",
		"testdata/3gp.3gp",
		"testdata/class.class",
	}
	for _, c := range corpus {
		data, err := os.ReadFile(c)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data[:100])
	}
	// First node is root. Remove it because it matches any input.
	detectors := root.flatten()[1:]
	f.Fuzz(func(t *testing.T, data []byte) {
		matched := false
		for _, d := range detectors {
			if d.detector(data, math.MaxUint32) {
				matched = true
			}
		}
		if !matched {
			t.Skip()
		}
	})
}
