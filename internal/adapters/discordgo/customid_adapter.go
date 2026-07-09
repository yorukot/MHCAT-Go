package discordgo

import (
	"fmt"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

type ComponentInput struct {
	CustomID string
	Values   []string
}

type ModalInput struct {
	CustomID string
	Fields   []customid.ModalField
}

func ComponentInputFromDiscord(data dgo.MessageComponentInteractionData) (ComponentInput, error) {
	switch data.ComponentType {
	case dgo.ButtonComponent, dgo.SelectMenuComponent, dgo.UserSelectMenuComponent, dgo.RoleSelectMenuComponent, dgo.MentionableSelectMenuComponent, dgo.ChannelSelectMenuComponent:
		return ComponentInput{CustomID: data.CustomID, Values: append([]string(nil), data.Values...)}, nil
	default:
		return ComponentInput{}, fmt.Errorf("%w: %d", customid.ErrUnsupportedComponent, data.ComponentType)
	}
}

func ModalInputFromDiscord(data dgo.ModalSubmitInteractionData) (ModalInput, error) {
	fields := make([]customid.ModalField, 0)
	for _, component := range data.Components {
		row, ok := component.(*dgo.ActionsRow)
		if !ok {
			return ModalInput{}, fmt.Errorf("%w: modal row", customid.ErrUnsupportedComponent)
		}
		for _, child := range row.Components {
			input, ok := child.(*dgo.TextInput)
			if !ok {
				return ModalInput{}, fmt.Errorf("%w: modal field", customid.ErrUnsupportedComponent)
			}
			fields = append(fields, customid.ModalField{CustomID: input.CustomID, Value: input.Value})
		}
	}
	return ModalInput{CustomID: data.CustomID, Fields: fields}, nil
}
