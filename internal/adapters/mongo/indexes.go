package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IndexKey struct {
	Field string `json:"field"`
	Order int    `json:"order"`
}

type IndexSpec struct {
	Collection             string     `json:"collection"`
	Name                   string     `json:"name"`
	Keys                   []IndexKey `json:"keys"`
	Unique                 bool       `json:"unique,omitempty"`
	Sparse                 bool       `json:"sparse,omitempty"`
	PartialFilter          string     `json:"partial_filter,omitempty"`
	TTLSeconds             *int32     `json:"ttl_seconds,omitempty"`
	Reason                 string     `json:"reason,omitempty"`
	RequiresDuplicateAudit bool       `json:"requires_duplicate_audit,omitempty"`
	RequiresRetentionADR   bool       `json:"requires_retention_adr,omitempty"`
	RetentionADR           string     `json:"retention_adr,omitempty"`
}

type IndexInfo struct {
	Collection    string     `json:"collection,omitempty"`
	Name          string     `json:"name"`
	Keys          []IndexKey `json:"keys"`
	Unique        bool       `json:"unique,omitempty"`
	Sparse        bool       `json:"sparse,omitempty"`
	PartialFilter string     `json:"partial_filter,omitempty"`
	TTLSeconds    *int32     `json:"ttl_seconds,omitempty"`
}

type IndexPlan struct {
	Indexes []IndexSpec `json:"indexes"`
}

func DefaultIndexPlan(catalog []CollectionSpec) IndexPlan {
	var specs []IndexSpec
	for _, collection := range catalog {
		for _, index := range collection.PlannedIndexes {
			if index.Collection == "" {
				index.Collection = collection.Name
			}
			specs = append(specs, index)
		}
	}
	sortIndexSpecs(specs)
	return IndexPlan{Indexes: specs}
}

func LoadIndexPlan(r io.Reader) (IndexPlan, error) {
	var plan IndexPlan
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return IndexPlan{}, fmt.Errorf("decode index plan: %w", err)
	}
	sortIndexSpecs(plan.Indexes)
	return plan, nil
}

