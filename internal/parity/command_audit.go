package parity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureannouncements "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/announcements"
	featureautochat "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/autochat"
	featurebalance "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/balance"
	featurebirthday "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/birthday"
	featureeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/economy"
	featuregacha "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/gacha"
	featurelogging "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/logging"
	featurelottery "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/lottery"
	featuremoderation "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/moderation"
	featurenotifications "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/notifications"
	featureonboarding "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/onboarding"
	featurepoll "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/poll"
	featureredeem "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/redeem"
	featuresafety "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/safety"
	featurestats "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/stats"
	featureticket "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/ticket"
	featurevoice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/voice"
	featurework "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/work"
	featurexp "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/xp"
)

type LegacyCommand struct {
	File        string         `json:"file"`
	Category    string         `json:"category"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	UserPerms   string         `json:"user_perms,omitempty"`
	Cooldown    string         `json:"cooldown,omitempty"`
	Options     []LegacyOption `json:"options,omitempty"`
	Warnings    []string       `json:"warnings,omitempty"`
}

type LegacyOption struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Type        string         `json:"type,omitempty"`
	Required    *bool          `json:"required,omitempty"`
	Options     []LegacyOption `json:"options,omitempty"`
}

type CommandAudit struct {
	LegacyFileCount       int                 `json:"legacy_file_count"`
	LegacyUniqueCommands  int                 `json:"legacy_unique_commands"`
	GoDefinitionCount     int                 `json:"go_definition_count"`
	MatchingCommands      []CommandComparison `json:"matching_commands"`
	CommandsWithDrift     []CommandComparison `json:"commands_with_drift"`
	MissingGoDefinitions  []LegacyCommand     `json:"missing_go_definitions"`
	ExtraGoDefinitions    []GoCommandSummary  `json:"extra_go_definitions"`
	DuplicateLegacyNames  []DuplicateName     `json:"duplicate_legacy_names,omitempty"`
	DuplicateGoNames      []DuplicateName     `json:"duplicate_go_names,omitempty"`
	LegacyParserWarnings  []string            `json:"legacy_parser_warnings,omitempty"`
	LegacyParseErrorCount int                 `json:"legacy_parse_error_count"`
}

type CommandComparison struct {
	Name     string         `json:"name"`
	File     string         `json:"file"`
	Status   string         `json:"status"`
	Findings []AuditFinding `json:"findings,omitempty"`
}

type AuditFinding struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Legacy  string `json:"legacy,omitempty"`
	Go      string `json:"go,omitempty"`
}

type GoCommandSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Ownership   string `json:"ownership,omitempty"`
}

type DuplicateName struct {
	Name  string   `json:"name"`
	Files []string `json:"files,omitempty"`
}

func CurrentGoDefinitions() []commands.Definition {
	definitions := commands.BuiltinDefinitions()
	definitions = append(definitions, commands.TranslateDefinition())
	definitions = append(definitions, featureticket.Definitions()...)
	definitions = append(definitions, featurepoll.Definitions()...)
	definitions = append(definitions, featureeconomy.Definitions()...)
	definitions = append(definitions, featureeconomy.SignInDefinitions()...)
	definitions = append(definitions, featureeconomy.SettingsDefinitions()...)
	definitions = append(definitions, featureeconomy.CoinAdminDefinitions()...)
	definitions = append(definitions, featureeconomy.CoinRankDefinitions()...)
	definitions = append(definitions, featureeconomy.RockPaperScissorsDefinitions()...)
	definitions = append(definitions, featureeconomy.ProfileDefinitions()...)
	definitions = append(definitions, featurework.Definitions()...)
	definitions = append(definitions, featuremoderation.Definitions()...)
	definitions = append(definitions, featuremoderation.SettingsDefinitions()...)
	definitions = append(definitions, featuremoderation.RemovalDefinitions()...)
	definitions = append(definitions, featuremoderation.IssueDefinitions()...)
	definitions = append(definitions, featuremoderation.CleanupDefinitions()...)
	definitions = append(definitions, featuremoderation.DeleteDataDefinitions()...)
	definitions = append(definitions, featurebalance.Definitions()...)
	definitions = append(definitions, featureredeem.Definitions()...)
	definitions = append(definitions, featureautochat.Definitions()...)
	definitions = append(definitions, featurenotifications.Definitions()...)
	definitions = append(definitions, featuresafety.Definitions()...)
	definitions = append(definitions, featurelogging.Definitions()...)
	definitions = append(definitions, featuregacha.AllDefinitions()...)
	definitions = append(definitions, featurelottery.Definitions()...)
	definitions = append(definitions, featurestats.Definitions()...)
	definitions = append(definitions, featurebirthday.Definitions()...)
	definitions = append(definitions, featureannouncements.ConfigDefinitions()...)
	definitions = append(definitions, featureannouncements.SendDefinitions()...)
	definitions = append(definitions, featurexp.TextDefinitions()...)
	definitions = append(definitions, featurexp.VoiceDefinitions()...)
	definitions = append(definitions, featurexp.RewardRoleDefinitions()...)
	definitions = append(definitions, featurexp.DisabledProfileDefinitions()...)
	definitions = append(definitions, featurexp.AdminDefinitions()...)
	definitions = append(definitions, featurexp.ResetDefinitions()...)
	definitions = append(definitions, featurevoice.Definitions()...)
	definitions = append(definitions, featurevoice.LockDefinitions()...)
	definitions = append(definitions, featureonboarding.JoinRoleDefinitions()...)
	definitions = append(definitions, featureonboarding.MessageDefinitions()...)
	definitions = append(definitions, featureonboarding.VerificationDefinitions()...)
	definitions = append(definitions, featureonboarding.VerificationFlowDefinitions()...)
	definitions = append(definitions, featureonboarding.AccountAgeDefinitions()...)
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "parity-audit"}, definitions)
	return registry.Commands
}

func LoadLegacySlashCommands(legacyRoot string) ([]LegacyCommand, error) {
	slashRoot := filepath.Join(legacyRoot, "slashCommands")
	entries := []LegacyCommand{}
	err := filepath.WalkDir(slashRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".js" {
			return nil
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(legacyRoot, path)
		if err != nil {
			return err
		}
		command := ParseLegacySlashCommand(filepath.ToSlash(relative), payload)
		entries = append(entries, command)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Category != entries[j].Category {
			return entries[i].Category < entries[j].Category
		}
		return entries[i].File < entries[j].File
	})
	return entries, nil
}

func ParseLegacySlashCommand(relativeFile string, payload []byte) LegacyCommand {
	command := LegacyCommand{
		File:     relativeFile,
		Category: legacyCategory(relativeFile),
	}
	body, err := extractModuleObject(string(payload))
	if err != nil {
		command.Warnings = append(command.Warnings, err.Error())
		return command
	}
	fields := objectFields(body)
	command.Name, _ = jsStringValue(fields["name"])
	command.Description, _ = jsStringValue(fields["description"])
	command.UserPerms, _ = jsStringValue(fields["UserPerms"])
	command.Cooldown = numberValue(fields["cooldown"])
	if command.Name == "" {
		command.Warnings = append(command.Warnings, "missing command name")
	}
	if command.Description == "" {
		command.Warnings = append(command.Warnings, "missing command description")
	}
	if rawOptions, ok := fields["options"]; ok {
		options, warnings := parseLegacyOptions(rawOptions)
		command.Options = options
		command.Warnings = append(command.Warnings, warnings...)
	}
	return command
}

func AuditSlashCommandParity(legacy []LegacyCommand, goDefinitions []commands.Definition) CommandAudit {
	audit := CommandAudit{
		LegacyFileCount:   len(legacy),
		GoDefinitionCount: len(goDefinitions),
	}
	legacyByName := map[string][]LegacyCommand{}
	for _, command := range legacy {
		if command.Name == "" {
			audit.LegacyParseErrorCount++
			if len(command.Warnings) == 0 {
				audit.LegacyParserWarnings = append(audit.LegacyParserWarnings, command.File+": missing command name")
			}
			for _, warning := range command.Warnings {
				audit.LegacyParserWarnings = append(audit.LegacyParserWarnings, command.File+": "+warning)
			}
			continue
		}
		if len(command.Warnings) > 0 {
			audit.LegacyParseErrorCount++
			for _, warning := range command.Warnings {
				audit.LegacyParserWarnings = append(audit.LegacyParserWarnings, command.File+": "+warning)
			}
		}
		legacyByName[command.Name] = append(legacyByName[command.Name], command)
	}
	audit.LegacyUniqueCommands = len(legacyByName)
	for name, commands := range legacyByName {
		if len(commands) > 1 {
			files := make([]string, 0, len(commands))
			for _, command := range commands {
				files = append(files, command.File)
			}
			audit.DuplicateLegacyNames = append(audit.DuplicateLegacyNames, DuplicateName{Name: name, Files: files})
		}
	}

	goByName := map[string][]commands.Definition{}
	for _, definition := range goDefinitions {
		goByName[definition.Name] = append(goByName[definition.Name], definition)
	}
	for name, definitions := range goByName {
		if len(definitions) > 1 {
			audit.DuplicateGoNames = append(audit.DuplicateGoNames, DuplicateName{Name: name})
		}
	}

	names := sortedKeys(legacyByName)
	for _, name := range names {
		legacyCommand := legacyByName[name][0]
		definitions := goByName[name]
		if len(definitions) == 0 {
			audit.MissingGoDefinitions = append(audit.MissingGoDefinitions, legacyCommand)
			continue
		}
		comparison := compareCommand(legacyCommand, definitions[0])
		if len(comparison.Findings) == 0 {
			comparison.Status = "matching-definition"
			audit.MatchingCommands = append(audit.MatchingCommands, comparison)
		} else {
			comparison.Status = "ui-drift-review"
			audit.CommandsWithDrift = append(audit.CommandsWithDrift, comparison)
		}
	}
	for _, name := range sortedKeys(goByName) {
		if _, ok := legacyByName[name]; !ok {
			definition := goByName[name][0]
			audit.ExtraGoDefinitions = append(audit.ExtraGoDefinitions, GoCommandSummary{
				Name:        definition.Name,
				Description: definition.Description,
				Ownership:   ownershipLabel(definition),
			})
		}
	}
	sort.Slice(audit.DuplicateLegacyNames, func(i, j int) bool { return audit.DuplicateLegacyNames[i].Name < audit.DuplicateLegacyNames[j].Name })
	sort.Slice(audit.DuplicateGoNames, func(i, j int) bool { return audit.DuplicateGoNames[i].Name < audit.DuplicateGoNames[j].Name })
	sort.Strings(audit.LegacyParserWarnings)
	return audit
}

func RenderMarkdown(audit CommandAudit) string {
	var out strings.Builder
	fmt.Fprintln(&out, "# Slash Command UI Parity Audit")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Generated from the current local legacy `MHCAT/slashCommands` tree and the current Go command definitions.")
	fmt.Fprintln(&out, "This is a static slash-command metadata audit for names, descriptions, options, and option flags; handler behavior remains covered by feature-specific tests and docs.")
	fmt.Fprintln(&out, "Rerun with `go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown`.")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "## Summary")
	fmt.Fprintln(&out)
	fmt.Fprintf(&out, "- Legacy slash command files: %d\n", audit.LegacyFileCount)
	fmt.Fprintf(&out, "- Legacy unique command names: %d\n", audit.LegacyUniqueCommands)
	fmt.Fprintf(&out, "- Current Go command definitions: %d\n", audit.GoDefinitionCount)
	fmt.Fprintf(&out, "- Matching command definitions: %d\n", len(audit.MatchingCommands))
	fmt.Fprintf(&out, "- Implemented definitions needing UI review: %d\n", len(audit.CommandsWithDrift))
	fmt.Fprintf(&out, "- Legacy commands without Go definitions: %d\n", len(audit.MissingGoDefinitions))
	fmt.Fprintf(&out, "- Go definitions without a legacy command name: %d\n", len(audit.ExtraGoDefinitions))
	fmt.Fprintf(&out, "- Legacy parse warning/error files: %d\n", audit.LegacyParseErrorCount)

	writeComparisons(&out, "Matching Definitions", audit.MatchingCommands, false)
	writeComparisons(&out, "Implemented Definitions Needing UI Review", audit.CommandsWithDrift, true)
	writeMissing(&out, audit.MissingGoDefinitions)
	writeExtra(&out, audit.ExtraGoDefinitions)
	writeDuplicates(&out, "Duplicate Legacy Command Names", audit.DuplicateLegacyNames)
	writeDuplicates(&out, "Duplicate Go Command Names", audit.DuplicateGoNames)
	writeWarnings(&out, audit.LegacyParserWarnings)
	return out.String()
}

func RenderJSON(audit CommandAudit) ([]byte, error) {
	return json.MarshalIndent(audit, "", "  ")
}

func compareCommand(legacy LegacyCommand, definition commands.Definition) CommandComparison {
	comparison := CommandComparison{Name: legacy.Name, File: legacy.File}
	if legacy.Description != "" && definition.Description != legacy.Description {
		comparison.Findings = append(comparison.Findings, AuditFinding{
			Path:    "description",
			Message: "command description differs",
			Legacy:  legacy.Description,
			Go:      definition.Description,
		})
	}
	compareOptions(&comparison, legacy.Options, definition.Options, "options")
	return comparison
}

func compareOptions(comparison *CommandComparison, legacy []LegacyOption, goOptions []commands.Option, path string) {
	if len(legacy) != len(goOptions) {
		comparison.Findings = append(comparison.Findings, AuditFinding{
			Path:    path,
			Message: "option count differs",
			Legacy:  strconv.Itoa(len(legacy)),
			Go:      strconv.Itoa(len(goOptions)),
		})
	}
	goByName := map[string]commands.Option{}
	for _, option := range goOptions {
		goByName[option.Name] = option
		for _, localized := range option.NameLocalizations {
			if localized != "" {
				goByName[localized] = option
			}
		}
	}
	for _, legacyOption := range legacy {
		optionPath := path + "." + legacyOption.Name
		goOption, ok := goByName[legacyOption.Name]
		if !ok {
			comparison.Findings = append(comparison.Findings, AuditFinding{
				Path:    optionPath,
				Message: "legacy option missing from Go definition",
				Legacy:  legacyOption.Type,
			})
			continue
		}
		if legacyOption.Type != "" && goOptionTypeName(goOption.Type) != legacyOption.Type {
			comparison.Findings = append(comparison.Findings, AuditFinding{
				Path:    optionPath + ".type",
				Message: "option type differs",
				Legacy:  legacyOption.Type,
				Go:      goOptionTypeName(goOption.Type),
			})
		}
		if legacyOption.Description != "" && goOption.Description != legacyOption.Description && goOption.DescriptionLocalizations["zh-TW"] != legacyOption.Description {
			comparison.Findings = append(comparison.Findings, AuditFinding{
				Path:    optionPath + ".description",
				Message: "option description differs",
				Legacy:  legacyOption.Description,
				Go:      goOption.Description,
			})
		}
		if legacyOption.Required != nil && goOption.Required != *legacyOption.Required {
			comparison.Findings = append(comparison.Findings, AuditFinding{
				Path:    optionPath + ".required",
				Message: "required flag differs",
				Legacy:  strconv.FormatBool(*legacyOption.Required),
				Go:      strconv.FormatBool(goOption.Required),
			})
		}
		compareOptions(comparison, legacyOption.Options, goOption.Options, optionPath+".options")
	}
}

func parseLegacyOptions(raw string) ([]LegacyOption, []string) {
	elements, warnings := arrayObjects(raw)
	options := make([]LegacyOption, 0, len(elements))
	for _, element := range elements {
		fields := objectFields(element)
		option := LegacyOption{}
		option.Name, _ = jsStringValue(fields["name"])
		option.Description, _ = jsStringValue(fields["description"])
		option.Type = normalizeLegacyOptionType(fields["type"])
		option.Required = boolPointer(fields["required"])
		if rawChildren, ok := fields["options"]; ok {
			children, childWarnings := parseLegacyOptions(rawChildren)
			option.Options = children
			warnings = append(warnings, childWarnings...)
		}
		if option.Name == "" {
			warnings = append(warnings, "option without parsed name")
		}
		options = append(options, option)
	}
	return options, warnings
}

func extractModuleObject(input string) (string, error) {
	index := strings.Index(input, "module.exports")
	if index < 0 {
		return "", errors.New("module.exports object not found")
	}
	start := strings.Index(input[index:], "{")
	if start < 0 {
		return "", errors.New("module.exports opening object not found")
	}
	start += index
	end, ok := matchingDelimiter(input, start, '{', '}')
	if ok {
		return input[start : end+1], nil
	}
	if prefix, ok := moduleObjectMetadataPrefix(input[start:]); ok {
		return prefix, nil
	}
	return "", errors.New("module.exports object did not close")
}

func objectFields(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		raw = strings.TrimSpace(raw[1 : len(raw)-1])
	}
	fields := map[string]string{}
	for position := 0; position < len(raw); {
		position = skipSpaceAndCommas(raw, position)
		key, next, ok := readFieldKey(raw, position)
		if !ok {
			break
		}
		next = skipSpace(raw, next)
		if next >= len(raw) || raw[next] != ':' {
			position = next + 1
			continue
		}
		next++
		start := skipSpace(raw, next)
		end := scanFieldValue(raw, start)
		if _, exists := fields[key]; !exists {
			fields[key] = strings.TrimSpace(raw[start:end])
		}
		position = end + 1
	}
	return fields
}

func moduleObjectMetadataPrefix(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") {
		return "", false
	}
	for position := 1; position < len(raw); {
		position = skipSpaceAndCommas(raw, position)
		if position >= len(raw) {
			break
		}
		key, next, ok := readFieldKey(raw, position)
		if !ok {
			return "", false
		}
		next = skipSpace(raw, next)
		if next >= len(raw) || raw[next] != ':' {
			return "", false
		}
		if key == "run" {
			return raw[:position] + "}", true
		}
		next++
		start := skipSpace(raw, next)
		end := scanFieldValue(raw, start)
		if end <= start {
			return "", false
		}
		position = end + 1
	}
	return "", false
}

func arrayObjects(raw string) ([]string, []string) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "[") {
		return nil, []string{"options value was not an array"}
	}
	end, ok := matchingDelimiter(raw, 0, '[', ']')
	if !ok {
		return nil, []string{"options array did not close"}
	}
	body := raw[1:end]
	var objects []string
	var warnings []string
	for position := 0; position < len(body); {
		position = skipSpaceAndCommas(body, position)
		if position >= len(body) {
			break
		}
		if body[position] != '{' {
			warnings = append(warnings, "non-object option entry skipped")
			position++
			continue
		}
		objectEnd, ok := matchingDelimiter(body, position, '{', '}')
		if !ok {
			warnings = append(warnings, "option object did not close")
			break
		}
		objects = append(objects, body[position:objectEnd+1])
		position = objectEnd + 1
	}
	return objects, warnings
}

func matchingDelimiter(input string, start int, open byte, close byte) (int, bool) {
	depth := 0
	var quote byte
	escaped := false
	for index := start; index < len(input); index++ {
		ch := input[index]
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' || ch == '`' {
			quote = ch
			continue
		}
		if ch == open {
			depth++
		}
		if ch == close {
			depth--
			if depth == 0 {
				return index, true
			}
		}
	}
	return 0, false
}

