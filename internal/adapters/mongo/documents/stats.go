package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type StatsConfigDocument struct {
	Guild             string  `bson:"guild" json:"guild"`
	Parent            string  `bson:"parent" json:"parent"`
	MemberNumber      *string `bson:"memberNumber" json:"memberNumber"`
	MemberNumberName  *string `bson:"memberNumber_name" json:"memberNumber_name"`
	UserNumber        *string `bson:"userNumber" json:"userNumber"`
	UserNumberName    *string `bson:"userNumber_name" json:"userNumber_name"`
	BotNumber         *string `bson:"BotNumber" json:"BotNumber"`
	BotNumberName     *string `bson:"BotNumber_name" json:"BotNumber_name"`
	ChannelNumber     *string `bson:"channelnumber" json:"channelnumber"`
	ChannelNumberName *string `bson:"channelnumber_name" json:"channelnumber_name"`
	TextNumber        *string `bson:"textnumber" json:"textnumber"`
	TextNumberName    *string `bson:"textnumber_name" json:"textnumber_name"`
	VoiceNumber       *string `bson:"voicenumber" json:"voicenumber"`
	VoiceNumberName   *string `bson:"voicenumber_name" json:"voicenumber_name"`
}

func (d StatsConfigDocument) ToDomain() domain.StatsConfig {
	return domain.StatsConfig{
		GuildID:           d.Guild,
		ParentID:          d.Parent,
		MemberNumberID:    statsString(d.MemberNumber),
		MemberNumberName:  statsString(d.MemberNumberName),
		UserNumberID:      statsString(d.UserNumber),
		UserNumberName:    statsString(d.UserNumberName),
		BotNumberID:       statsString(d.BotNumber),
		BotNumberName:     statsString(d.BotNumberName),
		ChannelNumberID:   statsString(d.ChannelNumber),
		ChannelNumberName: statsString(d.ChannelNumberName),
		TextNumberID:      statsString(d.TextNumber),
		TextNumberName:    statsString(d.TextNumberName),
		VoiceNumberID:     statsString(d.VoiceNumber),
		VoiceNumberName:   statsString(d.VoiceNumberName),
	}
}

func statsString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
