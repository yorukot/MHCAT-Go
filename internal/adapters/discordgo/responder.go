package discordgo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

type InteractionResponder struct {
	session     *dgo.Session
	interaction *dgo.Interaction
	state       *responses.State
}

func NewInteractionResponder(session *dgo.Session, interaction *dgo.Interaction) *InteractionResponder {
	return &InteractionResponder{
		session:     session,
		interaction: interaction,
		state:       responses.NewState(),
	}
}

func (r *InteractionResponder) Reply(ctx context.Context, msg responses.Message) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := r.state.MarkReply(ctx, msg); err != nil {
		return err
	}
	if err := r.session.InteractionRespond(r.interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: interactionResponseData(msg),
	}); err != nil {
		return fmt.Errorf("reply to interaction: %w", err)
	}
	return ctx.Err()
}

func (r *InteractionResponder) Defer(ctx context.Context, opts responses.DeferOptions) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := r.state.MarkDefer(ctx, opts); err != nil {
		return err
	}
	data := &dgo.InteractionResponseData{}
	if opts.Ephemeral {
		data.Flags = dgo.MessageFlagsEphemeral
	}
	if err := r.session.InteractionRespond(r.interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: data,
	}); err != nil {
		return fmt.Errorf("defer interaction: %w", err)
	}
	return ctx.Err()
}

func (r *InteractionResponder) DeferUpdate(ctx context.Context) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := r.state.MarkDeferUpdate(ctx); err != nil {
		return err
	}
	if err := r.session.InteractionRespond(r.interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		return fmt.Errorf("defer interaction message update: %w", err)
	}
	return ctx.Err()
}

func (r *InteractionResponder) ShowModal(ctx context.Context, modal responses.Modal) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := r.state.MarkModal(ctx, modal); err != nil {
		return err
	}
	if err := r.session.InteractionRespond(r.interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseModal,
		Data: modalResponseData(modal),
	}); err != nil {
		return fmt.Errorf("show interaction modal: %w", err)
	}
	return ctx.Err()
}

func (r *InteractionResponder) UpdateMessage(ctx context.Context, msg responses.Message) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := r.state.MarkUpdateMessage(ctx, msg); err != nil {
		return err
	}
	if err := r.session.InteractionRespond(r.interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseUpdateMessage,
		Data: interactionResponseData(msg),
	}); err != nil {
		return fmt.Errorf("update interaction message: %w", err)
	}
	return ctx.Err()
}

func (r *InteractionResponder) EditOriginal(ctx context.Context, msg responses.Message) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := r.state.MarkEditOriginal(ctx, msg); err != nil {
		return err
	}
	content := msg.Content
	embeds := toDiscordEmbeds(msg.Embeds)
	components := toDiscordComponents(msg.Components)
	files := toDiscordFiles(msg.Files)
	allowedMentions := toDiscordAllowedMentions(msg.AllowedMentions)
	if _, err := r.session.InteractionResponseEdit(r.interaction, &dgo.WebhookEdit{
		Content:         &content,
		Embeds:          &embeds,
		Components:      &components,
		Files:           files,
		AllowedMentions: allowedMentions,
	}); err != nil {
		return fmt.Errorf("edit original interaction response: %w", err)
	}
	return ctx.Err()
}

func (r *InteractionResponder) FollowUp(ctx context.Context, msg responses.Message) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := r.state.MarkFollowUp(ctx, msg); err != nil {
		return err
	}
	if _, err := r.session.FollowupMessageCreate(r.interaction, true, webhookParams(msg)); err != nil {
		return fmt.Errorf("create interaction follow-up: %w", err)
	}
	return ctx.Err()
}

func (r *InteractionResponder) Error(ctx context.Context, err error) error {
	msg := responses.SafeErrorMessage(err)
	if r.state.Status() == responses.StatusInitial {
		return r.Reply(ctx, msg)
	}
	return r.FollowUp(ctx, msg)
}

func (r *InteractionResponder) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.session == nil || r.interaction == nil {
		return errors.New("discordgo interaction responder is not configured")
	}
	return nil
}

