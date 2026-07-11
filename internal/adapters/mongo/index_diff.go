package mongo

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
)

type IndexOperationKind string

const (
	IndexOperationCreate        IndexOperationKind = "create"
	IndexOperationExists        IndexOperationKind = "exists"
	IndexOperationChanged       IndexOperationKind = "changed"
	IndexOperationDangerous     IndexOperationKind = "dangerous"
	IndexOperationSkipped       IndexOperationKind = "skipped"
	IndexOperationUnknownRemote IndexOperationKind = "unknown_remote"
)

type IndexRiskLevel string

const (
	IndexRiskLow    IndexRiskLevel = "low"
	IndexRiskMedium IndexRiskLevel = "medium"
	IndexRiskHigh   IndexRiskLevel = "high"
)

type IndexDiffOptions struct {
	AllowUnique         bool
	AllowTTL            bool
	DuplicateAuditClean map[string]bool
}

type IndexOperation struct {
	Operation              IndexOperationKind `json:"operation"`
	Collection             string             `json:"collection"`
	IndexName              string             `json:"index_name"`
	Keys                   []IndexKey         `json:"keys,omitempty"`
	Unique                 bool               `json:"unique,omitempty"`
	Sparse                 bool               `json:"sparse,omitempty"`
	Partial                string             `json:"partial,omitempty"`
	TTLSeconds             *int32             `json:"ttl_seconds,omitempty"`
	Reason                 string             `json:"reason,omitempty"`
	Risk                   IndexRiskLevel     `json:"risk_level"`
	RequiresDuplicateAudit bool               `json:"requires_duplicate_audit,omitempty"`
	RequiresRetentionADR   bool               `json:"requires_retention_adr,omitempty"`
}

type IndexDiffPlan struct {
	Operations []IndexOperation `json:"operations"`
}

func DiffIndexes(desired IndexPlan, live map[string][]IndexInfo, opts IndexDiffOptions) (IndexDiffPlan, error) {
	if err := desired.Validate(); err != nil {
		return IndexDiffPlan{}, err
	}
	liveByKey := map[string]IndexInfo{}
	for collection, indexes := range live {
		for _, index := range indexes {
			if index.Collection == "" {
				index.Collection = collection
			}
			liveByKey[indexKey(index.Collection, index.Name)] = index
		}
	}
	desiredByKey := map[string]IndexSpec{}
	for _, spec := range desired.Indexes {
		desiredByKey[indexKey(spec.Collection, spec.Name)] = spec
	}

	var operations []IndexOperation
	for _, spec := range desired.Indexes {
		liveIndex, ok := liveByKey[indexKey(spec.Collection, spec.Name)]
		if !ok {
			if covering, covered := coveringLiveIndex(spec, live[spec.Collection]); covered {
				operations = append(operations, indexOperation(IndexOperationExists, spec, "lookup is covered by existing index "+covering.Name, IndexRiskLow))
				continue
			}
			if conflict, blocked := conflictingLiveFallback(spec, live[spec.Collection]); blocked {
				op := indexOperation(IndexOperationDangerous, spec, "remove same-key non-unique fallback "+conflict.Name+" before creating unique index", IndexRiskHigh)
				op.RequiresDuplicateAudit = true
				operations = append(operations, op)
				continue
			}
			if covering, covered := approvedDesiredUniqueIndex(spec, desired.Indexes, opts); covered {
				operations = append(operations, indexOperation(IndexOperationExists, spec, "lookup will be covered by approved unique index "+covering.Name, IndexRiskLow))
				continue
			}
			operations = append(operations, missingIndexOperation(spec, opts))
			continue
		}
		if equivalentIndex(spec, liveIndex) {
			operations = append(operations, indexOperation(IndexOperationExists, spec, "index already exists with matching options", IndexRiskLow))
			continue
		}
		operations = append(operations, indexOperation(IndexOperationChanged, spec, "existing index differs; Wave 3 will not drop or modify indexes", IndexRiskHigh))
	}

	var remoteKeys []string
	for key := range liveByKey {
		remoteKeys = append(remoteKeys, key)
	}
	sort.Strings(remoteKeys)
	for _, key := range remoteKeys {
		if _, ok := desiredByKey[key]; ok {
			continue
		}
		liveIndex := liveByKey[key]
		operations = append(operations, IndexOperation{
			Operation:  IndexOperationUnknownRemote,
			Collection: liveIndex.Collection,
			IndexName:  liveIndex.Name,
			Keys:       append([]IndexKey(nil), liveIndex.Keys...),
			Unique:     liveIndex.Unique,
			Sparse:     liveIndex.Sparse,
			Partial:    liveIndex.PartialFilter,
			TTLSeconds: liveIndex.TTLSeconds,
			Reason:     "remote index is not in local plan; Wave 3 never drops indexes",
			Risk:       IndexRiskLow,
		})
	}
	sortIndexOperations(operations)
	return IndexDiffPlan{Operations: operations}, nil
}

func approvedDesiredUniqueIndex(spec IndexSpec, desired []IndexSpec, opts IndexDiffOptions) (IndexSpec, bool) {
	if spec.Unique || spec.Sparse || spec.PartialFilter != "" || spec.TTLSeconds != nil || !opts.AllowUnique {
		return IndexSpec{}, false
	}
	for _, candidate := range desired {
		if candidate.Collection != spec.Collection || !candidate.Unique || candidate.Sparse || candidate.PartialFilter != "" || candidate.TTLSeconds != nil {
			continue
		}
		if !opts.DuplicateAuditClean[indexKey(candidate.Collection, candidate.Name)] || !indexKeysHavePrefix(candidate.Keys, spec.Keys) {
			continue
		}
		return candidate, true
	}
	return IndexSpec{}, false
}

