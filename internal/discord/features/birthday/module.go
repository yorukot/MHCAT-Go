package birthday

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/birthday"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	configService  coreservice.ConfigService
	profileService coreservice.ProfileService
	cachedUsers    ports.DiscordCachedUserInfoProvider
	usage          ports.UsageTracker
	clock          ports.Clock
	pendingAdds    *birthdayAddStateStore
	color          func() int
}

func NewModule(repo ports.BirthdayConfigRepository, usage ports.UsageTracker) Module {
	return NewModuleWithClock(repo, usage, nil)
}

func NewModuleWithClock(repo ports.BirthdayConfigRepository, usage ports.UsageTracker, clock ports.Clock) Module {
	return NewModuleWithClockAndCachedUsers(repo, nil, usage, clock)
}

func NewModuleWithCachedUsers(repo ports.BirthdayConfigRepository, cachedUsers ports.DiscordCachedUserInfoProvider, usage ports.UsageTracker) Module {
	return NewModuleWithClockAndCachedUsers(repo, cachedUsers, usage, nil)
}

func NewModuleWithClockAndCachedUsers(repo ports.BirthdayConfigRepository, cachedUsers ports.DiscordCachedUserInfoProvider, usage ports.UsageTracker, clock ports.Clock) Module {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return Module{
		configService:  coreservice.NewConfigService(repo),
		profileService: coreservice.NewProfileService(repo),
		cachedUsers:    cachedUsers,
		usage:          usage,
		clock:          clock,
		pendingAdds:    newBirthdayAddStateStore(),
		color:          legacyRandomColor,
	}
}

func (m Module) Name() string {
	return "birthday-config"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if err := router.RegisterSlash(BirthdayCommandName, m.Handler()); err != nil {
		return err
	}
	if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "birthday", Action: "hour"}, m.HourSelectHandler()); err != nil {
		return err
	}
	return router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: customid.VersionV1, Feature: "birthday", Action: "minute"}, m.MinuteSelectHandler())
}

type pendingBirthdayAdd struct {
	OwnerUserID string
	Profile     domain.BirthdayProfile
	Hour        int
	HasHour     bool
	ExpiresAt   time.Time
}

type birthdayAddStateStore struct {
	mu      sync.Mutex
	next    uint64
	entries map[string]pendingBirthdayAdd
}

func newBirthdayAddStateStore() *birthdayAddStateStore {
	return &birthdayAddStateStore{entries: map[string]pendingBirthdayAdd{}}
}

func (s *birthdayAddStateStore) create(entry pendingBirthdayAdd) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	id := strconv.FormatUint(s.next, 36)
	s.entries[id] = entry
	return id
}

func (s *birthdayAddStateStore) get(id string, now time.Time) (pendingBirthdayAdd, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[id]
	if !ok {
		return pendingBirthdayAdd{}, false
	}
	if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
		delete(s.entries, id)
		return pendingBirthdayAdd{}, false
	}
	return entry, true
}

func (s *birthdayAddStateStore) setHour(id string, now time.Time, hour int) (pendingBirthdayAdd, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[id]
	if !ok {
		return pendingBirthdayAdd{}, false
	}
	if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
		delete(s.entries, id)
		return pendingBirthdayAdd{}, false
	}
	entry.Hour = hour
	entry.HasHour = true
	s.entries[id] = entry
	return entry, true
}

func (s *birthdayAddStateStore) delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, id)
}

func (m Module) now() time.Time {
	if m.clock == nil {
		return time.Now()
	}
	return m.clock.Now()
}

func (m Module) legacyColor() int {
	if m.color == nil {
		return legacyRandomColor()
	}
	return m.color()
}

func legacyRandomColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return birthdaySuccessColor
	}
	return int(value.Int64())
}
