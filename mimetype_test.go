package mimetype_test

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
	"class.class":        "application/x-java-applet; charset=binary",
	"crx.crx":            "application/x-chrome-extension",
	"csv.csv":            "text/csv",
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
	"fits.fits":          "application/fits",
	"flac.flac":          "audio/flac",
	"flv.flv":            "video/x-flv",
	"geojson.1.geojson":  "application/geo+json",
	"geojson.geojson":    "application/geo+json",
	"gml.gml":            "application/gml+xml",
	"gpx.gpx":            "application/gpx+xml",
	"gz.gz":              "application/gzip",
	"heic.single.heic":   "image/heic",
	"html.html":          "text/html; charset=utf-8",
	"html.withbr.html":   "text/html; charset=utf-8",
	"ico.ico":            "image/x-icon",
	"ics.dos.ics":        "text/calendar",
	"ics.ics":            "text/calendar",
	"jar.jar":            "application/jar",
	"jp2.jp2":            "image/jp2",
	"jpf.jpf":            "image/jpx",
	"jpg.jpg":            "image/jpeg",
	"jpm.jpm":            "image/jpm",
	"js.js":              "application/javascript",
	"json.json":          "application/json",
	"kml.kml":            "application/vnd.google-earth.kml+xml",
	"lit.lit":            "application/x-ms-reader",
	"ln":                 "application/x-executable",
	"lua.lua":            "text/x-lua",
	"lz.lz":              "application/lzip",
	"m4a.m4a":            "audio/x-m4a",
	"audio.mp4":          "audio/mp4",
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
	"ndjson.ndjson":      "application/x-ndjson",
	"nes.nes":            "application/vnd.nintendo.snes.rom",
	"elfobject":          "application/x-object",
	"odf.odf":            "application/vnd.oasis.opendocument.formula",
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
	"pdf.pdf":            "application/pdf",
	"php.php":            "text/x-php; charset=utf-8",
	"pl.pl":              "text/x-perl",
	"png.png":            "image/png",
	"ppt.ppt":            "application/vnd.ms-powerpoint",
	"pptx.pptx":          "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"ps.ps":              "application/postscript",
	"psd.psd":            "image/vnd.adobe.photoshop",
	"pub.pub":            "application/vnd.ms-publisher",
	"py.py":              "application/x-python",
	"qcp.qcp":            "audio/qcelp",
	"rar.rar":            "application/x-rar-compressed",
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
	"tsv.tsv":            "text/tab-separated-values",
	"ttf.ttf":            "font/ttf",
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
	"xlf.xlf":            "application/x-xliff+xml",
	"xls.xls":            "application/vnd.ms-excel",
	"xlsx.1.xlsx":        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"xlsx.xlsx":          "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"xml.withbr.xml":     "text/xml; charset=utf-8",
	"xz.xz":              "application/x-xz",
	"zip.zip":            "application/zip",
	"zst.zst":            "application/zstd",
}

func TestDetect(t *testing.T) {
	errStr := "File: %s; ExpectedMIME: %s != DetectedMIME: %s; err: %v"
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

		if mime := mimetype.Detect(data); !mime.Is(expected) {
			t.Errorf(errStr, fName, expected, mime.String(), nil)
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		if mime, err := mimetype.DetectReader(f); !mime.Is(expected) {
			t.Errorf(errStr, fName, expected, mime.String(), err)
		}
		f.Close()

		if mime, err := mimetype.DetectFile(fileName); !mime.Is(expected) {
			t.Errorf(errStr, fName, expected, mime.String(), err)
		} else if mime.Extension() != filepath.Ext(fName) {
			t.Errorf(extStr, fName, filepath.Ext(fName), mime.Extension())
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
		if mime, err := mimetype.DetectReader(&r); !mime.Is(expected) {
			t.Errorf(errStr, fName, expected, mime.String(), err)
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
