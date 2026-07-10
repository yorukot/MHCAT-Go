package economy

import (
	"strconv"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type coinGameSessionRecord struct {
	session coinGameSession
	busy    bool
	claimID uint64
}

type coinGameSessionStore struct {
	mu          sync.Mutex
	clock       ports.Clock
	nextID      int64
	nextClaimID uint64
	sessions    map[string]*coinGameSessionRecord
}

func newCoinGameSessionStore(clock ports.Clock) *coinGameSessionStore {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return &coinGameSessionStore{clock: clock, sessions: map[string]*coinGameSessionRecord{}}
}

func (s *coinGameSessionStore) Put(session coinGameSession) coinGameSession {
	if s == nil {
		return session
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	if session.ID == "" {
		s.nextID++
		session.ID = strconv.FormatInt(s.nextID, 10)
	}
	now := s.clock.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now
	s.sessions[session.ID] = &coinGameSessionRecord{session: session}
	return session
}

func (s *coinGameSessionStore) GetForComponent(guildID string, actorID string, channelID string, messageID string) (coinGameSession, bool) {
	if s == nil {
		return coinGameSession{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	record := s.componentRecordLocked(guildID, actorID, channelID, messageID)
	if record == nil {
		return coinGameSession{}, false
	}
	s.bindMessageLocked(record, channelID, messageID)
	return record.session, true
}

func (s *coinGameSessionStore) ClaimForComponent(guildID string, actorID string, channelID string, messageID string) (*coinGameSessionClaim, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	record := s.componentRecordLocked(guildID, actorID, channelID, messageID)
	if record == nil {
		return nil, false
	}
	s.bindMessageLocked(record, channelID, messageID)
	return s.claimLocked(record), true
}

type coinGameTimeoutClaimStatus uint8

const (
	coinGameTimeoutClaimMissing coinGameTimeoutClaimStatus = iota
	coinGameTimeoutClaimBusy
	coinGameTimeoutClaimNotDue
	coinGameTimeoutClaimStale
	coinGameTimeoutClaimed
)

func (s *coinGameSessionStore) ClaimForTimeout(sessionID string, generation uint64) (*coinGameSessionClaim, coinGameTimeoutClaimStatus) {
	if s == nil {
		return nil, coinGameTimeoutClaimMissing
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.sessions[sessionID]
	if record == nil || record.session.State != coinGameSessionActive {
		return nil, coinGameTimeoutClaimMissing
	}
	if record.session.TurnGeneration != generation {
		return nil, coinGameTimeoutClaimStale
	}
	if record.busy {
		return nil, coinGameTimeoutClaimBusy
	}
	if record.session.TurnDeadline.IsZero() || s.clock.Now().Before(record.session.TurnDeadline) {
		return nil, coinGameTimeoutClaimNotDue
	}
	return s.claimLocked(record), coinGameTimeoutClaimed
}

func (s *coinGameSessionStore) componentRecordLocked(guildID string, actorID string, channelID string, messageID string) *coinGameSessionRecord {
	var unbound *coinGameSessionRecord
	for _, record := range s.sessions {
		session := record.session
		if session.GuildID != guildID {
			continue
		}
		if actorID != session.ChallengerID && actorID != session.OpponentID {
			continue
		}
		if session.ChannelID != "" && channelID != "" && session.ChannelID != channelID {
			continue
		}
		if session.MessageID != "" {
			if session.MessageID != messageID {
				continue
			}
			if record.busy || s.expiredActiveLocked(session) {
				return nil
			}
			return record
		}
		if unbound == nil && !record.busy && !s.expiredActiveLocked(session) {
			unbound = record
		}
	}
	return unbound
}

func (s *coinGameSessionStore) bindMessageLocked(record *coinGameSessionRecord, channelID string, messageID string) {
	if record == nil {
		return
	}
	if record.session.ChannelID == "" {
		record.session.ChannelID = channelID
	}
	if record.session.MessageID == "" {
		record.session.MessageID = messageID
	}
	record.session.UpdatedAt = s.clock.Now()
}

func (s *coinGameSessionStore) claimLocked(record *coinGameSessionRecord) *coinGameSessionClaim {
	s.nextClaimID++
	record.busy = true
	record.claimID = s.nextClaimID
	return &coinGameSessionClaim{
		store:   s,
		id:      record.session.ID,
		claimID: record.claimID,
		session: record.session,
	}
}

func (s *coinGameSessionStore) finishClaim(id string, claimID uint64, session coinGameSession, deleteSession bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.sessions[id]
	if record == nil || !record.busy || record.claimID != claimID {
		return false
	}
	if deleteSession {
		delete(s.sessions, id)
		return true
	}
	session.ID = id
	session.UpdatedAt = s.clock.Now()
	record.session = session
	record.busy = false
	return true
}

func (s *coinGameSessionStore) expiredActiveLocked(session coinGameSession) bool {
	return session.State == coinGameSessionActive && !session.TurnDeadline.IsZero() && !s.clock.Now().Before(session.TurnDeadline)
}

func (s *coinGameSessionStore) pruneLocked() {
	now := s.clock.Now()
	for id, record := range s.sessions {
		if record.busy || record.session.State != coinGameSessionPending {
			continue
		}
		if !record.session.CreatedAt.IsZero() && !now.Before(record.session.CreatedAt.Add(coinGameInviteTTL)) {
			delete(s.sessions, id)
		}
	}
}

type coinGameSessionClaim struct {
	mu      sync.Mutex
	store   *coinGameSessionStore
	id      string
	claimID uint64
	session coinGameSession
	done    bool
}

func (c *coinGameSessionClaim) Session() coinGameSession {
	if c == nil {
		return coinGameSession{}
	}
	return c.session
}

func (c *coinGameSessionClaim) Commit(session coinGameSession) bool {
	return c.finish(session, false)
}

func (c *coinGameSessionClaim) Delete() bool {
	return c.finish(coinGameSession{}, true)
}

func (c *coinGameSessionClaim) Restore() {
	if c != nil {
		c.finish(c.session, false)
	}
}

func (c *coinGameSessionClaim) Abandon() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.done = true
	c.mu.Unlock()
}

func (c *coinGameSessionClaim) finish(session coinGameSession, deleteSession bool) bool {
	if c == nil || c.store == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.done {
		return false
	}
	if !c.store.finishClaim(c.id, c.claimID, session, deleteSession) {
		return false
	}
	c.done = true
	return true
}
