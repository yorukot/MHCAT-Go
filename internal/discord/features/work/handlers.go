package work

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	workservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/work"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	subcommandWorkSettings   = "打工系統設定"
	subcommandAddWork        = "新增打工事項"
	subcommandDeleteWork     = "打工事項刪除"
	subcommandWorkInterface  = "打工介面"
	subcommandAddUserEnergy  = "增加個人精力"
	subcommandAddAllEnergy   = "增加全體精力"
	permissionManageMessages = 8192
	workErrorColor           = 0xEA0000
	workSuccessColor         = 0x53FF53
	dashboardColor           = 0xDF1F2F
	dashboardTitle           = "<a:announcement:1005035747197337650> | 該指令已經移往控制面板，請前往控制面板進行設定"
	dashboardLabel           = "點我前往儀錶板設定!"
	dashboardEmoji           = "<a:arrow:986268851786375218>"
	workDocsURL              = "allcommands/%E6%89%93%E5%B7%A5%E7%B3%BB%E7%B5%B1/user_work"
	workSettingsDocsURL      = "allcommands/%E6%89%93%E5%B7%A5%E7%B3%BB%E7%B5%B1/work_set"
	workDeleteDocsURL        = "allcommands/%E6%89%93%E5%B7%A5%E7%B3%BB%E7%B5%B1/delete_work"
	workEnergyDocsURL        = "allcommands/%E6%89%93%E5%B7%A5%E7%B3%BB%E7%B5%B1/add_energy"
)

var ErrWorkSubcommandNotImplemented = errors.New("work subcommand not implemented")

type captchaGenerator func() (int, int)

func (m Module) Handler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		switch interaction.Subcommand {
		case subcommandWorkSettings:
			return m.handleWorkSettings(ctx, interaction, responder)
		case subcommandAddWork:
			if err := responder.Reply(ctx, dashboardRedirectMessage(interaction.Actor.GuildID)); err != nil {
				return err
			}
			return m.track(ctx, interaction)
		case subcommandDeleteWork:
			return m.handleDeleteWork(ctx, interaction, responder)
		case subcommandWorkInterface:
			return m.handleWorkInterface(ctx, interaction, responder)
		case subcommandAddUserEnergy:
			return m.handleAddUserEnergy(ctx, interaction, responder)
		case subcommandAddAllEnergy:
			return m.handleAddAllEnergy(ctx, interaction, responder)
		default:
			return responder.Reply(ctx, workNotImplementedMessage(interaction.Subcommand))
		}
	}
}

func (m Module) DetailHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.work == nil {
			return responder.UpdateMessage(ctx, workNotImplementedMessage("work detail"))
		}
		if !customIDAllowsUser(interaction.CustomID, customid.InteractionKindComponent, interaction.Actor.UserID) {
			return responder.Reply(ctx, workUnauthorizedMessage())
		}
		itemKey, err := payloadValue(interaction.CustomID, customid.InteractionKindComponent, "job")
		if err != nil {
			return responder.UpdateMessage(ctx, legacyWorkErrorMessage("很抱歉，找不到這個打工地點，請於幾秒鐘後重試!"))
		}
		view, item, err := m.work.Detail(ctx, m.workRequest(ctx, interaction), itemKey)
		if err != nil {
			return responder.UpdateMessage(ctx, workErrorMessageFromError(err))
		}
		return responder.UpdateMessage(ctx, workDetailMessage(view, item, m.work.CanStart(), m.randomColor()))
	}
}

func (m Module) StartHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		return m.handleWorkStart(ctx, interaction, responder, false)
	}
}

func (m Module) OverrideHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		return m.handleWorkStart(ctx, interaction, responder, true)
	}
}

func (m Module) CancelHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !customIDAllowsUser(interaction.CustomID, customid.InteractionKindComponent, interaction.Actor.UserID) {
			return responder.Reply(ctx, workUnauthorizedMessage())
		}
		return responder.UpdateMessage(ctx, workCancelMessage())
	}
}

