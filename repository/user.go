package repository

import (
	"context"

	"lumalog-backend/model"
)

func (r *Repository) GetUser(ctx context.Context, userID int64) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow(ctx, `
		SELECT id, email, display_name, theme_preference, language_preference, dashboard_view_mode,
			show_today_status,
			show_current_streak, show_longest_streak, show_completion_rate, show_total_checkins,
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
			to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.ThemePreference,
		&u.LanguagePreference,
		&u.DashboardViewMode,
		&u.ShowTodayStatus,
		&u.ShowCurrentStreak,
		&u.ShowLongestStreak,
		&u.ShowCompletionRate,
		&u.ShowTotalCheckins,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return u, err
}
