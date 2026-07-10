package economy

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

//go:embed legacy_topic.json
var legacyTopicJSON []byte

const (
	coinGameOptionOpponent = "跟誰玩"
	coinGameOptionWager    = "賭注"
	coinGameErrorColor     = 0xED4245
	coinGameSuccessColor   = 0x53FF53
	coinGameInviteTTL      = 30 * time.Second
	coinGameKnowledgeTTL   = 21 * time.Second
	coinGameBlackjackTTL   = 31 * time.Second
	coinGameTimeoutRetry   = 100 * time.Millisecond
)

type knowledgeQuestion struct {
	Question  string   `json:"question"`
	Type      string   `json:"type"`
	Answer    string   `json:"anser"`
	Unanswers []string `json:"unanser"`
}

type coinGameSessionState string

const (
	coinGameSessionPending coinGameSessionState = "pending"
	coinGameSessionActive  coinGameSessionState = "active"
)

type coinGameSession struct {
	ID             string
	GuildID        string
	ChannelID      string
	MessageID      string
	ChallengerID   string
	ChallengerName string
	OpponentID     string
	OpponentName   string
	Kind           domain.CoinGameKind
	Wager          int64
	State          coinGameSessionState
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Deck              []int
	DealerCards       []int
	ChallengerCards   []int
	OpponentCards     []int
	BlackjackTurn     string
	ChallengerHit     *bool
	OpponentHit       *bool
	KnowledgeRound    int
	KnowledgeQuestion knowledgeQuestion
	KnowledgeAnswers  []string
	QuestionStartedAt time.Time
	TurnStartedAt     time.Time
	TurnDeadline      time.Time
	TurnGeneration    uint64
	ChallengerChoice  string
	OpponentChoice    string
	ChallengerScore   int64
	OpponentScore     int64
}

func (m Module) CoinGameHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		command, ok := m.coinGameCommandFromInteraction(interaction)
		if !ok {
			return responder.EditOriginal(ctx, coinGameErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if command.Wager < 0 {
			return responder.EditOriginal(ctx, coinGameErrorMessage("賭注必須大於-1"))
		}
		if _, err := m.game.CheckBalances(ctx, command); err != nil {
			return responder.EditOriginal(ctx, coinGameBalanceErrorMessage(err))
		}
		session := coinGameSession{
			GuildID:        command.GuildID,
			ChannelID:      interaction.ChannelID,
			ChallengerID:   command.ChallengerID,
			ChallengerName: interaction.Actor.Username,
			OpponentID:     command.OpponentID,
			Kind:           command.Kind,
			Wager:          command.Wager,
			State:          coinGameSessionPending,
		}
		if err := responder.FollowUp(ctx, coinGameInviteMessage(session)); err != nil {
			return err
		}
		m.gameSessions.Put(session)
		return m.trackCommand(ctx, interaction, CoinGameCommandName)
	}
}

func (m Module) CoinGameComponentHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		switch interaction.CustomID {
		case "teach21point":
			return responder.Reply(ctx, coinGameEphemeralText("21點比誰的牌點數較大但不超過21點。"))
		case "thansize":
			return responder.Reply(ctx, coinGameEphemeralText("比大小會為雙方各抽一個0到100的數字，數字較大者獲勝。"))
		}
		if interaction.CustomID == "lookmenumber" {
			session, ok := m.gameSessions.GetForComponent(interaction.Actor.GuildID, interaction.Actor.UserID, interaction.ChannelID, interaction.MessageID)
			if !ok {
				return responder.Reply(ctx, coinGameEphemeralError("很抱歉，找不到這場遊戲，請重新開始!"))
			}
			return m.showBlackjackCards(ctx, interaction, responder, session)
		}
		claim, ok := m.gameSessions.ClaimForComponent(interaction.Actor.GuildID, interaction.Actor.UserID, interaction.ChannelID, interaction.MessageID)
		if !ok {
			return responder.Reply(ctx, coinGameEphemeralError("很抱歉，找不到這場遊戲，請重新開始!"))
		}
		defer claim.Restore()
		session := claim.Session()
		switch interaction.CustomID {
		case "nooooo":
			claim.Delete()
			m.cancelCoinGameTimeout(session.ID)
			if err := responder.UpdateMessage(ctx, coinGameDisableInviteMessage(session)); err != nil {
				return err
			}
			return responder.FollowUp(ctx, responses.Message{
				Content:         fmt.Sprintf("<a:green_tick:994529015652163614> | **<@%s>拒絕此次邀請!**", interaction.Actor.UserID),
				AllowedMentions: &responses.AllowedMentions{},
			})
		case "yesssss":
			return m.acceptCoinGame(ctx, interaction, responder, session, claim)
		case "main_get_card", "main_no_card", "user_get_card", "user_no_card":
			return m.handleBlackjackAction(ctx, interaction, responder, session, claim)
		default:
			if session.Kind == domain.CoinGameKindKnowledge {
				return m.handleKnowledgeAnswer(ctx, interaction, responder, session, claim)
			}
			return responder.Reply(ctx, coinGameEphemeralError("很抱歉，出現了未知的錯誤，請重試!"))
		}
	}
}

