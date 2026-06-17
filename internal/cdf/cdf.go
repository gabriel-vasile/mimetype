// Package cdf implements parsing of CDF (OLE2) files. It is greatly inspired
// by src/readcdf.c from libmagic. One difference is this implementation is
// permissive of truncated inputs. See readLimit in mimetype.go for the
// reason why truncated inputs need to be handled.
package cdf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// ErrNotCDF is returned when the input is not a CDF (OLE2) file.
var ErrNotCDF = errors.New("cdf: not a CDF (OLE2) file")

type CDFType int8

const (
	CDFTypeGeneric CDFType = iota
	CDFTypeInstaller
	CDFTypeDoc
	CDFTypePpt
	CDFTypeXls
	CDFTypeMsg
)

// Detect parses raw as a CDF (OLE2) compound file and returns the document type
// it contains. It returns CDFTypeGeneric for input that is not a CDF file or
// whose type cannot be narrowed down.
func Detect(raw []byte) CDFType {
	if len(raw) < 512 {
		return CDFTypeGeneric
	}
	c, err := parse(raw)
	if err != nil {
		return CDFTypeGeneric
	}
	return c.detect()
}

func (c *cdf) detect() CDFType {
	for _, name := range []string{"\x05SummaryInformation", "\x05DocumentSummaryInformation"} {
		if t, ok := c.detectFromSummary(name); ok {
			return t
		}
	}
	for i := range c.dir {
		d := &c.dir[i]
		if t, ok := lookupSection(d.nameBytes(), d.typ); ok {
			return t
		}
	}
	return CDFTypeGeneric
}

// detectFromSummary inspects a (Doc)SummaryInformation stream and tries to
// derive a CDFType from the root-storage CLSID, the property NameOfApplication,
// and finally the names of sibling user streams.
func (c *cdf) detectFromSummary(streamName string) (CDFType, bool) {
	if c.rootStorageUUID != nil && bytes.Equal(c.rootStorageUUID, msiCLSID) {
		return CDFTypeInstaller, true
	}
	raw, ok := c.userStream(streamName)
	if !ok {
		return CDFTypeGeneric, false
	}
	if app := summaryAppName(raw); len(app) > 0 {
		if t, ok := lookupSubstring(app, app2type); ok {
			return t, true
		}
	}
	for i := range c.dir {
		d := &c.dir[i]
		if d.nameLen == 0 {
			continue
		}
		if t, ok := lookupSubstring(d.nameBytes(), name2type); ok {
			return t, true
		}
	}
	return CDFTypeGeneric, true
}

const (
	cdfMagic uint64 = 0xE11AB1A1E011CFD0

	dirTypeUserStorage = 1
	dirTypeUserStream  = 2
	dirTypeRootStorage = 5

	dirEntrySize  = 128
	masterSATSize = 109 // first 109 SAT secids live in the file header
)

// dirEntry is a single CDF directory record. The UTF-16LE name is pre-decoded
// into an inline ASCII buffer at parse time, avoiding a per-entry heap
// allocation while keeping comparisons trivial. CDF names are at most 32
// UTF-16 code units, so 32 bytes always suffice.
type dirEntry struct {
	name        [32]byte
	nameLen     uint8
	typ         uint8
	streamFirst int32
	size        uint32
	storageUUID []byte
}

// nameBytes returns the decoded ASCII name without copying.
func (d *dirEntry) nameBytes() []byte { return d.name[:d.nameLen] }

// cdf holds everything we need from a CDF file to do detection.
type cdf struct {
	data            []byte
	secSize         int
	shortSecSize    int
	minStdStream    uint32
	satBytes        []byte // sector allocation table, kept as raw little-endian int32s
	ssat            []int32
	dir             []dirEntry
	sst             []byte // short-stream pool (root storage's stream)
	sstBuilt        bool   // whether sst was already loaded (it is loaded lazily)
	rootStreamFirst int32  // first sector of the root storage short-stream pool
	rootStreamSize  uint32 // size of the root storage short-stream pool
	rootStorageUUID []byte
}

