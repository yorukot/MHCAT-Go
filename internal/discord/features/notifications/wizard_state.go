package notifications

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

const autoNotificationWizardTTL = 5 * time.Minute

var errAutoNotificationWizardStateID = errors.New("generate auto-notification wizard state id")

type pendingAutoNotificationWizard struct {
	OwnerUserID      string
	GuildID          string
	ScheduleID       string
	PreviewChannelID string
	Message          domain.AutoNotificationMessage
	Week             string
	Hours            string
	ExpiresAt        time.Time
}

type autoNotificationWizardStateStore struct {
	mu      sync.Mutex
	entries map[string]pendingAutoNotificationWizard
}

func newAutoNotificationWizardStateStore() *autoNotificationWizardStateStore {
	return &autoNotificationWizardStateStore{entries: map[string]pendingAutoNotificationWizard{}}
}

func (s *autoNotificationWizardStateStore) create(now time.Time, entry pendingAutoNotificationWizard) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry.OwnerUserID = strings.TrimSpace(entry.OwnerUserID)
	entry.GuildID = strings.TrimSpace(entry.GuildID)
	entry.ScheduleID = strings.TrimSpace(entry.ScheduleID)
	entry.PreviewChannelID = strings.TrimSpace(entry.PreviewChannelID)
	for id, candidate := range s.entries {
		if !candidate.ExpiresAt.IsZero() && !now.Before(candidate.ExpiresAt) {
			delete(s.entries, id)
		}
	}
	for range 4 {
		id, err := randomAutoNotificationWizardStateID()
		if err != nil {
			return "", err
		}
		if _, exists := s.entries[id]; exists {
			continue
		}
		s.entries[id] = entry
		return id, nil
	}
	return "", errAutoNotificationWizardStateID
}

func randomAutoNotificationWizardStateID() (string, error) {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func (s *autoNotificationWizardStateStore) setWeek(id string, ownerUserID string, now time.Time, week string) (pendingAutoNotificationWizard, bool) {
	return s.update(id, ownerUserID, now, func(entry *pendingAutoNotificationWizard) {
		entry.Week = week
	})
}

func (s *autoNotificationWizardStateStore) setHours(id string, ownerUserID string, now time.Time, hours string) (pendingAutoNotificationWizard, bool) {
	return s.update(id, ownerUserID, now, func(entry *pendingAutoNotificationWizard) {
		entry.Hours = hours
	})
}

func (s *autoNotificationWizardStateStore) ready(id string, ownerUserID string, now time.Time) (pendingAutoNotificationWizard, bool) {
	entry, ok := s.update(id, ownerUserID, now, func(*pendingAutoNotificationWizard) {})
	return entry, ok && entry.Week != "" && entry.Hours != ""
}

func (s *autoNotificationWizardStateStore) update(id string, ownerUserID string, now time.Time, mutate func(*pendingAutoNotificationWizard)) (pendingAutoNotificationWizard, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id = strings.TrimSpace(id)
	ownerUserID = strings.TrimSpace(ownerUserID)
	entry, ok := s.entries[id]
	if !ok || entry.OwnerUserID != ownerUserID {
		return pendingAutoNotificationWizard{}, false
	}
	if !entry.ExpiresAt.IsZero() && !now.Before(entry.ExpiresAt) {
		delete(s.entries, id)
		return pendingAutoNotificationWizard{}, false
	}
	mutate(&entry)
	s.entries[id] = entry
	return entry, true
}

func (s *autoNotificationWizardStateStore) delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, id)
}
