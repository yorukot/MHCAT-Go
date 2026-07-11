package repositories

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestLotteryMongoIntegrationAtomicLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewLotteryRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new lottery repository: %v", err)
	}
	ctx := context.Background()
	collection := database.Collection(LotteryCollectionName)
	if _, err := collection.InsertOne(ctx, lotteryFixture("guild-1", "active", "Prize", int64(2), false)); err != nil {
		t.Fatalf("seed active lottery: %v", err)
	}

	joined, err := repository.JoinLottery(ctx, lotteryJoinRequest("active", "user-1"))
	if err != nil || len(joined.Participants) != 1 || joined.Participants[0].UserID != "user-1" {
		t.Fatalf("joined lottery = %#v err=%v", joined, err)
	}
	if _, err := repository.JoinLottery(ctx, lotteryJoinRequest("active", "user-1")); !errors.Is(err, ports.ErrLotteryAlreadyJoined) {
		t.Fatalf("duplicate lottery join error = %v", err)
	}
	if _, err := repository.JoinLottery(ctx, lotteryJoinRequest("active", "user-2")); err != nil {
		t.Fatalf("fill lottery: %v", err)
	}
	if _, err := repository.JoinLottery(ctx, lotteryJoinRequest("active", "user-3")); !errors.Is(err, ports.ErrLotteryFull) {
		t.Fatalf("full lottery join error = %v", err)
	}
	ended, err := repository.EndLottery(ctx, " guild-1 ", " active ")
	if err != nil || !ended.Ended || len(ended.Participants) != 2 {
		t.Fatalf("ended lottery = %#v err=%v", ended, err)
	}
	if _, err := repository.JoinLottery(ctx, lotteryJoinRequest("active", "user-4")); !errors.Is(err, ports.ErrLotteryFull) {
		t.Fatalf("full ended lottery join error = %v", err)
	}

	if _, err := collection.InsertMany(ctx, []any{
		lotteryFixture("guild-1", "ended-string", "String", int64(0), "true"),
		lotteryFixture("guild-1", "ended-number", "Number", int64(0), int64(1)),
		lotteryFixture("guild-1", "expired", "Expired", int64(0), false, bson.E{Key: "date", Value: "1"}),
		lotteryFixture("guild-1", "duplicate", "ended row", int64(0), true),
		lotteryFixture("guild-1", "duplicate", "active row", int64(0), false),
	}); err != nil {
		t.Fatalf("seed guarded lotteries: %v", err)
	}
	for _, id := range []string{"ended-string", "ended-number", "expired"} {
		if _, err := repository.JoinLottery(ctx, lotteryJoinRequest(id, "guard-user")); !errors.Is(err, ports.ErrLotteryEnded) {
			t.Fatalf("%s join error = %v", id, err)
		}
	}
	duplicate, err := repository.JoinLottery(ctx, lotteryJoinRequest("duplicate", "duplicate-user"))
	if err != nil || duplicate.Gift != "active row" || duplicate.Ended || !duplicate.HasParticipant("duplicate-user") {
		t.Fatalf("updated duplicate lottery = %#v err=%v", duplicate, err)
	}
	if _, err := repository.GetLottery(ctx, "guild-1", "missing"); !errors.Is(err, ports.ErrLotteryNotFound) {
		t.Fatalf("missing lottery error = %v", err)
	}
	if _, err := repository.EndLottery(ctx, "guild-1", "missing"); !errors.Is(err, ports.ErrLotteryNotFound) {
		t.Fatalf("end missing lottery error = %v", err)
	}
}

func TestLotteryMongoIntegrationConcurrentCapacity(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewLotteryRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new lottery repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(LotteryCollectionName).InsertOne(ctx, lotteryFixture("guild-1", "race", "Prize", int64(1), false)); err != nil {
		t.Fatalf("seed race lottery: %v", err)
	}

	const workers = 12
	var wait sync.WaitGroup
	errorsCh := make(chan error, workers)
	for index := range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := repository.JoinLottery(ctx, lotteryJoinRequest("race", fmt.Sprintf("user-%d", index)))
			errorsCh <- err
		}()
	}
	wait.Wait()
	close(errorsCh)
	successes := 0
	for err := range errorsCh {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ports.ErrLotteryFull):
		default:
			t.Fatalf("concurrent lottery join error = %v", err)
		}
	}
	lottery, err := repository.GetLottery(ctx, "guild-1", "race")
	if err != nil || successes != 1 || len(lottery.Participants) != 1 {
		t.Fatalf("concurrent lottery successes=%d lottery=%#v err=%v", successes, lottery, err)
	}
}

