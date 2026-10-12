package depreport

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Form fields the dependency-report Job posts. They are the Job-to-hub contract.
const (
	FieldExitCode = "exit_code"
	FieldPod      = "pod"
	FieldLogTail  = "log_tail"
	FieldReport   = "report"
)

// Errors returned by ParseSubmission. The handler answers ErrTooLarge with 413 and ErrInvalid with 400.
var (
	ErrTooLarge = errors.New("request body too large")
	ErrInvalid  = errors.New("invalid submission")
)

// Submission is one request of the Job: the Renovate exit code, the pod, the log tail and,
// when Renovate wrote one, the report.
type Submission struct {
	ExitCode  int
	Pod       string
	LogTail   []byte
	Report    []byte
	HasReport bool
}

// Failed reports whether the Renovate run failed, which means a failure message and no document.
func (s Submission) Failed() bool {
	return s.ExitCode != 0 || !s.HasReport
}

// ParseSubmission reads the multipart body of r, which may be at most maxBody bytes.
// A missing report is not an error; it makes the submission Failed.
func ParseSubmission(r *http.Request, maxBody int64) (Submission, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBody)
	if err := r.ParseMultipartForm(maxBody); err != nil { //nolint:gosec // G120: the body is capped by MaxBytesReader above
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return Submission{}, fmt.Errorf("%w: limit is %d bytes", ErrTooLarge, maxBody)
		}
		return Submission{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	exitCode, err := strconv.Atoi(strings.TrimSpace(r.PostFormValue(FieldExitCode)))
	if err != nil {
		return Submission{}, fmt.Errorf("%w: %s is not an integer", ErrInvalid, FieldExitCode)
	}
	logTail, _, err := readFile(r, FieldLogTail)
	if err != nil {
		return Submission{}, err
	}
	report, hasReport, err := readFile(r, FieldReport)
	if err != nil {
		return Submission{}, err
	}
	return Submission{
		ExitCode:  exitCode,
		Pod:       r.PostFormValue(FieldPod),
		LogTail:   logTail,
		Report:    report,
		HasReport: hasReport,
	}, nil
}

// readFile reads an uploaded file part. A missing part is not an error.
func readFile(r *http.Request, name string) ([]byte, bool, error) {
	file, _, err := r.FormFile(name)
	if errors.Is(err, http.ErrMissingFile) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("%w: %s: %w", ErrInvalid, name, err)
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, false, fmt.Errorf("%w: read %s: %w", ErrInvalid, name, err)
	}
	return data, true, nil
}
