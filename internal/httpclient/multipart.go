package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type MultipartPart struct {
	Name     string
	Value    string
	FilePath string
}

func BuildMultipart(parts []MultipartPart) (string, string, error) {
	if len(parts) == 0 {
		return "", "", fmt.Errorf("no multipart parts")
	}
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	for _, p := range parts {
		if p.Name == "" {
			continue
		}
		if p.FilePath != "" {
			if err := writeFilePart(w, p.Name, p.FilePath); err != nil {
				return "", "", err
			}
			continue
		}
		if err := w.WriteField(p.Name, p.Value); err != nil {
			return "", "", err
		}
	}
	if err := w.Close(); err != nil {
		return "", "", err
	}
	return buf.String(), w.FormDataContentType(), nil
}

func writeFilePart(w *multipart.Writer, fieldName, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	part, err := w.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	return err
}