func (p IndexPlan) Validate() error {
	seen := map[string]struct{}{}
	for _, spec := range p.Indexes {
		if strings.TrimSpace(spec.Collection) == "" {
			return errors.New("index collection is required")
		}
		if strings.TrimSpace(spec.Name) == "" {
			return errors.New("index name is required")
		}
		if len(spec.Keys) == 0 {
			return fmt.Errorf("index %s.%s requires at least one key", spec.Collection, spec.Name)
		}
		for _, key := range spec.Keys {
			if strings.TrimSpace(key.Field) == "" {
				return fmt.Errorf("index %s.%s contains an empty field", spec.Collection, spec.Name)
			}
			if key.Order != 1 && key.Order != -1 {
				return fmt.Errorf("index %s.%s field %s has unsupported order %d", spec.Collection, spec.Name, key.Field, key.Order)
			}
		}
		if spec.Unique && !spec.RequiresDuplicateAudit {
			return fmt.Errorf("unique index %s.%s must require duplicate audit", spec.Collection, spec.Name)
		}
		if spec.TTLSeconds != nil && (*spec.TTLSeconds <= 0 || !spec.RequiresRetentionADR) {
			return fmt.Errorf("ttl index %s.%s must have positive ttl and require retention adr", spec.Collection, spec.Name)
		}
		key := spec.Collection + "/" + spec.Name
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate index spec %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func ListIndexes(ctx context.Context, database *drivermongo.Database, collections []string) (map[string][]IndexInfo, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	result := make(map[string][]IndexInfo, len(collections))
	for _, collectionName := range collections {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		cursor, err := database.Collection(collectionName).Indexes().List(ctx)
		if err != nil {
			return nil, MapError(fmt.Errorf("list indexes for %s: %w", collectionName, err))
		}
		var raw []bson.M
		if err := cursor.All(ctx, &raw); err != nil {
			return nil, MapError(fmt.Errorf("decode indexes for %s: %w", collectionName, err))
		}
		indexes := make([]IndexInfo, 0, len(raw))
		for _, document := range raw {
			indexes = append(indexes, indexInfoFromDocument(collectionName, document))
		}
		sortIndexInfos(indexes)
		result[collectionName] = indexes
	}
	return result, nil
}

func EnsureIndexes(ctx context.Context, database *drivermongo.Database, plan IndexPlan, operations []IndexOperation) error {
	if database == nil {
		return errors.New("mongo database is required")
	}
	for _, op := range operations {
		if op.Operation != IndexOperationCreate {
			continue
		}
		spec, ok := findIndexSpec(plan, op.Collection, op.IndexName)
		if !ok {
			return fmt.Errorf("missing index spec for %s.%s", op.Collection, op.IndexName)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		model := drivermongo.IndexModel{
			Keys:    indexKeysToBSON(spec.Keys),
			Options: indexOptions(spec),
		}
		if _, err := database.Collection(spec.Collection).Indexes().CreateOne(ctx, model); err != nil {
			return MapError(fmt.Errorf("create index %s.%s: %w", spec.Collection, spec.Name, err))
		}
	}
	return nil
}

func indexOptions(spec IndexSpec) *options.IndexOptionsBuilder {
	opts := options.Index().SetName(spec.Name)
	if spec.Unique {
		opts.SetUnique(true)
	}
	if spec.Sparse {
		opts.SetSparse(true)
	}
	if spec.TTLSeconds != nil {
		opts.SetExpireAfterSeconds(int32(*spec.TTLSeconds))
	}
	return opts
}

func indexKeysToBSON(keys []IndexKey) bson.D {
	result := make(bson.D, 0, len(keys))
	for _, key := range keys {
		result = append(result, bson.E{Key: key.Field, Value: key.Order})
	}
	return result
}

func indexInfoFromDocument(collection string, document bson.M) IndexInfo {
	info := IndexInfo{Collection: collection}
	if name, ok := document["name"].(string); ok {
		info.Name = name
	}
	info.Keys = keysFromAny(document["key"])
	if unique, ok := document["unique"].(bool); ok {
		info.Unique = unique
	}
	if sparse, ok := document["sparse"].(bool); ok {
		info.Sparse = sparse
	}
	if expire, ok := int32FromAny(document["expireAfterSeconds"]); ok {
		info.TTLSeconds = &expire
	}
	if partial, ok := document["partialFilterExpression"]; ok {
		payload, _ := json.Marshal(partial)
		info.PartialFilter = string(payload)
	}
	return info
}

func keysFromAny(value any) []IndexKey {
	switch keys := value.(type) {
	case bson.D:
		return keysFromBSOND(keys)
	case bson.M:
		result := make([]IndexKey, 0, len(keys))
		for field, order := range keys {
			if parsed, ok := intFromAny(order); ok {
				result = append(result, IndexKey{Field: field, Order: parsed})
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].Field < result[j].Field })
		return result
	case map[string]any:
		result := make([]IndexKey, 0, len(keys))
		for field, order := range keys {
			if parsed, ok := intFromAny(order); ok {
				result = append(result, IndexKey{Field: field, Order: parsed})
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].Field < result[j].Field })
		return result
	default:
		return nil
	}
}

func keysFromBSOND(keys bson.D) []IndexKey {
	result := make([]IndexKey, 0, len(keys))
	for _, key := range keys {
		if parsed, ok := intFromAny(key.Value); ok {
			result = append(result, IndexKey{Field: key.Key, Order: parsed})
		}
	}
	return result
}

func intFromAny(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}

func int32FromAny(value any) (int32, bool) {
	switch typed := value.(type) {
	case int32:
		return typed, true
	case int64:
		return int32(typed), true
	case int:
		return int32(typed), true
	case float64:
		return int32(typed), true
	default:
		return 0, false
	}
}

func findIndexSpec(plan IndexPlan, collection, name string) (IndexSpec, bool) {
	for _, spec := range plan.Indexes {
		if spec.Collection == collection && spec.Name == name {
			return spec, true
		}
	}
	return IndexSpec{}, false
}

func sortIndexSpecs(indexes []IndexSpec) {
	sort.SliceStable(indexes, func(i, j int) bool {
		if indexes[i].Collection != indexes[j].Collection {
			return indexes[i].Collection < indexes[j].Collection
		}
		return indexes[i].Name < indexes[j].Name
	})
}

func sortIndexInfos(indexes []IndexInfo) {
	sort.SliceStable(indexes, func(i, j int) bool {
		if indexes[i].Collection != indexes[j].Collection {
			return indexes[i].Collection < indexes[j].Collection
		}
		return indexes[i].Name < indexes[j].Name
	})
}

func ttl(seconds int32) *int32 {
	return &seconds
}

func secondsFromDuration(duration time.Duration) *int32 {
	seconds := int32(duration.Seconds())
	return &seconds
}