// shortStream returns the root storage short-stream pool, loading it on first
// use. Detection often finishes (e.g. via the root CLSID or a long-stream
// summary) without ever reading a short stream, so building this eagerly would
// be wasted work.
func (c *cdf) shortStream() []byte {
	if !c.sstBuilt {
		c.sstBuilt = true
		if c.rootStreamFirst >= 0 {
			if sst, err := c.readLong(c.rootStreamFirst, c.rootStreamSize); err == nil {
				c.sst = sst
			}
		}
	}
	return c.sst
}

// parse reads the entire on-disk structure required for type detection.
func parse(raw []byte) (*cdf, error) {
	if len(raw) < 512 || binary.LittleEndian.Uint64(raw) != cdfMagic {
		return nil, fmt.Errorf("len(raw)=%d %w", len(raw), ErrNotCDF)
	}
	secP2 := binary.LittleEndian.Uint16(raw[30:32])
	shortP2 := binary.LittleEndian.Uint16(raw[32:34])
	if secP2 > 20 || shortP2 > 20 {
		return nil, fmt.Errorf("secP2=%d shortP2=%d, %w", secP2, shortP2, ErrNotCDF)
	}
	c := &cdf{
		data:         raw,
		secSize:      1 << secP2,
		shortSecSize: 1 << shortP2,
		minStdStream: binary.LittleEndian.Uint32(raw[56:60]),
	}
	if c.secSize < dirEntrySize {
		return nil, errors.New("cdf: sector smaller than directory entry")
	}
	firstDirSec := readSecID(raw[48:52])
	firstSSAT := readSecID(raw[60:64])
	firstMSAT := readSecID(raw[68:72])
	nMSAT := binary.LittleEndian.Uint32(raw[72:76])
	masterSAT := readInt32s(raw[76 : 76+4*masterSATSize])

	if err := c.buildSAT(masterSAT, firstMSAT, nMSAT); err != nil {
		return nil, err
	}
	if firstSSAT >= 0 {
		ssat, err := c.collectIDs(firstSSAT)
		if err != nil {
			return nil, err
		}
		c.ssat = ssat
	}
	dirBytes, err := c.readLong(firstDirSec, 0)
	if err != nil {
		return nil, err
	}
	c.dir = parseDir(dirBytes)

	c.rootStreamFirst = -1
	for _, d := range c.dir {
		if d.typ != dirTypeRootStorage || d.streamFirst < 0 {
			continue
		}
		c.rootStorageUUID = d.storageUUID
		// Record where the short-stream pool lives; it is loaded lazily by
		// shortStream the first time a short stream is actually read.
		c.rootStreamFirst = d.streamFirst
		c.rootStreamSize = d.size
		break
	}
	return c, nil
}

// readInt32s decodes a buffer as little-endian int32s.
func readInt32s(b []byte) []int32 {
	out := make([]int32, len(b)/4)
	for i := range out {
		out[i] = int32(binary.LittleEndian.Uint32(b[4*i:])) //nolint:gosec // intentional two's-complement reinterpretation of a sector id
	}
	return out
}

// appendInt32s decodes b as little-endian int32s and appends them to dst,
// avoiding the intermediate slice that readInt32s would allocate per sector.
func appendInt32s(dst []int32, b []byte) []int32 {
	for i := 0; i+4 <= len(b); i += 4 {
		dst = append(dst, int32(binary.LittleEndian.Uint32(b[i:]))) //nolint:gosec // intentional two's-complement reinterpretation of a sector id
	}
	return dst
}

// readSecID reinterprets four little-endian bytes as a signed sector id.
// Every 32-bit pattern is a valid id (values >= 0 are sector numbers,
// negatives are CDF sentinels such as -2 end-of-chain), so the conversion is
// an intentional two's-complement reinterpretation rather than an overflow.
func readSecID(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b)) //nolint:gosec // intentional two's-complement reinterpretation
}

// satLen is the number of sector ids in the SAT.
func (c *cdf) satLen() int { return len(c.satBytes) / 4 }

// satAt returns the i-th sector id from the SAT. Callers must ensure
// i < satLen().
func (c *cdf) satAt(i int32) int32 { return readSecID(c.satBytes[4*i:]) }

