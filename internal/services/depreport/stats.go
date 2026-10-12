package depreport

import (
	"fmt"
	"strconv"
)

// typesInBar is the order of the distribution bar and of the legend: patch, minor, major.
var typesInBar = []UpdateType{UpdatePatch, UpdateMinor, UpdateMajor}

// StatsView is the panel of a tab: the eyebrow, the number of listed updates, the sub line, the bar and
// the legend.
type StatsView struct {
	Eyebrow  string
	Big      string
	Sub      string
	Segments []BarSegment
	Legend   []LegendItem
}

// BarSegment is one type of the distribution bar. Percent is its share of the listed updates.
type BarSegment struct {
	Type    UpdateType
	Percent string // for example "7.7%"
}

// LegendItem is one type of the legend. A zero count is kept.
type LegendItem struct {
	Type  UpdateType
	Label string
	Count string
}

// segmentsOf lists the bar segments of the non-zero types. A tab without listed updates has none.
func segmentsOf(tab Tab) []BarSegment {
	var segments []BarSegment
	for _, t := range typesInBar {
		if n := tab.Counts[t]; n > 0 {
			segments = append(segments, BarSegment{Type: t, Percent: percentOf(n, tab.Updates)})
		}
	}
	return segments
}

// percentOf is n as a percentage of total, with one decimal.
func percentOf(n, total int) string {
	return fmt.Sprintf("%.1f%%", float64(n)*100/float64(total))
}

// legendOf lists every type with its count, zero included.
func legendOf(tab Tab, labels labelSet) []LegendItem {
	items := make([]LegendItem, 0, len(typesInBar))
	for _, t := range typesInBar {
		items = append(items, LegendItem{Type: t, Label: labels.typeName(t), Count: strconv.Itoa(tab.Counts[t])})
	}
	return items
}
