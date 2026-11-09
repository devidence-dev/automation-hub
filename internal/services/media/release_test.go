package media

import (
	"reflect"
	"testing"
)

func TestParseRelease(t *testing.T) {
	tests := []struct {
		name string
		want Release
	}{
		{"Coyote.Vs.Acme.2026.WEB-DL.4k.HDR-Dual-Lat", Release{Title: "Coyote Vs Acme", Year: "2026", Tags: []string{"WEB-DL", "4k", "HDR", "Dual", "Lat"}}},
		{"The.Bear.S03E02.1080p.WEB-DL.mkv", Release{Title: "The Bear", Episode: "S03E02", Tags: []string{"1080p", "WEB-DL"}}},
		{"2001.A.Space.Odyssey.1968.1080p.BluRay", Release{Title: "2001 A Space Odyssey", Year: "1968", Tags: []string{"1080p", "BluRay"}}},
		{"2160p.WEB-DL", Release{Tags: []string{"2160p", "WEB-DL"}}},
		{"The_Bear_2022_1080p", Release{Title: "The Bear", Year: "2022", Tags: []string{"1080p"}}},
		{"Blade.Runner.2049.2017.2160p", Release{Title: "Blade Runner 2049", Year: "2017", Tags: []string{"2160p"}}},
		{"Dune.2021.1080p.REMASTERED.2024.mkv", Release{Title: "Dune", Year: "2021", Tags: []string{"1080p", "REMASTERED", "2024"}}},
		{"Dune.2021.HDR-Dual-Lat.2024.mkv", Release{Title: "Dune", Year: "2021", Tags: []string{"HDR", "Dual", "Lat", "2024"}}},
		{"Dune.2021.WEB-DL.2024.mkv", Release{Title: "Dune", Year: "2021", Tags: []string{"WEB-DL", "2024"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseRelease(tt.name); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
