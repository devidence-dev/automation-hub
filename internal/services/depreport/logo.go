package depreport

import (
	_ "embed" // assets/logo.svg, assets/icon.svg
	"encoding/base64"
)

// logoSVG is the devidence logo, stripped of its metadata and inlined by Logo (the document is sent
// alone, so no image file can sit beside it). Its colors are the logo tokens of the stylesheet.
//
//go:embed assets/logo.svg
var logoSVG string

// iconSVG is the devidence icon, stripped the same way. It is the favicon.
//
//go:embed assets/icon.svg
var iconSVG string

// faviconHref is the favicon as a base64 data URI.
func faviconHref() string {
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(iconSVG))
}
