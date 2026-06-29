package service

import (
	"context"

	"lumalog-backend/model"
)

func ItemBadges(stats model.ItemStats) []model.Badge {
	return []model.Badge{
		{
			ID:          "first_light",
			Title:       "初次点亮",
			Description: "完成第一次签到",
			Level:       "bronze",
			Earned:      stats.TotalCheckins >= 1,
		},
		{
			ID:          "week_streak",
			Title:       "七日连光",
			Description: "最长连续签到达到 7 天",
			Level:       "silver",
			Earned:      stats.LongestStreak >= 7,
		},
		{
			ID:          "month_streak",
			Title:       "三十日微光",
			Description: "最长连续签到达到 30 天",
			Level:       "gold",
			Earned:      stats.LongestStreak >= 30,
		},
		{
			ID:          "hundred_lights",
			Title:       "百次记录",
			Description: "累计签到达到 100 次",
			Level:       "gold",
			Earned:      stats.TotalCheckins >= 100,
		},
		{
			ID:          "steady_flow",
			Title:       "稳定节奏",
			Description: "完成率达到 80%",
			Level:       "silver",
			Earned:      stats.ExpectedDays >= 7 && stats.CompletionRate >= 0.8,
		},
	}
}

func (s *Service) UserBadges(ctx context.Context, userID int64) ([]model.Badge, error) {
	items, err := s.Repo.ListItems(ctx, userID, false, false)
	if err != nil {
		return nil, err
	}

	totalCheckins := 0
	maxLongestStreak := 0
	activeCompletedHabits := 0
	for _, it := range items {
		di, err := s.BuildDashboardItem(ctx, userID, it)
		if err != nil {
			return nil, err
		}
		totalCheckins += di.Stats.TotalCheckins
		if di.Stats.LongestStreak > maxLongestStreak {
			maxLongestStreak = di.Stats.LongestStreak
		}
		if di.Stats.CompletedDays > 0 {
			activeCompletedHabits++
		}
	}

	return []model.Badge{
		{
			ID:          "first_habit_light",
			Title:       "第一束光",
			Description: "任意 habit 完成第一次签到",
			Level:       "bronze",
			Earned:      totalCheckins >= 1,
		},
		{
			ID:          "seven_day_runner",
			Title:       "七日同行",
			Description: "任意 habit 最长连续达到 7 天",
			Level:       "silver",
			Earned:      maxLongestStreak >= 7,
		},
		{
			ID:          "thirty_day_runner",
			Title:       "一月成线",
			Description: "任意 habit 最长连续达到 30 天",
			Level:       "gold",
			Earned:      maxLongestStreak >= 30,
		},
		{
			ID:          "three_habits_lit",
			Title:       "三线并进",
			Description: "至少 3 个 habit 有完成记录",
			Level:       "gold",
			Earned:      activeCompletedHabits >= 3,
		},
		{
			ID:          "hundred_total_lights",
			Title:       "百次点亮",
			Description: "所有 habit 累计签到达到 100 次",
			Level:       "gold",
			Earned:      totalCheckins >= 100,
		},
	}, nil
}
