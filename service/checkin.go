package service

import (
	"time"

	"lumalog-backend/model"
	"lumalog-backend/util"
)

func CheckinStatus(it model.Item, todayCount int, now time.Time) string {
	today := now.Format(util.DateLayout)
	if util.DateBefore(today, it.StartDate) {
		return "not_started"
	}
	if !it.IsUnlimited && it.EndDate != "" && util.DateBefore(it.EndDate, today) {
		return "ended"
	}
	if it.TimeMode == "time_range" {
		clock := now.Format(util.ClockLayout)
		if it.ValidStartTime != "" && clock < it.ValidStartTime {
			return "before_time_window"
		}
		if it.ValidEndTime != "" && clock > it.ValidEndTime {
			return "after_time_window"
		}
	}
	if todayCount >= it.DailyTargetCount {
		if it.AllowExtraCheckins {
			return "completed_can_continue"
		}
		return "completed"
	}
	return "available"
}

func StatusMessage(status string) string {
	switch status {
	case "not_started":
		return "该 item 还未开始"
	case "ended":
		return "该 item 已结束"
	case "before_time_window":
		return "还不到签到时间"
	case "after_time_window":
		return "今日签到时间已结束"
	case "completed":
		return "今日已完成签到"
	default:
		return "当前不能签到"
	}
}
