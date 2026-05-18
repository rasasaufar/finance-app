package store

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rasasaufar/finance-app/api/internal/types"
)

// Store wraps the database connection pool.
type Store struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{DB: db}
}

// EnsureSchema creates all required tables and indexes if they don't exist.
// Also handles migration of existing tables to add account_id column.
func (s *Store) EnsureSchema(ctx context.Context) error {
	statements := []string{
		// Accounts table (replaces user_profiles for auth purposes)
		`CREATE TABLE IF NOT EXISTS accounts (
			id BIGSERIAL PRIMARY KEY,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
			avatar_url TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_accounts_lower_email ON accounts ((LOWER(email)))`,

		// Sessions table for token-based auth
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		// Keep old user_profiles for backward compat (no-op if already exists)
		`CREATE TABLE IF NOT EXISTS user_profiles (
			id BIGSERIAL PRIMARY KEY,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL DEFAULT '',
			avatar_url TEXT NOT NULL DEFAULT '',
			avatar_initials TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}

	for _, stmt := range statements {
		if _, err := s.DB.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	// Migration: ensure categories table has account_id column
	// If the table doesn't exist yet, create it fresh with account_id.
	// If it exists without account_id, add the column.
	if err := s.migrateCategories(ctx); err != nil {
		return err
	}
	if err := s.migrateBudgetRules(ctx); err != nil {
		return err
	}
	if err := s.migrateTransactions(ctx); err != nil {
		return err
	}
	if err := s.migrateSalaryMasters(ctx); err != nil {
		return err
	}
	if err := s.migrateWeddingSavings(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Store) migrateCategories(ctx context.Context) error {
	// Create table if not exists (new schema with account_id)
	if _, err := s.DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS categories (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (account_id, name)
		)`); err != nil {
		return err
	}
	// Add account_id column if missing (migration for existing DB)
	if _, err := s.DB.Exec(ctx, `
		ALTER TABLE categories ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE`); err != nil {
		return err
	}
	// Drop old single-column unique constraint if it exists (ignore error if not found)
	_, _ = s.DB.Exec(ctx, `ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_name_key`)
	_, _ = s.DB.Exec(ctx, `DROP INDEX IF EXISTS uq_categories_lower_name`)
	if _, err := s.DB.Exec(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_account_lower_name ON categories (account_id, (LOWER(name)))`); err != nil {
		return err
	}
	return nil
}

func (s *Store) migrateBudgetRules(ctx context.Context) error {
	if _, err := s.DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS budget_rules (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			period TEXT NOT NULL CHECK (period IN ('daily', 'weekly', 'monthly')),
			limit_amount BIGINT NOT NULL CHECK (limit_amount > 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (account_id, category_id, period)
		)`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(ctx, `
		ALTER TABLE budget_rules ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE`); err != nil {
		return err
	}
	// Drop old unique constraint (category_id, period) if exists
	_, _ = s.DB.Exec(ctx, `ALTER TABLE budget_rules DROP CONSTRAINT IF EXISTS budget_rules_category_id_period_key`)
	return nil
}

func (s *Store) migrateTransactions(ctx context.Context) error {
	if _, err := s.DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS transactions (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
			category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
			amount BIGINT NOT NULL CHECK (amount > 0),
			trx_date DATE NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(ctx, `
		ALTER TABLE transactions ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_transactions_account_trx_date ON transactions (account_id, trx_date)`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions (category_id)`); err != nil {
		return err
	}
	return nil
}

func (s *Store) migrateSalaryMasters(ctx context.Context) error {
	if _, err := s.DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS salary_masters (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			month TEXT NOT NULL CHECK (month ~ '^[0-9]{4}-[0-9]{2}$'),
			amount BIGINT NOT NULL CHECK (amount > 0),
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (account_id, month)
		)`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(ctx, `
		ALTER TABLE salary_masters ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE`); err != nil {
		return err
	}
	// Drop old unique constraint (month) if exists
	_, _ = s.DB.Exec(ctx, `ALTER TABLE salary_masters DROP CONSTRAINT IF EXISTS salary_masters_month_key`)
	return nil
}

// SeedDefaults inserts the default admin account and its default categories/budget rules.
func (s *Store) SeedDefaults(ctx context.Context) error {
	// Upsert admin account
	var adminID int64
	err := s.DB.QueryRow(ctx, `
		INSERT INTO accounts (full_name, email, password_hash, role)
		VALUES ('Rasa Saufar', $1, $2, 'admin')
		ON CONFLICT (email) DO UPDATE SET role = 'admin'
		RETURNING id`,
		types.HardcodedEmail,
		types.HardcodedPassword,
	).Scan(&adminID)
	if err != nil {
		return err
	}

	// Migration: assign orphaned data (account_id IS NULL) to admin
	for _, table := range []string{"categories", "budget_rules", "transactions", "salary_masters"} {
		if _, err := s.DB.Exec(ctx,
			`UPDATE `+table+` SET account_id = $1 WHERE account_id IS NULL`,
			adminID,
		); err != nil {
			return err
		}
	}

	// Seed default categories for admin
	defaultCategories := []string{"Makan", "Bensin", "Transport", "Belanja", "Hiburan", "Tagihan", "Gaji"}
	for _, name := range defaultCategories {
		if _, err := s.DB.Exec(ctx, `
			INSERT INTO categories (account_id, name) VALUES ($1, $2)
			ON CONFLICT DO NOTHING`,
			adminID, name,
		); err != nil {
			return err
		}
	}

	// Seed default budget rules for admin
	if _, err := s.DB.Exec(ctx, `
		INSERT INTO budget_rules (account_id, category_id, period, limit_amount)
		SELECT $1, id, 'daily', 60000
		FROM categories
		WHERE account_id = $1 AND name = 'Makan'
		ON CONFLICT DO NOTHING`,
		adminID,
	); err != nil {
		return err
	}

	if _, err := s.DB.Exec(ctx, `
		INSERT INTO budget_rules (account_id, category_id, period, limit_amount)
		SELECT $1, id, 'monthly', 240000
		FROM categories
		WHERE account_id = $1 AND name = 'Bensin'
		ON CONFLICT DO NOTHING`,
		adminID,
	); err != nil {
		return err
	}

	return nil
}

// --- Auth / Sessions ---

// FindAccountByEmail returns account info for login.
func (s *Store) FindAccountByEmail(ctx context.Context, email string) (types.DBAccount, error) {
	item := types.DBAccount{}
	err := s.DB.QueryRow(ctx, `
		SELECT id, full_name, email, password_hash, role
		FROM accounts
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1`,
		strings.TrimSpace(email),
	).Scan(&item.ID, &item.FullName, &item.Email, &item.PasswordHash, &item.Role)
	if err != nil {
		return types.DBAccount{}, err
	}
	return item, nil
}

// CreateSession stores a new session token for an account.
func (s *Store) CreateSession(ctx context.Context, token string, accountID int64) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO sessions (token, account_id) VALUES ($1, $2)`,
		token, accountID,
	)
	return err
}

// FindAccountByToken returns the account associated with a session token.
func (s *Store) FindAccountByToken(ctx context.Context, token string) (types.DBAccount, error) {
	item := types.DBAccount{}
	err := s.DB.QueryRow(ctx, `
		SELECT a.id, a.full_name, a.email, a.password_hash, a.role
		FROM sessions s
		JOIN accounts a ON a.id = s.account_id
		WHERE s.token = $1
		LIMIT 1`,
		token,
	).Scan(&item.ID, &item.FullName, &item.Email, &item.PasswordHash, &item.Role)
	if err != nil {
		return types.DBAccount{}, err
	}
	return item, nil
}

// DeleteSession removes a session token (logout).
func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

// --- Admin: Account Management ---

func (s *Store) ListAccounts(ctx context.Context) ([]types.DBAccount, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, full_name, email, role
		FROM accounts
		ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.DBAccount, 0)
	for rows.Next() {
		item := types.DBAccount{}
		if err := rows.Scan(&item.ID, &item.FullName, &item.Email, &item.Role); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateAccount(ctx context.Context, name, email, passwordHash, role string) (types.DBAccount, error) {
	item := types.DBAccount{}
	err := s.DB.QueryRow(ctx, `
		INSERT INTO accounts (full_name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, full_name, email, role`,
		name, email, passwordHash, role,
	).Scan(&item.ID, &item.FullName, &item.Email, &item.Role)
	if err != nil {
		return types.DBAccount{}, err
	}
	return item, nil
}

func (s *Store) UpdateAccount(ctx context.Context, id int64, name, email, passwordHash, role string) (types.DBAccount, error) {
	item := types.DBAccount{}
	var err error
	if passwordHash != "" {
		err = s.DB.QueryRow(ctx, `
			UPDATE accounts
			SET full_name = $1, email = $2, password_hash = $3, role = $4, updated_at = NOW()
			WHERE id = $5
			RETURNING id, full_name, email, role`,
			name, email, passwordHash, role, id,
		).Scan(&item.ID, &item.FullName, &item.Email, &item.Role)
	} else {
		err = s.DB.QueryRow(ctx, `
			UPDATE accounts
			SET full_name = $1, email = $2, role = $3, updated_at = NOW()
			WHERE id = $4
			RETURNING id, full_name, email, role`,
			name, email, role, id,
		).Scan(&item.ID, &item.FullName, &item.Email, &item.Role)
	}
	if err != nil {
		return types.DBAccount{}, err
	}
	return item, nil
}

func (s *Store) DeleteAccount(ctx context.Context, id int64) error {
	result, err := s.DB.Exec(ctx, `DELETE FROM accounts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("akun tidak ditemukan")
	}
	return nil
}

// SeedDefaultCategoriesForAccount seeds default categories for a newly created account.
func (s *Store) SeedDefaultCategoriesForAccount(ctx context.Context, accountID int64) error {
	defaultCategories := []string{"Makan", "Bensin", "Transport", "Belanja", "Hiburan", "Tagihan", "Gaji"}
	for _, name := range defaultCategories {
		if _, err := s.DB.Exec(ctx, `
			INSERT INTO categories (account_id, name) VALUES ($1, $2)
			ON CONFLICT DO NOTHING`,
			accountID, name,
		); err != nil {
			return err
		}
	}

	// Default budget rules
	if _, err := s.DB.Exec(ctx, `
		INSERT INTO budget_rules (account_id, category_id, period, limit_amount)
		SELECT $1, id, 'daily', 60000
		FROM categories WHERE account_id = $1 AND name = 'Makan'
		ON CONFLICT DO NOTHING`,
		accountID,
	); err != nil {
		return err
	}

	if _, err := s.DB.Exec(ctx, `
		INSERT INTO budget_rules (account_id, category_id, period, limit_amount)
		SELECT $1, id, 'monthly', 240000
		FROM categories WHERE account_id = $1 AND name = 'Bensin'
		ON CONFLICT DO NOTHING`,
		accountID,
	); err != nil {
		return err
	}

	return nil
}

// --- User Profile (self-update) ---

func (s *Store) FindAccountByID(ctx context.Context, id int64) (types.DBAccount, error) {
	item := types.DBAccount{}
	err := s.DB.QueryRow(ctx, `
		SELECT id, full_name, email, password_hash, role
		FROM accounts WHERE id = $1`,
		id,
	).Scan(&item.ID, &item.FullName, &item.Email, &item.PasswordHash, &item.Role)
	if err != nil {
		return types.DBAccount{}, err
	}
	return item, nil
}

func (s *Store) UpdateSelfProfile(ctx context.Context, id int64, name, email, passwordHash string) (types.DBAccount, error) {
	item := types.DBAccount{}
	var err error
	if passwordHash != "" {
		err = s.DB.QueryRow(ctx, `
			UPDATE accounts
			SET full_name = $1, email = $2, password_hash = $3, updated_at = NOW()
			WHERE id = $4
			RETURNING id, full_name, email, role`,
			name, email, passwordHash, id,
		).Scan(&item.ID, &item.FullName, &item.Email, &item.Role)
	} else {
		err = s.DB.QueryRow(ctx, `
			UPDATE accounts
			SET full_name = $1, email = $2, updated_at = NOW()
			WHERE id = $3
			RETURNING id, full_name, email, role`,
			name, email, id,
		).Scan(&item.ID, &item.FullName, &item.Email, &item.Role)
	}
	if err != nil {
		return types.DBAccount{}, err
	}
	return item, nil
}

// --- Transactions ---

func (s *Store) ListTransactions(ctx context.Context, accountID int64) ([]types.DBTransaction, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			t.id,
			t.type,
			c.id,
			c.name,
			t.amount,
			t.trx_date,
			t.note
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.account_id = $1
		ORDER BY t.trx_date DESC, t.id DESC`,
		accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.DBTransaction, 0)
	for rows.Next() {
		item := types.DBTransaction{}
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.CategoryID,
			&item.Category,
			&item.Amount,
			&item.Date,
			&item.Note,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// --- Categories ---

func (s *Store) ListCategories(ctx context.Context, accountID int64) ([]types.Category, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name FROM categories
		WHERE account_id = $1
		ORDER BY name ASC`,
		accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.Category, 0)
	for rows.Next() {
		item := types.Category{}
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *Store) FindCategoryByName(ctx context.Context, accountID int64, name string) (types.Category, error) {
	item := types.Category{}
	err := s.DB.QueryRow(ctx, `
		SELECT id, name FROM categories
		WHERE account_id = $1 AND LOWER(name) = LOWER($2)
		ORDER BY id LIMIT 1`,
		accountID, strings.TrimSpace(name),
	).Scan(&item.ID, &item.Name)
	if err != nil {
		return types.Category{}, err
	}
	return item, nil
}

func (s *Store) FindCategoryByID(ctx context.Context, accountID int64, id int64) (types.Category, error) {
	item := types.Category{}
	err := s.DB.QueryRow(ctx, `
		SELECT id, name FROM categories
		WHERE account_id = $1 AND id = $2`,
		accountID, id,
	).Scan(&item.ID, &item.Name)
	if err != nil {
		return types.Category{}, err
	}
	return item, nil
}

func (s *Store) CategoryNameExists(ctx context.Context, accountID int64, name string, exceptID int64) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM categories
			WHERE account_id = $1
			  AND LOWER(name) = LOWER($2)
			  AND ($3 = 0 OR id <> $3)
		)`,
		accountID, strings.TrimSpace(name), exceptID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// --- Budget Rules ---

func (s *Store) ListBudgetRules(ctx context.Context, accountID int64) ([]types.DBBudgetRule, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			br.id,
			c.id,
			c.name,
			br.period,
			br.limit_amount
		FROM budget_rules br
		JOIN categories c ON c.id = br.category_id
		WHERE br.account_id = $1
		ORDER BY c.name ASC, br.period ASC`,
		accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.DBBudgetRule, 0)
	for rows.Next() {
		item := types.DBBudgetRule{}
		if err := rows.Scan(
			&item.ID,
			&item.CategoryID,
			&item.Category,
			&item.Period,
			&item.Limit,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *Store) ListBudgetRulesByCategoryID(ctx context.Context, accountID int64, categoryID int64) ([]types.DBBudgetRule, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			br.id,
			c.id,
			c.name,
			br.period,
			br.limit_amount
		FROM budget_rules br
		JOIN categories c ON c.id = br.category_id
		WHERE br.account_id = $1 AND br.category_id = $2`,
		accountID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.DBBudgetRule, 0)
	for rows.Next() {
		item := types.DBBudgetRule{}
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Category, &item.Period, &item.Limit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *Store) SumExpenseForRule(ctx context.Context, accountID int64, categoryID int64, period string, referenceDate time.Time, excludeTransactionID int64) (int64, error) {
	var query string
	switch period {
	case "daily":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE account_id = $1
			  AND type = 'expense'
			  AND category_id = $2
			  AND ($3 = 0 OR id <> $3)
			  AND trx_date = $4::date`
	case "weekly":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE account_id = $1
			  AND type = 'expense'
			  AND category_id = $2
			  AND ($3 = 0 OR id <> $3)
			  AND DATE_TRUNC('week', trx_date) = DATE_TRUNC('week', $4::date)`
	case "monthly":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE account_id = $1
			  AND type = 'expense'
			  AND category_id = $2
			  AND ($3 = 0 OR id <> $3)
			  AND DATE_TRUNC('month', trx_date) = DATE_TRUNC('month', $4::date)`
	default:
		return 0, errors.New("periode tidak valid")
	}

	var total int64
	err := s.DB.QueryRow(ctx, query, accountID, categoryID, excludeTransactionID, referenceDate).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Store) CheckBudgetForTransaction(ctx context.Context, accountID int64, categoryID int64, trxType string, referenceDate time.Time, additionalAmount int64, excludeTransactionID int64) ([]types.BudgetCheck, error) {
	if trxType != "expense" {
		return []types.BudgetCheck{}, nil
	}

	rules, err := s.ListBudgetRulesByCategoryID(ctx, accountID, categoryID)
	if err != nil {
		return nil, err
	}

	checks := make([]types.BudgetCheck, 0, len(rules))
	for _, rule := range rules {
		used, err := s.SumExpenseForRule(ctx, accountID, categoryID, rule.Period, referenceDate, excludeTransactionID)
		if err != nil {
			return nil, err
		}
		used += additionalAmount
		checks = append(checks, BuildBudgetCheck(rule.Category, rule.Period, rule.Limit, used))
	}

	sort.Slice(checks, func(i, j int) bool {
		return PeriodRank(checks[i].Period) < PeriodRank(checks[j].Period)
	})

	return checks, nil
}

