package json

import (
	"bytes"
	"sync"
)

// ParserType dictates what type of parser to use.
type ParserType string

const (
	ParserJSON    ParserType = "json"
	ParserGeoJSON ParserType = "geojson"
	ParserHARJSON ParserType = "harjson"
	maxRecursion             = 4096
)

// scannerPools keeps a pool of parsers for each JSON type.
// Parser queries can get big and the pools avoid allocating the same objects
// over and over.
var scannerPools = map[ParserType]*sync.Pool{
	ParserJSON: {New: func() any {
		return &parserState{
			maxRecursion: maxRecursion,
		}
	}},
	ParserGeoJSON: {New: func() any {
		return &parserState{
			maxRecursion: maxRecursion,
			queries: []query{{
				SearchPath: [][]byte{[]byte("type")},
				SearchVals: [][]byte{
					[]byte(`"Feature"`),
					[]byte(`"FeatureCollection"`),
					[]byte(`"Point"`),
					[]byte(`"LineString"`),
					[]byte(`"Polygon"`),
					[]byte(`"MultiPoint"`),
					[]byte(`"MultiLineString"`),
					[]byte(`"MultiPolygon"`),
					[]byte(`"GeometryCollection"`),
				},
			}},
		}
	}},
	ParserHARJSON: {New: func() any {
		return &parserState{
			maxRecursion: maxRecursion,
			queries: []query{{
				SearchPath: [][]byte{[]byte("log"), []byte("version")},
			}, {
				SearchPath: [][]byte{[]byte("log"), []byte("creator")},
			}, {
				SearchPath: [][]byte{[]byte("log"), []byte("entries")},
			}},
		}
	}},
}

// parserState holds the state of JSON parsing. The number of inspected bytes,
// the current path inside the JSON object, etc.
type parserState struct {
	// ib represents the number of inspected bytes.
	// Because mimetype limits itself to only reading the header of the file,
	// it means sometimes the input JSON can be truncated. In that case, we want
	// to still detect it as JSON, even if it's invalid/truncated.
	// When ib == len(input) it means the JSON was valid (at least the header).
	ib           int
	queries      []query
	maxRecursion int
	// currPath keeps a track of the JSON keys parsed up.
	// It works only for JSON objects. JSON arrays are ignored
	// mainly because the functionality is not needed.
	currPath [][]byte
	// firstToken stores the first JSON token encountered in input.
	// TODO: performance would be better if we would stop parsing as soon
	// as we see that first token is not what we are interested in.
	firstToken int
}

// query holds information about a combination of {"key": "val"} that we're trying
// to search for inside the JSON.
type query struct {
	// SearchPath represents the whole path to look for inside the JSON.
	// ex: [][]byte{[]byte("foo"), []byte("bar")} matches {"foo": {"bar": "baz"}}
	SearchPath [][]byte
	// SearchVals represents values to look for when the SearchPath is found.
	// Each SearchVal element is tried until one of them matches (logical OR.)
	SearchVals [][]byte

	searchPathSatisfied bool
	searchValSatisfied  bool
}

func (d *parserState) anyQuerySatisfied() bool {
	for i := range d.queries {
		if d.queries[i].searchPathSatisfied && d.queries[i].searchValSatisfied {
			return true
		}
	}
	return false
}
func (d *parserState) resetQueries() {
	for i := range d.queries {
		d.queries[i].searchPathSatisfied = false
		d.queries[i].searchValSatisfied = false
	}
}
func (d *parserState) resetQueryPaths() {
	for i := range d.queries {
		if !d.queries[i].searchValSatisfied {
			d.queries[i].searchPathSatisfied = false
		}
	}
}
func (d *parserState) markQueryPaths() {
	for i := range d.queries {
		if eq(d.queries[i].SearchPath, d.currPath) {
			d.queries[i].searchPathSatisfied = true
		}
	}
}
func (d *parserState) markQueryVals(key []byte) {
	for i := range d.queries {
		if !d.queries[i].searchPathSatisfied {
			continue
		}
		if len(d.queries[i].SearchVals) == 0 {
			d.queries[i].searchValSatisfied = true
		}
		for _, val := range d.queries[i].SearchVals {
			if bytes.Equal(val, key) {
				d.queries[i].searchValSatisfied = true
			}
		}
	}
}
func eq(path1, path2 [][]byte) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i := range path1 {
		if !bytes.Equal(path1[i], path2[i]) {
			return false
		}
	}
	return true
}

