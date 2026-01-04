package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ParseFile extracts text content from a file based on its extension.
// Supports .pdf and .md (and .txt).
func ParseFile(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return parsePDF(filePath)
	case ".md", ".txt":
		return parseText(filePath)
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func parseText(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

func parsePDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open pdf: %w", err)
	}
	defer f.Close()

	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to get plain text from pdf: %w", err)
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(b)
	if err != nil {
		return "", fmt.Errorf("failed to read pdf buffer: %w", err)
	}

	return buf.String(), nil
}
