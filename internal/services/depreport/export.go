package depreport

import (
	"fmt"
	"strconv"
	"strings"
)

// Fixed names of the YAML export. They are the machine format, so they are not catalog texts.
const (
	yamlScannedAt    = "scanned_at"
	yamlImages       = "images"
	yamlDependencies = "dependencies"
	yamlRepo         = "repo"
	yamlFork         = "fork"
	yamlFiles        = "files"
	yamlPath         = "path"
	yamlUpdates      = "updates"
	yamlName         = "name"
	yamlCurrent      = "current"
	yamlNew          = "new"
	yamlType         = "type"
)

// ExportView is the YAML export of the document. The dialog offers one checkbox per repo in
// Sections; the script joins the checked fragments under their section keys. With no sections the
// document has no export button and no dialog.
type ExportView struct {
	Header   string // the scanned_at line, which starts the export
	FileName string // the download name
	Sections []ExportSection
}

// ExportSection is one section of the dialog: the images or the dependencies, in card order.
type ExportSection struct {
	Key   string // YAML key of the section
	Label string // the catalog name of the tab
	Repos []ExportRepo
}

// ExportRepo is one repo of a section: its checkbox label and the YAML fragment of its entry.
type ExportRepo struct {
	Repo  string // full repo name, shown as the tooltip
	Short string // short repo name, the checkbox label
	Note  string // fork note, empty when the repo is not a fork
	YAML  string // the fragment, one list item of the section
}

// buildExport builds the export from the images and the dependencies tab. Empty tabs are left out, so
// the images section comes first and both sections keep the card order of their tab.
func buildExport(scanTimestamp, fileName string, images, deps TabView) ExportView {
	view := ExportView{Header: exportHeader(scanTimestamp), FileName: fileName}
	for _, s := range []struct {
		key string
		tab TabView
	}{{yamlImages, images}, {yamlDependencies, deps}} {
		if len(s.tab.Groups) == 0 {
			continue
		}
		section := ExportSection{Key: s.key, Label: s.tab.Label}
		for _, g := range s.tab.Groups {
			section.Repos = append(section.Repos, ExportRepo{Repo: g.Repo, Short: g.Short, Note: g.Note, YAML: repoFragment(g)})
		}
		view.Sections = append(view.Sections, section)
	}
	return view
}

// exportHeader is the first line of the export: the scan time, double-quoted.
func exportHeader(scanTimestamp string) string {
	return yamlScannedAt + ": " + strconv.Quote(scanTimestamp) + "\n"
}

// repoFragment is the list item of one repo card: its fork flag, then its package files and their
// updates. Every string is a double-quoted scalar, so report text cannot add keys or list items.
func repoFragment(g RepoGroupView) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  - %s: %s\n", yamlRepo, strconv.Quote(g.Repo))
	fmt.Fprintf(&b, "    %s: %t\n", yamlFork, g.Fork)
	fmt.Fprintf(&b, "    %s:\n", yamlFiles)
	for _, f := range g.Files {
		fmt.Fprintf(&b, "      - %s: %s\n", yamlPath, strconv.Quote(f.Path))
		fmt.Fprintf(&b, "        %s:\n", yamlUpdates)
		for _, e := range f.Entries {
			fmt.Fprintf(&b, "          - %s: %s\n", yamlName, strconv.Quote(e.Name))
			fmt.Fprintf(&b, "            %s: %s\n", yamlCurrent, strconv.Quote(e.Current))
			fmt.Fprintf(&b, "            %s: %s\n", yamlNew, strconv.Quote(e.New))
			fmt.Fprintf(&b, "            %s: %s\n", yamlType, strconv.Quote(e.Type.String()))
		}
	}
	return b.String()
}
