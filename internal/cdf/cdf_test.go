package cdf

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"math/rand/v2"
	"testing"
)

const testSecSize = 512

// testSecSizes are the two sector sizes a CDF file may use: 512 (v3) and
// 4096 (v4). Table tests generate inputs for each.
var testSecSizes = []int{512, 4096}

// secP2 returns the power-of-two exponent stored in the header for secSize.
func secP2(secSize int) uint16 {
	return uint16(bits.TrailingZeros(uint(secSize)))
}

// testHeader builds a 512-byte CDF header with secSize-byte sectors and
// 64-byte short sectors.
func testHeader(secSize int, firstDirSec, firstSSAT int32, minStdStream uint32, masterSAT []int32) []byte {
	h := make([]byte, 512)
	binary.LittleEndian.PutUint64(h, cdfMagic)
	binary.LittleEndian.PutUint16(h[30:], secP2(secSize))
	binary.LittleEndian.PutUint16(h[32:], 6) // 64-byte short sectors
	binary.LittleEndian.PutUint32(h[48:], uint32(firstDirSec))
	binary.LittleEndian.PutUint32(h[56:], minStdStream)
	binary.LittleEndian.PutUint32(h[60:], uint32(firstSSAT))
	binary.LittleEndian.PutUint32(h[68:], uint32(0xFFFFFFFE)) // no extra master SAT sectors
	binary.LittleEndian.PutUint32(h[72:], 0)
	for i := 0; i < masterSATSize; i++ {
		v := int32(-1)
		if i < len(masterSAT) {
			v = masterSAT[i]
		}
		binary.LittleEndian.PutUint32(h[76+4*i:], uint32(v))
	}
	return h
}

// idSector encodes ids as a secSize-byte sector of little-endian int32s,
// padding with -1 (free).
func idSector(secSize int, ids ...int32) []byte {
	s := make([]byte, secSize)
	for i := 0; i < secSize/4; i++ {
		v := int32(-1)
		if i < len(ids) {
			v = ids[i]
		}
		binary.LittleEndian.PutUint32(s[4*i:], uint32(v))
	}
	return s
}

// padSector pads b with zeros to a full secSize-byte sector.
func padSector(secSize int, b []byte) []byte {
	s := make([]byte, secSize)
	copy(s, b)
	return s
}

// dirEntryBytes encodes a single 128-byte directory entry. The name is
// encoded as NUL-terminated UTF-16LE.
func dirEntryBytes(name string, typ uint8, first int32, size uint32, uuid []byte) []byte {
	e := make([]byte, dirEntrySize)
	for i := 0; i < len(name); i++ {
		binary.LittleEndian.PutUint16(e[2*i:], uint16(name[i]))
	}
	binary.LittleEndian.PutUint16(e[64:], uint16(2*(len(name)+1)))
	e[66] = typ
	copy(e[80:96], uuid)
	binary.LittleEndian.PutUint32(e[116:], uint32(first))
	binary.LittleEndian.PutUint32(e[120:], size)
	return e
}

// summaryStream builds a minimal (Doc)SummaryInformation property-set stream
// containing a single NameOfApplication (0x12) property. If wide is true the
// value is encoded as UTF-16LE, otherwise as ASCII.
func summaryStream(appName string, wide bool) []byte {
	strType, step := uint32(typeStringASCII), 1
	if wide {
		strType, step = typeStringWide, 2
	}
	strData := make([]byte, (len(appName)+1)*step) // NUL-terminated
	for i := 0; i < len(appName); i++ {
		strData[i*step] = appName[i]
	}
	// Section: 8-byte header, one 8-byte property declaration, then the value.
	section := make([]byte, 24, 24+len(strData))
	binary.LittleEndian.PutUint32(section[4:], 1) // one property
	binary.LittleEndian.PutUint32(section[8:], propIDNameOfApplication)
	binary.LittleEndian.PutUint32(section[12:], 16) // value offset in section
	binary.LittleEndian.PutUint32(section[16:], strType)
	binary.LittleEndian.PutUint32(section[20:], uint32(len(appName)+1))
	section = append(section, strData...)
	binary.LittleEndian.PutUint32(section, uint32(len(section))) // section length

	// 48-byte property-set header; the first section declaration starts at
	// 0x1c and its offset field (at 0x2c) points right after the header.
	stream := make([]byte, 48+len(section))
	binary.LittleEndian.PutUint32(stream[44:], 48)
	copy(stream[48:], section)
	return stream
}

