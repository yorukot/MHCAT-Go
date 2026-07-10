package repositories

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPollVoteFiltersPreserveChoiceWhitespace(t *testing.T) {
	const choice = " A "

	removeFilter := pollRemoveVoteFilter("guild-1", "message-1", "user-1", choice)
	removeMatch := documentValue(t, documentValue(t, removeFilter, "join_member"), "$elemMatch")
	if got := documentValue(t, removeMatch, "choise"); got != choice {
		t.Fatalf("remove choice = %q", got)
	}

	addFilter := pollAddVoteFilter("guild-1", "message-1", "user-1", choice)
	if got := documentValue(t, addFilter, "choose_data"); got != choice {
		t.Fatalf("stored choice match = %q", got)
	}
	notMatch := documentValue(t, documentValue(t, addFilter, "join_member"), "$not")
	addMatch := documentValue(t, notMatch, "$elemMatch")
	if got := documentValue(t, addMatch, "choise"); got != choice {
		t.Fatalf("duplicate choice match = %q", got)
	}
}

func TestPollVoteFiltersUseMongooseScalarGuards(t *testing.T) {
	wantTrue := pollMongooseTrueValues()
	activeFilter := pollActiveFilter("guild-1", "message-1")
	endCondition := documentValue(t, activeFilter, "end")
	if got := documentValue(t, endCondition, "$nin"); !reflect.DeepEqual(got, wantTrue) {
		t.Fatalf("active true values = %#v", got)
	}

	removeFilter := pollRemoveVoteFilter("guild-1", "message-1", "user-1", "A")
	changeCondition := documentValue(t, removeFilter, "can_change_choose")
	if got := documentValue(t, changeCondition, "$in"); !reflect.DeepEqual(got, wantTrue) {
		t.Fatalf("change true values = %#v", got)
	}

	addFilter := pollAddVoteFilter("guild-1", "message-1", "user-1", "A")
	expression := documentValue(t, addFilter, "$expr")
	lessThan := pollTestBSONArray(t, documentValue(t, expression, "$lt"))
	maxExpression := lessThan[1]
	letExpression := documentValue(t, maxExpression, "$let")
	variables := documentValue(t, letExpression, "vars")
	converted := documentValue(t, variables, "maxChoices")
	conversion := documentValue(t, converted, "$convert")
	if got := documentValue(t, conversion, "input"); got != "$many_choose" {
		t.Fatalf("max input = %#v", got)
	}
	if got := documentValue(t, conversion, "onError"); got != 1 {
		t.Fatalf("max fallback = %#v", got)
	}
	condition := pollTestBSONArray(t, documentValue(t, documentValue(t, letExpression, "in"), "$cond"))
	if condition[2] != 1 {
		t.Fatalf("nonpositive max fallback = %#v", condition[2])
	}
}

func pollTestBSONArray(t *testing.T, value any) bson.A {
	t.Helper()
	array, ok := value.(bson.A)
	if !ok {
		t.Fatalf("value type = %T, want bson.A", value)
	}
	return array
}