func TestTicketMongoIntegrationCreationRollbackAndDelete(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewTicketConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new ticket repository: %v", err)
	}
	ctx := context.Background()
	config := domain.TicketConfig{
		GuildID: " guild-1 ", CategoryID: " category-1 ", AdminRoleID: " admin-1 ", EveryoneRoleID: " everyone-1 ",
	}
	first, err := repository.CreateTicketConfig(ctx, config)
	if err != nil || first.GuildID != "guild-1" || first.ID == "" {
		t.Fatalf("ticket creation = %#v err=%v", first, err)
	}
	loaded, err := repository.GetTicketConfig(ctx, " guild-1 ")
	if err != nil || loaded.CategoryID != "category-1" || loaded.AdminRoleID != "admin-1" || loaded.EveryoneRoleID != "everyone-1" {
		t.Fatalf("ticket config = %#v err=%v", loaded, err)
	}
	if _, err := repository.CreateTicketConfig(ctx, domain.TicketConfig{
		GuildID: "guild-1", CategoryID: "other", AdminRoleID: "other", EveryoneRoleID: "other",
	}); !errors.Is(err, ports.ErrTicketConfigExists) {
		t.Fatalf("duplicate ticket config error = %v", err)
	}
	if err := repository.RollbackTicketConfigCreation(ctx, first); err != nil {
		t.Fatalf("rollback first ticket config: %v", err)
	}
	replacement, err := repository.CreateTicketConfig(ctx, domain.TicketConfig{
		GuildID: "guild-1", CategoryID: "category-2", AdminRoleID: "admin-2", EveryoneRoleID: "everyone-2",
	})
	if err != nil {
		t.Fatalf("create replacement ticket config: %v", err)
	}
	if err := repository.RollbackTicketConfigCreation(ctx, first); err != nil {
		t.Fatalf("stale ticket rollback: %v", err)
	}
	loaded, err = repository.GetTicketConfig(ctx, "guild-1")
	if err != nil || loaded.CategoryID != "category-2" {
		t.Fatalf("replacement ticket config = %#v err=%v", loaded, err)
	}
	if _, err := database.Collection(TicketConfigCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "ticket_channel", Value: "duplicate"},
	}); err != nil {
		t.Fatalf("seed duplicate ticket config: %v", err)
	}
	if err := repository.DeleteTicketConfig(ctx, " guild-1 "); err != nil {
		t.Fatalf("delete ticket configs: %v", err)
	}
	if _, err := repository.GetTicketConfig(ctx, "guild-1"); !errors.Is(err, ports.ErrTicketConfigNotFound) {
		t.Fatalf("deleted ticket config error = %v", err)
	}
	if err := repository.DeleteTicketConfig(ctx, "guild-1"); !errors.Is(err, ports.ErrTicketConfigNotFound) {
		t.Fatalf("second ticket delete error = %v", err)
	}
	if err := repository.RollbackTicketConfigCreation(ctx, ports.TicketConfigCreation{GuildID: replacement.GuildID, ID: "invalid"}); !errors.Is(err, domain.ErrInvalidTicketConfig) {
		t.Fatalf("invalid ticket rollback error = %v", err)
	}
}

func TestLoggingMongoIntegrationUpsertDuplicateAlignmentAndCache(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewLoggingConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new logging repository: %v", err)
	}
	ctx := context.Background()
	config := domain.LoggingConfig{GuildID: " guild-1 ", ChannelID: " channel-1 ", MessageUpdate: true}
	if err := repository.SaveLoggingConfig(ctx, config); err != nil {
		t.Fatalf("create logging config: %v", err)
	}
	loaded, err := repository.GetLoggingConfig(ctx, " guild-1 ")
	if err != nil || loaded.GuildID != "guild-1" || loaded.ChannelID != "channel-1" || !loaded.MessageUpdate {
		t.Fatalf("cached logging config = %#v err=%v", loaded, err)
	}
	if _, err := database.Collection(LoggingConfigCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "channel_id", Value: "old"}, {Key: "dashboard", Value: "preserve"},
	}); err != nil {
		t.Fatalf("seed duplicate logging config: %v", err)
	}
	config.ChannelID = " channel-2 "
	config.MessageUpdate = false
	config.MessageDelete = true
	config.ChannelUpdate = true
	config.MemberVoiceUpdate = true
	if err := repository.SaveLoggingConfig(ctx, config); err != nil {
		t.Fatalf("align logging configs: %v", err)
	}
	matched, err := database.Collection(LoggingConfigCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "channel_id", Value: "channel-2"}, {Key: "message_delete", Value: true},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned logging configs = %d err=%v", matched, err)
	}
	loaded, err = repository.GetLoggingConfig(ctx, "guild-1")
	if err != nil || loaded.ChannelID != "channel-2" || !loaded.MessageDelete || !loaded.ChannelUpdate || !loaded.MemberVoiceUpdate {
		t.Fatalf("updated cached logging config = %#v err=%v", loaded, err)
	}
	fresh, err := NewLoggingConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new fresh logging repository: %v", err)
	}
	loaded, err = fresh.GetLoggingConfig(ctx, "guild-1")
	if err != nil || loaded.ChannelID != "channel-2" {
		t.Fatalf("stored logging config = %#v err=%v", loaded, err)
	}
	if err := repository.SaveLoggingConfig(ctx, domain.LoggingConfig{}); !errors.Is(err, domain.ErrInvalidLoggingConfig) {
		t.Fatalf("invalid logging config error = %v", err)
	}
}

func lotteryFixture(guildID string, id string, gift string, maxParticipants int64, ended any, overrides ...bson.E) bson.D {
	document := bson.D{
		{Key: "guild", Value: guildID},
		{Key: "id", Value: id},
		{Key: "date", Value: "4102444800"},
		{Key: "gift", Value: gift},
		{Key: "howmanywinner", Value: "1"},
		{Key: "member", Value: bson.A{}},
		{Key: "end", Value: ended},
		{Key: "message_channel", Value: "channel-1"},
		{Key: "maxNumber", Value: maxParticipants},
	}
	for _, override := range overrides {
		for index := range document {
			if document[index].Key == override.Key {
				document[index] = override
				break
			}
		}
	}
	return document
}

func lotteryJoinRequest(id string, userID string) domain.LotteryJoinRequest {
	return domain.LotteryJoinRequest{
		GuildID: " guild-1 ", ID: " " + id + " ", UserID: " " + userID + " ",
		JoinedAtMillis: 1_700_000_000_000, NowUnix: 1_700_000_000,
	}
}
