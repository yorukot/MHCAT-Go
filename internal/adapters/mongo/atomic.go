package mongo

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrEmptyUpdate       = errors.New("mongo update has no operations")
	ErrInvalidField      = errors.New("mongo update field is invalid")
	ErrConflictingUpdate = errors.New("mongo update has conflicting field operations")
)

type UpdateBuilder struct {
	operations map[string]map[string]any
	fieldOps   map[string]string
	err        error
}

func NewUpdate() *UpdateBuilder {
	return &UpdateBuilder{
		operations: map[string]map[string]any{},
		fieldOps:   map[string]string{},
	}
}

func (b *UpdateBuilder) Inc(field string, value int64) *UpdateBuilder {
	return b.add("$inc", field, value)
}

func (b *UpdateBuilder) Set(field string, value any) *UpdateBuilder {
	return b.add("$set", field, value)
}

func (b *UpdateBuilder) SetOnInsert(field string, value any) *UpdateBuilder {
	return b.add("$setOnInsert", field, value)
}

func (b *UpdateBuilder) Unset(field string) *UpdateBuilder {
	return b.add("$unset", field, "")
}

func (b *UpdateBuilder) AddToSet(field string, value any) *UpdateBuilder {
	return b.add("$addToSet", field, value)
}

func (b *UpdateBuilder) Pull(field string, value any) *UpdateBuilder {
	return b.add("$pull", field, value)
}

func (b *UpdateBuilder) Push(field string, value any) *UpdateBuilder {
	return b.add("$push", field, value)
}

func (b *UpdateBuilder) Build() (bson.D, error) {
	if b.err != nil {
		return nil, b.err
	}
	if len(b.operations) == 0 {
		return nil, ErrEmptyUpdate
	}
	opNames := make([]string, 0, len(b.operations))
	for op := range b.operations {
		opNames = append(opNames, op)
	}
	sort.Strings(opNames)
	result := make(bson.D, 0, len(opNames))
	for _, op := range opNames {
		fields := b.operations[op]
		fieldNames := make([]string, 0, len(fields))
		for field := range fields {
			fieldNames = append(fieldNames, field)
		}
		sort.Strings(fieldNames)
		document := make(bson.D, 0, len(fieldNames))
		for _, field := range fieldNames {
			document = append(document, bson.E{Key: field, Value: fields[field]})
		}
		result = append(result, bson.E{Key: op, Value: document})
	}
	return result, nil
}

func (b *UpdateBuilder) add(operation string, field string, value any) *UpdateBuilder {
	if b.err != nil {
		return b
	}
	field = strings.TrimSpace(field)
	if field == "" || strings.HasPrefix(field, "$") {
		b.err = ErrInvalidField
		return b
	}
	if existing, ok := b.fieldOps[field]; ok && existing != operation {
		b.err = fmt.Errorf("%w: %s already used by %s", ErrConflictingUpdate, field, existing)
		return b
	}
	if _, ok := b.operations[operation]; !ok {
		b.operations[operation] = map[string]any{}
	}
	b.operations[operation][field] = value
	b.fieldOps[field] = operation
	return b
}
