package depreport

import "testing"

func TestParseUpdateType(t *testing.T) {
	tests := []struct {
		in       string
		want     UpdateType
		wantList bool
	}{
		{in: "major", want: UpdateMajor, wantList: true},
		{in: "minor", want: UpdateMinor, wantList: true},
		{in: "patch", want: UpdatePatch, wantList: true},
		{in: "digest", want: UpdateUnknown},
		{in: "pin", want: UpdateUnknown},
		{in: "", want: UpdateUnknown},
	}
	for _, tt := range tests {
		got := ParseUpdateType(tt.in)
		if got != tt.want || got.Listed() != tt.wantList {
			t.Errorf("ParseUpdateType(%q) = %v (listed %v), want %v (listed %v)", tt.in, got, got.Listed(), tt.want, tt.wantList)
		}
	}
	if UpdateMajor.String() != "major" || UpdatePatch.String() != "patch" {
		t.Errorf("String() = %q, %q", UpdateMajor, UpdatePatch)
	}
}

func TestUpdateTypeRank(t *testing.T) {
	if UpdateMajor.Rank() >= UpdateMinor.Rank() || UpdateMinor.Rank() >= UpdatePatch.Rank() {
		t.Errorf("ranks = %d, %d, %d; want major < minor < patch", UpdateMajor.Rank(), UpdateMinor.Rank(), UpdatePatch.Rank())
	}
}
