package media

import (
	"regexp"
	"strings"
)

// Release is the identifying information extracted from a torrent name.
type Release struct {
	Title, Year, Episode string
	Tags                 []string
}

// QualityTokens identifies release information that terminates a title.
var QualityTokens = []string{"2160p", "1080p", "720p", "480p", "4k", "web-dl", "webrip", "bluray", "bdrip", "hdtv", "hdr", "dv", "dual", "lat", "x264", "x265", "h264", "h265", "hevc", "aac", "dts"}

// containerExtensions lists the file extensions removed from the end of a torrent name.
var containerExtensions = []string{"mkv", "mp4", "avi", "mov", "wmv"}

var yearPattern = regexp.MustCompile(`^(19|20)\d{2}$`)
var episodePattern = regexp.MustCompile(`(?i)^s\d{2}e\d{2}$`)

func isQualityToken(part string) bool {
	for _, token := range QualityTokens {
		if strings.EqualFold(part, token) || strings.HasPrefix(strings.ToLower(part), token+"-") {
			return true
		}
	}
	return false
}

// isPartSeparator reports whether r separates the words of a torrent name.
func isPartSeparator(r rune) bool { return r == '.' || r == '_' || r == ' ' }

// stripContainerExtension drops a trailing container extension from the name parts.
func stripContainerExtension(parts []string) []string {
	if len(parts) == 0 {
		return parts
	}
	last := strings.ToLower(parts[len(parts)-1])
	for _, ext := range containerExtensions {
		if last == ext {
			return parts[:len(parts)-1]
		}
	}
	return parts
}

// isReleaseYear reports whether part is the release year after the title words.
func isReleaseYear(part string, title, later []string) bool {
	if !yearPattern.MatchString(part) || len(title) == 0 {
		return false
	}
	// A number followed by title words is part of the title if another year follows.
	for _, next := range later {
		if yearPattern.MatchString(next) {
			return false
		}
		if episodePattern.MatchString(next) || isQualityToken(next) {
			break
		}
	}
	return true
}

// releaseTags splits one release part into tags. A web-dl part keeps its "web-dl" tag whole.
func releaseTags(part string) []string {
	if strings.HasPrefix(strings.ToLower(part), "web-dl") {
		tags := []string{part[:6]}
		if len(part) > 6 {
			tags = append(tags, strings.Split(strings.TrimPrefix(part[6:], "-"), "-")...)
		}
		return tags
	}
	return strings.Split(part, "-")
}

// addReleaseField stores a part that follows the title: the first episode code, or else tags.
func addReleaseField(result *Release, part string) {
	if episodePattern.MatchString(part) && result.Episode == "" {
		result.Episode = strings.ToUpper(part)
		return
	}
	result.Tags = append(result.Tags, releaseTags(part)...)
}

// ParseRelease extracts a conservative title and ordered release tags.
func ParseRelease(name string) Release {
	parts := stripContainerExtension(strings.FieldsFunc(name, isPartSeparator))
	var result Release
	var title []string
	ended := false
	for i, part := range parts {
		switch {
		case ended:
			addReleaseField(&result, part)
		case isReleaseYear(part, title, parts[i+1:]):
			result.Year = part
			ended = true
		case episodePattern.MatchString(part):
			result.Episode = strings.ToUpper(part)
			ended = true
		case isQualityToken(part):
			ended = true
			addReleaseField(&result, part)
		default:
			title = append(title, part)
		}
	}
	result.Title = strings.Join(title, " ")
	return result
}
