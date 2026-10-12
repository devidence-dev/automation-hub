package depreport

// UpdateType is the type of one Renovate update. Only the listed types become rows and caption counts.
type UpdateType int

const (
	UpdateUnknown UpdateType = iota // digest, pin and any type Renovate adds later
	UpdateMajor
	UpdateMinor
	UpdatePatch
)

// ParseUpdateType maps Renovate's updateType to an UpdateType. Any other value is UpdateUnknown.
func ParseUpdateType(s string) UpdateType {
	switch s {
	case "major":
		return UpdateMajor
	case "minor":
		return UpdateMinor
	case "patch":
		return UpdatePatch
	default:
		return UpdateUnknown
	}
}

// Listed reports whether the report shows the update: major, minor or patch.
func (t UpdateType) Listed() bool {
	switch t {
	case UpdateMajor, UpdateMinor, UpdatePatch:
		return true
	default:
		return false
	}
}

// Rank orders the listed types for the tables and the caption: major before minor before patch.
func (t UpdateType) Rank() int {
	switch t {
	case UpdateMajor:
		return 0
	case UpdateMinor:
		return 1
	case UpdatePatch:
		return 2
	default:
		return 3
	}
}

// String is the Renovate name of the type, which is also the badge text.
func (t UpdateType) String() string {
	switch t {
	case UpdateMajor:
		return "major"
	case UpdateMinor:
		return "minor"
	case UpdatePatch:
		return "patch"
	default:
		return "unknown"
	}
}
