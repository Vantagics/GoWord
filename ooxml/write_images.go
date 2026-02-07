package ooxml

import (
	"fmt"
	"os"
	"strings"
)

func (w *Writer) writeImages() error {
	for _, img := range w.doc.Images() {
		var data []byte
		var ext string

		if len(img.Data) > 0 {
			data = img.Data
			ext = imageExtension(img.MimeType)
		} else if img.Source != "" {
			var err error
			data, err = os.ReadFile(img.Source)
			if err != nil {
				return fmt.Errorf("reading image %s: %w", img.Source, err)
			}
			ext = strings.TrimPrefix(strings.ToLower(getFileExt(img.Source)), ".")
			if ext == "jpg" {
				ext = "jpeg"
			}
			if img.MimeType == "" {
				img.MimeType = imageContentType(ext)
			}
		} else {
			continue
		}

		imgName := fmt.Sprintf("image%d.%s", img.ID, ext)
		img.Name = imgName
		imgPath := "word/media/" + imgName

		// Add relationship
		relID := w.addDocRel(relImage, "media/"+imgName)
		img.RelID = relID

		if err := w.addZipFile(imgPath, data); err != nil {
			return err
		}
	}
	return nil
}
