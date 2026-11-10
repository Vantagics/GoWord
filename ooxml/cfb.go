package ooxml

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

const (
	cfbSignature  = "\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1"
	cfbSector     = 512
	cfbMiniSector = 64
	cfbEndChain   = 0xFFFFFFFE
	cfbFreeSect   = 0xFFFFFFFF
	cfbFATSect    = 0xFFFFFFFD
	cfbDIFSect    = 0xFFFFFFFC
	cfbNoStream   = 0xFFFFFFFF
)

func isCompoundFile(data []byte) bool {
	return len(data) >= 8 && string(data[:8]) == cfbSignature
}

type cfbDir struct {
	name             string
	typ              byte
	left, right, kid int
	start            uint32
	size             uint64
	data             []byte
}

// readCompoundFile returns every stream in a version-3 (or version-4) compound file.
func readCompoundFile(data []byte) (map[string][]byte, error) {
	if !isCompoundFile(data) || len(data) < cfbSector {
		return nil, fmt.Errorf("not a compound file")
	}
	sectorShift := binary.LittleEndian.Uint16(data[0x1E:])
	if sectorShift < 9 || sectorShift > 12 {
		return nil, fmt.Errorf("unsupported compound sector shift %d", sectorShift)
	}
	secSize := 1 << sectorShift
	major := binary.LittleEndian.Uint16(data[0x1A:])
	headerSize := cfbSector
	if major >= 4 {
		headerSize = secSize
	}
	if len(data) < headerSize {
		return nil, fmt.Errorf("compound file header is truncated")
	}
	u := func(off int) uint32 { return binary.LittleEndian.Uint32(data[off:]) }
	numFAT := int(u(0x2C))
	dirStart := u(0x30)
	cutoff := u(0x38)
	miniFATStart := u(0x3C)
	difatStart := u(0x44)
	numDIFAT := int(u(0x48))
	if numFAT <= 0 {
		return nil, fmt.Errorf("compound file has no FAT")
	}

	sector := func(n uint32) ([]byte, error) {
		off := headerSize + int(n)*secSize
		if n >= 0xFFFFFFF0 || off < 0 || off+secSize > len(data) {
			return nil, fmt.Errorf("compound sector %d is outside the file", n)
		}
		return data[off : off+secSize], nil
	}
	var difat []uint32
	slots := (headerSize - 0x4C) / 4
	for i := 0; i < slots; i++ {
		v := u(0x4C + 4*i)
		if v >= 0xFFFFFFF0 {
			break
		}
		difat = append(difat, v)
	}
	next := difatStart
	for n := 0; n < numDIFAT && next < 0xFFFFFFF0; n++ {
		raw, err := sector(next)
		if err != nil {
			return nil, err
		}
		for i := 0; i < (secSize/4)-1; i++ {
			v := binary.LittleEndian.Uint32(raw[i*4:])
			if v >= 0xFFFFFFF0 {
				break
			}
			difat = append(difat, v)
		}
		next = binary.LittleEndian.Uint32(raw[secSize-4:])
	}
	if len(difat) > numFAT {
		difat = difat[:numFAT]
	}
	fat := make([]byte, 0, len(difat)*secSize)
	for _, s := range difat {
		raw, err := sector(s)
		if err != nil {
			return nil, err
		}
		fat = append(fat, raw...)
	}
	fatAt := func(n uint32) (uint32, error) {
		off := int(n) * 4
		if off < 0 || off+4 > len(fat) {
			return 0, fmt.Errorf("FAT index %d is outside the FAT", n)
		}
		return binary.LittleEndian.Uint32(fat[off:]), nil
	}
	readChain := func(start uint32, size int) ([]byte, error) {
		var out []byte
		s := start
		seen := map[uint32]bool{}
		for s < 0xFFFFFFF0 {
			if seen[s] {
				return nil, fmt.Errorf("compound sector chain loops")
			}
			seen[s] = true
			raw, err := sector(s)
			if err != nil {
				return nil, err
			}
			out = append(out, raw...)
			if size > 0 && len(out) >= size {
				return out[:size], nil
			}
			var errNext error
			s, errNext = fatAt(s)
			if errNext != nil {
				return nil, errNext
			}
			if len(seen) > 1_000_000 {
				return nil, fmt.Errorf("compound sector chain is too long")
			}
		}
		if size > 0 && len(out) > size {
			return out[:size], nil
		}
		return out, nil
	}

	dirBytes, err := readChain(dirStart, 0)
	if err != nil {
		return nil, err
	}
	var entries []cfbDir
	for off := 0; off+128 <= len(dirBytes); off += 128 {
		e := dirBytes[off : off+128]
		nameLen := int(binary.LittleEndian.Uint16(e[64:]))
		typ := e[66]
		if typ == 0 {
			continue
		}
		if nameLen < 2 {
			nameLen = 2
		}
		if nameLen > 64 {
			nameLen = 64
		}
		units := (nameLen / 2) - 1
		if units < 0 {
			units = 0
		}
		raw := make([]uint16, units)
		for i := range raw {
			raw[i] = binary.LittleEndian.Uint16(e[i*2:])
		}
		entries = append(entries, cfbDir{
			name:  string(utf16.Decode(raw)),
			typ:   typ,
			left:  int(binary.LittleEndian.Uint32(e[68:])),
			right: int(binary.LittleEndian.Uint32(e[72:])),
			kid:   int(binary.LittleEndian.Uint32(e[76:])),
			start: binary.LittleEndian.Uint32(e[116:]),
			size:  binary.LittleEndian.Uint64(e[120:]),
		})
	}
	var root *cfbDir
	for i := range entries {
		if entries[i].typ == 5 {
			root = &entries[i]
			break
		}
	}
	if root == nil {
		return nil, fmt.Errorf("compound file has no root entry")
	}
	miniFAT, err := readChain(miniFATStart, 0)
	if err != nil && miniFATStart < 0xFFFFFFF0 {
		return nil, err
	}
	rootData, err := readChain(root.start, int(root.size))
	if err != nil {
		return nil, err
	}
	miniAt := func(n uint32) uint32 {
		off := int(n) * 4
		if off < 0 || off+4 > len(miniFAT) {
			return cfbEndChain
		}
		return binary.LittleEndian.Uint32(miniFAT[off:])
	}
	readMini := func(start uint32, size int) ([]byte, error) {
		var out []byte
		s := start
		seen := map[uint32]bool{}
		for s < 0xFFFFFFF0 && (size == 0 || len(out) < size) {
			if seen[s] {
				return nil, fmt.Errorf("mini sector chain loops")
			}
			seen[s] = true
			off := int(s) * cfbMiniSector
			if off+cfbMiniSector > len(rootData) {
				break
			}
			out = append(out, rootData[off:off+cfbMiniSector]...)
			s = miniAt(s)
		}
		if size > 0 && len(out) > size {
			return out[:size], nil
		}
		return out, nil
	}
	out := map[string][]byte{}
	for _, e := range entries {
		if e.typ != 2 {
			continue
		}
		var blob []byte
		var err error
		if uint32(e.size) < cutoff {
			blob, err = readMini(e.start, int(e.size))
		} else {
			blob, err = readChain(e.start, int(e.size))
		}
		if err != nil {
			return nil, err
		}
		out[e.name] = blob
	}
	return out, nil
}

