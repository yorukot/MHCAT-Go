package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type Operation string

const (
	OperationCreate    Operation = "create"
	OperationUpdate    Operation = "update"
	OperationDelete    Operation = "delete"
	OperationUnchanged Operation = "unchanged"
	OperationSkipped   Operation = "skipped"
	OperationDangerous Operation = "dangerous"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

type RemoteCommand struct {
	ID            string     `json:"id,omitempty"`
	ApplicationID string     `json:"application_id,omitempty"`
	GuildID       string     `json:"guild_id,omitempty"`
	Version       string     `json:"version,omitempty"`
	Definition    Definition `json:"definition"`
	Owned         bool       `json:"owned,omitempty"`
}

type DiffOptions struct {
	Scope              Scope
	AllowDelete        bool
	AllowBulkOverwrite bool
}

type PlannedOperation struct {
	Operation   Operation   `json:"operation"`
	Scope       string      `json:"scope"`
	GuildID     string      `json:"guild_id,omitempty"`
	CommandType CommandType `json:"command_type"`
	CommandName string      `json:"command_name"`
	RemoteID    string      `json:"remote_id,omitempty"`
	Reason      string      `json:"reason"`
	BeforeHash  string      `json:"before_hash,omitempty"`
	AfterHash   string      `json:"after_hash,omitempty"`
	Risk        RiskLevel   `json:"risk"`
}

type Plan struct {
	Operations []PlannedOperation `json:"operations"`
}

func (p Plan) HasDangerous() bool {
	for _, op := range p.Operations {
		if op.Operation == OperationDangerous || op.Risk == RiskHigh {
			return true
		}
	}
	return false
}

func Diff(desired Registry, remote []RemoteCommand, opts DiffOptions) (Plan, error) {
	if err := ValidateRegistry(desired); err != nil {
		return Plan{}, err
	}
	scope := opts.Scope
	if scope.Kind == "" {
		scope = desired.Scope
	}
	if scope.Kind == "" {
		scope.Kind = ScopeGlobal
	}

	desiredCommands := EnabledDefinitions(desired)
	desiredByKey := map[string]Definition{}
	for _, definition := range desiredCommands {
		desiredByKey[commandKey(definition)] = definition
	}

	remoteByKey := map[string]RemoteCommand{}
	for _, command := range remote {
		remoteByKey[commandKey(command.Definition)] = command
	}

	var operations []PlannedOperation
	for _, definition := range desiredCommands {
		key := commandKey(definition)
		afterHash := StableHash(definition)
		if remoteCommand, ok := remoteByKey[key]; ok {
			beforeHash := StableHash(remoteCommand.Definition)
			if beforeHash == afterHash {
				operations = append(operations, planned(scope, OperationUnchanged, definition, remoteCommand.ID, "remote command matches desired definition", beforeHash, afterHash, RiskLow))
			} else {
				operations = append(operations, planned(scope, OperationUpdate, definition, remoteCommand.ID, "remote command differs from desired definition", beforeHash, afterHash, RiskMedium))
			}
			continue
		}
		operations = append(operations, planned(scope, OperationCreate, definition, "", "desired command is missing remotely", "", afterHash, RiskLow))
	}

	remoteKeys := make([]string, 0, len(remoteByKey))
	for key := range remoteByKey {
		remoteKeys = append(remoteKeys, key)
	}
	sort.Strings(remoteKeys)
	for _, key := range remoteKeys {
		if _, ok := desiredByKey[key]; ok {
			continue
		}
		remoteCommand := remoteByKey[key]
		beforeHash := StableHash(remoteCommand.Definition)
		if !remoteCommand.Owned {
			operations = append(operations, planned(scope, OperationSkipped, remoteCommand.Definition, remoteCommand.ID, "remote command is unknown or not owned by local registry", beforeHash, "", RiskHigh))
			continue
		}
		if !opts.AllowDelete {
			operations = append(operations, planned(scope, OperationSkipped, remoteCommand.Definition, remoteCommand.ID, "delete skipped because allow-delete is false", beforeHash, "", RiskHigh))
			continue
		}
		operations = append(operations, planned(scope, OperationDelete, remoteCommand.Definition, remoteCommand.ID, "owned remote command is absent locally and allow-delete is true", beforeHash, "", RiskHigh))
	}

	sortPlan(operations)
	return Plan{Operations: operations}, nil
}

func BulkOverwritePlan(desired Registry, opts DiffOptions) (Plan, error) {
	if err := ValidateRegistry(desired); err != nil {
		return Plan{}, err
	}
	scope := opts.Scope
	if scope.Kind == "" {
		scope = desired.Scope
	}
	if !opts.AllowBulkOverwrite {
		return Plan{Operations: []PlannedOperation{
			{
				Operation: OperationDangerous,
				Scope:     scope.Kind,
				GuildID:   scope.GuildID,
				Reason:    "bulk overwrite requested without allow-bulk-overwrite",
				Risk:      RiskHigh,
			},
		}}, nil
	}
	return Plan{Operations: []PlannedOperation{
		{
			Operation: OperationDangerous,
			Scope:     scope.Kind,
			GuildID:   scope.GuildID,
			Reason:    "bulk overwrite replaces the full remote command set",
			Risk:      RiskHigh,
		},
	}}, nil
}

func StableHash(definition Definition) string {
	canonical := stripLocalOnly(definition)
	payload, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func commandKey(definition Definition) string {
	return fmt.Sprintf("%d:%s", definition.Type, definition.Name)
}

func planned(scope Scope, operation Operation, definition Definition, remoteID, reason, beforeHash, afterHash string, risk RiskLevel) PlannedOperation {
	return PlannedOperation{
		Operation:   operation,
		Scope:       scope.Kind,
		GuildID:     scope.GuildID,
		CommandType: definition.Type,
		CommandName: definition.Name,
		RemoteID:    remoteID,
		Reason:      reason,
		BeforeHash:  beforeHash,
		AfterHash:   afterHash,
		Risk:        risk,
	}
}

func sortPlan(ops []PlannedOperation) {
	sort.SliceStable(ops, func(i, j int) bool {
		left := ops[i]
		right := ops[j]
		if left.Operation != right.Operation {
			return left.Operation < right.Operation
		}
		if left.CommandType != right.CommandType {
			return left.CommandType < right.CommandType
		}
		return left.CommandName < right.CommandName
	})
}