func scanFieldValue(input string, start int) int {
	depth := 0
	var quote byte
	escaped := false
	for index := start; index < len(input); index++ {
		ch := input[index]
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}
		switch ch {
		case '\'', '"', '`':
			quote = ch
		case '{', '[', '(':
			depth++
		case '}', ']', ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				return index
			}
		}
	}
	return len(input)
}

func readFieldKey(input string, start int) (string, int, bool) {
	if start >= len(input) {
		return "", start, false
	}
	if input[start] == '\'' || input[start] == '"' || input[start] == '`' {
		value, end, ok := readQuoted(input, start)
		return value, end, ok
	}
	end := start
	for end < len(input) {
		ch := input[end]
		if !(ch == '_' || ch == '$' || ch == '-' || ch >= '0' && ch <= '9' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z') {
			break
		}
		end++
	}
	if end == start {
		return "", start, false
	}
	return input[start:end], end, true
}

func jsStringValue(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	value, _, ok := readQuoted(raw, 0)
	return value, ok
}

func readQuoted(input string, start int) (string, int, bool) {
	if start >= len(input) {
		return "", start, false
	}
	quote := input[start]
	if quote != '\'' && quote != '"' && quote != '`' {
		return "", start, false
	}
	var out bytes.Buffer
	escaped := false
	for index := start + 1; index < len(input); index++ {
		ch := input[index]
		if escaped {
			switch ch {
			case 'n':
				out.WriteByte('\n')
			case 't':
				out.WriteByte('\t')
			default:
				out.WriteByte(ch)
			}
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == quote {
			return out.String(), index + 1, true
		}
		out.WriteByte(ch)
	}
	return "", start, false
}

func numberValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	end := 0
	for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
		end++
	}
	return raw[:end]
}

