package depreport

import (
	"errors"
	"reflect"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// fakeSender records the sends in order. docErr fails SendDocument and msgErr fails SendHTML.
type fakeSender struct {
	calls  []string // "document" or "message", then the text
	docErr error
	msgErr error
}

func (f *fakeSender) SendHTML(chatID, text string) error {
	f.calls = append(f.calls, "message:"+text)
	return f.msgErr
}

func (f *fakeSender) SendDocument(chatID, filename, caption string, data []byte) error {
	f.calls = append(f.calls, "document:"+caption)
	return f.docErr
}

func TestDeliverSendsDocumentWithCaption(t *testing.T) {
	sender := &fakeSender{}
	NewDeliverer(sender, "555", zap.NewNop()).Deliver(Delivery{Text: "caption", Filename: "r.html", Document: []byte("<html>")})
	if want := []string{"document:caption"}; !reflect.DeepEqual(sender.calls, want) {
		t.Errorf("calls = %q, want %q", sender.calls, want)
	}
}

func TestDeliverFallsBackToCaptionText(t *testing.T) {
	sender := &fakeSender{docErr: errors.New("rejected")}
	NewDeliverer(sender, "555", zap.NewNop()).Deliver(Delivery{Text: "caption", Filename: "r.html", Document: []byte("<html>")})
	if want := []string{"document:caption", "message:caption"}; !reflect.DeepEqual(sender.calls, want) {
		t.Errorf("calls = %q, want %q", sender.calls, want)
	}
}

func TestDeliverFailureMessageIsTextOnly(t *testing.T) {
	sender := &fakeSender{}
	NewDeliverer(sender, "555", zap.NewNop()).Deliver(Delivery{Text: "❌ failed"})
	if want := []string{"message:❌ failed"}; !reflect.DeepEqual(sender.calls, want) {
		t.Errorf("calls = %q, want %q", sender.calls, want)
	}
}

func TestDeliverLogsSenderError(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	sender := &fakeSender{msgErr: errors.New("telegram down")}
	NewDeliverer(sender, "555", zap.New(core)).Deliver(Delivery{Text: "❌ failed"})
	if n := logs.FilterMessage("Failed to send dependency report message").Len(); n != 1 {
		t.Errorf("send failure logged %d times, want once", n)
	}
}
