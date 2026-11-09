package media

import (
	"regexp"
	"strconv"
)

// posterWidthPattern matches the width segment of an OMDb poster URL, for example _SX300.
var posterWidthPattern = regexp.MustCompile(`_SX\d+`)

// ResizePoster replaces the width segment of url with _SX<width>. It returns the resized URL and the original
// URL as fallback. A width of 0 or less, or a URL without the segment, returns url unchanged and no fallback.
func ResizePoster(url string, width int) (resized, original string) {
	if width <= 0 || !posterWidthPattern.MatchString(url) {
		return url, ""
	}
	return posterWidthPattern.ReplaceAllLiteralString(url, "_SX"+strconv.Itoa(width)), url
}
