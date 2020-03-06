package mimetype

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testDataDir = "testdata"

var files = map[string]*MIME{
	// archives
	"pdf.pdf":     pdf,
	"zip.zip":     zip,
	"tar.tar":     tar,
	"xls.xls":     xls,
	"xlsx.xlsx":   xlsx,
	"xlsx.1.xlsx": xlsx,
	"doc.doc":     doc,
	"doc.1.doc":   doc,
	"docx.docx":   docx,
	"docx.1.docx": docx,
	"ppt.ppt":     ppt,
	"pptx.pptx":   pptx,
	"pub.pub":     pub,
	"odt.odt":     odt,
	"ott.ott":     ott,
	"ods.ods":     ods,
	"ots.ots":     ots,
	"odp.odp":     odp,
	"otp.otp":     otp,
	"odg.odg":     odg,
	"otg.otg":     otg,
	"odf.odf":     odf,
	"epub.epub":   epub,
	"7z.7z":       sevenZ,
	"jar.jar":     jar,
	"gz.gz":       gzip,
	"fits.fits":   fits,
	"xar.xar":     xar,
	"bz2.bz2":     bz2,
	"a.a":         ar,
	"deb.deb":     deb,
	"rpm.rpm":     rpm,
	"drpm.rpm":    rpm,
	"rar.rar":     rar,
	"djvu.djvu":   djvu,
	"mobi.mobi":   mobi,
	"lit.lit":     lit,
	"warc.warc":   warc,
	"zst.zst":     zstd,
	"cab.cab":     cab,
	"xz.xz":       xz,

	// images
	"png.png":          png,
	"jpg.jpg":          jpg,
	"jp2.jp2":          jp2,
	"jpf.jpf":          jpx,
	"jpm.jpm":          jpm,
	"psd.psd":          psd,
	"webp.webp":        webp,
	"tif.tif":          tiff,
	"ico.ico":          ico,
	"bmp.bmp":          bmp,
	"bpg.bpg":          bpg,
	"heic.single.heic": heic,

	// video
	"mp4.mp4":   mp4,
	"mp4.1.mp4": mp4,
	"webm.webm": webM,
	"3gp.3gp":   threeGP,
	"3g2.3g2":   threeG2,
	"flv.flv":   flv,
	"avi.avi":   avi,
	"mov.mov":   quickTime,
	"mqv.mqv":   mqv,
	"mpeg.mpeg": mpeg,
	"mkv.mkv":   mkv,
	"asf.asf":   asf,

	// audio
	"mp3.mp3":            mp3,
	"mp3.v1.notag.mp3":   mp3,
	"mp3.v2.notag.mp3":   mp3,
	"mp3.v2.5.notag.mp3": mp3,
	"wav.wav":            wav,
	"flac.flac":          flac,
	"midi.midi":          midi,
	"ape.ape":            ape,
	"aiff.aiff":          aiff,
	"au.au":              au,
	"ogg.oga":            oggAudio,
	"ogg.spx.oga":        oggAudio,
	"ogg.ogv":            oggVideo,
	"amr.amr":            amr,
	"mpc.mpc":            musePack,
	"aac.aac":            aac,
	"voc.voc":            voc,
	"m4a.m4a":            m4a,
	"m4b.m4b":            aMp4,
	"qcp.qcp":            qcp,

	// source code
	"html.html":         html,
	"html.withbr.html":  html,
	"svg.svg":           svg,
	"svg.1.svg":         svg,
	"utf8.txt":          utf8,
	"utf16lebom.txt":    utf16le,
	"utf16bebom.txt":    utf16be,
	"utf32bebom.txt":    utf32be,
	"utf32lebom.txt":    utf32le,
	"php.php":           php,
	"ps.ps":             ps,
	"json.json":         json,
	"geojson.geojson":   geoJson,
	"geojson.1.geojson": geoJson,
	"ndjson.ndjson":     ndJson,
	"csv.csv":           csv,
	"tsv.tsv":           tsv,
	"rtf.rtf":           rtf,
	"js.js":             js,
	"lua.lua":           lua,
	"pl.pl":             perl,
	"py.py":             python,
	"tcl.tcl":           tcl,
	"vCard.vCard":       vCard,
	"vCard.dos.vCard":   vCard,
	"ics.ics":           iCalendar,
	"ics.dos.ics":       iCalendar,

	// binary
	"class.class": class,
	"swf.swf":     swf,
	"crx.crx":     crx,
	"wasm.wasm":   wasm,
	"exe.exe":     exe,
	"ln":          elfExe,
	"so.so":       elfLib,
	"o.o":         elfObj,
	"dcm.dcm":     dcm,
	"mach.o":      macho,
	"sample32":    macho,
	"sample64":    macho,
	"mrc.mrc":     mrc,

	// fonts
	"ttf.ttf":     ttf,
	"woff.woff":   woff,
	"woff2.woff2": woff2,
	"otf.otf":     otf,
	"eot.eot":     eot,

	// XML and subtypes of XML
	"xml.withbr.xml": xml,
	"kml.kml":        kml,
	"xlf.xlf":        xliff,
	"dae.dae":        collada,
	"gml.gml":        gml,
	"gpx.gpx":        gpx,
	"tcx.tcx":        tcx,
	"x3d.x3d":        x3d,
	"amf.amf":        amf,
	"3mf.3mf":        threemf,
	"rss.rss":        rss,
	"atom.atom":      atom,

	"shp.shp": shp,
	"shx.shx": shx,
	"dbf.dbf": dbf,

	"sqlite3.sqlite3": sqlite3,
	"dwg.dwg":         dwg,
	"dwg.1.dwg":       dwg,
	"nes.nes":         nes,
	"mdb.mdb":         mdb,
	"accdb.accdb":     accdb,
}

func TestDetect(t *testing.T) {
	errStr := "File: %s; Mime: %s != DetectedMime: %s; err: %v"
	for fName, node := range files {
		fileName := filepath.Join(testDataDir, fName)
		f, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		if mime := Detect(data); mime.String() != node.mime {
			t.Errorf(errStr, fName, node.mime, mime.String(), nil)
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			t.Errorf(errStr, fName, node.mime, root.mime, err)
		}

		if mime, err := DetectReader(f); mime.String() != node.mime {
			t.Errorf(errStr, fName, node.mime, mime.String(), err)
		}
		f.Close()

		if mime, err := DetectFile(fileName); mime.String() != node.mime {
			t.Errorf(errStr, fName, node.mime, mime.String(), err)
		}
	}
}

func TestFaultyInput(t *testing.T) {
	inexistent := "inexistent.file"
	if _, err := DetectFile(inexistent); err == nil {
		t.Errorf("%s should not match successfully", inexistent)
	}

	f, _ := os.Open(inexistent)
	if _, err := DetectReader(f); err == nil {
		t.Errorf("%s reader should not match successfully", inexistent)
	}
}

func TestBadBdfInput(t *testing.T) {
	if mime, _ := DetectFile("testdata/bad.dbf"); mime.String() != "application/octet-stream" {
		t.Errorf("failed to detect bad DBF file")
	}
}

func TestGenerateSupportedMimesFile(t *testing.T) {
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

func TestIndexOutOfRange(t *testing.T) {
	for _, n := range root.flatten() {
		_ = n.matchFunc(nil)
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
			t.Fatalf("")
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
