package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

var ErrUnsafeOperation = errors.New("unsafe command sync operation")

type SyncClient interface {
	ListCommands(ctx context.Context, scope Scope) ([]RemoteCommand, error)
	CreateCommand(ctx context.Context, scope Scope, definition Definition) (RemoteCommand, error)
	UpdateCommand(ctx context.Context, scope Scope, remoteID string, definition Definition) (RemoteCommand, error)
	DeleteCommand(ctx context.Context, scope Scope, remoteID string) error
	BulkOverwriteCommands(ctx context.Context, scope Scope, definitions []Definition) ([]RemoteCommand, error)
}

type SyncOptions struct {
	Scope              Scope
	DryRun             bool
	AllowDelete        bool
	AllowBulkOverwrite bool
}

type SyncResult struct {
	Plan    Plan `json:"plan"`
	Applied bool `json:"applied"`
	Writes  int  `json:"writes"`
}

func PlanSync(ctx context.Context, client SyncClient, registry Registry, opts SyncOptions) (Plan, error) {
	if client == nil {
		return Plan{}, errors.New("command sync client is required")
	}
	remoteCommands, err := client.ListCommands(ctx, opts.Scope)
	if err != nil {
		return Plan{}, fmt.Errorf("list remote commands: %w", err)
	}
	return Diff(registry, remoteCommands, DiffOptions{
		Scope:              opts.Scope,
		AllowDelete:        opts.AllowDelete,
		AllowBulkOverwrite: opts.AllowBulkOverwrite,
	})
}

func ExecutePlan(ctx context.Context, client SyncClient, registry Registry, plan Plan, opts SyncOptions) (SyncResult, error) {
	result := SyncResult{Plan: plan, Applied: !opts.DryRun}
	if opts.DryRun {
		return result, nil
	}

	desired := map[string]Definition{}
	for _, definition := range EnabledDefinitions(registry) {
		desired[commandKey(definition)] = definition
	}

	for _, operation := range plan.Operations {
		switch operation.Operation {
		case OperationCreate:
			definition, ok := desired[operationKey(operation)]
			if !ok {
				return result, fmt.Errorf("desired command missing for create %s", operation.CommandName)
			}
			if _, err := client.CreateCommand(ctx, opts.Scope, definition); err != nil {
				return result, fmt.Errorf("create command %s: %w", operation.CommandName, err)
			}
			result.Writes++
		case OperationUpdate:
			definition, ok := desired[operationKey(operation)]
			if !ok {
				return result, fmt.Errorf("desired command missing for update %s", operation.CommandName)
			}
			if _, err := client.UpdateCommand(ctx, opts.Scope, operation.RemoteID, definition); err != nil {
				return result, fmt.Errorf("update command %s: %w", operation.CommandName, err)
			}
			result.Writes++
		case OperationDelete:
			if !opts.AllowDelete {
				return result, fmt.Errorf("%w: delete requires allow-delete", ErrUnsafeOperation)
			}
			if err := client.DeleteCommand(ctx, opts.Scope, operation.RemoteID); err != nil {
				return result, fmt.Errorf("delete command %s: %w", operation.CommandName, err)
			}
			result.Writes++
		case OperationDangerous:
			return result, fmt.Errorf("%w: dangerous operation requires explicit purpose-built path", ErrUnsafeOperation)
		}
	}
	return result, nil
}

func FormatPlan(w io.Writer, plan Plan, format string) error {
	normalized := Plan{Operations: append([]PlannedOperation(nil), plan.Operations...)}
	sortPlan(normalized.Operations)
	if format == "json" {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(normalized)
	}
	for _, op := range normalized.Operations {
		if _, err := fmt.Fprintf(
			w,
			"%s scope=%s guild=%s type=%d name=%s remote_id=%s risk=%s reason=%q before=%s after=%s\n",
			op.Operation,
			op.Scope,
			op.GuildID,
			op.CommandType,
			op.CommandName,
			op.RemoteID,
			op.Risk,
			op.Reason,
			op.BeforeHash,
			op.AfterHash,
		); err != nil {
			return err
		}
	}
	return nil
}

func operationKey(operation PlannedOperation) string {
	return fmt.Sprintf("%d:%s", operation.CommandType, operation.CommandName)
}

func SortRemote(remote []RemoteCommand) {
	sort.SliceStable(remote, func(i, j int) bool {
		left := remote[i].Definition
		right := remote[j].Definition
		if left.Type != right.Type {
			return left.Type < right.Type
		}
		return left.Name < right.Name
	})
}
