package server

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gashasystem/internal/config"
	"gashasystem/internal/domain"
	"gashasystem/internal/persistence"
	"gashasystem/internal/security"
	"gashasystem/internal/session"
)

type accountRepository interface {
	CreateAccount(loginID, hash, role string, credit int) (domain.Account, error)
	FindAccountByLoginID(loginID string) (domain.Account, error)
	Inventory(accountID int64) ([]domain.InventoryItem, int, error)
	Draw(accountID int64, count int, cost int) ([]domain.RewardResult, int, error)
	ListUserAccounts() ([]domain.AccountSummary, error)
	AccountDetail(accountID int64) (domain.Account, []domain.AccountReward, error)
}

type sessionStore interface {
	Save(token string, payload session.Payload, ttl time.Duration) error
	Get(token string) (session.Payload, error)
	Delete(token string) error
}

type API struct {
	repo     accountRepository
	sessions sessionStore
	cfg      config.Config
}

func NewMux(repo *persistence.Repository, sessions *session.Store, cfg config.Config) http.Handler {
	return NewMuxWithDeps(repo, sessions, cfg)
}

func NewMuxWithDeps(repo accountRepository, sessions sessionStore, cfg config.Config) http.Handler {
	a := &API{repo: repo, sessions: sessions, cfg: cfg}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/regist", a.handleUserRegister)
	mux.HandleFunc("/llogin", a.handleUserLogin)
	mux.HandleFunc("/logout", a.handleLogout)
	mux.HandleFunc("/inventory", a.handleInventory)
	mux.HandleFunc("/gasha", a.handleGasha)
	mux.HandleFunc("/gasha/ten", a.handleGashaTen)
	mux.HandleFunc("/admin/regist", a.handleAdminRegister)
	mux.HandleFunc("/admin/login", a.handleAdminLogin)
	mux.HandleFunc("/account/list", a.handleAccountList)
	mux.HandleFunc("/account/detail", a.handleAccountDetail)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type registerRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

func (a *API) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.register(w, r, "user", 1000, true)
}

func (a *API) handleAdminRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// 学習用途で未認証許可。
	a.register(w, r, "admin", 0, false)
}

func (a *API) register(w http.ResponseWriter, r *http.Request, role string, credit int, includeCredit bool) {
	defer r.Body.Close()

	var req registerRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "invalid JSON body")
		return
	}

	loginID := strings.TrimSpace(req.ID)
	password := strings.TrimSpace(req.Password)
	if !isValidCredential(loginID, password) {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "id and password are required")
		return
	}

	hash := security.HashPassword(password)

	_, err := a.repo.CreateAccount(loginID, hash, role, credit)
	if err != nil {
		switch {
		case errors.Is(err, persistence.ErrAlreadyExists):
			writeError(w, http.StatusConflict, "ALREADY_EXISTS", "id already exists")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to create account")
		}
		return
	}

	if includeCredit {
		writeJSON(w, http.StatusCreated, map[string]any{
			"id":     loginID,
			"credit": credit,
			"role":   role,
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":   loginID,
		"role": role,
	})
}

func (a *API) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.login(w, r, false)
}

func (a *API) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.login(w, r, true)
}

func (a *API) login(w http.ResponseWriter, r *http.Request, requireAdmin bool) {
	loginID := strings.TrimSpace(r.URL.Query().Get("id"))
	password := strings.TrimSpace(r.URL.Query().Get("password"))
	if !isValidCredential(loginID, password) {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "id and password are required")
		return
	}

	acc, err := a.repo.FindAccountByLoginID(loginID)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "invalid id or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load account")
		return
	}

	if !security.ComparePassword(acc.PasswordHash, password) {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "invalid id or password")
		return
	}

	if requireAdmin && acc.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "admin role required")
		return
	}

	token, err := session.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to create token")
		return
	}
	payload := session.Payload{
		AccountID: acc.AccountID,
		Role:      acc.Role,
		Exp:       time.Now().UTC().Add(a.cfg.SessionTTL).Unix(),
	}
	if err := a.sessions.Save(token, payload, a.cfg.SessionTTL); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to save session")
		return
	}

	if requireAdmin {
		writeJSON(w, http.StatusOK, map[string]string{"adminSessionToken": token})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"sessionToken": token})
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, token, ok := a.requireAuth(w, r, "user")
	if !ok {
		return
	}
	if err := a.sessions.Delete(token); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to delete session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) handleInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	payload, _, ok := a.requireAuth(w, r, "user")
	if !ok {
		return
	}

	items, credit, err := a.repo.Inventory(payload.AccountID)
	if err != nil {
		switch {
		case errors.Is(err, persistence.ErrNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "account not found")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load inventory")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"credit": credit,
	})
}

func (a *API) handleGasha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.draw(w, r, 1, 10)
}

func (a *API) handleGashaTen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.draw(w, r, 10, 100)
}

func (a *API) draw(w http.ResponseWriter, r *http.Request, count int, cost int) {
	payload, _, ok := a.requireAuth(w, r, "user")
	if !ok {
		return
	}

	rewards, remainingCredit, err := a.repo.Draw(payload.AccountID, count, cost)
	if err != nil {
		switch {
		case errors.Is(err, persistence.ErrInsufficient):
			writeError(w, http.StatusPaymentRequired, "INSUFFICIENT_DIAMONDS", "insufficient diamonds")
		case errors.Is(err, persistence.ErrNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "account not found")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to run gasha")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"consumedCredit":  cost,
		"remainingCredit": remainingCredit,
		"rewards":         rewards,
	})
}

func (a *API) handleAccountList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !a.requireAdminHost(w, r) {
		return
	}
	_, _, ok := a.requireAuth(w, r, "admin")
	if !ok {
		return
	}

	accounts, err := a.repo.ListUserAccounts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to list accounts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"accounts": accounts})
}

func (a *API) handleAccountDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !a.requireAdminHost(w, r) {
		return
	}
	_, _, ok := a.requireAuth(w, r, "admin")
	if !ok {
		return
	}

	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "id is required")
		return
	}
	accountID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || accountID <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "id must be positive integer")
		return
	}

	acc, rewards, err := a.repo.AccountDetail(accountID)
	if err != nil {
		switch {
		case errors.Is(err, persistence.ErrNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "account not found")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load account detail")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"account_id": acc.AccountID,
		"login_id":   acc.LoginID,
		"rewards":    rewards,
	})
}

func (a *API) requireAuth(w http.ResponseWriter, r *http.Request, role string) (session.Payload, string, bool) {
	token, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "invalid authorization header")
		return session.Payload{}, "", false
	}

	payload, err := a.sessions.Get(token)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "invalid session token")
			return session.Payload{}, "", false
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to get session")
		return session.Payload{}, "", false
	}

	if role != "" && payload.Role != role {
		if role == "admin" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "admin role required")
		} else {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "user role required")
		}
		return session.Payload{}, "", false
	}
	return payload, token, true
}

func (a *API) requireAdminHost(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(a.cfg.AdminHost) == "" {
		return true
	}
	host := hostOnly(r.Host)
	if host != a.cfg.AdminHost {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "admin domain required")
		return false
	}
	return true
}

func hostOnly(h string) string {
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h
	}
	return host
}

func bearerToken(header string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid header")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("empty token")
	}
	return token, nil
}

func isValidCredential(loginID, password string) bool {
	if loginID == "" || password == "" {
		return false
	}
	if len(loginID) > 64 || len(password) > 128 {
		return false
	}
	return true
}

type apiErrorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiErrorResponse{Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
