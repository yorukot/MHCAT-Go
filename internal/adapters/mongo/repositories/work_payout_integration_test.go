package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestWorkPayoutMongoIntegrationCrashRetryAndMarkerOrdering(t *testing.T) {
	database := workPayoutIntegrationDatabase(t)
	repository, err := NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	workID := bson.NewObjectID()
	coinID := bson.NewObjectID()
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "_id", Value: coinID},
		{Key: "guild", Value: "guild-crash"},
		{Key: "member", Value: "user-crash"},
		{Key: "coin", Value: int64(100)},
		{Key: "today", Value: int64(1)},
	}); err != nil {
		t.Fatalf("insert coin: %v", err)
	}
	if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, workPayoutWorkDocument(workID, "guild-crash", "user-crash", "job-a", 100, 50)); err != nil {
		t.Fatalf("insert work user: %v", err)
	}

	var pending workUserPayoutDocument
	if err := database.Collection(WorkUserCollectionName).FindOne(ctx, bson.D{{Key: "_id", Value: workID}}).Decode(&pending); err != nil {
		t.Fatalf("decode pending work: %v", err)
	}
	firstIdentity, err := newWorkPayoutIdentity(pending)
	if err != nil {
		t.Fatalf("first identity: %v", err)
	}
	coinResult, err := database.Collection(CoinCollectionName).UpdateOne(
		ctx,
		workPayoutCoinFilter(coinID, pending.Guild, pending.User, firstIdentity),
		workPayoutCoinPipeline(pending.Guild, pending.User, 1, firstIdentity),
	)
	if err != nil {
		t.Fatalf("simulate coin commit before crash: %v", err)
	}
	if coinResult.ModifiedCount != 1 {
		t.Fatalf("initial coin update = %#v", coinResult)
	}
	assertWorkPayoutBalance(t, database, coinID, 150)
	assertWorkPayoutState(t, database, workID, "job-a")

	result, err := repository.RunWorkPayout(ctx, 200)
	if err != nil {
		t.Fatalf("retry payout: %v", err)
	}
	if result.ProcessedJobs != 1 || result.IdempotentReplays != 1 || result.CoinModified != 0 || result.StateModified != 1 {
		t.Fatalf("retry result = %#v", result)
	}
	assertWorkPayoutBalance(t, database, coinID, 150)
	assertWorkPayoutState(t, database, workID, LegacyIdleWorkState)

	if _, err := database.Collection(WorkUserCollectionName).UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: workID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "state", Value: "job-b"},
			{Key: "end_time", Value: int64(150)},
			{Key: "get_coin", Value: int64(25)},
		}}},
	); err != nil {
		t.Fatalf("start newer work job: %v", err)
	}
	result, err = repository.RunWorkPayout(ctx, 200)
	if err != nil {
		t.Fatalf("newer payout: %v", err)
	}
	if result.ProcessedJobs != 1 || result.IdempotentReplays != 0 || result.CoinModified != 1 {
		t.Fatalf("newer payout result = %#v", result)
	}
	assertWorkPayoutBalance(t, database, coinID, 175)

	staleResult, err := database.Collection(CoinCollectionName).UpdateOne(
		ctx,
		workPayoutCoinFilter(coinID, pending.Guild, pending.User, firstIdentity),
		workPayoutCoinPipeline(pending.Guild, pending.User, 1, firstIdentity),
	)
	if err != nil {
		t.Fatalf("stale payout attempt: %v", err)
	}
	if staleResult.MatchedCount != 0 || staleResult.ModifiedCount != 0 {
		t.Fatalf("stale payout must be rejected: %#v", staleResult)
	}
	assertWorkPayoutBalance(t, database, coinID, 175)

	var newer workUserPayoutDocument
	if err := database.Collection(WorkUserCollectionName).FindOne(ctx, bson.D{{Key: "_id", Value: workID}}).Decode(&newer); err != nil {
		t.Fatalf("decode newer work: %v", err)
	}
	newer.State = "job-b"
	newer.EndTime = rawValue(t, int64(150))
	newer.GetCoin = rawValue(t, int64(25))
	newerIdentity, err := newWorkPayoutIdentity(newer)
	if err != nil {
		t.Fatalf("newer identity: %v", err)
	}
	markerPath := WorkPayoutMarkerField + "." + newerIdentity.MarkerKey
	count, err := database.Collection(CoinCollectionName).CountDocuments(ctx, bson.D{
		{Key: "_id", Value: coinID},
		{Key: markerPath + ".token", Value: newerIdentity.Token},
		{Key: markerPath + ".end_time", Value: int64(150)},
	})
	if err != nil || count != 1 {
		t.Fatalf("retained marker count=%d err=%v", count, err)
	}
}

