package domain

import (
	"errors"
	"testing"
)

func TestRedeemCodeValidation(t *testing.T) {
	valid := RedeemCode{Code: "abc", Price: 1.5, CreatedAtMillis: 1700000000000}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid code: %v", err)
	}
	for _, code := range []RedeemCode{
		{Price: 1, CreatedAtMillis: 1},
		{Code: "abc", Price: -1, CreatedAtMillis: 1},
		{Code: "abc", Price: 1},
	} {
		if err := code.Validate(); !errors.Is(err, ErrInvalidRedeemCode) {
			t.Fatalf("expected invalid redeem code for %#v, got %v", code, err)
		}
	}
}

func TestRedeemCommandValidation(t *testing.T) {
	valid := RedeemCommand{GuildID: "guild", Code: "abc", NowMS: 1700000000000}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid command: %v", err)
	}
	for _, command := range []RedeemCommand{
		{Code: "abc", NowMS: 1},
		{GuildID: "guild", NowMS: 1},
		{GuildID: "guild", Code: "abc"},
	} {
		if err := command.Validate(); !errors.Is(err, ErrInvalidRedeemCode) {
			t.Fatalf("expected invalid command for %#v, got %v", command, err)
		}
	}
}
