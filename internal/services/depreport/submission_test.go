package depreport

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// multipartRequest builds the Job's multipart request: text fields and file parts.
func multipartRequest(t *testing.T, fields, files map[string]string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for k, v := range files {
		fw, err := mw.CreateFormFile(k, k)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fw.Write([]byte(v))
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/webhook/dependency-report", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestParseSubmissionValid(t *testing.T) {
	req := multipartRequest(t,
		map[string]string{FieldExitCode: " 0 ", FieldPod: "pod-1"},
		map[string]string{FieldLogTail: "last line", FieldReport: "{}"})
	sub, err := ParseSubmission(req, 1<<20)
	if err != nil {
		t.Fatalf("ParseSubmission() error = %v", err)
	}
	if sub.ExitCode != 0 || sub.Pod != "pod-1" || string(sub.LogTail) != "last line" || string(sub.Report) != "{}" || !sub.HasReport {
		t.Errorf("submission = %+v", sub)
	}
	if sub.Failed() {
		t.Error("a successful run reports Failed")
	}
}

func TestParseSubmissionTooLarge(t *testing.T) {
	req := multipartRequest(t, map[string]string{FieldExitCode: "0"}, map[string]string{FieldReport: "0123456789012345678901234567890123456789"})
	_, err := ParseSubmission(req, 20)
	if !errors.Is(err, ErrTooLarge) {
		t.Errorf("ParseSubmission() error = %v, want ErrTooLarge", err)
	}
}

func TestParseSubmissionBadExitCode(t *testing.T) {
	req := multipartRequest(t, map[string]string{FieldExitCode: "abc"}, nil)
	if _, err := ParseSubmission(req, 1<<20); !errors.Is(err, ErrInvalid) {
		t.Errorf("ParseSubmission() error = %v, want ErrInvalid", err)
	}
}

func TestParseSubmissionMissingReport(t *testing.T) {
	req := multipartRequest(t, map[string]string{FieldExitCode: "0", FieldPod: "p"}, nil)
	sub, err := ParseSubmission(req, 1<<20)
	if err != nil {
		t.Fatalf("ParseSubmission() error = %v", err)
	}
	if sub.HasReport || !sub.Failed() {
		t.Errorf("submission = %+v, want no report and Failed", sub)
	}
}

func TestParseSubmissionNotMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("x=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if _, err := ParseSubmission(req, 1<<20); !errors.Is(err, ErrInvalid) {
		t.Errorf("ParseSubmission() error = %v, want ErrInvalid", err)
	}
}
