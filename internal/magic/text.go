package magic

import (
	"bytes"
	"time"

	"github.com/gabriel-vasile/mimetype/internal/charset"
	"github.com/gabriel-vasile/mimetype/internal/json"
)

// HTML matches a Hypertext Markup Language file.
func HTML(raw []byte, _ uint32) bool {
	return markup(raw,
		[]byte("<!DOCTYPE HTML"),
		[]byte("<HTML"),
		[]byte("<HEAD"),
		[]byte("<SCRIPT"),
		[]byte("<IFRAME"),
		[]byte("<H1"),
		[]byte("<DIV"),
		[]byte("<FONT"),
		[]byte("<TABLE"),
		[]byte("<A"),
		[]byte("<STYLE"),
		[]byte("<TITLE"),
		[]byte("<B"),
		[]byte("<BODY"),
		[]byte("<BR"),
		[]byte("<P"),
	)
}

// XML matches an Extensible Markup Language file.
func XML(raw []byte, _ uint32) bool {
	return markup(raw, []byte("<?XML"))
}

// Owl2 matches an Owl ontology file.
func Owl2(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<Ontology"), []byte(`xmlns="http://www.w3.org/2002/07/owl#"`)},
	)
}

// Rss matches a Rich Site Summary file.
func Rss(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<rss"), []byte{}},
	)
}

// Atom matches an Atom Syndication Format file.
func Atom(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<feed"), []byte(`xmlns="http://www.w3.org/2005/Atom"`)},
	)
}

// Kml matches a Keyhole Markup Language file.
func Kml(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<kml"), []byte(`xmlns="http://www.opengis.net/kml/2.2"`)},
		xmlSig{[]byte("<kml"), []byte(`xmlns="http://earth.google.com/kml/2.0"`)},
		xmlSig{[]byte("<kml"), []byte(`xmlns="http://earth.google.com/kml/2.1"`)},
		xmlSig{[]byte("<kml"), []byte(`xmlns="http://earth.google.com/kml/2.2"`)},
	)
}

// Xliff matches a XML Localization Interchange File Format file.
func Xliff(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<xliff"), []byte(`xmlns="urn:oasis:names:tc:xliff:document:1.2"`)},
	)
}

// Collada matches a COLLAborative Design Activity file.
func Collada(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<COLLADA"), []byte(`xmlns="http://www.collada.org/2005/11/COLLADASchema"`)},
	)
}

// Gml matches a Geography Markup Language file.
func Gml(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte{}, []byte(`xmlns:gml="http://www.opengis.net/gml"`)},
		xmlSig{[]byte{}, []byte(`xmlns:gml="http://www.opengis.net/gml/3.2"`)},
		xmlSig{[]byte{}, []byte(`xmlns:gml="http://www.opengis.net/gml/3.3/exr"`)},
	)
}

// Gpx matches a GPS Exchange Format file.
func Gpx(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<gpx"), []byte(`xmlns="http://www.topografix.com/GPX/1/1"`)},
	)
}

// Tcx matches a Training Center XML file.
func Tcx(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<TrainingCenterDatabase"), []byte(`xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2"`)},
	)
}

// X3d matches an Extensible 3D Graphics file.
func X3d(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<X3D"), []byte(`xmlns:xsd="http://www.w3.org/2001/XMLSchema-instance"`)},
	)
}

// Amf matches an Additive Manufacturing XML file.
func Amf(raw []byte, _ uint32) bool {
	return xml(raw, xmlSig{[]byte("<amf"), []byte{}})
}

// Threemf matches a 3D Manufacturing Format file.
func Threemf(raw []byte, _ uint32) bool {
	return xml(raw,
		xmlSig{[]byte("<model"), []byte(`xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02"`)},
	)
}

// Xfdf matches a XML Forms Data Format file.
func Xfdf(raw []byte, _ uint32) bool {
	return xml(raw, xmlSig{[]byte("<xfdf"), []byte(`xmlns="http://ns.adobe.com/xfdf/"`)})
}

// VCard matches a Virtual Contact File.
func VCard(raw []byte, _ uint32) bool {
	return ciPrefix(raw, []byte("BEGIN:VCARD\n"), []byte("BEGIN:VCARD\r\n"))
}

// ICalendar matches a iCalendar file.
func ICalendar(raw []byte, _ uint32) bool {
	return ciPrefix(raw, []byte("BEGIN:VCALENDAR\n"), []byte("BEGIN:VCALENDAR\r\n"))
}
func phpPageF(raw []byte, _ uint32) bool {
	return ciPrefix(raw,
		[]byte("<?PHP"),
		[]byte("<?\n"),
		[]byte("<?\r"),
		[]byte("<? "),
	)
}
func phpScriptF(raw []byte, _ uint32) bool {
	return shebang(raw,
		[]byte("/usr/local/bin/php"),
		[]byte("/usr/bin/php"),
		[]byte("/usr/bin/env php"),
	)
}

