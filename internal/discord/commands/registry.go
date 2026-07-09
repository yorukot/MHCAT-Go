package commands

func DefaultRegistry(scope Scope) Registry {
	return BuiltinRegistry(scope)
}

func EnabledDefinitions(registry Registry) []Definition {
	enabled := make([]Definition, 0, len(registry.Commands))
	for _, definition := range registry.Commands {
		if definition.Disabled || definition.Hidden || definition.Internal {
			continue
		}
		enabled = append(enabled, stripLocalOnly(definition))
	}
	sorted := NewRegistry(registry.Scope, enabled)
	return sorted.Commands
}

func stripLocalOnly(definition Definition) Definition {
	if definition.Type == 0 {
		definition.Type = CommandTypeChatInput
	}
	definition.Disabled = false
	definition.Hidden = false
	definition.Internal = false
	definition.DocsURL = ""
	definition.Ownership = nil
	return definition
}