// errTruncated is returned when a sector starts past the end of the input.
// Callers treat this as a graceful stop rather than a hard failure so that
// detection can still succeed from whatever data was already collected.
var errTruncated = errors.New("cdf: input truncated and sector is past EOF")

// sector returns the bytes of long sector secid. If the file is truncated
// inside the requested sector the result is the available bytes (no padding).
// If the sector starts past EOF, errTruncated is returned.
func (c *cdf) sector(secid int32) ([]byte, error) {
	if secid < 0 {
		return nil, fmt.Errorf("cdf: negative secid %d", secid)
	}
	off := int64(c.secSize) * (1 + int64(secid))
	if off >= int64(len(c.data)) {
		return nil, errTruncated
	}
	end := off + int64(c.secSize)
	return c.data[off:min(end, int64(len(c.data)))], nil
}

func (c *cdf) sectorIDs(secid int32) ([]int32, error) {
	buf, err := c.sector(secid)
	if err != nil {
		return nil, err
	}
	return readInt32s(buf), nil
}

// buildSAT assembles the sector allocation table from the 109 entries in the
// header plus any extension blocks chained via the master SAT.
// If the input is truncated, the SAT is built from whatever sectors are
// available and no error is returned.
func (c *cdf) buildSAT(masterSAT []int32, firstMSAT int32, nMSAT uint32) error {
	// Common case: the whole SAT lives in a single sector referenced by the
	// first master-SAT entry, with no extension blocks. Point straight at that
	// sector's bytes inside the input, avoiding any allocation.
	if firstMSAT < 0 && len(masterSAT) > 0 && masterSAT[0] >= 0 &&
		(len(masterSAT) == 1 || masterSAT[1] < 0) {
		if buf, err := c.sector(masterSAT[0]); err == nil {
			c.satBytes = buf
			return nil
		}
	}

	// The SAT has exactly one entry per long sector, so its byte length can
	// never exceed len(data). Capping at that bound keeps malformed inputs with
	// cyclic master-SAT chains from amplifying into unbounded allocations.
	maxBytes := len(c.data)
	// A full SAT sector holds secSize bytes; preallocating that avoids a regrow
	// in the common single-sector case.
	sat := make([]byte, 0, c.secSize)
	for _, sec := range masterSAT {
		if sec < 0 {
			c.satBytes = sat
			return nil
		}
		buf, err := c.sector(sec)
		if errors.Is(err, errTruncated) {
			break // file ends before this SAT sector; use what we have
		}
		if err != nil {
			return fmt.Errorf("cdf: SAT: %w", err)
		}
		sat = append(sat, buf...)
		if len(sat) > maxBytes {
			c.satBytes = sat
			return nil
		}
	}
	perSec := c.secSize/4 - 1
	mid := firstMSAT
	for j := uint32(0); j < nMSAT && mid >= 0; j++ {
		msa, err := c.sectorIDs(mid)
		if errors.Is(err, errTruncated) {
			break
		}
		if err != nil {
			return fmt.Errorf("cdf: master SAT: %w", err)
		}
		for k := 0; k < perSec; k++ {
			if k >= len(msa) {
				c.satBytes = sat
				return nil // master SAT sector truncated; use what we have
			}
			if msa[k] < 0 {
				c.satBytes = sat
				return nil
			}
			ids, err := c.sector(msa[k])
			if errors.Is(err, errTruncated) {
				c.satBytes = sat
				return nil
			}
			if err != nil {
				return fmt.Errorf("cdf: SAT: %w", err)
			}
			sat = append(sat, ids...)
			if len(sat) > maxBytes {
				c.satBytes = sat
				return nil
			}
		}
		if perSec >= len(msa) {
			c.satBytes = sat
			return nil // no next-MSAT pointer available
		}
		mid = msa[perSec]
	}
	c.satBytes = sat
	return nil
}

