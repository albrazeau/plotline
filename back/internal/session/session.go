package session

import "main/internal/session/store"

type Session struct {
	s store.Store
}

func New(s store.Store) *Session {
	return &Session{s: s}
}
