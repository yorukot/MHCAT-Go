package poll

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestPollResultMessageIncludesLegacyAttachments(t *testing.T) {
	poll := pollWithVotes(t)
	members := fakediscord.NewSideEffects()
	members.MemberTagValues["user-1"] = "Alice#1234"

	message := pollResultMessage(context.Background(), poll, members, 0x123456)

	if len(message.Embeds) != 1 || message.Embeds[0].Image == nil || message.Embeds[0].Image.URL != "attachment://file.jpg" {
		t.Fatalf("embed image = %#v", message.Embeds)
	}
	if message.Embeds[0].Color != 0x123456 {
		t.Fatalf("embed color = %#x", message.Embeds[0].Color)
	}
	if len(message.Files) != 2 {
		t.Fatalf("files = %#v", message.Files)
	}
	if message.Files[0].Name != pollChartFileName || message.Files[0].ContentType != "image/jpeg" || len(message.Files[0].Data) == 0 {
		t.Fatalf("chart file = %#v", message.Files[0])
	}
	if !bytes.HasPrefix(message.Files[0].Data, []byte{0xff, 0xd8, 0xff}) {
		t.Fatalf("chart is not jpeg: first bytes=%#v", message.Files[0].Data[:3])
	}
	if message.Files[1].Name != pollTextFileName || !strings.Contains(string(message.Files[1].Data), "Alice#1234") {
		t.Fatalf("text file = %#v content=%s", message.Files[1], string(message.Files[1].Data))
	}
}

func TestBuildPollTextExportPreservesAnonymousText(t *testing.T) {
	poll := pollWithVotes(t)
	poll.Anonymous = true
	members := fakediscord.NewSideEffects()
	members.Err = errors.New("member lookup should not be called")
	text := buildPollTextExport(context.Background(), poll, members)

	if !strings.Contains(text, "使用者id:該投票為匿名") || !strings.Contains(text, "使用者名稱:該投票為匿名") {
		t.Fatalf("anonymous text = %s", text)
	}
	if strings.Contains(text, "user-1") {
		t.Fatalf("anonymous text leaked user id: %s", text)
	}
}

func TestPollExcelMessageCreatesValidWorkbook(t *testing.T) {
	poll := pollWithVotes(t)
	members := fakediscord.NewSideEffects()
	members.MemberTagValues["user-1"] = "Alice#1234"

	message, err := pollExcelMessage(context.Background(), poll, members)
	if err != nil {
		t.Fatalf("excel message: %v", err)
	}
	if !strings.Contains(message.Content, "以下是該投票的excel表格") || len(message.Files) != 1 {
		t.Fatalf("message = %#v", message)
	}
	file := message.Files[0]
	if file.Name != pollXLSXFileName || file.ContentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("file = %#v", file)
	}
	reader, err := zip.NewReader(bytes.NewReader(file.Data), int64(len(file.Data)))
	if err != nil {
		t.Fatalf("open xlsx zip: %v", err)
	}
	sheet := readZipFile(t, reader, "xl/worksheets/sheet1.xml")
	for _, want := range []string{"使用者ID", "Alice#1234", "A", "台北標準時間"} {
		if !strings.Contains(sheet, want) {
			t.Fatalf("sheet missing %q: %s", want, sheet)
		}
	}
}

func TestBuildXLSXEscapesXML(t *testing.T) {
	data, err := buildXLSX([][]string{{"A&B", "<user>"}})
	if err != nil {
		t.Fatalf("build xlsx: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open xlsx zip: %v", err)
	}
	sheet := readZipFile(t, reader, "xl/worksheets/sheet1.xml")
	if !strings.Contains(sheet, "A&amp;B") || !strings.Contains(sheet, "&lt;user&gt;") {
		t.Fatalf("sheet did not escape XML: %s", sheet)
	}
}

func readZipFile(t *testing.T, reader *zip.Reader, name string) string {
	t.Helper()
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer handle.Close()
		data, err := io.ReadAll(handle)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return string(data)
	}
	t.Fatalf("zip file %s not found", name)
	return ""
}

func pollWithVotes(t *testing.T) domain.Poll {
	t.Helper()
	poll := domain.NewPoll(domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "問題",
		CreatorID: "owner-1",
		Choices:   []string{"A", "B"},
	})
	poll.CanSeeResult = true
	poll.Votes = []domain.PollVote{{
		UserID: "user-1",
		Choice: "A",
		Time:   domain.LegacyVoteTime(time.UnixMilli(1700000000000)),
	}}
	return poll
}
