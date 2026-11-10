package ooxml

import (
	"bytes"
	"encoding/binary"
)

type ttfRec struct {
	tag    string
	data   []byte
	offset int
	length int
	sum    uint32
}

func buildMinimalTTF() []byte {
	tables := []struct {
		tag  string
		data []byte
	}{
		{"cmap", cmapTable()},
		{"glyf", []byte{}},
		{"head", headTable()},
		{"hhea", hheaTable()},
		{"hmtx", hmtxTable()},
		{"loca", []byte{0, 0, 0, 0}},
		{"maxp", maxpTable()},
		{"name", nameTable("GoWord Test")},
		{"post", postTable()},
	}
	// Place tables after the offset table and directory.
	n := len(tables)
	off := 12 + 16*n
	recs := make([]ttfRec, n)
	var body []byte
	for i, t := range tables {
		pad := (4 - len(t.data)%4) % 4
		recs[i] = ttfRec{tag: t.tag, data: append(append([]byte{}, t.data...), make([]byte, pad)...), offset: off, length: len(t.data)}
		recs[i].sum = checksum(recs[i].data)
		body = append(body, recs[i].data...)
		off += len(recs[i].data)
	}
	font := assembleTTF(recs, body, 0)
	sum := checksum(font)
	adj := 0xB1B0AFBA - sum
	// Patch head checksumAdjustment at offset 8 of the head table.
	for i := range recs {
		if recs[i].tag == "head" {
			binary.BigEndian.PutUint32(recs[i].data[8:], adj)
			recs[i].sum = checksum(recs[i].data)
		}
	}
	return assembleTTF(recs, nil, adj)
}

func assembleTTF(recs []ttfRec, _ []byte, _ uint32) []byte {
	n := len(recs)
	// Recompute offsets.
	off := 12 + 16*n
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(0x00010000))
	binary.Write(&buf, binary.BigEndian, uint16(n))
	sr := uint16(1)
	es := uint16(0)
	for (sr << 1) <= uint16(n) {
		sr <<= 1
		es++
	}
	sr *= 16
	binary.Write(&buf, binary.BigEndian, sr)
	binary.Write(&buf, binary.BigEndian, es)
	binary.Write(&buf, binary.BigEndian, uint16(n*16)-sr)
	var body bytes.Buffer
	for i := range recs {
		recs[i].offset = off
		binary.Write(&buf, binary.BigEndian, []byte(recs[i].tag))
		binary.Write(&buf, binary.BigEndian, recs[i].sum)
		binary.Write(&buf, binary.BigEndian, uint32(recs[i].offset))
		binary.Write(&buf, binary.BigEndian, uint32(recs[i].length))
		body.Write(recs[i].data)
		off += len(recs[i].data)
	}
	buf.Write(body.Bytes())
	return buf.Bytes()
}

func checksum(b []byte) uint32 {
	var sum uint32
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	for i := 0; i < len(b); i += 4 {
		sum += binary.BigEndian.Uint32(b[i:])
	}
	return sum
}

func headTable() []byte {
	b := make([]byte, 54)
	binary.BigEndian.PutUint32(b[0:], 0x00010000)
	binary.BigEndian.PutUint32(b[4:], 0x00010000)
	binary.BigEndian.PutUint32(b[12:], 0x5F0F3CF5)
	binary.BigEndian.PutUint16(b[18:], 1000)
	binary.BigEndian.PutUint16(b[44:], 8)
	binary.BigEndian.PutUint16(b[46:], 2)
	return b
}

func hheaTable() []byte {
	b := make([]byte, 36)
	binary.BigEndian.PutUint32(b[0:], 0x00010000)
	binary.BigEndian.PutUint16(b[4:], 800)
	binary.BigEndian.PutUint16(b[34:], 1) // numberOfHMetrics
	return b
}

func maxpTable() []byte {
	b := make([]byte, 32)
	binary.BigEndian.PutUint32(b[0:], 0x00010000)
	binary.BigEndian.PutUint16(b[4:], 1)
	return b
}

func hmtxTable() []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[0:], 500)
	return b
}

func postTable() []byte {
	b := make([]byte, 32)
	binary.BigEndian.PutUint32(b[0:], 0x00030000)
	return b
}

func cmapTable() []byte {
	// format 4, two segments: map U+0000 to glyph 0, then the 0xFFFF sentinel.
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, uint16(0)) // version
	binary.Write(&b, binary.BigEndian, uint16(1)) // numTables
	binary.Write(&b, binary.BigEndian, uint16(3)) // platform
	binary.Write(&b, binary.BigEndian, uint16(1)) // encoding
	binary.Write(&b, binary.BigEndian, uint32(12))
	binary.Write(&b, binary.BigEndian, uint16(4))
	lengthPos := b.Len()
	binary.Write(&b, binary.BigEndian, uint16(0))
	binary.Write(&b, binary.BigEndian, uint16(0))
	binary.Write(&b, binary.BigEndian, uint16(4)) // segCountX2
	binary.Write(&b, binary.BigEndian, uint16(4)) // searchRange
	binary.Write(&b, binary.BigEndian, uint16(1)) // entrySelector
	binary.Write(&b, binary.BigEndian, uint16(0)) // rangeShift
	binary.Write(&b, binary.BigEndian, uint16(0))
	binary.Write(&b, binary.BigEndian, uint16(0xFFFF))
	binary.Write(&b, binary.BigEndian, uint16(0)) // reserved
	binary.Write(&b, binary.BigEndian, uint16(0))
	binary.Write(&b, binary.BigEndian, uint16(0xFFFF))
	binary.Write(&b, binary.BigEndian, uint16(0))
	binary.Write(&b, binary.BigEndian, uint16(1))
	binary.Write(&b, binary.BigEndian, uint16(0))
	binary.Write(&b, binary.BigEndian, uint16(0))
	out := b.Bytes()
	binary.BigEndian.PutUint16(out[lengthPos:], uint16(len(out)-12))
	return out
}

func nameTable(family string) []byte {
	rec := make([]byte, 12)
	binary.BigEndian.PutUint16(rec[0:], 1) // platform Mac
	binary.BigEndian.PutUint16(rec[6:], 1) // nameID family
	str := []byte(family)
	binary.BigEndian.PutUint16(rec[8:], uint16(len(str)))
	binary.BigEndian.PutUint16(rec[10:], 0)
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, uint16(0))
	binary.Write(&b, binary.BigEndian, uint16(1))
	binary.Write(&b, binary.BigEndian, uint16(6+len(rec)))
	b.Write(rec)
	b.Write(str)
	return b.Bytes()
}
