package interactions

import (
	"fmt"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

type Type string

const (
	TypeSlash        Type = "slash"
	TypeComponent    Type = "component"
	TypeModal        Type = "modal"
	TypeAutocomplete Type = "autocomplete"
)

type ComponentKey struct {
	Version string
	Feature string
	Action  string
}

func (k ComponentKey) String() string {
	return fmt.Sprintf("%s:%s:%s", k.Version, k.Feature, k.Action)
}

type ModalKey struct {
	Version string
	Feature string
	Action  string
}

func (k ModalKey) String() string {
	return fmt.Sprintf("%s:%s:%s", k.Version, k.Feature, k.Action)
}

type Actor struct {
	UserID         string
	Username       string
	GuildID        string
	UserTag        string
	AvatarURL      string
	VoiceChannelID string
	RoleIDs        []string
	PermissionBits int64
}

func (a Actor) HasPermission(permission int64) bool {
	return permission != 0 && a.PermissionBits&permission == permission
}

type Route struct {
	Type         Type
	Name         string
	ComponentKey ComponentKey
	ModalKey     ModalKey
	RouteKey     RouteKey
}

func (r Route) String() string {
	if !r.RouteKey.IsZero() {
		return r.RouteKey.String()
	}
	switch r.Type {
	case TypeComponent:
		return "component:" + r.ComponentKey.String()
	case TypeModal:
		return "modal:" + r.ModalKey.String()
	default:
		return string(r.Type) + ":" + r.Name
	}
}

type Interaction struct {
	ID                        string
	Type                      Type
	CommandName               string
	SubcommandGroup           string
	Subcommand                string
	Options                   map[string]string
	CommandOptions            map[string]CommandOptionValue
	CustomID                  string
	Values                    []string
	ComponentKey              ComponentKey
	ModalKey                  ModalKey
	RouteKey                  RouteKey
	ModalFields               []customid.ModalField
	CreatedAt                 time.Time
	ChannelID                 string
	ChannelName               string
	MessageID                 string
	OriginalInteractionID     string
	OriginalInteractionUserID string
	Locale                    string
	GuildLocale               string
	Actor                     Actor
}

func (i Interaction) Route() Route {
	return Route{
		Type:         i.Type,
		Name:         i.CommandName,
		ComponentKey: i.ComponentKey,
		ModalKey:     i.ModalKey,
		RouteKey:     i.RouteKey,
	}
}
