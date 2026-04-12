package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxImageSize = 5 * 1024 * 1024

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func SaveImage(dir string, file io.Reader, filename string, size int64) (string, error) {
	if size > MaxImageSize {
		return "", fmt.Errorf("image too large")
	}

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	}
	head = head[:n]

	contentType := http.DetectContentType(head)
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type")
	}

	if originalExt := strings.ToLower(filepath.Ext(filename)); originalExt != "" {
		for _, allowedExt := range allowedImageTypes {
			if originalExt == allowedExt || originalExt == ".jpeg" && allowedExt == ".jpg" {
				ext = originalExt
				break
			}
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, name)
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := dst.Write(head); err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/" + filepath.ToSlash(path), nil
}
