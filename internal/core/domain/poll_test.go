package domain

import (
	"strings"
	"testing"
)

func TestPollCreateValidateUsesUTF16ChoiceLength(t *testing.T) {
	tests := []struct {
		name    string
		choice  string
		wantErr bool
	}{
		{name: "eighty BMP units", choice: strings.Repeat("界", 80)},
		{name: "eighty emoji units", choice: strings.Repeat("😀", 40)},
		{name: "eighty two emoji units", choice: strings.Repeat("😀", 41), wantErr: true},
		{name: "eighty one BMP units", choice: strings.Repeat("界", 81), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := (PollCreate{
				GuildID:   "guild-1",
				MessageID: "message-1",
				Question:  "question",
				CreatorID: "user-1",
				Choices:   []string{tc.choice, "other"},
			}).Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
