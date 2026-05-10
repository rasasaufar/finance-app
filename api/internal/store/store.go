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
func (s *Store) EnsureSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_lower_name ON categories ((LOWER(name)))`,
		`CREATE TABLE IF NOT EXISTS budget_rules (
			id BIGSERIAL PRIMARY KEY,
			category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			period TEXT NOT NULL CHECK (period IN ('daily', 'weekly', 'monthly')),
			limit_amount BIGINT NOT NULL CHECK (limit_amount > 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (category_id, period)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id BIGSERIAL PRIMARY KEY,
			type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
			category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
			amount BIGINT NOT NULL CHECK (amount > 0),
			trx_date DATE NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_trx_date ON transactions (trx_date)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions (category_id)`,
		`CREATE TABLE IF NOT EXISTS salary_masters (
			id BIGSERIAL PRIMARY KEY,
			month TEXT NOT NULL UNIQUE CHECK (month ~ '^[0-9]{4}-[0-9]{2}$'),
			amount BIGINT NOT NULL CHECK (amount > 0),
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
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
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_user_profiles_lower_email ON user_profiles ((LOWER(email)))`,
	}

	for _, stmt := range statements {
		if _, err := s.DB.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}

// SeedDefaults inserts default categories, budget rules, and user profile.
func (s *Store) SeedDefaults(ctx context.Context) error {
	if _, err := s.DB.Exec(ctx, `
		INSERT INTO categories (name) VALUES
		('Makan'),
		('Bensin'),
		('Transport'),
		('Belanja'),
		('Hiburan'),
		('Tagihan'),
		('Gaji')
		ON CONFLICT (name) DO NOTHING`); err != nil {
		return err
	}

	if _, err := s.DB.Exec(ctx, `
		INSERT INTO budget_rules (category_id, period, limit_amount)
		SELECT id, 'daily', 60000
		FROM categories
		WHERE name = 'Makan'
		ON CONFLICT (category_id, period) DO NOTHING`); err != nil {
		return err
	}

	if _, err := s.DB.Exec(ctx, `
		INSERT INTO budget_rules (category_id, period, limit_amount)
		SELECT id, 'monthly', 240000
		FROM categories
		WHERE name = 'Bensin'
		ON CONFLICT (category_id, period) DO NOTHING`); err != nil {
		return err
	}

	if _, err := s.DB.Exec(ctx, `
		INSERT INTO user_profiles (full_name, email, password_hash, avatar_initials)
		VALUES ('Rasa Saufar', $1, $2, 'RS')
		ON CONFLICT (email) DO NOTHING`,
		types.HardcodedEmail,
		types.HardcodedPassword,
	); err != nil {
		return err
	}

	return nil
}

// --- Transactions ---

func (s *Store) ListTransactions(ctx context.Context) ([]types.DBTransaction, error) {
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
		ORDER BY t.trx_date DESC, t.id DESC`)
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

func (s *Store) ListCategories(ctx context.Context) ([]types.Category, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, name FROM categories ORDER BY name ASC`)
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

func (s *Store) FindCategoryByName(ctx context.Context, name string) (types.Category, error) {
	item := types.Category{}
	err := s.DB.QueryRow(
		ctx,
		`SELECT id, name FROM categories WHERE LOWER(name) = LOWER($1) ORDER BY id LIMIT 1`,
		strings.TrimSpace(name),
	).Scan(&item.ID, &item.Name)
	if err != nil {
		return types.Category{}, err
	}
	return item, nil
}

func (s *Store) FindCategoryByID(ctx context.Context, id int64) (types.Category, error) {
	item := types.Category{}
	err := s.DB.QueryRow(ctx, `SELECT id, name FROM categories WHERE id = $1`, id).Scan(&item.ID, &item.Name)
	if err != nil {
		return types.Category{}, err
	}
	return item, nil
}

func (s *Store) CategoryNameExists(ctx context.Context, name string, exceptID int64) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1 FROM categories
			WHERE LOWER(name) = LOWER($1)
			  AND ($2 = 0 OR id <> $2)
		)`,
		strings.TrimSpace(name),
		exceptID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// --- Budget Rules ---

func (s *Store) ListBudgetRules(ctx context.Context) ([]types.DBBudgetRule, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			br.id,
			c.id,
			c.name,
			br.period,
			br.limit_amount
		FROM budget_rules br
		JOIN categories c ON c.id = br.category_id
		ORDER BY c.name ASC, br.period ASC`)
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

func (s *Store) ListBudgetRulesByCategoryID(ctx context.Context, categoryID int64) ([]types.DBBudgetRule, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			br.id,
			c.id,
			c.name,
			br.period,
			br.limit_amount
		FROM budget_rules br
		JOIN categories c ON c.id = br.category_id
		WHERE br.category_id = $1`, categoryID)
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

func (s *Store) SumExpenseForRule(ctx context.Context, categoryID int64, period string, referenceDate time.Time, excludeTransactionID int64) (int64, error) {
	var query string
	switch period {
	case "daily":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE type = 'expense'
			  AND category_id = $1
			  AND ($2 = 0 OR id <> $2)
			  AND trx_date = $3::date`
	case "weekly":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE type = 'expense'
			  AND category_id = $1
			  AND ($2 = 0 OR id <> $2)
			  AND DATE_TRUNC('week', trx_date) = DATE_TRUNC('week', $3::date)`
	case "monthly":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE type = 'expense'
			  AND category_id = $1
			  AND ($2 = 0 OR id <> $2)
			  AND DATE_TRUNC('month', trx_date) = DATE_TRUNC('month', $3::date)`
	default:
		return 0, errors.New("periode tidak valid")
	}

	var total int64
	err := s.DB.QueryRow(ctx, query, categoryID, excludeTransactionID, referenceDate).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Store) CheckBudgetForTransaction(ctx context.Context, categoryID int64, trxType string, referenceDate time.Time, additionalAmount int64, excludeTransactionID int64) ([]types.BudgetCheck, error) {
	if trxType != "expense" {
		return []types.BudgetCheck{}, nil
	}

	rules, err := s.ListBudgetRulesByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	checks := make([]types.BudgetCheck, 0, len(rules))
	for _, rule := range rules {
		used, err := s.SumExpenseForRule(ctx, categoryID, rule.Period, referenceDate, excludeTransactionID)
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

func (s *Store) ListSalaryMasters(ctx context.Context) ([]types.SalaryMaster, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, month, amount, note
		FROM salary_masters
		ORDER BY month DESC, id DESC`)
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

func (s *Store) SalaryForMonth(ctx context.Context, month string) (int64, error) {
	var total int64
	err := s.DB.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM salary_masters WHERE month = $1`,
		month,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Store) SalaryToMonth(ctx context.Context, month string) (int64, error) {
	var total int64
	err := s.DB.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM salary_masters WHERE month <= $1`,
		month,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// --- User Profile ---

func (s *Store) FindUserProfileByEmail(ctx context.Context, email string) (types.DBUserProfile, error) {
	item := types.DBUserProfile{}
	err := s.DB.QueryRow(
		ctx,
		`SELECT id, full_name, email
		 FROM user_profiles
		 WHERE LOWER(email) = LOWER($1)
		 ORDER BY id
		 LIMIT 1`,
		strings.TrimSpace(email),
	).Scan(&item.ID, &item.FullName, &item.Email)
	if err != nil {
		return types.DBUserProfile{}, err
	}
	return item, nil
}

func (s *Store) UpdateUserProfile(ctx context.Context, currentEmail, newName, newEmail, newPassword string) (types.DBUserProfile, error) {
	var item types.DBUserProfile

	if newPassword != "" {
		err := s.DB.QueryRow(
			ctx,
			`UPDATE user_profiles
			 SET full_name = $1,
			     email = $2,
			     password_hash = $3,
			     updated_at = NOW()
			 WHERE LOWER(email) = LOWER($4)
			 RETURNING id, full_name, email`,
			newName, newEmail, newPassword, currentEmail,
		).Scan(&item.ID, &item.FullName, &item.Email)
		if err != nil {
			return types.DBUserProfile{}, err
		}
	} else {
		err := s.DB.QueryRow(
			ctx,
			`UPDATE user_profiles
			 SET full_name = $1,
			     email = $2,
			     updated_at = NOW()
			 WHERE LOWER(email) = LOWER($3)
			 RETURNING id, full_name, email`,
			newName, newEmail, currentEmail,
		).Scan(&item.ID, &item.FullName, &item.Email)
		if err != nil {
			return types.DBUserProfile{}, err
		}
	}

	return item, nil
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