func boolPointer(raw string) *bool {
	switch strings.TrimSpace(raw) {
	case "true":
		value := true
		return &value
	case "false":
		value := false
		return &value
	default:
		return nil
	}
}

func normalizeLegacyOptionType(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if index := strings.LastIndex(raw, "."); index >= 0 {
		raw = raw[index+1:]
	}
	raw = strings.Trim(raw, ", \n\r\t")
	switch strings.ToLower(raw) {
	case "subcommand", "sub_command":
		return "subcommand"
	case "subcommandgroup", "sub_command_group":
		return "subcommandgroup"
	case "string":
		return "string"
	case "integer":
		return "integer"
	case "boolean":
		return "boolean"
	case "user":
		return "user"
	case "channel":
		return "channel"
	case "role":
		return "role"
	case "mentionable":
		return "mentionable"
	case "number":
		return "number"
	case "attachment":
		return "attachment"
	default:
		return strings.ToLower(raw)
	}
}

func goOptionTypeName(optionType commands.OptionType) string {
	switch optionType {
	case commands.OptionTypeSubCommand:
		return "subcommand"
	case commands.OptionTypeSubCommandGroup:
		return "subcommandgroup"
	case commands.OptionTypeString:
		return "string"
	case commands.OptionTypeInteger:
		return "integer"
	case commands.OptionTypeBoolean:
		return "boolean"
	case commands.OptionTypeUser:
		return "user"
	case commands.OptionTypeChannel:
		return "channel"
	case commands.OptionTypeRole:
		return "role"
	case commands.OptionTypeMentionable:
		return "mentionable"
	case commands.OptionTypeNumber:
		return "number"
	case commands.OptionTypeAttachment:
		return "attachment"
	default:
		return strconv.Itoa(int(optionType))
	}
}

