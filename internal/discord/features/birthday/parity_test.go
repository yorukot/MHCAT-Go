package birthday

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
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

func TestBirthdayStaticResponsePayloadsMatchLegacy(t *testing.T) {
	config := configSuccessMessage(domain.BirthdayConfig{
		Message:                    "{user} 生日快樂",
		UTCOffset:                  "+08:00",
		ChannelID:                  "channel-1",
		EveryoneCanSetBirthdayDate: true,
	})
	wantConfig := "**你成功設定了祝福語!**\n" +
		"<:confetti:1065654294071738399> **祝福語為:**\n{user} 生日快樂" +
		"\n<:utc:1065654078417412168> **時區為:** `UTC+08:00`" +
		"\n**<:decisionmaking:1065935264352063559> 使用者是否可以自行設定生日日期:** `true`" +
		"\n <:Channel:994524759289233438> **通知頻道: <#channel-1>**" +
		"\n <:roleplaying:985945121264635964> 身分組: null"
	if len(config.Embeds) != 1 || config.Embeds[0].Title != "<:cake:1065654305983570041> 生日系統祝福語設定" || config.Embeds[0].Description != wantConfig || config.Embeds[0].Color != birthdaySuccessColor || config.AllowedMentions == nil {
		t.Fatalf("config response = %#v", config)
	}

	deleted := deleteSuccessMessage("user-1")
	if len(deleted.Embeds) != 1 || deleted.Embeds[0].Title != "<:trashbin:995991389043163257> 刪除生日日期資料" || deleted.Embeds[0].Description != "<a:green_tick:994529015652163614> **你成功刪除了<@user-1>的資料!**" || deleted.Embeds[0].Color != birthdaySuccessColor || deleted.AllowedMentions == nil {
		t.Fatalf("delete response = %#v", deleted)
	}

	allowed := allowAdminSuccessMessage(false)
	if len(allowed.Embeds) != 1 || allowed.Embeds[0].Title != "<a:green_tick:994529015652163614> 成功變更資料" || allowed.Embeds[0].Description != "<a:green_tick:994529015652163614> **你成功將是否允許管理員設定生日資料設為**`false`!" || allowed.Embeds[0].Footer == nil || allowed.Embeds[0].Footer.Text != "本人還是可以設定喔!" || allowed.Embeds[0].Color != birthdaySuccessColor || allowed.AllowedMentions == nil {
		t.Fatalf("allow-admin response = %#v", allowed)
	}

	year, month, day := 2000, 7, 9
	module := NewModule(&fakemongo.BirthdayConfigRepository{})
	module.color = func() int { return 0x123456 }
	added := module.birthdayAddSuccessMessage(domain.BirthdayProfile{UserID: "user-1", BirthdayYear: &year, BirthdayMonth: &month, BirthdayDay: &day}, 8, 5)
	wantAdded := "<a:green_tick:994529015652163614> 恭喜你設定完成了!\n" +
		"**<a:arrow_pink:996242460294512690> 以下是<@user-1>的生日日期:**`2000/7/9`\n" +
		"**通知時間為:**`8:5`"
	if len(added.Embeds) != 1 || added.Embeds[0].Description != wantAdded || added.Embeds[0].Color != 0x123456 || len(added.Components) != 0 || added.AllowedMentions == nil {
		t.Fatalf("add response = %#v", added)
	}
}

