package birthday

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestBirthdaySelectOptionsMatchLegacyPayloads(t *testing.T) {
	wantHours := []responses.SelectOption{
		{Label: "1點", Description: "凌晨1點", Value: "1", Emoji: "<:moon:1022055227194605599>"},
		{Label: "2點", Description: "凌晨2點", Value: "2", Emoji: "<:moon:1022055227194605599>"},
		{Label: "3點", Description: "凌晨3點", Value: "3", Emoji: "<:moon:1022055227194605599>"},
		{Label: "4點", Description: "凌晨4點", Value: "4", Emoji: "<:moon:1022055227194605599>"},
		{Label: "5點", Description: "早上5點", Value: "5", Emoji: "<:morning:1022055616203726888>"},
		{Label: "6點", Description: "早上6點", Value: "6", Emoji: "<:morning:1022055616203726888>"},
		{Label: "7點", Description: "早上7點", Value: "7", Emoji: "<:morning:1022055616203726888>"},
		{Label: "8點", Description: "早上8點", Value: "8", Emoji: "<:morning:1022055616203726888>"},
		{Label: "9點", Description: "早上9點", Value: "9", Emoji: "<:morning:1022055616203726888>"},
		{Label: "10點", Description: "早上10點", Value: "10", Emoji: "<:morning:1022055616203726888>"},
		{Label: "11點", Description: "中午11點", Value: "11", Emoji: "<:sun:1022055614458904596>"},
		{Label: "12點", Description: "中午12點", Value: "12", Emoji: "<:sun:1022055614458904596>"},
		{Label: "13點", Description: "中午1點", Value: "13", Emoji: "<:sun:1022055614458904596>"},
		{Label: "14點", Description: "下午2點", Value: "14", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "15點", Description: "下午3點", Value: "15", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "16點", Description: "下午4點", Value: "16", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "17點", Description: "下午5點", Value: "17", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "18點", Description: "晚上6點", Value: "18", Emoji: "<:forest:1022055611044732998>"},
		{Label: "19點", Description: "晚上7點", Value: "19", Emoji: "<:forest:1022055611044732998>"},
		{Label: "20點", Description: "晚上8點", Value: "20", Emoji: "<:forest:1022055611044732998>"},
		{Label: "21點", Description: "晚上9點", Value: "21", Emoji: "<:forest:1022055611044732998>"},
		{Label: "22點", Description: "晚上10點", Value: "22", Emoji: "<:forest:1022055611044732998>"},
		{Label: "23點", Description: "晚上11點", Value: "23", Emoji: "<:forest:1022055611044732998>"},
		{Label: "24點(0點)", Description: "凌晨12點(0點)", Value: "0", Emoji: "<:moon:1022055227194605599>"},
	}
	if got := legacyBirthdayHourOptions(); !reflect.DeepEqual(got, wantHours) {
		t.Fatalf("hour options = %#v", got)
	}

	wantMinutes := []responses.SelectOption{
		{Label: "0分", Description: "每個你選取的小時的0分", Value: "0", Emoji: "<:time:1022057997515640852>"},
		{Label: "5分", Description: "每個你選取的小時的5分", Value: "5", Emoji: "<:time:1022057997515640852>"},
		{Label: "10分", Description: "每個你選取的小時的10分", Value: "10", Emoji: "<:time:1022057997515640852>"},
		{Label: "15分", Description: "每個你選取的小時的15分", Value: "15", Emoji: "<:15minutes:1022058003752570933>"},
		{Label: "20分", Description: "每個你選取的小時的20分", Value: "20", Emoji: "<:15minutes:1022058003752570933>"},
		{Label: "25分", Description: "每個你選取的小時的25分", Value: "25", Emoji: "<:15minutes:1022058003752570933>"},
		{Label: "30分", Description: "每個你選取的小時的30分", Value: "30", Emoji: "<:30minutes:1022058001722527744>"},
		{Label: "35分", Description: "每個你選取的小時的35分", Value: "35", Emoji: "<:30minutes:1022058001722527744>"},
		{Label: "40分", Description: "每個你選取的小時的40分", Value: "40", Emoji: "<:30minutes:1022058001722527744>"},
		{Label: "45分", Description: "每個你選取的小時的45分", Value: "45", Emoji: "<:45minutes:1022057999881228288>"},
		{Label: "50分", Description: "每個你選取的小時的50分", Value: "50", Emoji: "<:45minutes:1022057999881228288>"},
		{Label: "55分", Description: "每個你選取的小時的55分", Value: "55", Emoji: "<:45minutes:1022057999881228288>"},
	}
	if got := legacyBirthdayMinuteOptions(); !reflect.DeepEqual(got, wantMinutes) {
		t.Fatalf("minute options = %#v", got)
	}
}

func TestBirthdayListPreservesLegacyDisplayCutoff(t *testing.T) {
	year, month, day := 2000, 1, 2
	profiles := make([]domain.BirthdayProfile, 100)
	for i := range profiles {
		profiles[i] = domain.BirthdayProfile{
			GuildID:       "guild-1",
			UserID:        fmt.Sprintf("user-%03d", i),
			BirthdayYear:  &year,
			BirthdayMonth: &month,
			BirthdayDay:   &day,
		}
	}

	underCutoff := listMessage(profiles[:99], nil, 0x123456)
	if !strings.Contains(underCutoff.Embeds[0].Description, "<@user-098>") || strings.Contains(underCutoff.Embeds[0].Description, "由於人數過多") {
		t.Fatalf("99-row description = %q", underCutoff.Embeds[0].Description)
	}
	atCutoff := listMessage(profiles, nil, 0x123456)
	if !strings.Contains(atCutoff.Embeds[0].Description, "由於人數過多") || strings.Contains(atCutoff.Embeds[0].Description, "<@user-000>") {
		t.Fatalf("100-row description = %q", atCutoff.Embeds[0].Description)
	}
	if lines := strings.Count(string(atCutoff.Files[0].Data), "\n") + 1; lines != 100 {
		t.Fatalf("attachment lines = %d", lines)
	}
}