// --- Salary Masters ---

func (s *Store) ListSalaryMasters(ctx context.Context, accountID int64) ([]types.SalaryMaster, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, month, amount, note
		FROM salary_masters
		WHERE account_id = $1
		ORDER BY month DESC, id DESC`,
		accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.SalaryMaster, 0)
	for rows.Next() {
		item := types.SalaryMaster{}
		if err := rows.Scan(&item.ID, &item.Month, &item.Amount, &item.Note); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *Store) SalaryForMonth(ctx context.Context, accountID int64, month string) (int64, error) {
	var total int64
	err := s.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM salary_masters
		WHERE account_id = $1 AND month = $2`,
		accountID, month,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Store) SalaryToMonth(ctx context.Context, accountID int64, month string) (int64, error) {
	var total int64
	err := s.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM salary_masters
		WHERE account_id = $1 AND month <= $2`,
		accountID, month,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// --- Helpers ---

func BuildBudgetCheck(category, period string, limit, used int64) types.BudgetCheck {
	remaining := limit - used
	percentage := 0.0
	if limit > 0 {
		percentage = (float64(used) / float64(limit)) * 100
	}

	return types.BudgetCheck{
		Category:   category,
		Period:     period,
		Limit:      limit,
		Used:       used,
		Remaining:  remaining,
		Percentage: percentage,
		OverBudget: used > limit,
	}
}

func PeriodRank(period string) int {
	switch period {
	case "daily":
		return 1
	case "weekly":
		return 2
	case "monthly":
		return 3
	default:
		return 99
	}
}

// --- Wedding Savings ---

func (s *Store) migrateWeddingSavings(ctx context.Context) error {
	if _, err := s.DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wedding_config (
			account_id BIGINT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
			target_amount BIGINT NOT NULL DEFAULT 50000000,
			target_date DATE,
			bride_name TEXT NOT NULL DEFAULT '',
			groom_name TEXT NOT NULL DEFAULT '',
			venue TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return err
	}

	if _, err := s.DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wedding_deposits (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			deposit_date DATE NOT NULL,
			amount BIGINT NOT NULL CHECK (amount > 0),
			note TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL DEFAULT 'self' CHECK (source IN ('self', 'partner', 'gift', 'other')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return err
	}

	if _, err := s.DB.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_wedding_deposits_account ON wedding_deposits (account_id, deposit_date DESC)`); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetWeddingConfig(ctx context.Context, accountID int64) (types.WeddingConfig, error) {
	cfg := types.WeddingConfig{TargetAmount: 50000000}
	var targetDate *time.Time
	err := s.DB.QueryRow(ctx, `
		SELECT target_amount, target_date, bride_name, groom_name, venue
		FROM wedding_config
		WHERE account_id = $1`,
		accountID,
	).Scan(&cfg.TargetAmount, &targetDate, &cfg.BrideName, &cfg.GroomName, &cfg.Venue)
	if err != nil {
		// Return defaults if no config exists
		return types.WeddingConfig{TargetAmount: 50000000}, nil
	}
	if targetDate != nil {
		cfg.TargetDate = targetDate.Format(types.DateLayout)
	}
	return cfg, nil
}

