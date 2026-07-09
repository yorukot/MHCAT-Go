package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultMongoAuditSampleLimit    = 20
	DefaultMongoAuditLargeDocBytes  = 1048576
	DefaultMongoAuditTimeout        = 30 * time.Second
	DefaultMongoIndexDryRun         = true
	DefaultMongoIndexApply          = false
	DefaultMongoIndexAllowUnique    = false
	DefaultMongoIndexAllowTTL       = false
	DefaultMongoIndexTimeout        = 60 * time.Second
	DefaultMongoIntegrationTestsRun = false
)

var ErrInvalidMongoAdminConfig = errors.New("invalid mongo admin config")

type MongoAdminConfig struct {
	Env                      string
	LogLevel                 string
	LogFormat                string
	MongoDBURI               string
	MongoDBDatabase          string
	MongoConnectTimeout      time.Duration
	MongoPingTimeout         time.Duration
	AuditSampleLimit         int
	AuditLargeDocumentBytes  int
	AuditTimeout             time.Duration
	IndexDryRun              bool
	IndexApply               bool
	IndexAllowUnique         bool
	IndexAllowTTL            bool
	IndexTimeout             time.Duration
	RunMongoIntegrationTests bool
	AliasWarnings            []AliasWarning
}

func LoadMongoAdmin() (MongoAdminConfig, error) {
	return LoadMongoAdminWithLookup(os.LookupEnv)
}

func LoadMongoAdminWithLookup(lookup LookupFunc) (MongoAdminConfig, error) {
	cfg, err := LoadMongoAdminRawWithLookup(lookup)
	if err != nil {
		return MongoAdminConfig{}, err
	}
	if err := ValidateMongoAdmin(cfg); err != nil {
		return MongoAdminConfig{}, err
	}
	return cfg, nil
}

func LoadMongoAdminRawWithLookup(lookup LookupFunc) (MongoAdminConfig, error) {
	cfg := MongoAdminConfig{
		Env:                      getString(lookup, "MHCAT_ENV", DefaultEnv),
		LogLevel:                 getString(lookup, "MHCAT_LOG_LEVEL", DefaultLogLevel),
		LogFormat:                getString(lookup, "MHCAT_LOG_FORMAT", DefaultLogFormat),
		MongoConnectTimeout:      DefaultMongoConnectTimeout,
		MongoPingTimeout:         DefaultMongoPingTimeout,
		AuditSampleLimit:         DefaultMongoAuditSampleLimit,
		AuditLargeDocumentBytes:  DefaultMongoAuditLargeDocBytes,
		AuditTimeout:             DefaultMongoAuditTimeout,
		IndexDryRun:              DefaultMongoIndexDryRun,
		IndexApply:               DefaultMongoIndexApply,
		IndexAllowUnique:         DefaultMongoIndexAllowUnique,
		IndexAllowTTL:            DefaultMongoIndexAllowTTL,
		IndexTimeout:             DefaultMongoIndexTimeout,
		RunMongoIntegrationTests: DefaultMongoIntegrationTestsRun,
	}
	cfg.MongoDBURI = getAliasedMongoAdminString(lookup, "MHCAT_MONGODB_URI", "MONGOOSE_CONNECTION_STRING", &cfg)
	cfg.MongoDBDatabase = getString(lookup, "MHCAT_MONGODB_DATABASE", "")

	var err error
	if cfg.MongoConnectTimeout, err = getDuration(lookup, "MHCAT_MONGO_CONNECT_TIMEOUT", DefaultMongoConnectTimeout); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.MongoPingTimeout, err = getDuration(lookup, "MHCAT_MONGO_PING_TIMEOUT", DefaultMongoPingTimeout); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.AuditSampleLimit, err = getInt(lookup, "MHCAT_MONGO_AUDIT_SAMPLE_LIMIT", DefaultMongoAuditSampleLimit); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.AuditLargeDocumentBytes, err = getInt(lookup, "MHCAT_MONGO_AUDIT_LARGE_DOC_BYTES", DefaultMongoAuditLargeDocBytes); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.AuditTimeout, err = getDuration(lookup, "MHCAT_MONGO_AUDIT_TIMEOUT", DefaultMongoAuditTimeout); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.IndexDryRun, err = getBool(lookup, "MHCAT_MONGO_INDEX_DRY_RUN", DefaultMongoIndexDryRun); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.IndexApply, err = getBool(lookup, "MHCAT_MONGO_INDEX_APPLY", DefaultMongoIndexApply); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.IndexAllowUnique, err = getBool(lookup, "MHCAT_MONGO_INDEX_ALLOW_UNIQUE", DefaultMongoIndexAllowUnique); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.IndexAllowTTL, err = getBool(lookup, "MHCAT_MONGO_INDEX_ALLOW_TTL", DefaultMongoIndexAllowTTL); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.IndexTimeout, err = getDuration(lookup, "MHCAT_MONGO_INDEX_TIMEOUT", DefaultMongoIndexTimeout); err != nil {
		return MongoAdminConfig{}, err
	}
	if cfg.RunMongoIntegrationTests, err = getBool(lookup, "MHCAT_RUN_MONGO_INTEGRATION_TESTS", DefaultMongoIntegrationTestsRun); err != nil {
		return MongoAdminConfig{}, err
	}
	return cfg, nil
}

