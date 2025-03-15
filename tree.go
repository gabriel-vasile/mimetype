package mimetype

import (
	"sync"

	"github.com/gabriel-vasile/mimetype/internal/magic"
	"github.com/gabriel-vasile/mimetype/types"
)

// mimetype stores the list of MIME types in a tree structure with
// "application/octet-stream" at the root of the hierarchy. The hierarchy
// approach minimizes the number of checks that need to be done on the input
// and allows for more precise results once the base type of file has been
// identified.
//
// root is a detector which passes for any slice of bytes.
// When a detector passes the check, the children detectors
// are tried in order to find a more accurate MIME type.
var root = newMIME(types.OCTET_STREAM, "",
	func([]byte, uint32) bool { return true },
	xpm, sevenZ, zip, pdf, fdf, ole, ps, psd, p7s, ogg, png, jpg, jxl, jp2, jpx,
	jpm, jxs, gif, webp, exe, elf, ar, tar, xar, bz2, fits, tiff, bmp, ico, mp3,
	flac, midi, ape, musePack, amr, wav, aiff, au, mpeg, quickTime, mp4, webM,
	avi, flv, mkv, asf, aac, voc, m3u, rmvb, gzip, class, swf, crx, ttf, woff,
	woff2, otf, ttc, eot, wasm, shx, dbf, dcm, rar, djvu, mobi, lit, bpg, cbor,
	sqlite3, dwg, nes, lnk, macho, qcp, icns, hdr, mrc, mdb, accdb, zstd, cab,
	rpm, xz, lzip, torrent, cpio, tzif, xcf, pat, gbr, glb, cabIS, jxr, parquet,
	// Keep text last because it is the slowest check.
	text,
)

// errMIME is returned from Detect functions when err is not nil.
// Detect could return root for erroneous cases, but it needs to lock mu in order to do so.
// errMIME is same as root but it does not require locking.
var errMIME = newMIME("application/octet-stream", "", func([]byte, uint32) bool { return false })

// mu guards access to the root MIME tree. Access to root must be synchronized with this lock.
var mu = &sync.RWMutex{}

