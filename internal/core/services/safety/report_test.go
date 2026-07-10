package safety

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestReportScamURLSendsWhenURLIsNew(t *testing.T) {
	catalog := &fakeScamURLCatalog{}
	sender := &fakeScamReportSender{}
	service := NewReportService(catalog, sender)

	report, err := service.Report(context.Background(), "ftp://bad.example/path", " user-1 ")
	if err != nil {
		t.Fatalf("report scam URL: %v", err)
	}
	if report.URL != "ftp://bad.example/path" || report.ReporterUserID != "user-1" {
		t.Fatalf("report = %#v", report)
	}
	if catalog.Checked != "ftp://bad.example/path" {
		t.Fatalf("checked = %q", catalog.Checked)
	}
	if len(sender.Sent) != 1 || sender.Sent[0] != report {
		t.Fatalf("sent = %#v", sender.Sent)
	}
}

func TestReportScamURLRejectsInvalidURL(t *testing.T) {
	sender := &fakeScamReportSender{}
	service := NewReportService(&fakeScamURLCatalog{}, sender)

	_, err := service.Report(context.Background(), "not-a-url", "user-1")
	if !errors.Is(err, domain.ErrInvalidScamURLReport) {
		t.Fatalf("expected invalid report, got %v", err)
	}
	if len(sender.Sent) != 0 {
		t.Fatalf("sent = %#v", sender.Sent)
	}
}

func TestReportScamURLDoesNotTrimLegacyInput(t *testing.T) {
	sender := &fakeScamReportSender{}
	service := NewReportService(&fakeScamURLCatalog{}, sender)
	_, err := service.Report(context.Background(), " https://bad.example ", "user-1")
	if !errors.Is(err, domain.ErrInvalidScamURLReport) || len(sender.Sent) != 0 {
		t.Fatalf("error=%v sent=%#v", err, sender.Sent)
	}
}

func TestReportScamURLRejectsKnownURL(t *testing.T) {
	sender := &fakeScamReportSender{}
	service := NewReportService(&fakeScamURLCatalog{Exists: true}, sender)

	_, err := service.Report(context.Background(), "https://bad.example", "user-1")
	if !errors.Is(err, domain.ErrScamURLAlreadyReported) {
		t.Fatalf("expected already reported, got %v", err)
	}
	if len(sender.Sent) != 0 {
		t.Fatalf("sent = %#v", sender.Sent)
	}
}

func TestReportScamURLPropagatesSenderError(t *testing.T) {
	want := errors.New("send failed")
	service := NewReportService(&fakeScamURLCatalog{}, &fakeScamReportSender{Err: want})

	_, err := service.Report(context.Background(), "https://bad.example", "user-1")
	if !errors.Is(err, want) {
		t.Fatalf("expected sender error, got %v", err)
	}
}

type fakeScamURLCatalog struct {
	Checked string
	Exists  bool
	Err     error
}

func (f *fakeScamURLCatalog) ContainsScamURL(ctx context.Context, rawURL string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	f.Checked = rawURL
	if f.Err != nil {
		return false, f.Err
	}
	return f.Exists, nil
}

func (f *fakeScamURLCatalog) FindScamURLInContent(ctx context.Context, content string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	return "", false, nil
}

type fakeScamReportSender struct {
	Sent []domain.ScamURLReport
	Err  error
}

func (f *fakeScamReportSender) SendScamURLReport(ctx context.Context, report domain.ScamURLReport) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if f.Err != nil {
		return f.Err
	}
	f.Sent = append(f.Sent, report)
	return nil
}