// writeOfficePackage builds a compound file with the streams Office looks up
// when it opens an encrypted document: EncryptionInfo, EncryptedPackage, and
// the DataSpaces tree those streams are registered in.
func writeOfficePackage(encryptionInfo, encryptedPackage []byte) ([]byte, error) {
	// Streams shorter than 4096 live in the mini stream. The encrypted package
	// is a FAT stream; pad a tiny one so readers do not look for it there.
	if len(encryptedPackage) < 4096 {
		padded := make([]byte, 4096)
		copy(padded, encryptedPackage)
		encryptedPackage = padded
	}
	entries := []cfbDir{
		{name: "Root Entry", typ: 5, left: -1, right: -1, kid: 10},
		{name: "EncryptedPackage", typ: 2, left: -1, right: -1, kid: -1, data: encryptedPackage},
		{name: "\x06DataSpaces", typ: 1, left: -1, right: -1, kid: 4},
		{name: "Version", typ: 2, left: -1, right: -1, kid: -1, data: dataSpaceVersion},
		{name: "DataSpaceMap", typ: 2, left: 3, right: 5, kid: -1, data: dataSpaceMap},
		{name: "DataSpaceInfo", typ: 1, left: -1, right: 7, kid: 6},
		{name: "StrongEncryptionDataSpace", typ: 2, left: -1, right: -1, kid: -1, data: strongEncryptionDataSpace},
		{name: "TransformInfo", typ: 1, left: -1, right: -1, kid: 8},
		{name: "StrongEncryptionTransform", typ: 1, left: -1, right: -1, kid: 9},
		{name: "\x06Primary", typ: 2, left: -1, right: -1, kid: -1, data: dataSpacePrimary},
		{name: "EncryptionInfo", typ: 2, left: 2, right: 1, kid: -1, data: encryptionInfo},
	}
	miniStream := make([]byte, 0)
	miniFAT := make([]uint32, 0)
	for i := range entries {
		if entries[i].typ != 2 || len(entries[i].data) == 0 || len(entries[i].data) >= 4096 {
			continue
		}
		entries[i].start = uint32(len(miniFAT))
		entries[i].size = uint64(len(entries[i].data))
		padded := padTo(entries[i].data, cfbMiniSector)
		for s := 0; s < len(padded)/cfbMiniSector; s++ {
			if s == len(padded)/cfbMiniSector-1 {
				miniFAT = append(miniFAT, cfbEndChain)
			} else {
				miniFAT = append(miniFAT, uint32(len(miniFAT)+1))
			}
		}
		miniStream = append(miniStream, padded...)
	}
	entries[0].size = uint64(len(miniStream))

	dirSectors := sectorsNeeded(len(entries)*128, cfbSector)
	miniFATSectors := sectorsNeeded(len(miniFAT)*4, cfbSector)
	if miniFATSectors == 0 {
		miniFATSectors = 1
	}
	rootSectors := sectorsNeeded(len(miniStream), cfbSector)
	pkgSectors := sectorsNeeded(len(encryptedPackage), cfbSector)
	fatSectors, difatSectors := 1, 0
	for n := 0; n < 8; n++ {
		total := fatSectors + difatSectors + dirSectors + miniFATSectors + rootSectors + pkgSectors
		needFat := sectorsNeeded(total*4, cfbSector)
		if needFat < 1 {
			needFat = 1
		}
		needDifat := 0
		if needFat > 109 {
			needDifat = (needFat - 109 + 126) / 127
		}
		if needFat == fatSectors && needDifat == difatSectors {
			break
		}
		fatSectors, difatSectors = needFat, needDifat
	}
	total := fatSectors + difatSectors + dirSectors + miniFATSectors + rootSectors + pkgSectors
	at := 0
	fatAt := at
	at += fatSectors
	difatAt := at
	at += difatSectors
	dirAt := at
	at += dirSectors
	miniAt := at
	at += miniFATSectors
	rootAt := at
	at += rootSectors
	pkgAt := at

	entries[0].start = uint32(rootAt)
	entries[1].start = uint32(pkgAt)
	entries[1].size = uint64(len(encryptedPackage))

	fat := make([]uint32, fatSectors*(cfbSector/4))
	for i := range fat {
		fat[i] = cfbFreeSect
	}
	for i := 0; i < fatSectors; i++ {
		fat[fatAt+i] = cfbFATSect
	}
	for i := 0; i < difatSectors; i++ {
		fat[difatAt+i] = cfbDIFSect
	}
	chain := func(start, n int) {
		for i := 0; i < n; i++ {
			if i == n-1 {
				fat[start+i] = cfbEndChain
			} else {
				fat[start+i] = uint32(start + i + 1)
			}
		}
	}
	chain(dirAt, dirSectors)
	chain(miniAt, miniFATSectors)
	chain(rootAt, rootSectors)
	chain(pkgAt, pkgSectors)

	sectors := make([]byte, total*cfbSector)
	putSector := func(index int, b []byte) {
		off := index * cfbSector
		copy(sectors[off:off+cfbSector], b)
	}
	for i := 0; i < fatSectors; i++ {
		raw := make([]byte, cfbSector)
		for j := 0; j < cfbSector/4; j++ {
			binary.LittleEndian.PutUint32(raw[j*4:], fat[i*(cfbSector/4)+j])
		}
		putSector(fatAt+i, raw)
	}
	// DIFAT sectors list FAT sectors that did not fit in the header.
	fatList := make([]uint32, fatSectors)
	for i := range fatList {
		fatList[i] = uint32(fatAt + i)
	}
	if difatSectors > 0 {
		extra := fatList[109:]
		for i := 0; i < difatSectors; i++ {
			raw := make([]byte, cfbSector)
			for j := 0; j < (cfbSector/4)-1; j++ {
				idx := i*((cfbSector/4)-1) + j
				v := uint32(cfbFreeSect)
				if idx < len(extra) {
					v = extra[idx]
				}
				binary.LittleEndian.PutUint32(raw[j*4:], v)
			}
			next := uint32(cfbEndChain)
			if i+1 < difatSectors {
				next = uint32(difatAt + i + 1)
			}
			binary.LittleEndian.PutUint32(raw[cfbSector-4:], next)
			putSector(difatAt+i, raw)
		}
	}
	dirRaw := make([]byte, dirSectors*cfbSector)
	for i, e := range entries {
		writeDirEntry(dirRaw[i*128:(i+1)*128], e)
	}
	for i := 0; i < dirSectors; i++ {
		putSector(dirAt+i, dirRaw[i*cfbSector:(i+1)*cfbSector])
	}
	miniRaw := make([]byte, miniFATSectors*cfbSector)
	for i := range miniRaw {
		miniRaw[i] = 0xFF
	}
	for i, v := range miniFAT {
		binary.LittleEndian.PutUint32(miniRaw[i*4:], v)
	}
	for i := 0; i < miniFATSectors; i++ {
		putSector(miniAt+i, miniRaw[i*cfbSector:(i+1)*cfbSector])
	}
	for i := 0; i < rootSectors; i++ {
		chunk := miniStream[i*cfbSector:]
		if len(chunk) > cfbSector {
			chunk = chunk[:cfbSector]
		}
		putSector(rootAt+i, chunk)
	}
	for i := 0; i < pkgSectors; i++ {
		chunk := encryptedPackage[i*cfbSector:]
		if len(chunk) > cfbSector {
			chunk = chunk[:cfbSector]
		}
		putSector(pkgAt+i, chunk)
	}

	header := make([]byte, cfbSector)
	copy(header, cfbSignature)
	binary.LittleEndian.PutUint16(header[0x18:], 0x003E)
	binary.LittleEndian.PutUint16(header[0x1A:], 3)
	binary.LittleEndian.PutUint16(header[0x1C:], 0xFFFE)
	binary.LittleEndian.PutUint16(header[0x1E:], 9)
	binary.LittleEndian.PutUint16(header[0x20:], 6)
	binary.LittleEndian.PutUint32(header[0x2C:], uint32(fatSectors))
	binary.LittleEndian.PutUint32(header[0x30:], uint32(dirAt))
	binary.LittleEndian.PutUint32(header[0x38:], 4096)
	binary.LittleEndian.PutUint32(header[0x3C:], uint32(miniAt))
	binary.LittleEndian.PutUint32(header[0x40:], uint32(miniFATSectors))
	if difatSectors == 0 {
		binary.LittleEndian.PutUint32(header[0x44:], cfbEndChain)
	} else {
		binary.LittleEndian.PutUint32(header[0x44:], uint32(difatAt))
	}
	binary.LittleEndian.PutUint32(header[0x48:], uint32(difatSectors))
	for i := 0; i < 109; i++ {
		v := uint32(cfbFreeSect)
		if i < len(fatList) && i < 109 {
			v = fatList[i]
		}
		binary.LittleEndian.PutUint32(header[0x4C+4*i:], v)
	}
	return append(header, sectors...), nil
}

