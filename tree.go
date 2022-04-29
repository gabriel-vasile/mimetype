package mimetype

import (
	"sync"

	"github.com/gabriel-vasile/mimetype/internal/magic"
)

// mimetype stores the list of MIME types in a tree structure with
// "application/octet-stream" at the root of the hierarchy. The hierarchy
// approach minimizes the number of checks that need to be done on the input
// and allows for more precise results once the base type of file has been
// identified.

// OctetStream is a detector which passes for any slice of bytes.
// When a detector passes the check, the children detectors
// are tried in order to find a more accurate MIME type.
var OctetStream = newMIME("application/octet-stream", "",
	func([]byte, uint32) bool { return true },
	Xpm, SevenZ, Zip, Pdf, Fdf, Ole, Ps, Psd, P7s, Ogg, Png, Jpg, Jxl, Jp2, Jpx,
	Jpm, Gif, WebP, Exe, Elf, Ar, Tar, Xar, Bz2, Fits, Tiff, Bmp, Ico, Mp3, Flac,
	Midi, Ape, MusePack, Amr, Wav, Aiff, Au, Mpeg, QuickTime, Mqv, Mp4, WebM,
	ThreeGp, ThreeG2, Avi, Flv, Mkv, Asf, Aac, Voc, AMp4, M4a, M3u, M4v, Rmvb,
	Gzip, Class, Swf, Crx, Ttf, Woff, Woff2, Otf, Ttc, Eot, Wasm, Shx, Dbf, Dcm, Rar,
	DjVu, Mobi, Lit, Bpg, Sqlite3, Dwg, Nes, Lnk, MachO, Qcp, Icns, Heic,
	HeicSeq, Heif, HeifSeq, Hdr, Mrc, Mdb, Accdb, Zstd, Cab, Rpm, Xz, Lzip,
	Torrent, Cpio, TzIf, Xcf, Pat, Gbr, Glb, Avif,
	// Keep text last because it is the slowest check
	Text,
)

// errMIME is returned from Detect functions when err is not nil.
// Detect could return root for erroneous cases, but it needs to lock mu in order to do so.
// errMIME is same as root but it does not require locking.
var errMIME = newMIME("application/octet-stream", "", func([]byte, uint32) bool { return false })

// mu guards access to the root MIME tree. Access to root must be synchonized with this lock.
var mu = &sync.RWMutex{}

