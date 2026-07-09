package gacha

import (
	"fmt"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDrawDefinitionMatchesLegacy(t *testing.T) {
	definition := DrawDefinition()
	if definition.Name != GachaDrawCommandName || definition.Description != "進行扭蛋，有機會抽中各種大獎喔!!!!" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("expected one option, got %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeString || option.Name != gachaDrawMultiOption || option.Description != "如果需要連抽的話可以使用這個指令" || option.Required {
		t.Fatalf("option = %#v", option)
	}
	if len(option.Choices) != 4 {
		t.Fatalf("choices = %#v", option.Choices)
	}
	for index, want := range []struct {
		name  string
		value string
	}{
		{"5連抽(無buff)", "5"},
		{"10連抽(多送一抽)", "11"},
		{"15連抽(多送兩抽)", "17"},
		{"20連抽(多送三抽)", "23"},
	} {
		choice := option.Choices[index]
		if choice.Name != want.name || fmt.Sprint(choice.Value) != want.value {
			t.Fatalf("choice %d = %#v", index, choice)
		}
	}
	if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "gacha-draw" || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestPrizeListDefinitionMatchesLegacy(t *testing.T) {
	definition := PrizeListDefinition()
	if definition.Name != GachaPrizeListCommandName || definition.Description != "增加扭蛋的獎池" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("expected no options, got %#v", definition.Options)
	}
	if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "gacha-prize-list" || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, PrizeListDefinitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestPrizeDeleteDefinitionMatchesLegacy(t *testing.T) {
	definition := PrizeDeleteDefinition()
	if definition.Name != GachaPrizeDeleteCommandName || definition.Description != "刪除扭蛋的獎池" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != gachaManageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("expected one option, got %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeString || option.Name != gachaPrizeNameOption || option.Description != "輸入這個獎品叫甚麼，以及簡單概述" || !option.Required {
		t.Fatalf("option = %#v", option)
	}
	if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "gacha-prize-delete" || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, AllDefinitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestPrizeCreateDefinitionMatchesLegacy(t *testing.T) {
	definition := PrizeCreateDefinition()
	if definition.Name != GachaPrizeCreateCommandName || definition.Description != "增加扭蛋的獎池" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != gachaManageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 6 {
		t.Fatalf("expected six options, got %#v", definition.Options)
	}
	expected := []struct {
		name        string
		optionType  commands.OptionType
		description string
		required    bool
	}{
		{gachaPrizeNameOption, commands.OptionTypeString, "輸入這個獎品叫甚麼，以及簡單概述", true},
		{gachaPrizeChanceOption, commands.OptionTypeNumber, "輸入中獎機率(百分比)ex:10 代表10% 0.1代表0.1%", true},
		{gachaPrizeCodeOption, commands.OptionTypeString, "填上獎品的代碼ex:stram序號nitro連結等", false},
		{gachaPrizeAutoDeleteOption, commands.OptionTypeBoolean, "抽中後是否自動刪除(預設為true，如果填否的話連獎品數量都不會變)", false},
		{gachaPrizeCountOption, commands.OptionTypeInteger, "該獎品的數量", false},
		{gachaPrizeGiveCoinOption, commands.OptionTypeInteger, "當抽中後是否要給予代幣", false},
	}
	for index, want := range expected {
		option := definition.Options[index]
		if option.Name != want.name || option.Type != want.optionType || option.Description != want.description || option.Required != want.required {
			t.Fatalf("option %d = %#v", index, option)
		}
	}
	if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "gacha-prize-create" || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
}

func TestPrizeEditDefinitionMatchesLegacy(t *testing.T) {
	definition := PrizeEditDefinition()
	if definition.Name != GachaPrizeEditCommandName || definition.Description != "增加扭蛋的獎池裡的獎品的數量" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != gachaManageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 6 {
		t.Fatalf("expected six options, got %#v", definition.Options)
	}
	expected := []struct {
		name        string
		optionType  commands.OptionType
		description string
		required    bool
	}{
		{gachaPrizeNameOption, commands.OptionTypeString, "輸入這個獎品叫甚麼，以及簡單概述", true},
		{gachaPrizeChanceOption, commands.OptionTypeNumber, "輸入中獎機率(百分比)ex:10 代表10% 0.1代表0.1%", false},
		{gachaPrizeCodeOption, commands.OptionTypeString, "填上獎品的代碼ex:stram序號nitro連結等", false},
		{gachaPrizeAutoDeleteOption, commands.OptionTypeBoolean, "抽中後是否自動刪除(預設為true，如果填否的話連獎品數量都不會變)", false},
		{gachaPrizeCountOption, commands.OptionTypeInteger, "該獎品的數量", false},
		{gachaPrizeGiveCoinOption, commands.OptionTypeInteger, "當抽中後是否要給予代幣", false},
	}
	for index, want := range expected {
		option := definition.Options[index]
		if option.Name != want.name || option.Type != want.optionType || option.Description != want.description || option.Required != want.required {
			t.Fatalf("option %d = %#v", index, option)
		}
	}
	if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "gacha-prize-edit" || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, AllDefinitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
