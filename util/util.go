package util

import (
	"strconv"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"
const ClockLayout = "15:04"

func LevelForCount(count, target int) int {
	if count <= 0 {
		return 0
	}
	if target < 1 {
		target = 1
	}
	level := (count*4 + target - 1) / target
	if level < 1 {
		return 1
	}
	if level > 4 {
		return 4
	}
	return level
}

func NormalizeDate(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if _, err := time.Parse(DateLayout, value); err != nil {
		return fallback
	}
	return value
}

func NormalizeClock(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) >= 5 {
		value = value[:5]
	}
	if _, err := time.Parse(ClockLayout, value); err != nil {
		return ""
	}
	return value
}

func NormalizeTimeMode(value string) string {
	if strings.TrimSpace(value) == "time_range" {
		return "time_range"
	}
	return "all_day"
}

func ParseDateOr(value string, fallback time.Time) time.Time {
	parsed, err := time.ParseInLocation(DateLayout, value, time.Local)
	if err != nil {
		return fallback
	}
	return parsed
}

func DateBefore(a, b string) bool {
	aa, errA := time.Parse(DateLayout, a)
	bb, errB := time.Parse(DateLayout, b)
	if errA != nil || errB != nil {
		return false
	}
	return aa.Before(bb)
}

func DaysBetween(start, end time.Time) int {
	start = DateOnly(start)
	end = DateOnly(end)
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func DateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.Local)
}

func StringValue(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return strings.TrimSpace(*value)
}

func IntValue(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func Int64Value(value *int64, fallback int64) int64 {
	if value == nil {
		return fallback
	}
	return *value
}

func BoolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func MakeSlug(name string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			builder.WriteRune(r)
			continue
		}
		if builder.Len() > 0 {
			builder.WriteByte('-')
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		slug = "category"
	}
	return slug + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
