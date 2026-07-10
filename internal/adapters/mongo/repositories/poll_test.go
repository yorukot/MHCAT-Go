package repositories

import "testing"

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