// The list of nodes appended to the root node.
var (
	Xz   = newMIME("application/x-xz", ".xz", magic.Xz)
	Gzip = newMIME("application/gzip", ".gz", magic.Gzip).alias(
		"application/x-gzip", "application/x-gunzip", "application/gzipped",
		"application/gzip-compressed", "application/x-gzip-compressed",
		"gzip/document")
	SevenZ = newMIME("application/x-7z-compressed", ".7z", magic.SevenZ)
	Zip    = newMIME("application/zip", ".zip", magic.Zip, Xlsx, Docx, Pptx, Epub, Jar, Odt, Ods, Odp, Odg, Odf, Odc, Sxc).
		alias("application/x-zip", "application/x-zip-compressed")
	Tar = newMIME("application/x-tar", ".tar", magic.Tar)
	Xar = newMIME("application/x-xar", ".xar", magic.Xar)
	Bz2 = newMIME("application/x-bzip2", ".bz2", magic.Bz2)
	Pdf = newMIME("application/pdf", ".pdf", magic.Pdf).
		alias("application/x-pdf")
	Fdf  = newMIME("application/vnd.fdf", ".fdf", magic.Fdf)
	Xlsx = newMIME("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ".xlsx", magic.Xlsx)
	Docx = newMIME("application/vnd.openxmlformats-officedocument.wordprocessingml.document", ".docx", magic.Docx)
	Pptx = newMIME("application/vnd.openxmlformats-officedocument.presentationml.presentation", ".pptx", magic.Pptx)
	Epub = newMIME("application/epub+zip", ".epub", magic.Epub)
	Jar  = newMIME("application/jar", ".jar", magic.Jar)
	Ole  = newMIME("application/x-ole-storage", "", magic.Ole, Msi, Aaf, Msg, Xls, Pub, Ppt, Doc)
	Msi  = newMIME("application/x-ms-installer", ".msi", magic.Msi).
		alias("application/x-windows-installer", "application/x-msi")
	Aaf = newMIME("application/octet-stream", ".aaf", magic.Aaf)
	Doc = newMIME("application/msword", ".doc", magic.Doc).
		alias("application/vnd.ms-word")
	Ppt = newMIME("application/vnd.ms-powerpoint", ".ppt", magic.Ppt).
		alias("application/mspowerpoint")
	Pub = newMIME("application/vnd.ms-publisher", ".pub", magic.Pub)
	Xls = newMIME("application/vnd.ms-excel", ".xls", magic.Xls).
		alias("application/msexcel")
	Msg  = newMIME("application/vnd.ms-outlook", ".msg", magic.Msg)
	Ps   = newMIME("application/postscript", ".ps", magic.Ps)
	Fits = newMIME("application/fits", ".fits", magic.Fits)
	Ogg  = newMIME("application/ogg", ".ogg", magic.Ogg, OggAudio, OggVideo).
		alias("application/x-ogg")
	OggAudio = newMIME("audio/ogg", ".oga", magic.OggAudio)
	OggVideo = newMIME("video/ogg", ".ogv", magic.OggVideo)
	Text     = newMIME("text/plain", ".txt", magic.Text, Html, Svg, Xml, Php, Js, Lua, Perl, Python, Json, NdJson, Rtf, Srt, Tcl, Csv, Tsv, VCard, ICalendar, Warc, Vtt)
	Xml      = newMIME("text/xml", ".xml", magic.XML, Rss, Atom, X3d, Kml, Xliff, Collada, Gml, Gpx, Tcx, Amf, ThreeMF, Xfdf, Owl2)
	Json     = newMIME("application/json", ".json", magic.JSON, GeoJson, Har)
	Har      = newMIME("application/json", ".har", magic.HAR)
	Csv      = newMIME("text/csv", ".csv", magic.Csv)
	Tsv      = newMIME("text/tab-separated-values", ".tsv", magic.Tsv)
	GeoJson  = newMIME("application/geo+json", ".geojson", magic.GeoJSON)
	NdJson   = newMIME("application/x-ndjson", ".ndjson", magic.NdJSON)
	Html     = newMIME("text/html", ".html", magic.HTML)
	Php      = newMIME("text/x-php", ".php", magic.Php)
	Rtf      = newMIME("text/rtf", ".rtf", magic.Rtf)
	Js       = newMIME("application/javascript", ".js", magic.Js).
			alias("application/x-javascript", "text/javascript")
	Srt = newMIME("application/x-subrip", ".srt", magic.Srt).
		alias("application/x-srt", "text/x-srt")
	Vtt    = newMIME("text/vtt", ".vtt", magic.Vtt)
	Lua    = newMIME("text/x-lua", ".lua", magic.Lua)
	Perl   = newMIME("text/x-perl", ".pl", magic.Perl)
	Python = newMIME("application/x-python", ".py", magic.Python)
	Tcl    = newMIME("text/x-tcl", ".tcl", magic.Tcl).
		alias("application/x-tcl")
	VCard     = newMIME("text/vcard", ".vcf", magic.VCard)
	ICalendar = newMIME("text/calendar", ".ics", magic.ICalendar)
	Svg       = newMIME("image/svg+xml", ".svg", magic.Svg)
	Rss       = newMIME("application/rss+xml", ".rss", magic.Rss).
			alias("text/rss")
	Owl2    = newMIME("application/owl+xml", ".owl", magic.Owl2)
	Atom    = newMIME("application/atom+xml", ".atom", magic.Atom)
	X3d     = newMIME("model/x3d+xml", ".x3d", magic.X3d)
	Kml     = newMIME("application/vnd.google-earth.kml+xml", ".kml", magic.Kml)
	Xliff   = newMIME("application/x-xliff+xml", ".xlf", magic.Xliff)
	Collada = newMIME("model/vnd.collada+xml", ".dae", magic.Collada)
	Gml     = newMIME("application/gml+xml", ".gml", magic.Gml)
	Gpx     = newMIME("application/gpx+xml", ".gpx", magic.Gpx)
	Tcx     = newMIME("application/vnd.garmin.tcx+xml", ".tcx", magic.Tcx)
	Amf     = newMIME("application/x-amf", ".amf", magic.Amf)
	ThreeMF = newMIME("application/vnd.ms-package.3dmanufacturing-3dmodel+xml", ".3mf", magic.Threemf)
	Png     = newMIME("image/png", ".png", magic.Png, APng)
	APng    = newMIME("image/vnd.mozilla.apng", ".png", magic.Apng)
	Jpg     = newMIME("image/jpeg", ".jpg", magic.Jpg)
	Jxl     = newMIME("image/jxl", ".jxl", magic.Jxl)
	Jp2     = newMIME("image/jp2", ".jp2", magic.Jp2)
	Jpx     = newMIME("image/jpx", ".jpf", magic.Jpx)
	Jpm     = newMIME("image/jpm", ".jpm", magic.Jpm).
		alias("video/jpm")
	Xpm  = newMIME("image/x-xpixmap", ".xpm", magic.Xpm)
	Bpg  = newMIME("image/bpg", ".bpg", magic.Bpg)
	Gif  = newMIME("image/gif", ".gif", magic.Gif)
	WebP = newMIME("image/webp", ".webp", magic.Webp)
	Tiff = newMIME("image/tiff", ".tiff", magic.Tiff)
	Bmp  = newMIME("image/bmp", ".bmp", magic.Bmp).
		alias("image/x-bmp", "image/x-ms-bmp")
	Ico  = newMIME("image/x-icon", ".ico", magic.Ico)
	Icns = newMIME("image/x-icns", ".icns", magic.Icns)
	Psd  = newMIME("image/vnd.adobe.photoshop", ".psd", magic.Psd).
		alias("image/x-psd", "application/photoshop")
	Heic    = newMIME("image/heic", ".heic", magic.Heic)
	HeicSeq = newMIME("image/heic-sequence", ".heic", magic.HeicSequence)
	Heif    = newMIME("image/heif", ".heif", magic.Heif)
	HeifSeq = newMIME("image/heif-sequence", ".heif", magic.HeifSequence)
	Hdr     = newMIME("image/vnd.radiance", ".hdr", magic.Hdr)
	Avif    = newMIME("image/avif", ".avif", magic.AVIF)
	Mp3     = newMIME("audio/mpeg", ".mp3", magic.Mp3).
		alias("audio/x-mpeg", "audio/mp3")
	Flac = newMIME("audio/flac", ".flac", magic.Flac)
	Midi = newMIME("audio/midi", ".midi", magic.Midi).
		alias("audio/mid", "audio/sp-midi", "audio/x-mid", "audio/x-midi")
	Ape      = newMIME("audio/ape", ".ape", magic.Ape)
	MusePack = newMIME("audio/musepack", ".mpc", magic.MusePack)
	Wav      = newMIME("audio/wav", ".wav", magic.Wav).
			alias("audio/x-wav", "audio/vnd.wave", "audio/wave")
	Aiff = newMIME("audio/aiff", ".aiff", magic.Aiff).alias("audio/x-aiff")
	Au   = newMIME("audio/basic", ".au", magic.Au)
	Amr  = newMIME("audio/amr", ".amr", magic.Amr).
		alias("audio/amr-nb")
	Aac  = newMIME("audio/aac", ".aac", magic.AAC)
	Voc  = newMIME("audio/x-unknown", ".voc", magic.Voc)
	AMp4 = newMIME("audio/mp4", ".mp4", magic.AMp4).
		alias("audio/x-m4a", "audio/x-mp4a")
	M4a = newMIME("audio/x-m4a", ".m4a", magic.M4a)
	M3u = newMIME("application/vnd.apple.mpegurl", ".m3u", magic.M3u).
		alias("audio/mpegurl")
	M4v  = newMIME("video/x-m4v", ".m4v", magic.M4v)
	Mp4  = newMIME("video/mp4", ".mp4", magic.Mp4)
	WebM = newMIME("video/webm", ".webm", magic.WebM).
		alias("audio/webm")
	Mpeg      = newMIME("video/mpeg", ".mpeg", magic.Mpeg)
	QuickTime = newMIME("video/quicktime", ".mov", magic.QuickTime)
	Mqv       = newMIME("video/quicktime", ".mqv", magic.Mqv)
	ThreeGp   = newMIME("video/3gpp", ".3gp", magic.ThreeGP).
			alias("video/3gp", "audio/3gpp")
	ThreeG2 = newMIME("video/3gpp2", ".3g2", magic.ThreeG2).
		alias("video/3g2", "audio/3gpp2")
	Avi = newMIME("video/x-msvideo", ".avi", magic.Avi).
		alias("video/avi", "video/msvideo")
	Flv = newMIME("video/x-flv", ".flv", magic.Flv)
	Mkv = newMIME("video/x-matroska", ".mkv", magic.Mkv)
	Asf = newMIME("video/x-ms-asf", ".asf", magic.Asf).
		alias("video/asf", "video/x-ms-wmv")
	Rmvb  = newMIME("application/vnd.rn-realmedia-vbr", ".rmvb", magic.Rmvb)
	Class = newMIME("application/x-java-applet", ".class", magic.Class)
	Swf   = newMIME("application/x-shockwave-flash", ".swf", magic.SWF)
	Crx   = newMIME("application/x-chrome-extension", ".crx", magic.CRX)
	Ttf   = newMIME("font/ttf", ".ttf", magic.Ttf).
		alias("font/sfnt", "application/x-font-ttf", "application/font-sfnt")
	Woff    = newMIME("font/woff", ".woff", magic.Woff)
	Woff2   = newMIME("font/woff2", ".woff2", magic.Woff2)
	Otf     = newMIME("font/otf", ".otf", magic.Otf)
	Ttc     = newMIME("font/collection", ".ttc", magic.Ttc)
	Eot     = newMIME("application/vnd.ms-fontobject", ".eot", magic.Eot)
	Wasm    = newMIME("application/wasm", ".wasm", magic.Wasm)
	Shp     = newMIME("application/octet-stream", ".shp", magic.Shp)
	Shx     = newMIME("application/octet-stream", ".shx", magic.Shx, Shp)
	Dbf     = newMIME("application/x-dbf", ".dbf", magic.Dbf)
	Exe     = newMIME("application/vnd.microsoft.portable-executable", ".exe", magic.Exe)
	Elf     = newMIME("application/x-elf", "", magic.Elf, ElfObj, ElfExe, ElfLib, ElfDump)
	ElfObj  = newMIME("application/x-object", "", magic.ElfObj)
	ElfExe  = newMIME("application/x-executable", "", magic.ElfExe)
	ElfLib  = newMIME("application/x-sharedlib", ".so", magic.ElfLib)
	ElfDump = newMIME("application/x-coredump", "", magic.ElfDump)
	Ar      = newMIME("application/x-archive", ".a", magic.Ar, Deb).
		alias("application/x-unix-archive")
	Deb = newMIME("application/vnd.debian.binary-package", ".deb", magic.Deb)
	Rpm = newMIME("application/x-rpm", ".rpm", magic.RPM)
	Dcm = newMIME("application/dicom", ".dcm", magic.Dcm)
	Odt = newMIME("application/vnd.oasis.opendocument.text", ".odt", magic.Odt, Ott).
		alias("application/x-vnd.oasis.opendocument.text")
	Ott = newMIME("application/vnd.oasis.opendocument.text-template", ".ott", magic.Ott).
		alias("application/x-vnd.oasis.opendocument.text-template")
	Ods = newMIME("application/vnd.oasis.opendocument.spreadsheet", ".ods", magic.Ods, Ots).
		alias("application/x-vnd.oasis.opendocument.spreadsheet")
	Ots = newMIME("application/vnd.oasis.opendocument.spreadsheet-template", ".ots", magic.Ots).
		alias("application/x-vnd.oasis.opendocument.spreadsheet-template")
	Odp = newMIME("application/vnd.oasis.opendocument.presentation", ".odp", magic.Odp, Otp).
		alias("application/x-vnd.oasis.opendocument.presentation")
	Otp = newMIME("application/vnd.oasis.opendocument.presentation-template", ".otp", magic.Otp).
		alias("application/x-vnd.oasis.opendocument.presentation-template")
	Odg = newMIME("application/vnd.oasis.opendocument.graphics", ".odg", magic.Odg, Otg).
		alias("application/x-vnd.oasis.opendocument.graphics")
	Otg = newMIME("application/vnd.oasis.opendocument.graphics-template", ".otg", magic.Otg).
		alias("application/x-vnd.oasis.opendocument.graphics-template")
	Odf = newMIME("application/vnd.oasis.opendocument.formula", ".odf", magic.Odf).
		alias("application/x-vnd.oasis.opendocument.formula")
	Odc = newMIME("application/vnd.oasis.opendocument.chart", ".odc", magic.Odc).
		alias("application/x-vnd.oasis.opendocument.chart")
	Sxc = newMIME("application/vnd.sun.xml.calc", ".sxc", magic.Sxc)
	Rar = newMIME("application/x-rar-compressed", ".rar", magic.RAR).
		alias("application/x-rar")
	DjVu    = newMIME("image/vnd.djvu", ".djvu", magic.DjVu)
	Mobi    = newMIME("application/x-mobipocket-ebook", ".mobi", magic.Mobi)
	Lit     = newMIME("application/x-ms-reader", ".lit", magic.Lit)
	Sqlite3 = newMIME("application/vnd.sqlite3", ".sqlite", magic.Sqlite).
		alias("application/x-sqlite3")
	Dwg = newMIME("image/vnd.dwg", ".dwg", magic.Dwg).
		alias("image/x-dwg", "application/acad", "application/x-acad",
			"application/autocad_dwg", "application/dwg", "application/x-dwg",
			"application/x-autocad", "drawing/dwg")
	Warc    = newMIME("application/warc", ".warc", magic.Warc)
	Nes     = newMIME("application/vnd.nintendo.snes.rom", ".nes", magic.Nes)
	Lnk     = newMIME("application/x-ms-shortcut", ".lnk", magic.Lnk)
	MachO   = newMIME("application/x-mach-binary", ".macho", magic.MachO)
	Qcp     = newMIME("audio/qcelp", ".qcp", magic.Qcp)
	Mrc     = newMIME("application/marc", ".mrc", magic.Marc)
	Mdb     = newMIME("application/x-msaccess", ".mdb", magic.MsAccessMdb)
	Accdb   = newMIME("application/x-msaccess", ".accdb", magic.MsAccessAce)
	Zstd    = newMIME("application/zstd", ".zst", magic.Zstd)
	Cab     = newMIME("application/vnd.ms-cab-compressed", ".cab", magic.Cab)
	Lzip    = newMIME("application/lzip", ".lz", magic.Lzip).alias("application/x-lzip")
	Torrent = newMIME("application/x-bittorrent", ".torrent", magic.Torrent)
	Cpio    = newMIME("application/x-cpio", ".cpio", magic.Cpio)
	TzIf    = newMIME("application/tzif", "", magic.TzIf)
	P7s     = newMIME("application/pkcs7-signature", ".p7s", magic.P7s)
	Xcf     = newMIME("image/x-xcf", ".xcf", magic.Xcf)
	Pat     = newMIME("image/x-gimp-pat", ".pat", magic.Pat)
	Gbr     = newMIME("image/x-gimp-gbr", ".gbr", magic.Gbr)
	Xfdf    = newMIME("application/vnd.adobe.xfdf", ".xfdf", magic.Xfdf)
	Glb     = newMIME("model/gltf-binary", ".glb", magic.Glb)
)
