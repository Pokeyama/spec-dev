CREATE TABLE IF NOT EXISTS accounts (
  account_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  login_id VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role ENUM('user', 'admin') NOT NULL DEFAULT 'user',
  credit INT NOT NULL DEFAULT 1000,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (account_id),
  UNIQUE KEY uk_accounts_login_id (login_id),
  KEY idx_accounts_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS rewards (
  reward_id INT UNSIGNED NOT NULL,
  name VARCHAR(64) NOT NULL,
  PRIMARY KEY (reward_id),
  KEY idx_rewards_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reward_history (
  reward_history_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id BIGINT UNSIGNED NOT NULL,
  reward_id INT UNSIGNED NOT NULL,
  obtained_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (reward_history_id),
  KEY idx_reward_history_account_id_obtained_at (account_id, obtained_at DESC),
  KEY idx_reward_history_reward_id (reward_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
