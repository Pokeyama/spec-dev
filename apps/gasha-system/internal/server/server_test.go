package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"gashasystem/internal/config"
	"gashasystem/internal/domain"
	"gashasystem/internal/persistence"
	"gashasystem/internal/session"
)

type fakeRepository struct {
	mu sync.Mutex

	nextAccountID int64

	accountsByLogin map[string]domain.Account
	accountsByID    map[int64]domain.Account
	historyByID     map[int64][]domain.AccountReward
	rewardNames     []string
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		nextAccountID:   1,
		accountsByLogin: map[string]domain.Account{},
		accountsByID:    map[int64]domain.Account{},
		historyByID:     map[int64][]domain.AccountReward{},
		rewardNames:     []string{"Pikachu", "Bulbasaur", "Charmander", "Squirtle"},
	}
}

func (r *fakeRepository) CreateAccount(loginID, hash, role string, credit int) (domain.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accountsByLogin[loginID]; exists {
		return domain.Account{}, persistence.ErrAlreadyExists
	}

	acc := domain.Account{
		AccountID:    r.nextAccountID,
		LoginID:      loginID,
		PasswordHash: hash,
		Role:         role,
		Credit:       credit,
		CreatedAt:    time.Now().UTC(),
	}
	r.nextAccountID++

	r.accountsByLogin[loginID] = acc
	r.accountsByID[acc.AccountID] = acc

	return acc, nil
}

func (r *fakeRepository) FindAccountByLoginID(loginID string) (domain.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accountsByLogin[loginID]
	if !ok {
		return domain.Account{}, persistence.ErrNotFound
	}
	return acc, nil
}

