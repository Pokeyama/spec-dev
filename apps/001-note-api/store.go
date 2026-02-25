package main

import (
	"sync"
	"time"
)

type NoteStore struct {
	mu       sync.Mutex
	notes    []Note
	nextID   int64
	maxNotes int
	now      func() time.Time
}

func NewNoteStore(maxNotes int, now func() time.Time) *NoteStore {
	if maxNotes <= 0 {
		maxNotes = 50
	}
	if now == nil {
		now = time.Now
	}
	return &NoteStore{
		notes:    make([]Note, 0, maxNotes),
		nextID:   1,
		maxNotes: maxNotes,
		now:      now,
	}
}

func (s *NoteStore) Create(title string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.notes) >= s.maxNotes {
		s.notes = append([]Note(nil), s.notes[1:]...)
	}

	note := Note{
		ID:        s.nextID,
		Title:     title,
		CreatedAt: s.now().UTC(),
	}
	s.nextID++
	s.notes = append(s.notes, note)
	return note
}

func (s *NoteStore) ListNewestFirst() []Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]Note, 0, len(s.notes))
	for i := len(s.notes) - 1; i >= 0; i-- {
		result = append(result, s.notes[i])
	}
	return result
}
