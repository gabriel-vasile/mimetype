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

type testcase struct {
	name         string
	data         string
	expectedMIME string
	// If bench is true, then this entry will be used in benchmarks.
	bench bool
}

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

var testcases = []testcase{
	{"3gpp2", "\x00\x00\x00\x18ftyp3g24", "video/3gpp2", true},
	{"3gpp2 without ftyp", "\x00\x00\x00\x18mtyp3g24", "application/octet-stream", false},
	{"3gp", "\x00\x00\x00\x18ftyp3gp1", "video/3gpp", true},
	{
		"3mf",
		`<?xml version="1.0"?><model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02">`,
		"application/vnd.ms-package.3dmanufacturing-3dmodel+xml",
		true,
	},
	{"7z", "\x37\x7A\xBC\xAF\x27\x1C", "application/x-7z-compressed", true},
	{"a", "\x21\x3C\x61\x72\x63\x68\x3E", "application/x-archive", true},
	{"aac 1", "\xFF\xF1", "audio/aac", true},
	{"aac 2", "\xFF\xF9", "audio/aac", false},
	{"accdb", offset(4, "Standard ACE DB"), "application/x-msaccess", false}, // false because accdb and mdb share the same MIME
	{"aiff", "\x46\x4F\x52\x4D\x00\x00\x00\x00\x41\x49\x46\x46\x00", "audio/aiff", true},
	{"amf", `<?xml version="1.0"?><amf>`, "application/x-amf", true},
	{"amr", "\x23\x21\x41\x4D\x52", "audio/amr", true},
	{"ape", "\x4D\x41\x43\x20\x96\x0F\x00\x00\x34\x00\x00\x00\x18\x00\x00\x00\x90\xE3", "audio/ape", true},
	{"apng", "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A" + offset(29, "acTL"), "image/vnd.mozilla.apng", true},
	{"asf", "\x30\x26\xB2\x75\x8E\x66\xCF\x11\xA6\xD9\x00\xAA\x00\x62\xCE\x6C", "video/x-ms-asf", true},
	{"atom", `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom">`, "application/atom+xml", true},
	{"au", "\x2E\x73\x6E\x64", "audio/basic", true},
	{"avi", "RIFF\x00\x00\x00\x00AVI LIST\x00", "video/x-msvideo", true},
	{"avif", "\x00\x00\x00\x18ftypavif", "image/avif", true},
	{"avis", "\x00\x00\x00\x18ftypavis", "image/avif", false},
	{"bmp", "\x42\x4D", "image/bmp", true},
	{"bpg", "\x42\x50\x47\xFB", "image/bpg", true},
	{"bz2", "\x42\x5A\x68", "application/x-bzip2", true},
	{"cab", "MSCF\x00\x00\x00\x00", "application/vnd.ms-cab-compressed", true},
	{"cab.is", "ISc(\x00\x00\x00\x01", "application/x-installshield", true},
	{"class", "\xCA\xFE\xBA\xBE\x00\x00\x00\xFF", "application/x-java-applet", true},
	{
		"crx",
		"Cr24\x00\x00\x00\x00\x01\x00\x00\x00\x0F\x00\x00\x00" + offset(16, "") + "\x50\x4B\x03\x04",
		"application/x-chrome-extension",
		true,
	},
	{"csv", "1,2,3,4\n5,6,7,8\na,b,c,d", "text/csv", true},
	{"cpio 7", "070707", "application/x-cpio", true},
	{"cpio 1", "070701", "application/x-cpio", false},
	{"cpio 2", "070702", "application/x-cpio", false},
	{"dae", `<?xml version="1.0"?><COLLADA xmlns="http://www.collada.org/2005/11/COLLADASchema">`, "model/vnd.collada+xml", true},
	{"dbf", "\x03\x5f\x07\x1a\x96\x0f\x00\x00\xc1\x00\xa3\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x6f\x73\x6d\x5f\x69\x64\x00\x00\x00\x00\x00\x43\x00\x00\x00\x00\x0a\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x63\x6f\x64\x65", "application/x-dbf", true},
	{"dcm", offset(128, "\x44\x49\x43\x4D"), "application/dicom", true},
	{"deb", "\x21\x3c\x61\x72\x63\x68\x3e\x0a\x64\x65\x62\x69\x61\x6e\x2d\x62\x69\x6e\x61\x72\x79", "application/vnd.debian.binary-package", true},
	{"djvu", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00DJVU", "image/vnd.djvu", true},
	{"djvuM", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00DJVM", "image/vnd.djvu", false},
	{"djvuI", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00DJVI", "image/vnd.djvu", false},
	{"djvuTHUM", "\x41\x54\x26\x54\x46\x4F\x52\x4D\x00\x00\x00\x00THUM", "image/vnd.djvu", false},
	{"doc", fromDisk("doc.doc"), "application/msword", true},
	{"docx", fromDisk("docx.docx"), "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true},
	{"rpm 1", "\xed\xab\xee\xdb", "application/x-rpm", true},
	{"rpm 2", "drpm", "application/x-rpm", false},
	{"dwg", "\x41\x43\x31\x30\x32\x34", "image/vnd.dwg", false},
	{"eot", "\xbe\x45\x00\x00\xfa\x44\x00\x00\x02\x00\x02\x00\x04\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x90\x01\x00\x00\x00\x00\x4c\x50", "application/vnd.ms-fontobject", true},
	{"epub", "\x50\x4B\x03\x04" + offset(26, "mimetypeapplication/epub+zip"), "application/epub+zip", true},
	{"fdf", "%FDF", "application/vnd.fdf", true},
	{"fits", "\x53\x49\x4d\x50\x4c\x45\x20\x20\x3d\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x54", "application/fits", true},
	{"flac", "\x66\x4C\x61\x43\x00\x00\x00\x22", "audio/flac", true},
	{"flv", "\x46\x4C\x56\x01", "video/x-flv", true},
	{"gbr", offset(20, "GIMP"), "image/x-gimp-gbr", true},
	{"geojson", `{"type":"Feature"}`, "application/geo+json", false},
	{"gif 87", "GIF87a", "image/gif", true},
	{"gif 89", "GIF89a", "image/gif", false},
	{"glb 1", "\x67\x6C\x54\x46\x02\x00\x00\x00", "model/gltf-binary", true},
	{"glb 2", "\x67\x6C\x54\x46\x01\x00\x00\x00", "model/gltf-binary", false},
	{"gml", `<?xml version="1.0"?><any xmlns:gml="http://www.opengis.net/gml">`, "application/gml+xml", true},
	{"gml3.2", `<?xml version="1.0"?><any xmlns:gml="http://www.opengis.net/gml/3.2">`, "application/gml+xml", false},
	{"gml3.3", `<?xml version="1.0"?><any xmlns:gml="http://www.opengis.net/gml/3.3/exr">`, "application/gml+xml", false},
	{"gpx", `<?xml version="1.0"?><gpx xmlns="http://www.topografix.com/GPX/1/1">`, "application/gpx+xml", true},
	{"gz", "\x1F\x8B", "application/gzip", true},
	{"har", `{"log":{ "version": "1.2"}}`, "application/json", true},
	{"hdr", "#?RADIANCE\n", "image/vnd.radiance", true},
	{"heic", "\x00\x00\x00\x18ftypheic", "image/heic", true},
	{"heix", "\x00\x00\x00\x18ftypheix", "image/heic", false},
	{"heif mif1", "\x00\x00\x00\x18ftypmif1", "image/heif", true},
	{"heif heim", "\x00\x00\x00\x18ftypheim", "image/heif", false},
	{"heif heis", "\x00\x00\x00\x18ftypheis", "image/heif", false},
	{"heif avic", "\x00\x00\x00\x18ftypavic", "image/heif", false},
	{"html", `<HtMl><bOdY>blah blah blah</body></html>`, "text/html; charset=utf-8", true},
	{"html empty", `<HTML></HTML>`, "text/html; charset=utf-8", false},
	{"html just header", `   <!DOCTYPE HTML>...`, "text/html; charset=utf-8", false},
	{"line ending before html", "\r\n<html>...", "text/html; charset=utf-8", false},
	{
		"html with encoding",
		`<html><head><meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1">`,
		"text/html; charset=iso-8859-1",
		false,
	},
	{"ico 01", "\x00\x00\x01\x00", "image/x-icon", true},
	{"ico 02", "\x00\x00\x02\x00", "image/x-icon", false},
	{"ics", "BEGIN:VCALENDAR\n00", "text/calendar", true},
	{"ics dos", "BEGIN:VCALENDAR\r\n00", "text/calendar", false},
	{"txt iso88591", "\x0a\xe6\xf8\xe6\xf8\xe5\xe6\xf8\xe5\xe5\x0a", "text/plain; charset=iso-8859-1", false},
	{"jar", fromDisk("jar.jar"), "application/jar", true},
	{"jp2", "\x00\x00\x00\x0c\x6a\x50\x20\x20\x0d\x0a\x87\x0a\x00\x00\x00\x14\x66\x74\x79\x70\x6a\x70\x32\x20", "image/jp2", true},
	{"jpf", "\x00\x00\x00\x0c\x6a\x50\x20\x20\x0d\x0a\x87\x0a\x00\x00\x00\x1c\x66\x74\x79\x70\x6a\x70\x78\x20", "image/jpx", true},
	{"jpg", "\xFF\xD8\xFF", "image/jpeg", true},
	{"jpm", "\x00\x00\x00\x0c\x6a\x50\x20\x20\x0d\x0a\x87\x0a\x00\x00\x00\x14\x66\x74\x79\x70\x6a\x70\x6d\x20", "image/jpm", true},
	{"jxl 1", "\xFF\x0A", "image/jxl", true},
	{"jxl 2", "\x00\x00\x00\x0cJXL\x20\x0d\x0a\x87\x0a", "image/jxl", false},
	{"jxr", "\x49\x49\xBC\x01", "image/jxr", true},
	{"xpm", "\x2F\x2A\x20\x58\x50\x4D\x20\x2A\x2F", "image/x-xpixmap", true},
	{"js", "#!/bin/node ", "text/javascript", true},
	{"json", `{"key":"val"}`, "application/json", true},
	{"json issue#239", "{\x0A\x09\x09\"key\":\"val\"}\x0A", "application/json", false},
	// json.{int,string}.txt contain a single JSON value. They are valid JSON
	// documentsthey should not be detected as application/json. This mimics
	// the behaviour of the file utility and seems the correct thing to do.
	{"json.int.txt", "1", "text/plain; charset=utf-8", false},
	{"json.float.txt", "1.5", "text/plain; charset=utf-8", false},
	{"json.string.txt", `"some string"`, "text/plain; charset=utf-8", false},
	{"kml 2.2", `<?xml version="1.0"?><kml xmlns="http://www.opengis.net/kml/2.2">`, "application/vnd.google-earth.kml+xml", true},
	{"kml 2.0", `<?xml version="1.0"?><kml xmlns="http://earth.google.com/kml/2.0">`, "application/vnd.google-earth.kml+xml", false},
	{"kml 2.1", `<?xml version="1.0"?><kml xmlns="http://earth.google.com/kml/2.1">`, "application/vnd.google-earth.kml+xml", false},
	{"kml 2.2", `<?xml version="1.0"?><kml xmlns="http://earth.google.com/kml/2.2">`, "application/vnd.google-earth.kml+xml", false},
	{"lit", "ITOLITLS", "application/x-ms-reader", true},
	{"lua", "#!/usr/bin/lua", "text/x-lua", true},
	{"lua space", "#! /usr/bin/lua", "text/x-lua", false},
	{"lz", "\x4c\x5a\x49\x50", "application/lzip", true},
	{"m3u", "#EXTM3U", "application/vnd.apple.mpegurl", true},
	{"m4a", "\x00\x00\x00\x18ftypM4A ", "audio/x-m4a", true},
	{"audio mp4 F4A", "\x00\x00\x00\x18ftypF4A ", "audio/mp4", true},
	{"audio mp4 F4B", "\x00\x00\x00\x18ftypF4B ", "audio/mp4", false},
	{"audio mp4 M4B", "\x00\x00\x00\x18ftypM4B ", "audio/mp4", false},
	{"audio mp4 M4P", "\x00\x00\x00\x18ftypM4P ", "audio/mp4", false},
	{"audio mp4 MSNV", "\x00\x00\x00\x18ftypMSNV", "audio/mp4", false},
	{"audio mp4 NDAS", "\x00\x00\x00\x18ftypNDAS", "audio/mp4", false},
	{"lnk", "\x4C\x00\x00\x00\x01\x14\x02\x00", "application/x-ms-shortcut", true},
	{"mdb", offset(4, "Standard Jet DB"), "application/x-msaccess", true},
	{"midi", "\x4D\x54\x68\x64", "audio/midi", true},
	{"mkv", "\x1a\x45\xdf\xa3\x01\x00\x00\x00\x00\x00\x00\x23\x42\x86\x81\x01\x42\xf7\x81\x01\x42\xf2\x81\x04\x42\xf3\x81\x08\x42\x82\x88\x6d\x61\x74\x72\x6f\x73\x6b\x61", "video/x-matroska", true},
	{"mobi", offset(60, "BOOKMOBI"), "application/x-mobipocket-ebook", true},
	{"mov", "\x00\x00\x00\x14\x66\x74\x79\x70\x71\x74\x20\x20", "video/quicktime", true},
	{"mp3", "\x49\x44\x33\x03", "audio/mpeg", true},
	{"mp3 v1 notag", "\xff\xfb\xc8\x00", "audio/mpeg", false},
	{"mp3 v2.5 notag", "\xff\xe3\x18\xc4", "audio/mpeg", false},
	{"mp3 v2 notag", "\xff\xf3\x82\xc4", "audio/mpeg", false},
	{"mp4 1", "\x00\x00\x00\x18ftyp0000", "video/mp4", false},
	{"mpc", "MPCK", "audio/musepack", true},
	{"mpeg", "\x00\x00\x01\xba", "video/mpeg", true},
	{"mqv", "\x00\x00\x00\x18ftypmqt ", "video/quicktime", false},
	{"mrc", "00057     2200037   4500245001900000\x1e", "application/marc", true},
	{"msi", fromDisk("msi.msi"), "application/x-ms-installer", true},
	{"msg", fromDisk("msg.msg"), "application/vnd.ms-outlook", true},
	{"ndjson", `{"key":"val"}` + "\n" + `{"key":"val"}`, "application/x-ndjson", true},
	{"nes", "NES\x1a", "application/vnd.nintendo.snes.rom", true},
	{"elfobject", "\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00", "application/x-object", true},
	{"odf", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xb1Z\xa8N\x07\x8a\xa8[*\x00\x00\x00*\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.formula", "application/vnd.oasis.opendocument.formula", true},
	{"sxc", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xbb\x03\x5eGE\xbc\x13\x94\x1c\x00\x00\x00\x1c\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.sun.xml.calc", "application/vnd.sun.xml.calc", true},
	{"odg", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xcbY\xa8N\x9f\x03.\xc4\x2b\x00\x00\x00\x2b\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.graphics", "application/vnd.oasis.opendocument.graphics", true},
	{"odp", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xbdX\xa8N3&\xac\xa8/\x00\x00\x00/\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.presentation", "application/vnd.oasis.opendocument.presentation", true},
	{"ods", "PK\x03\x04\x14\x00\x00\x08\x00\x00\x14V\xa8N\x85l9\x8a.\x00\x00\x00.\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.spreadsheet", "application/vnd.oasis.opendocument.spreadsheet", true},
	{"odt", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xbbP\xa8N\x5e\xc62\n'\x00\x00\x00'\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.text", "application/vnd.oasis.opendocument.text", true},
	{"ogg", "OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\xce\xc6AI\x00\x00\x00\x00py\xf3\x3d\x01\x1e\x01vorbis\x00\x00", "audio/ogg", true},
	{"ogg", "OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x80\xbc\x81_\x00\x00\x00\x00\xd0\xfbP\x84\x01@fishead\x00\x03", "video/ogg", true},
	{"ogg spx oga", "OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\xc7w\xaa\x15\x00\x00\x00\x00V&\x88\x89\x01PSpeex   1", "audio/ogg", true},
	{"otf", "OTTO\x00", "font/otf", true},
	{"otg", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xd1Y\xa8N\xdf%\xad\xe94\x00\x00\x004\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.graphics-template", "application/vnd.oasis.opendocument.graphics-template", true},
	{"otp", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xc4X\xa8N\xef\n\x14:8\x00\x00\x008\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.presentation-template", "application/vnd.oasis.opendocument.presentation-template", true},
	{"ots", "PK\x03\x04\x14\x00\x00\x08\x00\x00\x1bV\xa8N{\x96\xa3N7\x00\x00\x007\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.spreadsheet-template", "application/vnd.oasis.opendocument.spreadsheet-template", true},
	{"ott", "PK\x03\x04\x14\x00\x00\x08\x00\x00\xcfP\xa8N\xe4\x11\x92)0\x00\x00\x000\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.text-template", "application/vnd.oasis.opendocument.text-template", true},
	{"odc", "PK\x03\x04\x14\x00\x00\x08\x00\x00zp2R\xab\xb8\xb2l(\x00\x00\x00(\x00\x00\x00\x08\x00\x00\x00mimetypeapplication/vnd.oasis.opendocument.chart", "application/vnd.oasis.opendocument.chart", true},
	{"owl", `<?xml version="1.0"?><Ontology xmlns="http://www.w3.org/2002/07/owl#">`, "application/owl+xml", true},
	{"pat", "\x00\x00\x00\x1c\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x03GPAT", "image/x-gimp-pat", true},
	{"pdf", "%PDF-", "application/pdf", true},
	{"php", "#!/usr/bin/env php", "text/x-php", true},
	{"pl", "#!/usr/bin/perl", "text/x-perl", true},
	{"png", "\x89PNG\x0d\x0a\x1a\x0a", "image/png", true},
	{"ppt", fromDisk("ppt.ppt"), "application/vnd.ms-powerpoint", true},
	{"pptx", fromDisk("pptx.pptx"), "application/vnd.openxmlformats-officedocument.presentationml.presentation", true},
	{"ps", "%!PS-Adobe-", "application/postscript", true},
	{"psd", "8BPS", "image/vnd.adobe.photoshop", true},
	{"p7s_pem", "-----BEGIN PKCS7", "application/pkcs7-signature", true},
	{"p7s_der", "\x30\x82\x01\x26\x06\x09\x2a\x86\x48\x86\xf7\x0d\x01\x07\x02\xa0\x82\x01\x17\x30", "application/pkcs7-signature", true},
	{"pub", fromDisk("pub.pub"), "application/vnd.ms-publisher", true},
	{"py", "#!/usr/bin/python", "text/x-python", true},
	{"qcp", "RIFF\xc0\xcf\x00\x00QLCMf", "audio/qcelp", true},
	{"rar", "Rar!\x1a\x07\x01\x00", "application/x-rar-compressed", true},
	{"rmvb", ".RMF", "application/vnd.rn-realmedia-vbr", true},
	{"rpm", "\xed\xab\xee\xdb", "application/x-rpm", true},
	{"rss", "\x3c\x3f\x78\x6d\x6c\x20\x76\x65\x72\x73\x69\x6f\x6e\x3d\x22\x31\x2e\x30\x22\x20\x65\x6e\x63\x6f\x64\x69\x6e\x67\x3d\x22\x55\x54\x46\x2d\x38\x22\x3f\x3e\x0a\x3c\x72\x73\x73", "application/rss+xml", true},
	{"rtf", "{\\rtf", "text/rtf", true},
	{"shp", fromDisk("shp.shp"), "application/vnd.shp", true},
	{"shx", "\x00\x00\x27\x0a", "application/vnd.shx", true},
	{"so", "\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x03\x00", "application/x-sharedlib", true},
	{"sqlite", "SQLite format 3\x00", "application/vnd.sqlite3", true},
	{"srt", "1\n00:02:16,612 --\x3e 00:02:19,376\nS", "application/x-subrip", true},
	{"svg", "<svg", "image/svg+xml", true},
	{"swf", "CWS", "application/x-shockwave-flash", true},
	{"tar", fromDisk("tar.tar"), "application/x-tar", true},
	{"tcl", "#!/usr/bin/tcl", "text/x-tcl", true},
	{"tcx", `<?xml version="1.0"?><TrainingCenterDatabase xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2">`, "application/vnd.garmin.tcx+xml", true},
	{"tiff", "II*\x00", "image/tiff", true},
	{"tsv", "a\tb\tc\n1\t2\t3", "text/tab-separated-values", true},
	{"ttc", "ttcf\x00\x01\x00\x00", "font/collection", true},
	{"ttf", "\x00\x01\x00\x00", "font/ttf", true},
	{"tzfile", fromDisk("tzfile"), "application/tzif", true},
	{"utf16bebom txt", "\xfe\xff\x00\x74\x00\x68\x00\x69\x00\x73", "text/plain; charset=utf-16be", false},
	{"utf16lebom txt", "\xff\xfe\x74\x00\x68\x00\x69\x00\x73\x00", "text/plain; charset=utf-16le", false},
	{"utf32bebom txt", "\x00\x00\xfe\xff\x00\x00\x00\x74\x00\x00\x00\x68\x00\x00\x00\x69\x00\x00\x00\x73", "text/plain; charset=utf-32be", false},
	{"utf32lebom txt", "\xff\xfe\x00\x00\x74\x00\x00\x00\x68\x00\x00\x00\x69\x00\x00\x00\x73\x00\x00\x00", "text/plain; charset=utf-32le", false},
	{"utf8 txt", fromDisk("utf8.txt"), "text/plain; charset=utf-8", true},
	{"utf8ctrlchars", "\xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd\x10", "application/octet-stream", false},
	{"vcf", "BEGIN:VCARD\nV", "text/vcard", true},
	{"vcf dos", "BEGIN:VCARD\r\nV", "text/vcard", false},
	{"voc", "Creative Voice File", "audio/x-unknown", true},
	{"vtt", "WEBVTT", "text/vtt", true},
	{"warc", "WARC/1.1", "application/warc", true},
	{"wasm", "\x00asm", "application/wasm", true},
	{"wav", "RIFF\xba\xa5\x04\x00WAVEf", "audio/wav", true},
	{"webm", "\x1aE\xdf\xa3\x01\x00\x00\x00\x00\x00\x00\x1fB\x86\x81\x01B\xf7\x81\x01B\xf2\x81\x04B\xf3\x81\x08B\x82\x84webm", "video/webm", true},
	{"webp", "RIFFhv\x00\x00WEBPV", "image/webp", true},
	{"woff", "wOFF", "font/woff", true},
	{"woff2", "wOF2", "font/woff2", true},
	{"x3d", `<?xml version="1.0"?><X3D xmlns:xsd="http://www.w3.org/2001/XMLSchema-instance">`, "model/x3d+xml", true},
	{"xar", "xar!", "application/x-xar", true},
	{"xcf", "gimp xcf", "image/x-xcf", true},
	{"xfdf", `<?xml version="1.0"?><xfdf xmlns="http://ns.adobe.com/xfdf/">`, "application/vnd.adobe.xfdf", true},
	{"xlf", `<?xml version="1.0"?><xliff xmlns="urn:oasis:names:tc:xliff:document:1.2">`, "application/x-xliff+xml", true},
	{"xls", fromDisk("xls.xls"), "application/vnd.ms-excel", true},
	{"xlsx", fromDisk("xlsx.xlsx"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true},
	{"xml", "<?xml ", "text/xml; charset=utf-8", true},
	{"xml withbr", "\x0D\x0A<?xml ", "text/xml; charset=utf-8", false},
	{"xz", "\xfd7zXZ\x00", "application/x-xz", true},
	{"zip", "PK\x03\x04", "application/zip", true},
	{"zst", "(\xb5/\xfd", "application/zstd", true},
	{"zst skippable frame", "\x50\x2A\x4D\x18", "application/zstd", false},
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
		data := []byte(tc.data)
		parsed, _, _ := mime.ParseMediaType(tc.expectedMIME)
		mtype := Lookup(parsed)
		if mtype == nil || mtype.detector == nil {
			b.Fatalf("nu e bine %s %s", mtype, tc.expectedMIME)
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
	for _, tc := range testcases {
		if len(tc.data) < 100 {
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