// LooksLikeObjectOrArray reports if first non white space character from raw
// is either { or [. Parsing raw as JSON is a heavy operation. When receiving some
// text input we can skip parsing if the input does not even look like JSON.
func LooksLikeObjectOrArray(raw []byte) bool {
	for i := range raw {
		if isSpace(raw[i]) {
			continue
		}
		return raw[i] == '{' || raw[i] == '['
	}

	return false
}

// Parse will take out a parser from the pool depending on typ and tries to parse
// raw bytes as JSON.
func Parse(typ ParserType, raw []byte) (parsed, inspected, firstToken int, querySatisfied bool) {
	d := scannerPools[typ].Get().(*parserState)
	defer func() {
		// Avoid hanging on to too much memory in extreme input cases.
		if len(d.currPath) > 128 {
			d.currPath = nil
		}
		scannerPools[typ].Put(d)
	}()
	d.reset()
	got := d.consumeAny(raw, 0)
	return got, d.ib, d.firstToken, d.anyQuerySatisfied()
}

func (d *parserState) reset() {
	d.currPath = d.currPath[0:0]
	d.ib = 0
	d.firstToken = TokInvalid
	d.resetQueries()
}

func (d *parserState) consumeSpace(b []byte) (n int) {
	for len(b) > 0 && isSpace(b[0]) {
		b = b[1:]
		n++
		d.ib++
	}
	return n
}

func (d *parserState) consumeConst(b, cnst []byte) int {
	lb := len(b)
	for i, c := range cnst {
		if lb > i && b[i] == c {
			d.ib++
		} else {
			return 0
		}
	}
	return len(cnst)
}

func (d *parserState) consumeString(b []byte) (n int) {
	var c byte
	for len(b[n:]) > 0 {
		c, n = b[n], n+1
		d.ib++
		switch c {
		case '\\':
			if len(b[n:]) == 0 {
				return 0
			}
			switch b[n] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				n++
				d.ib++
				continue
			case 'u':
				n++
				d.ib++
				for j := 0; j < 4 && len(b[n:]) > 0; j++ {
					if !isXDigit(b[n]) {
						return 0
					}
					n++
					d.ib++
				}
				continue
			default:
				return 0
			}
		case '"':
			return n
		default:
			continue
		}
	}
	return 0
}

func (d *parserState) consumeNumber(b []byte) (n int) {
	got := false
	var i int

	if len(b) == 0 {
		goto out
	}
	if b[0] == '-' {
		b, i = b[1:], i+1
		d.ib++
	}

	for len(b) > 0 {
		if !isDigit(b[0]) {
			break
		}
		got = true
		b, i = b[1:], i+1
		d.ib++
	}
	if len(b) == 0 {
		goto out
	}
	if b[0] == '.' {
		b, i = b[1:], i+1
		d.ib++
	}
	for len(b) > 0 {
		if !isDigit(b[0]) {
			break
		}
		got = true
		b, i = b[1:], i+1
		d.ib++
	}
	if len(b) == 0 {
		goto out
	}
	if got && (b[0] == 'e' || b[0] == 'E') {
		b, i = b[1:], i+1
		d.ib++
		got = false
		if len(b) == 0 {
			goto out
		}
		if b[0] == '+' || b[0] == '-' {
			b, i = b[1:], i+1
			d.ib++
		}
		for len(b) > 0 {
			if !isDigit(b[0]) {
				break
			}
			got = true
			b, i = b[1:], i+1
			d.ib++
		}
	}
out:
	if got {
		return i
	}
	return 0
}

