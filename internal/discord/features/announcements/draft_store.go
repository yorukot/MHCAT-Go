package announcements

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var ErrAnnouncementDraftNotFound = errors.New("announcement draft not found")
var ErrAnnouncementDraftUnauthorized = errors.New("announcement draft unauthorized")

const (
	defaultDraftLimit = 512
	defaultDraftTTL   = 6 * time.Second
)

type AnnouncementDraft struct {
	GuildID   string
	UserID    string
	UserTag   string
	AvatarURL string
	Tag       string
	Color     int
	Title     string
	Content   string
	ExpiresAt time.Time
}

type DraftStore struct {
	mu     sync.Mutex
	now    func() time.Time
	limit  int
	ttl    time.Duration
	order  []string
	drafts map[string]AnnouncementDraft
}

func NewDraftStore() *DraftStore {
	return &DraftStore{
		now:    time.Now,
		limit:  defaultDraftLimit,
		ttl:    defaultDraftTTL,
		drafts: map[string]AnnouncementDraft{},
	}
}

func (s *DraftStore) Put(draft AnnouncementDraft) (string, error) {
	if s == nil {
		return "", ErrAnnouncementDraftNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure()
	id, err := randomDraftID()
	if err != nil {
		return "", err
	}
	if draft.ExpiresAt.IsZero() {
		draft.ExpiresAt = s.now().Add(s.ttl)
	}
	if len(s.drafts) >= s.limit && len(s.order) > 0 {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.drafts, oldest)
	}
	s.drafts[id] = draft
	s.order = append(s.order, id)
	return id, nil
}

func (s *DraftStore) Take(id string) (AnnouncementDraft, error) {
	return s.take(id, "", "", false)
}

func (s *DraftStore) TakeForActor(id string, guildID string, userID string) (AnnouncementDraft, error) {
	return s.take(id, guildID, userID, true)
}

func (s *DraftStore) take(id string, guildID string, userID string, authorize bool) (AnnouncementDraft, error) {
	if s == nil || id == "" {
		return AnnouncementDraft{}, ErrAnnouncementDraftNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure()
	draft, ok := s.drafts[id]
	if !ok {
		return AnnouncementDraft{}, ErrAnnouncementDraftNotFound
	}
	if !draft.ExpiresAt.IsZero() && !s.now().Before(draft.ExpiresAt) {
		delete(s.drafts, id)
		s.removeOrder(id)
		return AnnouncementDraft{}, ErrAnnouncementDraftNotFound
	}
	if authorize && ((draft.GuildID != "" && draft.GuildID != guildID) || (draft.UserID != "" && draft.UserID != userID)) {
		return AnnouncementDraft{}, ErrAnnouncementDraftUnauthorized
	}
	delete(s.drafts, id)
	s.removeOrder(id)
	return draft, nil
}

func (s *DraftStore) Delete(id string) {
	if s == nil || id == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure()
	delete(s.drafts, id)
	s.removeOrder(id)
}

func (s *DraftStore) ensure() {
	if s.drafts == nil {
		s.drafts = map[string]AnnouncementDraft{}
	}
	if s.now == nil {
		s.now = time.Now
	}
	if s.limit <= 0 {
		s.limit = defaultDraftLimit
	}
	if s.ttl <= 0 {
		s.ttl = defaultDraftTTL
	}
}

func (s *DraftStore) removeOrder(id string) {
	for index, current := range s.order {
		if current == id {
			s.order = append(s.order[:index], s.order[index+1:]...)
			return
		}
	}
}

func randomDraftID() (string, error) {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