func ValidateMongoAdmin(cfg MongoAdminConfig) error {
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("%w: MHCAT_LOG_LEVEL must be debug, info, warn, or error", ErrInvalidMongoAdminConfig)
	}
	switch cfg.LogFormat {
	case "text", "json":
	default:
		return fmt.Errorf("%w: MHCAT_LOG_FORMAT must be text or json", ErrInvalidMongoAdminConfig)
	}
	if strings.TrimSpace(cfg.MongoDBURI) == "" {
		return fmt.Errorf("%w: missing required env: MHCAT_MONGODB_URI", ErrInvalidMongoAdminConfig)
	}
	if strings.TrimSpace(cfg.MongoDBDatabase) == "" {
		return fmt.Errorf("%w: missing required env: MHCAT_MONGODB_DATABASE", ErrInvalidMongoAdminConfig)
	}
	if cfg.MongoConnectTimeout <= 0 {
		return fmt.Errorf("%w: mongo connect timeout must be positive", ErrInvalidMongoAdminConfig)
	}
	if cfg.MongoPingTimeout <= 0 {
		return fmt.Errorf("%w: mongo ping timeout must be positive", ErrInvalidMongoAdminConfig)
	}
	if cfg.AuditSampleLimit < 0 {
		return fmt.Errorf("%w: audit sample limit must be non-negative", ErrInvalidMongoAdminConfig)
	}
	if cfg.AuditLargeDocumentBytes <= 0 {
		return fmt.Errorf("%w: audit large document bytes must be positive", ErrInvalidMongoAdminConfig)
	}
	if cfg.AuditTimeout <= 0 {
		return fmt.Errorf("%w: audit timeout must be positive", ErrInvalidMongoAdminConfig)
	}
	if cfg.IndexTimeout <= 0 {
		return fmt.Errorf("%w: index timeout must be positive", ErrInvalidMongoAdminConfig)
	}
	if cfg.IndexApply && cfg.IndexDryRun {
		return fmt.Errorf("%w: index apply and dry-run cannot both be true", ErrInvalidMongoAdminConfig)
	}
	if cfg.IndexAllowUnique && !cfg.IndexApply {
		return fmt.Errorf("%w: allow-unique requires apply mode", ErrInvalidMongoAdminConfig)
	}
	if cfg.IndexAllowTTL && !cfg.IndexApply {
		return fmt.Errorf("%w: allow-ttl requires apply mode", ErrInvalidMongoAdminConfig)
	}
	return nil
}

func getAliasedMongoAdminString(lookup LookupFunc, primary, alias string, cfg *MongoAdminConfig) string {
	primaryValue, primaryOK := lookup(primary)
	aliasValue, aliasOK := lookup(alias)
	primaryValue = strings.TrimSpace(primaryValue)
	aliasValue = strings.TrimSpace(aliasValue)
	if primaryOK {
		if aliasOK && primaryValue != aliasValue {
			cfg.AliasWarnings = append(cfg.AliasWarnings, AliasWarning{
				Primary:      primary,
				Alias:        alias,
				PrimaryValue: primaryValue,
				AliasValue:   aliasValue,
			})
		}
		return primaryValue
	}
	if aliasOK {
		return aliasValue
	}
	return ""
}

func getInt(lookup LookupFunc, key string, fallback int) (int, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("parse %s as int: %w", key, err)
	}
	return parsed, nil
}
