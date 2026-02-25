package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPostNotesSuccess(t *testing.T) {
	server := newTestServer()

	res := performJSON(t, server, http.MethodPost, "/notes", `{"title":"buy milk"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("status=%d, want=%d", res.Code, http.StatusCreated)
	}

	var got Note
	decodeJSON(t, res, &got)

	if got.ID != 1 {
		t.Fatalf("id=%d, want=1", got.ID)
	}
	if got.Title != "buy milk" {
		t.Fatalf("title=%q, want=%q", got.Title, "buy milk")
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("createdAt is zero")
	}
	if got.CreatedAt.Location() != time.UTC {
		t.Fatalf("createdAt location=%v, want=%v", got.CreatedAt.Location(), time.UTC)
	}
}

func TestPostNotesValidation(t *testing.T) {
	server := newTestServer()

	cases := []struct {
		name string
		body string
	}{
		{name: "empty", body: `{"title":""}`},
		{name: "too_long", body: fmt.Sprintf(`{"title":"%s"}`, strings.Repeat("a", 51))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := performJSON(t, server, http.MethodPost, "/notes", tc.body)
			if res.Code != http.StatusBadRequest {
				t.Fatalf("status=%d, want=%d", res.Code, http.StatusBadRequest)
			}
			var got apiErrorResponse
			decodeJSON(t, res, &got)
			if got.Error.Code != "INVALID_ARGUMENT" {
				t.Fatalf("code=%q, want=%q", got.Error.Code, "INVALID_ARGUMENT")
			}
		})
	}
}

func TestGetNotesReturnsNewestFirst(t *testing.T) {
	server := newTestServer()

	performJSON(t, server, http.MethodPost, "/notes", `{"title":"old"}`)
	performJSON(t, server, http.MethodPost, "/notes", `{"title":"new"}`)

	res := performJSON(t, server, http.MethodGet, "/notes", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d, want=%d", res.Code, http.StatusOK)
	}

	var got listNotesResponse
	decodeJSON(t, res, &got)

	if len(got.Notes) != 2 {
		t.Fatalf("len(notes)=%d, want=2", len(got.Notes))
	}
	if got.Notes[0].Title != "new" || got.Notes[1].Title != "old" {
		t.Fatalf("order=%q,%q, want=%q,%q", got.Notes[0].Title, got.Notes[1].Title, "new", "old")
	}
}

func TestPostNotesOverCapacityDropsOldest(t *testing.T) {
	server := newTestServer()

	for i := 1; i <= 51; i++ {
		body := fmt.Sprintf(`{"title":"note-%d"}`, i)
		res := performJSON(t, server, http.MethodPost, "/notes", body)
		if res.Code != http.StatusCreated {
			t.Fatalf("status=%d, want=%d at i=%d", res.Code, http.StatusCreated, i)
		}
	}

	res := performJSON(t, server, http.MethodGet, "/notes", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d, want=%d", res.Code, http.StatusOK)
	}

	var got listNotesResponse
	decodeJSON(t, res, &got)

	if len(got.Notes) != 50 {
		t.Fatalf("len(notes)=%d, want=50", len(got.Notes))
	}
	if got.Notes[0].Title != "note-51" {
		t.Fatalf("newest=%q, want=%q", got.Notes[0].Title, "note-51")
	}
	if got.Notes[len(got.Notes)-1].Title != "note-2" {
		t.Fatalf("oldest=%q, want=%q", got.Notes[len(got.Notes)-1].Title, "note-2")
	}
}

func newTestServer() http.Handler {
	store := NewNoteStore(50, time.Now)
	return NewServer(store)
}

func performJSON(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}

func decodeJSON(t *testing.T, res *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}
