package domain

import (
	"errors"
	"math"
	"strings"
	"time"
)

var ErrInvalidEconomyQuery = errors.New("invalid economy query")
var ErrInvalidSignIn = errors.New("invalid sign in")
var ErrInvalidEconomySettings = errors.New("invalid economy settings")
var ErrInvalidCoinAdminCommand = errors.New("invalid coin admin command")
var ErrInvalidCoinRankQuery = errors.New("invalid coin rank query")
var ErrInvalidCoinResetCommand = errors.New("invalid coin reset command")
var ErrInvalidEconomyProfileQuery = errors.New("invalid economy profile query")
var ErrInvalidRockPaperScissorsCommand = errors.New("invalid rock paper scissors command")
var ErrInvalidShopItem = errors.New("invalid shop item")
var ErrInvalidShopPurchase = errors.New("invalid shop purchase")
var ErrInvalidCoinGameCommand = errors.New("invalid coin game command")

type CoinAdminOperation string
type RockPaperScissorsChoice string
type RockPaperScissorsOutcome string
type CoinGameKind string

const (
	CoinAdminOperationAdd    CoinAdminOperation = "add"
	CoinAdminOperationReduce CoinAdminOperation = "reduce"

	RockPaperScissorsChoiceScissors RockPaperScissorsChoice = "剪刀"
	RockPaperScissorsChoiceRock     RockPaperScissorsChoice = "石頭"
	RockPaperScissorsChoicePaper    RockPaperScissorsChoice = "布"

	RockPaperScissorsOutcomeWin  RockPaperScissorsOutcome = "win"
	RockPaperScissorsOutcomeLoss RockPaperScissorsOutcome = "loss"
	RockPaperScissorsOutcomeTie  RockPaperScissorsOutcome = "tie"

	CoinGameKindBlackjack   CoinGameKind = "21點"
	CoinGameKindKnowledge   CoinGameKind = "知識王"
	CoinGameKindHigherLower CoinGameKind = "比大小"
)

type CoinBalance struct {
	GuildID string
	UserID  string
	Coins   int64
	Today   int64
}

type EconomyConfig struct {
	GuildID     string
	GachaCost   int64
	SignCoins   int64
	ChannelID   string
	XPMultiple  float64
	ResetMarker int64
}

func (c EconomyConfig) EffectiveGachaCost() int64 {
	if c.GachaCost <= 0 {
		return 500
	}
	return c.GachaCost
}

type SignCalendar struct {
	GuildID string
	UserID  string
	Date    map[string]map[string][]string
}

func (c SignCalendar) HasDay(year string, month string, day string) bool {
	months, ok := c.Date[year]
	if !ok {
		return false
	}
	days, ok := months[month]
	if !ok {
		return false
	}
	for _, signedDay := range days {
		if signedDay == day {
			return true
		}
	}
	return false
}

type SignInCommand struct {
	GuildID string
	UserID  string
	Now     time.Time
	Year    string
	Month   string
	Day     string
}

type SignInResult struct {
	Balance     CoinBalance
	Config      EconomyConfig
	Calendar    SignCalendar
	Reward      int64
	ConfigFound bool
	SignedAt    time.Time
}

type SignInListEntry struct {
	UserID       string
	SignedAtUnix int64
	ShowSignedAt bool
}

type SignInListResult struct {
	GuildID       string
	ActorUserID   string
	Entries       []SignInListEntry
	RollingWindow bool
}

type EconomySettingsCommand struct {
	GuildID           string
	GachaCost         int64
	SignCooldownHours int64
	SignCoins         int64
	NotificationID    string
	XPMultiple        float64
}

type CoinAdminCommand struct {
	GuildID   string
	UserID    string
	Operation CoinAdminOperation
	Amount    int64
}

type CoinAdminResult struct {
	Balance CoinBalance
	Delta   int64
	Created bool
}

type CoinResetCommand struct {
	GuildID string
	Divisor int64
}

type CoinResetResult struct {
	GuildID       string
	Divisor       int64
	AffectedCount int64
	Deleted       bool
}

