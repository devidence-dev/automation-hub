package depreport

// iconMarkup is the inner SVG of each icon of the liquid glass design (liquid-glass-report, 14px).
// The fixed strings are rendered by Icon with templ.Raw; no input reaches them.
var iconMarkup = map[string]string{
	"chevron": `<path d="M3 5l4 4 4-4" stroke-width="1.6"/>`,
	"copy":    `<rect x="4.5" y="4.5" width="7" height="7" rx="1.5"/><path d="M9.5 2.5h-5a2 2 0 0 0-2 2v5"/>`,
	"check":   `<path d="M3 7.5l2.5 2.5L11 4.5" stroke-width="1.7"/>`,
	"system":  `<circle cx="7" cy="7" r="5.25"/><path d="M7 1.75a5.25 5.25 0 0 1 0 10.5z" fill="currentColor" stroke="none"/>`,
	"light":   `<circle cx="7" cy="7" r="2.6"/><path d="M7 1v1.4M7 11.6V13M1 7h1.4M11.6 7H13M2.8 2.8l1 1M10.2 10.2l1 1M2.8 11.2l1-1M10.2 3.8l1-1"/>`,
	"dark":    `<path d="M11.8 8.6A5 5 0 0 1 5.4 2.2a5 5 0 1 0 6.4 6.4z"/>`,
}