type entrySpec struct {
	name string
	typ  uint8
}

// makeCDF assembles a complete single-SAT-sector CDF file with secSize-byte
// sectors:
//
//	sector 0: SAT
//	sector 1: directory (root entry, optional summary entry, extras)
//	sector 2: root storage's stream (short-sector pool, zero-filled)
//	sector 3: summary stream (when summary != nil)
//
// The 512-byte header is zero-padded to a full sector so sector 0 starts at
// offset secSize. minStdStream is 0 so every stream is read through the
// long-sector path. Extra entries carry no stream data; detection only
// inspects their names.
func makeCDF(secSize int, rootUUID []byte, summaryName string, summary []byte, extras ...entrySpec) []byte {
	if len(summary) > secSize {
		panic("summary stream larger than one sector")
	}
	dir := dirEntryBytes("Root Entry", dirTypeRootStorage, 2, uint32(secSize), rootUUID)
	if summary != nil {
		dir = append(dir, dirEntryBytes(summaryName, dirTypeUserStream, 3, uint32(len(summary)), nil)...)
	}
	for _, e := range extras {
		dir = append(dir, dirEntryBytes(e.name, e.typ, -2, 0, nil)...)
	}
	if len(dir) > secSize {
		panic("directory larger than one sector")
	}

	out := padSector(secSize, testHeader(secSize, 1, -2, 0, []int32{0}))
	out = append(out, idSector(secSize, -3, -2, -2, -2)...) // all chains are single-sector
	out = append(out, padSector(secSize, dir)...)
	out = append(out, make([]byte, secSize)...)
	out = append(out, padSector(secSize, summary)...)
	return out
}

func TestDetectInvalidInput(t *testing.T) {
	badSectorSize := testHeader(testSecSize, 1, -2, 0, nil)
	binary.LittleEndian.PutUint16(badSectorSize[30:], 21) // secP2 > 20

	tooSmallSector := testHeader(testSecSize, 1, -2, 0, nil)
	binary.LittleEndian.PutUint16(tooSmallSector[30:], 5) // 32 < dirEntrySize

	// firstDirSec * secSize == 2^31 overflows int on 32bit arch and causes
	// slice-bounds panic.
	dirSecOverflow := testHeader(testSecSize, 1<<22-1, -2, 0, nil)

	// secSize=4096, masterSAT[0]=0, firstSSAT=0. sector(0) is only 1 byte (file
	// ends one byte into it), so collectIDs must not loop forever trying to read
	// int32s from a sub-4-byte buffer.
	secSize4K := 4096
	truncatedSSAT := append(
		padSector(secSize4K, testHeader(secSize4K, -2, 0, 0, []int32{0})),
		0x00, // one-byte sector 0 — too small for even one int32
	)

	tests := []struct {
		name string
		data []byte
	}{
		{"nil", nil},
		{"short", []byte{0xD0, 0xCF, 0x11, 0xE0}},
		{"bad magic", make([]byte, 512)},
		{"bad sector size", badSectorSize},
		{"sector smaller than dir entry", tooSmallSector},
		{"dir sector overflows int32", dirSecOverflow},
		{"truncated sector in SSAT chain loops forever", truncatedSSAT},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Detect(tt.data); got != CDFTypeGeneric {
				t.Errorf("Detect() = %v, want CDFTypeGeneric", got)
			}
		})
	}
}

