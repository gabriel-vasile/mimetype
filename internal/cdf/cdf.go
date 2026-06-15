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
	"strings"
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
	for _, d := range c.dir {
		if t, ok := sectionTypes[sectionKey{d.name, d.typ}]; ok {
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
	if app := summaryAppName(raw); app != "" {
		if t, ok := lookupSubstring(app, app2type); ok {
			return t, true
		}
	}
	for _, d := range c.dir {
		if d.name == "" {
			continue
		}
		if t, ok := lookupSubstring(d.name, name2type); ok {
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

// dirEntry is a single CDF directory record. We pre-decode the UTF-16LE name
// into a Go string at parse time so subsequent comparisons are trivial.
type dirEntry struct {
	name        string
	typ         uint8
	streamFirst int32
	size        uint32
	storageUUID []byte
}

// cdf holds everything we need from a CDF file to do detection.
type cdf struct {
	data            []byte
	secSize         int
	shortSecSize    int
	minStdStream    uint32
	sat             []int32
	ssat            []int32
	dir             []dirEntry
	sst             []byte // short-stream pool (root storage's stream)
	rootStorageUUID []byte
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

	for _, d := range c.dir {
		if d.typ != dirTypeRootStorage || d.streamFirst < 0 {
			continue
		}
		c.rootStorageUUID = d.storageUUID
		if sst, err := c.readLong(d.streamFirst, d.size); err == nil {
			c.sst = sst
		}
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
	// A valid SAT has exactly one entry per long sector, so it can never hold
	// more than len(data)/4 ids. Capping at that bound keeps malformed inputs
	// with cyclic master-SAT chains from amplifying into unbounded allocations.
	maxIDs := len(c.data) / 4
	// A full SAT sector holds secSize/4 ids; preallocating that avoids a regrow
	// in the common single-sector case.
	sat := make([]int32, 0, c.secSize/4)
	for _, sec := range masterSAT {
		if sec < 0 {
			c.sat = sat
			return nil
		}
		buf, err := c.sector(sec)
		if errors.Is(err, errTruncated) {
			break // file ends before this SAT sector; use what we have
		}
		if err != nil {
			return fmt.Errorf("cdf: SAT: %w", err)
		}
		sat = appendInt32s(sat, buf)
		if len(sat) > maxIDs {
			c.sat = sat
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
				c.sat = sat
				return nil // master SAT sector truncated; use what we have
			}
			if msa[k] < 0 {
				c.sat = sat
				return nil
			}
			ids, err := c.sector(msa[k])
			if errors.Is(err, errTruncated) {
				c.sat = sat
				return nil
			}
			if err != nil {
				return fmt.Errorf("cdf: SAT: %w", err)
			}
			sat = appendInt32s(sat, ids)
			if len(sat) > maxIDs {
				c.sat = sat
				return nil
			}
		}
		if perSec >= len(msa) {
			c.sat = sat
			return nil // no next-MSAT pointer available
		}
		mid = msa[perSec]
	}
	c.sat = sat
	return nil
}

// collectIDs walks the SAT chain at sid and returns every sector decoded as
// int32s. Used to build the SSAT. On truncation it returns whatever was
// collected rather than an error.
func (c *cdf) collectIDs(sid int32) ([]int32, error) {
	maxIDs := len(c.data) / 4
	out := make([]int32, 0, c.secSize/4)
	for sid >= 0 {
		if int(sid) >= len(c.sat) {
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
		sid = c.sat[sid]
	}
	return out, nil
}

// readLong reads a long-sector chain starting at sid. If length > 0 the
// result is truncated to that many bytes. On truncation it returns whatever
// sectors were readable rather than an error.
func (c *cdf) readLong(sid int32, length uint32) ([]byte, error) {
	// TODO: anyway to avoid allocating and copying the bytes?
	out := make([]byte, 0, c.secSize)
	for sid >= 0 {
		if int(sid) >= len(c.sat) {
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
		sid = c.sat[sid]
	}
	if length > 0 && int(length) < len(out) {
		out = out[:length]
	}
	return out, nil
}

// readShort reads a short-sector chain at sid by indexing into c.sst.
// On truncation it returns whatever short sectors were readable.
func (c *cdf) readShort(sid int32, length uint32) ([]byte, error) {
	if c.sst == nil {
		return nil, errors.New("cdf: short stream not loaded")
	}
	// TODO: anyway to avoid allocating and copying the bytes?
	out := make([]byte, 0, c.shortSecSize)
	for sid >= 0 {
		if int(sid) >= len(c.ssat) {
			break // SSAT truncated; return what we have
		}
		off := int(sid) * c.shortSecSize
		if off+c.shortSecSize > len(c.sst) {
			break // short-stream pool truncated
		}
		out = append(out, c.sst[off:off+c.shortSecSize]...)
		if len(out) >= len(c.sst) {
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
	if length < c.minStdStream && c.sst != nil {
		return c.readShort(sid, length)
	}
	return c.readLong(sid, length)
}

// parseDir splits the directory stream into 128-byte entries, decoding each
// UTF-16LE name into an ASCII Go string up to the first NUL.
func parseDir(b []byte) []dirEntry {
	n := len(b) / dirEntrySize
	out := make([]dirEntry, n)
	for i := 0; i < n; i++ {
		raw := b[i*dirEntrySize:]
		nameLen := int(binary.LittleEndian.Uint16(raw[64:]))
		if nameLen > 64 {
			nameLen = 64
		}
		nb := make([]byte, 0, 32) // names are at most 32 UTF-16 code units
		for j := 0; j < nameLen/2; j++ {
			// Names are ASCII; keep the low byte of each little-endian UTF-16
			// code unit and stop at the first NUL.
			lo, hi := raw[2*j], raw[2*j+1]
			if lo == 0 && hi == 0 {
				break
			}
			nb = append(nb, lo)
		}
		d := dirEntry{
			// TODO: avoid allocation and instead compare on read-only slices of data.
			name:        string(nb),
			typ:         raw[66],
			streamFirst: readSecID(raw[116:120]),
			size:        binary.LittleEndian.Uint32(raw[120:]),
		}
		d.storageUUID = raw[80:96]
		out[i] = d
	}
	return out
}

// userStream finds a user stream by name and returns its bytes.
func (c *cdf) userStream(name string) ([]byte, bool) {
	for _, d := range c.dir {
		if d.typ == dirTypeUserStream && d.name == name {
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
// value of property NameOfApplication (0x12) as printable ASCII, or "" if
// not present or the stream is malformed. This is the only summary property
// the detection logic ever consults.
func summaryAppName(stream []byte) string {
	if len(stream) < sectionDeclOffset+20 {
		return ""
	}
	sdOff := binary.LittleEndian.Uint32(stream[sectionDeclOffset+16:])
	if uint64(sdOff)+8 > uint64(len(stream)) {
		return ""
	}
	section := stream[sdOff:]
	shLen := binary.LittleEndian.Uint32(section[0:])
	nProps := binary.LittleEndian.Uint32(section[4:])
	if uint64(shLen) > uint64(len(section)) || nProps > 1<<16 || 8+8*nProps > shLen {
		return ""
	}
	for i := uint32(0); i < nProps; i++ {
		base := 8 + 8*i
		id := binary.LittleEndian.Uint32(section[base:])
		if id != propIDNameOfApplication {
			continue
		}
		off := binary.LittleEndian.Uint32(section[base+4:])
		if uint64(off)+8 > uint64(shLen) {
			return ""
		}
		typ := binary.LittleEndian.Uint32(section[off:])
		if typ&typeVector != 0 {
			return ""
		}
		step := uint32(0)
		switch typ & typeMask {
		case typeStringASCII:
			step = 1
		case typeStringWide:
			step = 2
		default:
			return ""
		}
		slen := binary.LittleEndian.Uint32(section[off+4:])
		start := uint64(off) + 8
		end := start + uint64(slen)*uint64(step)
		if end > uint64(shLen) {
			return ""
		}
		return printableLowBytes(section[start:end], int(step))
	}
	return ""
}

// printableLowBytes copies the printable low byte of each step-byte unit
// in b, stopping at the first NUL.
func printableLowBytes(b []byte, step int) string {
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
	return string(out)
}

// pattern is a case-insensitive substring → CDFType mapping. Entries are
// tested in order; first match wins.
type pattern struct {
	needle string
	typ    CDFType
}

// app2type maps NameOfApplication values to CDFTypes.
// Mirrors app2mime[] in libmagic.
var app2type = []pattern{
	{"Word", CDFTypeDoc},
	{"Excel", CDFTypeXls},
	{"Powerpoint", CDFTypePpt},
	{"Advanced Installer", CDFTypeInstaller},
	{"InstallShield", CDFTypeInstaller},
	{"Microsoft Patch Compiler", CDFTypeInstaller},
	{"NAnt", CDFTypeInstaller},
	{"Windows Installer", CDFTypeInstaller},
}

// name2type maps directory entry names to CDFTypes.
// Mirrors name2mime[] in libmagic.
var name2type = []pattern{
	{"Book", CDFTypeXls},
	{"Workbook", CDFTypeXls},
	{"WordDocument", CDFTypeDoc},
	{"PowerPoint", CDFTypePpt},
	{"DigitalSignature", CDFTypeInstaller},
}

// lookupSubstring returns the CDFType for the first entry in t whose needle
// is a case-insensitive substring of v. Mirrors C's strcasestr semantics
// under the C locale.
func lookupSubstring(v string, t []pattern) (CDFType, bool) {
	lv := strings.ToLower(v)
	for _, p := range t {
		if strings.Contains(lv, strings.ToLower(p.needle)) {
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

// sectionKey is a (directory entry name, type) pair.
type sectionKey struct {
	name string
	typ  uint8
}

// sectionTypes maps distinctive directory entries to CDFTypes — a flattened
// equivalent of sectioninfo[] in libmagic. Used as a fallback when no
// SummaryInformation stream is present.
var sectionTypes = map[sectionKey]CDFType{
	// libmagic uses application/encrypted, but that is not a registered media type.
	// For now, we skip identifying that and fall-back on CDFTypeGeneric
	// {"EncryptedPackage", dirTypeUserStream}:              CDFTypeEncrypted,
	// {"EncryptedSummary", dirTypeUserStream}:              CDFTypeEncrypted,
	{"Book", dirTypeUserStream}:                          CDFTypeXls,
	{"Workbook", dirTypeUserStream}:                      CDFTypeXls,
	{"WordDocument", dirTypeUserStream}:                  CDFTypeDoc,
	{"PowerPoint Document", dirTypeUserStream}:           CDFTypePpt,
	{"__properties_version1.0", dirTypeUserStream}:       CDFTypeMsg,
	{"__recip_version1.0_#00000000", dirTypeUserStorage}: CDFTypeMsg,
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