func writeDirEntry(dst []byte, e cfbDir) {
	for i := range dst {
		dst[i] = 0
	}
	encoded := utf16.Encode([]rune(e.name))
	for i, u := range encoded {
		if i >= 31 {
			break
		}
		binary.LittleEndian.PutUint16(dst[i*2:], u)
	}
	n := len(encoded)
	if n > 31 {
		n = 31
	}
	binary.LittleEndian.PutUint16(dst[64:], uint16((n+1)*2))
	dst[66] = e.typ
	dst[67] = 1 // black
	putLink := func(off, v int) {
		u := uint32(cfbNoStream)
		if v >= 0 {
			u = uint32(v)
		}
		binary.LittleEndian.PutUint32(dst[off:], u)
	}
	putLink(68, e.left)
	putLink(72, e.right)
	putLink(76, e.kid)
	binary.LittleEndian.PutUint32(dst[116:], e.start)
	binary.LittleEndian.PutUint64(dst[120:], e.size)
}

func sectorsNeeded(n, size int) int {
	if n <= 0 {
		return 0
	}
	return (n + size - 1) / size
}

func padTo(b []byte, n int) []byte {
	if n <= 1 || len(b)%n == 0 {
		out := make([]byte, len(b))
		copy(out, b)
		return out
	}
	out := make([]byte, ((len(b)+n-1)/n)*n)
	copy(out, b)
	return out
}

