package interactions

import (
	"errors"
	"fmt"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

var ErrBadInteractionID = errors.New("interaction custom id is invalid")

type CustomIDParser interface {
	ParseComponent(raw string) (customid.ID, error)
	ParseModal(raw string, fields []customid.ModalField) (customid.ID, error)
}

type DefaultCustomIDParser struct{}

func (DefaultCustomIDParser) ParseComponent(raw string) (customid.ID, error) {
	return customid.ParseComponent(raw)
}

func (DefaultCustomIDParser) ParseModal(raw string, fields []customid.ModalField) (customid.ID, error) {
	return customid.ParseModal(raw, fields)
}

func RouteKeyFromCustomID(id customid.ID) RouteKey {
	kind := TypeComponent
	if id.Kind == customid.InteractionKindModal {
		kind = TypeModal
	}
	return RouteKey{
		Kind:    kind,
		Feature: id.Feature,
		Action:  id.Action,
		Version: id.Version,
		Legacy:  id.Legacy,
	}
}

func ApplyParsedRoute(interaction Interaction, parser CustomIDParser, fields []customid.ModalField) (Interaction, error) {
	if parser == nil {
		return interaction, nil
	}
	if !interaction.RouteKey.IsZero() {
		return interaction, nil
	}
	switch interaction.Type {
	case TypeComponent:
		parsed, err := parser.ParseComponent(interaction.CustomID)
		if err != nil {
			return interaction, fmt.Errorf("%w: %w", ErrBadInteractionID, err)
		}
		interaction.RouteKey = RouteKeyFromCustomID(parsed)
	case TypeModal:
		parsed, err := parser.ParseModal(interaction.CustomID, fields)
		if err != nil {
			return interaction, fmt.Errorf("%w: %w", ErrBadInteractionID, err)
		}
		interaction.RouteKey = RouteKeyFromCustomID(parsed)
	}
	return interaction, nil
}