func ownershipLabel(definition commands.Definition) string {
	if definition.Ownership == nil {
		return ""
	}
	return definition.Ownership.SinceWave
}

func legacyCategory(relativeFile string) string {
	parts := strings.Split(filepath.ToSlash(relativeFile), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func skipSpaceAndCommas(input string, position int) int {
	for position < len(input) {
		ch := input[position]
		if ch == ',' || ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
			position++
			continue
		}
		if strings.HasPrefix(input[position:], "//") {
			position += len("//")
			for position < len(input) && input[position] != '\n' {
				position++
			}
			continue
		}
		if strings.HasPrefix(input[position:], "/*") {
			end := strings.Index(input[position+len("/*"):], "*/")
			if end < 0 {
				return len(input)
			}
			position += len("/*") + end + len("*/")
			continue
		}
		break
	}
	return position
}

func skipSpace(input string, position int) int {
	for position < len(input) {
		ch := input[position]
		if ch != ' ' && ch != '\n' && ch != '\r' && ch != '\t' {
			break
		}
		position++
	}
	return position
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func writeComparisons(out *strings.Builder, title string, comparisons []CommandComparison, includeFindings bool) {
	fmt.Fprintln(out)
	fmt.Fprintf(out, "## %s\n\n", title)
	if len(comparisons) == 0 {
		fmt.Fprintln(out, "None.")
		return
	}
	fmt.Fprintln(out, "| Command | Legacy file | Status | Findings |")
	fmt.Fprintln(out, "| --- | --- | --- | --- |")
	for _, comparison := range comparisons {
		findings := "none"
		if includeFindings && len(comparison.Findings) > 0 {
			parts := make([]string, 0, len(comparison.Findings))
			for _, finding := range comparison.Findings {
				parts = append(parts, markdownEscape(finding.Path+": "+finding.Message+" ["+finding.Legacy+" -> "+finding.Go+"]"))
			}
			findings = strings.Join(parts, "<br>")
		}
		fmt.Fprintf(out, "| `%s` | `%s` | %s | %s |\n", markdownEscape(comparison.Name), markdownEscape(comparison.File), comparison.Status, findings)
	}
}

func writeMissing(out *strings.Builder, commands []LegacyCommand) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "## Legacy Commands Without Go Definitions")
	fmt.Fprintln(out)
	if len(commands) == 0 {
		fmt.Fprintln(out, "None.")
		return
	}
	fmt.Fprintln(out, "| Command | Category | Legacy file | Description |")
	fmt.Fprintln(out, "| --- | --- | --- | --- |")
	for _, command := range commands {
		fmt.Fprintf(out, "| `%s` | %s | `%s` | %s |\n", markdownEscape(command.Name), markdownEscape(command.Category), markdownEscape(command.File), markdownEscape(command.Description))
	}
}