func TestWorkPayoutMongoIntegrationCreatesMissingCoinWithMarker(t *testing.T) {
	database := workPayoutIntegrationDatabase(t)
	repository, err := NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	workID := bson.NewObjectID()
	if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, workPayoutWorkDocument(workID, "guild-new", "user-new", "job", 100, "40")); err != nil {
		t.Fatalf("insert work user: %v", err)
	}
	result, err := repository.RunWorkPayout(ctx, 200)
	if err != nil {
		t.Fatalf("run payout: %v", err)
	}
	if result.ProcessedJobs != 1 || result.CoinUpserted != 1 || result.CoinModified != 0 {
		t.Fatalf("payout result = %#v", result)
	}
	expectedID, err := newWorkPayoutCoinID("guild-new", "user-new")
	if err != nil {
		t.Fatalf("expected coin id: %v", err)
	}
	assertWorkPayoutBalance(t, database, expectedID, 40)
	var coin struct {
		Today int64 `bson:"today"`
	}
	if err := database.Collection(CoinCollectionName).FindOne(ctx, bson.D{{Key: "_id", Value: expectedID}}).Decode(&coin); err != nil {
		t.Fatalf("decode new coin: %v", err)
	}
	if coin.Today != 1 {
		t.Fatalf("new coin today = %d", coin.Today)
	}
	assertWorkPayoutState(t, database, workID, LegacyIdleWorkState)
}

func TestWorkPayoutMongoIntegrationPreservesDecimalScalars(t *testing.T) {
	database := workPayoutIntegrationDatabase(t)
	repository, err := NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	workID := bson.NewObjectID()
	coinID := bson.NewObjectID()
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "_id", Value: coinID}, {Key: "guild", Value: "guild-decimal"},
		{Key: "member", Value: "user-decimal"}, {Key: "coin", Value: 1.25},
	}); err != nil {
		t.Fatalf("insert coin: %v", err)
	}
	if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, bson.D{
		{Key: "_id", Value: workID}, {Key: "guild", Value: "guild-decimal"},
		{Key: "user", Value: "user-decimal"}, {Key: "state", Value: "job"},
		{Key: "end_time", Value: 0.5}, {Key: "get_coin", Value: 2.5},
	}); err != nil {
		t.Fatalf("insert work user: %v", err)
	}
	result, err := repository.RunWorkPayout(ctx, 1)
	if err != nil || result.ProcessedJobs != 1 {
		t.Fatalf("run payout = %#v, err=%v", result, err)
	}
	var coin struct {
		Balance float64 `bson:"coin"`
	}
	if err := database.Collection(CoinCollectionName).FindOne(ctx, bson.D{{Key: "_id", Value: coinID}}).Decode(&coin); err != nil {
		t.Fatalf("decode coin: %v", err)
	}
	if coin.Balance != 3.75 {
		t.Fatalf("coin balance = %v", coin.Balance)
	}
	assertWorkPayoutState(t, database, workID, LegacyIdleWorkState)
}

func TestWorkPayoutMongoIntegrationConcurrentSameTokenCreditsOnce(t *testing.T) {
	database := workPayoutIntegrationDatabase(t)
	ctx := context.Background()
	workID := bson.NewObjectID()
	coinID := bson.NewObjectID()
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "_id", Value: coinID},
		{Key: "guild", Value: "guild-concurrent"},
		{Key: "member", Value: "user-concurrent"},
		{Key: "coin", Value: int64(100)},
		{Key: "today", Value: int64(1)},
	}); err != nil {
		t.Fatalf("insert coin: %v", err)
	}
	if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, workPayoutWorkDocument(workID, "guild-concurrent", "user-concurrent", "job", 100, 50)); err != nil {
		t.Fatalf("insert work user: %v", err)
	}
	var pending workUserPayoutDocument
	if err := database.Collection(WorkUserCollectionName).FindOne(ctx, bson.D{{Key: "_id", Value: workID}}).Decode(&pending); err != nil {
		t.Fatalf("decode pending work: %v", err)
	}
	identity, err := newWorkPayoutIdentity(pending)
	if err != nil {
		t.Fatalf("payout identity: %v", err)
	}

	const contenders = 16
	results := make(chan *drivermongo.UpdateResult, contenders)
	errorsFound := make(chan error, contenders)
	var wait sync.WaitGroup
	for range contenders {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := database.Collection(CoinCollectionName).UpdateOne(
				ctx,
				workPayoutCoinFilter(coinID, pending.Guild, pending.User, identity),
				workPayoutCoinPipeline(pending.Guild, pending.User, 1, identity),
			)
			if err != nil {
				errorsFound <- err
				return
			}
			results <- result
		}()
	}
	wait.Wait()
	close(results)
	close(errorsFound)
	for err := range errorsFound {
		t.Errorf("concurrent payout: %v", err)
	}
	var modified int64
	var matched int64
	for result := range results {
		modified += result.ModifiedCount
		matched += result.MatchedCount
	}
	if matched != contenders || modified != 1 {
		t.Fatalf("concurrent results matched=%d modified=%d", matched, modified)
	}
	assertWorkPayoutBalance(t, database, coinID, 150)
}

