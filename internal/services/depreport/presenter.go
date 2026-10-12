package depreport

import (
	"strings"

	"automation-hub/internal/messages"
)

// Presenter renders a Summary as one payload: the caption or the HTML document.
type Presenter interface {
	Present(s Summary) ([]byte, error)
}

// htmlEscaper escapes the report values that go into HTML. Catalog texts are not escaped.
var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// renderer renders catalog texts and keeps the first error, so a builder reads its texts
// inline and checks the error once at the end.
type renderer struct {
	cat *messages.Catalog
	err error
}

func (r *renderer) text(key string, vars map[string]string) string {
	if r.err != nil {
		return ""
	}
	out, err := r.cat.Render(key, vars)
	r.err = err
	return out
}