func (s *Store) UpsertWeddingConfig(ctx context.Context, accountID int64, cfg types.WeddingConfig) (types.WeddingConfig, error) {
	var targetDate *time.Time
	if cfg.TargetDate != "" {
		t, _ := time.Parse(types.DateLayout, cfg.TargetDate)
		targetDate = &t
	}

	_, err := s.DB.Exec(ctx, `
		INSERT INTO wedding_config (account_id, target_amount, target_date, bride_name, groom_name, venue, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (account_id) DO UPDATE SET
			target_amount = EXCLUDED.target_amount,
			target_date = EXCLUDED.target_date,
			bride_name = EXCLUDED.bride_name,
			groom_name = EXCLUDED.groom_name,
			venue = EXCLUDED.venue,
			updated_at = NOW()`,
		accountID, cfg.TargetAmount, targetDate, cfg.BrideName, cfg.GroomName, cfg.Venue,
	)
	if err != nil {
		return types.WeddingConfig{}, err
	}
	return cfg, nil
}

func (s *Store) ListWeddingDeposits(ctx context.Context, accountID int64) ([]types.WeddingDeposit, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, deposit_date, amount, note, source
		FROM wedding_deposits
		WHERE account_id = $1
		ORDER BY deposit_date DESC, id DESC`,
		accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.WeddingDeposit, 0)
	for rows.Next() {
		item := types.WeddingDeposit{}
		var d time.Time
		if err := rows.Scan(&item.ID, &d, &item.Amount, &item.Note, &item.Source); err != nil {
			return nil, err
		}
		item.Date = d.Format(types.DateLayout)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateWeddingDeposit(ctx context.Context, accountID int64, dep types.WeddingDeposit) (types.WeddingDeposit, error) {
	d, _ := time.Parse(types.DateLayout, dep.Date)
	created := types.WeddingDeposit{}
	var dateOut time.Time
	err := s.DB.QueryRow(ctx, `
		INSERT INTO wedding_deposits (account_id, deposit_date, amount, note, source)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, deposit_date, amount, note, source`,
		accountID, d, dep.Amount, dep.Note, dep.Source,
	).Scan(&created.ID, &dateOut, &created.Amount, &created.Note, &created.Source)
	if err != nil {
		return types.WeddingDeposit{}, err
	}
	created.Date = dateOut.Format(types.DateLayout)
	return created, nil
}

func (s *Store) UpdateWeddingDeposit(ctx context.Context, accountID int64, id int64, dep types.WeddingDeposit) (types.WeddingDeposit, error) {
	d, _ := time.Parse(types.DateLayout, dep.Date)
	updated := types.WeddingDeposit{}
	var dateOut time.Time
	err := s.DB.QueryRow(ctx, `
		UPDATE wedding_deposits
		SET deposit_date = $1, amount = $2, note = $3, source = $4, updated_at = NOW()
		WHERE id = $5 AND account_id = $6
		RETURNING id, deposit_date, amount, note, source`,
		d, dep.Amount, dep.Note, dep.Source, id, accountID,
	).Scan(&updated.ID, &dateOut, &updated.Amount, &updated.Note, &updated.Source)
	if err != nil {
		return types.WeddingDeposit{}, err
	}
	updated.Date = dateOut.Format(types.DateLayout)
	return updated, nil
}

func (s *Store) DeleteWeddingDeposit(ctx context.Context, accountID int64, id int64) error {
	result, err := s.DB.Exec(ctx, `
		DELETE FROM wedding_deposits WHERE id = $1 AND account_id = $2`,
		id, accountID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("setoran tidak ditemukan")
	}
	return nil
}