func (m Module) coinGameCommandFromInteraction(interaction interactions.Interaction) (domain.CoinGameCommand, bool) {
	wager, ok := integerOption(interaction, coinGameOptionWager)
	if !ok {
		return domain.CoinGameCommand{}, false
	}
	opponentID := userIDOption(interaction, coinGameOptionOpponent)
	command := domain.CoinGameCommand{
		GuildID:      interaction.Actor.GuildID,
		ChallengerID: interaction.Actor.UserID,
		OpponentID:   opponentID,
		Wager:        wager,
		Kind:         domain.CoinGameKind(interaction.Subcommand),
	}.Normalize()
	if command.GuildID == "" || command.ChallengerID == "" || command.OpponentID == "" || command.ChallengerID == command.OpponentID || !command.Kind.Valid() {
		return domain.CoinGameCommand{}, false
	}
	return command, true
}

func userIDOption(interaction interactions.Interaction, name string) string {
	value := strings.TrimSpace(interaction.Options[name])
	if option, ok := interaction.CommandOptions[name]; ok && option.String != "" {
		value = strings.TrimSpace(option.String)
	}
	value = strings.TrimPrefix(value, "<@")
	value = strings.TrimPrefix(value, "!")
	value = strings.TrimSuffix(value, ">")
	return strings.TrimSpace(value)
}

func (m Module) acceptCoinGame(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, session coinGameSession, claim *coinGameSessionClaim) error {
	if interaction.Actor.UserID == session.ChallengerID {
		return responder.Reply(ctx, coinGameEphemeralError("你不是被邀請者，無法選擇接受!"))
	}
	if interaction.Actor.UserID != session.OpponentID {
		return responder.Reply(ctx, coinGameEphemeralError("你不是被邀請者，無法選擇接受!"))
	}
	if session.State != coinGameSessionPending {
		return responder.Reply(ctx, coinGameEphemeralError("很抱歉，這場遊戲已經開始了!"))
	}
	if _, err := m.game.Reserve(ctx, domain.CoinGameCommand{
		GuildID:      session.GuildID,
		ChallengerID: session.ChallengerID,
		OpponentID:   session.OpponentID,
		Wager:        session.Wager,
		Kind:         session.Kind,
	}); err != nil {
		if !errors.Is(err, ports.ErrCoinGameOpponent) && !errors.Is(err, ports.ErrCoinGameChallenger) {
			claim.Abandon()
			m.logCoinGameError(ctx, "coin game wager reserve failed", session, err)
		}
		return responder.UpdateMessage(ctx, coinGameBalanceErrorMessage(err))
	}
	session.OpponentName = interaction.Actor.Username
	session.State = coinGameSessionActive
	switch session.Kind {
	case domain.CoinGameKindHigherLower:
		return m.startHigherLower(ctx, responder, session, claim)
	case domain.CoinGameKindKnowledge:
		return m.startKnowledge(ctx, responder, session, claim)
	case domain.CoinGameKindBlackjack:
		return m.startBlackjack(ctx, responder, session, claim)
	default:
		return responder.UpdateMessage(ctx, coinGameErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
	}
}

func (m Module) startHigherLower(ctx context.Context, responder responses.Responder, session coinGameSession, claim *coinGameSessionClaim) error {
	challengerNumber := m.coinGameRandomInt(101)
	opponentNumber := m.coinGameRandomInt(101)
	challengerReturn := int64(0)
	opponentReturn := int64(0)
	result := ""
	switch {
	case challengerNumber < opponentNumber:
		opponentReturn = session.Wager * 2
		result = fmt.Sprintf("**裁判結果為:\n<@%s>輸了\n<@%s>取得賭注的**`2`**倍(共**`%d`**)**", session.ChallengerID, session.OpponentID, opponentReturn)
	case challengerNumber == opponentNumber:
		challengerReturn = session.Wager
		opponentReturn = session.Wager
		result = "**裁判結果為:\n平手，不加不減**"
	default:
		challengerReturn = session.Wager * 2
		result = fmt.Sprintf("**裁判結果為:\n<@%s>輸了\n<@%s>取得賭注的**`2`**倍(共**`%d`**)**", session.OpponentID, session.ChallengerID, challengerReturn)
	}
	if _, err := m.game.Settle(ctx, domain.CoinGameSettlementCommand{
		GuildID:          session.GuildID,
		ChallengerID:     session.ChallengerID,
		OpponentID:       session.OpponentID,
		ChallengerReturn: challengerReturn,
		OpponentReturn:   opponentReturn,
	}); err != nil {
		claim.Abandon()
		m.logCoinGameError(ctx, "coin game higher/lower settlement failed", session, err)
		return responder.UpdateMessage(ctx, coinGameErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
	}
	claim.Delete()
	return responder.UpdateMessage(ctx, responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:numberblocks:1044894385340416031> 比大小結果",
			Description: fmt.Sprintf("<@%s>的數字: %d\n<@%s>的數字: %d\n<:referee:1007236839524024340> %s\n",
				session.ChallengerID, challengerNumber, session.OpponentID, opponentNumber, result),
			Color: m.colorValue(),
		}},
		AllowedMentions: &responses.AllowedMentions{},
	})
}

func (m Module) startKnowledge(ctx context.Context, responder responses.Responder, session coinGameSession, claim *coinGameSessionClaim) error {
	if err := m.nextKnowledgeQuestion(&session); err != nil {
		claim.Abandon()
		m.logCoinGameError(ctx, "coin game knowledge question load failed", session, err)
		return responder.UpdateMessage(ctx, coinGameErrorMessage("很抱歉，題庫載入失敗，請重試!"))
	}
	claim.Commit(session)
	m.scheduleCoinGameTimeout(session)
	return responder.UpdateMessage(ctx, knowledgeQuestionMessage(session, m.colorValue()))
}

