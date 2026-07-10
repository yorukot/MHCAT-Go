package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VoiceRoomConfigDocument struct {
	Guild         string `bson:"guild" json:"guild"`
	TicketChannel string `bson:"ticket_channel" json:"ticket_channel"`
	Limit         int    `bson:"limit" json:"limit"`
	Name          string `bson:"name" json:"name"`
	Parent        string `bson:"parent,omitempty" json:"parent,omitempty"`
	Lock          bool   `bson:"lock" json:"lock"`
}

type VoiceRoomConfigReadDocument struct {
	Guild         string        `bson:"guild" json:"guild"`
	TicketChannel string        `bson:"ticket_channel" json:"ticket_channel"`
	Limit         bson.RawValue `bson:"limit" json:"limit"`
	Name          bson.RawValue `bson:"name" json:"name"`
	Parent        bson.RawValue `bson:"parent" json:"parent"`
	Lock          bson.RawValue `bson:"lock" json:"lock"`
}

type VoiceRoomLockDocument struct {
	Guild        string   `bson:"guild" json:"guild"`
	ChannelID    string   `bson:"channel_id" json:"channel_id"`
	LockAnswer   *string  `bson:"lock_anser" json:"lock_anser"`
	Owner        string   `bson:"owner" json:"owner"`
	TextChannel  *string  `bson:"text_channel" json:"text_channel"`
	AllowedUsers []string `bson:"ok_people" json:"ok_people"`
}

type VoiceRoomLockReadDocument struct {
	Guild        string        `bson:"guild" json:"guild"`
	ChannelID    string        `bson:"channel_id" json:"channel_id"`
	LockAnswer   bson.RawValue `bson:"lock_anser" json:"lock_anser"`
	Owner        bson.RawValue `bson:"owner" json:"owner"`
	TextChannel  bson.RawValue `bson:"text_channel" json:"text_channel"`
	AllowedUsers bson.RawValue `bson:"ok_people" json:"ok_people"`
}

type VoiceRoomStateDocument struct {
	Guild     string `bson:"guild" json:"guild"`
	ChannelID string `bson:"channel_id" json:"channel_id"`
}

func VoiceRoomConfigDocumentFromDomain(config domain.VoiceRoomConfig) VoiceRoomConfigDocument {
	return VoiceRoomConfigDocument{
		Guild:         config.GuildID,
		TicketChannel: config.TriggerChannelID,
		Limit:         config.Limit,
		Name:          config.Name,
		Parent:        config.ParentID,
		Lock:          config.Lock,
	}
}

func (d VoiceRoomConfigDocument) ToDomain() domain.VoiceRoomConfig {
	return domain.VoiceRoomConfig{
		GuildID:          d.Guild,
		TriggerChannelID: d.TicketChannel,
		ParentID:         d.Parent,
		Name:             d.Name,
		Limit:            d.Limit,
		Lock:             d.Lock,
	}
}

func (d VoiceRoomConfigReadDocument) ToDomain() domain.VoiceRoomConfig {
	limit, ok := LegacyExactInt64(d.Limit)
	if !ok {
		limit = 0
	}
	name, _ := legacyMongooseString(d.Name)
	parent, _ := legacyNullableString(d.Parent)
	return domain.VoiceRoomConfig{
		GuildID:          d.Guild,
		TriggerChannelID: d.TicketChannel,
		ParentID:         parent,
		Name:             name,
		Limit:            int(limit),
		Lock:             legacyMongooseBoolean(d.Lock),
	}
}

func VoiceRoomLockDocumentFromDomain(lock domain.VoiceRoomLock) VoiceRoomLockDocument {
	lock = lock.Normalize()
	var password *string
	if lock.HasPassword() {
		password = &lock.Password
	}
	var textChannel *string
	if lock.TextChannelID != "" {
		textChannel = &lock.TextChannelID
	}
	return VoiceRoomLockDocument{
		Guild:        lock.GuildID,
		ChannelID:    lock.ChannelID,
		LockAnswer:   password,
		Owner:        lock.OwnerID,
		TextChannel:  textChannel,
		AllowedUsers: append([]string{}, lock.AllowedUserIDs...),
	}
}

func (d VoiceRoomLockDocument) ToDomain() domain.VoiceRoomLock {
	password := ""
	if d.LockAnswer != nil {
		password = *d.LockAnswer
	}
	textChannel := ""
	if d.TextChannel != nil {
		textChannel = *d.TextChannel
	}
	return domain.VoiceRoomLock{
		GuildID:         d.Guild,
		ChannelID:       d.ChannelID,
		Password:        password,
		PasswordPresent: d.LockAnswer != nil,
		OwnerID:         d.Owner,
		TextChannelID:   textChannel,
		AllowedUserIDs:  append([]string(nil), d.AllowedUsers...),
	}.Normalize()
}

func (d VoiceRoomLockReadDocument) ToDomain() domain.VoiceRoomLock {
	password, passwordPresent := legacyMongooseString(d.LockAnswer)
	owner, _ := legacyMongooseString(d.Owner)
	textChannel, _ := legacyNullableString(d.TextChannel)
	return domain.VoiceRoomLock{
		GuildID:         d.Guild,
		ChannelID:       d.ChannelID,
		Password:        password,
		PasswordPresent: passwordPresent,
		OwnerID:         owner,
		TextChannelID:   textChannel,
		AllowedUserIDs:  legacyVoiceRoomAllowedUsers(d.AllowedUsers),
	}.Normalize()
}

func legacyVoiceRoomAllowedUsers(value bson.RawValue) []string {
	array, ok := value.ArrayOK()
	if !ok {
		if text, ok := value.StringValueOK(); ok {
			return []string{text}
		}
		return nil
	}
	values, err := array.Values()
	if err != nil {
		return nil
	}
	users := make([]string, 0, len(values))
	for _, candidate := range values {
		if text, ok := candidate.StringValueOK(); ok {
			users = append(users, text)
		}
	}
	return users
}

func VoiceRoomStateDocumentFromDomain(state domain.VoiceRoomState) VoiceRoomStateDocument {
	return VoiceRoomStateDocument{
		Guild:     state.GuildID,
		ChannelID: state.ChannelID,
	}
}

func (d VoiceRoomStateDocument) ToDomain() domain.VoiceRoomState {
	return domain.VoiceRoomState{
		GuildID:   d.Guild,
		ChannelID: d.ChannelID,
	}
}