type RockPaperScissorsCommand struct {
	GuildID        string
	UserID         string
	Wager          int64
	PlayerChoice   RockPaperScissorsChoice
	ComputerChoice RockPaperScissorsChoice
}

type RockPaperScissorsResult struct {
	Balance         CoinBalance
	PreviousBalance int64
	Delta           int64
	Outcome         RockPaperScissorsOutcome
	PlayerChoice    RockPaperScissorsChoice
	ComputerChoice  RockPaperScissorsChoice
}

type ShopItem struct {
	GuildID     string
	CommodityID int64
	Name        string
	NeedCoins   int64
	Description string
	Code        string
	AutoDelete  bool
	RoleID      string
	Count       int64
}

type ShopPurchaseCommand struct {
	GuildID     string
	UserID      string
	CommodityID int64
	Quantity    int64
}

type ShopPurchaseResult struct {
	Item            ShopItem
	Quantity        int64
	TotalCost       int64
	PreviousBalance int64
	Balance         CoinBalance
}

type CoinGameCommand struct {
	GuildID      string
	ChallengerID string
	OpponentID   string
	Wager        int64
	Kind         CoinGameKind
}

type CoinGameBalanceResult struct {
	Challenger CoinBalance
	Opponent   CoinBalance
	Wager      int64
}

type CoinGameSettlementCommand struct {
	GuildID          string
	ChallengerID     string
	OpponentID       string
	ChallengerReturn int64
	OpponentReturn   int64
}

type CoinGameSettlementResult struct {
	Challenger CoinBalance
	Opponent   CoinBalance
}

func (c CoinAdminCommand) Normalize() CoinAdminCommand {
	return CoinAdminCommand{
		GuildID:   strings.TrimSpace(c.GuildID),
		UserID:    strings.TrimSpace(c.UserID),
		Operation: CoinAdminOperation(strings.TrimSpace(string(c.Operation))),
		Amount:    c.Amount,
	}
}

func (c CoinAdminCommand) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.UserID == "" || c.Amount <= 0 {
		return ErrInvalidCoinAdminCommand
	}
	switch c.Operation {
	case CoinAdminOperationAdd, CoinAdminOperationReduce:
		return nil
	default:
		return ErrInvalidCoinAdminCommand
	}
}

func (c CoinAdminCommand) SignedDelta() (int64, error) {
	c = c.Normalize()
	if err := c.Validate(); err != nil {
		return 0, err
	}
	if c.Operation == CoinAdminOperationReduce {
		return -c.Amount, nil
	}
	return c.Amount, nil
}

func (c CoinResetCommand) Normalize() CoinResetCommand {
	return CoinResetCommand{
		GuildID: strings.TrimSpace(c.GuildID),
		Divisor: c.Divisor,
	}
}

func (c CoinResetCommand) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" {
		return ErrInvalidCoinResetCommand
	}
	return nil
}

// LegacyJavaScriptRound preserves JavaScript Math.round semantics, including negative half values.
func LegacyJavaScriptRound(value float64) int64 {
	return int64(math.Floor(value + 0.5))
}

func (c RockPaperScissorsChoice) Normalize() RockPaperScissorsChoice {
	return RockPaperScissorsChoice(strings.TrimSpace(string(c)))
}

func (c RockPaperScissorsChoice) Valid() bool {
	switch c.Normalize() {
	case RockPaperScissorsChoiceScissors, RockPaperScissorsChoiceRock, RockPaperScissorsChoicePaper:
		return true
	default:
		return false
	}
}

func (c RockPaperScissorsCommand) Normalize() RockPaperScissorsCommand {
	return RockPaperScissorsCommand{
		GuildID:        strings.TrimSpace(c.GuildID),
		UserID:         strings.TrimSpace(c.UserID),
		Wager:          c.Wager,
		PlayerChoice:   c.PlayerChoice.Normalize(),
		ComputerChoice: c.ComputerChoice.Normalize(),
	}
}

func (c RockPaperScissorsCommand) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.UserID == "" || c.Wager <= 0 || !c.PlayerChoice.Valid() || !c.ComputerChoice.Valid() {
		return ErrInvalidRockPaperScissorsCommand
	}
	return nil
}

