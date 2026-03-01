package export

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Options configures per-export behaviour. Pass a zero value for defaults.
type Options struct {
	// Credits is displayed left-aligned in the export footer alongside the
	// version attribution. Use it to claim recipe authorship.
	Credits string
}

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

// SafeFilename returns a URL-safe slug from a recipe name (no extension).
func SafeFilename(name string) string {
	s := strings.ToLower(name)
	s = nonAlphanumRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// UniqueFilePath returns a path in dir for base.ext that does not already
// exist. If base.ext is taken it tries base-2.ext, base-3.ext, and so on.
func UniqueFilePath(dir, base, ext string) string {
	candidate := filepath.Join(dir, base+"."+ext)
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}
	for n := 2; ; n++ {
		candidate = filepath.Join(dir, base+"-"+strconv.Itoa(n)+"."+ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// DownloadsDir returns ~/Downloads, creating it if needed.
func DownloadsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Downloads")
	return dir, os.MkdirAll(dir, 0o755)
}

// TagContextLabel returns a human-readable label for a tag context.
func TagContextLabel(ctx string) string {
	switch ctx {
	case "courses":
		return "Courses"
	case "cooking_methods":
		return "Cooking methods"
	case "cultural_influences":
		return "Cultural influences"
	case "dietary_restrictions":
		return "Dietary"
	default:
		return ctx
	}
}

// FormatMins formats a minute count as a human-readable duration string.
func FormatMins(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	h := minutes / 60
	m := minutes % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}
