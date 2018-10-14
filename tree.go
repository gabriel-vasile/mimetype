package mimetype

import "github.com/gabriel-vasile/mimetype/matchers"

// Root is a matcher which passes for any slice of bytes.
// When a matcher passes the check, the children matchers are tried in order to
// find a more accurate mime type
var Root = NewNode("application/octet-stream", "", matchers.True,
	SevenZ, Zip, Pdf, Png, Jpg, Gif, Webp, Tiff, Mp3, Flac, Midi, Ape, MusePack,
	Wav, Aiff, Mpeg, Au, Quicktime, Mp4, Ogg, WebM, ThreeGP, Avi, Flv, Ps, Psd, Txt,
	Doc, Xls, Ppt)

var (
	SevenZ = NewNode("application/x-7z-compressed", "7z", matchers.SevenZ)
	Zip    = NewNode("application/zip", "zip", matchers.Zip, Xlsx, Docx, Pptx, Epub, Jar)
	Pdf    = NewNode("application/pdf", "pdf", matchers.Pdf)
	Xlsx   = NewNode("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "xlsx", matchers.Xlsx)
	Docx   = NewNode("application/vnd.openxmlformats-officedocument.wordprocessingml.document", "docx", matchers.Docx)
	Pptx   = NewNode("application/vnd.openxmlformats-officedocument.presentationml.presentation", "pptx", matchers.Pptx)
	Epub   = NewNode("application/epub+zip", "epub", matchers.Epub)
	Jar    = NewNode("application/jar", "jar", matchers.Jar, Apk)
	Apk    = NewNode("application/vnd.android.package-archive", "apk", matchers.False)
	Doc    = NewNode("application/msword", "doc", matchers.Doc)
	Ppt    = NewNode("application/vnd.ms-powerpoint", "ppt", matchers.Ppt)
	Xls    = NewNode("application/vnd.ms-excel", "xls", matchers.Xls)
	Ps     = NewNode("application/postscript", "ps", matchers.Ps)
	Psd    = NewNode("application/x-photoshop", "psd", matchers.Psd)
	Ogg    = NewNode("application/ogg", "ogg", matchers.Ogg)

	Txt = NewNode("text/plain", "txt", matchers.Txt,
		Html, Xml, Php, Js, Lua, Perl, Python, Json, Rtf)
	Xml = NewNode("text/xml; charset=utf-8", "xml", matchers.Xml,
		Svg, X3d, Kml, Collada, Gml, Gpx)
	Json = NewNode("application/json", "json", matchers.Json)
	Html = NewNode("text/html; charset=utf-8", "html", matchers.Html)
	Php  = NewNode("text/x-php; charset=utf-8", "php", matchers.Php)
	Rtf  = NewNode("text/rtf", "rtf", matchers.Rtf)

	Js     = NewNode("application/javascript", "js", matchers.Js)
	Lua    = NewNode("text/x-lua", "lua", matchers.Lua)
	Perl   = NewNode("text/x-perl", "pl", matchers.Perl)
	Python = NewNode("application/x-python", "py", matchers.Python)

	Svg     = NewNode("image/svg+xml", "svg", matchers.Svg)
	X3d     = NewNode("model/x3d+xml", "x3d", matchers.X3d)
	Kml     = NewNode("application/vnd.google-earth.kml+xml", "kml", matchers.False)
	Collada = NewNode("model/vnd.collada+xml", "dae", matchers.False)
	Gml     = NewNode("application/gml+xml", "gml", matchers.False)
	Gpx     = NewNode("application/gpx+xml", "gpx", matchers.False)

	Png  = NewNode("image/png", "png", matchers.Png)
	Jpg  = NewNode("image/jpeg", "jpg", matchers.Jpg)
	Gif  = NewNode("image/gif", "gif", matchers.Gif)
	Webp = NewNode("image/webp", "webp", matchers.Webp)
	Tiff = NewNode("image/tiff", "tiff", matchers.Tiff)

	Mp3      = NewNode("audio/mpeg", "mp3", matchers.Mp3)
	Flac     = NewNode("audio/flac", "flac", matchers.Flac)
	Midi     = NewNode("audio/midi", "midi", matchers.Midi)
	Ape      = NewNode("audio/ape", "ape", matchers.Ape)
	MusePack = NewNode("audio/musepack", "mpc", matchers.MusePack)
	Wav      = NewNode("audio/wav", "wav", matchers.Wav)
	Aiff     = NewNode("audio/aiff", "aiff", matchers.Aiff)
	Au       = NewNode("audio/basic", "au", matchers.Au)

	Mp4       = NewNode("video/mp4", "mp4", matchers.Mp4)
	WebM      = NewNode("video/webm", "webm", matchers.WebM)
	Mpeg      = NewNode("video/mpeg", "mpeg", matchers.Mpeg)
	Quicktime = NewNode("video/quicktime", "mov", matchers.Quicktime)
	ThreeGP   = NewNode("video/3gp", "3gp", matchers.ThreeGP)
	Avi       = NewNode("video/x-msvideo", "avi", matchers.Avi)
	Flv       = NewNode("video/x-flv", "flv", matchers.Flv)
)
