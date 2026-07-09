package commands

func BuiltinDefinitions() []Definition {
	return []Definition{
		PingDefinition(),
		HelpDefinition(),
		InfoDefinition(),
	}
}

func TranslateDefinition() Definition {
	return Definition{
		Type:        CommandTypeChatInput,
		Name:        "翻譯",
		Description: "翻譯成各種語言",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/translate",
		Ownership:   ManagedOwnership("translate", ScopeGuild),
		Options: []Option{
			{
				Type:        OptionTypeString,
				Name:        "要的翻譯",
				Description: "你要翻譯的句子或是單詞!",
				Required:    true,
			},
			{
				Type:        OptionTypeString,
				Name:        "目標語言",
				Description: "你要翻譯成的語言!",
				Required:    true,
				Choices: []Choice{
					{Name: "🇹🇼中文(traditional Chinese)", Value: "zh-TW"},
					{Name: "🇺🇸英文(English)", Value: "en"},
					{Name: "🇯🇵日文(Japanese)", Value: "ja"},
					{Name: "🇰🇷韓語(Korean)", Value: "ko"},
					{Name: "🇩🇪德語(German)", Value: "de"},
					{Name: "🇫🇷法語(French)", Value: "fr"},
					{Name: "🇷🇺俄語(Russian)", Value: "ru"},
					{Name: "🇪🇸西班牙語(Spanish)", Value: "es"},
					{Name: "🇨🇳簡體中文(Simplified Chinese)", Value: "zh-CN"},
				},
			},
		},
	}
}

func BuiltinRegistry(scope Scope) Registry {
	return NewRegistry(scope, BuiltinDefinitions())
}

func PingDefinition() Definition {
	return Definition{
		Type:        CommandTypeChatInput,
		Name:        "ping",
		Description: "查看我的ping",
		Ownership:   ManagedOwnership("5.1", ScopeGuild),
	}
}

func HelpDefinition() Definition {
	return Definition{
		Type:        CommandTypeChatInput,
		Name:        "help",
		Description: "使用我開始使用",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/help",
		Ownership:   ManagedOwnership("5.1", ScopeGuild),
		Options: []Option{
			{
				Type:        OptionTypeString,
				Name:        "指令名稱",
				Description: "輸入指令名稱(可不輸入)!",
				Required:    false,
			},
		},
	}
}

func InfoDefinition() Definition {
	return Definition{
		Type:        CommandTypeChatInput,
		Name:        "info",
		Description: "Check all informations.",
		Ownership:   ManagedOwnership("5.1", ScopeGuild),
		NameLocalizations: map[string]string{
			"zh-TW": "資訊",
			"zh-CN": "资讯",
			"en-US": "info",
		},
		DescriptionLocalizations: map[string]string{
			"zh-TW": "各種資訊查詢!",
			"zh-CN": "各种资讯查询!",
			"en-US": "Check all informations.",
		},
		Options: []Option{
			{
				Type:        OptionTypeSubCommand,
				Name:        "user",
				Description: "Check a user's information!",
				NameLocalizations: map[string]string{
					"zh-TW": "使用者",
					"zh-CN": "使用者",
					"en-US": "user",
				},
				DescriptionLocalizations: map[string]string{
					"zh-TW": "查看某位使用者的資訊!",
					"zh-CN": "查看某位使用者的资讯!",
					"en-US": "Check a user's information!",
				},
				Options: []Option{{
					Type:        OptionTypeUser,
					Name:        "user",
					Description: "User to check",
					NameLocalizations: map[string]string{
						"zh-TW": "使用者",
						"zh-CN": "使用者",
						"en-US": "user",
					},
					DescriptionLocalizations: map[string]string{
						"zh-TW": "要查詢的使用者",
						"zh-CN": "要查询的使用者",
						"en-US": "User to check",
					},
				}},
			},
			{
				Type:        OptionTypeSubCommand,
				Name:        "bot",
				Description: "MHCAT about",
				NameLocalizations: map[string]string{
					"zh-TW": "機器人",
					"zh-CN": "机器人",
					"en-US": "bot",
				},
				DescriptionLocalizations: map[string]string{
					"zh-TW": "有關MHCAT的各種資訊",
					"zh-CN": "有关MHCAT的各种资讯",
					"en-US": "MHCAT about",
				},
			},
			{
				Type:        OptionTypeSubCommand,
				Name:        "shard",
				Description: "MHCat shard informations",
				NameLocalizations: map[string]string{
					"zh-TW": "分片",
					"zh-CN": "分片",
					"en-US": "shard",
				},
				DescriptionLocalizations: map[string]string{
					"zh-TW": "有關MHCAT分片的各種資訊",
					"zh-CN": "有关MHCAT分片的各种资讯",
					"en-US": "MHCat shard informations",
				},
			},
			{
				Type:        OptionTypeSubCommand,
				Name:        "guild",
				Description: "Server informations",
				NameLocalizations: map[string]string{
					"zh-TW": "伺服器",
					"zh-CN": "伺服器",
					"en-US": "guild",
				},
				DescriptionLocalizations: map[string]string{
					"zh-TW": "有关这个伺服器的各种资讯",
					"zh-CN": "有关这个伺服器的各种资讯",
					"en-US": "Server informations",
				},
			},
		},
	}
}