func TestWorkPayoutMongoIntegrationRejectsDuplicateCoinsBeforeCredit(t *testing.T) {
	database := workPayoutIntegrationDatabase(t)
	repository, err := NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	workID := bson.NewObjectID()
	if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, workPayoutWorkDocument(workID, "guild-duplicate", "user-duplicate", "job", 100, 50)); err != nil {
		t.Fatalf("insert work user: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-duplicate"}, {Key: "member", Value: "user-duplicate"}, {Key: "coin", Value: int64(10)}, {Key: "today", Value: int64(1)}},
		bson.D{{Key: "guild", Value: "guild-duplicate"}, {Key: "member", Value: "user-duplicate"}, {Key: "coin", Value: int64(20)}, {Key: "today", Value: int64(1)}},
	}); err != nil {
		t.Fatalf("insert duplicate coins: %v", err)
	}
	result, err := repository.RunWorkPayout(ctx, 200)
	if !errors.Is(err, domain.ErrWorkPayoutCoinConflict) {
		t.Fatalf("expected coin conflict, got result=%#v err=%v", result, err)
	}
	if result.CoinMatched != 0 || result.CoinModified != 0 || result.StateModified != 0 {
		t.Fatalf("conflict must not write: %#v", result)
	}
	assertWorkPayoutState(t, database, workID, "job")
	var total []struct {
		Coin int64 `bson:"coin"`
	}
	cursor, err := database.Collection(CoinCollectionName).Find(ctx, bson.D{{Key: "guild", Value: "guild-duplicate"}, {Key: "member", Value: "user-duplicate"}})
	if err != nil {
		t.Fatalf("find duplicate coins: %v", err)
	}
	if err := cursor.All(ctx, &total); err != nil {
		t.Fatalf("decode duplicate coins: %v", err)
	}
	if len(total) != 2 || total[0].Coin+total[1].Coin != 30 {
		t.Fatalf("duplicate balances changed: %#v", total)
	}
}

func TestWorkPayoutMongoIntegrationRejectsNonnumericCoinBeforeCredit(t *testing.T) {
	database := workPayoutIntegrationDatabase(t)
	repository, err := NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	workID := bson.NewObjectID()
	coinID := bson.NewObjectID()
	if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, workPayoutWorkDocument(workID, "guild-invalid-coin", "user-invalid-coin", "job", 100, 50)); err != nil {
		t.Fatalf("insert work user: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "_id", Value: coinID},
		{Key: "guild", Value: "guild-invalid-coin"},
		{Key: "member", Value: "user-invalid-coin"},
		{Key: "coin", Value: nil},
		{Key: "today", Value: int64(1)},
	}); err != nil {
		t.Fatalf("insert invalid coin: %v", err)
	}
	result, err := repository.RunWorkPayout(ctx, 200)
	if !errors.Is(err, domain.ErrWorkPayoutCoinConflict) {
		t.Fatalf("expected coin conflict, got result=%#v err=%v", result, err)
	}
	if result.CoinModified != 0 || result.StateModified != 0 {
		t.Fatalf("invalid coin must not write: %#v", result)
	}
	assertWorkPayoutState(t, database, workID, "job")
	count, err := database.Collection(CoinCollectionName).CountDocuments(ctx, bson.D{
		{Key: "_id", Value: coinID},
		{Key: WorkPayoutMarkerField, Value: bson.D{{Key: "$exists", Value: true}}},
	})
	if err != nil || count != 0 {
		t.Fatalf("invalid coin marker count=%d err=%v", count, err)
	}
}

