package patch

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"lumalog-backend/repository"
	"lumalog-backend/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Runner struct {
	DB   *pgxpool.Pool
	Repo *repository.Repository
}

type Script func(context.Context, Runner) error

var scripts = map[string]Script{
	"yu-hua-reading-246": yuHuaReading246,
}

func NewRunner(db *pgxpool.Pool, repo *repository.Repository) Runner {
	return Runner{DB: db, Repo: repo}
}

func (r Runner) Run(ctx context.Context, name string) error {
	script, ok := scripts[name]
	if !ok {
		return fmt.Errorf("unknown patch %q, available: %s", name, AvailableNames())
	}
	return script(ctx, r)
}

func AvailableNames() string {
	names := make([]string, 0, len(scripts))
	for name := range scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func yuHuaReading246(ctx context.Context, runner Runner) error {
	const (
		email       = "demo@lumalog.local"
		password    = "123456"
		displayName = "Demo"
		itemName    = "余华阅读"
		checkins    = 246
	)

	tx, err := runner.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userID, err := ensurePatchUser(ctx, tx, email, password, displayName)
	if err != nil {
		return err
	}
	if err := runner.Repo.InsertDefaultCategoriesTx(ctx, tx, userID); err != nil {
		return err
	}

	categoryID, err := findCategoryBySlug(ctx, tx, userID, "reading")
	if err != nil {
		return err
	}

	today := time.Now()
	startDate := today.AddDate(0, 0, -(checkins - 1)).Format(util.DateLayout)
	itemID, err := upsertPatchItem(ctx, tx, userID, categoryID, itemName, startDate)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM checkins WHERE user_id = $1 AND item_id = $2`, userID, itemID); err != nil {
		return err
	}
	for i := 0; i < checkins; i++ {
		date := today.AddDate(0, 0, -(checkins-1)+i).Format(util.DateLayout)
		_, err := tx.Exec(ctx, `
			INSERT INTO checkins (user_id, item_id, checkin_date, checkin_time, count, note, source)
			VALUES ($1, $2, $3, $4, 1, $5, 'normal')
		`, userID, itemID, date, "21:00", "patch: 余华阅读")
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func ensurePatchUser(ctx context.Context, tx pgx.Tx, email, password, displayName string) (int64, error) {
	var userID int64
	err := tx.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	if err == nil {
		_, err = tx.Exec(ctx, `
			UPDATE users
			SET display_name = $1,
				show_current_streak = TRUE,
				show_longest_streak = TRUE,
				show_completion_rate = TRUE,
				show_total_checkins = TRUE,
				updated_at = NOW()
			WHERE id = $2
		`, displayName, userID)
		return userID, err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO users (
			email, password_hash, display_name,
			show_current_streak, show_longest_streak, show_completion_rate, show_total_checkins
		)
		VALUES ($1, $2, $3, TRUE, TRUE, TRUE, TRUE)
		RETURNING id
	`, email, string(hash), displayName).Scan(&userID)
	return userID, err
}

func findCategoryBySlug(ctx context.Context, tx pgx.Tx, userID int64, slug string) (int64, error) {
	var categoryID int64
	err := tx.QueryRow(ctx, `
		SELECT id
		FROM categories
		WHERE user_id = $1 AND slug = $2
	`, userID, slug).Scan(&categoryID)
	return categoryID, err
}

func upsertPatchItem(ctx context.Context, tx pgx.Tx, userID, categoryID int64, name, startDate string) (int64, error) {
	var itemID int64
	err := tx.QueryRow(ctx, `
		SELECT id
		FROM items
		WHERE user_id = $1 AND name = $2 AND deleted_at IS NULL
		ORDER BY id ASC
		LIMIT 1
	`, userID, name).Scan(&itemID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	if err == nil {
		_, err = tx.Exec(ctx, `
			UPDATE items
			SET category_id = $1,
				description = $2,
				color_theme = 'teal',
				start_date = $3,
				end_date = '',
				is_unlimited = TRUE,
				daily_target_count = 1,
				time_mode = 'all_day',
				valid_start_time = '',
				valid_end_time = '',
				allow_extra_checkins = FALSE,
				show_on_dashboard = TRUE,
				updated_at = NOW()
			WHERE id = $4 AND user_id = $5
		`, categoryID, "patch 模拟数据：已连续阅读余华作品 246 天", startDate, itemID, userID)
		return itemID, err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO items (
			user_id, category_id, name, description, color_theme, start_date, end_date,
			is_unlimited, daily_target_count, time_mode, valid_start_time, valid_end_time,
			allow_makeup, makeup_limit_days, allow_extra_checkins, show_on_dashboard, sort_order
		)
		VALUES ($1, $2, $3, $4, 'teal', $5, '', TRUE, 1, 'all_day', '', '', FALSE, 0, FALSE, TRUE, 0)
		RETURNING id
	`, userID, categoryID, name, "patch 模拟数据：已连续阅读余华作品 246 天", startDate).Scan(&itemID)
	return itemID, err
}