func TestDetectMSIRootCLSID(t *testing.T) {
	for _, secSize := range testSecSizes {
		data := makeCDF(secSize, msiCLSID, "", nil)
		if got := Detect(data); got != CDFTypeInstaller {
			t.Errorf("Detect(secSize=%d) = %v, want CDFTypeInstaller", secSize, got)
		}
	}
}

func TestDetectFromAppName(t *testing.T) {
	tests := []struct {
		app  string
		want CDFType
	}{
		{"Microsoft Office Word", CDFTypeDoc},
		{"Microsoft Excel", CDFTypeXls},
		{"Microsoft Powerpoint", CDFTypePpt},
		{"Advanced Installer 19.0", CDFTypeInstaller},
		{"InstallShield", CDFTypeInstaller},
		{"Microsoft Patch Compiler", CDFTypeInstaller},
		{"NAnt", CDFTypeInstaller},
		{"Windows Installer XML Toolset", CDFTypeInstaller},
		{"MICROSOFT WORD", CDFTypeDoc}, // matching is case-insensitive
	}
	summaryNames := []string{"\x05SummaryInformation", "\x05DocumentSummaryInformation"}
	for _, secSize := range testSecSizes {
		for _, summaryName := range summaryNames {
			for _, wide := range []bool{false, true} {
				for _, tt := range tests {
					t.Run(tt.app, func(t *testing.T) {
						data := makeCDF(secSize, nil, summaryName, summaryStream(tt.app, wide))
						if got := Detect(data); got != tt.want {
							t.Errorf("Detect(secSize=%d, summary=%q, app=%q, wide=%v) = %v, want %v",
								secSize, summaryName, tt.app, wide, got, tt.want)
						}
					})
				}
			}
		}
	}
}

// TestDetectFromSiblingNames exercises the fallback inside detectFromSummary:
// the application name does not match, so detection relies on the names of
// the other directory entries.
func TestDetectFromSiblingNames(t *testing.T) {
	tests := []struct {
		sibling string
		want    CDFType
	}{
		{"Book", CDFTypeXls},
		{"Workbook", CDFTypeXls},
		{"WordDocument", CDFTypeDoc},
		{"PowerPoint Document", CDFTypePpt},
		{"\x05DigitalSignature", CDFTypeInstaller},
		// A non-distinctive sibling name yields no match, so detection
		// falls back to the generic OLE storage type.
		{"Contents", CDFTypeGeneric},
	}
	for _, secSize := range testSecSizes {
		for _, tt := range tests {
			t.Run(tt.sibling, func(t *testing.T) {
				summary := summaryStream("Unknown Application", false)
				data := makeCDF(secSize, nil, "\x05SummaryInformation", summary, entrySpec{tt.sibling, dirTypeUserStream})
				if got := Detect(data); got != tt.want {
					t.Errorf("Detect(secSize=%d) = %v, want %v", secSize, got, tt.want)
				}
			})
		}
	}
}

// TestDetectFromSectionNames exercises the final fallback in Detect: no
// summary stream at all, detection relies on distinctive directory entries.
func TestDetectFromSectionNames(t *testing.T) {
	tests := []struct {
		entry entrySpec
		want  CDFType
	}{
		{entrySpec{"EncryptedPackage", dirTypeUserStream}, CDFTypeGeneric},
		{entrySpec{"EncryptedSummary", dirTypeUserStream}, CDFTypeGeneric},
		{entrySpec{"Book", dirTypeUserStream}, CDFTypeXls},
		{entrySpec{"Workbook", dirTypeUserStream}, CDFTypeXls},
		{entrySpec{"WordDocument", dirTypeUserStream}, CDFTypeDoc},
		{entrySpec{"PowerPoint Document", dirTypeUserStream}, CDFTypePpt},
		{entrySpec{"__properties_version1.0", dirTypeUserStream}, CDFTypeMsg},
		{entrySpec{"__recip_version1.0_#00000000", dirTypeUserStorage}, CDFTypeMsg},
		// A non-distinctive entry matches nothing and degrades to generic.
		{entrySpec{"Contents", dirTypeUserStream}, CDFTypeGeneric},
	}
	for _, secSize := range testSecSizes {
		for _, tt := range tests {
			t.Run(tt.entry.name, func(t *testing.T) {
				data := makeCDF(secSize, nil, "", nil, tt.entry)
				if got := Detect(data); got != tt.want {
					t.Errorf("Detect(secSize=%d) = %v, want %v", secSize, got, tt.want)
				}
			})
		}
	}
}