func interactionResponseData(msg responses.Message) *dgo.InteractionResponseData {
	data := &dgo.InteractionResponseData{
		Content:    msg.Content,
		Embeds:     toDiscordEmbeds(msg.Embeds),
		Components: toDiscordComponents(msg.Components),
		Files:      toDiscordFiles(msg.Files),
	}
	data.AllowedMentions = toDiscordAllowedMentions(msg.AllowedMentions)
	if msg.Ephemeral {
		data.Flags = dgo.MessageFlagsEphemeral
	}
	return data
}

func modalResponseData(modal responses.Modal) *dgo.InteractionResponseData {
	return &dgo.InteractionResponseData{
		CustomID:   modal.CustomID,
		Title:      modal.Title,
		Components: toDiscordModalComponents(modal.Rows),
	}
}

func webhookParams(msg responses.Message) *dgo.WebhookParams {
	params := &dgo.WebhookParams{
		Content:    msg.Content,
		Embeds:     toDiscordEmbeds(msg.Embeds),
		Components: toDiscordComponents(msg.Components),
		Files:      toDiscordFiles(msg.Files),
	}
	params.AllowedMentions = toDiscordAllowedMentions(msg.AllowedMentions)
	if msg.Ephemeral {
		params.Flags = dgo.MessageFlagsEphemeral
	}
	return params
}

func toDiscordEmbeds(embeds []responses.Embed) []*dgo.MessageEmbed {
	converted := make([]*dgo.MessageEmbed, 0, len(embeds))
	for _, embed := range embeds {
		out := &dgo.MessageEmbed{
			Title:       embed.Title,
			Description: embed.Description,
			Color:       embed.Color,
		}
		if !embed.Timestamp.IsZero() {
			out.Timestamp = embed.Timestamp.Format(time.RFC3339)
		}
		if embed.Author != nil {
			out.Author = &dgo.MessageEmbedAuthor{
				Name:    embed.Author.Name,
				IconURL: embed.Author.IconURL,
				URL:     embed.Author.URL,
			}
		}
		if embed.Footer != nil {
			out.Footer = &dgo.MessageEmbedFooter{
				Text:    embed.Footer.Text,
				IconURL: embed.Footer.IconURL,
			}
		}
		if embed.Thumbnail != nil {
			out.Thumbnail = &dgo.MessageEmbedThumbnail{URL: embed.Thumbnail.URL}
		}
		if embed.Image != nil {
			out.Image = &dgo.MessageEmbedImage{URL: embed.Image.URL}
		}
		for _, field := range embed.Fields {
			out.Fields = append(out.Fields, &dgo.MessageEmbedField{
				Name:   field.Name,
				Value:  field.Value,
				Inline: field.Inline,
			})
		}
		converted = append(converted, out)
	}
	return converted
}

func toDiscordComponents(rows []responses.ComponentRow) []dgo.MessageComponent {
	converted := make([]dgo.MessageComponent, 0, len(rows))
	for _, row := range rows {
		actionRow := dgo.ActionsRow{Components: make([]dgo.MessageComponent, 0, len(row.Components))}
		for _, component := range row.Components {
			switch component.Type {
			case responses.ComponentTypeButton:
				actionRow.Components = append(actionRow.Components, dgo.Button{
					Label:    component.Label,
					Style:    toDiscordButtonStyle(component.Style),
					URL:      component.URL,
					CustomID: component.CustomID,
					Emoji:    toDiscordEmoji(component.Emoji),
					Disabled: component.Disabled,
				})
			case responses.ComponentTypeSelect:
				options := make([]dgo.SelectMenuOption, 0, len(component.Options))
				for _, option := range component.Options {
					options = append(options, dgo.SelectMenuOption{
						Label:       option.Label,
						Value:       option.Value,
						Description: option.Description,
						Emoji:       toDiscordEmoji(option.Emoji),
						Default:     option.Default,
					})
				}
				selectMenu := dgo.SelectMenu{
					MenuType:    dgo.StringSelectMenu,
					CustomID:    component.CustomID,
					Placeholder: component.Placeholder,
					Options:     options,
					Disabled:    component.Disabled,
				}
				if component.MinValues > 0 {
					minValues := component.MinValues
					selectMenu.MinValues = &minValues
				}
				if component.MaxValues > 0 {
					selectMenu.MaxValues = component.MaxValues
				}
				actionRow.Components = append(actionRow.Components, selectMenu)
			}
		}
		if len(actionRow.Components) > 0 {
			converted = append(converted, actionRow)
		}
	}
	return converted
}

