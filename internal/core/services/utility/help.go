package utility

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

var ErrHelpCommandNotFound = errors.New("help command not found")

type HelpService struct {
	registry commands.Registry
}

func NewHelpService(registry commands.Registry) HelpService {
	registry.Sort()
	return HelpService{registry: registry}
}

func (s HelpService) Overview() string {
	definitions := publicDefinitions(s.registry)
	lines := []string{
		"MHCAT help",
		"Implemented commands:",
	}
	if len(definitions) == 0 {
		lines = append(lines, "- no commands implemented yet")
		return strings.Join(lines, "\n")
	}
	for _, definition := range definitions {
		lines = append(lines, fmt.Sprintf("- /%s: %s", definition.Name, definition.Description))
	}
	lines = append(lines, "Use /help 指令名稱:<command> for details.")
	return strings.Join(lines, "\n")
}

func (s HelpService) Detail(name string) (string, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return s.Overview(), nil
	}
	for _, definition := range publicDefinitions(s.registry) {
		if definition.Name != name {
			continue
		}
		lines := []string{
			"Command detail",
			fmt.Sprintf("Name: /%s", definition.Name),
			fmt.Sprintf("Description: %s", definition.Description),
		}
		if definition.DocsURL != "" {
			lines = append(lines, fmt.Sprintf("Docs: %s", definition.DocsURL))
		}
		if len(definition.Options) > 0 {
			lines = append(lines, "Options:")
			for _, option := range definition.Options {
				lines = append(lines, fmt.Sprintf("- %s: %s", option.Name, option.Description))
			}
		}
		return strings.Join(lines, "\n"), nil
	}
	return "", fmt.Errorf("%w: %s", ErrHelpCommandNotFound, name)
}

func publicDefinitions(registry commands.Registry) []commands.Definition {
	definitions := make([]commands.Definition, 0, len(registry.Commands))
	for _, definition := range registry.Commands {
		if definition.Disabled || definition.Hidden || definition.Internal {
			continue
		}
		definitions = append(definitions, definition)
	}
	sort.SliceStable(definitions, func(i, j int) bool {
		return definitions[i].Name < definitions[j].Name
	})
	return definitions
}
