package economy

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	signInCommandName           = "簽到"
	signSuccessText             = "🎉 | 今天有準時簽到很棒喔! 明天也要記得來簽到.w."
	signDailyDuplicateText      = "⚠ | 你今天已經簽到過了!請於隔天(0:00)後再來簽到!"
	signRollingDuplicateText    = "⚠ | 你今天已經簽到過了喔!"
	signCoinLimitText           = "⚠ | 不可以加超過`999999999`!!"
	signLoadingAuthor           = "正在努力為您尋找資料!"
	signLoadingIcon             = "https://media.discordapp.net/attachments/991337796960784424/1076582216127230053/6209-loading-online-circle.gif"
	signLoadingFooter           = "MHCAT 帶給你最好的discord體驗!"
	signDefaultAvatar           = "https://i.imgur.com/B91C90T.png"
	signLoadingColor            = 0xFF5809
	signFileName                = "sign.png"
	signFileContentType         = "image/png"
	legacySignYearBackEmoji     = "<:lefft:1079286176332136480>"
	legacySignMonthBackEmoji    = "<:left:1079286186570436609>"
	legacySignMonthForwardEmoji = "<:right:1079285288645447730>"
	legacySignYearForwardEmoji  = "<:right_r:1079285582263500920>"
)

var legacySignPageRe = regexp.MustCompile(`^/([0-9]{17,20})_sing\{([0-9]{4})\}-\[([0-9]{1,2})\]$`)

type signPageRequest struct {
	UserID string
	Year   int
	Month  time.Month
}

func (m Module) SignInHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.signIn.Repository == nil {
			return domain.ErrInvalidSignIn
		}
		if err := responder.Reply(ctx, signLoadingMessage(legacyCoinRankPNGURL(interaction.Actor.AvatarURL))); err != nil {
			return err
		}
		now := m.now()
		result, err := m.signIn.SignIn(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, now)
		if err != nil {
			message, ok, buildErr := m.signInErrorMessage(ctx, interaction, now, err)
			if buildErr != nil {
				return buildErr
			}
			if ok {
				if err := responder.EditOriginal(ctx, message); err != nil {
					return err
				}
				return m.trackCommand(ctx, interaction, signInCommandName)
			}
			return err
		}
		local := now.In(m.signInLocation())
		message, err := m.signCalendarMessage(result.Calendar, signPageRequest{
			UserID: interaction.Actor.UserID,
			Year:   local.Year(),
			Month:  local.Month(),
		}, actorDisplayName(interaction), fetchProfileAvatar(ctx, interaction.Actor.AvatarURL), signSuccessText)
		if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, message); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, signInCommandName)
	}
}

func (m Module) SignPageHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.signIn.Repository == nil {
			return domain.ErrInvalidSignIn
		}
		request, err := parseSignPageRequest(interaction.CustomID)
		if err != nil {
			return err
		}
		loading := signLoadingMessage(legacyCoinRankPNGURL(interaction.Actor.AvatarURL))
		loading.ClearAttachments = true
		if err := responder.UpdateMessage(ctx, loading); err != nil {
			return err
		}
		calendar, err := m.signIn.Calendar(ctx, interaction.Actor.GuildID, request.UserID, fmt.Sprintf("%04d", request.Year), fmt.Sprintf("%02d", int(request.Month)))
		if err != nil {
			return err
		}
		userInfo, userErr := m.lookupProfileUser(ctx, interaction, request.UserID)
		if userErr != nil {
			return userErr
		}
		username := strings.TrimSpace(userInfo.Username)
		if username == "" {
			username = request.UserID
		}
		message, err := m.signCalendarMessage(calendar, request, username, fetchProfileAvatar(ctx, userInfo.AvatarURL), "")
		if err != nil {
			return err
		}
		return responder.EditOriginal(ctx, message)
	}
}

func (m Module) signInErrorMessage(ctx context.Context, interaction interactions.Interaction, now time.Time, err error) (responses.Message, bool, error) {
	status := ""
	switch {
	case errors.Is(err, ports.ErrAlreadySignedIn):
		status = m.signDuplicateText(ctx, interaction.Actor.GuildID)
	case errors.Is(err, ports.ErrCoinLimitExceeded):
		status = signCoinLimitText
	default:
		return responses.Message{}, false, nil
	}
	local := now.In(m.signInLocation())
	calendar, calendarErr := m.signIn.Calendar(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, local.Format("2006"), local.Format("01"))
	if calendarErr != nil {
		return responses.Message{}, true, calendarErr
	}
	message, buildErr := m.signCalendarMessage(calendar, signPageRequest{
		UserID: interaction.Actor.UserID,
		Year:   local.Year(),
		Month:  local.Month(),
	}, actorDisplayName(interaction), fetchProfileAvatar(ctx, interaction.Actor.AvatarURL), status)
	return message, true, buildErr
}

func (m Module) signDuplicateText(ctx context.Context, guildID string) string {
	if m.query.Repository == nil {
		return signDailyDuplicateText
	}
	config, err := m.query.Repository.GetEconomyConfig(ctx, guildID)
	if err != nil {
		return signDailyDuplicateText
	}
	rolling, _ := coreeconomy.LegacySignWindow(config, true)
	if rolling {
		return signRollingDuplicateText
	}
	return signDailyDuplicateText
}

