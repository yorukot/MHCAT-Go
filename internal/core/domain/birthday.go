package domain

import (
	"errors"
	"strings"
)

var ErrInvalidBirthdayConfig = errors.New("invalid birthday config")
var ErrInvalidBirthdayProfile = errors.New("invalid birthday profile")
var ErrInvalidBirthdayYear = errors.New("invalid birthday year")
var ErrInvalidBirthdayMonth = errors.New("invalid birthday month")
var ErrInvalidBirthdayDay = errors.New("invalid birthday day")
var ErrInvalidBirthdayTime = errors.New("invalid birthday time")
var ErrBirthdayManageMessagesRequired = errors.New("birthday manage messages permission required")
var ErrBirthdaySelfOnly = errors.New("birthday self only")
var ErrBirthdayAdminNotAllowed = errors.New("birthday admin not allowed")

type BirthdayConfig struct {
	GuildID                    string
	Message                    string
	UTCOffset                  string
	ChannelID                  string
	EveryoneCanSetBirthdayDate bool
	RoleID                     string
}

type BirthdayProfile struct {
	GuildID       string
	UserID        string
	BirthdayYear  *int
	BirthdayMonth *int
	BirthdayDay   *int
	SendHour      *int
	SendMinute    *int
	AllowAdmin    bool
}

type BirthdayAddRequest struct {
	GuildID                string
	ActorUserID            string
	TargetUserID           string
	ActorCanManageMessages bool
	BirthdayYear           *int
	BirthdayMonth          int
	BirthdayDay            int
	CurrentYear            int
}

func (c BirthdayConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" ||
		strings.TrimSpace(c.Message) == "" ||
		strings.TrimSpace(c.UTCOffset) == "" ||
		strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidBirthdayConfig
	}
	if !validLegacyBirthdayUTCOffset(c.UTCOffset) {
		return ErrInvalidBirthdayConfig
	}
	return nil
}

func (p BirthdayProfile) ValidateDate() error {
	if err := p.ValidateIdentity(); err != nil {
		return err
	}
	if p.BirthdayMonth == nil {
		return ErrInvalidBirthdayMonth
	}
	if p.BirthdayDay == nil {
		return ErrInvalidBirthdayDay
	}
	return ValidateBirthdayDate(p.BirthdayYear, *p.BirthdayMonth, *p.BirthdayDay, 9999)
}

func (p BirthdayProfile) ValidateDateTime() error {
	if err := p.ValidateDate(); err != nil {
		return err
	}
	if p.SendHour == nil || *p.SendHour < 0 || *p.SendHour > 23 {
		return ErrInvalidBirthdayTime
	}
	if p.SendMinute == nil || *p.SendMinute < 0 || *p.SendMinute > 55 || *p.SendMinute%5 != 0 {
		return ErrInvalidBirthdayTime
	}
	return nil
}

func (p BirthdayProfile) ValidateIdentity() error {
	if strings.TrimSpace(p.GuildID) == "" || strings.TrimSpace(p.UserID) == "" {
		return ErrInvalidBirthdayProfile
	}
	return nil
}

func ValidateBirthdayAddRequest(request BirthdayAddRequest) error {
	if strings.TrimSpace(request.GuildID) == "" || strings.TrimSpace(request.ActorUserID) == "" {
		return ErrInvalidBirthdayProfile
	}
	targetUserID := strings.TrimSpace(request.TargetUserID)
	if targetUserID == "" {
		targetUserID = strings.TrimSpace(request.ActorUserID)
	}
	if targetUserID != strings.TrimSpace(request.ActorUserID) && !request.ActorCanManageMessages {
		return ErrBirthdaySelfOnly
	}
	return ValidateBirthdayDate(request.BirthdayYear, request.BirthdayMonth, request.BirthdayDay, request.CurrentYear)
}

func ValidateBirthdayDate(year *int, month int, day int, currentYear int) error {
	if year != nil && *year != 0 && (*year < 1900 || (currentYear > 0 && *year > currentYear)) {
		return ErrInvalidBirthdayYear
	}
	switch {
	case month < 1 || month > 12:
		return ErrInvalidBirthdayMonth
	case day < 1 || day > maxBirthdayDay(month):
		return ErrInvalidBirthdayDay
	default:
		return nil
	}
}

func maxBirthdayDay(month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	default:
		return 29
	}
}

func validLegacyBirthdayUTCOffset(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != len("+00:00") || value[0] != '+' || value[3:] != ":00" {
		return false
	}
	switch value[1:3] {
	case "00", "01", "02", "03", "04", "05", "06", "07", "08", "09",
		"10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
		"20", "21", "22", "23":
		return true
	default:
		return false
	}
}
