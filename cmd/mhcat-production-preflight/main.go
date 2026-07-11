package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	mongoadapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var requiredAssets = []string{
	"asset/background.png",
	"asset/background_profile.png",
	"asset/blue_discord.png",
	"asset/coin_rank_background.png",
	"asset/mhcat_white.png",
	"asset/rank_background.png",
	"asset/verify_icon.png",
	"asset/yellow_discord.png",
	"fonts/Comic-Sans-MS-copy-5-.ttf",
	"fonts/Oswald-Regular.ttf",
	"fonts/TaipeiSansTCBeta-Regular.ttf",
	"fonts/language/Arabic.ttf",
	"fonts/language/Bengali.ttf",
	"fonts/language/HK.otf",
	"fonts/language/JP.otf",
	"fonts/language/NotoSans.ttf",
	"fonts/language/SC.otf",
	"fonts/language/TC.otf",
	"fonts/language/emoji.ttf",
}

type report struct {
	Environment              string   `json:"environment"`
	MongoDatabase            string   `json:"mongo_database"`
	MongoReplicaSet          string   `json:"mongo_replica_set,omitempty"`
	MongoWritablePrimary     bool     `json:"mongo_writable_primary"`
	DiscordConfiguredShards  int      `json:"discord_configured_shards"`
	DiscordRecommendedShards int      `json:"discord_recommended_shards"`
	AssetsChecked            int      `json:"assets_checked"`
	Errors                   []string `json:"errors,omitempty"`
}

type helloResult struct {
	SetName           string `bson:"setName"`
	IsWritablePrimary bool   `bson:"isWritablePrimary"`
}

type gatewayBotResponse struct {
	Shards int `json:"shards"`
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr, http.DefaultClient))
}

func run(ctx context.Context, args []string, lookup config.LookupFunc, stdout io.Writer, stderr io.Writer, httpClient *http.Client) int {
	flags := flag.NewFlagSet("mhcat-production-preflight", flag.ContinueOnError)
	flags.SetOutput(stderr)
	format := flags.String("format", "text", "output format: text or json")
	assetRoot := flags.String("asset-root", ".", "root containing asset and fonts directories")
	if err := flags.Parse(args); err != nil {
		return 1
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintln(stderr, "production preflight: format must be text or json")
		return 1
	}

	cfg, err := config.LoadWithLookup(lookup)
	if err != nil {
		fmt.Fprintf(stderr, "production preflight config: %v\n", err)
		return 1
	}
	result := report{Environment: cfg.Env, MongoDatabase: cfg.MongoDBDatabase, DiscordConfiguredShards: cfg.DiscordShardCount}
	if err := config.Validate(cfg); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}
	if cfg.Env != "production" {
		result.Errors = append(result.Errors, "MHCAT_ENV must be production")
	}

	missingAssets := checkAssets(*assetRoot)
	result.AssetsChecked = len(requiredAssets) - len(missingAssets)
	for _, path := range missingAssets {
		result.Errors = append(result.Errors, "missing UI asset: "+path)
	}

	if len(result.Errors) == 0 {
		mongoClient, newErr := mongoadapter.NewClient(mongoadapter.Options{
			URI: cfg.MongoDBURI, Database: cfg.MongoDBDatabase,
			ConnectTimeout: cfg.MongoConnectTimeout, PingTimeout: cfg.MongoPingTimeout,
		})
		if newErr != nil {
			result.Errors = append(result.Errors, newErr.Error())
		} else if connectErr := mongoClient.Connect(ctx); connectErr != nil {
			result.Errors = append(result.Errors, connectErr.Error())
		} else {
			database, databaseErr := mongoClient.Database()
			if databaseErr != nil {
				result.Errors = append(result.Errors, databaseErr.Error())
			} else {
				var hello helloResult
				if helloErr := database.RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&hello); helloErr != nil {
					result.Errors = append(result.Errors, "mongo hello: "+helloErr.Error())
				} else {
					result.MongoReplicaSet = hello.SetName
					result.MongoWritablePrimary = hello.IsWritablePrimary
					if !hello.IsWritablePrimary {
						result.Errors = append(result.Errors, "MongoDB is not a writable primary")
					}
					if cfg.FeatureEconomyGameEnabled && strings.TrimSpace(hello.SetName) == "" {
						result.Errors = append(result.Errors, "economy game requires a MongoDB replica set")
					}
				}
			}
			disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if disconnectErr := mongoClient.Disconnect(disconnectCtx); disconnectErr != nil {
				result.Errors = append(result.Errors, disconnectErr.Error())
			}
			cancel()
		}
	}

	if len(result.Errors) == 0 {
		recommended, gatewayErr := recommendedShardCount(ctx, httpClient, cfg.DiscordToken)
		if gatewayErr != nil {
			result.Errors = append(result.Errors, gatewayErr.Error())
		} else {
			result.DiscordRecommendedShards = recommended
			if recommended != cfg.DiscordShardCount {
				result.Errors = append(result.Errors, fmt.Sprintf("configured shard count %d does not match Discord recommendation %d", cfg.DiscordShardCount, recommended))
			}
		}
	}

	writeReport(stdout, result, *format)
	if len(result.Errors) > 0 {
		return 1
	}
	return 0
}

func checkAssets(root string) []string {
	missing := make([]string, 0)
	for _, path := range requiredAssets {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			missing = append(missing, path)
		}
	}
	return missing
}

func recommendedShardCount(ctx context.Context, client *http.Client, token string) (int, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/v10/gateway/bot", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("User-Agent", "MHCAT production preflight")
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Discord gateway metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Discord gateway metadata returned HTTP %d", resp.StatusCode)
	}
	var payload gatewayBotResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return 0, fmt.Errorf("decode Discord gateway metadata: %w", err)
	}
	if payload.Shards <= 0 {
		return 0, errors.New("Discord returned an invalid shard recommendation")
	}
	return payload.Shards, nil
}

func writeReport(writer io.Writer, result report, format string) {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
		return
	}
	fmt.Fprintf(writer, "environment=%s\nmongo_database=%s\nmongo_replica_set=%s\nmongo_writable_primary=%t\ndiscord_shards=%d/%d\nassets=%d/%d\n",
		result.Environment, result.MongoDatabase, result.MongoReplicaSet, result.MongoWritablePrimary,
		result.DiscordConfiguredShards, result.DiscordRecommendedShards, result.AssetsChecked, len(requiredAssets))
	for _, err := range result.Errors {
		fmt.Fprintln(writer, "error="+err)
	}
}
