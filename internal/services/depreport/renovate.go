package depreport

import (
	"encoding/json"
	"fmt"
)

// The types below hold only the fields of Renovate's report that the format
// uses (lib/instrumentation/types.ts in Renovate 44.151.0).
type Report struct {
	Problems     []Problem             `json:"problems"`
	Repositories map[string]RepoReport `json:"repositories"`
}

type RepoReport struct {
	Problems     []Problem                `json:"problems"`
	PackageFiles map[string][]PackageFile `json:"packageFiles"`
}

type Problem struct {
	Level int    `json:"level"`
	Msg   string `json:"msg"`
}

type PackageFile struct {
	PackageFile string `json:"packageFile"`
	Deps        []Dep  `json:"deps"`
}

type Dep struct {
	DepName        *string  `json:"depName"`
	Datasource     *string  `json:"datasource"`
	CurrentVersion *string  `json:"currentVersion"`
	CurrentValue   *string  `json:"currentValue"`
	Updates        []Update `json:"updates"`
}

type Update struct {
	UpdateType string  `json:"updateType"`
	NewVersion *string `json:"newVersion"`
	NewValue   *string `json:"newValue"`
}

// ParseReport decodes the Renovate report file.
func ParseReport(data []byte) (Report, error) {
	var rep Report
	if err := json.Unmarshal(data, &rep); err != nil {
		return Report{}, fmt.Errorf("parse renovate report: %w", err)
	}
	return rep, nil
}
