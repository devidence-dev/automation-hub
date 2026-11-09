package depreport

import (
	"fmt"
	"strings"
	"testing"
)

// nothingPendingReport is a repo with 12 docker and 30 go deps, and charts helm deps, none with an update.
func nothingPendingReport(t *testing.T, charts int) Report {
	t.Helper()
	var docker, others []string
	for i := 0; i < 12; i++ {
		docker = append(docker, fmt.Sprintf(`{"depName":"img-%02d","currentValue":"1","currentVersion":"1","datasource":"docker","updates":[]}`, i))
	}
	for i := 0; i < 30; i++ {
		others = append(others, fmt.Sprintf(`{"depName":"dep-%02d","currentValue":"1","currentVersion":"1","datasource":"go","updates":[]}`, i))
	}
	files := `{"packageFile":"Dockerfile","deps":[` + strings.Join(docker, ",") + `]}, {"packageFile":"go.mod","deps":[` + strings.Join(others, ",") + `]}`
	if charts > 0 {
		var helm []string
		for i := 0; i < charts; i++ {
			helm = append(helm, fmt.Sprintf(`{"depName":"chart-%02d","currentValue":"1","currentVersion":"1","datasource":"helm","updates":[]}`, i))
		}
		files += `, {"packageFile":"application.yaml","deps":[` + strings.Join(helm, ",") + `]}`
	}
	return reportOf(t, `"devidence-dev/QuantWarden":{"packageFiles":{"dockerfile":[`+files+`]}}`)
}
