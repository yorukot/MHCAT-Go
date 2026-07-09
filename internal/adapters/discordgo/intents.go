package discordgo

import dgo "github.com/bwmarrin/discordgo"

type IntentOptions struct {
	GuildMembers     bool
	GuildMessages    bool
	MessageReactions bool
	VoiceStates      bool
	MessageContent   bool
}

func BuildIntents(opts IntentOptions) dgo.Intent {
	intents := dgo.IntentsGuilds
	if opts.GuildMembers {
		intents |= dgo.IntentsGuildMembers
	}
	if opts.GuildMessages {
		intents |= dgo.IntentsGuildMessages
	}
	if opts.MessageReactions {
		intents |= dgo.IntentsGuildMessageReactions
	}
	if opts.VoiceStates {
		intents |= dgo.IntentsGuildVoiceStates
	}
	if opts.MessageContent {
		intents |= dgo.IntentsMessageContent
	}
	return intents
}
