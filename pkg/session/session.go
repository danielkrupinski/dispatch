package session

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

var (
	CookieName = "session"

	Expiration      = time.Hour * 24 * 7
	RefreshInterval = time.Hour
)

type Session struct {
	UserID uint64

	key       string
	createdAt int64
	lock      sync.Mutex
}

func New(id uint64) (*Session, error) {
	key, err := newSessionKey()
	if err != nil {
		return nil, err
	}

	return &Session{
		key:       key,
		createdAt: time.Now().Unix(),
		UserID:    id,
	}, nil
}

func (s *Session) Key() string {
	s.lock.Lock()
	key := s.key
	s.lock.Unlock()
	return key
}

func (s *Session) SetCookie(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	created := time.Unix(s.createdAt, 0)
	s.lock.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    s.Key(),
		Path:     "/",
		Expires:  created.Add(Expiration),
		HttpOnly: true,
		Secure:   r.TLS != nil,
	})
}

func (s *Session) Expired() bool {
	s.lock.Lock()
	created := time.Unix(s.createdAt, 0)
	s.lock.Unlock()
	return time.Since(created) > Expiration
}

func (s *Session) Refresh() (string, bool, error) {
	s.lock.Lock()
	created := time.Unix(s.createdAt, 0)
	s.lock.Unlock()

	if time.Since(created) > Expiration {
		return "", true, nil
	}

	if time.Since(created) > RefreshInterval {
		key, err := newSessionKey()
		if err != nil {
			return "", false, err
		}

		s.lock.Lock()
		s.createdAt = time.Now().Unix()
		s.key = key
		s.lock.Unlock()
		return key, false, nil
	}

	return "", false, nil
}

func newSessionKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	return base64.RawURLEncoding.EncodeToString(key), err
}
