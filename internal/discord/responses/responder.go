package responses

import (
	"context"
	"time"
)

type Message struct {
	Content         string
	Embeds          []Embed
	Components      []ComponentRow
	Files           []File
	AllowedMentions *AllowedMentions
	Ephemeral       bool
}

type Modal struct {
	CustomID string
	Title    string
	Rows     []ModalRow
}

type ModalRow struct {
	Inputs []TextInput
}

type TextInputStyle string

const (
	TextInputStyleShort     TextInputStyle = "short"
	TextInputStyleParagraph TextInputStyle = "paragraph"
)

type TextInput struct {
	CustomID    string
	Label       string
	Style       TextInputStyle
	Placeholder string
	Value       string
	Required    bool
	MinLength   int
	MaxLength   int
}

type File struct {
	Name        string
	ContentType string
	Data        []byte
}

type AllowedMentions struct {
	ParseEveryone bool
	ParseUsers    bool
	ParseRoles    bool
	UserIDs       []string
	RoleIDs       []string
	RepliedUser   bool
}

type Embed struct {
	Title       string
	Description string
	Color       int
	Timestamp   time.Time
	Author      *EmbedAuthor
	Footer      *EmbedFooter
	Thumbnail   *EmbedImage
	Image       *EmbedImage
	Fields      []EmbedField
}

type EmbedAuthor struct {
	Name    string
	IconURL string
	URL     string
}

type EmbedFooter struct {
	Text    string
	IconURL string
}

type EmbedImage struct {
	URL string
}

type EmbedField struct {
	Name   string
	Value  string
	Inline bool
}

type ComponentRow struct {
	Components []Component
}

type ComponentType string

const (
	ComponentTypeButton ComponentType = "button"
	ComponentTypeSelect ComponentType = "select"
)

type ButtonStyle string

const (
	ButtonStyleLink      ButtonStyle = "link"
	ButtonStylePrimary   ButtonStyle = "primary"
	ButtonStyleSecondary ButtonStyle = "secondary"
	ButtonStyleSuccess   ButtonStyle = "success"
	ButtonStyleDanger    ButtonStyle = "danger"
)

type Component struct {
	Type        ComponentType
	CustomID    string
	Label       string
	URL         string
	Emoji       string
	Style       ButtonStyle
	Placeholder string
	Disabled    bool
	MinValues   int
	MaxValues   int
	Options     []SelectOption
}

type SelectOption struct {
	Label       string
	Value       string
	Description string
	Emoji       string
	Default     bool
}

type DeferOptions struct {
	Ephemeral bool
	Deadline  time.Time
}

type Responder interface {
	Reply(ctx context.Context, msg Message) error
	Defer(ctx context.Context, opts DeferOptions) error
	DeferUpdate(ctx context.Context) error
	ShowModal(ctx context.Context, modal Modal) error
	UpdateMessage(ctx context.Context, msg Message) error
	EditOriginal(ctx context.Context, msg Message) error
	FollowUp(ctx context.Context, msg Message) error
	CreateFollowUp(ctx context.Context, msg Message) (string, error)
	EditFollowUp(ctx context.Context, messageID string, msg Message) error
	DeleteFollowUp(ctx context.Context, messageID string) error
	Error(ctx context.Context, err error) error
}
