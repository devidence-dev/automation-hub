package depreport

import _ "embed" // depreport.min.css

// ReportCSS is the compiled, minified stylesheet of the report. It is built from assets/input.css
// by scripts/generate-depreport-assets.sh and committed, so the build needs no Node or Tailwind.
//
//go:embed assets/depreport.min.css
var ReportCSS string
