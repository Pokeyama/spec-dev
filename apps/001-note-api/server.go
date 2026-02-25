package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"
)

const (
	maxTitleChars = 50
)

type Server struct {
	store *NoteStore
}

func NewServer(store *NoteStore) http.Handler {
	s := &Server{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/notes", s.handleNotes)
	return mux
}

func (s *Server) handleNotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateNote(w, r)
	case http.MethodGet:
		s.handleListNotes(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type createNoteRequest struct {
	Title string `json:"title"`
}

type listNotesResponse struct {
	Notes []Note `json:"notes"`
}

type apiErrorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *Server) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req createNoteRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "invalid JSON body")
		return
	}

	if err := validateTitle(req.Title); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "title must be 1-50 characters")
		return
	}

	note := s.store.Create(strings.TrimSpace(req.Title))
	writeJSON(w, http.StatusCreated, note)
}

func (s *Server) handleListNotes(w http.ResponseWriter, r *http.Request) {
	notes := s.store.ListNewestFirst()
	writeJSON(w, http.StatusOK, listNotesResponse{Notes: notes})
}

func validateTitle(title string) error {
	trimmed := strings.TrimSpace(title)
	if utf8.RuneCountInString(trimmed) < 1 || utf8.RuneCountInString(trimmed) > maxTitleChars {
		return errInvalidTitle
	}
	return nil
}

var errInvalidTitle = &validationError{}

type validationError struct{}

func (e *validationError) Error() string {
	return "invalid title"
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiErrorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