// TestDetectMasterSATExtension builds a file whose SAT does not fit in the
// 109 header slots, so part of it must be loaded through the master SAT
// extension chain. The directory and streams live in sectors mapped only by
// the extension SAT sector, so detection fails unless the chain is followed.
func TestDetectMasterSATExtension(t *testing.T) {
	const (
		msatSec    = 109   // master SAT extension sector
		extSATSec  = 110   // SAT sector referenced from the extension
		dirSec     = 13952 // first sector mapped by extSATSec (109*128)
		rootSec    = 13953
		summarySec = 13954
		nSectors   = summarySec + 1
	)
	summary := summaryStream("Microsoft Excel", false)

	data := make([]byte, (1+nSectors)*testSecSize)
	sector := func(id int32) []byte {
		off := (1 + int(id)) * testSecSize
		return data[off : off+testSecSize]
	}

	masterSAT := make([]int32, masterSATSize)
	for i := range masterSAT {
		masterSAT[i] = int32(i) // SAT occupies sectors 0..108
	}
	copy(data, testHeader(testSecSize, dirSec, -2, 0, masterSAT))
	binary.LittleEndian.PutUint32(data[68:], msatSec) // first extension sector
	binary.LittleEndian.PutUint32(data[72:], 1)       // one extension sector

	// Mark every SAT entry free, then set the entries actually used.
	for i := int32(0); i < masterSATSize; i++ {
		copy(sector(i), idSector(testSecSize))
	}
	satIDs := make([]int32, testSecSize/4)
	for i := range satIDs {
		satIDs[i] = -1
	}
	for i := 0; i <= extSATSec; i++ {
		satIDs[i] = -3 // sectors 0..110 hold SAT/MSAT data
	}
	copy(sector(0), idSector(testSecSize, satIDs...))

	// The extension lists one extra SAT sector; remaining slots are free.
	copy(sector(msatSec), idSector(testSecSize, extSATSec))
	// extSATSec maps sectors 13952..14079: dir, root stream and summary are
	// all single-sector chains.
	copy(sector(extSATSec), idSector(testSecSize, -2, -2, -2))

	dir := dirEntryBytes("Root Entry", dirTypeRootStorage, rootSec, testSecSize, nil)
	dir = append(dir, dirEntryBytes("\x05SummaryInformation", dirTypeUserStream, summarySec, uint32(len(summary)), nil)...)
	copy(sector(dirSec), dir)
	copy(sector(summarySec), summary)

	if got := Detect(data); got != CDFTypeXls {
		t.Errorf("Detect() = %v, want CDFTypeXls", got)
	}
}

