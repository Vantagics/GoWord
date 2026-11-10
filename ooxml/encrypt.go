package ooxml

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"hash"
	"io"
	"unicode/utf16"

	"github.com/VantageDataChat/GoWord/document"
)

const (
	agileSpinCount = 100000
	agileKeyBits   = 256
	agileSaltSize  = 16
	agileBlockSize = 16
	agileHashName  = "SHA512"
	agileHashSize  = 64
	agileSegment   = 4096
	standardSpin   = 50000
)

var (
	blockVerifierInput = []byte{0xfe, 0xa7, 0xd2, 0x76, 0x3b, 0x4b, 0x9e, 0x79}
	blockVerifierValue = []byte{0xd7, 0xaa, 0x0f, 0x6d, 0x30, 0x61, 0x34, 0x4e}
	blockKeyValue      = []byte{0x14, 0x6e, 0x0b, 0xe7, 0xab, 0xac, 0xd0, 0xd6}
	blockHmacKey       = []byte{0x5f, 0xb2, 0xad, 0x01, 0x0c, 0xb9, 0xe1, 0xf6}
	blockHmacValue     = []byte{0xa0, 0x67, 0x7f, 0x02, 0xb2, 0x2c, 0x84, 0x33}
)

func decryptOfficePackage(data []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, document.ErrEncrypted
	}
	streams, err := readCompoundFile(data)
	if err != nil {
		return nil, err
	}
	info := streams["EncryptionInfo"]
	pkg := streams["EncryptedPackage"]
	if len(info) < 8 || len(pkg) < 8 {
		return nil, fmt.Errorf("encrypted package is missing EncryptionInfo")
	}
	major := binary.LittleEndian.Uint16(info[0:])
	minor := binary.LittleEndian.Uint16(info[2:])
	switch {
	case major == 4 && minor == 4:
		return decryptAgile(info[8:], pkg, password)
	case (major == 4 || major == 3) && minor == 2:
		return decryptStandard(info, pkg, password)
	default:
		return nil, fmt.Errorf("unsupported Office encryption version %d.%d", major, minor)
	}
}

func encryptOfficePackage(plain []byte, password string) ([]byte, error) {
	if password == "" {
		return plain, nil
	}
	info, pkg, err := encryptAgile(plain, password)
	if err != nil {
		return nil, err
	}
	return writeOfficePackage(info, pkg)
}

type agileKey struct {
	spin        int
	saltSize    int
	blockSize   int
	keyBits     int
	hashSize    int
	hashName    string
	salt        []byte
	encVerifier []byte
	encHash     []byte
	encKey      []byte
}

type agileInfo struct {
	keySalt     []byte
	keyHash     string
	keyBlock    int
	keyBits     int
	hmacKey     []byte
	hmacValue   []byte
	passwordKey agileKey
}

func decryptAgile(xmlBytes, pkg []byte, password string) ([]byte, error) {
	info, err := parseAgile(xmlBytes)
	if err != nil {
		return nil, err
	}
	if info.passwordKey.hashName == "" {
		return nil, fmt.Errorf("encrypted package has no password key")
	}
	secret, ok := agileSecret(password, info.passwordKey)
	if !ok {
		return nil, document.ErrWrongPassword
	}
	plain, err := agileDecryptPayload(pkg, secret, info.keySalt, info.keyHash, info.keyBits)
	if err != nil {
		return nil, err
	}
	if len(info.hmacKey) > 0 && len(info.hmacValue) > 0 {
		if !agileHMACOK(pkg, secret, info) {
			return nil, fmt.Errorf("encrypted package failed the integrity check")
		}
	}
	return plain, nil
}

