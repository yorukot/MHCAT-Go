package commands

const OwnerMHCATRefactor = "mhcat-refactor"

type CommandOwnership struct {
	Owner      string   `json:"owner,omitempty"`
	Managed    bool     `json:"managed,omitempty"`
	SinceWave  string   `json:"since_wave,omitempty"`
	SafeScopes []string `json:"safe_scopes,omitempty"`
}

func ManagedOwnership(sinceWave string, safeScopes ...string) *CommandOwnership {
	return &CommandOwnership{
		Owner:      OwnerMHCATRefactor,
		Managed:    true,
		SinceWave:  sinceWave,
		SafeScopes: append([]string(nil), safeScopes...),
	}
}

func IsManagedForScope(definition Definition, scope Scope) bool {
	ownership := definition.Ownership
	if ownership == nil || !ownership.Managed || ownership.Owner != OwnerMHCATRefactor {
		return false
	}
	for _, safeScope := range ownership.SafeScopes {
		if safeScope == scope.Kind {
			return true
		}
	}
	return false
}