func (m Module) handleKnowledgeAnswer(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, session coinGameSession, claim *coinGameSessionClaim) error {
	if session.State != coinGameSessionActive {
		return responder.Reply(ctx, coinGameEphemeralError("很抱歉，這場遊戲還沒開始!"))
	}
	answer := strings.TrimSpace(interaction.CustomID)
	if answer == "" {
		return responder.Reply(ctx, coinGameEphemeralError("很抱歉，出現了未知的錯誤，請重試!"))
	}
	isChallenger := interaction.Actor.UserID == session.ChallengerID
	if isChallenger && session.ChallengerChoice != "" || !isChallenger && session.OpponentChoice != "" {
		return responder.Reply(ctx, coinGameEphemeralError("你已經選取過了!!!"))
	}
	points := int64(0)
	if answer == session.KnowledgeQuestion.Answer {
		points = m.knowledgePoints(session)
	}
	if isChallenger {
		session.ChallengerChoice = answer
		session.ChallengerScore += points
	} else {
		session.OpponentChoice = answer
		session.OpponentScore += points
	}
	reply := knowledgeAnswerReply(answer, session.KnowledgeQuestion.Answer, points)
	if session.ChallengerChoice == "" || session.OpponentChoice == "" {
		claim.Commit(session)
		return responder.Reply(ctx, reply)
	}
	session.KnowledgeRound++
	if session.KnowledgeRound >= 5 {
		return m.finishKnowledge(ctx, responder, session, reply, claim)
	}
	if err := m.nextKnowledgeQuestion(&session); err != nil {
		claim.Abandon()
		m.cancelCoinGameTimeout(session.ID)
		m.logCoinGameError(ctx, "coin game knowledge question load failed", session, err)
		return responder.UpdateMessage(ctx, coinGameErrorMessage("很抱歉，題庫載入失敗，請重試!"))
	}
	claim.Commit(session)
	m.scheduleCoinGameTimeout(session)
	if err := responder.UpdateMessage(ctx, knowledgeQuestionMessage(session, m.colorValue())); err != nil {
		return err
	}
	return responder.FollowUp(ctx, reply)
}

