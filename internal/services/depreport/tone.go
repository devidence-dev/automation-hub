package depreport

// Tone classes of the liquid glass design. Each one sets the tokens of an update type, or of the
// errors: the dot, the chip and the bar segment read them, so the colors follow the color scheme.
const (
	toneMajor = "tone-major"
	toneMinor = "tone-minor"
	tonePatch = "tone-patch"
	toneError = "tone-err"
)

// toneOf returns the tone class of an update type. An unknown type has no tone.
func toneOf(t UpdateType) string {
	switch t {
	case UpdateMajor:
		return toneMajor
	case UpdateMinor:
		return toneMinor
	case UpdatePatch:
		return tonePatch
	default:
		return ""
	}
}
