package depreport

// tabIndex is the tabindex of a tab: the selected tab is in the tab order, the others are reached with
// the arrow keys (roving tabindex, WAI-ARIA tabs).
func tabIndex(selected bool) string {
	if selected {
		return "0"
	}
	return "-1"
}
