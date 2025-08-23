package mimetype

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"mime"
	"os"
	"strings"
	"sync"
	"testing"
)

// testcases are used for correctness and benchmarks.
// testcase data is provided as a string for convenience,
// but that makes benchmarks allocate more. It can be observed
// for testcases that call fromDisk. Those allocate more because
// they read a lot from disk.
type testcase struct {
	name         string
	data         string
	expectedMIME string
	bench        string
}

var testcases = []testcase{
	{"3gpp2", "\x00\x00\x00\x18ftyp3g24", "video/3gpp2", one},
	{"3gpp2 without ftyp", "\x00\x00\x00\x18mtyp3g24", "application/octet-stream", none},
	{"3gp", "\x00\x00\x00\x18ftyp3gp1", "video/3gpp", one},
	{
		"3mf",
		`<?xml version="1.0"?><model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02">`,
		"application/vnd.ms-package.3dmanufacturing-3dmodel+xml",
		one,
	},
	{"7z", "\x37\x7A\xBC\xAF\x27\x1C", "application/x-7z-compressed", all},
	{"a", "\x21\x3C\x61\x72\x63\x68\x3E", "application/x-archive", one},
	{"aac 1", "\xFF\xF1", "audio/aac", one},
	{"aac 2", "\xFF\xF9", "audio/aac", none},
	{"accdb", offset(4, "Standard ACE DB"), "application/x-msaccess", none}, // none because accdb and mdb share the same MIME
	{"aiff", "\x46\x4F\x52\x4D\x00\x00\x00\x00\x41\x49\x46\x46\x00", "audio/aiff", one},
	{"amf", `<?xml version="1.0"?><amf>`, "application/x-amf", one},
	{"amr", "\x23\x21\x41\x4D\x52", "audio/amr", one},
	{"ape", "\x4D\x41\x43\x20\x96\x0F\x00\x00\x34\x00\x00\x00\x18\x00\x00\x00\x90\xE3", "audio/ape", one},
	{"apng", "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A" + offset(29, "acTL"), "image/vnd.mozilla.apng", all},
	{"asf", "\x30\x26\xB2\x75\x8E\x66\xCF\x11\xA6\xD9\x00\xAA\x00\x62\xCE\x6C", "video/x-ms-asf", one},
	{"atom", `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom">`, "application/atom+xml", one},
	{"au", "\x2E\x73\x6E\x64", "audio/basic", one},
	{"avi", "RIFF\x00\x00\x00\x00AVI LIST\x00", "video/x-msvideo", all},
	{"avif", "\x00\x00\x00\x18ftypavif", "image/avif", all},
	{"avis", "\x00\x00\x00\x18ftypavis", "image/avif", all},
	{"bmp", "\x42\x4D", "image/bmp", all},
	{"bpg", "\x42\x50\x47\xFB", "image/bpg", one},
	{"bz2", "\x42\x5A\x68", "application/x-bzip2", one},
	{"cab", "MSCF\x00\x00\x00\x00", "application/vnd.ms-cab-compressed", one},
	{"cab.is", "ISc(\x00\x00\x00\x01", "application/x-installshield", one},
	{"chm", "ITSF\003\000\000\000\x60\000\000\000", "application/vnd.ms-htmlhelp", one},
	{"class", "\xCA\xFE\xBA\xBE\x00\x00\x00\xFF", "application/x-java-applet", one},
	{
		"crx",
		"Cr24\x00\x00\x00\x00\x01\x00\x00\x00\x0F\x00\x00\x00" + offset(16, "") + "\x50\x4B\x03\x04",
		"application/x-chrome-extension",
		one,
	},
	{
		"csv",
		`1,2
"abc","def"
a,"b`,
		"text/csv",
		all,
	},
	{
		`csv with \r\n`,
		"1,2\r\n3,4\r\na,b",
		"text/csv",
		none,
	},
	{"cpio 7", "070707", "application/x-cpio", one},
	{"cpio 1", "070701", "application/x-cpio", none},
	{"cpio 2", "070702", "application/x-cpio", none},
	{"dae", `<?xml version="1.0"?><COLLADA xmlns="http://www.collada.org/2005/11/COLLADASchema">`, "model/vnd.collada+xml", one},
	{"dbf", "\x03\x5f\x07\x1a\x96\x0f\x00\x00\xc1\x00\xa3\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x6f\x73\x6d\x5f\x69\x64\x00\x00\x00\x00\x00\x43\x00\x00\x00\x00\x0a\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x63\x6f\x64\x65", "application/x-dbf", one},
	{"dcm", offset(128, "\x44\x49\x43\x4D"), "application/dicom", one},
	{"deb", "\x21\x3c\x61\x72\x63\x68\x3e\x0a\x64\x65\x62\x69\x61\x6e\x2d\x62\x69\x6e\x61\x72\x79", "application/vnd.debian.binary-package", all},
	{"djvu", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00DJVU", "image/vnd.djvu", one},
	{"djvuM", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00DJVM", "image/vnd.djvu", none},
	{"djvuI", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00DJVI", "image/vnd.djvu", none},
	{"djvuTHUM", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00THUM", "image/vnd.djvu", none},
	{"doc", fromDisk("doc.doc"), "application/msword", all},
	{"docx", fromDisk("docx.docx"), "application/vnd.openxmlformats-officedocument.wordprocessingml.document", all},
	{"rpm 1", "\xed\xab\xee\xdb", "application/x-rpm", one},
	{"rpm 2", "drpm", "application/x-rpm", none},
	{"dwg", "\x41\x43\x31\x30\x32\x34", "image/vnd.dwg", none},
	{"eot", "\xbe\x45\x00\x00\xfa\x44\x00\x00\x02\x00\x02\x00\x04\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x90\x01\x00\x00\x00\x00\x4c\x50", "application/vnd.ms-fontobject", one},
	{"epub", "\x50\x4B\x03\x04" + offset(26, "mimetypeapplication/epub+zip"), "application/epub+zip", all},
	{"exe", "\x4D\x5A", "application/vnd.microsoft.portable-executable", all},
	{"fdf", "%FDF", "application/vnd.fdf", one},
	{"fits", "\x53\x49\x4d\x50\x4c\x45\x20\x20\x3d\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x54", "application/fits", one},
	{"flac", "\x66\x4C\x61\x43\x00\x00\x00\x22", "audio/flac", one},
	{"flv", "\x46\x4C\x56\x01", "video/x-flv", one},
	{"gbr", offset(20, "GIMP"), "image/x-gimp-gbr", one},
	{"geojson", `{"type":"Feature"}`, "application/geo+json", one},
	{"geojson with space", `{ "type" : "Feature" }`, "application/geo+json", none},
	{"gltf1", `{"asset":{"version":"1.0"}}`, "model/gltf+json", none},
	{"gltf2", `{"asset":{"version":"2.0"}}`, "model/gltf+json", none},
	{"gif 87", "GIF87a", "image/gif", all},
	{"gif 89", "GIF89a", "image/gif", none},
	{"glb 1", "\x67\x6C\x54\x46\x02\x00\x00\x00", "model/gltf-binary", one},
	{"glb 2", "\x67\x6C\x54\x46\x01\x00\x00\x00", "model/gltf-binary", none},
	{"gml", `<?xml version="1.0"?><any xmlns:gml="http://www.opengis.net/gml">`, "application/gml+xml", one},
	{"gml3.2", `<?xml version="1.0"?><any xmlns:gml="http://www.opengis.net/gml/3.2">`, "application/gml+xml", none},
	{"gml3.3", `<?xml version="1.0"?><any xmlns:gml="http://www.opengis.net/gml/3.3/exr">`, "application/gml+xml", none},
	{"gpx", `<?xml version="1.0"?><gpx xmlns="http://www.topografix.com/GPX/1/1">`, "application/gpx+xml", one},
	{"gz", "\x1F\x8B", "application/gzip", all},
	{"har", `{"log":{ "version": "1.2"}}`, "application/json", one},
	{"hdr", "#?RADIANCE\n", "image/vnd.radiance", one},
	{"heic", "\x00\x00\x00\x18ftypheic", "image/heic", one},
	{"heix", "\x00\x00\x00\x18ftypheix", "image/heic", none},
	{"heif mif1", "\x00\x00\x00\x18ftypmif1", "image/heif", one},
	{"heif heim", "\x00\x00\x00\x18ftypheim", "image/heif", none},
	{"heif heis", "\x00\x00\x00\x18ftypheis", "image/heif", none},
	{"heif avic", "\x00\x00\x00\x18ftypavic", "image/heif", none},
	{"html", `<HtMl><bOdY>blah blah blah</body></html>`, "text/html; charset=utf-8", all},
	{"html empty", `<HTML></HTML>`, "text/html; charset=utf-8", none},
	{"html just header", `   <!DOCTYPE HTML>...`, "text/html; charset=utf-8", none},
	{"line ending before html", "\r\n<html>...", "text/html; charset=utf-8", none},
	{
		"html with encoding",
		`<html><head><meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1">`,
		"text/html; charset=iso-8859-1",
		none,
	},
	{
		"html with comment prefix",
		`<!-- this comment should not affect --><html><head>`,
		"text/html; charset=utf-8",
		none,
	},
	{"ico 01", "\x00\x00\x01\x00", "image/x-icon", one},
	{"ico 02", "\x00\x00\x02\x00", "image/x-icon", none},
	{"ics", "BEGIN:VCALENDAR\n00", "text/calendar", one},
	{"ics dos", "BEGIN:VCALENDAR\r\n00", "text/calendar", none},
	{"txt iso88591", "\x0a\xe6\xf8\xe6\xf8\xe5\xe6\xf8\xe5\xe5\x0a", "text/plain; charset=iso-8859-1", none},
	{"jar", fromDisk("jar.jar"), "application/java-archive", all},
	{"jar executable", "PK\x03\x04" + offset(0x1A, "\xFE\xCA"), "application/java-archive", none},
	{"jar in zip #639", fromDisk("jar_in_zip.zip"), "application/zip", none},
	{"jp2", "\x00\x00\x00\x0c\x6a\x50\x20\x20\x0d\x0a\x87\x0a\x00\x00\x00\x14\x66\x74\x79\x70\x6a\x70\x32\x20", "image/jp2", one},
	{"jpf", "\x00\x00\x00\x0c\x6a\x50\x20\x20\x0d\x0a\x87\x0a\x00\x00\x00\x1c\x66\x74\x79\x70\x6a\x70\x78\x20", "image/jpx", one},
	{"jpg", "\xFF\xD8\xFF", "image/jpeg", one},
	{"jpm", "\x00\x00\x00\x0c\x6a\x50\x20\x20\x0d\x0a\x87\x0a\x00\x00\x00\x14\x66\x74\x79\x70\x6a\x70\x6d\x20", "image/jpm", one},
	{"jxl 1", "\xFF\x0A", "image/jxl", one},
	{"jxl 2", "\x00\x00\x00\x0cJXL\x20\x0d\x0a\x87\x0a", "image/jxl", none},
	{"jxr", "\x49\x49\xBC\x01", "image/jxr", one},
	{"xpm", "\x2F\x2A\x20\x58\x50\x4D\x20\x2A\x2F", "image/x-xpixmap", one},
	{"js", "#!/bin/node ", "text/javascript", one},
	{"json", `{"a":"b", "c":[{"a":"b"},1,true,false,"abc"]}`, "application/json", all},
	{"json issue#239", "{\x0A\x09\x09\"key\":\"val\"}\x0A", "application/json", none},
	// json.{int,string}.txt contain a single JSON value. They are valid JSON
	// documents but they should not be detected as application/json. This mimics
	// the behaviour of the file utility and seems the correct thing to do.
	{"json.int.txt", "1", "text/plain; charset=utf-8", none},
	{"json.float.txt", "1.5", "text/plain; charset=utf-8", none},
	{"json.string.txt", `"some string"`, "text/plain; charset=utf-8", none},
	{"kml 2.2", `<?xml version="1.0"?><kml xmlns="http://www.opengis.net/kml/2.2">`, "application/vnd.google-earth.kml+xml", one},
	{"kml 2.0", `<?xml version="1.0"?><kml xmlns="http://earth.google.com/kml/2.0">`, "application/vnd.google-earth.kml+xml", none},
	{"kml 2.1", `<?xml version="1.0"?><kml xmlns="http://earth.google.com/kml/2.1">`, "application/vnd.google-earth.kml+xml", none},
	{"kml 2.2", `<?xml version="1.0"?><kml xmlns="http://earth.google.com/kml/2.2">`, "application/vnd.google-earth.kml+xml", none},
	{"kmz", "\x50\x4b\x03\x04\x14\x00\x00\x00\x08\x00\xe6\x6c\x04\x5b\xfd\xf4\xf2\x45\x41\x00\x00\x00\x43\x00\x00\x00\x07\x00\x1c\x00doc.kml", "application/vnd.google-earth.kmz", none},
	{"lit", "ITOLITLS", "application/x-ms-reader", one},
	{"lotus1", "\x00\x00\x02\x00456\x00" + offset(13, ""), "application/vnd.lotus-1-2-3", one},
	{"lotus2", "\x00\x00\x1a\x00" + offset(16, "\x01"), "application/vnd.lotus-1-2-3", one},
	{"lua", "#!/usr/bin/lua", "text/x-lua", one},
	{"lua space", "#! /usr/bin/lua", "text/x-lua", none},
	{"lz", "\x4c\x5a\x49\x50", "application/lzip", one},
	{"m3u", "#EXTM3U", "application/vnd.apple.mpegurl", one},
	{"m4a", "\x00\x00\x00\x18ftypM4A ", "audio/x-m4a", one},
	{"audio mp4 F4A", "\x00\x00\x00\x18ftypF4A ", "audio/mp4", one},
	{"audio mp4 F4B", "\x00\x00\x00\x18ftypF4B ", "audio/mp4", none},
	{"audio mp4 M4B", "\x00\x00\x00\x18ftypM4B ", "audio/mp4", none},
	{"audio mp4 M4P", "\x00\x00\x00\x18ftypM4P ", "audio/mp4", none},
	{"audio mp4 MSNV", "\x00\x00\x00\x18ftypMSNV", "audio/mp4", none},
	{"audio mp4 NDAS", "\x00\x00\x00\x18ftypNDAS", "audio/mp4", none},
	{"lnk", "\x4C\x00\x00\x00\x01\x14\x02\x00", "application/x-ms-shortcut", one},
	{"mdb", offset(4, "Standard Jet DB"), "application/x-msaccess", one},
	{"midi", "\x4D\x54\x68\x64", "audio/midi", one},
	{"mkv", "\x1a\x45\xdf\xa3\x01\x00\x00\x00\x00\x00\x00\x23\x42\x86\x81\x01\x42\xf7\x81\x01\x42\xf2\x81\x04\x42\xf3\x81\x08\x42\x82\x88\x6d\x61\x74\x72\x6f\x73\x6b\x61", "video/x-matroska", one},
	{"mobi", offset(60, "BOOKMOBI"), "application/x-mobipocket-ebook", one},
	{"mov", "\x00\x00\x00\x14\x66\x74\x79\x70\x71\x74\x20\x20", "video/quicktime", one},
	{"mp3", "\x49\x44\x33\x03", "audio/mpeg", all},
	{"mp3 v1 notag", "\xff\xfb\xc8\x00", "audio/mpeg", none},
	{"mp3 v2.5 notag", "\xff\xe3\x18\xc4", "audio/mpeg", none},
	{"mp3 v2 notag", "\xff\xf3\x82\xc4", "audio/mpeg", none},
	{"mp4 1", "\x00\x00\x00\x18ftyp0000", "video/mp4", all},
	{"mpc", "MPCK", "audio/musepack", one},
	{"mpeg", "\x00\x00\x01\xba", "video/mpeg", one},
	{"mqv", "\x00\x00\x00\x18ftypmqt ", "video/quicktime", none},
	{"mrc", "00057     2200037   4500245001900000\x1e", "application/marc", one},
	{"msi", fromDisk("msi.msi"), "application/x-ms-installer", all},
	{"msg", fromDisk("msg.msg"), "application/vnd.ms-outlook", one},
	{"ndjson", `{"key":"val"}` + "\n" + `{"key":"val"}`, "application/x-ndjson", one},
	{"ndjson spaces", `{ "key" : "val" }` + "\n" + ` { "key" : "val" }`, "application/x-ndjson", one},
	{"nes", "NES\x1a", "application/vnd.nintendo.snes.rom", one},
	{"elfobject", "\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00", "application/x-object", one},
	{"odf", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xb1Z\xa8N\x07\x8a\xa8[*\x00\x00\x00*\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.formula", "application/vnd.oasis.opendocument.formula", one},
	{"sxc", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xbb\x03\x5eGE\xbc\x13\x94\x1c\x00\x00\x00\x1c\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.sun.xml.calc", "application/vnd.sun.xml.calc", one},
	{"odg", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xcbY\xa8N\x9f\x03.\xc4\x2b\x00\x00\x00\x2b\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.graphics", "application/vnd.oasis.opendocument.graphics", one},
	{"odp", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xbdX\xa8N3&\xac\xa8/\x00\x00\x00/\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.presentation", "application/vnd.oasis.opendocument.presentation", one},
	{"ods", "PK\x03\x04\x14\x00\x00\x08\x00\x00\x14V\xa8N\x85l9\x8a.\x00\x00\x00.\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.spreadsheet", "application/vnd.oasis.opendocument.spreadsheet", one},
	{"odt", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xbbP\xa8N\x5e\xc62\n'\x00\x00\x00'\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.text", "application/vnd.oasis.opendocument.text", one},
	{"ogg", "OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\xce\xc6AI\x00\x00\x00\x00py\xf3\x3d\x01\x1e\x01vorbis\x00\x00", "audio/ogg", one},
	{"ogg", "OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x80\xbc\x81_\x00\x00\x00\x00\xd0\xfbP\x84\x01@fishead\x00\x03", "video/ogg", one},
	{"ogg spx oga", "OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\xc7w\xaa\x15\x00\x00\x00\x00V&\x88\x89\x01PSpeex   1", "audio/ogg", one},
	{"one", "\xe4\x52\x5c\x7b\x8c\xd8\xa7\x4d\xae\xb1\x53\x78\xd0\x29\x96\xd3", "application/onenote", one},
	{"otf", "OTTO\x00", "font/otf", one},
	{"otg", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xd1Y\xa8N\xdf%\xad\xe94\x00\x00\x004\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.graphics-template", "application/vnd.oasis.opendocument.graphics-template", one},
	{"otp", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xc4X\xa8N\xef\n\x14:8\x00\x00\x008\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.presentation-template", "application/vnd.oasis.opendocument.presentation-template", one},
	{"ots", "PK\x03\x04\x14\x00\x00\x08\x00\x00\x1bV\xa8N{\x96\xa3N7\x00\x00\x007\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.spreadsheet-template", "application/vnd.oasis.opendocument.spreadsheet-template", one},
	{"ott", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xcfP\xa8N\xe4\x11\x92)0\x00\x00\x000\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.text-template", "application/vnd.oasis.opendocument.text-template", one},
	{"odc", "PK\x03\x04\x14\x00\x00\x08\x00\x00zp2R\xab\xb8\xb2l(\x00\x00\x00(\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.chart", "application/vnd.oasis.opendocument.chart", one},
	{"owl", `<?xml version="1.0"?><Ontology xmlns="http://www.w3.org/2002/07/owl#">`, "application/owl+xml", one},
	{"pat", "\x00\x00\x00\x1c\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x03GPAT", "image/x-gimp-pat", one},
	{"pdf", "%PDF-", "application/pdf", all},
	{"php", "#!/usr/bin/env php", "text/x-php", one},
	{"pl", "#!/usr/bin/perl", "text/x-perl", one},
	{"png", "\x89PNG\x0d\x0a\x1a\x0a", "image/png", all},
	{"ppt", fromDisk("ppt.ppt"), "application/vnd.ms-powerpoint", all},
	{"pptx", fromDisk("pptx.pptx"), "application/vnd.openxmlformats-officedocument.presentationml.presentation", all},
	{"pbm", "P1\n# comment\n\n6 10", "image/x-portable-bitmap", one},
	{"pgm", "P2\n# comment\n\n6 10", "image/x-portable-graymap", one},
	{"ppm", "P3\n# comment\n\n6 10", "image/x-portable-pixmap", one},
	{
		"pam",
		`P7
WIDTH 4
HEIGHT 2
DEPTH 4
MAXVAL 255
TUPLTYPE RGB_ALPHA
ENDHDR`,
		"image/x-portable-arbitrarymap",
		one,
	},
	{"ps", "%!PS-Adobe-", "application/postscript", one},
	{"psd", "8BPS", "image/vnd.adobe.photoshop", all},
	{"p7s_pem", "-----BEGIN PKCS7", "application/pkcs7-signature", one},
	{"p7s_der", "\x30\x82\x01\x26\x06\x09\x2a\x86\x48\x86\xf7\x0d\x01\x07\x02\xa0\x82\x01\x17\x30", "application/pkcs7-signature", one},
	{"pub", fromDisk("pub.pub"), "application/vnd.ms-publisher", one},
	{"py", "#!/usr/bin/python", "text/x-python", one},
	{"py3", "#!/usr/bin/env python3", "text/x-python", one},
	{"qcp", "RIFF\xc0\xcf\x00\x00QLCMf", "audio/qcelp", one},
	{"rar", "Rar!\x1a\x07\x01\x00", "application/x-rar-compressed", all},
	{"rb", "#!/usr/local/bin/ruby", "text/x-ruby", one},
	{"rmvb", ".RMF", "application/vnd.rn-realmedia-vbr", one},
	{"rpm", "\xed\xab\xee\xdb", "application/x-rpm", one},
	{"rss", "\x3c\x3f\x78\x6d\x6c\x20\x76\x65\x72\x73\x69\x6f\x6e\x3d\x22\x31\x2e\x30\x22\x20\x65\x6e\x63\x6f\x64\x69\x6e\x67\x3d\x22\x55\x54\x46\x2d\x38\x22\x3f\x3e\x0a\x3c\x72\x73\x73", "application/rss+xml", one},
	{"rtf", "{\\rtf", "text/rtf", one},
	{"sh", "#!/bin/sh", "text/x-shellscript", one},
	{"shp", fromDisk("shp.shp"), "application/vnd.shp", one},
	{"shx", "\x00\x00\x27\x0a", "application/vnd.shx", one},
	{"so", "\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x03\x00", "application/x-sharedlib", all},
	{"sqlite", "SQLite format 3\x00", "application/vnd.sqlite3", one},
	{"srt", "1\n00:02:16,612 --\x3e 00:02:19,376\nS", "application/x-subrip", one},
	{"svg no xml header", `<svg xmlns="http://www.w3.org/2000/svg"`, "image/svg+xml", all},
	{
		"svg xml header",
		`
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="391" height="391" viewBox="-70.5 -70.5 391 391" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <rect fill="#fff" stroke="#000" x="-70" y="-70" width="390" height="390"/>
</svg>
`,
		"image/svg+xml",
		all,
	},
	{
		"svg with comment prefix",
		`<!-- this comment should not affect --><svg xmlns="http://www.w3.org/2000/svg"`,
		"image/svg+xml",
		none,
	},

	{"swf", "CWS", "application/x-shockwave-flash", one},
	{"tar", fromDisk("tar.tar"), "application/x-tar", all},
	{"tcl", "#!/usr/bin/tcl", "text/x-tcl", one},
	{"tcx", `<?xml version="1.0"?><TrainingCenterDatabase xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2">`, "application/vnd.garmin.tcx+xml", one},
	{"tiff", "II*\x00", "image/tiff", one},
	{"tsv", "a\t\"b\"\tc\n1\t2\t3", "text/tab-separated-values", all},
	{"ttc", "ttcf\x00\x01\x00\x00", "font/collection", one},
	{"ttf", "\x00\x01\x00\x00", "font/ttf", one},
	{"tzfile", fromDisk("tzfile"), "application/tzif", one},
	{"utf16bebom txt", "\xfe\xff\x00\x74\x00\x68\x00\x69\x00\x73", "text/plain; charset=utf-16be", none},
	{"utf16lebom txt", "\xff\xfe\x74\x00\x68\x00\x69\x00\x73\x00", "text/plain; charset=utf-16le", none},
	{"utf32bebom txt", "\x00\x00\xfe\xff\x00\x00\x00\x74\x00\x00\x00\x68\x00\x00\x00\x69\x00\x00\x00\x73", "text/plain; charset=utf-32be", none},
	{"utf32lebom txt", "\xff\xfe\x00\x00\x74\x00\x00\x00\x68\x00\x00\x00\x69\x00\x00\x00\x73\x00\x00\x00", "text/plain; charset=utf-32le", none},
	{"utf8 txt", fromDisk("utf8.txt"), "text/plain; charset=utf-8", all},
	{"utf8ctrlchars", "\xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd\x10", "application/octet-stream", none},
	{"vcf", "BEGIN:VCARD\nV", "text/vcard", one},
	{"vcf dos", "BEGIN:VCARD\r\nV", "text/vcard", none},
	{"visio", "\x50\x4b\x03\x04\x14\x00\x00\x00\x00\x00\x83\x93\x11\x5b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x06\x00\x20\x00visio/", "application/vnd.ms-visio.drawing.main+xml", one},
	{"voc", "Creative Voice File", "audio/x-unknown", one},
	{"vtt", "WEBVTT", "text/vtt", one},
	{"warc", "WARC/1.1", "application/warc", one},
	{"wasm", "\x00asm", "application/wasm", one},
	{"wav", "RIFF\xba\xa5\x04\x00WAVEf", "audio/wav", all},
	{"webm", "\x1aE\xdf\xa3\x01\x00\x00\x00\x00\x00\x00\x1fB\x86\x81\x01B\xf7\x81\x01B\xf2\x81\x04B\xf3\x81\x08B\x82\x84webm", "video/webm", all},
	{"webp", "RIFFhv\x00\x00WEBPV", "image/webp", all},
	{"woff", "wOFF", "font/woff", one},
	{"woff2", "wOF2", "font/woff2", one},
	{"x3d", `<?xml version="1.0"?><X3D xmlns:xsd="http://www.w3.org/2001/XMLSchema-instance">`, "model/x3d+xml", one},
	{"xar", "xar!", "application/x-xar", one},
	{"xcf", "gimp xcf", "image/x-xcf", one},
	{"xfdf", `<?xml version="1.0"?><xfdf xmlns="http://ns.adobe.com/xfdf/">`, "application/vnd.adobe.xfdf", one},
	{"xhtml1", `<?xml version="1.0"?><!DOCTYPE html`, "application/xhtml+xml", one},
	{"xhtml2", `<?xml version="1.0"?><HtMl 	XMLNS=`, "application/xhtml+xml", none},
	{"xlf", `<?xml version="1.0"?><xliff xmlns="urn:oasis:names:tc:xliff:document:1.2">`, "application/x-xliff+xml", one},
	{"xls", fromDisk("xls.xls"), "application/vnd.ms-excel", one},
	{"xlsx", fromDisk("xlsx.xlsx"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", all},
	{"xml", "<?xml ", "text/xml; charset=utf-8", all},
	{"xml withbr", "\x0D\x0A<?xml ", "text/xml; charset=utf-8", none},
	{"xz", "\xfd7zXZ\x00", "application/x-xz", one},
	{"zip", "PK\x03\x04", "application/zip", all},
	{"zst", "(\xb5/\xfd", "application/zstd", all},
	{"zst skippable frame", "\x50\x2A\x4D\x18", "application/zstd", none},
}

func TestDetect(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if mtype := Detect([]byte(tc.data)); mtype.String() != tc.expectedMIME {
				t.Errorf("Detect: Expected: %s != Detected: %s", tc.expectedMIME, mtype.String())
			}
			if mtype, err := DetectReader(strings.NewReader(tc.data)); err != nil {
				t.Errorf("DetectReader: unexpected error: %s", err)
			} else if mtype.String() != tc.expectedMIME {
				t.Errorf("DetectReader: Expected: %s != Detected: %s", tc.expectedMIME, mtype.String())
			}
		})
	}
}

func TestDetectBreakReader(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			br := &breakReader{
				r:         strings.NewReader(tc.data),
				breakSize: 3,
			}
			if mtype, err := DetectReader(br); err != nil {
				t.Errorf("Unexpected error: %s", err)
			} else if mtype.String() != tc.expectedMIME {
				t.Errorf("Expected: %s != Detected: %s", tc.expectedMIME, mtype.String())
			}
		})
	}
}

// This test generates the doc file containing the table with the supported MIMEs.
func TestGenerateSupportedFormats(t *testing.T) {
	f, err := os.OpenFile("supported_mimes.md", os.O_WRONLY|os.O_TRUNC, 0o644)
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
	detectedMIME := Detect([]byte("<html></html>"))
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

	n := 1000
	Extend(func([]byte, uint32) bool { return false }, "e", ".e")
	go func() {
		for i := 0; i < n; i++ {
			Detect([]byte("text content"))
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < n; i++ {
			SetLimit(5000 + uint32(i))
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < n; i++ {
			Lookup("text/plain")
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < n; i++ {
			Lookup("e").Extend(func([]byte, uint32) bool { return false }, "e", ".e")
		}
		wg.Done()
	}()

	wg.Wait()
	// Reset to the original limit and MIME tree structure for benchmarks.
	SetLimit(defaultLimit)
	root.children = root.children[1:]
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

func BenchmarkOne(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	// randData is used for the negative case benchmark.
	randData := make([]byte, defaultLimit)
	if _, err := io.ReadFull(r, randData); err != io.ErrUnexpectedEOF && err != nil {
		b.Fatal(err)
	}

	for _, tc := range testcases {
		if shouldRun := tc.bench == one || tc.bench == all; !shouldRun {
			continue
		}
		data := []byte(tc.data)
		parsed, _, _ := mime.ParseMediaType(tc.expectedMIME)
		mtype := Lookup(parsed)
		if mtype == nil || mtype.detector == nil {
			b.Fatalf("mime should always be non-nil %s %s", mtype, tc.expectedMIME)
		}
		// data is used for the positive case benchmark.
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				if !mtype.detector(data, defaultLimit) {
					b.Fatalf("positive detection should never fail")
				}
				if mtype.detector(randData, defaultLimit) {
					b.Fatalf("negative detection should always fail")
				}
			}
		})
	}
}

func BenchmarkAll(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	// randData is used for the negative case benchmark.
	randData := make([]byte, defaultLimit)
	if _, err := io.ReadFull(r, randData); err != io.ErrUnexpectedEOF && err != nil {
		b.Fatal(err)
	}
	for _, tc := range testcases {
		if tc.bench != all {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				// Benchmark both positive and negative.
				Detect([]byte(tc.data))
				Detect(randData)
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
			// Revert the Extend to restore previous MIME tree structure.
			tt.parent.children = tt.parent.children[1:]
		})
	}
}

// Because of the random nature of fuzzing I don't think there is a way to test
// the correctness of the Detect results. Still there is value in fuzzing in
// search for panics.
func FuzzMimetype(f *testing.F) {
	for _, tc := range testcases {
		if len(tc.data) < 100 && tc.bench == one {
			f.Add([]byte(tc.data))
		}
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

func TestInputIsNotMutated(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			passedBytes := []byte(tc.data)
			Detect(passedBytes)

			if pbs := string(passedBytes); pbs != tc.data {
				t.Errorf("input should not be mutated; before: %s, after: %s", tc.data, pbs)
			}
		})
	}
}

const (
	// none means the testcase will not be benchmarked.
	none = "none"
	// one means just the detector function will be called for the testcase.
	// For example: .txt input will be fed into .txt detector.
	one = "one"
	// all means all the detector functions will be called for the testcase.
	// For example: .txt input will be fed into all detectors.
	all = "all"
)

// offset prepends n nul bytes to s.
func offset(n int, s string) string {
	prepend := make([]byte, n)
	return string(prepend) + s
}
func fromDisk(path string) string {
	data, err := os.ReadFile("testdata/" + path)
	if err != nil {
		panic(err)
	}
	return string(data)
}