// The list of nodes appended to the root node.
var (
	xz   = newMIME(types.XZ, ".xz", magic.Xz)
	gzip = newMIME(types.GZIP, ".gz", magic.Gzip).alias(
		"application/x-gzip", "application/x-gunzip", "application/gzipped",
		"application/gzip-compressed", "application/x-gzip-compressed",
		"gzip/document")
	sevenZ = newMIME(types.SEVENZ, ".7z", magic.SevenZ)
	// APK must be checked before JAR because APK is a subset of JAR.
	// This means APK should be a child of JAR detector, but in practice,
	// the decisive signature for JAR might be located at the end of the file
	// and not reachable because of library readLimit.
	zip = newMIME(types.ZIP, ".zip", magic.Zip, xlsx, docx, pptx, epub, apk, jar, odt, ods, odp, odg, odf, odc, sxc).
		alias("application/x-zip", "application/x-zip-compressed")
	tar = newMIME(types.TAR, ".tar", magic.Tar)
	xar = newMIME(types.XAR, ".xar", magic.Xar)
	bz2 = newMIME(types.BZIP2, ".bz2", magic.Bz2)
	pdf = newMIME(types.PDF, ".pdf", magic.Pdf).
		alias("application/x-pdf")
	fdf  = newMIME(types.FDF, ".fdf", magic.Fdf)
	xlsx = newMIME(types.XLSX, ".xlsx", magic.Xlsx)
	docx = newMIME(types.DOCX, ".docx", magic.Docx)
	pptx = newMIME(types.PPTX, ".pptx", magic.Pptx)
	epub = newMIME(types.EPUB, ".epub", magic.Epub)
	jar  = newMIME(types.JAR, ".jar", magic.Jar)
	apk  = newMIME(types.APK, ".apk", magic.APK)
	ole  = newMIME(types.OLE, "", magic.Ole, msi, aaf, msg, xls, pub, ppt, doc)
	msi  = newMIME(types.MSI, ".msi", magic.Msi).
		alias("application/x-windows-installer", "application/x-msi")
	aaf = newMIME(types.OCTET_STREAM, ".aaf", magic.Aaf)
	doc = newMIME(types.DOC, ".doc", magic.Doc).
		alias("application/vnd.ms-word")
	ppt = newMIME(types.PPT, ".ppt", magic.Ppt).
		alias("application/mspowerpoint")
	pub = newMIME(types.PUB, ".pub", magic.Pub)
	xls = newMIME(types.XLS, ".xls", magic.Xls).
		alias("application/msexcel")
	msg  = newMIME(types.MSG, ".msg", magic.Msg)
	ps   = newMIME(types.POSTSCRIPT, ".ps", magic.Ps)
	fits = newMIME(types.FITS, ".fits", magic.Fits)
	ogg  = newMIME(types.OGG, ".ogg", magic.Ogg, oggAudio, oggVideo).
		alias("application/x-ogg")
	oggAudio = newMIME(types.OGGAUDIO, ".oga", magic.OggAudio)
	oggVideo = newMIME(types.OGGVIDEO, ".ogv", magic.OggVideo)
	text     = newMIME(types.TEXT, ".txt", magic.Text, html, svg, xml, php, js, lua, perl, python, json, ndJSON, rtf, srt, tcl, csv, tsv, vCard, iCalendar, warc, vtt)
	xml      = newMIME(types.XML, ".xml", magic.XML, rss, atom, x3d, kml, xliff, collada, gml, gpx, tcx, amf, threemf, xfdf, owl2).
			alias("application/xml")
	json    = newMIME(types.JSON, ".json", magic.JSON, geoJSON, har)
	har     = newMIME(types.JSON, ".har", magic.HAR)
	csv     = newMIME(types.CSV, ".csv", magic.Csv)
	tsv     = newMIME(types.TSV, ".tsv", magic.Tsv)
	geoJSON = newMIME(types.GEOJSON, ".geojson", magic.GeoJSON)
	ndJSON  = newMIME(types.NDJSON, ".ndjson", magic.NdJSON)
	html    = newMIME(types.HTML, ".html", magic.HTML)
	php     = newMIME(types.PHP, ".php", magic.Php)
	rtf     = newMIME(types.RTF, ".rtf", magic.Rtf).alias("application/rtf")
	js      = newMIME(types.JS, ".js", magic.Js).
		alias("application/x-javascript", "application/javascript")
	srt = newMIME(types.SRT, ".srt", magic.Srt).
		alias("application/x-srt", "text/x-srt")
	vtt    = newMIME(types.VTT, ".vtt", magic.Vtt)
	lua    = newMIME(types.LUA, ".lua", magic.Lua)
	perl   = newMIME(types.PERL, ".pl", magic.Perl)
	python = newMIME(types.PYTHON, ".py", magic.Python).
		alias("text/x-script.python", "application/x-python")
	tcl = newMIME(types.TCL, ".tcl", magic.Tcl).
		alias("application/x-tcl")
	vCard     = newMIME(types.VCARD, ".vcf", magic.VCard)
	iCalendar = newMIME(types.ICALENDAR, ".ics", magic.ICalendar)
	svg       = newMIME(types.SVG, ".svg", magic.Svg)
	rss       = newMIME(types.RSS, ".rss", magic.Rss).
			alias("text/rss")
	owl2    = newMIME(types.OWL, ".owl", magic.Owl2)
	atom    = newMIME(types.ATOM, ".atom", magic.Atom)
	x3d     = newMIME(types.X3D, ".x3d", magic.X3d)
	kml     = newMIME(types.KML, ".kml", magic.Kml)
	xliff   = newMIME(types.XLIFF, ".xlf", magic.Xliff)
	collada = newMIME(types.COLLADA, ".dae", magic.Collada)
	gml     = newMIME(types.GML, ".gml", magic.Gml)
	gpx     = newMIME(types.GPX, ".gpx", magic.Gpx)
	tcx     = newMIME(types.TCX, ".tcx", magic.Tcx)
	amf     = newMIME(types.AMF, ".amf", magic.Amf)
	threemf = newMIME(types.THREEMF, ".3mf", magic.Threemf)
	png     = newMIME(types.PNG, ".png", magic.Png, apng)
	apng    = newMIME(types.APNG, ".png", magic.Apng)
	jpg     = newMIME(types.JPG, ".jpg", magic.Jpg)
	jxl     = newMIME(types.JXL, ".jxl", magic.Jxl)
	jp2     = newMIME(types.JP2, ".jp2", magic.Jp2)
	jpx     = newMIME(types.JPX, ".jpf", magic.Jpx)
	jpm     = newMIME(types.JPM, ".jpm", magic.Jpm).
		alias("video/jpm")
	jxs  = newMIME(types.JXS, ".jxs", magic.Jxs)
	xpm  = newMIME(types.XPM, ".xpm", magic.Xpm)
	bpg  = newMIME(types.BPG, ".bpg", magic.Bpg)
	gif  = newMIME("image/gif", ".gif", magic.Gif)
	webp = newMIME(types.WEBP, ".webp", magic.Webp)
	tiff = newMIME(types.TIFF, ".tiff", magic.Tiff)
	bmp  = newMIME(types.BMP, ".bmp", magic.Bmp).
		alias("image/x-bmp", "image/x-ms-bmp")
	ico  = newMIME(types.ICO, ".ico", magic.Ico)
	icns = newMIME(types.ICNS, ".icns", magic.Icns)
	psd  = newMIME(types.PSD, ".psd", magic.Psd).
		alias("image/x-psd", "application/photoshop")
	heic    = newMIME(types.HEIC, ".heic", magic.Heic)
	heicSeq = newMIME(types.HEICSEQ, ".heic", magic.HeicSequence)
	heif    = newMIME(types.HEIF, ".heif", magic.Heif)
	heifSeq = newMIME(types.HEIFSEQ, ".heif", magic.HeifSequence)
	hdr     = newMIME(types.HDR, ".hdr", magic.Hdr)
	avif    = newMIME(types.AVIF, ".avif", magic.AVIF)
	mp3     = newMIME(types.MP3, ".mp3", magic.Mp3).
		alias("audio/x-mpeg", "audio/mp3")
	flac = newMIME(types.FLAC, ".flac", magic.Flac)
	midi = newMIME(types.MIDI, ".midi", magic.Midi).
		alias("audio/mid", "audio/sp-midi", "audio/x-mid", "audio/x-midi")
	ape      = newMIME(types.APE, ".ape", magic.Ape)
	musePack = newMIME(types.MUSEPACK, ".mpc", magic.MusePack)
	wav      = newMIME(types.WAV, ".wav", magic.Wav).
			alias("audio/x-wav", "audio/vnd.wave", "audio/wave")
	aiff = newMIME(types.AIFF, ".aiff", magic.Aiff).
		alias("audio/x-aiff")
	au  = newMIME(types.AU, ".au", magic.Au)
	amr = newMIME(types.AMR, ".amr", magic.Amr).
		alias("audio/amr-nb")
	aac  = newMIME(types.AAC, ".aac", magic.AAC)
	voc  = newMIME(types.VOC, ".voc", magic.Voc)
	aMp4 = newMIME(types.AMP4, ".mp4", magic.AMp4).
		alias("audio/x-mp4a")
	m4a = newMIME(types.M4A, ".m4a", magic.M4a)
	m3u = newMIME(types.M3U, ".m3u", magic.M3u).
		alias("audio/mpegurl")
	m4v  = newMIME(types.M4V, ".m4v", magic.M4v)
	mj2  = newMIME(types.MJ2, ".mj2", magic.Mj2)
	dvb  = newMIME(types.DVB, ".dvb", magic.Dvb)
	mp4  = newMIME(types.MP4, ".mp4", magic.Mp4, avif, threeGP, threeG2, aMp4, mqv, m4a, m4v, heic, heicSeq, heif, heifSeq, mj2, dvb)
	webM = newMIME(types.WEBM, ".webm", magic.WebM).
		alias("audio/webm")
	mpeg      = newMIME(types.MPEG, ".mpeg", magic.Mpeg)
	quickTime = newMIME(types.QUICKTIME, ".mov", magic.QuickTime)
	mqv       = newMIME(types.QUICKTIME, ".mqv", magic.Mqv)
	threeGP   = newMIME(types.THREEGP, ".3gp", magic.ThreeGP).
			alias("video/3gp", "audio/3gpp")
	threeG2 = newMIME(types.THREEG2, ".3g2", magic.ThreeG2).
		alias("video/3g2", "audio/3gpp2")
	avi = newMIME(types.AVI, ".avi", magic.Avi).
		alias("video/avi", "video/msvideo")
	flv = newMIME(types.FLV, ".flv", magic.Flv)
	mkv = newMIME(types.MKV, ".mkv", magic.Mkv)
	asf = newMIME(types.ASF, ".asf", magic.Asf).
		alias("video/asf", "video/x-ms-wmv")
	rmvb  = newMIME(types.RMVB, ".rmvb", magic.Rmvb)
	class = newMIME(types.CLASS, ".class", magic.Class)
	swf   = newMIME(types.SWF, ".swf", magic.SWF)
	crx   = newMIME(types.CRX, ".crx", magic.CRX)
	ttf   = newMIME(types.TTF, ".ttf", magic.Ttf).
		alias("font/sfnt", "application/x-font-ttf", "application/font-sfnt")
	woff    = newMIME(types.WOFF, ".woff", magic.Woff)
	woff2   = newMIME(types.WOFF2, ".woff2", magic.Woff2)
	otf     = newMIME(types.OTF, ".otf", magic.Otf)
	ttc     = newMIME(types.TTC, ".ttc", magic.Ttc)
	eot     = newMIME(types.EOT, ".eot", magic.Eot)
	wasm    = newMIME(types.WASM, ".wasm", magic.Wasm)
	shp     = newMIME(types.SHP, ".shp", magic.Shp)
	shx     = newMIME(types.SHX, ".shx", magic.Shx, shp)
	dbf     = newMIME(types.DBF, ".dbf", magic.Dbf)
	exe     = newMIME(types.EXE, ".exe", magic.Exe)
	elf     = newMIME(types.ELF, "", magic.Elf, elfObj, elfExe, elfLib, elfDump)
	elfObj  = newMIME(types.ELFOBJ, "", magic.ElfObj)
	elfExe  = newMIME(types.ELFEXE, "", magic.ElfExe)
	elfLib  = newMIME(types.ELFLIB, ".so", magic.ElfLib)
	elfDump = newMIME(types.ELFDUMP, "", magic.ElfDump)
	ar      = newMIME(types.AR, ".a", magic.Ar, deb).
		alias("application/x-unix-archive")
	deb = newMIME(types.DEB, ".deb", magic.Deb)
	rpm = newMIME(types.RPM, ".rpm", magic.RPM)
	dcm = newMIME(types.DCM, ".dcm", magic.Dcm)
	odt = newMIME(types.ODT, ".odt", magic.Odt, ott).
		alias("application/x-vnd.oasis.opendocument.text")
	ott = newMIME(types.OTT, ".ott", magic.Ott).
		alias("application/x-vnd.oasis.opendocument.text-template")
	ods = newMIME(types.ODS, ".ods", magic.Ods, ots).
		alias("application/x-vnd.oasis.opendocument.spreadsheet")
	ots = newMIME(types.OTS, ".ots", magic.Ots).
		alias("application/x-vnd.oasis.opendocument.spreadsheet-template")
	odp = newMIME(types.ODP, ".odp", magic.Odp, otp).
		alias("application/x-vnd.oasis.opendocument.presentation")
	otp = newMIME(types.OTP, ".otp", magic.Otp).
		alias("application/x-vnd.oasis.opendocument.presentation-template")
	odg = newMIME(types.ODG, ".odg", magic.Odg, otg).
		alias("application/x-vnd.oasis.opendocument.graphics")
	otg = newMIME(types.OTG, ".otg", magic.Otg).alias("application/x-vnd.oasis.opendocument.graphics-template")
	odf = newMIME(types.ODF, ".odf", magic.Odf).
		alias("application/x-vnd.oasis.opendocument.formula")
	odc = newMIME(types.ODC, ".odc", magic.Odc).
		alias("application/x-vnd.oasis.opendocument.chart")
	sxc = newMIME(types.SXC, ".sxc", magic.Sxc)
	rar = newMIME(types.RAR, ".rar", magic.RAR).
		alias("application/x-rar")
	djvu    = newMIME(types.DJVU, ".djvu", magic.DjVu)
	mobi    = newMIME(types.MOBI, ".mobi", magic.Mobi)
	lit     = newMIME(types.LIT, ".lit", magic.Lit)
	sqlite3 = newMIME(types.SQLITE3, ".sqlite", magic.Sqlite).
		alias("application/x-sqlite3")
	dwg = newMIME(types.DWG, ".dwg", magic.Dwg).
		alias("image/x-dwg", "application/acad", "application/x-acad",
			"application/autocad_dwg", "application/dwg", "application/x-dwg",
			"application/x-autocad", "drawing/dwg")
	warc  = newMIME(types.WARC, ".warc", magic.Warc)
	nes   = newMIME(types.NES, ".nes", magic.Nes)
	lnk   = newMIME(types.LNK, ".lnk", magic.Lnk)
	macho = newMIME(types.MACHO, ".macho", magic.MachO)
	qcp   = newMIME(types.QCP, ".qcp", magic.Qcp)
	mrc   = newMIME(types.MRC, ".mrc", magic.Marc)
	mdb   = newMIME(types.MDB, ".mdb", magic.MsAccessMdb)
	accdb = newMIME(types.ACCDB, ".accdb", magic.MsAccessAce)
	zstd  = newMIME(types.ZSTD, ".zst", magic.Zstd)
	cab   = newMIME(types.CAB, ".cab", magic.Cab)
	cabIS = newMIME(types.CABIS, ".cab", magic.InstallShieldCab)
	lzip  = newMIME(types.LZIP, ".lz", magic.Lzip).
		alias("application/x-lzip")
	torrent = newMIME(types.TORRENT, ".torrent", magic.Torrent)
	cpio    = newMIME(types.CPIO, ".cpio", magic.Cpio)
	tzif    = newMIME(types.TZIF, "", magic.TzIf)
	p7s     = newMIME(types.P7S, ".p7s", magic.P7s)
	xcf     = newMIME(types.XCF, ".xcf", magic.Xcf)
	pat     = newMIME(types.PAT, ".pat", magic.Pat)
	gbr     = newMIME(types.GBR, ".gbr", magic.Gbr)
	xfdf    = newMIME(types.XFDF, ".xfdf", magic.Xfdf)
	glb     = newMIME(types.GLB, ".glb", magic.Glb)
	jxr     = newMIME(types.JXR, ".jxr", magic.Jxr).
		alias("image/vnd.ms-photo")
	parquet = newMIME(types.PARQUET, ".parquet", magic.Par1).
		alias("application/x-parquet")
	cbor = newMIME(types.CBOR, ".cbor", magic.CBOR)
)