func conflictingLiveFallback(spec IndexSpec, indexes []IndexInfo) (IndexInfo, bool) {
	if !spec.Unique || spec.Sparse || spec.PartialFilter != "" || spec.TTLSeconds != nil {
		return IndexInfo{}, false
	}
	for _, index := range indexes {
		if index.Unique || index.Sparse || index.PartialFilter != "" || index.TTLSeconds != nil || len(index.Keys) != len(spec.Keys) {
			continue
		}
		if indexKeysHavePrefix(index.Keys, spec.Keys) {
			return index, true
		}
	}
	return IndexInfo{}, false
}

func coveringLiveIndex(spec IndexSpec, indexes []IndexInfo) (IndexInfo, bool) {
	if spec.Unique || spec.Sparse || spec.PartialFilter != "" || spec.TTLSeconds != nil {
		return IndexInfo{}, false
	}
	for _, index := range indexes {
		if index.Sparse || index.PartialFilter != "" || index.TTLSeconds != nil || len(index.Keys) < len(spec.Keys) {
			continue
		}
		if indexKeysHavePrefix(index.Keys, spec.Keys) {
			return index, true
		}
	}
	return IndexInfo{}, false
}

func indexKeysHavePrefix(keys []IndexKey, prefix []IndexKey) bool {
	if len(keys) < len(prefix) {
		return false
	}
	for index := range prefix {
		if keys[index] != prefix[index] {
			return false
		}
	}
	return true
}

func FormatIndexDiffPlan(w io.Writer, plan IndexDiffPlan, format string) error {
	normalized := IndexDiffPlan{Operations: append([]IndexOperation(nil), plan.Operations...)}
	sortIndexOperations(normalized.Operations)
	if format == "json" {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(normalized)
	}
	for _, op := range normalized.Operations {
		if _, err := fmt.Fprintf(
			w,
			"%s collection=%s index=%s unique=%t ttl=%v risk=%s reason=%q\n",
			op.Operation,
			op.Collection,
			op.IndexName,
			op.Unique,
			op.TTLSeconds,
			op.Risk,
			op.Reason,
		); err != nil {
			return err
		}
	}
	return nil
}

func SafeIndexApplyOperations(plan IndexDiffPlan) []IndexOperation {
	var operations []IndexOperation
	for _, op := range plan.Operations {
		if op.Operation == IndexOperationCreate && op.Risk != IndexRiskHigh {
			operations = append(operations, op)
		}
	}
	sortIndexOperations(operations)
	return operations
}

func missingIndexOperation(spec IndexSpec, opts IndexDiffOptions) IndexOperation {
	if spec.Unique {
		if !opts.AllowUnique {
			op := indexOperation(IndexOperationDangerous, spec, "unique index requires --allow-unique and clean duplicate audit", IndexRiskHigh)
			op.RequiresDuplicateAudit = true
			return op
		}
		if !opts.DuplicateAuditClean[indexKey(spec.Collection, spec.Name)] {
			op := indexOperation(IndexOperationDangerous, spec, "unique index requires clean duplicate audit", IndexRiskHigh)
			op.RequiresDuplicateAudit = true
			return op
		}
	}
	if spec.TTLSeconds != nil {
		if !opts.AllowTTL {
			op := indexOperation(IndexOperationDangerous, spec, "ttl index requires --allow-ttl and retention ADR", IndexRiskHigh)
			op.RequiresRetentionADR = true
			return op
		}
		if spec.RequiresRetentionADR && spec.RetentionADR == "" {
			op := indexOperation(IndexOperationDangerous, spec, "ttl index requires retention ADR note", IndexRiskHigh)
			op.RequiresRetentionADR = true
			return op
		}
	}
	return indexOperation(IndexOperationCreate, spec, "planned index is missing remotely", IndexRiskMedium)
}

func indexOperation(kind IndexOperationKind, spec IndexSpec, reason string, risk IndexRiskLevel) IndexOperation {
	return IndexOperation{
		Operation:              kind,
		Collection:             spec.Collection,
		IndexName:              spec.Name,
		Keys:                   append([]IndexKey(nil), spec.Keys...),
		Unique:                 spec.Unique,
		Sparse:                 spec.Sparse,
		Partial:                spec.PartialFilter,
		TTLSeconds:             spec.TTLSeconds,
		Reason:                 reason,
		Risk:                   risk,
		RequiresDuplicateAudit: spec.RequiresDuplicateAudit,
		RequiresRetentionADR:   spec.RequiresRetentionADR,
	}
}

func equivalentIndex(spec IndexSpec, live IndexInfo) bool {
	return reflect.DeepEqual(spec.Keys, live.Keys) &&
		spec.Unique == live.Unique &&
		spec.Sparse == live.Sparse &&
		spec.PartialFilter == live.PartialFilter &&
		ttlEqual(spec.TTLSeconds, live.TTLSeconds)
}

func ttlEqual(left, right *int32) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func indexKey(collection, name string) string {
	return collection + "/" + name
}

func sortIndexOperations(operations []IndexOperation) {
	sort.SliceStable(operations, func(i, j int) bool {
		if operations[i].Collection != operations[j].Collection {
			return operations[i].Collection < operations[j].Collection
		}
		if operations[i].IndexName != operations[j].IndexName {
			return operations[i].IndexName < operations[j].IndexName
		}
		return operations[i].Operation < operations[j].Operation
	})
}