func agileSecret(password string, k agileKey) ([]byte, bool) {
	h := agileIterated(password, k.salt, k.spin, k.hashName)
	k1 := agileBlockKey(h, blockVerifierInput, k.hashName, k.keyBits)
	k2 := agileBlockKey(h, blockVerifierValue, k.hashName, k.keyBits)
	k3 := agileBlockKey(h, blockKeyValue, k.hashName, k.keyBits)
	iv := fitIV(k.salt, k.blockSize)
	verifier, err := aesCBC(k.encVerifier, k1, iv, false)
	if err != nil {
		return nil, false
	}
	got := hashSum(k.hashName, verifier)
	encHash, err := aesCBC(k.encHash, k2, iv, false)
	if err != nil || len(encHash) < len(got) || !hmac.Equal(got, encHash[:len(got)]) {
		return nil, false
	}
	secret, err := aesCBC(k.encKey, k3, iv, false)
	if err != nil || len(secret) < k.keyBits/8 {
		return nil, false
	}
	return secret[:k.keyBits/8], true
}

func agileDecryptPayload(pkg, secret, salt []byte, hashName string, keyBits int) ([]byte, error) {
	if len(pkg) < 8 {
		return nil, fmt.Errorf("encrypted package is truncated")
	}
	total := binary.LittleEndian.Uint64(pkg[:8])
	if total > uint64(len(pkg)-8) || total > 1<<32 {
		return nil, fmt.Errorf("encrypted package size is invalid")
	}
	key := secret
	if len(key) > keyBits/8 && keyBits > 0 {
		key = key[:keyBits/8]
	}
	out := make([]byte, 0, int(total))
	rest := pkg[8:]
	remaining := int(total)
	for i := 0; remaining > 0 && len(rest) > 0; i++ {
		n := agileSegment
		if n > len(rest) {
			n = len(rest)
		}
		chunk := rest[:n]
		rest = rest[n:]
		if len(chunk)%aes.BlockSize != 0 {
			return nil, fmt.Errorf("encrypted segment is not block aligned")
		}
		iv := hashSum(hashName, append(append([]byte{}, salt...), putU32(uint32(i))...))
		iv = iv[:aes.BlockSize]
		dec, err := aesCBC(chunk, key, iv, false)
		if err != nil {
			return nil, err
		}
		if remaining < len(dec) {
			dec = dec[:remaining]
		}
		out = append(out, dec...)
		remaining -= len(dec)
	}
	if remaining != 0 {
		return nil, fmt.Errorf("encrypted package ended early")
	}
	return out, nil
}

func agileHMACOK(pkg, secret []byte, info agileInfo) bool {
	hashName := info.keyHash
	size := hashLen(hashName)
	iv1 := hashSum(hashName, append(append([]byte{}, info.keySalt...), blockHmacKey...))[:info.keyBlock]
	iv2 := hashSum(hashName, append(append([]byte{}, info.keySalt...), blockHmacValue...))[:info.keyBlock]
	hmacKey, err := aesCBC(info.hmacKey, secret, iv1, false)
	if err != nil || len(hmacKey) < size {
		return false
	}
	want, err := aesCBC(info.hmacValue, secret, iv2, false)
	if err != nil || len(want) < size {
		return false
	}
	mac := hmac.New(hashNew(hashName), hmacKey[:size])
	mac.Write(pkg)
	return hmac.Equal(mac.Sum(nil), want[:size])
}

