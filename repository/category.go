package repository

import (
	"context"

	"lumalog-backend/model"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) InsertDefaultCategoriesTx(ctx context.Context, tx pgx.Tx, userID int64) error {
	for _, cat := range model.DefaultCategories {
		_, err := tx.Exec(ctx, `
			INSERT INTO categories (user_id, name, slug, color_theme, sort_order, is_default, is_hidden)
			VALUES ($1, $2, $3, $4, $5, TRUE, FALSE)
			ON CONFLICT (user_id, slug) DO NOTHING
		`, userID, cat.Name, cat.Slug, cat.ColorTheme, cat.SortOrder)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) EnsureDefaultCategories(ctx context.Context, userID int64) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := r.InsertDefaultCategoriesTx(ctx, tx, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListCategories(ctx context.Context, userID int64, includeHidden bool) ([]model.Category, error) {
	query := `
		SELECT id, user_id, name, slug, color_theme, sort_order, is_default, is_hidden,
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
			to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM categories
		WHERE user_id = $1
	`
	if !includeHidden {
		query += ` AND is_hidden = FALSE`
	}
	query += ` ORDER BY sort_order ASC, id ASC`

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.Category{}
	for rows.Next() {
		var cat model.Category
		err := rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Slug, &cat.ColorTheme, &cat.SortOrder, &cat.IsDefault, &cat.IsHidden, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, cat)
	}
	return items, rows.Err()
}

func (r *Repository) GetCategory(ctx context.Context, userID, id int64) (model.Category, error) {
	var cat model.Category
	err := r.DB.QueryRow(ctx, `
		SELECT id, user_id, name, slug, color_theme, sort_order, is_default, is_hidden,
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
			to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM categories
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Slug, &cat.ColorTheme, &cat.SortOrder, &cat.IsDefault, &cat.IsHidden, &cat.CreatedAt, &cat.UpdatedAt)
	return cat, err
}

func (r *Repository) FirstVisibleCategory(ctx context.Context, userID int64) (model.Category, error) {
	var cat model.Category
	err := r.DB.QueryRow(ctx, `
		SELECT id, user_id, name, slug, color_theme, sort_order, is_default, is_hidden,
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
			to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM categories
		WHERE user_id = $1 AND is_hidden = FALSE
		ORDER BY sort_order ASC, id ASC
		LIMIT 1
	`, userID).Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Slug, &cat.ColorTheme, &cat.SortOrder, &cat.IsDefault, &cat.IsHidden, &cat.CreatedAt, &cat.UpdatedAt)
	return cat, err
}
