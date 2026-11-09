package media

import (
	"regexp"
	"strconv"
)

// posterTransformPattern matches an OMDb poster URL from "._V1_" to the file extension, capturing the prefix and the
// extension. The transform segment between them can be _SX300, _QL75_UX380_CR0,0,380,562_, or empty.
var posterTransformPattern = regexp.MustCompile(`(\._V1_)[^/]*?(\.[A-Za-z0-9]+)$`)

// ResizePoster replaces the image transform segment of url (everything between "._V1_" and the file extension) with
// SX<width>. It returns the resized URL and the original URL as fallback. A width of 0 or less, or a URL without
// "._V1_", returns url unchanged and no fallback.
func ResizePoster(url string, width int) (resized, original string) {
	if width <= 0 || !posterTransformPattern.MatchString(url) {
		return url, ""
	}
	return posterTransformPattern.ReplaceAllString(url, "${1}SX"+strconv.Itoa(width)+"${2}"), url
}
