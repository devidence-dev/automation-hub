package depreport

import (
	"slices"
	"strconv"
	"strings"
)

// RepoGroupView is one repo card of a tab: its summary row and its package files.
type RepoGroupView struct {
	Repo  string // full repo name, shown when the card is open
	Short string // short repo name, shown in the summary row
	Note  string // fork note of the summary row, empty when the repo is not a fork
	Dot   UpdateType
	Chips []CountChip // one chip per update type with dependencies in the card
	Files []FileGroupView
}

// CountChip is one text chip of the summary row: a count and its type, colored by the type.
type CountChip struct {
	Type UpdateType
	Text string
}

// FileGroupView is one package file of a repo card and its entries. The location shows once, under the
// entries: Dir is the muted folder part (empty for a path of one or two segments), Short the rest.
type FileGroupView struct {
	Path    string
	Dir     string
	Short   string
	Entries []EntryView
}

// EntryView is one listed update: a dependency, or an image with its versions, and its copy button.
// For an image Name is the image name and Current and New are its tags.
type EntryView struct {
	Name         string
	Current      string // current version
	New          string // new version
	Type         UpdateType
	TypeLabel    string // the catalog name of Type, shown in the chip
	CurrentLabel string // screen-reader label of Current
	NewLabel     string // screen-reader label of New
	Copy         CopyView
}

// CopyView is the icon-only copy button of an entry. Label is its accessible name and tooltip.
type CopyView struct {
	Value        string // the text copied to the clipboard
	Label        string
	Copied       string // the text announced after a copy
	CopiedBefore string // the title shown after the first copy
}

// groupTab regroups the rows of a tab by repo, then package file, then entry. The rows are sorted by
// repo and then package file, so each repo and each package file is one run of rows. The cards follow
// the caption order of the tab's pending repos. images tells whether the tab lists images.
func groupTab(r *renderer, tab Tab, labels labelSet, images bool) []RepoGroupView {
	byRepo := map[string]*RepoGroupView{}
	for _, row := range tab.Rows {
		g := byRepo[row.Repo]
		if g == nil {
			g = &RepoGroupView{Repo: row.Repo, Short: ShortName(row.Repo)}
			byRepo[row.Repo] = g
		}
		if n := len(g.Files); n == 0 || g.Files[n-1].Path != row.PackageFile {
			dir, short := splitPath(row.PackageFile)
			g.Files = append(g.Files, FileGroupView{Path: row.PackageFile, Dir: dir, Short: short})
		}
		f := &g.Files[len(g.Files)-1]
		f.Entries = append(f.Entries, entryOf(row, labels, images))
	}
	groups := make([]RepoGroupView, 0, len(tab.Pending))
	for _, p := range tab.Pending {
		g := byRepo[p.Repo]
		if g == nil {
			continue
		}
		if p.Fork {
			g.Note = strings.TrimSpace(r.text(KeyForkTagDocument, nil))
		}
		sortFiles(g.Files)
		g.Dot = dotOf(p)
		g.Chips = chipsOf(r, p)
		groups = append(groups, *g)
	}
	return groups
}

// splitPath splits a package file path into its folder, which is every segment but the last two, and the
// rest. A path of one or two segments has no folder.
func splitPath(path string) (dir, short string) {
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return "", path
	}
	n := len(parts) - 2
	return strings.Join(parts[:n], "/") + "/", strings.Join(parts[n:], "/")
}

// sortFiles orders the entries of each file major before minor before patch, then the files by their
// highest type and, on a tie, the files of one entry first. The sorts are stable, so equal keys keep the
// order of the rows.
func sortFiles(files []FileGroupView) {
	for _, f := range files {
		slices.SortStableFunc(f.Entries, func(a, b EntryView) int { return a.Type.Rank() - b.Type.Rank() })
	}
	slices.SortStableFunc(files, func(a, b FileGroupView) int {
		if d := a.Entries[0].Type.Rank() - b.Entries[0].Type.Rank(); d != 0 {
			return d
		}
		return min(len(a.Entries), 2) - min(len(b.Entries), 2)
	})
}

// dotOf is the type of the summary dot: major when a dependency has a major update, otherwise minor,
// otherwise patch.
func dotOf(p RepoPending) UpdateType {
	switch {
	case p.Major > 0:
		return UpdateMajor
	case p.Minor > 0:
		return UpdateMinor
	default:
		return UpdatePatch
	}
}

// chipsOf lists one chip per non-zero count of a repo, major first.
func chipsOf(r *renderer, p RepoPending) []CountChip {
	var chips []CountChip
	for _, c := range []struct {
		typ UpdateType
		n   int
	}{{UpdateMajor, p.Major}, {UpdateMinor, p.Minor}, {UpdatePatch, p.Patch}} {
		if c.n > 0 {
			chips = append(chips, CountChip{Type: c.typ, Text: r.text(countKey(c.typ), map[string]string{"count": strconv.Itoa(c.n)})})
		}
	}
	return chips
}

// countKey is the catalog key of the "{count} type" text of an update type.
func countKey(t UpdateType) string {
	switch t {
	case UpdateMajor:
		return KeyCountMajor
	case UpdateMinor:
		return KeyCountMinor
	default:
		return KeyCountPatch
	}
}

// entryOf builds the entry of one row. Each entry has one copy button: an image copies its new
// reference, a dependency copies "name@new version".
func entryOf(row Row, labels labelSet, images bool) EntryView {
	e := EntryView{
		Name:         row.Name,
		Current:      row.Version,
		New:          row.NewVersion,
		Type:         row.Type,
		TypeLabel:    labels.typeName(row.Type),
		CurrentLabel: labels.Current,
		NewLabel:     labels.New,
	}
	if images {
		e.Copy = labels.copyOf(row.Name+":"+row.NewVersion, labels.CopyLatest)
		return e
	}
	e.Copy = labels.copyOf(row.Name+"@"+row.NewVersion, labels.CopyNewVersion)
	return e
}

// copyOf is a copy button that copies value and carries the labels of the document.
func (l labelSet) copyOf(value, label string) CopyView {
	return CopyView{Value: value, Label: label, Copied: l.Copied, CopiedBefore: l.CopiedBefore}
}