// TestDetectShortStreamSummary stores the summary stream in the short-sector
// pool, exercising the SSAT and readShort paths.
func TestDetectShortStreamSummary(t *testing.T) {
	summary := summaryStream("Microsoft Office Word", false)
	if len(summary) <= 64 || len(summary) > testSecSize {
		t.Fatalf("summary len %d, want >64 and <=%d to span short sectors", len(summary), testSecSize)
	}

	dir := dirEntryBytes("Root Entry", dirTypeRootStorage, 2, testSecSize, nil)
	// streamFirst is a short-sector id because size < minStdStream.
	dir = append(dir, dirEntryBytes("\x05SummaryInformation", dirTypeUserStream, 0, uint32(len(summary)), nil)...)

	// sector 0: SAT, sector 1: dir, sector 2: short-sector pool, sector 3: SSAT
	data := testHeader(testSecSize, 1, 3, 4096, []int32{0})
	data = append(data, idSector(testSecSize, -3, -2, -2, -2)...)
	data = append(data, padSector(testSecSize, dir)...)
	data = append(data, padSector(testSecSize, summary)...) // pool holds the summary in 64-byte short sectors
	nShort := int32((len(summary) + 63) / 64)
	ssat := make([]int32, nShort)
	for i := int32(0); i < nShort-1; i++ {
		ssat[i] = i + 1
	}
	ssat[nShort-1] = -2
	data = append(data, idSector(testSecSize, ssat...)...)

	if got := Detect(data); got != CDFTypeDoc {
		t.Errorf("Detect() = %v, want CDFTypeDoc", got)
	}
}

func put16(b []byte, off int, v uint16) { binary.LittleEndian.PutUint16(b[off:], v) }
func put32(b []byte, off int, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }

// TestBuildSATTruncatedMSATNoPanic feeds a CDF whose master-SAT points at a
// partial (truncated) sector. The inner SAT loop walks perSec entries without
// checking len(msa), so a short sector triggers an index-out-of-range panic.
// CDF files use a sector size of either 512 (secP2 9) or 4096 (secP2 12); both
// are exercised here.
func TestBuildSATTruncatedMSATNoPanic(t *testing.T) {
	for _, tc := range []struct {
		name    string
		secP2   uint16
		secSize int
	}{
		{"sector 512", 9, 512},
		{"sector 4096", 12, 4096},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// One extra int32 (4 bytes) past the header: sector 0 starts at
			// offset secSize and is only 4 bytes long, so readInt32s yields a
			// single entry while the loop expects perSec of them.
			data := make([]byte, tc.secSize+4)
			binary.LittleEndian.PutUint64(data, cdfMagic)
			put16(data, 30, tc.secP2)
			put16(data, 32, 6)          // shortP2
			put32(data, 48, 0xFFFFFFFF) // firstDirSec = -1
			put32(data, 56, 4096)       // minStdStream
			put32(data, 60, 0xFFFFFFFF) // firstSSAT = -1
			put32(data, 68, 0)          // firstMSAT = sector 0
			put32(data, 72, 1)          // nMSAT = 1
			// masterSAT[0] points past EOF so the header loop breaks
			// (errTruncated) and execution falls through to the master-SAT
			// extension loop.
			put32(data, 76, 100)
			for i := 1; i < masterSATSize; i++ {
				put32(data, 76+4*i, 0xFFFFFFFF)
			}
			// data[secSize:secSize+4] = 0 -> msa = [0], a non-negative secid so
			// the loop keeps going and then dereferences msa[1].
			put32(data, tc.secSize, 0)

			if cdf := Detect(data); cdf != CDFTypeGeneric {
				t.Errorf("expected: %q, got: %q", CDFTypeGeneric, cdf)
			}
		})
	}
}

