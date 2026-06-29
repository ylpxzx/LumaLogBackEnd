package repository

import (
	"context"

	"lumalog-backend/model"
)

const itemColumns = `
	i.id, i.user_id, i.category_id, COALESCE(c.name, ''),
	i.name, i.description, i.color_theme, i.start_date, i.end_date, i.is_unlimited,
	i.daily_target_count, i.time_mode, i.valid_start_time, i.valid_end_time,
	i.allow_makeup, i.makeup_limit_days, i.allow_extra_checkins, i.show_on_dashboard,
	i.sort_order, COALESCE(to_char(i.archived_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'), ''),
	to_char(i.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
	to_char(i.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
`

func (r *Repository) ListItems(ctx context.Context, userID int64, dashboardOnly bool) ([]model.Item, error) {
	query := `SELECT ` + itemColumns + `
		FROM items i
		LEFT JOIN categories c ON c.id = i.category_id
		WHERE i.user_id = $1 AND i.deleted_at IS NULL AND i.archived_at IS NULL
	`
	if dashboardOnly {
		query += ` AND i.show_on_dashboard = TRUE`
	}
	query += ` ORDER BY i.sort_order ASC, i.created_at DESC`

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *Repository) GetItem(ctx context.Context, userID, id int64) (model.Item, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+itemColumns+`
		FROM items i
		LEFT JOIN categories c ON c.id = i.category_id
		WHERE i.id = $1 AND i.user_id = $2 AND i.deleted_at IS NULL
	`, id, userID)

	var it model.Item
	err := row.Scan(&it.ID, &it.UserID, &it.CategoryID, &it.CategoryName, &it.Name, &it.Description,
		&it.ColorTheme, &it.StartDate, &it.EndDate, &it.IsUnlimited, &it.DailyTargetCount,
		&it.TimeMode, &it.ValidStartTime, &it.ValidEndTime, &it.AllowMakeup, &it.MakeupLimitDays,
		&it.AllowExtraCheckins, &it.ShowOnDashboard, &it.SortOrder, &it.ArchivedAt, &it.CreatedAt, &it.UpdatedAt)
	return it, err
}

type itemScanner interface {
	Scan(dest ...any) error
}

func scanItem(row itemScanner) (model.Item, error) {
	var it model.Item
	err := row.Scan(&it.ID, &it.UserID, &it.CategoryID, &it.CategoryName, &it.Name, &it.Description,
		&it.ColorTheme, &it.StartDate, &it.EndDate, &it.IsUnlimited, &it.DailyTargetCount,
		&it.TimeMode, &it.ValidStartTime, &it.ValidEndTime, &it.AllowMakeup, &it.MakeupLimitDays,
		&it.AllowExtraCheckins, &it.ShowOnDashboard, &it.SortOrder, &it.ArchivedAt, &it.CreatedAt, &it.UpdatedAt)
	return it, err
}