func (m Module) handleWorkStart(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, override bool) error {
	if m.work == nil {
		return responder.UpdateMessage(ctx, workNotImplementedMessage("work start"))
	}
	if !customIDAllowsUser(interaction.CustomID, customid.InteractionKindComponent, interaction.Actor.UserID) {
		return responder.Reply(ctx, workUnauthorizedMessage())
	}
	itemKey, err := payloadValue(interaction.CustomID, customid.InteractionKindComponent, "job")
	if err != nil {
		return responder.UpdateMessage(ctx, legacyWorkErrorMessage("很抱歉，找不到這個打工地點，請於幾秒鐘後重試!"))
	}
	_, item, updated, err := m.work.Start(ctx, m.workRequest(ctx, interaction), itemKey, override)
	if err != nil {
		if errors.Is(err, domain.ErrWorkAlreadyBusy) {
			return responder.UpdateMessage(ctx, workOverrideMessage(item, interaction.Actor.UserID))
		}
		return responder.UpdateMessage(ctx, workErrorMessageFromError(err))
	}
	if err := responder.UpdateMessage(ctx, workStartSuccessMessage(item, updated)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) CaptchaHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.work == nil {
			return responder.Reply(ctx, workNotImplementedMessage("work captcha"))
		}
		expected, err := payloadValue(interaction.CustomID, customid.InteractionKindModal, "sum")
		if err != nil {
			return responder.Reply(ctx, legacyWorkErrorMessage("驗證碼錯誤!"))
		}
		expectedNumber, expectedOK := legacyCaptchaNumber(expected)
		answerNumber, answerOK := legacyCaptchaNumber(modalFieldValue(interaction.ModalFields, "captcha"))
		if !expectedOK || !answerOK || expectedNumber != answerNumber {
			return responder.Reply(ctx, responses.Message{
				Content:         "<a:error:980086028113182730> | 驗證碼錯誤!",
				AllowedMentions: &responses.AllowedMentions{},
			})
		}
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		view, err := m.work.Interface(ctx, m.workRequest(ctx, interaction))
		if err != nil {
			return responder.EditOriginal(ctx, workErrorMessageFromError(err))
		}
		if err := responder.EditOriginal(ctx, workInterfaceMessage(view, m.randomColor())); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
}

func legacyCaptchaNumber(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, true
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil && !math.IsNaN(parsed) {
		return parsed, true
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "0x") || strings.HasPrefix(lower, "+0x") || strings.HasPrefix(lower, "-0x") ||
		strings.HasPrefix(lower, "0b") || strings.HasPrefix(lower, "+0b") || strings.HasPrefix(lower, "-0b") ||
		strings.HasPrefix(lower, "0o") || strings.HasPrefix(lower, "+0o") || strings.HasPrefix(lower, "-0o") {
		parsed, err := strconv.ParseInt(value, 0, 64)
		return float64(parsed), err == nil
	}
	return 0, false
}