func TestWorkPayoutMongoIntegrationPaysDuplicateWorkRowsIndependently(t *testing.T) {
	database := workPayoutIntegrationDatabase(t)
	repository, err := NewWorkPayoutRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	firstID := bson.NewObjectID()
	secondID := bson.NewObjectID()
	coinID := bson.NewObjectID()
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "_id", Value: coinID},
		{Key: "guild", Value: "guild-work-duplicates"},
		{Key: "member", Value: "user-work-duplicates"},
		{Key: "coin", Value: int64(10)},
		{Key: "today", Value: int64(1)},
	}); err != nil {
		t.Fatalf("insert coin: %v", err)
	}
	if _, err := database.Collection(WorkUserCollectionName).InsertMany(ctx, []any{
		workPayoutWorkDocument(firstID, "guild-work-duplicates", "user-work-duplicates", "job-a", 100, 5),
		workPayoutWorkDocument(secondID, "guild-work-duplicates", "user-work-duplicates", "job-b", 110, 7),
	}); err != nil {
		t.Fatalf("insert duplicate work rows: %v", err)
	}
	result, err := repository.RunWorkPayout(ctx, 200)
	if err != nil {
		t.Fatalf("run payout: %v", err)
	}
	if result.EligibleJobs != 2 || result.ProcessedJobs != 2 || result.CoinModified != 2 || result.StateModified != 2 {
		t.Fatalf("payout result = %#v", result)
	}
	assertWorkPayoutBalance(t, database, coinID, 22)
	assertWorkPayoutState(t, database, firstID, LegacyIdleWorkState)
	assertWorkPayoutState(t, database, secondID, LegacyIdleWorkState)

	for _, id := range []bson.ObjectID{firstID, secondID} {
		var document workUserPayoutDocument
		if err := database.Collection(WorkUserCollectionName).FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&document); err != nil {
			t.Fatalf("decode work row %s: %v", id.Hex(), err)
		}
		if id == firstID {
			document.State = "job-a"
		} else {
			document.State = "job-b"
		}
		identity, err := newWorkPayoutIdentity(document)
		if err != nil {
			t.Fatalf("identity for %s: %v", id.Hex(), err)
		}
		count, err := database.Collection(CoinCollectionName).CountDocuments(ctx, bson.D{
			{Key: "_id", Value: coinID},
			{Key: WorkPayoutMarkerField + "." + identity.MarkerKey + ".token", Value: identity.Token},
		})
		if err != nil || count != 1 {
			t.Fatalf("marker for %s count=%d err=%v", id.Hex(), count, err)
		}
	}
}

func workPayoutIntegrationDatabase(t *testing.T) *drivermongo.Database {
	t.Helper()
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	if uri == "" {
		t.Fatal("MHCAT_MONGODB_URI is required")
	}
	databaseName := fmt.Sprintf("mhcat_work_payout_test_%d", time.Now().UnixNano())
	client, err := mhcatmongo.NewClient(mhcatmongo.Options{
		URI:            uri,
		Database:       databaseName,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new Mongo client: %v", err)
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("connect Mongo: %v", err)
	}
	database, err := client.Database()
	if err != nil {
		t.Fatalf("get Mongo database: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.Drop(ctx); err != nil {
			t.Errorf("drop integration database: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("disconnect integration Mongo: %v", err)
		}
	})
	return database
}

func workPayoutWorkDocument(id bson.ObjectID, guildID string, userID string, state string, endTime int64, reward any) bson.D {
	return bson.D{
		{Key: "_id", Value: id},
		{Key: "guild", Value: guildID},
		{Key: "user", Value: userID},
		{Key: "state", Value: state},
		{Key: "end_time", Value: endTime},
		{Key: "energi", Value: int64(1)},
		{Key: "get_coin", Value: reward},
	}
}

func assertWorkPayoutBalance(t *testing.T, database *drivermongo.Database, coinID any, want int64) {
	t.Helper()
	var coin struct {
		Balance int64 `bson:"coin"`
	}
	if err := database.Collection(CoinCollectionName).FindOne(context.Background(), bson.D{{Key: "_id", Value: coinID}}).Decode(&coin); err != nil {
		t.Fatalf("decode coin: %v", err)
	}
	if coin.Balance != want {
		t.Fatalf("coin balance = %d, want %d", coin.Balance, want)
	}
}

func assertWorkPayoutState(t *testing.T, database *drivermongo.Database, workID any, want string) {
	t.Helper()
	var workUser struct {
		State string `bson:"state"`
	}
	if err := database.Collection(WorkUserCollectionName).FindOne(context.Background(), bson.D{{Key: "_id", Value: workID}}).Decode(&workUser); err != nil {
		t.Fatalf("decode work state: %v", err)
	}
	if workUser.State != want {
		t.Fatalf("work state = %q, want %q", workUser.State, want)
	}
}
