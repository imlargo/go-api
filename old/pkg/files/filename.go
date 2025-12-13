package files

import (
	"net/url"
	"path/filepath"
	"strings"
)

// ExtractFileName extracts the file name and extension from a full filename.
func ExtractFileName(path string) (filename, extension string) {

	normalized := normalizeString(path)

	// Get the base file name from the path
	base := filepath.Base(normalized)

	// Extract the extension
	ext := filepath.Ext(base)

	// Remove the extension from the name to get only the base name
	name := strings.TrimSuffix(base, ext)

	// Remove the dot from the extension if it exists
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
	}

	return name, ext
}

func normalizeString(filename string) string {
	filename = strings.TrimSpace(filename)
	filename = strings.ToLower(filename)
	filename = strings.ReplaceAll(filename, " ", "-")
	return filename
}

func ExtractFileNameFromURL(rawurl string) (string, string) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", ""
	}
	segments := strings.Split(u.Path, "/")
	lastPart := segments[len(segments)-1]
	filename, ext := ExtractFileName(lastPart)
	return filename, ext
}