// Js matches a Javascript file.
func Js(raw []byte, _ uint32) bool {
	return shebang(raw,
		[]byte("/bin/node"),
		[]byte("/usr/bin/node"),
		[]byte("/bin/nodejs"),
		[]byte("/usr/bin/nodejs"),
		[]byte("/usr/bin/env node"),
		[]byte("/usr/bin/env nodejs"),
	)
}

// Lua matches a Lua programming language file.
func Lua(raw []byte, _ uint32) bool {
	return shebang(raw,
		[]byte("/usr/bin/lua"),
		[]byte("/usr/local/bin/lua"),
		[]byte("/usr/bin/env lua"),
	)
}

// Perl matches a Perl programming language file.
func Perl(raw []byte, _ uint32) bool {
	return shebang(raw,
		[]byte("/usr/bin/perl"),
		[]byte("/usr/bin/env perl"),
	)
}

// Python matches a Python programming language file.
func Python(raw []byte, _ uint32) bool {
	return shebang(raw,
		[]byte("/usr/bin/python"),
		[]byte("/usr/local/bin/python"),
		[]byte("/usr/bin/env python"),
	)
}

// Tcl matches a Tcl programming language file.
func Tcl(raw []byte, _ uint32) bool {
	return shebang(raw,
		[]byte("/usr/bin/tcl"),
		[]byte("/usr/local/bin/tcl"),
		[]byte("/usr/bin/env tcl"),
		[]byte("/usr/bin/tclsh"),
		[]byte("/usr/local/bin/tclsh"),
		[]byte("/usr/bin/env tclsh"),
		[]byte("/usr/bin/wish"),
		[]byte("/usr/local/bin/wish"),
		[]byte("/usr/bin/env wish"),
	)
}

// Rtf matches a Rich Text Format file.
func Rtf(raw []byte, _ uint32) bool {
	return bytes.HasPrefix(raw, []byte("{\\rtf"))
}

// Text matches a plain text file.
//
// TODO: This function does not parse BOM-less UTF16 and UTF32 files. Not really
// sure it should. Linux file utility also requires a BOM for UTF16 and UTF32.
func Text(raw []byte, limit uint32) bool {
	// First look for BOM.
	if cset := charset.FromBOM(raw); cset != "" {
		return true
	}
	// Binary data bytes as defined here: https://mimesniff.spec.whatwg.org/#binary-data-byte
	for _, b := range raw {
		if b <= 0x08 ||
			b == 0x0B ||
			0x0E <= b && b <= 0x1A ||
			0x1C <= b && b <= 0x1F {
			return false
		}
	}
	return true
}

// JSON matches a JavaScript Object Notation file.
func JSON(raw []byte, limit uint32) bool {
	if !json.LooksLikeObjectOrArray(raw) {
		return false
	}
	lraw := len(raw)
	parsed, inspected, firstToken, _ := json.Parse(json.ParserJSON, raw)
	// #175 A single JSON string, number or bool is not considered JSON.
	// JSON objects and arrays are reported as JSON.
	if firstToken != json.TokArray && firstToken != json.TokObject {
		return false
	}

	// If the full file content was provided, check that the whole input was parsed.
	if limit == 0 || lraw < int(limit) {
		return parsed == len(raw)
	}

	// If a section of the file was provided, check if all of it was inspected.
	// In other words, check that if there was a problem parsing, that problem
	// occured at the last byte in the input.
	return inspected == len(raw) && len(raw) > 0
}

// Php matches a PHP: Hypertext Preprocessor file.
func Php(raw []byte, limit uint32) bool {
	if res := phpPageF(raw, limit); res {
		return res
	}
	return phpScriptF(raw, limit)
}

// GeoJSON matches a RFC 7946 GeoJSON file.
//
// GeoJSON detection implies searching for key:value pairs like: `"type": "Feature"`
// in the input.
func GeoJSON(raw []byte, limit uint32) bool {
	if !json.LooksLikeObjectOrArray(raw) {
		return false
	}
	lraw := len(raw)
	parsed, inspected, firstToken, querySatisfied := json.Parse(json.ParserGeoJSON, raw)
	if !querySatisfied || firstToken != json.TokObject {
		return false
	}
	// If the full file content was provided, check that the whole input was parsed.
	if limit == 0 || lraw < int(limit) {
		return parsed == len(raw)
	}

	// If a section of the file was provided, check if all of it was inspected.
	// In other words, check that if there was a problem parsing, that problem
	// occured at the last byte in the input.
	return inspected == lraw && lraw > 0
}

