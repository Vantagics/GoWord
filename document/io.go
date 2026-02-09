package document

import "errors"

// Open reads an existing .docx file. Implementation is in the ooxml package;
// this is wired up via the goword top-level package.
var openFunc func(path string) (*Document, error)
var openFromBytesFunc func(data []byte) (*Document, error)

// Open reads a .docx file and returns a Document.
func Open(path string) (*Document, error) {
	if openFunc == nil {
		return nil, ErrNoReader
	}
	return openFunc(path)
}

// OpenFromBytes reads a .docx from byte data.
func OpenFromBytes(data []byte) (*Document, error) {
	if openFromBytesFunc == nil {
		return nil, ErrNoReader
	}
	return openFromBytesFunc(data)
}

// Save writes the document to a .docx file. Implementation is in the ooxml package.
var saveFunc func(doc *Document, path string) error
var writeToBytesFunc func(doc *Document) ([]byte, error)

// Save writes the document to a .docx file.
func (d *Document) Save(path string) error {
	if saveFunc == nil {
		return ErrNoWriter
	}
	return saveFunc(d, path)
}

// ToBytes returns the document as .docx bytes.
func (d *Document) ToBytes() ([]byte, error) {
	if writeToBytesFunc == nil {
		return nil, ErrNoWriter
	}
	return writeToBytesFunc(d)
}

// Sentinel errors for missing I/O registration.
var (
	ErrNoReader = errors.New("goword: reader not registered. Import github.com/VantageDataChat/GoWord to auto-register")
	ErrNoWriter = errors.New("goword: writer not registered. Import github.com/VantageDataChat/GoWord to auto-register")
)

// RegisterIO is called by the ooxml package to wire up read/write functions.
func RegisterIO(
	open func(string) (*Document, error),
	openBytes func([]byte) (*Document, error),
	save func(*Document, string) error,
	toBytes func(*Document) ([]byte, error),
) {
	openFunc = open
	openFromBytesFunc = openBytes
	saveFunc = save
	writeToBytesFunc = toBytes
}
