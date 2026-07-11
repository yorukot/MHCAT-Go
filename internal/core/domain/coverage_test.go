package domain

import (
	"errors"
	"testing"
)

func TestValidateBirthdayAddRequestEnforcesActorPermission(t *testing.T) {
	request := BirthdayAddRequest{GuildID: " guild-1 ", ActorUserID: "actor", TargetUserID: "target", BirthdayMonth: 7, BirthdayDay: 11, CurrentYear: 2026}
	if err := ValidateBirthdayAddRequest(request); !errors.Is(err, ErrBirthdaySelfOnly) {
		t.Fatalf("unprivileged target add: %v", err)
	}
	request.ActorCanManageMessages = true
	if err := ValidateBirthdayAddRequest(request); err != nil {
		t.Fatalf("privileged target add: %v", err)
	}
	request.GuildID = ""
	if err := ValidateBirthdayAddRequest(request); !errors.Is(err, ErrInvalidBirthdayProfile) {
		t.Fatalf("missing guild: %v", err)
	}
}

func TestGachaDrawResultNonAirPrizes(t *testing.T) {
	result := GachaDrawResult{Prizes: []GachaDrawPrizeResult{{Name: "air", Air: true}, {Name: "coin"}, {Name: "code", Code: "ABC"}}}
	prizes := result.NonAirPrizes()
	if len(prizes) != 2 || prizes[0].Name != "coin" || prizes[1].Name != "code" {
		t.Fatalf("non-air prizes = %#v", prizes)
	}
}

func TestWarningHistoryValidateQuery(t *testing.T) {
	if err := (WarningHistory{GuildID: " guild ", UserID: " user "}).ValidateQuery(); err != nil {
		t.Fatalf("valid query: %v", err)
	}
	if err := (WarningHistory{GuildID: "guild"}).ValidateQuery(); !errors.Is(err, ErrInvalidWarningQuery) {
		t.Fatalf("missing user query: %v", err)
	}
}
