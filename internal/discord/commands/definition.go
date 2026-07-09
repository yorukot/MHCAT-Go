package commands

import "sort"

type CommandType int

const (
	CommandTypeChatInput CommandType = 1
	CommandTypeUser      CommandType = 2
	CommandTypeMessage   CommandType = 3
)

type OptionType int

const (
	OptionTypeSubCommand      OptionType = 1
	OptionTypeSubCommandGroup OptionType = 2
	OptionTypeString          OptionType = 3
	OptionTypeInteger         OptionType = 4
	OptionTypeBoolean         OptionType = 5
	OptionTypeUser            OptionType = 6
	OptionTypeChannel         OptionType = 7
	OptionTypeRole            OptionType = 8
	OptionTypeMentionable     OptionType = 9
	OptionTypeNumber          OptionType = 10
	OptionTypeAttachment      OptionType = 11
)

type Scope struct {
	Kind    string `json:"kind"`
	GuildID string `json:"guild_id,omitempty"`
}

const (
	ScopeGlobal = "global"
	ScopeGuild  = "guild"
)

type Definition struct {
	Type                     CommandType       `json:"type"`
	Name                     string            `json:"name"`
	Description              string            `json:"description,omitempty"`
	NameLocalizations        map[string]string `json:"name_localizations,omitempty"`
	DescriptionLocalizations map[string]string `json:"description_localizations,omitempty"`
	Options                  []Option          `json:"options,omitempty"`
	DefaultMemberPermissions *string           `json:"default_member_permissions,omitempty"`
	Contexts                 []int             `json:"contexts,omitempty"`
	IntegrationTypes         []int             `json:"integration_types,omitempty"`
	NSFW                     bool              `json:"nsfw,omitempty"`

	Disabled  bool              `json:"disabled,omitempty"`
	Hidden    bool              `json:"hidden,omitempty"`
	Internal  bool              `json:"internal,omitempty"`
	DocsURL   string            `json:"docs_url,omitempty"`
	Ownership *CommandOwnership `json:"ownership,omitempty"`
}

type Option struct {
	Type                     OptionType        `json:"type"`
	Name                     string            `json:"name"`
	Description              string            `json:"description,omitempty"`
	NameLocalizations        map[string]string `json:"name_localizations,omitempty"`
	DescriptionLocalizations map[string]string `json:"description_localizations,omitempty"`
	Required                 bool              `json:"required,omitempty"`
	Options                  []Option          `json:"options,omitempty"`
	Choices                  []Choice          `json:"choices,omitempty"`
	ChannelTypes             []int             `json:"channel_types,omitempty"`
}

type Choice struct {
	Name              string            `json:"name"`
	NameLocalizations map[string]string `json:"name_localizations,omitempty"`
	Value             any               `json:"value"`
}

type Registry struct {
	Scope    Scope        `json:"scope"`
	Commands []Definition `json:"commands"`
}

func NewRegistry(scope Scope, definitions []Definition) Registry {
	registry := Registry{
		Scope:    scope,
		Commands: append([]Definition(nil), definitions...),
	}
	registry.Sort()
	return registry
}

func (r *Registry) Sort() {
	sort.SliceStable(r.Commands, func(i, j int) bool {
		left := r.Commands[i]
		right := r.Commands[j]
		if left.Type != right.Type {
			return left.Type < right.Type
		}
		return left.Name < right.Name
	})
}

func EmptyRegistry(scope Scope) Registry {
	return Registry{Scope: scope}
}