func ResolveRockPaperScissors(command RockPaperScissorsCommand) (RockPaperScissorsOutcome, int64, error) {
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return "", 0, err
	}
	if command.PlayerChoice == command.ComputerChoice {
		return RockPaperScissorsOutcomeTie, -(command.Wager / 2), nil
	}
	if rockPaperScissorsPlayerWins(command.PlayerChoice, command.ComputerChoice) {
		return RockPaperScissorsOutcomeWin, command.Wager, nil
	}
	return RockPaperScissorsOutcomeLoss, -command.Wager, nil
}

func (i ShopItem) Normalize() ShopItem {
	return ShopItem{
		GuildID:     strings.TrimSpace(i.GuildID),
		CommodityID: i.CommodityID,
		Name:        strings.TrimSpace(i.Name),
		NeedCoins:   i.NeedCoins,
		Description: strings.TrimSpace(i.Description),
		Code:        strings.TrimSpace(i.Code),
		AutoDelete:  i.AutoDelete,
		RoleID:      strings.TrimSpace(i.RoleID),
		Count:       i.Count,
	}
}

func (i ShopItem) Validate() error {
	i = i.Normalize()
	if i.GuildID == "" || i.CommodityID <= 0 || i.Name == "" || i.Description == "" || i.NeedCoins <= 0 || i.Count <= 0 {
		return ErrInvalidShopItem
	}
	return nil
}

func (c ShopPurchaseCommand) Normalize() ShopPurchaseCommand {
	return ShopPurchaseCommand{
		GuildID:     strings.TrimSpace(c.GuildID),
		UserID:      strings.TrimSpace(c.UserID),
		CommodityID: c.CommodityID,
		Quantity:    c.Quantity,
	}
}

func (c ShopPurchaseCommand) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.UserID == "" || c.CommodityID <= 0 || c.Quantity <= 0 {
		return ErrInvalidShopPurchase
	}
	return nil
}

func (k CoinGameKind) Normalize() CoinGameKind {
	return CoinGameKind(strings.TrimSpace(string(k)))
}

func (k CoinGameKind) Valid() bool {
	switch k.Normalize() {
	case CoinGameKindBlackjack, CoinGameKindKnowledge, CoinGameKindHigherLower:
		return true
	default:
		return false
	}
}

func (c CoinGameCommand) Normalize() CoinGameCommand {
	return CoinGameCommand{
		GuildID:      strings.TrimSpace(c.GuildID),
		ChallengerID: strings.TrimSpace(c.ChallengerID),
		OpponentID:   strings.TrimSpace(c.OpponentID),
		Wager:        c.Wager,
		Kind:         c.Kind.Normalize(),
	}
}

func (c CoinGameCommand) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.ChallengerID == "" || c.OpponentID == "" || c.ChallengerID == c.OpponentID || c.Wager < 0 || !c.Kind.Valid() {
		return ErrInvalidCoinGameCommand
	}
	return nil
}

func (c CoinGameSettlementCommand) Normalize() CoinGameSettlementCommand {
	return CoinGameSettlementCommand{
		GuildID:          strings.TrimSpace(c.GuildID),
		ChallengerID:     strings.TrimSpace(c.ChallengerID),
		OpponentID:       strings.TrimSpace(c.OpponentID),
		ChallengerReturn: c.ChallengerReturn,
		OpponentReturn:   c.OpponentReturn,
	}
}

func (c CoinGameSettlementCommand) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.ChallengerID == "" || c.OpponentID == "" || c.ChallengerID == c.OpponentID || c.ChallengerReturn < 0 || c.OpponentReturn < 0 {
		return ErrInvalidCoinGameCommand
	}
	return nil
}

func rockPaperScissorsPlayerWins(player RockPaperScissorsChoice, computer RockPaperScissorsChoice) bool {
	return (player == RockPaperScissorsChoiceScissors && computer == RockPaperScissorsChoicePaper) ||
		(player == RockPaperScissorsChoiceRock && computer == RockPaperScissorsChoiceScissors) ||
		(player == RockPaperScissorsChoicePaper && computer == RockPaperScissorsChoiceRock)
}
