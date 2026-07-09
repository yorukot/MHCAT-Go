package parity

import (
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestParseLegacySlashCommandExtractsMetadataAndOptions(t *testing.T) {
	payload := []byte(`
module.exports = {
  name: '測試',
  cooldown: 10,
  description: '測試描述',
  UserPerms: '訊息管理',
  options: [{
    name: '設定',
    type: ApplicationCommandOptionType.Subcommand,
    description: '設定描述',
    options: [{
      name: '頻道',
      type: ApplicationCommandOptionType.Channel,
      description: '輸入頻道',
      required: true,
    }]
  }],
  run: async () => {}
}`)
	command := ParseLegacySlashCommand("slashCommands/分類/test.js", payload)
	if command.Name != "測試" || command.Description != "測試描述" || command.UserPerms != "訊息管理" || command.Cooldown != "10" {
		t.Fatalf("command = %#v", command)
	}
	if command.Category != "分類" {
		t.Fatalf("category = %q", command.Category)
	}
	if len(command.Options) != 1 || command.Options[0].Name != "設定" || command.Options[0].Type != "subcommand" {
		t.Fatalf("options = %#v", command.Options)
	}
	child := command.Options[0].Options[0]
	if child.Name != "頻道" || child.Type != "channel" || child.Required == nil || !*child.Required {
		t.Fatalf("child option = %#v", child)
	}
}

func TestObjectFieldsKeepsFirstMetadataValue(t *testing.T) {
	fields := objectFields(`{
  name: '代幣遊戲',
  description: '遊玩有關代幣的小遊戲',
  run: async () => {
    interaction.reply({ name: client.user.username })
  },
  name: client.user.username
}`)
	name, ok := jsStringValue(fields["name"])
	if !ok || name != "代幣遊戲" {
		t.Fatalf("name = %q ok=%v fields=%#v", name, ok, fields)
	}
}

func TestParseLegacySlashCommandFallsBackToMetadataBeforeRun(t *testing.T) {
	payload := []byte("module.exports = {\n" +
		"  name: '代幣遊戲',\n" +
		"  cooldown: 10,\n" +
		"  description: '遊玩有關代幣的小遊戲',\n" +
		"  options: [{ name: '21點', type: ApplicationCommandOptionType.Subcommand, description: '跟真人遊玩21點!!' }],\n" +
		"  run: async () => {\n" +
		"    const result = `${winner ? `${winner.username}` : `平手`}`\n" +
		"    return result\n" +
		"  }\n" +
		"}")
	command := ParseLegacySlashCommand("slashCommands/代幣系統/game.js", payload)
	if command.Name != "代幣遊戲" || command.Description != "遊玩有關代幣的小遊戲" || command.Cooldown != "10" {
		t.Fatalf("command = %#v", command)
	}
	if len(command.Options) != 1 || command.Options[0].Name != "21點" {
		t.Fatalf("options = %#v", command.Options)
	}
}

func TestParseLegacySlashCommandExtractsReportWebOptions(t *testing.T) {
	payload := []byte("module.exports = {\n" +
		"  name: '詐騙網址回報',\n" +
		"  cooldown: 10,\n" +
		"  description: '回報詐騙網站',\n" +
		"  //video: 'https://docsmhcat.yorukot.me/commands/statistics.html',\n" +
		"  emoji: '<:fraudalert:1000408260777611355>',\n" +
		"  options: [{\n" +
		"    name: '網址',\n" +
		"    type: ApplicationCommandOptionType.String,\n" +
		"    description: '回報網址',\n" +
		"    required: true,\n" +
		"  }],\n" +
		"  run: async (client, interaction, options, perms) => {\n" +
		"    try {\n" +
		"      await interaction.deferReply();\n" +
		"      const web = interaction.options.getString(\"網址\")\n" +
		"      const dsadsa = new WebhookClient({url:`${process.env.REPORT_WEBHOOK}`})\n" +
		"      dsadsa.send(`\\`\\`\\`${web}\\`\\`\\`\\nby:<@${interaction.user.id}>`)\n" +
		"    } catch (error) {\n" +
		"      error_send(error, interaction)\n" +
		"    }\n" +
		"  }\n" +
		"}")
	command := ParseLegacySlashCommand("slashCommands/群組防護/report_web.js", payload)
	if command.Name != "詐騙網址回報" || command.Description != "回報詐騙網站" {
		t.Fatalf("command = %#v", command)
	}
	if len(command.Options) != 1 {
		t.Fatalf("options = %#v warnings=%#v", command.Options, command.Warnings)
	}
	option := command.Options[0]
	if option.Name != "網址" || option.Type != "string" || option.Description != "回報網址" || option.Required == nil || !*option.Required {
		t.Fatalf("option = %#v", option)
	}
}

func TestAuditSlashCommandParityClassifiesMissingMatchingAndDrift(t *testing.T) {
	required := true
	legacy := []LegacyCommand{
		{Name: "match", File: "slashCommands/a/match.js", Description: "same", Options: []LegacyOption{{Name: "value", Type: "string", Description: "value", Required: &required}}},
		{Name: "drift", File: "slashCommands/a/drift.js", Description: "legacy"},
		{Name: "missing", File: "slashCommands/a/missing.js", Description: "missing"},
	}
	definitions := []commands.Definition{
		{Name: "match", Type: commands.CommandTypeChatInput, Description: "same", Options: []commands.Option{{Name: "value", Type: commands.OptionTypeString, Description: "value", Required: true}}},
		{Name: "drift", Type: commands.CommandTypeChatInput, Description: "go"},
		{Name: "extra", Type: commands.CommandTypeChatInput, Description: "extra"},
	}
	audit := AuditSlashCommandParity(legacy, definitions)
	if len(audit.MatchingCommands) != 1 || audit.MatchingCommands[0].Name != "match" {
		t.Fatalf("matching = %#v", audit.MatchingCommands)
	}
	if len(audit.CommandsWithDrift) != 1 || audit.CommandsWithDrift[0].Name != "drift" || len(audit.CommandsWithDrift[0].Findings) == 0 {
		t.Fatalf("drift = %#v", audit.CommandsWithDrift)
	}
	if len(audit.MissingGoDefinitions) != 1 || audit.MissingGoDefinitions[0].Name != "missing" {
		t.Fatalf("missing = %#v", audit.MissingGoDefinitions)
	}
	if len(audit.ExtraGoDefinitions) != 1 || audit.ExtraGoDefinitions[0].Name != "extra" {
		t.Fatalf("extra = %#v", audit.ExtraGoDefinitions)
	}
}

func TestCurrentGoDefinitionsIncludesSplitFeatureDefinitions(t *testing.T) {
	var foundWarningSettings bool
	var foundWarningRemove bool
	var foundWarningRemoveAll bool
	var foundCleanup bool
	var foundDeleteData bool
	var foundCoinAdmin bool
	var foundCoinRank bool
	var foundGachaPrizeCreate bool
	var foundGachaPrizeDelete bool
	for _, definition := range CurrentGoDefinitions() {
		if definition.Name == "警告設定" {
			foundWarningSettings = true
		}
		if definition.Name == "警告清除" {
			foundWarningRemove = true
		}
		if definition.Name == "警告全部清除" {
			foundWarningRemoveAll = true
		}
		if definition.Name == "刪除訊息" {
			foundCleanup = true
		}
		if definition.Name == "刪除資料" {
			foundDeleteData = true
		}
		if definition.Name == "代幣增加" {
			foundCoinAdmin = true
		}
		if definition.Name == "代幣排行榜" {
			foundCoinRank = true
		}
		if definition.Name == "扭蛋獎池增加" {
			foundGachaPrizeCreate = true
		}
		if definition.Name == "扭蛋獎池刪除" {
			foundGachaPrizeDelete = true
		}
	}
	if !foundWarningSettings {
		t.Fatal("current Go definitions should include warning settings")
	}
	if !foundWarningRemove || !foundWarningRemoveAll {
		t.Fatal("current Go definitions should include warning removal commands")
	}
	if !foundCleanup {
		t.Fatal("current Go definitions should include message cleanup")
	}
	if !foundDeleteData {
		t.Fatal("current Go definitions should include delete data")
	}
	if !foundCoinAdmin {
		t.Fatal("current Go definitions should include coin admin")
	}
	if !foundCoinRank {
		t.Fatal("current Go definitions should include coin rank")
	}
	if !foundGachaPrizeCreate {
		t.Fatal("current Go definitions should include gacha prize create")
	}
	if !foundGachaPrizeDelete {
		t.Fatal("current Go definitions should include gacha prize delete")
	}
}

func TestAuditSlashCommandParityPreservesUnnamedParserWarnings(t *testing.T) {
	audit := AuditSlashCommandParity([]LegacyCommand{
		{File: "slashCommands/a/bad.js", Warnings: []string{"module.exports object did not close"}},
	}, nil)
	if audit.LegacyParseErrorCount != 1 {
		t.Fatalf("parse error count = %d", audit.LegacyParseErrorCount)
	}
	if len(audit.LegacyParserWarnings) != 1 || !strings.Contains(audit.LegacyParserWarnings[0], "module.exports object did not close") {
		t.Fatalf("warnings = %#v", audit.LegacyParserWarnings)
	}
}

func TestRenderMarkdownIncludesSummary(t *testing.T) {
	markdown := RenderMarkdown(CommandAudit{
		LegacyFileCount:      2,
		LegacyUniqueCommands: 2,
		GoDefinitionCount:    1,
		MissingGoDefinitions: []LegacyCommand{{Name: "missing", File: "slashCommands/a/missing.js"}},
	})
	for _, want := range []string{"# Slash Command UI Parity Audit", "Legacy slash command files: 2", "`missing`"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown missing %q:\n%s", want, markdown)
		}
	}
}
