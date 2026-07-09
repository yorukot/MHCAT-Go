package commands

import (
	"fmt"
	"sort"
)

type StagingSyncOptions struct {
	Scope              Scope
	ExpectedCommands   []string
	AllowDelete        bool
	AllowBulkOverwrite bool
}

func ValidateStagingSync(registry Registry, opts StagingSyncOptions) error {
	if opts.Scope.Kind != ScopeGuild {
		return fmt.Errorf("%w: staging sync requires guild scope", ErrUnsafeOperation)
	}
	if opts.Scope.GuildID == "" {
		return fmt.Errorf("%w: staging sync requires guild id", ErrUnsafeOperation)
	}
	if opts.AllowDelete {
		return fmt.Errorf("%w: staging sync rejects delete", ErrUnsafeOperation)
	}
	if opts.AllowBulkOverwrite {
		return fmt.Errorf("%w: staging sync rejects bulk overwrite", ErrUnsafeOperation)
	}
	expected := stringSet(opts.ExpectedCommands)
	if len(expected) == 0 {
		return fmt.Errorf("%w: staging expected command list is empty", ErrUnsafeOperation)
	}
	seen := map[string]struct{}{}
	for _, definition := range registry.Commands {
		if definition.Disabled || definition.Hidden || definition.Internal {
			continue
		}
		if _, ok := expected[definition.Name]; !ok {
			return fmt.Errorf("%w: command %q is not allowed in staging sync", ErrUnsafeOperation, definition.Name)
		}
		if !IsManagedForScope(definition, opts.Scope) {
			return fmt.Errorf("%w: command %q is not marked managed for guild staging", ErrUnsafeOperation, definition.Name)
		}
		seen[definition.Name] = struct{}{}
	}
	missing := missingNames(expected, seen)
	if len(missing) > 0 {
		return fmt.Errorf("%w: staging registry missing expected commands: %v", ErrUnsafeOperation, missing)
	}
	return nil
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			result[value] = struct{}{}
		}
	}
	return result
}

func missingNames(expected, seen map[string]struct{}) []string {
	var missing []string
	for name := range expected {
		if _, ok := seen[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}