// NdJSON matches a Newline delimited JSON file. All complete lines from raw
// must be valid JSON documents meaning they contain one of the valid JSON data
// types.
func NdJSON(raw []byte, limit uint32) bool {
	lCount, objOrArr := 0, 0
	raw = dropLastLine(raw, limit)
	var l []byte
	for len(raw) != 0 {
		l, raw = scanLine(raw)
		_, inspected, firstToken, _ := json.Parse(json.ParserJSON, l)
		if len(l) != inspected {
			return false
		}
		if firstToken == json.TokArray || firstToken == json.TokObject {
			objOrArr++
		}
		lCount++
	}

	return lCount > 1 && objOrArr > 0
}

// HAR matches a HAR Spec file.
// Spec: http://www.softwareishard.com/blog/har-12-spec/
func HAR(raw []byte, limit uint32) bool {
	if !json.LooksLikeObjectOrArray(raw) {
		return false
	}
	lraw := len(raw)
	parsed, inspected, firstToken, querySatisfied := json.Parse(json.ParserHARJSON, raw)
	if !querySatisfied || firstToken != json.TokObject {
		return false
	}
	// If the full file content was provided, check that the whole input was parsed.
	if limit == 0 || lraw < int(limit) {
		return parsed == len(raw)
	}

	// If a section of the file was provided, check if all of it was inspected.
	// In other words, check that if there was a problem parsing, that problem
	// occured at the last byte in the input.
	return inspected == lraw && lraw > 0
}

// Svg matches a SVG file.
func Svg(raw []byte, limit uint32) bool {
	return bytes.Contains(raw, []byte("<svg"))
}

// Srt matches a SubRip file.
func Srt(raw []byte, _ uint32) bool {
	line, raw := scanLine(raw)

	// First line must be 1.
	if len(line) != 1 || line[0] != '1' {
		return false
	}
	line, raw = scanLine(raw)
	// Timestamp format (e.g: 00:02:16,612 --> 00:02:19,376) limits second line
	// length to exactly 29 characters.
	if len(line) != 29 {
		return false
	}
	// Decimal separator of fractional seconds in the timestamps must be a
	// comma, not a period.
	if bytes.IndexByte(line, '.') != -1 {
		return false
	}
	sep := []byte(" --> ")
	i := bytes.Index(line, sep)
	if i == -1 {
		return false
	}
	const layout = "15:04:05,000"
	t0, err := time.Parse(layout, string(line[:i]))
	if err != nil {
		return false
	}
	t1, err := time.Parse(layout, string(line[i+len(sep):]))
	if err != nil {
		return false
	}
	if t0.After(t1) {
		return false
	}

	line, _ = scanLine(raw)
	// A third line must exist and not be empty. This is the actual subtitle text.
	return len(line) != 0
}

// Vtt matches a Web Video Text Tracks (WebVTT) file. See
// https://www.iana.org/assignments/media-types/text/vtt.
func Vtt(raw []byte, limit uint32) bool {
	// Prefix match.
	prefixes := [][]byte{
		{0xEF, 0xBB, 0xBF, 0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x0A}, // UTF-8 BOM, "WEBVTT" and a line feed
		{0xEF, 0xBB, 0xBF, 0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x0D}, // UTF-8 BOM, "WEBVTT" and a carriage return
		{0xEF, 0xBB, 0xBF, 0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x20}, // UTF-8 BOM, "WEBVTT" and a space
		{0xEF, 0xBB, 0xBF, 0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x09}, // UTF-8 BOM, "WEBVTT" and a horizontal tab
		{0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x0A},                   // "WEBVTT" and a line feed
		{0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x0D},                   // "WEBVTT" and a carriage return
		{0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x20},                   // "WEBVTT" and a space
		{0x57, 0x45, 0x42, 0x56, 0x54, 0x54, 0x09},                   // "WEBVTT" and a horizontal tab
	}
	for _, p := range prefixes {
		if bytes.HasPrefix(raw, p) {
			return true
		}
	}

	// Exact match.
	return bytes.Equal(raw, []byte{0xEF, 0xBB, 0xBF, 0x57, 0x45, 0x42, 0x56, 0x54, 0x54}) || // UTF-8 BOM and "WEBVTT"
		bytes.Equal(raw, []byte{0x57, 0x45, 0x42, 0x56, 0x54, 0x54}) // "WEBVTT"
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
func scanLine(b []byte) (line, remainder []byte) {
	line, remainder, _ = bytes.Cut(b, []byte("\n"))
	return dropCR(line), remainder
}