// TestCyclicChainBounded feeds a CDF whose SAT contains a self-referential
// entry (sector 1 points at itself) together with the maximum sector size of
// 1MiB (secP2 20). Walking such a cycle must be bounded by the input size: a
// stream can never be longer than the file. Without that bound the reader would
// follow the cycle thousands of times, each step appending a whole 1MiB sector,
// and attempt a multi-gigabyte allocation.
func TestCyclicChainBounded(t *testing.T) {
	const secP2 = 20
	secSize := 1 << secP2
	// Header region [0,secSize), then sector 0 and sector 1.
	data := make([]byte, 3*secSize)
	binary.LittleEndian.PutUint64(data, cdfMagic)
	put16(data, 30, secP2)
	put16(data, 32, 6)          // shortP2
	put32(data, 48, 1)          // firstDirSec = sector 1
	put32(data, 56, 4096)       // minStdStream
	put32(data, 60, 0xFFFFFFFF) // firstSSAT = -1
	put32(data, 68, 0xFFFFFFFF) // firstMSAT = -1
	put32(data, 72, 0)          // nMSAT = 0
	// masterSAT[0] = sector 0 holds the SAT; the rest are unused (-1).
	put32(data, 76, 0)
	for i := 1; i < masterSATSize; i++ {
		put32(data, 76+4*i, 0xFFFFFFFF)
	}
	// SAT lives in sector 0 (file offset secSize). Make sat[1] = 1 so the
	// directory chain starting at sector 1 loops forever.
	put32(data, secSize+4, 1)

	if cdf := Detect(data); cdf != CDFTypeGeneric {
		t.Errorf("expected: %q, got: %q", CDFTypeGeneric, cdf)
	}
}

// TestSummaryAppNameNoPanic feeds summaryAppName streams whose length/offset
// fields are crafted to overflow or go negative, plus one well-formed stream.
// Malformed inputs must not panic and must yield "". The sdOff case is
// 386-specific: int(sdOff) is negative there, so the bounds check is bypassed
// and stream[sdOff:] would panic.
func TestSummaryAppNameNoPanic(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stream func() []byte
		want   string
	}{{
		// A well-formed stream with NameOfApplication = "Word".
		name: "valid app name",
		stream: func() []byte {
			stream := make([]byte, 80)
			put32(stream, sectionDeclOffset+16, 48) // sdOff -> section at 48
			sec := stream[48:]
			put32(sec, 0, 29) // shLen
			put32(sec, 4, 1)  // nProps
			put32(sec, 8, propIDNameOfApplication)
			put32(sec, 12, 16)              // off -> value at section[16:]
			put32(sec, 16, typeStringASCII) // step = 1
			put32(sec, 20, 5)               // slen ("Word\0")
			copy(sec[24:], "Word\x00")
			return stream
		},
		want: "Word",
	}, {
		// off is near math.MaxUint32 so that off+8 overflows uint32 and
		// bypasses the bounds check, then section[off:] panics.
		name: "offset overflow",
		stream: func() []byte {
			stream := make([]byte, 64)
			put32(stream, sectionDeclOffset+16, 48) // sdOff -> section at 48
			sec := stream[48:]
			put32(sec, 0, 16) // shLen
			put32(sec, 4, 1)  // nProps
			put32(sec, 8, propIDNameOfApplication)
			put32(sec, 12, 0xFFFFFFFF) // off; off+8 wraps to 7
			return stream
		},
		want: "",
	}, {
		// slen*step overflows uint32, producing end < start so that
		// section[start:end] panics.
		name: "length overflow",
		stream: func() []byte {
			stream := make([]byte, 80)
			put32(stream, sectionDeclOffset+16, 48) // sdOff -> section at 48
			sec := stream[48:]
			put32(sec, 0, 32) // shLen
			put32(sec, 4, 1)  // nProps
			put32(sec, 8, propIDNameOfApplication)
			put32(sec, 12, 16)             // off -> value at section[16:]
			put32(sec, 16, typeStringWide) // step = 2
			put32(sec, 20, 0xFFFFFFFF)     // slen; slen*2 overflows uint32
			return stream
		},
		want: "",
	}, {
		// sdOff above math.MaxInt32. On 386 int(sdOff) is negative, the
		// bounds check passes and stream[sdOff:] panics.
		name: "sdOff negative on 386",
		stream: func() []byte {
			stream := make([]byte, 64)
			put32(stream, sectionDeclOffset+16, 0x80000000) // sdOff > MaxInt32
			return stream
		},
		want: "",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(summaryAppName(tc.stream())); got != tc.want {
				t.Errorf("expected: %q, got: %q", tc.want, got)
			}
		})
	}
}