func TestBirthdaySelectorPayloadsMatchLegacy(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{})
	module.color = func() int { return 0x123456 }
	expiresAt := time.Unix(1700000300, 600*int64(time.Millisecond))

	hour := module.hourSelectMessage("hour-id", expiresAt, "https://example.test/avatar.png")
	if len(hour.Embeds) != 1 || hour.Embeds[0].Description != "**<:24hours:1022059604747747379> 請選取你的生日通知要在幾點發送**\n**<a:warn:1000814885506129990> 你必須在<t:1700000301:R>選取完畢(超過時間將會無法選取)**" || hour.Embeds[0].Color != 0x123456 || hour.Embeds[0].Footer == nil || hour.Embeds[0].Footer.IconURL != "https://example.test/avatar.png" || len(hour.Components) != 1 || hour.Components[0].Components[0].CustomID != "hour-id" || !reflect.DeepEqual(hour.Components[0].Components[0].Options, legacyBirthdayHourOptions()) || hour.AllowedMentions == nil {
		t.Fatalf("hour response = %#v", hour)
	}

	minute := module.minuteSelectMessage("minute-id", expiresAt, "https://example.test/avatar.png")
	if len(minute.Embeds) != 1 || minute.Embeds[0].Description != "<:60minutes:1022059603153924156> **請選取你的生日通知要在幾分發送**\n**<a:warn:1000814885506129990> 你必須在<t:1700000301:R>選取完畢(超過時間將會無法選取)**" || minute.Embeds[0].Color != 0x123456 || len(minute.Components) != 1 || minute.Components[0].Components[0].CustomID != "minute-id" || !reflect.DeepEqual(minute.Components[0].Components[0].Options, legacyBirthdayMinuteOptions()) || minute.AllowedMentions == nil {
		t.Fatalf("minute response = %#v", minute)
	}
}

func TestBirthdayRandomColorUsesFullDiscordRange(t *testing.T) {
	for range 256 {
		color := legacyRandomColor()
		if color < 0 || color > 0xFFFFFF {
			t.Fatalf("color = %#x", color)
		}
	}
}

func TestBirthdayListPayloadAndAttachmentMatchLegacy(t *testing.T) {
	year, month, day := 2002, 3, 4
	message := listMessage([]domain.BirthdayProfile{{
		GuildID: "guild-1", UserID: "user-1", BirthdayYear: &year, BirthdayMonth: &month, BirthdayDay: &day,
	}}, map[string]string{"user-1": "Yoru#1234"}, 0x123456)
	wantDescription := "<:list:992002476360343602>**目前共有**`1`**人的生日數據**\n\n" +
		"┃ <@user-1>  | 生日日期(YYYY/MM/DD):2002/3/4┃"
	wantFile := "Yoru#1234(user-1)  | 生日日期(YYYY/MM/DD):2002/3/4"
	if len(message.Embeds) != 1 || message.Embeds[0].Title != "🎂 生日列表" || message.Embeds[0].Description != wantDescription || message.Embeds[0].Color != 0x123456 {
		t.Fatalf("list embed = %#v", message.Embeds)
	}
	if len(message.Files) != 1 || message.Files[0].Name != "discord.txt" || message.Files[0].ContentType != "text/plain; charset=utf-8" || string(message.Files[0].Data) != wantFile || message.AllowedMentions == nil {
		t.Fatalf("list attachment = %#v", message)
	}
}

func TestBirthdayExpiredSelectorPreservesExistingProfile(t *testing.T) {
	year, month, day, hour, minute := 1990, 1, 2, 3, 5
	repo := birthdayAddRepo(true)
	repo.Profiles = map[string]domain.BirthdayProfile{
		"guild-1/user-1": {
			GuildID: "guild-1", UserID: "user-1", BirthdayYear: &year, BirthdayMonth: &month,
			BirthdayDay: &day, SendHour: &hour, SendMinute: &minute, AllowAdmin: false,
		},
	}
	clock := &mutableBirthdayClock{now: time.Unix(1700000000, 0)}
	module := NewModuleWithClock(repo, clock)
	start := fakediscord.NewResponder()
	if err := module.Handler()(context.Background(), birthdayAddSlash(), start); err != nil {
		t.Fatalf("start handler: %v", err)
	}
	clock.now = clock.now.Add(5 * time.Minute)
	interaction := fakediscord.ComponentInteractionFromID(start.Edits[0].Components[0].Components[0].CustomID)
	interaction.Values = []string{"8"}
	responder := fakediscord.NewResponder()
	if err := module.HourSelectHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("hour handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || len(responder.Updates) != 0 {
		t.Fatalf("replies=%#v updates=%#v", responder.Replies, responder.Updates)
	}
	preserved := repo.Profiles["guild-1/user-1"]
	if preserved.BirthdayYear == nil || *preserved.BirthdayYear != 1990 || preserved.SendHour == nil || *preserved.SendHour != 3 || preserved.AllowAdmin {
		t.Fatalf("profile changed on timeout: %#v", preserved)
	}
}

type mutableBirthdayClock struct {
	now time.Time
}

func (c *mutableBirthdayClock) Now() time.Time {
	return c.now
}
