package document

import (
	"crypto/rand"
	"fmt"
)

// EmbeddedFont is a TrueType font stored obfuscated in the package, as
// ECMA-376 requires for embedRegular. Data is the clear font bytes. Key is
// the fontKey GUID, including braces.
type EmbeddedFont struct {
	Name string
	Data []byte
	Key  string
	// Stored is true when the font was loaded from the package and does not
	// need to be written again.
	Stored bool
}

// EmbedFont stores a font and assigns a fontKey. The package writer obfuscates
// the first 32 bytes with that key.
func (d *Document) EmbedFont(name string, ttf []byte) (*EmbeddedFont, error) {
	if d == nil {
		return nil, fmt.Errorf("nil document")
	}
	if len(ttf) < 32 {
		return nil, fmt.Errorf("font data is shorter than the 32-byte obfuscation window")
	}
	key, err := newFontKey()
	if err != nil {
		return nil, err
	}
	cp := make([]byte, len(ttf))
	copy(cp, ttf)
	ef := EmbeddedFont{Name: name, Data: cp, Key: key}
	d.embeddedFonts = append(d.embeddedFonts, ef)
	return &d.embeddedFonts[len(d.embeddedFonts)-1], nil
}

// EmbeddedFonts returns fonts embedded in this document.
func (d *Document) EmbeddedFonts() []EmbeddedFont {
	if d == nil {
		return nil
	}
	return d.embeddedFonts
}

// AddEmbeddedFont records a font loaded from a package. Data is clear.
func (d *Document) AddEmbeddedFont(ef EmbeddedFont) {
	if d == nil {
		return
	}
	ef.Stored = true
	d.embeddedFonts = append(d.embeddedFonts, ef)
}

func newFontKey() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("{%02X%02X%02X%02X-%02X%02X-%02X%02X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15]), nil
}

// ObfuscateFont XORs the first 32 bytes with the fontKey GUID, per ECMA-376.
// Applying it twice restores the original bytes.
func ObfuscateFont(fontKey string, data []byte) ([]byte, error) {
	key, err := fontKeyBytes(fontKey)
	if err != nil {
		return nil, err
	}
	if len(data) < 32 {
		return nil, fmt.Errorf("font data is shorter than the 32-byte obfuscation window")
	}
	out := make([]byte, len(data))
	copy(out, data)
	for i := 0; i < 32; i++ {
		out[i] ^= key[i%16]
	}
	return out, nil
}

// fontKeyBytes is the .NET Guid.ToByteArray order, reversed, which is the
// key Office uses for the XOR.
func fontKeyBytes(fontKey string) ([]byte, error) {
	s := fontKey
	s = trimBraces(s)
	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return nil, fmt.Errorf("fontKey %q is not a GUID", fontKey)
	}
	raw, err := parseHex(s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:])
	if err != nil {
		return nil, err
	}
	// Mixed-endian GUID, then reversed.
	guid := []byte{raw[3], raw[2], raw[1], raw[0], raw[5], raw[4], raw[7], raw[6]}
	guid = append(guid, raw[8:]...)
	for i, j := 0, len(guid)-1; i < j; i, j = i+1, j-1 {
		guid[i], guid[j] = guid[j], guid[i]
	}
	return guid, nil
}

func trimBraces(s string) string {
	if len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}' {
		return s[1 : len(s)-1]
	}
	return s
}

func parseHex(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("odd hex length")
	}
	out := make([]byte, len(s)/2)
	for i := 0; i < len(out); i++ {
		n, err := parseHexByte(s[i*2], s[i*2+1])
		if err != nil {
			return nil, err
		}
		out[i] = n
	}
	return out, nil
}

func parseHexByte(a, b byte) (byte, error) {
	hi, err := hexVal(a)
	if err != nil {
		return 0, err
	}
	lo, err := hexVal(b)
	if err != nil {
		return 0, err
	}
	return hi<<4 | lo, nil
}

func hexVal(c byte) (byte, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	default:
		return 0, fmt.Errorf("bad hex digit %q", c)
	}
}
