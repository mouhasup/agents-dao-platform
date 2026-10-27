package service

import (
	"agents-dao-platform/model"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// SessionRepository gère la persistance des tokens OAuth par session
type SessionRepository interface {
	Create(token *oauth2.Token) (string, error)
	GetToken(sessionID string) (*oauth2.Token, error)
	Delete(sessionID string) error
}

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toSession(id string, t *oauth2.Token) *model.UserSession {
	return &model.UserSession{
		ID:           id,
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		Expiry:       t.Expiry,
	}
}

func toToken(s *model.UserSession) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		TokenType:    s.TokenType,
		Expiry:       s.Expiry,
	}
}

// ---------- GORM ----------

type gormSessionRepository struct{ db *gorm.DB }

func NewGormSessionRepository(db *gorm.DB) SessionRepository {
	return &gormSessionRepository{db: db}
}

func (r *gormSessionRepository) Create(token *oauth2.Token) (string, error) {
	id, err := newSessionID()
	if err != nil {
		return "", err
	}
	return id, r.db.Create(toSession(id, token)).Error
}

func (r *gormSessionRepository) GetToken(sessionID string) (*oauth2.Token, error) {
	var s model.UserSession
	if err := r.db.First(&s, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}
	return toToken(&s), nil
}

func (r *gormSessionRepository) Delete(sessionID string) error {
	return r.db.Delete(&model.UserSession{}, "id = ?", sessionID).Error
}

// ---------- Mémoire ----------

type memorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*model.UserSession
}

func NewMemorySessionRepository() SessionRepository {
	return &memorySessionRepository{sessions: make(map[string]*model.UserSession)}
}

func (r *memorySessionRepository) Create(token *oauth2.Token) (string, error) {
	id, err := newSessionID()
	if err != nil {
		return "", err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[id] = toSession(id, token)
	return id, nil
}

func (r *memorySessionRepository) GetToken(sessionID string) (*oauth2.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[sessionID]
	if !ok {
		return nil, errors.New("session introuvable")
	}
	return toToken(s), nil
}

func (r *memorySessionRepository) Delete(sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}
