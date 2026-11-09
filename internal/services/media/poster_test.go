package media

import "testing"

// posterCase is one poster URL and width, with the resized poster and fallback that ResizePoster must return.
type posterCase struct {
	name, url           string
	width               int
	wantResized, wantOg string
}

func TestResizePoster(t *testing.T) {
	const amazon = "https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX300.jpg"
	const cropped = "https://m.media-amazon.com/images/M/MV5BMDg5NDFkNTctZWE2ZC00MDYzLThlYjItZTljZDY3NTcwNjA3XkEyXkFqcGc@._V1_QL75_UX380_CR0,0,380,562_.jpg"
	for _, tt := range []posterCase{
		{name: "resized", url: amazon, width: 1000, wantResized: "https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX1000.jpg", wantOg: amazon},
		{name: "cropped real URL", url: cropped, width: 1000, wantResized: "https://m.media-amazon.com/images/M/MV5BMDg5NDFkNTctZWE2ZC00MDYzLThlYjItZTljZDY3NTcwNjA3XkEyXkFqcGc@._V1_SX1000.jpg", wantOg: cropped},
		{name: "bare transform", url: "https://m.media-amazon.com/images/M/MV5BNzk@._V1_.jpg", width: 1000, wantResized: "https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX1000.jpg", wantOg: "https://m.media-amazon.com/images/M/MV5BNzk@._V1_.jpg"},
		{name: "unknown pattern", url: "https://example.com/poster.jpg", width: 1000, wantResized: "https://example.com/poster.jpg"},
		{name: "width zero", url: amazon, width: 0, wantResized: amazon},
		{name: "negative width", url: amazon, width: -1, wantResized: amazon},
		{name: "empty URL", url: "", width: 1000, wantResized: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resized, original := ResizePoster(tt.url, tt.width)
			if resized != tt.wantResized || original != tt.wantOg {
				t.Fatalf("ResizePoster(%q, %d) = %q, %q; want %q, %q", tt.url, tt.width, resized, original, tt.wantResized, tt.wantOg)
			}
		})
	}
}