func (r *fakeRepository) Inventory(accountID int64) ([]domain.InventoryItem, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accountsByID[accountID]
	if !ok || acc.Role != "user" {
		return nil, 0, persistence.ErrNotFound
	}

	counter := map[string]int{}
	for _, rw := range r.historyByID[accountID] {
		counter[rw.Name]++
	}

	items := make([]domain.InventoryItem, 0, len(counter))
	for name, count := range counter {
		items = append(items, domain.InventoryItem{Name: name, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, acc.Credit, nil
}

func (r *fakeRepository) Draw(accountID int64, count int, cost int) ([]domain.RewardResult, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accountsByID[accountID]
	if !ok || acc.Role != "user" {
		return nil, 0, persistence.ErrNotFound
	}
	if acc.Credit < cost {
		return nil, 0, persistence.ErrInsufficient
	}

	history := r.historyByID[accountID]
	now := time.Now().UTC()
	results := make([]domain.RewardResult, 0, count)

	for i := 0; i < count; i++ {
		name := r.rewardNames[(len(history)+i)%len(r.rewardNames)]
		history = append(history, domain.AccountReward{
			Name:       name,
			ObtainedAt: now.Add(time.Duration(i) * time.Millisecond),
		})
		results = append(results, domain.RewardResult{Name: name})
	}

	acc.Credit -= cost
	r.accountsByID[accountID] = acc
	r.accountsByLogin[acc.LoginID] = acc
	r.historyByID[accountID] = history

	return results, acc.Credit, nil
}

func (r *fakeRepository) ListUserAccounts() ([]domain.AccountSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	list := make([]domain.AccountSummary, 0)
	for _, acc := range r.accountsByID {
		if acc.Role != "user" {
			continue
		}
		list = append(list, domain.AccountSummary{
			AccountID: acc.AccountID,
			LoginID:   acc.LoginID,
			Credit:    acc.Credit,
			CreatedAt: acc.CreatedAt,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].AccountID < list[j].AccountID
	})

	return list, nil
}

func (r *fakeRepository) AccountDetail(accountID int64) (domain.Account, []domain.AccountReward, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accountsByID[accountID]
	if !ok || acc.Role != "user" {
		return domain.Account{}, nil, persistence.ErrNotFound
	}

	history := append([]domain.AccountReward(nil), r.historyByID[accountID]...)
	sort.Slice(history, func(i, j int) bool {
		return history[i].ObtainedAt.After(history[j].ObtainedAt)
	})

	return acc, history, nil
}

type fakeSessionStore struct {
	mu    sync.Mutex
	items map[string]fakeSessionItem
}

type fakeSessionItem struct {
	payload session.Payload
	exp     time.Time
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{items: map[string]fakeSessionItem{}}
}

func (s *fakeSessionStore) Save(token string, payload session.Payload, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[token] = fakeSessionItem{payload: payload, exp: time.Now().UTC().Add(ttl)}
	return nil
}

func (s *fakeSessionStore) Get(token string) (session.Payload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok := s.items[token]
	if !ok {
		return session.Payload{}, session.ErrNotFound
	}
	if time.Now().UTC().After(it.exp) || it.payload.Exp <= time.Now().UTC().Unix() {
		delete(s.items, token)
		return session.Payload{}, session.ErrNotFound
	}
	return it.payload, nil
}

func (s *fakeSessionStore) Delete(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, token)
	return nil
}

func TestAPIUserAdminFlowAndRoleGuard(t *testing.T) {
	repo := newFakeRepository()
	sessions := newFakeSessionStore()
	h := NewMuxWithDeps(repo, sessions, config.Config{SessionTTL: time.Hour})

	status, _ := requestJSON(t, h, http.MethodPost, "/regist", "", map[string]string{
		"id":       "alice",
		"password": "pass123",
	})
	if status != http.StatusCreated {
		t.Fatalf("regist status=%d", status)
	}

	status, loginBody := requestJSON(t, h, http.MethodGet, "/llogin?id=alice&password=pass123", "", nil)
	if status != http.StatusOK {
		t.Fatalf("user login status=%d", status)
	}
	userToken := stringAt(loginBody, "sessionToken")
	if userToken == "" {
		t.Fatalf("user sessionToken is empty")
	}

	status, gashaBody := requestJSON(t, h, http.MethodPost, "/gasha", userToken, nil)
	if status != http.StatusOK {
		t.Fatalf("gasha status=%d", status)
	}
	if numberAt(gashaBody, "remainingCredit") != 990 {
		t.Fatalf("remaining credit=%v", gashaBody["remainingCredit"])
	}

	status, invBody := requestJSON(t, h, http.MethodGet, "/inventory", userToken, nil)
	if status != http.StatusOK {
		t.Fatalf("inventory status=%d", status)
	}
	if numberAt(invBody, "credit") != 990 {
		t.Fatalf("inventory credit=%v", invBody["credit"])
	}

	status, _ = requestJSON(t, h, http.MethodPost, "/admin/regist", "", map[string]string{
		"id":       "admin",
		"password": "pass123",
	})
	if status != http.StatusCreated {
		t.Fatalf("admin regist status=%d", status)
	}

	status, adminLoginBody := requestJSON(t, h, http.MethodGet, "/admin/login?id=admin&password=pass123", "", nil)
	if status != http.StatusOK {
		t.Fatalf("admin login status=%d", status)
	}
	adminToken := stringAt(adminLoginBody, "adminSessionToken")
	if adminToken == "" {
		t.Fatalf("adminSessionToken is empty")
	}

	status, listBody := requestJSON(t, h, http.MethodGet, "/account/list", adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("account list status=%d", status)
	}
	if len(arrayAt(listBody, "accounts")) != 1 {
		t.Fatalf("accounts len=%d", len(arrayAt(listBody, "accounts")))
	}

	status, forbiddenBody := requestJSON(t, h, http.MethodGet, "/account/list", userToken, nil)
	if status != http.StatusForbidden {
		t.Fatalf("account list by user status=%d", status)
	}
	if errorCode(forbiddenBody) != "FORBIDDEN" {
		t.Fatalf("forbidden code=%q", errorCode(forbiddenBody))
	}

	status, detailBody := requestJSON(t, h, http.MethodGet, "/account/detail?id=1", adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("account detail status=%d", status)
	}
	if stringAt(detailBody, "login_id") != "alice" {
		t.Fatalf("detail login_id=%q", stringAt(detailBody, "login_id"))
	}

	status, _ = requestJSON(t, h, http.MethodGet, "/logout", userToken, nil)
	if status != http.StatusOK {
		t.Fatalf("logout status=%d", status)
	}

	status, afterLogout := requestJSON(t, h, http.MethodGet, "/inventory", userToken, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("inventory after logout status=%d", status)
	}
	if errorCode(afterLogout) != "UNAUTHENTICATED" {
		t.Fatalf("after logout code=%q", errorCode(afterLogout))
	}
}

func TestAPIInsufficientDiamonds(t *testing.T) {
	repo := newFakeRepository()
	sessions := newFakeSessionStore()
	h := NewMuxWithDeps(repo, sessions, config.Config{SessionTTL: time.Hour})

	status, _ := requestJSON(t, h, http.MethodPost, "/regist", "", map[string]string{
		"id":       "bob",
		"password": "pass123",
	})
	if status != http.StatusCreated {
		t.Fatalf("regist status=%d", status)
	}

	status, loginBody := requestJSON(t, h, http.MethodGet, "/llogin?id=bob&password=pass123", "", nil)
	if status != http.StatusOK {
		t.Fatalf("login status=%d", status)
	}
	token := stringAt(loginBody, "sessionToken")

	for i := 0; i < 10; i++ {
		status, _ = requestJSON(t, h, http.MethodPost, "/gasha/ten", token, nil)
		if status != http.StatusOK {
			t.Fatalf("gasha/ten #%d status=%d", i+1, status)
		}
	}

	status, body := requestJSON(t, h, http.MethodPost, "/gasha/ten", token, nil)
	if status != http.StatusPaymentRequired {
		t.Fatalf("insufficient status=%d", status)
	}
	if errorCode(body) != "INSUFFICIENT_DIAMONDS" {
		t.Fatalf("insufficient code=%q", errorCode(body))
	}

	status, invBody := requestJSON(t, h, http.MethodGet, "/inventory", token, nil)
	if status != http.StatusOK {
		t.Fatalf("inventory status=%d", status)
	}
	if numberAt(invBody, "credit") != 0 {
		t.Fatalf("credit=%v", invBody["credit"])
	}
}

func TestAPIInvalidArgumentsAndExpiredSession(t *testing.T) {
	repo := newFakeRepository()
	sessions := newFakeSessionStore()
	h := NewMuxWithDeps(repo, sessions, config.Config{SessionTTL: time.Hour})

	status, invalidBody := requestRawJSON(t, h, http.MethodPost, "/regist", "", `{"id":"a"`)
	if status != http.StatusBadRequest {
		t.Fatalf("invalid json status=%d", status)
	}
	if errorCode(invalidBody) != "INVALID_ARGUMENT" {
		t.Fatalf("invalid json code=%q", errorCode(invalidBody))
	}

	status, _ = requestJSON(t, h, http.MethodPost, "/regist", "", map[string]string{
		"id":       "eve",
		"password": "pass123",
	})
	if status != http.StatusCreated {
		t.Fatalf("regist status=%d", status)
	}

	if err := sessions.Save("expired-token", session.Payload{
		AccountID: 1,
		Role:      "user",
		Exp:       time.Now().UTC().Add(-1 * time.Minute).Unix(),
	}, time.Hour); err != nil {
		t.Fatalf("save expired token: %v", err)
	}

	status, expiredBody := requestJSON(t, h, http.MethodGet, "/inventory", "expired-token", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("expired session status=%d", status)
	}
	if errorCode(expiredBody) != "UNAUTHENTICATED" {
		t.Fatalf("expired session code=%q", errorCode(expiredBody))
	}

	status, noAuthBody := requestJSON(t, h, http.MethodGet, "/inventory", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("no auth status=%d", status)
	}
	if errorCode(noAuthBody) != "UNAUTHENTICATED" {
		t.Fatalf("no auth code=%q", errorCode(noAuthBody))
	}
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{name: "ok", header: "Bearer abc", want: "abc"},
		{name: "case_insensitive", header: "bearer xyz", want: "xyz"},
		{name: "missing", header: "", wantErr: true},
		{name: "invalid", header: "Token abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bearerToken(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestIsValidCredential(t *testing.T) {
	if !isValidCredential("alice", "password") {
		t.Fatal("expected valid")
	}
	if isValidCredential("", "password") {
		t.Fatal("expected invalid")
	}
	if isValidCredential("alice", "") {
		t.Fatal("expected invalid")
	}
}

func TestHostOnly(t *testing.T) {
	if got := hostOnly("example.com:8080"); got != "example.com" {
		t.Fatalf("got=%q", got)
	}
	if got := hostOnly("example.com"); got != "example.com" {
		t.Fatalf("got=%q", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	repo := newFakeRepository()
	sessions := newFakeSessionStore()
	h := NewMuxWithDeps(repo, sessions, config.Config{SessionTTL: time.Hour})

	req := httptest.NewRequest(http.MethodOptions, "/regist", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow-origin=%q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Content-Type") {
		t.Fatalf("allow-headers=%q", got)
	}
}

func requestJSON(t *testing.T, h http.Handler, method, path, token string, body any) (int, map[string]any) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	parsed := map[string]any{}
	trimmed := strings.TrimSpace(string(data))
	if trimmed != "" {
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("unmarshal response: %v body=%s", err, string(data))
		}
	}

	return resp.StatusCode, parsed
}

func requestRawJSON(t *testing.T, h http.Handler, method, path, token, raw string) (int, map[string]any) {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	parsed := map[string]any{}
	trimmed := strings.TrimSpace(string(data))
	if trimmed != "" {
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("unmarshal response: %v body=%s", err, string(data))
		}
	}

	return resp.StatusCode, parsed
}

func stringAt(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func numberAt(m map[string]any, key string) int {
	v, ok := m[key].(float64)
	if !ok {
		return 0
	}
	return int(v)
}

func arrayAt(m map[string]any, key string) []any {
	v, _ := m[key].([]any)
	return v
}

func errorCode(m map[string]any) string {
	errObj, ok := m["error"].(map[string]any)
	if !ok {
		return ""
	}
	code, _ := errObj["code"].(string)
	return code
}
