package poll

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

const (
	pollTestPermission  = int64(permissionManageMessages)
	pollTestRandomColor = 0x123456
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestCreateHandlerSendsLegacyPollUIAndSavesDocument(t *testing.T) {
	repo := fakemongo.NewPollRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.NonBotMembers = 10
	module := NewModuleWithSideEffects(repo, sideEffects, sideEffects, fixedClock{now: time.UnixMilli(1700000000000)})
	module.randomColor = func() int { return pollTestRandomColor }
	responder := fakediscord.NewResponder()
	interaction := pollCreateInteraction("今天吃什麼?", "拉麵^壽司^咖哩")

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("create handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0]
	if len(sent.Message.Embeds) != 1 || sent.Message.Embeds[0].Title != "<:poll:1023968837965709312> | 投票\n今天吃什麼?" {
		t.Fatalf("poll embed = %#v", sent.Message.Embeds)
	}
	if sent.Message.Embeds[0].Color != pollTestRandomColor {
		t.Fatalf("poll color = %#x", sent.Message.Embeds[0].Color)
	}
	description := sent.Message.Embeds[0].Description
	for _, want := range []string{"總投票人數:`0` / `10`", "每人可以投給`1`個選項", "`不能`改投其他選項", "`無法`看到投票結果", "`實名`投票"} {
		if !strings.Contains(description, want) {
			t.Fatalf("description missing %q: %s", want, description)
		}
	}
	if len(sent.Message.Components) != 2 {
		t.Fatalf("components = %#v", sent.Message.Components)
	}
	firstRow := sent.Message.Components[0].Components
	if len(firstRow) != 4 {
		t.Fatalf("first row = %#v", firstRow)
	}
	if firstRow[0].Label != "拉麵" || firstRow[0].Style != "secondary" || !strings.HasPrefix(firstRow[0].CustomID, "mhcat:v1:poll:vote:") {
		t.Fatalf("first choice button = %#v", firstRow[0])
	}
	if firstRow[3].Label != "查看投票結果" || firstRow[3].Emoji != "<:analysis:1023965999357243432>" || firstRow[3].Style != "success" {
		t.Fatalf("result button = %#v", firstRow[3])
	}
	menu := sent.Message.Components[1].Components[0]
	if menu.Type != "select" || menu.Placeholder != "🔧投票發起人操作" || len(menu.Options) != 6 {
		t.Fatalf("owner menu = %#v", menu)
	}
	saved, err := repo.GetPoll(context.Background(), interaction.Actor.GuildID, sent.Ref.MessageID)
	if err != nil {
		t.Fatalf("get saved poll: %v", err)
	}
	if saved.Question != "今天吃什麼?" || len(saved.Choices) != 3 || saved.CreatorID != interaction.Actor.UserID {
		t.Fatalf("saved poll = %#v", saved)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral || len(responder.Edits) != 1 {
		t.Fatalf("responses defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "成功創建投票") {
		t.Fatalf("success edit = %#v", responder.Edits)
	}
	if got := responder.Edits[0].Embeds[0].Color; got != 0x57F287 {
		t.Fatalf("success color = %#x", got)
	}
	if len(sideEffects.Edited) != 1 || sideEffects.Edited[0].Message.Embeds[0].Color != pollTestRandomColor {
		t.Fatalf("persisted poll edits = %#v", sideEffects.Edited)
	}
}

func TestCreateHandlerPreservesQuestionAndChoiceWhitespace(t *testing.T) {
	repo := fakemongo.NewPollRepository()
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, sideEffects, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := pollCreateInteraction("  今天吃什麼?  ", " A ^ ")

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("create handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 || len(sideEffects.Sent[0].Message.Embeds) != 1 {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	if got := sideEffects.Sent[0].Message.Embeds[0].Title; got != "<:poll:1023968837965709312> | 投票\n  今天吃什麼?  " {
		t.Fatalf("poll title = %q", got)
	}
	components := sideEffects.Sent[0].Message.Components[0].Components
	if components[0].Label != " A " || components[1].Label != " " {
		t.Fatalf("choice labels = %#v", components)
	}
	saved, err := repo.GetPoll(context.Background(), interaction.Actor.GuildID, sideEffects.Sent[0].Ref.MessageID)
	if err != nil {
		t.Fatalf("get saved poll: %v", err)
	}
	if saved.Question != "  今天吃什麼?  " {
		t.Fatalf("saved question = %q", saved.Question)
	}
	if len(saved.Choices) != 2 || saved.Choices[0] != " A " || saved.Choices[1] != " " {
		t.Fatalf("saved choices = %#v", saved.Choices)
	}
}

func TestCreateHandlerCleansSentPollAfterCanceledRepositoryFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wantErr := errors.New("create poll failed")
	repo := &cancelingCreatePollRepository{
		PollRepository: fakemongo.NewPollRepository(),
		cancel:         cancel,
		err:            wantErr,
	}
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, sideEffects, nil, nil)

	err := module.CreateHandler()(ctx, pollCreateInteraction("問題", "A^B"), fakediscord.NewResponder())
	if !errors.Is(err, wantErr) {
		t.Fatalf("create handler error = %v", err)
	}
	if len(sideEffects.Sent) != 1 || len(sideEffects.DeletedMessage) != 1 || sideEffects.DeletedMessage[0] != sideEffects.Sent[0].Ref {
		t.Fatalf("sent = %#v deleted = %#v", sideEffects.Sent, sideEffects.DeletedMessage)
	}
}

func TestCreateHandlerRequiresManageMessages(t *testing.T) {
	module := NewModuleWithSideEffects(fakemongo.NewPollRepository(), fakediscord.NewSideEffects(), nil, nil)
	responder := fakediscord.NewResponder()
	interaction := pollCreateInteraction("問題", "A^B")
	interaction.Actor.PermissionBits = 0

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("create handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("permission defers = %#v", responder.Defers)
	}
	if len(responder.Replies) != 0 {
		t.Fatalf("permission replies = %#v", responder.Replies)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("permission edits = %#v", responder.Edits)
	}
	if got := responder.Edits[0].Embeds[0].Color; got != 0xED4245 {
		t.Fatalf("error color = %#x", got)
	}
}

func TestCreateHandlerRejectsInvalidOptionsBeforeSend(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(fakemongo.NewPollRepository(), sideEffects, nil, nil)
	cases := []struct {
		name    string
		options string
		want    string
	}{
		{name: "too few", options: "A", want: "最少需要2個選項!"},
		{name: "duplicate", options: "A^A", want: "選項名稱不可以重複!"},
		{name: "duplicate empty", options: "^^B", want: "選項名稱不可以重複!"},
		{name: "duplicate overlong", options: strings.Repeat("A", 81) + "^" + strings.Repeat("A", 81), want: "選項名稱不可以重複!"},
		{name: "empty", options: "A^^B", want: "^跟^中間請填入選項，不可為空"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			responder := fakediscord.NewResponder()
			if err := module.CreateHandler()(context.Background(), pollCreateInteraction("問題", tc.options), responder); err != nil {
				t.Fatalf("create handler: %v", err)
			}
			if len(sideEffects.Sent) != 0 {
				t.Fatalf("unexpected sends = %#v", sideEffects.Sent)
			}
			if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, tc.want) {
				t.Fatalf("edits = %#v", responder.Edits)
			}
		})
	}
}