func encryptAgile(plain []byte, password string) (infoXML, encrypted []byte, err error) {
	keySalt := random(agileSaltSize)
	passSalt := random(agileSaltSize)
	verifier := random(agileSaltSize)
	secret := random(agileKeyBits / 8)
	h := agileIterated(password, passSalt, agileSpinCount, agileHashName)
	k1 := agileBlockKey(h, blockVerifierInput, agileHashName, agileKeyBits)
	k2 := agileBlockKey(h, blockVerifierValue, agileHashName, agileKeyBits)
	k3 := agileBlockKey(h, blockKeyValue, agileHashName, agileKeyBits)
	iv := fitIV(passSalt, agileBlockSize)
	encVerifier, err := aesCBC(verifier, k1, iv, true)
	if err != nil {
		return nil, nil, err
	}
	hashed := hashSum(agileHashName, verifier)
	encHash, err := aesCBC(hashed, k2, iv, true)
	if err != nil {
		return nil, nil, err
	}
	encSecret, err := aesCBC(secret, k3, iv, true)
	if err != nil {
		return nil, nil, err
	}
	pkg := encryptAgilePayload(plain, secret, keySalt, agileHashName)
	hmacSalt := random(agileHashSize)
	iv1 := hashSum(agileHashName, append(append([]byte{}, keySalt...), blockHmacKey...))[:agileBlockSize]
	iv2 := hashSum(agileHashName, append(append([]byte{}, keySalt...), blockHmacValue...))[:agileBlockSize]
	encHmacKey, err := aesCBC(hmacSalt, secret, iv1, true)
	if err != nil {
		return nil, nil, err
	}
	mac := hmac.New(sha512.New, hmacSalt)
	mac.Write(pkg)
	encHmacValue, err := aesCBC(mac.Sum(nil), secret, iv2, true)
	if err != nil {
		return nil, nil, err
	}
	xmlText := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+"\r\n"+
		`<encryption xmlns="http://schemas.microsoft.com/office/2006/encryption" xmlns:p="http://schemas.microsoft.com/office/2006/keyEncryptor/password">`+
		`<keyData saltSize="%d" blockSize="%d" keyBits="%d" hashSize="%d" cipherAlgorithm="AES" cipherChaining="ChainingModeCBC" hashAlgorithm="%s" saltValue="%s"/>`+
		`<dataIntegrity encryptedHmacKey="%s" encryptedHmacValue="%s"/>`+
		`<keyEncryptors><keyEncryptor uri="http://schemas.microsoft.com/office/2006/keyEncryptor/password">`+
		`<p:encryptedKey spinCount="%d" saltSize="%d" blockSize="%d" keyBits="%d" hashSize="%d" cipherAlgorithm="AES" cipherChaining="ChainingModeCBC" hashAlgorithm="%s" saltValue="%s" encryptedVerifierHashInput="%s" encryptedVerifierHashValue="%s" encryptedKeyValue="%s"/>`+
		`</keyEncryptor></keyEncryptors></encryption>`,
		agileSaltSize, agileBlockSize, agileKeyBits, agileHashSize, agileHashName, b64(keySalt),
		b64(encHmacKey), b64(encHmacValue),
		agileSpinCount, agileSaltSize, agileBlockSize, agileKeyBits, agileHashSize, agileHashName, b64(passSalt),
		b64(encVerifier), b64(encHash), b64(encSecret),
	)
	header := make([]byte, 8)
	binary.LittleEndian.PutUint16(header[0:], 4)
	binary.LittleEndian.PutUint16(header[2:], 4)
	binary.LittleEndian.PutUint32(header[4:], 0x40)
	return append(header, []byte(xmlText)...), pkg, nil
}

func encryptAgilePayload(plain, secret, salt []byte, hashName string) []byte {
	out := make([]byte, 8+len(plain)+agileBlockSize)
	binary.LittleEndian.PutUint64(out[:8], uint64(len(plain)))
	n := 8
	for i, off := 0, 0; off < len(plain); i++ {
		end := off + agileSegment
		if end > len(plain) {
			end = len(plain)
		}
		chunk := zeroPad(plain[off:end], agileBlockSize)
		iv := hashSum(hashName, append(append([]byte{}, salt...), putU32(uint32(i))...))[:agileBlockSize]
		enc, err := aesCBC(chunk, secret, iv, true)
		if err != nil {
			return nil
		}
		copy(out[n:], enc)
		n += len(enc)
		off = end
	}
	return out[:n]
}