// collectIDs walks the SAT chain at sid and returns every sector decoded as
// int32s. Used to build the SSAT. On truncation it returns whatever was
// collected rather than an error.
func (c *cdf) collectIDs(sid int32) ([]int32, error) {
	maxIDs := len(c.data) / 4
	out := make([]int32, 0, c.secSize/4)
	for sid >= 0 {
		if int(sid) >= c.satLen() {
			break // SAT is truncated; stop collecting
		}
		buf, err := c.sector(sid)
		if errors.Is(err, errTruncated) {
			break
		}
		if err != nil {
			return nil, err
		}
		out = appendInt32s(out, buf)
		if len(out) > maxIDs {
			break // cyclic SAT chain; stop allocating
		}
		sid = c.satAt(sid)
	}
	return out, nil
}

// readLong reads a long-sector chain starting at sid. If length > 0 the
// result is truncated to that many bytes. On truncation it returns whatever
// sectors were readable rather than an error.
func (c *cdf) readLong(sid int32, length uint32) ([]byte, error) {
	// Fast path: when the chain is a single physically contiguous run of
	// sectors (the common case for the directory and summary streams) the data
	// is already laid out sequentially in the input, so return a sub-slice of
	// it instead of allocating a buffer and copying every sector.
	if sid >= 0 {
		maxSec := len(c.data)/c.secSize + 1
		n, s := 0, sid
		contiguous := true
		for s >= 0 {
			if int(s) >= c.satLen() {
				break // SAT truncated; what remains is still contiguous
			}
			n++
			if n > maxSec {
				contiguous = false // cyclic chain; let the slow path guard it
				break
			}
			next := c.satAt(s)
			if next >= 0 && next != s+1 {
				contiguous = false
				break
			}
			s = next
		}
		if contiguous {
			off := c.secSize * (1 + int(sid))
			if off >= len(c.data) {
				return nil, nil
			}
			end := off + n*c.secSize
			if end > len(c.data) {
				end = len(c.data)
			}
			out := c.data[off:end]
			if length > 0 && int(length) < len(out) {
				out = out[:length]
			}
			return out, nil
		}
	}

	// Slow path: gather a fragmented chain into a fresh buffer.
	// TODO: anyway to avoid allocating and copying the bytes?
	out := make([]byte, 0, c.secSize)
	for sid >= 0 {
		if int(sid) >= c.satLen() {
			break // SAT truncated; return what we have
		}
		buf, err := c.sector(sid)
		if errors.Is(err, errTruncated) {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, buf...)
		if len(out) >= len(c.data) {
			break // chain longer than the file: cyclic SAT, stop
		}
		sid = c.satAt(sid)
	}
	if length > 0 && int(length) < len(out) {
		out = out[:length]
	}
	return out, nil
}

// readShort reads a short-sector chain at sid by indexing into the short-stream
// pool. On truncation it returns whatever short sectors were readable.
func (c *cdf) readShort(sid int32, length uint32) ([]byte, error) {
	sst := c.shortStream()
	if sst == nil {
		return nil, errors.New("cdf: short stream not loaded")
	}
	// TODO: anyway to avoid allocating and copying the bytes?
	out := make([]byte, 0, c.shortSecSize)
	for sid >= 0 {
		if int(sid) >= len(c.ssat) {
			break // SSAT truncated; return what we have
		}
		off := int(sid) * c.shortSecSize
		if off+c.shortSecSize > len(sst) {
			break // short-stream pool truncated
		}
		out = append(out, sst[off:off+c.shortSecSize]...)
		if len(out) >= len(sst) {
			break // chain longer than the pool: cyclic SSAT, stop
		}
		sid = c.ssat[sid]
	}
	if length > 0 && int(length) < len(out) {
		out = out[:length]
	}
	return out, nil
}

// readChain dispatches to the long or short reader depending on stream size.
func (c *cdf) readChain(sid int32, length uint32) ([]byte, error) {
	if length < c.minStdStream && c.rootStreamFirst >= 0 {
		return c.readShort(sid, length)
	}
	return c.readLong(sid, length)
}