func (m Module) signCalendarMessage(calendar domain.SignCalendar, request signPageRequest, username string, avatarData []byte, status string) (responses.Message, error) {
	image, err := renderSignPNG(signCalendarView{
		Year:       request.Year,
		Month:      request.Month,
		Username:   username,
		AvatarData: avatarData,
		StatusText: status,
		Calendar:   calendar,
	})
	if err != nil {
		return responses.Message{}, err
	}
	components, err := signNavigationButtons(request.UserID, request.Year, request.Month)
	if err != nil {
		return responses.Message{}, err
	}
	return responses.Message{
		Files: []responses.File{{
			Name:        signFileName,
			ContentType: signFileContentType,
			Data:        image,
		}},
		Components:      components,
		AllowedMentions: &responses.AllowedMentions{},
	}, nil
}

func signLoadingMessage(avatarURL string) responses.Message {
	if strings.TrimSpace(avatarURL) == "" {
		avatarURL = signDefaultAvatar
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Author: &responses.EmbedAuthor{
				Name:    signLoadingAuthor,
				IconURL: signLoadingIcon,
			},
			Footer: &responses.EmbedFooter{
				Text:    signLoadingFooter,
				IconURL: avatarURL,
			},
			Color: signLoadingColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func signNavigationButtons(userID string, year int, month time.Month) ([]responses.ComponentRow, error) {
	previousYear, err := signButton(userID, year-1, month, legacySignYearBackEmoji)
	if err != nil {
		return nil, err
	}
	prevYear, prevMonth := shiftSignMonth(year, month, -1)
	previousMonth, err := signButton(userID, prevYear, prevMonth, legacySignMonthBackEmoji)
	if err != nil {
		return nil, err
	}
	nextYear, nextMonth := shiftSignMonth(year, month, 1)
	nextMonthButton, err := signButton(userID, nextYear, nextMonth, legacySignMonthForwardEmoji)
	if err != nil {
		return nil, err
	}
	followingYear, err := signButton(userID, year+1, month, legacySignYearForwardEmoji)
	if err != nil {
		return nil, err
	}
	return []responses.ComponentRow{{Components: []responses.Component{previousYear, previousMonth, nextMonthButton, followingYear}}}, nil
}

func signButton(userID string, year int, month time.Month, emoji string) (responses.Component, error) {
	payload, err := customid.KeyValuePayload(map[string]string{
		"u": userID,
		"y": fmt.Sprintf("%04d", year),
		"m": fmt.Sprintf("%02d", int(month)),
	})
	if err != nil {
		return responses.Component{}, err
	}
	id, err := customid.Encode(customid.InteractionKindComponent, "economy", "sign_page", payload)
	if err != nil {
		return responses.Component{}, err
	}
	return responses.Component{
		Type:     responses.ComponentTypeButton,
		CustomID: id,
		Emoji:    emoji,
		Style:    responses.ButtonStyleSecondary,
	}, nil
}

func legacySignCustomID(userID string, year int, month time.Month) string {
	return fmt.Sprintf("/%s_sing{%04d}-[%d]", userID, year, int(month))
}

func shiftSignMonth(year int, month time.Month, delta int) (int, time.Month) {
	target := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, delta, 0)
	return target.Year(), target.Month()
}

func parseSignPageRequest(raw string) (signPageRequest, error) {
	parsed, err := customid.ParseComponent(raw)
	if err != nil {
		return signPageRequest{}, err
	}
	if parsed.Legacy {
		return parseLegacySignPageRequest(raw)
	}
	year, err := strconv.Atoi(parsed.Payload.Values["y"])
	if err != nil || year < 1 {
		return signPageRequest{}, domain.ErrInvalidSignIn
	}
	monthInt, err := strconv.Atoi(parsed.Payload.Values["m"])
	if err != nil || monthInt < 1 || monthInt > 12 {
		return signPageRequest{}, domain.ErrInvalidSignIn
	}
	userID := strings.TrimSpace(parsed.Payload.Values["u"])
	if !customid.IsSnowflake(userID) {
		return signPageRequest{}, domain.ErrInvalidSignIn
	}
	return signPageRequest{UserID: userID, Year: year, Month: time.Month(monthInt)}, nil
}

func parseLegacySignPageRequest(raw string) (signPageRequest, error) {
	matches := legacySignPageRe.FindStringSubmatch(raw)
	if matches == nil {
		return signPageRequest{}, domain.ErrInvalidSignIn
	}
	year, err := strconv.Atoi(matches[2])
	if err != nil || year < 1 {
		return signPageRequest{}, domain.ErrInvalidSignIn
	}
	monthInt, err := strconv.Atoi(matches[3])
	if err != nil || monthInt < 1 || monthInt > 12 {
		return signPageRequest{}, domain.ErrInvalidSignIn
	}
	return signPageRequest{UserID: matches[1], Year: year, Month: time.Month(monthInt)}, nil
}

func (m Module) now() time.Time {
	if m.clock == nil {
		return time.Now()
	}
	return m.clock.Now()
}

func (m Module) signInLocation() *time.Location {
	if m.signIn.Location != nil {
		return m.signIn.Location
	}
	location, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.FixedZone("Asia/Taipei", 8*60*60)
	}
	return location
}

func actorDisplayName(interaction interactions.Interaction) string {
	tag := strings.TrimSpace(interaction.Actor.UserTag)
	if tag == "" {
		return interaction.Actor.UserID
	}
	name, _, ok := strings.Cut(tag, "#")
	if ok && name != "" {
		return name
	}
	return tag
}