func (m Module) finishKnowledge(ctx context.Context, responder responses.Responder, session coinGameSession, reply responses.Message, claim *coinGameSessionClaim) error {
	challengerReturn, opponentReturn := coinGameCompareReturns(session.Wager, session.ChallengerScore, session.OpponentScore)
	if _, err := m.game.Settle(ctx, domain.CoinGameSettlementCommand{
		GuildID:          session.GuildID,
		ChallengerID:     session.ChallengerID,
		OpponentID:       session.OpponentID,
		ChallengerReturn: challengerReturn,
		OpponentReturn:   opponentReturn,
	}); err != nil {
		claim.Abandon()
		m.cancelCoinGameTimeout(session.ID)
		m.logCoinGameError(ctx, "coin game knowledge settlement failed", session, err)
		return responder.UpdateMessage(ctx, coinGameErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
	}
	claim.Delete()
	m.cancelCoinGameTimeout(session.ID)
	if err := responder.UpdateMessage(ctx, knowledgeFinalMessage(session, m.colorValue())); err != nil {
		return err
	}
	return responder.FollowUp(ctx, reply)
}

func (m Module) nextKnowledgeQuestion(session *coinGameSession) error {
	questions, err := legacyKnowledgeQuestions()
	if err != nil || len(questions) == 0 {
		return err
	}
	question := questions[m.coinGameRandomInt(len(questions))]
	answers := append([]string(nil), question.Unanswers...)
	for i := len(answers) - 1; i > 0; i-- {
		j := m.coinGameRandomInt(i + 1)
		answers[i], answers[j] = answers[j], answers[i]
	}
	session.KnowledgeQuestion = question
	session.KnowledgeAnswers = answers
	now := m.clockNow()
	session.QuestionStartedAt = now
	session.TurnStartedAt = now
	session.TurnDeadline = now.Add(coinGameKnowledgeTTL)
	session.TurnGeneration++
	session.ChallengerChoice = ""
	session.OpponentChoice = ""
	return nil
}

func (m Module) knowledgePoints(session coinGameSession) int64 {
	elapsed := int64(m.clockNow().Sub(session.QuestionStartedAt).Seconds())
	remaining := int64(20) - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return remaining * 50
}

func (m Module) startBlackjack(ctx context.Context, responder responses.Responder, session coinGameSession, claim *coinGameSessionClaim) error {
	session.Deck = legacyBlackjackDeck()
	for blackjackSum(session.DealerCards) < 13 {
		session.DealerCards = append(session.DealerCards, m.drawBlackjackCard(&session))
	}
	session.ChallengerCards = append(session.ChallengerCards, m.drawBlackjackCard(&session))
	session.OpponentCards = append(session.OpponentCards, m.drawBlackjackCard(&session))
	m.beginBlackjackTurn(&session, session.ChallengerID)
	claim.Commit(session)
	m.scheduleCoinGameTimeout(session)
	return responder.UpdateMessage(ctx, blackjackTurnMessage(session, true, m.colorValue()))
}

func (m Module) showBlackjackCards(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, session coinGameSession) error {
	if session.Kind != domain.CoinGameKindBlackjack || session.State != coinGameSessionActive {
		return responder.Reply(ctx, coinGameEphemeralError("很抱歉，這場遊戲還沒開始!"))
	}
	cards := session.OpponentCards
	if interaction.Actor.UserID == session.ChallengerID {
		cards = session.ChallengerCards
	}
	return responder.Reply(ctx, responses.Message{
		Embeds: []responses.Embed{{
			Title:       fmt.Sprintf("共:`%d`點", blackjackSum(cards)),
			Description: blackjackCards(cards),
			Color:       m.colorValue(),
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	})
}

func (m Module) handleBlackjackAction(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, session coinGameSession, claim *coinGameSessionClaim) error {
	if session.Kind != domain.CoinGameKindBlackjack || session.State != coinGameSessionActive {
		return responder.Reply(ctx, coinGameEphemeralError("很抱歉，這場遊戲還沒開始!"))
	}
	challengerAction := interaction.CustomID == "main_get_card" || interaction.CustomID == "main_no_card"
	if challengerAction && interaction.Actor.UserID != session.ChallengerID || !challengerAction && interaction.Actor.UserID != session.OpponentID {
		return responder.Reply(ctx, coinGameEphemeralError("還沒輪到你啦!!"))
	}
	expectedTurn := session.ChallengerID
	if !challengerAction {
		expectedTurn = session.OpponentID
	}
	if session.BlackjackTurn != expectedTurn {
		return responder.Reply(ctx, coinGameEphemeralError("還沒輪到你啦!!"))
	}
	hit := strings.HasSuffix(interaction.CustomID, "get_card")
	drawn := 0
	if challengerAction {
		session.ChallengerHit = boolPtr(hit)
		if hit {
			drawn = m.drawBlackjackCard(&session)
			session.ChallengerCards = append(session.ChallengerCards, drawn)
		}
		m.beginBlackjackTurn(&session, session.OpponentID)
		claim.Commit(session)
		m.scheduleCoinGameTimeout(session)
		if err := responder.UpdateMessage(ctx, blackjackTurnMessage(session, false, m.colorValue())); err != nil {
			return err
		}
		return responder.FollowUp(ctx, blackjackActionReply(hit, drawn, "略過"))
	}
	session.OpponentHit = boolPtr(hit)
	if hit {
		drawn = m.drawBlackjackCard(&session)
		session.OpponentCards = append(session.OpponentCards, drawn)
	}
	if blackjackShouldFinish(session) {
		return m.finishBlackjack(ctx, responder, session, blackjackActionReply(hit, drawn, "不抽獎"), claim)
	}
	m.beginBlackjackTurn(&session, session.ChallengerID)
	claim.Commit(session)
	m.scheduleCoinGameTimeout(session)
	if err := responder.UpdateMessage(ctx, blackjackTurnMessage(session, true, m.colorValue())); err != nil {
		return err
	}
	return responder.FollowUp(ctx, blackjackActionReply(hit, drawn, "不抽獎"))
}

func (m Module) finishBlackjack(ctx context.Context, responder responses.Responder, session coinGameSession, reply responses.Message, claim *coinGameSessionClaim) error {
	challengerReturn, opponentReturn := blackjackReturns(session)
	if _, err := m.game.Settle(ctx, domain.CoinGameSettlementCommand{
		GuildID:          session.GuildID,
		ChallengerID:     session.ChallengerID,
		OpponentID:       session.OpponentID,
		ChallengerReturn: challengerReturn,
		OpponentReturn:   opponentReturn,
	}); err != nil {
		claim.Abandon()
		m.cancelCoinGameTimeout(session.ID)
		m.logCoinGameError(ctx, "coin game blackjack settlement failed", session, err)
		return responder.UpdateMessage(ctx, coinGameErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
	}
	claim.Delete()
	m.cancelCoinGameTimeout(session.ID)
	if err := responder.UpdateMessage(ctx, blackjackFinalMessage(session, challengerReturn, opponentReturn, m.colorValue())); err != nil {
		return err
	}
	return responder.FollowUp(ctx, reply)
}

func (m Module) beginBlackjackTurn(session *coinGameSession, userID string) {
	now := m.clockNow()
	session.BlackjackTurn = userID
	session.TurnStartedAt = now
	session.TurnDeadline = now.Add(coinGameBlackjackTTL)
	session.TurnGeneration++
}

func (m Module) scheduleCoinGameTimeout(session coinGameSession) {
	if m.gameTimeouts == nil || session.ID == "" || session.TurnGeneration == 0 || session.TurnDeadline.IsZero() {
		return
	}
	m.gameTimeouts.Schedule(session.ID, session.TurnGeneration, session.TurnDeadline, func(ctx context.Context) {
		m.handleCoinGameTimeout(ctx, session.ID, session.TurnGeneration)
	})
}

func (m Module) cancelCoinGameTimeout(sessionID string) {
	if m.gameTimeouts != nil {
		m.gameTimeouts.Cancel(sessionID)
	}
}

func (m Module) handleCoinGameTimeout(ctx context.Context, sessionID string, generation uint64) {
	claim, status := m.gameSessions.ClaimForTimeout(sessionID, generation)
	switch status {
	case coinGameTimeoutClaimBusy, coinGameTimeoutClaimNotDue:
		m.gameTimeouts.Schedule(sessionID, generation, m.clockNow().Add(coinGameTimeoutRetry), func(ctx context.Context) {
			m.handleCoinGameTimeout(ctx, sessionID, generation)
		})
		return
	case coinGameTimeoutClaimed:
	default:
		return
	}
	session := claim.Session()
	challengerReturn, opponentReturn := coinGameTimeoutReturns(session)
	if _, err := m.game.Settle(ctx, domain.CoinGameSettlementCommand{
		GuildID:          session.GuildID,
		ChallengerID:     session.ChallengerID,
		OpponentID:       session.OpponentID,
		ChallengerReturn: challengerReturn,
		OpponentReturn:   opponentReturn,
	}); err != nil {
		claim.Abandon()
		m.cancelCoinGameTimeout(session.ID)
		m.logCoinGameError(ctx, "coin game timeout settlement failed", session, err)
		return
	}
	claim.Delete()
	m.cancelCoinGameTimeout(session.ID)
	if m.messages == nil || session.ChannelID == "" || session.MessageID == "" {
		return
	}
	if err := m.messages.EditMessage(ctx, ports.MessageRef{ChannelID: session.ChannelID, MessageID: session.MessageID}, coinGameTimeoutOutbound(session)); err != nil {
		m.logCoinGameError(ctx, "coin game timeout message edit failed", session, err)
	}
}

func (m Module) logCoinGameError(ctx context.Context, message string, session coinGameSession, err error) {
	logger := m.logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.ErrorContext(ctx, message, "session_id", session.ID, "guild_id", session.GuildID, "kind", session.Kind, "error", err)
}

func coinGameTimeoutReturns(session coinGameSession) (int64, int64) {
	switch session.Kind {
	case domain.CoinGameKindKnowledge:
		challengerReturn := int64(0)
		opponentReturn := int64(0)
		if session.ChallengerChoice != "" {
			challengerReturn = session.Wager * 2
		}
		if session.OpponentChoice != "" {
			opponentReturn = session.Wager * 2
		}
		return challengerReturn, opponentReturn
	case domain.CoinGameKindBlackjack:
		if session.BlackjackTurn == session.ChallengerID {
			return 0, session.Wager * 2
		}
		return session.Wager * 2, 0
	default:
		return 0, 0
	}
}

func coinGameTimeoutOutbound(session coinGameSession) ports.OutboundMessage {
	content := "<:idea:1007312008179351624> **| 知識王**"
	title := knowledgeTimeoutTitle(session)
	if session.Kind == domain.CoinGameKindBlackjack {
		content = fmt.Sprintf("這回合是<@%s>的，另一位只能查看牌組喔!", session.BlackjackTurn)
		title = blackjackTimeoutTitle(session)
	}
	return ports.OutboundMessage{
		Content: content,
		Embeds: []ports.OutboundEmbed{{
			Title: title,
			Color: coinGameErrorColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func knowledgeTimeoutTitle(session coinGameSession) string {
	names := make([]string, 0, 2)
	if session.OpponentChoice == "" {
		names = append(names, coinGamePlayerName(session.OpponentName, session.OpponentID))
	}
	if session.ChallengerChoice == "" {
		names = append(names, coinGamePlayerName(session.ChallengerName, session.ChallengerID))
	}
	return fmt.Sprintf("<a:error:980086028113182730> | %s 超過回應時間，自動判定棄賽!", strings.Join(names, " "))
}

func blackjackTimeoutTitle(session coinGameSession) string {
	name := session.OpponentName
	if session.BlackjackTurn == session.ChallengerID {
		name = session.ChallengerName
	}
	return fmt.Sprintf("<a:error:980086028113182730> | `%s` 超過回應時間，自動判定棄賽!", coinGamePlayerName(name, session.BlackjackTurn))
}

func coinGamePlayerName(name string, userID string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	return "<@" + strings.TrimSpace(userID) + ">"
}

func (m Module) drawBlackjackCard(session *coinGameSession) int {
	if len(session.Deck) == 0 {
		session.Deck = legacyBlackjackDeck()
	}
	index := m.coinGameRandomInt(len(session.Deck))
	card := session.Deck[index]
	session.Deck = append(session.Deck[:index], session.Deck[index+1:]...)
	return card
}

func (m Module) coinGameRandomInt(maxExclusive int) int {
	if maxExclusive <= 0 {
		return 0
	}
	random := m.gameRandInt
	if random == nil {
		random = legacyRandomInt
	}
	value := random(maxExclusive)
	if value < 0 {
		value = -value
	}
	return value % maxExclusive
}

func (m Module) colorValue() int {
	if m.color == nil {
		return legacyRandomColor()
	}
	return m.color()
}

func (m Module) clockNow() time.Time {
	if m.clock == nil {
		return time.Now()
	}
	return m.clock.Now()
}

func coinGameInviteMessage(session coinGameSession) responses.Message {
	return responses.Message{
		Content: fmt.Sprintf("<@%s>", session.OpponentID),
		Embeds: []responses.Embed{{
			Title:       coinGameTitle(session.Kind),
			Description: fmt.Sprintf("<@%s>**邀請<@%s>玩%s\n將會消耗你**`%d`**進行賭注\n是否願意?\n<a:warn:1000814885506129990> 一但同意如中途放棄則視為敗北**", session.ChallengerID, session.OpponentID, session.Kind, session.Wager),
			Footer:      &responses.EmbedFooter{Text: "請於30秒內回覆，如無回復則視為拒絕"},
			Color:       0,
		}},
		Components:      coinGameInviteRows(session.Kind, false),
		AllowedMentions: &responses.AllowedMentions{UserIDs: []string{session.OpponentID}},
	}
}

func coinGameDisableInviteMessage(session coinGameSession) responses.Message {
	msg := coinGameInviteMessage(session)
	msg.Components = coinGameInviteRows(session.Kind, true)
	return msg
}

func coinGameInviteRows(kind domain.CoinGameKind, disabled bool) []responses.ComponentRow {
	components := []responses.Component{
		{Type: responses.ComponentTypeButton, CustomID: "yesssss", Label: "點我接受遊玩", Emoji: "<:halloween_yes:1005480105642041354>", Style: responses.ButtonStyleSuccess, Disabled: disabled},
		{Type: responses.ComponentTypeButton, CustomID: "nooooo", Label: "點我拒絕遊玩", Emoji: "<a:YuiHeadShake:1005480366167040021>", Style: responses.ButtonStyleDanger, Disabled: disabled},
	}
	if kind == domain.CoinGameKindBlackjack {
		components = append(components, responses.Component{Type: responses.ComponentTypeButton, CustomID: "teach21point", Label: "甚麼是21點", Emoji: "<:question:997374195229003776>", Style: responses.ButtonStyleSecondary, Disabled: disabled})
	}
	if kind == domain.CoinGameKindHigherLower {
		components = append(components, responses.Component{Type: responses.ComponentTypeButton, CustomID: "thansize", Label: "甚麼是比大小", Emoji: "<:question:997374195229003776>", Style: responses.ButtonStyleSecondary, Disabled: disabled})
	}
	return []responses.ComponentRow{{Components: components}}
}

func coinGameTitle(kind domain.CoinGameKind) string {
	switch kind {
	case domain.CoinGameKindBlackjack:
		return "<:blackjack:1005469849180459131> 21點小遊戲"
	case domain.CoinGameKindKnowledge:
		return "<:idea:1007312008179351624> 知識王"
	default:
		return "<:numberblocks:1044894385340416031> 比大小"
	}
}

func coinGameBalanceErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrCoinGameOpponent):
		return coinGameErrorMessage("對方沒有這麼多代幣可以玩喔!!")
	case errors.Is(err, ports.ErrCoinGameChallenger):
		return coinGameErrorMessage("你沒有這麼多代幣可以玩喔!!")
	default:
		return coinGameErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func coinGameErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:error:980086028113182730> | " + content,
			Color: coinGameErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func coinGameEphemeralError(content string) responses.Message {
	msg := coinGameErrorMessage(content)
	msg.Embeds[0].Title = "<a:Discord_AnimatedNo:1015989839809757295> | " + content
	msg.Ephemeral = true
	return msg
}

func coinGameEphemeralText(content string) responses.Message {
	return responses.Message{Content: content, Ephemeral: true, AllowedMentions: &responses.AllowedMentions{}}
}

func knowledgeQuestionMessage(session coinGameSession, color int) responses.Message {
	var builder strings.Builder
	if session.KnowledgeRound > 0 || session.ChallengerScore > 0 || session.OpponentScore > 0 {
		fmt.Fprintf(&builder, "**<:speedometer:1007522466995912734><@%s> 目前得分:**`%d`**\n<:speedometer:1007522466995912734> <@%s> 目前得分:**`%d`**\n", session.ChallengerID, session.ChallengerScore, session.OpponentID, session.OpponentScore)
	}
	fmt.Fprintf(&builder, "**類型:**`%s`**\n<:q_:1007244629923598377> 題目:\n%s\n\n", session.KnowledgeQuestion.Type, session.KnowledgeQuestion.Question)
	emojis := []string{"<:lettera:1007246307674570753>", "<:letterb:1007245313758740530>", "<:c_:1007245311695126588>", "<:d1:1007245309719625788>"}
	for i, answer := range session.KnowledgeAnswers {
		fmt.Fprintf(&builder, "%s %s\n", emojis[i], answer)
	}
	displayDeadline := session.QuestionStartedAt.Add(15 * time.Second).Unix()
	fmt.Fprintf(&builder, "\n<a:warn:1000814885506129990> 請於<t:%d:R>選擇，超過時間則視為棄賽**", displayDeadline)
	buttons := make([]responses.Component, 0, len(session.KnowledgeAnswers))
	for i, answer := range session.KnowledgeAnswers {
		buttons = append(buttons, responses.Component{Type: responses.ComponentTypeButton, CustomID: answer, Emoji: emojis[i], Style: responses.ButtonStylePrimary})
	}
	return responses.Message{
		Content: "<:idea:1007312008179351624> **| 知識王**",
		Embeds: []responses.Embed{{
			Title:       "<:startbutton1:1005838813274325022> 遊戲已開始",
			Description: builder.String(),
			Color:       color,
		}},
		Components:      []responses.ComponentRow{{Components: buttons}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func knowledgeAnswerReply(answer string, correct string, points int64) responses.Message {
	if answer == correct {
		return responses.Message{
			Embeds: []responses.Embed{{
				Title:       "<a:green_tick:994529015652163614> | 你選擇了正確的答案:" + correct,
				Description: fmt.Sprintf("根據你選取的正確以及時間給予你`%d`分", points),
				Color:       coinGameSuccessColor,
			}},
			Ephemeral:       true,
			AllowedMentions: &responses.AllowedMentions{},
		}
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:error:980086028113182730> | 你選擇了錯誤的答案，正確答案:" + correct,
			Description: fmt.Sprintf("根據你選取的正確以及時間給予你`%d`分", points),
			Color:       coinGameErrorColor,
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func knowledgeFinalMessage(session coinGameSession, color int) responses.Message {
	result := fmt.Sprintf("<@%s>獲勝!", session.OpponentID)
	if session.ChallengerScore > session.OpponentScore {
		result = fmt.Sprintf("<@%s>獲勝!", session.ChallengerID)
	} else if session.ChallengerScore == session.OpponentScore {
		result = "平手，不加也不減**"
	}
	gain := ""
	if session.ChallengerScore != session.OpponentScore {
		gain = fmt.Sprintf("\n取得:`%d`", session.Wager*2)
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: fmt.Sprintf("<a:green_tick:994529015652163614> **| 遊戲已結束**"),
			Description: fmt.Sprintf("**\n<:businesscreditscore:1007236532421275688> 開始進行結算!\n<:speedometer:1007522466995912734> <@%s>得分:**`%d`**\n<:speedometer:1007522466995912734> <@%s>得分:**`%d`**\n\n**<:referee:1007236839524024340>裁判結果:\n%s%s",
				session.ChallengerID, session.ChallengerScore, session.OpponentID, session.OpponentScore, result, gain),
			Color: color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func blackjackTurnMessage(session coinGameSession, challengerTurn bool, color int) responses.Message {
	target := session.ChallengerID
	components := blackjackMainRows(false)
	if !challengerTurn {
		target = session.OpponentID
		components = blackjackUserRows(false)
	}
	displayDeadline := session.TurnStartedAt.Add(30 * time.Second).Unix()
	return responses.Message{
		Content: fmt.Sprintf("這回合是<@%s>的，另一位只能查看牌組喔!", target),
		Embeds: []responses.Embed{{
			Title:       "<:startbutton1:1005838813274325022> 遊戲已開始",
			Description: fmt.Sprintf("\n**已為各位各發一張牌\n請選擇要抽牌還是不抽\n<a:warn:1000814885506129990>請於<t:%d:R>選擇，超過時間則視為棄賽(你的賭注會全輸)**", displayDeadline),
			Color:       color,
		}},
		Components:      components,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func blackjackMainRows(disabled bool) []responses.ComponentRow {
	return []responses.ComponentRow{{Components: []responses.Component{
		{Type: responses.ComponentTypeButton, CustomID: "main_no_card", Label: "略過", Emoji: "<a:YuiHeadShake:1005480366167040021>", Style: responses.ButtonStyleDanger, Disabled: disabled},
		{Type: responses.ComponentTypeButton, CustomID: "main_get_card", Label: "抽牌", Emoji: "<:playingcard:1006058772634009700>", Style: responses.ButtonStyleSuccess, Disabled: disabled},
		{Type: responses.ComponentTypeButton, CustomID: "lookmenumber", Label: "查看我的牌", Emoji: "<:searching:986107902777491497>", Style: responses.ButtonStylePrimary, Disabled: disabled},
	}}}
}

func blackjackUserRows(disabled bool) []responses.ComponentRow {
	return []responses.ComponentRow{{Components: []responses.Component{
		{Type: responses.ComponentTypeButton, CustomID: "user_no_card", Label: "略過", Emoji: "<a:YuiHeadShake:1005480366167040021>", Style: responses.ButtonStyleDanger, Disabled: disabled},
		{Type: responses.ComponentTypeButton, CustomID: "user_get_card", Label: "抽牌", Emoji: "<:playingcard:1006058772634009700>", Style: responses.ButtonStyleSuccess, Disabled: disabled},
		{Type: responses.ComponentTypeButton, CustomID: "lookmenumber", Label: "查看我的牌", Emoji: "<:searching:986107902777491497>", Style: responses.ButtonStylePrimary, Disabled: disabled},
	}}}
}

func blackjackActionReply(hit bool, card int, standText string) responses.Message {
	title := fmt.Sprintf("<a:green_tick:994529015652163614> | 你選擇了%s", standText)
	if hit {
		title = fmt.Sprintf("<a:green_tick:994529015652163614> | 你抽到了: %s", blackjackNumberEmoji(card))
	}
	return responses.Message{
		Embeds:          []responses.Embed{{Title: title, Color: coinGameSuccessColor}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func blackjackFinalMessage(session coinGameSession, challengerReturn int64, opponentReturn int64, color int) responses.Message {
	content := blackjackResultText(session, challengerReturn, opponentReturn)
	return responses.Message{
		Content: "<a:green_tick:994529015652163614> **| 遊戲已結束!**",
		Embeds: []responses.Embed{{
			Title: "<:startbutton1:1005838813274325022> 21點小遊戲",
			Description: fmt.Sprintf("**<:businesscreditscore:1007236532421275688> 遊戲已經結束，現在開始結算成績**\n\n**莊家總共:**\n%s**(共**`%d`**點)**\n**<@%s>總共:**\n%s**(共**`%d`**點)**\n**<@%s>總共:**\n%s**(共**`%d`**點)**\n\n<:referee:1007236839524024340> %s",
				blackjackCards(session.DealerCards), blackjackSum(session.DealerCards), session.ChallengerID, blackjackCards(session.ChallengerCards), blackjackSum(session.ChallengerCards), session.OpponentID, blackjackCards(session.OpponentCards), blackjackSum(session.OpponentCards), content),
			Color: color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func blackjackResultText(session coinGameSession, challengerReturn int64, opponentReturn int64) string {
	challenger := blackjackSum(session.ChallengerCards)
	opponent := blackjackSum(session.OpponentCards)
	switch {
	case challenger > 21 && opponent > 21:
		return "**裁判結果為:\n兩個都爆，不加不減**"
	case challenger > 21:
		return fmt.Sprintf("**裁判結果為:\n<@%s>爆了\n<@%s>取得賭注的**`2`**倍(共**`%d`**)**", session.ChallengerID, session.OpponentID, opponentReturn)
	case opponent > 21:
		return fmt.Sprintf("**裁判結果為:\n<@%s>爆了\n<@%s>取得賭注的**`2`**倍(共**`%d`**)**", session.OpponentID, session.ChallengerID, challengerReturn)
	case challenger > opponent:
		return fmt.Sprintf("**裁判結果為:\n<@%s>大於<@%s>\n<@%s>取得賭注的**`2`**倍(共**`%d`**)**", session.ChallengerID, session.OpponentID, session.ChallengerID, challengerReturn)
	case challenger < opponent:
		return fmt.Sprintf("**裁判結果為:\n<@%s>大於<@%s>\n<@%s>取得賭注的**`2`**倍**(共**`%d`**)**", session.OpponentID, session.ChallengerID, session.OpponentID, opponentReturn)
	default:
		return fmt.Sprintf("**裁判結果為:\n<@%s>等於<@%s>\n平手，因此不加也不減**", session.ChallengerID, session.OpponentID)
	}
}

func blackjackReturns(session coinGameSession) (int64, int64) {
	challenger := blackjackSum(session.ChallengerCards)
	opponent := blackjackSum(session.OpponentCards)
	return coinGameCompareBlackjackReturns(session.Wager, challenger, opponent)
}

func coinGameCompareBlackjackReturns(wager int64, challenger int, opponent int) (int64, int64) {
	switch {
	case challenger > 21 && opponent > 21:
		return wager, wager
	case challenger > 21:
		return 0, wager * 2
	case opponent > 21:
		return wager * 2, 0
	case challenger > opponent:
		return wager * 2, 0
	case challenger < opponent:
		return 0, wager * 2
	default:
		return wager, wager
	}
}

func coinGameCompareReturns(wager int64, challenger int64, opponent int64) (int64, int64) {
	switch {
	case challenger > opponent:
		return wager * 2, 0
	case challenger < opponent:
		return 0, wager * 2
	default:
		return wager, wager
	}
}

func blackjackShouldFinish(session coinGameSession) bool {
	challenger := blackjackSum(session.ChallengerCards)
	opponent := blackjackSum(session.OpponentCards)
	bothStand := session.ChallengerHit != nil && session.OpponentHit != nil && !*session.ChallengerHit && !*session.OpponentHit
	return bothStand || challenger > 21 || opponent > 21 || (challenger > 21 && opponent > 21)
}

func legacyBlackjackDeck() []int {
	deck := make([]int, 0, 40)
	for suit := 0; suit < 4; suit++ {
		for value := 1; value <= 10; value++ {
			deck = append(deck, value)
		}
	}
	return deck
}

func blackjackSum(cards []int) int {
	total := 0
	for _, card := range cards {
		total += card
	}
	return total
}

func blackjackCards(cards []int) string {
	values := make([]string, 0, len(cards))
	for _, card := range cards {
		values = append(values, blackjackNumberEmoji(card))
	}
	return strings.Join(values, ",")
}

func blackjackNumberEmoji(card int) string {
	switch card {
	case 1:
		return "<:numberone:1005471516407906324>"
	case 2:
		return "<:number2:1005471518018510950>"
	case 3:
		return "<:number3:1005471519574597672>"
	case 4:
		return "<:numberfour:1005471521147473950>"
	case 5:
		return "<:number5:1005471522649022517>"
	case 6:
		return "<:six:1005471524721020948>"
	case 7:
		return "<:seven:1005471526222581760>"
	case 8:
		return "<:number8:1005471527891898398>"
	case 9:
		return "<:number9:1005471529699655780>"
	default:
		return "<:number10:1005471531377360957>"
	}
}

func boolPtr(value bool) *bool {
	return &value
}

var (
	legacyKnowledgeOnce           sync.Once
	legacyKnowledgeQuestionsCache []knowledgeQuestion
	legacyKnowledgeErr            error
)

func legacyKnowledgeQuestions() ([]knowledgeQuestion, error) {
	legacyKnowledgeOnce.Do(func() {
		var payload map[string][]knowledgeQuestion
		if err := json.Unmarshal(legacyTopicJSON, &payload); err != nil {
			legacyKnowledgeErr = err
			return
		}
		for _, question := range payload["1"] {
			if question.Question == "" || question.Answer == "" || len(question.Unanswers) != 4 {
				continue
			}
			legacyKnowledgeQuestionsCache = append(legacyKnowledgeQuestionsCache, question)
		}
		if len(legacyKnowledgeQuestionsCache) == 0 {
			legacyKnowledgeErr = errors.New("legacy knowledge topic bank is empty")
		}
	})
	return legacyKnowledgeQuestionsCache, legacyKnowledgeErr
}