// parseDir splits the directory stream into 128-byte entries, decoding each
// UTF-16LE name into the entry's inline ASCII buffer up to the first NUL.
func parseDir(b []byte) []dirEntry {
	n := len(b) / dirEntrySize
	out := make([]dirEntry, n)
	for i := 0; i < n; i++ {
		raw := b[i*dirEntrySize:]
		nameLen := int(binary.LittleEndian.Uint16(raw[64:]))
		if nameLen > 64 {
			nameLen = 64
		}
		d := &out[i]
		k := uint8(0)
		for j := 0; j < nameLen/2; j++ {
			// Names are ASCII; keep the low byte of each little-endian UTF-16
			// code unit and stop at the first NUL.
			lo, hi := raw[2*j], raw[2*j+1]
			if lo == 0 && hi == 0 {
				break
			}
			d.name[k] = lo
			k++
		}
		d.nameLen = k
		d.typ = raw[66]
		d.streamFirst = readSecID(raw[116:120])
		d.size = binary.LittleEndian.Uint32(raw[120:])
		d.storageUUID = raw[80:96]
	}
	return out
}

// userStream finds a user stream by name and returns its bytes.
func (c *cdf) userStream(name string) ([]byte, bool) {
	for i := range c.dir {
		d := &c.dir[i]
		if d.typ == dirTypeUserStream && string(d.nameBytes()) == name {
			buf, err := c.readChain(d.streamFirst, d.size)
			if err != nil {
				return nil, false
			}
			return buf, true
		}
	}
	return nil, false
}

const (
	propIDNameOfApplication = 0x12

	typeMask        = 0x0fff
	typeVector      = 0x1000
	typeStringASCII = 0x1e
	typeStringWide  = 0x1f

	sectionDeclOffset = 0x1c // section declaration in property-set header
)

// summaryAppName parses a (Doc)SummaryInformation stream and returns the
// value of property NameOfApplication (0x12) as printable ASCII, or nil if
// not present or the stream is malformed. This is the only summary property
// the detection logic ever consults.
func summaryAppName(stream []byte) []byte {
	if len(stream) < sectionDeclOffset+20 {
		return nil
	}
	sdOff := binary.LittleEndian.Uint32(stream[sectionDeclOffset+16:])
	if uint64(sdOff)+8 > uint64(len(stream)) {
		return nil
	}
	section := stream[sdOff:]
	shLen := binary.LittleEndian.Uint32(section[0:])
	nProps := binary.LittleEndian.Uint32(section[4:])
	if uint64(shLen) > uint64(len(section)) || nProps > 1<<16 || 8+8*nProps > shLen {
		return nil
	}
	for i := uint32(0); i < nProps; i++ {
		base := 8 + 8*i
		id := binary.LittleEndian.Uint32(section[base:])
		if id != propIDNameOfApplication {
			continue
		}
		off := binary.LittleEndian.Uint32(section[base+4:])
		if uint64(off)+8 > uint64(shLen) {
			return nil
		}
		typ := binary.LittleEndian.Uint32(section[off:])
		if typ&typeVector != 0 {
			return nil
		}
		step := uint32(0)
		switch typ & typeMask {
		case typeStringASCII:
			step = 1
		case typeStringWide:
			step = 2
		default:
			return nil
		}
		slen := binary.LittleEndian.Uint32(section[off+4:])
		start := uint64(off) + 8
		end := start + uint64(slen)*uint64(step)
		if end > uint64(shLen) {
			return nil
		}
		return printableLowBytes(section[start:end], int(step))
	}
	return nil
}

// printableLowBytes copies the printable low byte of each step-byte unit
// in b, stopping at the first NUL.
func printableLowBytes(b []byte, step int) []byte {
	out := make([]byte, 0, len(b)/step)
	for i := 0; i+step <= len(b); i += step {
		c := b[i]
		if c == 0 {
			break
		}
		if c >= 0x20 && c < 0x7f {
			out = append(out, c)
		}
	}
	return out
}

// pattern is a case-insensitive substring → CDFType mapping. Entries are
// tested in order; first match wins. needle is stored upper-cased so it can be
// matched case-insensitively by scan.Bytes.Search with scan.IgnoreCase.
type pattern struct {
	needle []byte
	typ    CDFType
}

