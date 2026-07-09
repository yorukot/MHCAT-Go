package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestWorkDocumentsConvertLegacyFields(t *testing.T) {
	cfg := WorkConfigDocument{
		Guild:      "guild-1",
		GetEnergy:  rawInt32(t, 5),
		MaxEnergy:  rawDouble(t, 12.9),
		CaptchaRaw: rawBool(t, true),
	}.ToDomain()
	if cfg.DailyEnergy != 5 || cfg.MaxEnergy != 12 || !cfg.Captcha {
		t.Fatalf("config = %#v", cfg)
	}
	item := WorkItemDocument{
		Guild:  "guild-1",
		Name:   "礦坑",
		Time:   rawString(t, "3600"),
		Energy: rawInt64(t, 3),
		Coin:   rawDouble(t, 88),
		Role:   "role-1",
	}.ToDomain()
	if item.DurationSec != 3600 || item.EnergyCost != 3 || item.CoinReward != 88 || item.RoleID != "role-1" {
		t.Fatalf("item = %#v", item)
	}
	user := WorkUserDocument{Guild: "guild-1", User: "user-1", State: "", EndTime: rawInt64(t, 10), Energi: rawString(t, "7"), GetCoin: rawInt64(t, 99)}.ToDomain()
	if user.State != domain.WorkIdleState || user.Energy != 7 || user.GetCoin != 99 || !user.Initialized {
		t.Fatalf("user = %#v", user)
	}
}

func rawInt32(t *testing.T, value int32) bson.RawValue {
	t.Helper()
	data, err := bson.Marshal(bson.D{{Key: "v", Value: value}})
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	return raw.Lookup("v")
}

func rawInt64(t *testing.T, value int64) bson.RawValue {
	t.Helper()
	data, err := bson.Marshal(bson.D{{Key: "v", Value: value}})
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	return raw.Lookup("v")
}

func rawDouble(t *testing.T, value float64) bson.RawValue {
	t.Helper()
	data, err := bson.Marshal(bson.D{{Key: "v", Value: value}})
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	return raw.Lookup("v")
}

func rawString(t *testing.T, value string) bson.RawValue {
	t.Helper()
	data, err := bson.Marshal(bson.D{{Key: "v", Value: value}})
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	return raw.Lookup("v")
}

func rawBool(t *testing.T, value bool) bson.RawValue {
	t.Helper()
	data, err := bson.Marshal(bson.D{{Key: "v", Value: value}})
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	return raw.Lookup("v")
}