// DataSpaces bytes copied from the structure Microsoft Word writes next to an
// agile encrypted package. They name the EncryptedPackage stream; they do not
// contain document content.
var dataSpaceVersion = []byte{
	0x3c, 0x00, 0x00, 0x00, 0x4d, 0x00, 0x69, 0x00, 0x63, 0x00, 0x72, 0x00, 0x6f, 0x00, 0x73, 0x00,
	0x6f, 0x00, 0x66, 0x00, 0x74, 0x00, 0x2e, 0x00, 0x43, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x74, 0x00,
	0x61, 0x00, 0x69, 0x00, 0x6e, 0x00, 0x65, 0x00, 0x72, 0x00, 0x2e, 0x00, 0x44, 0x00, 0x61, 0x00,
	0x74, 0x00, 0x61, 0x00, 0x53, 0x00, 0x70, 0x00, 0x61, 0x00, 0x63, 0x00, 0x65, 0x00, 0x73, 0x00,
	0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
}

var dataSpaceMap = []byte{
	0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x45, 0x00, 0x6e, 0x00, 0x63, 0x00, 0x72, 0x00,
	0x79, 0x00, 0x70, 0x00, 0x74, 0x00, 0x65, 0x00, 0x64, 0x00, 0x50, 0x00, 0x61, 0x00, 0x63, 0x00,
	0x6b, 0x00, 0x61, 0x00, 0x67, 0x00, 0x65, 0x00, 0x32, 0x00, 0x00, 0x00, 0x53, 0x00, 0x74, 0x00,
	0x72, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x67, 0x00, 0x45, 0x00, 0x6e, 0x00, 0x63, 0x00, 0x72, 0x00,
	0x79, 0x00, 0x70, 0x00, 0x74, 0x00, 0x69, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x44, 0x00, 0x61, 0x00,
	0x74, 0x00, 0x61, 0x00, 0x53, 0x00, 0x70, 0x00, 0x61, 0x00, 0x63, 0x00, 0x65, 0x00, 0x00, 0x00,
}