func parseAgile(raw []byte) (agileInfo, error) {
	var info agileInfo
	dec := xml.NewDecoder(bytes.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return info, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "keyData":
			info.keyHash = attrVal(se, "hashAlgorithm")
			info.keyBlock = atoi(attrVal(se, "blockSize"))
			info.keyBits = atoi(attrVal(se, "keyBits"))
			info.keySalt, err = base64.StdEncoding.DecodeString(attrVal(se, "saltValue"))
		case "dataIntegrity":
			info.hmacKey, err = base64.StdEncoding.DecodeString(attrVal(se, "encryptedHmacKey"))
			if err != nil {
				return info, err
			}
			info.hmacValue, err = base64.StdEncoding.DecodeString(attrVal(se, "encryptedHmacValue"))
		case "encryptedKey":
			k := agileKey{
				spin:      atoi(attrVal(se, "spinCount")),
				saltSize:  atoi(attrVal(se, "saltSize")),
				blockSize: atoi(attrVal(se, "blockSize")),
				keyBits:   atoi(attrVal(se, "keyBits")),
				hashSize:  atoi(attrVal(se, "hashSize")),
				hashName:  attrVal(se, "hashAlgorithm"),
			}
			k.salt, err = base64.StdEncoding.DecodeString(attrVal(se, "saltValue"))
			if err != nil {
				return info, err
			}
			k.encVerifier, err = base64.StdEncoding.DecodeString(attrVal(se, "encryptedVerifierHashInput"))
			if err != nil {
				return info, err
			}
			k.encHash, err = base64.StdEncoding.DecodeString(attrVal(se, "encryptedVerifierHashValue"))
			if err != nil {
				return info, err
			}
			k.encKey, err = base64.StdEncoding.DecodeString(attrVal(se, "encryptedKeyValue"))
			info.passwordKey = k
		}
		if err != nil {
			return info, err
		}
	}
	if info.keyBlock == 0 {
		info.keyBlock = agileBlockSize
	}
	return info, nil
}

func decryptStandard(info, pkg []byte, password string) ([]byte, error) {
	if len(info) < 12 {
		return nil, fmt.Errorf("standard encryption header is truncated")
	}
	headerSize := int(binary.LittleEndian.Uint32(info[8:]))
	off := 12
	if off+headerSize > len(info) {
		return nil, fmt.Errorf("standard encryption header is truncated")
	}
	header := info[off : off+headerSize]
	off += headerSize
	if len(header) < 24 {
		return nil, fmt.Errorf("standard encryption header is truncated")
	}
	alg := binary.LittleEndian.Uint32(header[8:])
	keyBits := int(binary.LittleEndian.Uint32(header[16:]))
	switch alg {
	case 0x660E:
		if keyBits == 0 {
			keyBits = 128
		}
	case 0x660F:
		if keyBits == 0 {
			keyBits = 192
		}
	case 0x6610:
		if keyBits == 0 {
			keyBits = 256
		}
	default:
		return nil, fmt.Errorf("unsupported standard cipher 0x%x", alg)
	}
	if off+4 > len(info) {
		return nil, fmt.Errorf("standard encryption verifier is truncated")
	}
	saltSize := int(binary.LittleEndian.Uint32(info[off:]))
	off += 4
	if saltSize <= 0 || off+saltSize+16+4 > len(info) {
		return nil, fmt.Errorf("standard encryption verifier is truncated")
	}
	salt := info[off : off+saltSize]
	off += saltSize
	encVerifier := info[off : off+16]
	off += 16
	off += 4 // verifier hash size
	encHash := info[off:]
	key := standardKey(password, salt, standardSpin, []byte{0, 0, 0, 0}, keyBits/8)
	verifier, err := aesECB(encVerifier, key, false)
	if err != nil {
		return nil, err
	}
	got := sha1.Sum(verifier)
	hashPlain, err := aesECB(encHash, key, false)
	if err != nil || len(hashPlain) < len(got) || !hmac.Equal(got[:], hashPlain[:len(got)]) {
		return nil, document.ErrWrongPassword
	}
	if len(pkg) < 8 {
		return nil, fmt.Errorf("encrypted package is truncated")
	}
	total := int(binary.LittleEndian.Uint64(pkg[:8]))
	if total < 0 || total > len(pkg)-8 {
		return nil, fmt.Errorf("encrypted package size is invalid")
	}
	plain, err := aesECB(pkg[8:], key, false)
	if err != nil {
		return nil, err
	}
	if len(plain) < total {
		return nil, fmt.Errorf("encrypted package ended early")
	}
	return plain[:total], nil
}