func (m Module) handleWorkInterface(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if m.work == nil {
		return responder.Reply(ctx, workNotImplementedMessage(interaction.Subcommand))
	}
	settings, err := m.work.Settings(ctx, interaction.Actor.GuildID)
	if err != nil {
		return responder.Reply(ctx, workErrorMessageFromError(err))
	}
	if settings.Captcha {
		left, right := m.captcha()
		if err := responder.ShowModal(ctx, workCaptchaModal(left, right)); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	view, err := m.work.Interface(ctx, m.workRequest(ctx, interaction))
	if err != nil {
		return responder.EditOriginal(ctx, workErrorMessageFromError(err))
	}
	if err := responder.EditOriginal(ctx, workInterfaceMessage(view, m.randomColor())); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleWorkSettings(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if m.work == nil || !m.work.CanAdmin() {
		return responder.Reply(ctx, workNotImplementedMessage(interaction.Subcommand))
	}
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	if !interaction.Actor.HasPermission(permissionManageMessages) {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("你需要有`訊息管理`才能使用此指令", workSettingsDocsURL))
	}
	dailyEnergy, ok := intOption(interaction, "每天可獲得多少精力")
	if !ok {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("很抱歉，出現了未知的錯誤，請重試!", workSettingsDocsURL))
	}
	maxEnergy, ok := intOption(interaction, "精力上限為多少")
	if !ok {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("很抱歉，出現了未知的錯誤，請重試!", workSettingsDocsURL))
	}
	captcha := boolOption(interaction, "是否需要驗證", false)
	config, err := m.work.SaveConfig(ctx, domain.WorkConfigCommand{
		GuildID:     interaction.Actor.GuildID,
		DailyEnergy: dailyEnergy,
		MaxEnergy:   maxEnergy,
		Captcha:     captcha,
	})
	if err != nil {
		return responder.EditOriginal(ctx, workErrorMessageFromErrorWithURL(err, workSettingsDocsURL))
	}
	if err := responder.EditOriginal(ctx, workSettingsSuccessMessage(config)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleDeleteWork(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if m.work == nil || !m.work.CanAdmin() {
		return responder.Reply(ctx, workNotImplementedMessage(interaction.Subcommand))
	}
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	if !interaction.Actor.HasPermission(permissionManageMessages) {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("你需要有`訊息管理`才能使用此指令", workDeleteDocsURL))
	}
	name := stringOption(interaction, "打工地點名稱")
	if err := m.work.DeleteItem(ctx, domain.WorkDeleteItemCommand{GuildID: interaction.Actor.GuildID, Name: name}); err != nil {
		return responder.EditOriginal(ctx, workErrorMessageFromErrorWithURL(err, workDeleteDocsURL))
	}
	if err := responder.EditOriginal(ctx, workDeleteSuccessMessage(name)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleAddUserEnergy(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if m.work == nil || !m.work.CanAdmin() {
		return responder.Reply(ctx, workNotImplementedMessage(interaction.Subcommand))
	}
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	if !interaction.Actor.HasPermission(permissionManageMessages) {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("你需要有`訊息管理`才能使用此指令", workEnergyDocsURL))
	}
	amount, ok := intOption(interaction, "要給多少精力")
	if !ok {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("很抱歉，出現了未知的錯誤，請重試!", workEnergyDocsURL))
	}
	userID := stringOption(interaction, "使用者")
	_, err := m.work.GrantEnergy(ctx, domain.WorkEnergyGrantCommand{GuildID: interaction.Actor.GuildID, UserID: userID, Amount: amount})
	if err != nil {
		return responder.EditOriginal(ctx, workErrorMessageFromErrorWithURL(err, workEnergyDocsURL))
	}
	if err := responder.EditOriginal(ctx, workGrantUserEnergySuccessMessage(userID, amount)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleAddAllEnergy(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if m.work == nil || !m.work.CanAdmin() {
		return responder.Reply(ctx, workNotImplementedMessage(interaction.Subcommand))
	}
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	if !interaction.Actor.HasPermission(permissionManageMessages) {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("你需要有`訊息管理`才能使用此指令", workEnergyDocsURL))
	}
	amount, ok := intOption(interaction, "要給多少精力")
	if !ok {
		return responder.EditOriginal(ctx, legacyWorkErrorMessageWithURL("很抱歉，出現了未知的錯誤，請重試!", workEnergyDocsURL))
	}
	if _, err := m.work.GrantEnergyToAll(ctx, domain.WorkEnergyGrantAllCommand{GuildID: interaction.Actor.GuildID, Amount: amount}); err != nil {
		return responder.EditOriginal(ctx, workErrorMessageFromErrorWithURL(err, workEnergyDocsURL))
	}
	if err := responder.EditOriginal(ctx, workGrantAllEnergySuccessMessage(amount)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func dashboardRedirectMessage(guildID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: dashboardTitle,
			Color: dashboardColor,
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:  responses.ComponentTypeButton,
				Style: responses.ButtonStyleLink,
				URL:   fmt.Sprintf("https://mhcat.yorukot.me//guilds/%s/work", strings.TrimSpace(guildID)),
				Label: dashboardLabel,
				Emoji: dashboardEmoji,
			}},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) workRequest(ctx context.Context, interaction interactions.Interaction) workservice.InterfaceRequest {
	return workservice.InterfaceRequest{
		GuildID:       interaction.Actor.GuildID,
		GuildName:     m.guildName(ctx, interaction),
		UserID:        interaction.Actor.UserID,
		UserTag:       interaction.Actor.UserTag,
		UserAvatarURL: interaction.Actor.AvatarURL,
		RoleIDs:       interaction.Actor.RoleIDs,
	}
}

func (m Module) guildName(ctx context.Context, interaction interactions.Interaction) string {
	if m.discord != nil && strings.TrimSpace(interaction.Actor.GuildID) != "" {
		info, err := m.discord.GuildInfo(ctx, interaction.Actor.GuildID)
		if err == nil && strings.TrimSpace(info.Name) != "" {
			return info.Name
		}
	}
	if strings.TrimSpace(interaction.ChannelName) != "" {
		return interaction.ChannelName
	}
	return interaction.Actor.GuildID
}

func workCaptchaModal(left int, right int) responses.Modal {
	payload, err := customid.KeyValuePayload(map[string]string{"sum": strconv.Itoa(left + right)})
	if err != nil {
		payload = customid.EmptyPayload()
	}
	id, err := customid.Encode(customid.InteractionKindModal, "work", "captcha", payload)
	if err != nil {
		id = "mhcat:v1:work:captcha:"
	}
	return responses.Modal{
		CustomID: id,
		Title:    "認證你不是機器人!",
		Rows: []responses.ModalRow{{
			Inputs: []responses.TextInput{{
				CustomID: "captcha",
				Label:    fmt.Sprintf("請計算%d + %d", left, right),
				Style:    responses.TextInputStyleShort,
				Required: true,
			}},
		}},
	}
}

func workInterfaceMessage(view domain.WorkInterfaceView, colors ...int) responses.Message {
	guildName := strings.TrimSpace(view.GuildName)
	if guildName == "" {
		guildName = view.Config.GuildID
	}
	userTag := strings.TrimSpace(view.UserTag)
	if userTag == "" {
		userTag = view.User.UserID
	}
	color := randomWorkColor()
	if len(colors) > 0 {
		color = colors[0]
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       fmt.Sprintf("<:list:992002476360343602> 以下是%s的打工簡章", guildName),
			Description: workInterfaceDescription(view),
			Color:       color,
			Fields:      workItemFields(view.VisibleItems),
			Footer: &responses.EmbedFooter{
				Text:    userTag + "的查詢",
				IconURL: view.UserAvatarURL,
			},
		}},
		Components:      workItemButtons(view.VisibleItems, view.User.UserID),
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workInterfaceDescription(view domain.WorkInterfaceView) string {
	return fmt.Sprintf("<:status:1048643690572283965> **你目前的打工狀態 :** `%s`\n<:chronometer:986065703369080884> **剩餘時間:** %s\n<:lighting:1048626093994803200> **剩餘體力:** `%s \\ %s`\n<a:arrow_pink:996242460294512690> **點擊下方的按扭進行打工!**",
		view.User.EffectiveState(view.NowUnix),
		view.User.RemainingTimeText(view.NowUnix),
		workScalarText(view.User.EnergyText, view.User.Energy),
		workScalarText(view.Config.MaxEnergyText, view.Config.MaxEnergy),
	)
}

func workItemFields(items []domain.WorkItem) []responses.EmbedField {
	limit := len(items)
	if limit > 25 {
		limit = 25
	}
	fields := make([]responses.EmbedField, 0, limit)
	for _, item := range items[:limit] {
		role := "無"
		if strings.TrimSpace(item.RoleID) != "" {
			role = "<@&" + item.RoleID + ">"
		}
		fields = append(fields, responses.EmbedField{
			Name: fmt.Sprintf("<:id:985950321975128094> **打工地點名稱 :** `%s`", item.Name),
			Value: fmt.Sprintf("<:lighting:1048626093994803200> **打工所需精力 : **`%s` \n<:ontime:981966857718353950> **耗費時間 : **`%s分(%s小時)` \n<:id:985950321975128094> **打工報酬 : **`%s`(代幣)\n<:roleplaying:985945121264635964> **所需身分組 : ** %s",
				workScalarText(item.EnergyCostText, item.EnergyCost),
				legacyDurationNumber(workScalarNumber(item.DurationText, item.DurationSec)/60),
				legacyDurationNumber(workScalarNumber(item.DurationText, item.DurationSec)/60/60),
				workScalarText(item.CoinRewardText, item.CoinReward),
				role,
			),
			Inline: true,
		})
	}
	return fields
}

func workItemButtons(items []domain.WorkItem, userID string) []responses.ComponentRow {
	limit := len(items)
	if limit > 25 {
		limit = 25
	}
	rows := make([]responses.ComponentRow, 0, 5)
	for start := 0; start < limit; start += 5 {
		end := start + 5
		if end > limit {
			end = limit
		}
		row := responses.ComponentRow{}
		for _, item := range items[start:end] {
			row.Components = append(row.Components, responses.Component{
				Type:     responses.ComponentTypeButton,
				CustomID: workDetailCustomID(item, userID),
				Label:    buttonLabel(item.Name),
				Style:    responses.ButtonStylePrimary,
			})
		}
		rows = append(rows, row)
	}
	return rows
}

func workDetailMessage(view domain.WorkInterfaceView, item domain.WorkItem, startEnabled bool, colors ...int) responses.Message {
	color := randomWorkColor()
	if len(colors) > 0 {
		color = colors[0]
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: fmt.Sprintf("<:creativeteaching:986060052949524600> 以下是%s打工的詳細資料", item.Name),
			Description: fmt.Sprintf("<:id:1010884394791207003> **打工地點名稱:**`%s`\n<:money:997374193026994236> **可獲得代幣:**`%s 個代幣`\n<:lighting:1048626093994803200> **耗費精力:**`%s`\n",
				item.Name,
				workScalarText(item.CoinRewardText, item.CoinReward),
				workScalarText(item.EnergyCostText, item.EnergyCost),
			),
			Color: color,
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:     responses.ComponentTypeButton,
				CustomID: workStartCustomID(item, view.User.UserID),
				Label:    "確認打工",
				Emoji:    "<:tickmark:985949769224556614>",
				Style:    responses.ButtonStyleSuccess,
				Disabled: !startEnabled,
			}},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workOverrideMessage(item domain.WorkItem, userID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "⚠️ | 你目前有其他工作，如果堅持繼續將會強制停止之前的工作，並且不返還體力，是否繼續?",
			Color: workErrorColor,
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{
				{
					Type:     responses.ComponentTypeButton,
					CustomID: workOverrideCustomID(item, userID),
					Label:    "是",
					Emoji:    "✅",
					Style:    responses.ButtonStylePrimary,
				},
				{
					Type:     responses.ComponentTypeButton,
					CustomID: workCancelCustomID(item, userID),
					Label:    "否",
					Emoji:    "❎",
					Style:    responses.ButtonStyleDanger,
				},
			},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workCancelMessage() responses.Message {
	return responses.Message{
		Content:         ":x: | **成功取消!**",
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workStartSuccessMessage(item domain.WorkItem, updated domain.WorkUserState) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:working:1048617967799242772> 成功取得該工作!",
			Description: fmt.Sprintf("<a:green_tick:994529015652163614>**你已經成功取得**`%s`**的工作**\n<:tickmark:985949769224556614> **預計於:<t:%d:R>打工完成**",
				item.Name,
				updated.EndTimeUnix,
			),
			Color: workSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workSettingsSuccessMessage(config domain.WorkConfig) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:working:1048617967799242772> 打工系統",
			Description: fmt.Sprintf("<a:green_tick:994529015652163614>**成功設定打工系統!**\n<:lighting:1048626093994803200> **每天可獲得精力:**`%d`\n**精力上限:**`%d`\n**是否需要驗證:**`%t`",
				config.DailyEnergy,
				config.MaxEnergy,
				config.Captcha,
			),
			Color: workSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workDeleteSuccessMessage(name string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:working:1048617967799242772> 打工事項",
			Description: fmt.Sprintf("<:trashbin:995991389043163257> **成功刪除打工事項**:`%s`", strings.TrimSpace(name)),
			Color:       workSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workGrantAllEnergySuccessMessage(amount int64) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:working:1048617967799242772> 打工系統",
			Description: fmt.Sprintf("<a:green_tick:994529015652163614>**成功增加精力!!**\n**成功為所有已建檔的使用者增加**`%d`**精力**", amount),
			Color:       workSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workGrantUserEnergySuccessMessage(userID string, amount int64) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:working:1048617967799242772> 打工系統",
			Description: fmt.Sprintf("<a:green_tick:994529015652163614>**成功增加精力!!**\n**成功為**<@%s>**增加**`%d`**精力**", strings.TrimSpace(userID), amount),
			Color:       workSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workErrorMessageFromError(err error) responses.Message {
	return workErrorMessageFromErrorWithURL(err, workDocsURL)
}

func workErrorMessageFromErrorWithURL(err error, docsURL string) responses.Message {
	switch {
	case errors.Is(err, ports.ErrWorkConfigMissing):
		if docsURL == workDocsURL {
			return legacyWorkErrorMessageWithURL("請先請管理員使用`/打工系統 打工系統設定`進行初始設定!", docsURL)
		}
		return legacyWorkErrorMessageWithURL("請先使用`/打工系統 打工系統設定`進行初始設定!", docsURL)
	case errors.Is(err, ports.ErrWorkItemsMissing):
		return legacyWorkErrorMessageWithURL("目前沒有任何打工給你做喔!", docsURL)
	case errors.Is(err, ports.ErrNoVisibleWorkItem):
		return legacyWorkErrorMessageWithURL("還沒有適合你身分組的打工喔!可以請管理員增加打工!", docsURL)
	case errors.Is(err, ports.ErrWorkItemMissing):
		return legacyWorkErrorMessageWithURL("很抱歉，找不到這個名字的資料!", docsURL)
	case errors.Is(err, domain.ErrWorkEnergyInsufficient):
		return legacyWorkErrorMessageWithURL("你的精力不夠!", docsURL)
	case errors.Is(err, domain.ErrWorkAlreadyBusy):
		return legacyWorkErrorMessageWithURL("你目前有其他工作，如果堅持繼續將會強制停止之前的工作，並且不返還體力，是否繼續?", docsURL)
	case errors.Is(err, domain.ErrWorkStartUnavailable):
		return workNotImplementedMessage("work start")
	case errors.Is(err, domain.ErrWorkAdminUnavailable):
		return workNotImplementedMessage("work admin")
	default:
		return legacyWorkErrorMessageWithURL("很抱歉，找不到這個打工地點，請於幾秒鐘後重試!", docsURL)
	}
}

func legacyWorkErrorMessage(content string) responses.Message {
	return legacyWorkErrorMessageWithURL(content, workDocsURL)
}

func legacyWorkErrorMessageWithURL(content string, docsURL string) responses.Message {
	if strings.TrimSpace(docsURL) == "" {
		docsURL = workDocsURL
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Description: "<:MHCATdarkdsadsadsadsadsadas1:1079853990541529208> [立即前往官方文檔查看問題](https://docsmhcat.yorukot.me/" + docsURL + ")",
			Color:       workErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workUnauthorizedMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: "<a:error:980086028113182730> | 你不是查詢者無法使用!",
			Color: workErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workNotImplementedMessage(subcommand string) responses.Message {
	if strings.TrimSpace(subcommand) == "" {
		subcommand = "unknown"
	}
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: fmt.Sprintf("%s: %s", ErrWorkSubcommandNotImplemented.Error(), subcommand),
			Color: workErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func workDetailCustomID(item domain.WorkItem, userID ...string) string {
	encoded, err := workActionID(item, "detail", userID...)
	if err != nil {
		return "mhcat:v1:work:detail:"
	}
	return encoded
}

func workStartCustomID(item domain.WorkItem, userID ...string) string {
	return workActionCustomID(item, "start", userID...)
}

func workOverrideCustomID(item domain.WorkItem, userID ...string) string {
	return workActionCustomID(item, "override", userID...)
}

func workCancelCustomID(item domain.WorkItem, userID ...string) string {
	return workActionCustomID(item, "cancel", userID...)
}

func workActionCustomID(item domain.WorkItem, action string, userID ...string) string {
	encoded, err := workActionID(item, action, userID...)
	if err != nil {
		return "mhcat:v1:work:" + action + ":"
	}
	return encoded
}

func workActionID(item domain.WorkItem, action string, userID ...string) (string, error) {
	values := map[string]string{"job": item.Key()}
	if len(userID) > 0 && strings.TrimSpace(userID[0]) != "" {
		values["user"] = strings.TrimSpace(userID[0])
	}
	payload, err := customid.KeyValuePayload(values)
	if err != nil {
		payload = customid.EmptyPayload()
	}
	return customid.Encode(customid.InteractionKindComponent, "work", action, payload)
}

func customIDAllowsUser(raw string, kind customid.InteractionKind, actorUserID string) bool {
	values, err := payloadValues(raw, kind)
	if err != nil {
		return true
	}
	expected := strings.TrimSpace(values["user"])
	if expected == "" {
		expected = strings.TrimSpace(values["u"])
	}
	return expected == "" || strings.TrimSpace(actorUserID) == "" || expected == strings.TrimSpace(actorUserID)
}

func payloadValue(raw string, kind customid.InteractionKind, key string) (string, error) {
	values, err := payloadValues(raw, kind)
	if err != nil {
		return "", err
	}
	value := values[key]
	if value == "" && key == "sum" {
		value = values["s"]
	}
	if value == "" {
		return "", customid.ErrInvalidPayload
	}
	return value, nil
}

func payloadValues(raw string, kind customid.InteractionKind) (map[string]string, error) {
	var (
		id  customid.ID
		err error
	)
	if kind == customid.InteractionKindModal {
		id, err = customid.ParseModal(raw, nil)
	} else {
		id, err = customid.ParseComponent(raw)
	}
	if err != nil {
		return nil, err
	}
	if id.Payload.Values == nil {
		return map[string]string{}, nil
	}
	return id.Payload.Values, nil
}

func modalFieldValue(fields []customid.ModalField, key string) string {
	for _, field := range fields {
		if field.CustomID == key {
			return strings.TrimSpace(field.Value)
		}
	}
	return ""
}

func stringOption(interaction interactions.Interaction, name string) string {
	if value, ok := interaction.CommandOptions[name]; ok {
		return strings.TrimSpace(value.String)
	}
	return strings.TrimSpace(interaction.Options[name])
}

func intOption(interaction interactions.Interaction, name string) (int64, bool) {
	if value, ok := interaction.CommandOptions[name]; ok {
		if value.Type == interactions.CommandOptionInteger {
			return value.Int, true
		}
		if strings.TrimSpace(value.String) != "" {
			parsed, err := strconv.ParseInt(strings.TrimSpace(value.String), 10, 64)
			return parsed, err == nil
		}
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(interaction.Options[name]), 10, 64)
	return parsed, err == nil
}

func boolOption(interaction interactions.Interaction, name string, fallback bool) bool {
	if value, ok := interaction.CommandOptions[name]; ok {
		if value.Type == interactions.CommandOptionBoolean {
			return value.Bool
		}
		if strings.TrimSpace(value.String) != "" {
			parsed, err := strconv.ParseBool(strings.TrimSpace(value.String))
			if err == nil {
				return parsed
			}
		}
	}
	raw := strings.TrimSpace(interaction.Options[name])
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func legacyDurationNumber(value float64) string {
	if math.IsNaN(value) {
		return "NaN"
	}
	if math.IsInf(value, 1) {
		return "Infinity"
	}
	if math.IsInf(value, -1) {
		return "-Infinity"
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func workScalarText(text string, fallback int64) string {
	if text = strings.TrimSpace(text); text != "" {
		return text
	}
	return strconv.FormatInt(fallback, 10)
}

func workScalarNumber(text string, fallback int64) float64 {
	text = strings.TrimSpace(text)
	if text == "" {
		return float64(fallback)
	}
	if text == "null" {
		return 0
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return math.NaN()
	}
	return value
}

func buttonLabel(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= 80 {
		return value
	}
	return string(runes[:80])
}

func randomCaptcha() (int, int) {
	return secureDigit(), secureDigit()
}

func randomWorkColor() int {
	value, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		return 0
	}
	return int(value.Int64())
}

func secureDigit() int {
	value, err := rand.Int(rand.Reader, big.NewInt(10))
	if err != nil {
		return 0
	}
	return int(value.Int64())
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: CommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "work",
	})
}