var strongEncryptionDataSpace = []byte{
	0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x32, 0x00, 0x00, 0x00, 0x53, 0x00, 0x74, 0x00,
	0x72, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x67, 0x00, 0x45, 0x00, 0x6e, 0x00, 0x63, 0x00, 0x72, 0x00,
	0x79, 0x00, 0x70, 0x00, 0x74, 0x00, 0x69, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x54, 0x00, 0x72, 0x00,
	0x61, 0x00, 0x6e, 0x00, 0x73, 0x00, 0x66, 0x00, 0x6f, 0x00, 0x72, 0x00, 0x6d, 0x00, 0x00, 0x00,
}

var dataSpacePrimary = []byte{
	0x58, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x4c, 0x00, 0x00, 0x00, 0x7b, 0x00, 0x46, 0x00,
	0x46, 0x00, 0x39, 0x00, 0x41, 0x00, 0x33, 0x00, 0x46, 0x00, 0x30, 0x00, 0x33, 0x00, 0x2d, 0x00,
	0x35, 0x00, 0x36, 0x00, 0x45, 0x00, 0x46, 0x00, 0x2d, 0x00, 0x34, 0x00, 0x36, 0x00, 0x31, 0x00,
	0x33, 0x00, 0x2d, 0x00, 0x42, 0x00, 0x44, 0x00, 0x44, 0x00, 0x35, 0x00, 0x2d, 0x00, 0x35, 0x00,
	0x41, 0x00, 0x34, 0x00, 0x31, 0x00, 0x43, 0x00, 0x31, 0x00, 0x44, 0x00, 0x30, 0x00, 0x37, 0x00,
	0x32, 0x00, 0x34, 0x00, 0x36, 0x00, 0x7d, 0x00, 0x4e, 0x00, 0x00, 0x00, 0x4d, 0x00, 0x69, 0x00,
	0x63, 0x00, 0x72, 0x00, 0x6f, 0x00, 0x73, 0x00, 0x6f, 0x00, 0x66, 0x00, 0x74, 0x00, 0x2e, 0x00,
	0x43, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x74, 0x00, 0x61, 0x00, 0x69, 0x00, 0x6e, 0x00, 0x65, 0x00,
	0x72, 0x00, 0x2e, 0x00, 0x45, 0x00, 0x6e, 0x00, 0x63, 0x00, 0x72, 0x00, 0x79, 0x00, 0x70, 0x00,
	0x74, 0x00, 0x69, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x54, 0x00, 0x72, 0x00, 0x61, 0x00, 0x6e, 0x00,
	0x73, 0x00, 0x66, 0x00, 0x6f, 0x00, 0x72, 0x00, 0x6d, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
}
