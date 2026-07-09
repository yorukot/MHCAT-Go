package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FieldSpec struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type LogicalKeySpec struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`
}

type CollectionSpec struct {
	Name                string           `json:"name"`
	LegacyModelFile     string           `json:"legacy_model_file,omitempty"`
	LegacyMongooseModel string           `json:"legacy_mongoose_model,omitempty"`
	RequiredFields      []FieldSpec      `json:"required_fields,omitempty"`
	LogicalKeys         []LogicalKeySpec `json:"logical_keys,omitempty"`
	PlannedIndexes      []IndexSpec      `json:"planned_indexes,omitempty"`
	Incomplete          bool             `json:"incomplete,omitempty"`
	Notes               string           `json:"notes,omitempty"`
}

type AuditOptions struct {
	SampleLimit           int
	LargeDocumentBytes    int
	DuplicateAuditEnabled bool
}

type AuditReport struct {
	Database           string            `json:"database"`
	Collections        []CollectionAudit `json:"collections"`
	UnknownCollections []string          `json:"unknown_collections,omitempty"`
	MissingCollections []string          `json:"missing_collections,omitempty"`
	Warnings           []AuditWarning    `json:"warnings,omitempty"`
}

type CollectionAudit struct {
	Name                  string               `json:"name"`
	DocumentCount         int64                `json:"document_count"`
	Indexes               []IndexInfo          `json:"indexes,omitempty"`
	SampledDocuments      int                  `json:"sampled_documents"`
	FieldTypes            map[string][]string  `json:"field_types,omitempty"`
	MissingRequiredFields []FieldIssue         `json:"missing_required_fields,omitempty"`
	DuplicateKeyRisks     []DuplicateKeyRisk   `json:"duplicate_key_risks,omitempty"`
	LargeDocuments        []LargeDocumentIssue `json:"large_documents,omitempty"`
}

type FieldIssue struct {
	Field string `json:"field"`
	Count int    `json:"count"`
}

type DuplicateKeyRisk struct {
	KeyName         string   `json:"key_name"`
	Fields          []string `json:"fields"`
	DuplicateGroups int64    `json:"duplicate_groups"`
	Examples        []string `json:"examples,omitempty"`
}

type LargeDocumentIssue struct {
	DocumentRef string `json:"document_ref,omitempty"`
	Bytes       int    `json:"bytes"`
}

type AuditWarning struct {
	Collection string `json:"collection,omitempty"`
	Field      string `json:"field,omitempty"`
	Message    string `json:"message"`
}

type SampleDocument struct {
	Ref       string            `json:"ref,omitempty"`
	SizeBytes int               `json:"size_bytes"`
	Fields    map[string]string `json:"fields"`
}

type CollectionSnapshot struct {
	Name           string
	DocumentCount  int64
	Indexes        []IndexInfo
	Samples        []SampleDocument
	DuplicateRisks []DuplicateKeyRisk
}

func AuditDatabase(ctx context.Context, database *drivermongo.Database, catalog []CollectionSpec, opts AuditOptions) (AuditReport, error) {
	if database == nil {
		return AuditReport{}, errors.New("mongo database is required")
	}
	if opts.SampleLimit < 0 {
		return AuditReport{}, errors.New("sample limit must be non-negative")
	}
	names, err := database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return AuditReport{}, MapError(fmt.Errorf("list collections: %w", err))
	}
	sort.Strings(names)
	indexes, err := ListIndexes(ctx, database, names)
	if err != nil {
		return AuditReport{}, err
	}
	snapshots := make([]CollectionSnapshot, 0, len(names))
	catalogByName := catalogMap(catalog)
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return AuditReport{}, err
		}
		collection := database.Collection(name)
		count, err := collection.CountDocuments(ctx, bson.D{})
		if err != nil {
			return AuditReport{}, MapError(fmt.Errorf("count documents for %s: %w", name, err))
		}
		samples, err := sampleCollection(ctx, collection, opts.SampleLimit)
		if err != nil {
			return AuditReport{}, err
		}
		var duplicateRisks []DuplicateKeyRisk
		if opts.DuplicateAuditEnabled {
			if spec, ok := catalogByName[name]; ok {
				duplicateRisks, err = auditDuplicateKeys(ctx, collection, spec.LogicalKeys)
				if err != nil {
					return AuditReport{}, err
				}
			}
		}
		snapshots = append(snapshots, CollectionSnapshot{
			Name:           name,
			DocumentCount:  count,
			Indexes:        indexes[name],
			Samples:        samples,
			DuplicateRisks: duplicateRisks,
		})
	}
	return AnalyzeAudit(database.Name(), catalog, snapshots, opts), nil
}

func AnalyzeAudit(database string, catalog []CollectionSpec, snapshots []CollectionSnapshot, opts AuditOptions) AuditReport {
	catalogByName := catalogMap(catalog)
	liveByName := map[string]CollectionSnapshot{}
	for _, snapshot := range snapshots {
		liveByName[snapshot.Name] = snapshot
	}

	var report AuditReport
	report.Database = database
	for _, snapshot := range snapshots {
		audit := CollectionAudit{
			Name:              snapshot.Name,
			DocumentCount:     snapshot.DocumentCount,
			Indexes:           append([]IndexInfo(nil), snapshot.Indexes...),
			SampledDocuments:  len(snapshot.Samples),
			FieldTypes:        fieldTypes(snapshot.Samples),
			DuplicateKeyRisks: append([]DuplicateKeyRisk(nil), snapshot.DuplicateRisks...),
			LargeDocuments:    largeDocuments(snapshot.Samples, opts.LargeDocumentBytes),
		}
		if spec, ok := catalogByName[snapshot.Name]; ok {
			audit.MissingRequiredFields = missingRequiredFields(spec, snapshot.Samples)
		} else {
			report.UnknownCollections = append(report.UnknownCollections, snapshot.Name)
		}
		for field, types := range audit.FieldTypes {
			if len(types) > 1 {
				report.Warnings = append(report.Warnings, AuditWarning{
					Collection: snapshot.Name,
					Field:      field,
					Message:    "mixed field types detected in sample",
				})
			}
		}
		if len(audit.LargeDocuments) > 0 {
			report.Warnings = append(report.Warnings, AuditWarning{
				Collection: snapshot.Name,
				Message:    "large documents detected in sample",
			})
		}
		for _, risk := range audit.DuplicateKeyRisks {
			report.Warnings = append(report.Warnings, AuditWarning{
				Collection: snapshot.Name,
				Message:    fmt.Sprintf("duplicate logical key detected: %s groups=%d", risk.KeyName, risk.DuplicateGroups),
			})
		}
		report.Collections = append(report.Collections, audit)
	}
	for _, spec := range catalog {
		if _, ok := liveByName[spec.Name]; !ok {
			report.MissingCollections = append(report.MissingCollections, spec.Name)
		}
	}
	sortAuditReport(&report)
	return report
}

func FormatAuditReport(w io.Writer, report AuditReport, format string) error {
	sortAuditReport(&report)
	if format == "json" {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	for _, collection := range report.Collections {
		if _, err := fmt.Fprintf(w, "collection=%s documents=%d sampled=%d indexes=%d\n", collection.Name, collection.DocumentCount, collection.SampledDocuments, len(collection.Indexes)); err != nil {
			return err
		}
		for _, risk := range collection.DuplicateKeyRisks {
			if _, err := fmt.Fprintf(w, "duplicate_key_risk collection=%s key=%s fields=%s groups=%d\n", collection.Name, risk.KeyName, strings.Join(risk.Fields, ","), risk.DuplicateGroups); err != nil {
				return err
			}
		}
	}
	for _, name := range report.UnknownCollections {
		if _, err := fmt.Fprintf(w, "unknown_collection=%s\n", name); err != nil {
			return err
		}
	}
	for _, name := range report.MissingCollections {
		if _, err := fmt.Fprintf(w, "missing_collection=%s\n", name); err != nil {
			return err
		}
	}
	for _, warning := range report.Warnings {
		if _, err := fmt.Fprintf(w, "warning collection=%s field=%s message=%q\n", warning.Collection, warning.Field, warning.Message); err != nil {
			return err
		}
	}
	return nil
}

func sampleCollection(ctx context.Context, collection *drivermongo.Collection, limit int) ([]SampleDocument, error) {
	if limit == 0 {
		return nil, nil
	}
	findOpts := options.Find().SetLimit(int64(limit))
	cursor, err := collection.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		return nil, MapError(fmt.Errorf("sample collection %s: %w", collection.Name(), err))
	}
	defer cursor.Close(ctx)
	var samples []SampleDocument
	for cursor.Next(ctx) {
		raw := cursor.Current
		samples = append(samples, sampleFromRaw(raw))
	}
	if err := cursor.Err(); err != nil {
		return nil, MapError(fmt.Errorf("iterate sample collection %s: %w", collection.Name(), err))
	}
	return samples, nil
}

func sampleFromRaw(raw bson.Raw) SampleDocument {
	fields := map[string]string{}
	elements, err := raw.Elements()
	if err != nil {
		return SampleDocument{SizeBytes: len(raw), Fields: fields}
	}
	ref := ""
	for _, element := range elements {
		key := element.Key()
		value := element.Value()
		if key == "_id" {
			ref = value.String()
		}
		fields[key] = value.Type.String()
	}
	return SampleDocument{Ref: ref, SizeBytes: len(raw), Fields: fields}
}

func auditDuplicateKeys(ctx context.Context, collection *drivermongo.Collection, keys []LogicalKeySpec) ([]DuplicateKeyRisk, error) {
	var risks []DuplicateKeyRisk
	for _, key := range keys {
		if !key.Unique || len(key.Fields) == 0 {
			continue
		}
		group := bson.D{}
		for _, field := range key.Fields {
			group = append(group, bson.E{Key: field, Value: "$" + field})
		}
		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: group}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}}
		matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "count", Value: bson.D{{Key: "$gt", Value: 1}}}}}}
		pipeline := drivermongo.Pipeline{
			groupStage,
			matchStage,
			{{Key: "$count", Value: "groups"}},
		}
		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, MapError(fmt.Errorf("audit duplicate key %s.%s: %w", collection.Name(), key.Name, err))
		}
		var rows []struct {
			Groups int64 `bson:"groups"`
		}
		if err := cursor.All(ctx, &rows); err != nil {
			return nil, MapError(fmt.Errorf("decode duplicate key audit %s.%s: %w", collection.Name(), key.Name, err))
		}
		if len(rows) > 0 && rows[0].Groups > 0 {
			risks = append(risks, DuplicateKeyRisk{
				KeyName:         key.Name,
				Fields:          append([]string(nil), key.Fields...),
				DuplicateGroups: rows[0].Groups,
			})
		}
	}
	return risks, nil
}

func catalogMap(catalog []CollectionSpec) map[string]CollectionSpec {
	result := make(map[string]CollectionSpec, len(catalog))
	for _, spec := range catalog {
		result[spec.Name] = spec
	}
	return result
}

func fieldTypes(samples []SampleDocument) map[string][]string {
	typesByField := map[string]map[string]struct{}{}
	for _, sample := range samples {
		for field, typ := range sample.Fields {
			if _, ok := typesByField[field]; !ok {
				typesByField[field] = map[string]struct{}{}
			}
			typesByField[field][typ] = struct{}{}
		}
	}
	result := make(map[string][]string, len(typesByField))
	for field, typeSet := range typesByField {
		for typ := range typeSet {
			result[field] = append(result[field], typ)
		}
		sort.Strings(result[field])
	}
	return result
}

func missingRequiredFields(spec CollectionSpec, samples []SampleDocument) []FieldIssue {
	counts := map[string]int{}
	for _, field := range spec.RequiredFields {
		if !field.Required {
			continue
		}
		for _, sample := range samples {
			if _, ok := sample.Fields[field.Name]; !ok {
				counts[field.Name]++
			}
		}
	}
	var issues []FieldIssue
	for field, count := range counts {
		if count > 0 {
			issues = append(issues, FieldIssue{Field: field, Count: count})
		}
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].Field < issues[j].Field })
	return issues
}

func largeDocuments(samples []SampleDocument, threshold int) []LargeDocumentIssue {
	if threshold <= 0 {
		return nil
	}
	var issues []LargeDocumentIssue
	for _, sample := range samples {
		if sample.SizeBytes > threshold {
			issues = append(issues, LargeDocumentIssue{DocumentRef: sample.Ref, Bytes: sample.SizeBytes})
		}
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].DocumentRef != issues[j].DocumentRef {
			return issues[i].DocumentRef < issues[j].DocumentRef
		}
		return issues[i].Bytes < issues[j].Bytes
	})
	return issues
}

func sortAuditReport(report *AuditReport) {
	sort.SliceStable(report.Collections, func(i, j int) bool {
		return report.Collections[i].Name < report.Collections[j].Name
	})
	sort.Strings(report.UnknownCollections)
	sort.Strings(report.MissingCollections)
	sort.SliceStable(report.Warnings, func(i, j int) bool {
		if report.Warnings[i].Collection != report.Warnings[j].Collection {
			return report.Warnings[i].Collection < report.Warnings[j].Collection
		}
		if report.Warnings[i].Field != report.Warnings[j].Field {
			return report.Warnings[i].Field < report.Warnings[j].Field
		}
		return report.Warnings[i].Message < report.Warnings[j].Message
	})
}