func writeExtra(out *strings.Builder, definitions []GoCommandSummary) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "## Go Definitions Without Legacy Command Names")
	fmt.Fprintln(out)
	if len(definitions) == 0 {
		fmt.Fprintln(out, "None.")
		return
	}
	fmt.Fprintln(out, "| Command | Description | Ownership |")
	fmt.Fprintln(out, "| --- | --- | --- |")
	for _, definition := range definitions {
		fmt.Fprintf(out, "| `%s` | %s | %s |\n", markdownEscape(definition.Name), markdownEscape(definition.Description), markdownEscape(definition.Ownership))
	}
}

func writeDuplicates(out *strings.Builder, title string, duplicates []DuplicateName) {
	if len(duplicates) == 0 {
		return
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "## %s\n\n", title)
	fmt.Fprintln(out, "| Command | Files |")
	fmt.Fprintln(out, "| --- | --- |")
	for _, duplicate := range duplicates {
		fmt.Fprintf(out, "| `%s` | %s |\n", markdownEscape(duplicate.Name), markdownEscape(strings.Join(duplicate.Files, "<br>")))
	}
}

func writeWarnings(out *strings.Builder, warnings []string) {
	if len(warnings) == 0 {
		return
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "## Parser Warnings")
	fmt.Fprintln(out)
	for _, warning := range warnings {
		fmt.Fprintf(out, "- %s\n", markdownEscape(warning))
	}
}

func markdownEscape(value string) string {
	value = strings.ReplaceAll(value, "\n", "<br>")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}