func toDiscordModalComponents(rows []responses.ModalRow) []dgo.MessageComponent {
	converted := make([]dgo.MessageComponent, 0, len(rows))
	for _, row := range rows {
		actionRow := dgo.ActionsRow{Components: make([]dgo.MessageComponent, 0, len(row.Inputs))}
		for _, input := range row.Inputs {
			actionRow.Components = append(actionRow.Components, dgo.TextInput{
				CustomID:    input.CustomID,
				Label:       input.Label,
				Style:       toDiscordTextInputStyle(input.Style),
				Placeholder: input.Placeholder,
				Value:       input.Value,
				Required:    input.Required,
				MinLength:   input.MinLength,
				MaxLength:   input.MaxLength,
			})
		}
		if len(actionRow.Components) > 0 {
			converted = append(converted, actionRow)
		}
	}
	return converted
}

func toDiscordTextInputStyle(style responses.TextInputStyle) dgo.TextInputStyle {
	if style == responses.TextInputStyleParagraph {
		return dgo.TextInputParagraph
	}
	return dgo.TextInputShort
}

func toDiscordButtonStyle(style responses.ButtonStyle) dgo.ButtonStyle {
	switch style {
	case responses.ButtonStyleLink:
		return dgo.LinkButton
	case responses.ButtonStyleSecondary:
		return dgo.SecondaryButton
	case responses.ButtonStyleSuccess:
		return dgo.SuccessButton
	case responses.ButtonStyleDanger:
		return dgo.DangerButton
	default:
		return dgo.PrimaryButton
	}
}

func toDiscordFiles(files []responses.File) []*dgo.File {
	converted := make([]*dgo.File, 0, len(files))
	for _, file := range files {
		if file.Name == "" {
			continue
		}
		converted = append(converted, &dgo.File{
			Name:        file.Name,
			ContentType: file.ContentType,
			Reader:      bytes.NewReader(file.Data),
		})
	}
	return converted
}

func toDiscordAllowedMentions(allowed *responses.AllowedMentions) *dgo.MessageAllowedMentions {
	out := &dgo.MessageAllowedMentions{}
	if allowed == nil {
		return out
	}
	if allowed.ParseEveryone {
		out.Parse = append(out.Parse, dgo.AllowedMentionTypeEveryone)
	}
	if allowed.ParseUsers {
		out.Parse = append(out.Parse, dgo.AllowedMentionTypeUsers)
	}
	if allowed.ParseRoles {
		out.Parse = append(out.Parse, dgo.AllowedMentionTypeRoles)
	}
	out.Users = append([]string(nil), allowed.UserIDs...)
	out.Roles = append([]string(nil), allowed.RoleIDs...)
	out.RepliedUser = allowed.RepliedUser
	return out
}

var customEmojiRE = regexp.MustCompile(`^<(?P<animated>a?):(?P<name>[A-Za-z0-9_]+):(?P<id>[0-9]+)>$`)

func toDiscordEmoji(value string) *dgo.ComponentEmoji {
	if value == "" {
		return nil
	}
	matches := customEmojiRE.FindStringSubmatch(value)
	if matches == nil {
		return &dgo.ComponentEmoji{Name: value}
	}
	animated := matches[1] == "a"
	id, _ := strconv.ParseUint(matches[3], 10, 64)
	return &dgo.ComponentEmoji{Name: matches[2], ID: strconv.FormatUint(id, 10), Animated: animated}
}