func standardKey(password string, salt []byte, spin int, block []byte, keyLen int) []byte {
	h := sha1Sum(salt, utf16LE(password))
	for i := 0; i < spin; i++ {
		h = sha1Sum(putU32(uint32(i)), h)
	}
	h = sha1Sum(h, block)
	buf1 := bytes.Repeat([]byte{0x36}, 64)
	buf2 := bytes.Repeat([]byte{0x5c}, 64)
	for i, b := range h {
		buf1[i] ^= b
		buf2[i] ^= b
	}
	out := append(sha1Sum(buf1), sha1Sum(buf2)...)
	if keyLen > len(out) {
		keyLen = len(out)
	}
	return out[:keyLen]
}

func agileIterated(password string, salt []byte, spin int, hashName string) []byte {
	h := hashSum(hashName, append(append([]byte{}, salt...), utf16LE(password)...))
	for i := 0; i < spin; i++ {
		h = hashSum(hashName, append(putU32(uint32(i)), h...))
	}
	return h
}

func agileBlockKey(iterated, block []byte, hashName string, keyBits int) []byte {
	h := hashSum(hashName, append(append([]byte{}, iterated...), block...))
	need := keyBits / 8
	if len(h) >= need {
		return h[:need]
	}
	out := make([]byte, need)
	copy(out, h)
	for i := len(h); i < need; i++ {
		out[i] = 0x36
	}
	return out
}

func fitIV(salt []byte, block int) []byte {
	if block <= 0 {
		block = aes.BlockSize
	}
	if len(salt) == block {
		return salt
	}
	out := make([]byte, block)
	for i := range out {
		out[i] = 0x36
	}
	copy(out, salt)
	return out
}

func aesCBC(data, key, iv []byte, encrypt bool) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() || len(data) == 0 || len(data)%block.BlockSize() != 0 {
		return nil, errors.New("aes-cbc input is not block aligned")
	}
	out := make([]byte, len(data))
	var mode cipher.BlockMode
	if encrypt {
		mode = cipher.NewCBCEncrypter(block, iv)
	} else {
		mode = cipher.NewCBCDecrypter(block, iv)
	}
	mode.CryptBlocks(out, data)
	return out, nil
}

func aesECB(data, key []byte, encrypt bool) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data)%block.BlockSize() != 0 {
		return nil, errors.New("aes-ecb input is not block aligned")
	}
	out := make([]byte, len(data))
	for i := 0; i < len(data); i += block.BlockSize() {
		if encrypt {
			block.Encrypt(out[i:], data[i:])
		} else {
			block.Decrypt(out[i:], data[i:])
		}
	}
	return out, nil
}

func hashNew(name string) func() hash.Hash {
	switch name {
	case "SHA256", "SHA-256":
		return sha256.New
	case "SHA384", "SHA-384":
		return sha512.New384
	case "SHA512", "SHA-512":
		return sha512.New
	default:
		return sha1.New
	}
}

func hashSum(name string, data []byte) []byte {
	h := hashNew(name)()
	h.Write(data)
	return h.Sum(nil)
}

func hashLen(name string) int {
	return hashNew(name)().Size()
}

func sha1Sum(parts ...[]byte) []byte {
	h := sha1.New()
	for _, p := range parts {
		h.Write(p)
	}
	return h.Sum(nil)
}

func utf16LE(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2)
	for i, r := range u {
		binary.LittleEndian.PutUint16(b[i*2:], r)
	}
	return b
}

func putU32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func zeroPad(b []byte, n int) []byte {
	if len(b)%n == 0 {
		return b
	}
	out := make([]byte, ((len(b)+n-1)/n)*n)
	copy(out, b)
	return out
}

func random(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func attrVal(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