func TestValidatePollInputUsesLegacyUTF16Lengths(t *testing.T) {
	tests := []struct {
		name     string
		question string
		choices  string
		want     string
	}{
		{name: "question at emoji limit", question: strings.Repeat("😀", 1250), choices: "A^B"},
		{name: "question over emoji limit", question: strings.Repeat("😀", 1251), choices: "A^B", want: "問題字數不可超過2500"},
		{name: "choice at emoji limit", question: "Q", choices: strings.Repeat("😀", 40) + "^B"},
		{name: "choice over emoji limit", question: "Q", choices: strings.Repeat("😀", 41) + "^B", want: "你輸入的選項字數不能超過80"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, got := validatePollInput(tc.question, tc.choices)
			if got != tc.want {
				t.Fatalf("validation message = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVoteHandlerAddsVoteAndRerendersPoll(t *testing.T) {
	repo := seededPollRepo(t)
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.NonBotMembers = 4
	module := NewModuleWithSideEffects(repo, sideEffects, sideEffects, fixedClock{now: time.UnixMilli(1700000000000)})
	responder := fakediscord.NewResponder()
	interaction := pollButtonInteraction("mhcat:v1:poll:vote:i=1")

	if err := module.VoteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("vote handler: %v", err)
	}
	poll, err := repo.GetPoll(context.Background(), interaction.Actor.GuildID, interaction.MessageID)
	if err != nil {
		t.Fatalf("get poll: %v", err)
	}
	if len(poll.Votes) != 1 || poll.Votes[0].Choice != "B" || poll.Votes[0].Time != "1700000000000" {
		t.Fatalf("votes = %#v", poll.Votes)
	}
	if len(sideEffects.Edited) != 1 || !strings.Contains(sideEffects.Edited[0].Message.Embeds[0].Description, "總投票人數:`1` / `4`") {
		t.Fatalf("edited messages = %#v", sideEffects.Edited)
	}
	if !strings.Contains(sideEffects.Edited[0].Message.Embeds[0].Description, "`無法`改投其他選項") {
		t.Fatalf("rerendered description = %q", sideEffects.Edited[0].Message.Embeds[0].Description)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你成功投給`B`") {
		t.Fatalf("vote response = %#v", responder.Edits)
	}
}

func TestVoteHandlerLegacyIDSupportsEightyCharacterChoice(t *testing.T) {
	choice := strings.Repeat("選", 80)
	repo := fakemongo.NewPollRepository()
	if _, err := repo.CreatePoll(context.Background(), domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "問題",
		CreatorID: "owner-1",
		Choices:   []string{choice, "B"},
	}); err != nil {
		t.Fatalf("seed poll: %v", err)
	}
	module := NewModuleWithSideEffects(repo, fakediscord.NewSideEffects(), nil, fixedClock{now: time.UnixMilli(1700000000000)})
	responder := fakediscord.NewResponder()
	interaction := pollButtonInteraction("poll_" + choice)

	if err := module.VoteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("vote handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, choice) {
		t.Fatalf("response = %#v", responder.Edits)
	}
}

func TestVoteHandlerPreservesChoiceWhitespace(t *testing.T) {
	repo := fakemongo.NewPollRepository()
	if _, err := repo.CreatePoll(context.Background(), domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  " ",
		CreatorID: "owner-1",
		Choices:   []string{" A ", " "},
	}); err != nil {
		t.Fatalf("seed poll: %v", err)
	}
	module := NewModuleWithSideEffects(repo, fakediscord.NewSideEffects(), nil, fixedClock{now: time.UnixMilli(1700000000000)})
	responder := fakediscord.NewResponder()

	if err := module.VoteHandler()(context.Background(), pollButtonInteraction("mhcat:v1:poll:vote:i=0"), responder); err != nil {
		t.Fatalf("vote handler: %v", err)
	}
	poll, err := repo.GetPoll(context.Background(), "guild-1", "message-1")
	if err != nil {
		t.Fatalf("get poll: %v", err)
	}
	if len(poll.Votes) != 1 || poll.Votes[0].Choice != " A " {
		t.Fatalf("votes = %#v", poll.Votes)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "` A `") {
		t.Fatalf("response = %#v", responder.Edits)
	}
}

func TestVoteHandlerDuplicateChoiceWithoutChangeUsesLegacyError(t *testing.T) {
	repo := seededPollRepo(t)
	_, err := repo.Vote(context.Background(), "guild-1", "message-1", "user-1", "A", "1")
	if err != nil {
		t.Fatalf("seed vote: %v", err)
	}
	module := NewModuleWithSideEffects(repo, fakediscord.NewSideEffects(), nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.VoteHandler()(context.Background(), pollButtonInteraction("mhcat:v1:poll:vote:i=0"), responder); err != nil {
		t.Fatalf("vote handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "不支援更改選項") {
		t.Fatalf("duplicate response = %#v", responder.Edits)
	}
}

func TestOwnerMenuTogglesPublicResultAndRerenders(t *testing.T) {
	repo := seededPollRepo(t)
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, sideEffects, sideEffects, nil)
	responder := fakediscord.NewResponder()
	interaction := pollMenuInteraction("poll_public_result")

	if err := module.OwnerMenuHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("owner menu handler: %v", err)
	}
	poll, err := repo.GetPoll(context.Background(), "guild-1", "message-1")
	if err != nil {
		t.Fatalf("get poll: %v", err)
	}
	if !poll.CanSeeResult {
		t.Fatalf("poll = %#v", poll)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:green_tick:994529015652163614>成功將投票結果設為公開!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(sideEffects.Edited) != 1 || sideEffects.Edited[0].Message.Components[1].Components[0].Options[0].Label != "隱藏投票結果" {
		t.Fatalf("rerendered menu = %#v", sideEffects.Edited)
	}
}

func TestOwnerMenuRejectsNonCreator(t *testing.T) {
	repo := seededPollRepo(t)
	module := NewModuleWithSideEffects(repo, fakediscord.NewSideEffects(), nil, nil)
	responder := fakediscord.NewResponder()
	interaction := pollMenuInteraction("poll_public_result")
	interaction.Actor.UserID = "not-owner"

	if err := module.OwnerMenuHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("owner menu handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你不是投票發起人") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestOwnerMenuManyChoiceReturnsVersionedSelect(t *testing.T) {
	repo := seededPollRepo(t)
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, sideEffects, nil, nil)
	module.randomColor = func() int { return pollTestRandomColor }
	responder := fakediscord.NewResponder()

	if err := module.OwnerMenuHandler()(context.Background(), pollMenuInteraction("poll_can_choose_many"), responder); err != nil {
		t.Fatalf("owner menu handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Components) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	selectMenu := responder.Edits[0].Components[0].Components[0]
	if selectMenu.Placeholder != "請選擇可以最多選擇數!" || !strings.HasPrefix(selectMenu.CustomID, "mhcat:v1:poll:max_choices:") {
		t.Fatalf("select = %#v", selectMenu)
	}
	if responder.Edits[0].Embeds[0].Color != pollTestRandomColor {
		t.Fatalf("max-choice menu color = %#x", responder.Edits[0].Embeds[0].Color)
	}
	if len(sideEffects.Edited) != 1 {
		t.Fatalf("refreshed polls = %#v", sideEffects.Edited)
	}
}

func TestMaxChoicesHandlerUpdatesPollAndMenuMessage(t *testing.T) {
	repo := seededPollRepo(t)
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, sideEffects, sideEffects, nil)
	module.randomColor = func() int { return pollTestRandomColor }
	responder := fakediscord.NewResponder()
	interaction := pollButtonInteraction("mhcat:v1:poll:max_choices:m=message-1")
	interaction.Values = []string{"2"}

	if err := module.MaxChoicesHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("max choices handler: %v", err)
	}
	poll, err := repo.GetPoll(context.Background(), "guild-1", "message-1")
	if err != nil {
		t.Fatalf("get poll: %v", err)
	}
	if poll.MaxChoices != 2 {
		t.Fatalf("poll max choices = %d", poll.MaxChoices)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "成功將最多選擇數量設為2") {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	if responder.Updates[0].Embeds[0].Color != pollTestRandomColor {
		t.Fatalf("max-choice success color = %#x", responder.Updates[0].Embeds[0].Color)
	}
	if len(sideEffects.Edited) != 1 || sideEffects.Edited[0].Message.Embeds[0].Color != pollTestRandomColor {
		t.Fatalf("rerendered poll = %#v", sideEffects.Edited)
	}
}

func TestResultHandlerReturnsLegacyTextFields(t *testing.T) {
	repo := seededPollRepo(t)
	_, _ = repo.TogglePoll(context.Background(), "guild-1", "message-1", domain.PollTogglePublicResult)
	_, _ = repo.Vote(context.Background(), "guild-1", "message-1", "user-1", "A", "1")
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.NonBotMembers = 4
	module := NewModuleWithSideEffects(repo, sideEffects, sideEffects, nil)
	module.randomColor = func() int { return pollTestRandomColor }
	responder := fakediscord.NewResponder()

	if err := module.ResultHandler()(context.Background(), pollButtonInteraction("mhcat:v1:poll:result:"), responder); err != nil {
		t.Fatalf("result handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || len(responder.Edits[0].Embeds[0].Fields) != 3 {
		t.Fatalf("result edits = %#v", responder.Edits)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Fields[0].Value, "<@user-1>") {
		t.Fatalf("result fields = %#v", responder.Edits[0].Embeds[0].Fields)
	}
	if responder.Edits[0].Embeds[0].Image == nil || responder.Edits[0].Embeds[0].Image.URL != "attachment://file.jpg" {
		t.Fatalf("result image = %#v", responder.Edits[0].Embeds[0].Image)
	}
	if responder.Edits[0].Embeds[0].Color != pollTestRandomColor {
		t.Fatalf("result color = %#x", responder.Edits[0].Embeds[0].Color)
	}
	if len(responder.Edits[0].Files) != 2 || responder.Edits[0].Files[0].Name != "file.jpg" || responder.Edits[0].Files[1].Name != "discord.txt" {
		t.Fatalf("result files = %#v", responder.Edits[0].Files)
	}
	if len(sideEffects.Edited) != 1 || !strings.Contains(sideEffects.Edited[0].Message.Embeds[0].Description, "總投票人數:`1` / `4`") {
		t.Fatalf("refreshed poll = %#v", sideEffects.Edited)
	}
}

func TestOwnerMenuExcelExportReturnsLegacyWorkbookAttachment(t *testing.T) {
	repo := seededPollRepo(t)
	_, _ = repo.Vote(context.Background(), "guild-1", "message-1", "user-1", "A", "1700000000000")
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.MemberTagValues["user-1"] = "Alice#1234"
	module := NewModuleWithSideEffects(repo, sideEffects, sideEffects, nil)
	responder := fakediscord.NewResponder()

	if err := module.OwnerMenuHandler()(context.Background(), pollMenuInteraction("poll_excel_result"), responder); err != nil {
		t.Fatalf("owner menu handler: %v", err)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	edit := responder.Edits[0]
	if edit.Content != "<:sheets:1023972957330100324> | **以下是該投票的excel表格!**" {
		t.Fatalf("content = %q", edit.Content)
	}
	if len(edit.Files) != 1 || edit.Files[0].Name != "poll_info.xlsx" || !strings.Contains(edit.Files[0].ContentType, "spreadsheetml") || len(edit.Files[0].Data) == 0 {
		t.Fatalf("files = %#v", edit.Files)
	}
	if len(sideEffects.Edited) != 1 {
		t.Fatalf("refreshed polls = %#v", sideEffects.Edited)
	}
}

func TestOwnerMenuAnonymousLockRerendersPoll(t *testing.T) {
	repo := seededPollRepo(t)
	_, _ = repo.TogglePoll(context.Background(), "guild-1", "message-1", domain.PollToggleAnonymous)
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, sideEffects, nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.OwnerMenuHandler()(context.Background(), pollMenuInteraction("poll_anonymous"), responder); err != nil {
		t.Fatalf("owner menu handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "匿名的投票無法改為實名") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(sideEffects.Edited) != 1 {
		t.Fatalf("refreshed polls = %#v", sideEffects.Edited)
	}
}

func TestOwnerMenuExcelExportRejectsAnonymousPoll(t *testing.T) {
	repo := seededPollRepo(t)
	_, _ = repo.TogglePoll(context.Background(), "guild-1", "message-1", domain.PollToggleAnonymous)
	module := NewModuleWithSideEffects(repo, fakediscord.NewSideEffects(), nil, nil)
	responder := fakediscord.NewResponder()

	if err := module.OwnerMenuHandler()(context.Background(), pollMenuInteraction("poll_excel_result"), responder); err != nil {
		t.Fatalf("owner menu handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "該投票為匿名") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func pollCreateInteraction(question, choices string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions("投票創建", "", map[string]string{
		"問題": question,
		"選項": choices,
	})
	interaction.Actor.PermissionBits = pollTestPermission
	interaction.Actor.GuildID = "guild-1"
	interaction.ChannelID = "channel-1"
	return interaction
}

func pollButtonInteraction(customID string) interactions.Interaction {
	interaction := fakediscord.ComponentInteractionFromID(customID)
	interaction.Actor.GuildID = "guild-1"
	interaction.Actor.UserID = "user-1"
	interaction.ChannelID = "channel-1"
	interaction.MessageID = "message-1"
	return interaction
}

func pollMenuInteraction(value string) interactions.Interaction {
	interaction := pollButtonInteraction("mhcat:v1:poll:owner_menu:")
	interaction.Values = []string{value}
	interaction.Actor.UserID = "owner-1"
	return interaction
}

func seededPollRepo(t *testing.T) *fakemongo.PollRepository {
	t.Helper()
	repo := fakemongo.NewPollRepository()
	_, err := repo.CreatePoll(context.Background(), domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "問題",
		CreatorID: "owner-1",
		Choices:   []string{"A", "B", "C"},
	})
	if err != nil {
		t.Fatalf("seed poll: %v", err)
	}
	return repo
}

var _ ports.PollRepository = (*fakemongo.PollRepository)(nil)
var _ responses.Responder = (*fakediscord.Responder)(nil)

type cancelingCreatePollRepository struct {
	ports.PollRepository
	cancel context.CancelFunc
	err    error
}

func (r *cancelingCreatePollRepository) CreatePoll(context.Context, domain.PollCreate) (domain.Poll, error) {
	r.cancel()
	return domain.Poll{}, r.err
}
