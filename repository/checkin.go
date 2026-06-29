package repository

import (
	"context"
	"time"
)

func (r *Repository) HeatmapCounts(ctx context.Context, userID, itemID int64, from string) (map[string]int, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT checkin_date, COALESCE(SUM(count), 0)::INT
		FROM checkins
		WHERE user_id = $1 AND item_id = $2 AND checkin_date >= $3
		GROUP BY checkin_date
	`, userID, itemID, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			return nil, err
		}
		counts[date] = count
	}
	return counts, rows.Err()
}

func (r *Repository) CountForDate(ctx context.Context, userID, itemID int64, date string) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(count), 0)::INT
		FROM checkins
		WHERE user_id = $1 AND item_id = $2 AND checkin_date = $3
	`, userID, itemID, date).Scan(&count)
	return count, err
}

func (r *Repository) CountMakeupsCreatedBetween(ctx context.Context, userID, itemID int64, from time.Time, to time.Time) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx, `
		SELECT COUNT(*)::INT
		FROM checkins
		WHERE user_id = $1 AND item_id = $2 AND source = 'makeup' AND created_at >= $3 AND created_at < $4
	`, userID, itemID, from, to).Scan(&count)
	return count, err
}
