package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type Payload struct {
	AccountID int64  `json:"account_id"`
	Role      string `json:"role"`
	Exp       int64  `json:"exp"`
}

type Store struct {
	client *memcache.Client
}

func NewStore(addr string) *Store {
	c := memcache.New(addr)
	c.Timeout = 2 * time.Second
	return &Store{client: c}
}

func (s *Store) Save(token string, payload Payload, ttl time.Duration) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	exp := int32(ttl.Seconds())
	if exp <= 0 {
		exp = 1
	}

	return s.client.Set(&memcache.Item{
		Key:        sessionKey(token),
		Value:      b,
		Expiration: exp,
	})
}

func (s *Store) Get(token string) (Payload, error) {
	it, err := s.client.Get(sessionKey(token))
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return Payload{}, ErrNotFound
		}
		return Payload{}, err
	}

	var payload Payload
	if err := json.Unmarshal(it.Value, &payload); err != nil {
		return Payload{}, err
	}

	if payload.Exp <= time.Now().UTC().Unix() {
		_ = s.client.Delete(sessionKey(token))
		return Payload{}, ErrNotFound
	}
	return payload, nil
}

func (s *Store) Delete(token string) error {
	err := s.client.Delete(sessionKey(token))
	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil
	}
	return err
}

func sessionKey(token string) string {
	return "session:" + token
}

func NewToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

var ErrNotFound = errors.New("session not found")