func BenchmarkDetect(b *testing.B) {
	cases := []struct {
		name  string
		build func(secSize int) []byte
	}{{
		name: "MSIRootCLSID",
		build: func(secSize int) []byte {
			return makeCDF(secSize, msiCLSID, "", nil)
		},
	}, {
		name: "AppNameASCII",
		build: func(secSize int) []byte {
			return makeCDF(secSize, nil, "\x05SummaryInformation",
				summaryStream("Microsoft Office Word", false))
		},
	}, {
		name: "AppNameWide",
		build: func(secSize int) []byte {
			return makeCDF(secSize, nil, "\x05SummaryInformation",
				summaryStream("Microsoft Excel", true))
		},
	}, {
		name: "SiblingName",
		build: func(secSize int) []byte {
			return makeCDF(secSize, nil, "\x05SummaryInformation",
				summaryStream("Unknown Application", false),
				entrySpec{"WordDocument", dirTypeUserStream})
		},
	}, {
		name: "SectionName",
		build: func(secSize int) []byte {
			return makeCDF(secSize, nil, "", nil,
				entrySpec{"EncryptedPackage", dirTypeUserStream})
		},
	}, {
		name: "Generic",
		build: func(secSize int) []byte {
			return makeCDF(secSize, nil, "", nil,
				entrySpec{"Contents", dirTypeUserStream})
		},
	}}

	for _, secSize := range testSecSizes {
		for _, bc := range cases {
			data := bc.build(secSize)
			b.Run(fmt.Sprintf("%s/sec%d", bc.name, secSize), func(b *testing.B) {
				b.SetBytes(int64(len(data)))
				b.ReportAllocs()
				for b.Loop() {
					Detect(data)
				}
			})
		}
	}
}

func FuzzDetect(f *testing.F) {
	for _, secSize := range testSecSizes {
		// MSI installer (root-storage CLSID).
		f.Add(makeCDF(secSize, msiCLSID, "", nil))

		// Detection from the summary stream's application name, both
		// ASCII and UTF-16LE encodings.
		for _, wide := range []bool{false, true} {
			f.Add(makeCDF(secSize, nil, "\x05SummaryInformation",
				summaryStream("Microsoft Office Word", wide)))
			f.Add(makeCDF(secSize, nil, "\x05DocumentSummaryInformation",
				summaryStream("Microsoft Excel", wide)))
		}

		// Detection from sibling entry names when the app name is unknown.
		f.Add(makeCDF(secSize, nil, "\x05SummaryInformation",
			summaryStream("Unknown Application", false),
			entrySpec{"WordDocument", dirTypeUserStream}))

		// Summary present but no recognizable names -> Generic fallback.
		f.Add(makeCDF(secSize, nil, "\x05SummaryInformation",
			summaryStream("Unknown Application", false),
			entrySpec{"Contents", dirTypeUserStream}))

		// Detection from section/entry names with no summary stream.
		f.Add(makeCDF(secSize, nil, "", nil, entrySpec{"EncryptedPackage", dirTypeUserStream}))
		f.Add(makeCDF(secSize, nil, "", nil, entrySpec{"Workbook", dirTypeUserStream}))
		f.Add(makeCDF(secSize, nil, "", nil, entrySpec{"PowerPoint Document", dirTypeUserStream}))
		f.Add(makeCDF(secSize, nil, "", nil, entrySpec{"__properties_version1.0", dirTypeUserStream}))

		// Generic OLE storage with no distinctive entries.
		f.Add(makeCDF(secSize, nil, "", nil, entrySpec{"Contents", dirTypeUserStream}))
	}

	// A bare header with no body and a non-CDF input.
	f.Add(testHeader(512, 1, -2, 0, []int32{0}))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		for i := 0; i < 100 && i < len(data); i++ {
			j := rand.IntN(len(data))
			_ = Detect(data[:j]) // must not panic on any input
		}
	})
}
