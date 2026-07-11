package repositories

import (
	"math"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestWorkInterfaceCollectionNames(t *testing.T) {
	if WorkSetCollectionName != "work_sets" {
		t.Fatalf("work set collection = %s", WorkSetCollectionName)
	}
	if WorkSomethingCollectionName != "work_somethings" {
		t.Fatalf("work something collection = %s", WorkSomethingCollectionName)
	}
	if WorkUserCollectionName != "work_users" {
		t.Fatalf("work user collection = %s", WorkUserCollectionName)
	}
}

func TestNewWorkInterfaceRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewWorkInterfaceRepository(nil, nil, nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestWorkStartFilterRequiresIdleUnlessOverride(t *testing.T) {
	command := testWorkStartCommand()
	filter := workStartFilter(command)
	if !containsKey(filter, "state") {
		t.Fatalf("expected state filter for non-override start: %#v", filter)
	}
	command.Override = true
	filter = workStartFilter(command)
	if containsKey(filter, "state") {
		t.Fatalf("did not expect state filter for override start: %#v", filter)
	}
	if !containsKey(filter, "energi") {
		t.Fatalf("expected energy filter: %#v", filter)
	}
}

func TestValidateWorkStartCommand(t *testing.T) {
	if err := validateWorkStartCommand(testWorkStartCommand()); err != nil {
		t.Fatalf("valid command: %v", err)
	}
	invalid := testWorkStartCommand()
	invalid.EnergyCost = -1
	invalid.DurationSec = -1
	invalid.CoinReward = -1
	if err := validateWorkStartCommand(invalid); err != nil {
		t.Fatalf("legacy signed work values must be accepted: %v", err)
	}
}

func TestWorkStartUsesLegacyScalarArithmetic(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		want       float64
		wantEnergy bool
	}{
		{name: "decimal", text: "2.5", want: 2.5, wantEnergy: true},
		{name: "negative", text: "-2", want: -2, wantEnergy: true},
		{name: "null", text: "null", want: 0, wantEnergy: true},
		{name: "infinity", text: "Infinity", want: math.Inf(1), wantEnergy: true},
		{name: "malformed", text: "undefined", want: math.NaN(), wantEnergy: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := testWorkStartCommand()
			command.DurationText = test.text
			command.EnergyCostText = test.text
			command.CoinRewardText = test.text
			filter := workStartFilter(command)
			if containsKey(filter, "energi") != test.wantEnergy {
				t.Fatalf("energy filter = %#v", filter)
			}
			update := workStartUpdate(command)
			inc := lookupD(update, "$inc").(bson.D)
			set := lookupD(update, "$set").(bson.D)
			assertWorkFloat(t, lookupD(inc, "energi"), -test.want)
			assertWorkFloat(t, lookupD(set, "end_time"), 100+test.want)
			assertWorkFloat(t, lookupD(set, "get_coin"), test.want)
		})
	}
}

func assertWorkFloat(t *testing.T, got any, want float64) {
	t.Helper()
	value, ok := got.(float64)
	if !ok || !(value == want || math.IsNaN(value) && math.IsNaN(want)) {
		t.Fatalf("value = %#v, want %v", got, want)
	}
}

func TestValidateWorkAdminCommands(t *testing.T) {
	if err := validateWorkConfigCommand(domain.WorkConfigCommand{GuildID: "guild-1", DailyEnergy: 1, MaxEnergy: 10}); err != nil {
		t.Fatalf("valid config: %v", err)
	}
	if err := validateWorkConfigCommand(domain.WorkConfigCommand{GuildID: "guild-1", DailyEnergy: -1, MaxEnergy: 10}); err == nil {
		t.Fatal("expected invalid negative daily energy")
	}
	if err := validateWorkDeleteItemCommand(domain.WorkDeleteItemCommand{GuildID: "guild-1", Name: "礦坑"}); err != nil {
		t.Fatalf("valid delete: %v", err)
	}
	if err := validateWorkDeleteItemCommand(domain.WorkDeleteItemCommand{GuildID: "guild-1"}); err == nil {
		t.Fatal("expected invalid empty name")
	}
	if err := validateWorkEnergyGrantCommand("guild-1", "user-1", 1, 10); err != nil {
		t.Fatalf("valid grant: %v", err)
	}
	if err := validateWorkEnergyGrantCommand("guild-1", "user-1", -1, 10); err != nil {
		t.Fatalf("legacy signed grant must be valid: %v", err)
	}
}

func TestWorkEnergyGrantPipelineClampsEnergy(t *testing.T) {
	pipeline := workEnergyGrantPipeline(5, 20)
	if len(pipeline) != 1 {
		t.Fatalf("pipeline = %#v", pipeline)
	}
	stage := pipeline[0]
	if len(stage) != 1 || stage[0].Key != "$set" {
		t.Fatalf("unexpected stage = %#v", stage)
	}
	set, ok := stage[0].Value.(bson.D)
	if !ok || !containsKey(set, "energi") {
		t.Fatalf("expected energi set expression: %#v", stage[0].Value)
	}
}

func testWorkStartCommand() domain.WorkStartCommand {
	return domain.WorkStartCommand{
		GuildID:     "guild-1",
		UserID:      "user-1",
		WorkName:    "礦坑",
		DurationSec: 3600,
		EnergyCost:  3,
		CoinReward:  88,
		MaxEnergy:   20,
		NowUnix:     100,
	}
}

func containsKey(filter any, key string) bool {
	switch typed := filter.(type) {
	case bson.D:
		for _, element := range typed {
			if element.Key == key {
				return true
			}
		}
	}
	return false
}
