package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mysql "github.com/go-sql-driver/mysql"

	"gashasystem/internal/config"
	"gashasystem/internal/domain"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInsufficient  = errors.New("insufficient diamonds")
)

type rewardMaster struct {
	RewardID int
	Name     string
}

type Repository struct {
	db *sql.DB

	rngMu sync.Mutex
	rng   *rand.Rand
}

func NewRepository(cfg config.Config) (*Repository, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Repository{
		db:  db,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (r *Repository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *Repository) CreateAccount(loginID, hash, role string, credit int) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO accounts (login_id, password_hash, role, credit) VALUES (?, ?, ?, ?)`,
		loginID,
		hash,
		role,
		credit,
	)
	if err != nil {
		var myErr *mysql.MySQLError
		if errors.As(err, &myErr) && myErr.Number == 1062 {
			return domain.Account{}, ErrAlreadyExists
		}
		return domain.Account{}, err
	}

	return r.FindAccountByLoginID(loginID)
}

func (r *Repository) FindAccountByLoginID(loginID string) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(
		ctx,
		`SELECT account_id, login_id, password_hash, role, credit, created_at
		 FROM accounts
		 WHERE login_id = ?
		 LIMIT 1`,
		loginID,
	)

	var acc domain.Account
	if err := row.Scan(&acc.AccountID, &acc.LoginID, &acc.PasswordHash, &acc.Role, &acc.Credit, &acc.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, ErrNotFound
		}
		return domain.Account{}, err
	}
	return acc, nil
}

func (r *Repository) FindUserAccountByID(accountID int64) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(
		ctx,
		`SELECT account_id, login_id, password_hash, role, credit, created_at
		 FROM accounts
		 WHERE account_id = ? AND role = 'user'
		 LIMIT 1`,
		accountID,
	)

	var acc domain.Account
	if err := row.Scan(&acc.AccountID, &acc.LoginID, &acc.PasswordHash, &acc.Role, &acc.Credit, &acc.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, ErrNotFound
		}
		return domain.Account{}, err
	}
	return acc, nil
}

func (r *Repository) ListUserAccounts() ([]domain.AccountSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT account_id, login_id, credit, created_at
		 FROM accounts
		 WHERE role = 'user'
		 ORDER BY account_id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.AccountSummary, 0)
	for rows.Next() {
		var acc domain.AccountSummary
		if err := rows.Scan(&acc.AccountID, &acc.LoginID, &acc.Credit, &acc.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repository) Inventory(accountID int64) ([]domain.InventoryItem, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var credit int
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT credit FROM accounts WHERE account_id = ? AND role = 'user' LIMIT 1`,
		accountID,
	).Scan(&credit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT r.name, COUNT(*) AS cnt
		 FROM reward_history h
		 JOIN rewards r ON r.reward_id = h.reward_id
		 WHERE h.account_id = ?
		 GROUP BY r.name
		 ORDER BY r.name ASC`,
		accountID,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.InventoryItem, 0)
	for rows.Next() {
		var item domain.InventoryItem
		if err := rows.Scan(&item.Name, &item.Count); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, credit, nil
}

func (r *Repository) Draw(accountID int64, count int, cost int) ([]domain.RewardResult, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var credit int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT credit FROM accounts WHERE account_id = ? AND role = 'user' FOR UPDATE`,
		accountID,
	).Scan(&credit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, err
	}

	if credit < cost {
		return nil, 0, ErrInsufficient
	}

	masters, err := loadRewardMasters(ctx, tx)
	if err != nil {
		return nil, 0, err
	}
	if len(masters) == 0 {
		return nil, 0, errors.New("rewards not initialized")
	}

	selected := r.pickRewards(masters, count)
	if err := insertRewardHistory(ctx, tx, accountID, selected); err != nil {
		return nil, 0, err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE accounts
		 SET credit = credit - ?, updated_at = UTC_TIMESTAMP(6)
		 WHERE account_id = ?`,
		cost,
		accountID,
	); err != nil {
		return nil, 0, err
	}

	remaining := credit - cost
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	results := make([]domain.RewardResult, 0, len(selected))
	for _, s := range selected {
		results = append(results, domain.RewardResult{Name: s.Name})
	}
	return results, remaining, nil
}

func loadRewardMasters(ctx context.Context, tx *sql.Tx) ([]rewardMaster, error) {
	rows, err := tx.QueryContext(ctx, `SELECT reward_id, name FROM rewards ORDER BY reward_id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	masters := make([]rewardMaster, 0, 1024)
	for rows.Next() {
		var m rewardMaster
		if err := rows.Scan(&m.RewardID, &m.Name); err != nil {
			return nil, err
		}
		masters = append(masters, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return masters, nil
}

func (r *Repository) pickRewards(masters []rewardMaster, count int) []rewardMaster {
	r.rngMu.Lock()
	defer r.rngMu.Unlock()

	picked := make([]rewardMaster, 0, count)
	for i := 0; i < count; i++ {
		picked = append(picked, masters[r.rng.Intn(len(masters))])
	}
	return picked
}

func insertRewardHistory(ctx context.Context, tx *sql.Tx, accountID int64, selected []rewardMaster) error {
	if len(selected) == 0 {
		return nil
	}

	query := "INSERT INTO reward_history (account_id, reward_id, obtained_at) VALUES "
	args := make([]any, 0, len(selected)*3)
	for i, s := range selected {
		if i > 0 {
			query += ","
		}
		query += "(?, ?, UTC_TIMESTAMP(6))"
		args = append(args, accountID, s.RewardID)
	}

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) AccountDetail(accountID int64) (domain.Account, []domain.AccountReward, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(
		ctx,
		`SELECT account_id, login_id, password_hash, role, credit, created_at
		 FROM accounts
		 WHERE account_id = ? AND role = 'user'
		 LIMIT 1`,
		accountID,
	)

	var acc domain.Account
	if err := row.Scan(&acc.AccountID, &acc.LoginID, &acc.PasswordHash, &acc.Role, &acc.Credit, &acc.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, nil, ErrNotFound
		}
		return domain.Account{}, nil, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT r.name, h.obtained_at
		 FROM reward_history h
		 JOIN rewards r ON r.reward_id = h.reward_id
		 WHERE h.account_id = ?
		 ORDER BY h.obtained_at DESC, h.reward_history_id DESC`,
		accountID,
	)
	if err != nil {
		return domain.Account{}, nil, err
	}
	defer rows.Close()

	rewards := make([]domain.AccountReward, 0)
	for rows.Next() {
		var rw domain.AccountReward
		if err := rows.Scan(&rw.Name, &rw.ObtainedAt); err != nil {
			return domain.Account{}, nil, err
		}
		rewards = append(rewards, rw)
	}
	if err := rows.Err(); err != nil {
		return domain.Account{}, nil, err
	}

	return acc, rewards, nil
}
