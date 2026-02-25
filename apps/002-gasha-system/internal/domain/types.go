package domain

import "time"

type Account struct {
	AccountID    int64
	LoginID      string
	PasswordHash string
	Role         string
	Credit       int
	CreatedAt    time.Time
}

type InventoryItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type RewardResult struct {
	Name string `json:"name"`
}

type AccountSummary struct {
	AccountID int64     `json:"account_id"`
	LoginID   string    `json:"login_id"`
	Credit    int       `json:"credit"`
	CreatedAt time.Time `json:"createdAt"`
}

type AccountReward struct {
	Name       string    `json:"name"`
	ObtainedAt time.Time `json:"obtainedAt"`
}
