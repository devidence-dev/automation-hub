package depreport

import _ "embed" // assets/report.js

// reportScript holds the theme control, the tabs, the motion and the copy handler. ReportDocument
// inlines it at the end of <body>. It is linted by Biome (biome.json).
//
//go:embed assets/report.js
var reportScript string