// app2type maps NameOfApplication values to CDFTypes.
// Mirrors app2mime[] in libmagic. Needles are upper-cased for case-insensitive
// matching via scan.IgnoreCase.
var app2type = []pattern{
	{[]byte("WORD"), CDFTypeDoc},
	{[]byte("EXCEL"), CDFTypeXls},
	{[]byte("POWERPOINT"), CDFTypePpt},
	{[]byte("ADVANCED INSTALLER"), CDFTypeInstaller},
	{[]byte("INSTALLSHIELD"), CDFTypeInstaller},
	{[]byte("MICROSOFT PATCH COMPILER"), CDFTypeInstaller},
	{[]byte("NANT"), CDFTypeInstaller},
	{[]byte("WINDOWS INSTALLER"), CDFTypeInstaller},
}

// name2type maps directory entry names to CDFTypes.
// Mirrors name2mime[] in libmagic. Needles are upper-cased for case-insensitive
// matching via scan.IgnoreCase.
var name2type = []pattern{
	{[]byte("BOOK"), CDFTypeXls},
	{[]byte("WORKBOOK"), CDFTypeXls},
	{[]byte("WORDDOCUMENT"), CDFTypeDoc},
	{[]byte("POWERPOINT"), CDFTypePpt},
	{[]byte("DIGITALSIGNATURE"), CDFTypeInstaller},
}

// lookupSubstring returns the CDFType for the first entry in t whose needle
// is a case-insensitive substring of v. Mirrors C's strcasestr semantics
// under the C locale. It allocates nothing: scan.IgnoreCase matches the
// upper-cased needle against input of either case.
func lookupSubstring(v []byte, t []pattern) (CDFType, bool) {
	s := scan.Bytes(v)
	for _, p := range t {
		if i, _ := s.Search(p.needle, scan.IgnoreCase); i != -1 {
			return p.typ, true
		}
	}
	return CDFTypeGeneric, false
}

// msiCLSID is the Microsoft Installer root-storage CLSID, in on-disk byte
// order (cdf_directory_t.d_storage_uuid stores two little-endian uint64s).
var msiCLSID = []byte{
	0x84, 0x10, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46,
}

// section is a (directory entry name, type) → CDFType mapping.
type section struct {
	name string
	typ  uint8
	cdf  CDFType
}

// sectionTypes maps distinctive directory entries to CDFTypes — a flattened
// equivalent of sectioninfo[] in libmagic. Used as a fallback when no
// SummaryInformation stream is present. A slice (rather than a map) lets
// lookupSection compare entry names without allocating a string key.
var sectionTypes = []section{
	// libmagic uses application/encrypted, but that is not a registered media type.
	// For now, we skip identifying that and fall-back on CDFTypeGeneric
	// {"EncryptedPackage", dirTypeUserStream, CDFTypeEncrypted},
	// {"EncryptedSummary", dirTypeUserStream, CDFTypeEncrypted},
	{"Book", dirTypeUserStream, CDFTypeXls},
	{"Workbook", dirTypeUserStream, CDFTypeXls},
	{"WordDocument", dirTypeUserStream, CDFTypeDoc},
	{"PowerPoint Document", dirTypeUserStream, CDFTypePpt},
	{"__properties_version1.0", dirTypeUserStream, CDFTypeMsg},
	{"__recip_version1.0_#00000000", dirTypeUserStorage, CDFTypeMsg},
}

// lookupSection returns the CDFType for a directory entry whose name and type
// match a sectionTypes entry exactly. The string(name) == comparison is
// optimized by the compiler to avoid allocating.
func lookupSection(name []byte, typ uint8) (CDFType, bool) {
	for _, s := range sectionTypes {
		if s.typ == typ && string(name) == s.name {
			return s.cdf, true
		}
	}
	return CDFTypeGeneric, false
}

func (c CDFType) String() string {
	switch c {
	case CDFTypeGeneric:
		return "application/x-ole-storage"
	case CDFTypeInstaller:
		return "application/vnd.ms-msi"
	case CDFTypeDoc:
		return "application/msword"
	case CDFTypePpt:
		return "application/vnd.ms-powerpoint"
	case CDFTypeXls:
		return "application/vnd.ms-excel"
	case CDFTypeMsg:
		return "application/vnd.ms-outlook"
	}

	return "unknown CDF"
}