func (d *parserState) consumeArray(b []byte, lvl int) (n int) {
	if len(b) == 0 {
		return 0
	}

	for n < len(b) {
		n += d.consumeSpace(b[n:])
		if len(b[n:]) == 0 {
			return 0
		}
		if b[n] == ']' {
			d.ib++
			return n + 1
		}
		innerParsed := d.consumeAny(b[n:], lvl)
		if innerParsed == 0 {
			return 0
		}
		n += innerParsed
		if len(b[n:]) == 0 {
			return 0
		}
		switch b[n] {
		case ',':
			n += 1
			d.ib++
			continue
		case ']':
			d.ib++
			return n + 1
		default:
			return 0
		}
	}
	return 0
}

func (d *parserState) consumeObject(b []byte, lvl int) (n int) {
	for n < len(b) {
		n += d.consumeSpace(b[n:])
		if len(b[n:]) == 0 {
			return 0
		}
		if b[n] == '}' {
			d.ib++
			return n + 1
		}
		if b[n] != '"' {
			return 0
		} else {
			n += 1
			d.ib++
		}
		if keyLen := d.consumeString(b[n:]); keyLen == 0 {
			return 0
		} else {
			d.currPath = append(d.currPath, b[n:n+keyLen-1])
			d.resetQueryPaths()
			d.markQueryPaths()
			n += keyLen
		}
		n += d.consumeSpace(b[n:])
		if len(b[n:]) == 0 {
			return 0
		}
		if b[n] != ':' {
			return 0
		} else {
			n += 1
			d.ib++
		}
		n += d.consumeSpace(b[n:])
		if len(b[n:]) == 0 {
			return 0
		}

		if valLen := d.consumeAny(b[n:], lvl); valLen == 0 {
			return 0
		} else {
			d.markQueryVals(b[n : n+valLen])
			n += valLen
		}
		if len(b[n:]) == 0 {
			return 0
		}
		switch b[n] {
		case ',':
			d.currPath = d.currPath[:len(d.currPath)-1]
			n++
			d.ib++
			continue
		case '}':
			d.currPath = d.currPath[:len(d.currPath)-1]
			d.ib++
			return n + 1
		default:
			return 0
		}
	}
	return 0
}

func (d *parserState) consumeAny(b []byte, lvl int) (n int) {
	// Avoid too much recursion.
	if d.maxRecursion != 0 && lvl > d.maxRecursion {
		return 0
	}
	n += d.consumeSpace(b)
	if len(b[n:]) == 0 {
		return 0
	}

	var t, rv int
	switch b[n] {
	case '"':
		n++
		d.ib++
		rv = d.consumeString(b[n:])
		t = TokString
	case '[':
		n++
		d.ib++
		rv = d.consumeArray(b[n:], lvl+1)
		t = TokArray
	case '{':
		n++
		d.ib++
		rv = d.consumeObject(b[n:], lvl+1)
		t = TokObject
	case 't':
		rv = d.consumeConst(b[n:], []byte("true"))
		t = TokTrue
	case 'f':
		rv = d.consumeConst(b[n:], []byte("false"))
		t = TokFalse
	case 'n':
		rv = d.consumeConst(b[n:], []byte("null"))
		t = TokNull
	default:
		rv = d.consumeNumber(b[n:])
		t = TokNumber
	}
	if lvl == 0 {
		d.firstToken = t
	}
	if rv <= 0 {
		return n
	}
	n += rv
	n += d.consumeSpace(b[n:])
	return n
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isXDigit(c byte) bool {
	if isDigit(c) {
		return true
	}
	return ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

const (
	TokInvalid = iota
	TokNull
	TokTrue
	TokFalse
	TokNumber
	TokString
	TokArray
	TokObject
	TokComma
)
