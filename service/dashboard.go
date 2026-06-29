package service

import (
	"context"
	"time"

	"lumalog-backend/model"
	"lumalog-backend/util"
)

func (s *Service) BuildDashboardItem(ctx context.Context, userID int64, it model.Item) (model.DashboardItem, error) {
	from := time.Now().AddDate(-1, 0, 1).Format(util.DateLayout)
	counts, err := s.Repo.HeatmapCounts(ctx, userID, it.ID, from)
	if err != nil {
		return model.DashboardItem{}, err
	}
	heatmap := BuildHeatmap(counts, it.DailyTargetCount)
	stats := ComputeStats(it, heatmap)
	today := time.Now().Format(util.DateLayout)
	todayCount := counts[today]
	return model.DashboardItem{
		Item:       it,
		Stats:      stats,
		Heatmap:    heatmap,
		TodayCount: todayCount,
		Status:     CheckinStatus(it, todayCount, time.Now()),
	}, nil
}

func BuildHeatmap(counts map[string]int, target int) []model.HeatmapDay {
	if target < 1 {
		target = 1
	}
	end := time.Now()
	start := end.AddDate(-1, 0, 1)
	total := int(end.Sub(start).Hours()/24) + 1
	days := make([]model.HeatmapDay, 0, total)
	for i := 0; i < total; i++ {
		date := start.AddDate(0, 0, i).Format(util.DateLayout)
		count := counts[date]
		level := util.LevelForCount(count, target)
		days = append(days, model.HeatmapDay{
			Date:      date,
			Count:     count,
			Level:     level,
			Completed: count >= target,
		})
	}
	return days
}

func ComputeStats(it model.Item, days []model.HeatmapDay) model.ItemStats {
	completedSet := map[string]bool{}
	total := 0
	completedDays := 0
	for _, day := range days {
		total += day.Count
		if day.Completed {
			completedDays++
			completedSet[day.Date] = true
		}
	}

	today := time.Now()
	start := util.ParseDateOr(it.StartDate, today)
	end := today
	if !it.IsUnlimited && it.EndDate != "" {
		parsedEnd := util.ParseDateOr(it.EndDate, today)
		if parsedEnd.Before(end) {
			end = parsedEnd
		}
	}
	expected := util.DaysBetween(start, end)
	if expected < 0 {
		expected = 0
	}

	streakStart := today
	if !completedSet[today.Format(util.DateLayout)] {
		streakStart = today.AddDate(0, 0, -1)
	}
	current := 0
	for date := streakStart; ; date = date.AddDate(0, 0, -1) {
		if date.Before(start) || !completedSet[date.Format(util.DateLayout)] {
			break
		}
		current++
	}

	longest := 0
	running := 0
	for _, day := range days {
		if day.Completed {
			running++
			if running > longest {
				longest = running
			}
		} else {
			running = 0
		}
	}

	rate := 0.0
	if expected > 0 {
		rate = float64(completedDays) / float64(expected)
		if rate > 1 {
			rate = 1
		}
	}

	return model.ItemStats{
		CurrentStreak:  current,
		LongestStreak:  longest,
		TotalCheckins:  total,
		CompletedDays:  completedDays,
		ExpectedDays:   expected,
		CompletionRate: rate,
	}
}
