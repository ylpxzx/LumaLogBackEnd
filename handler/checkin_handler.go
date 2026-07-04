package handler

import (
	"net/http"
	"strings"
	"time"

	"lumalog-backend/model"
	"lumalog-backend/service"
	"lumalog-backend/util"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCheckins(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	userID := currentUserID(c)
	if _, err := h.Repo.GetItem(c.Request.Context(), userID, id); err != nil {
		jsonError(c, http.StatusNotFound, "item 不存在")
		return
	}

	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT id, user_id, item_id, checkin_date, checkin_time, count, note, source,
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM checkins
		WHERE user_id = $1 AND item_id = $2
		ORDER BY checkin_date DESC, id DESC
	`, userID, id)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取签到记录失败")
		return
	}
	defer rows.Close()

	items := []model.Checkin{}
	for rows.Next() {
		var ck model.Checkin
		if err := rows.Scan(&ck.ID, &ck.UserID, &ck.ItemID, &ck.CheckinDate, &ck.CheckinTime, &ck.Count, &ck.Note, &ck.Source, &ck.CreatedAt); err != nil {
			jsonError(c, http.StatusInternalServerError, "读取签到记录失败")
			return
		}
		items = append(items, ck)
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateCheckin(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req model.CheckinRequest
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}

	userID := currentUserID(c)
	it, err := h.Repo.GetItem(c.Request.Context(), userID, id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "item 不存在")
		return
	}

	now := time.Now()
	today := now.Format(util.DateLayout)
	checkinDate := util.NormalizeDate(util.StringValue(req.CheckinDate, today), today)
	isMakeup := checkinDate != today
	dateCount, err := h.Repo.CountForDate(c.Request.Context(), userID, id, checkinDate)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取签到进度失败")
		return
	}

	count := util.IntValue(req.Count, 1)
	if count < 1 {
		count = 1
	}
	note := strings.TrimSpace(util.StringValue(req.Note, ""))
	source := strings.TrimSpace(util.StringValue(req.Source, "normal"))
	if source != "normal" && source != "makeup" {
		source = "normal"
	}
	if isMakeup {
		source = "makeup"
		if it.ArchivedAt != "" {
			jsonError(c, http.StatusBadRequest, "已归档 habit 不能补签")
			return
		}
		parsedDate := util.ParseDateOr(checkinDate, now)
		todayDate := util.DateOnly(now)
		currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		if parsedDate.After(todayDate) {
			jsonError(c, http.StatusBadRequest, "不能补签未来日期")
			return
		}
		if parsedDate.Before(currentMonthStart) {
			jsonError(c, http.StatusBadRequest, "只能补签本月日期")
			return
		}
		if util.DateBefore(checkinDate, it.StartDate) {
			jsonError(c, http.StatusBadRequest, "不能补签开始日期之前")
			return
		}
		if !it.IsUnlimited && it.EndDate != "" && util.DateBefore(it.EndDate, checkinDate) {
			jsonError(c, http.StatusBadRequest, "不能补签结束日期之后")
			return
		}
		if !it.AllowMakeup {
			jsonError(c, http.StatusBadRequest, "该 habit 未开启补签")
			return
		}
		if dateCount >= it.DailyTargetCount && !it.AllowExtraCheckins {
			jsonError(c, http.StatusBadRequest, "该日期已完成签到")
			return
		}
		if it.MakeupMonthlyLimit > 0 {
			monthEnd := currentMonthStart.AddDate(0, 1, 0)
			used, err := h.Repo.CountMakeupsCreatedBetween(c.Request.Context(), userID, id, currentMonthStart, monthEnd)
			if err != nil {
				jsonError(c, http.StatusInternalServerError, "读取补签次数失败")
				return
			}
			if used >= it.MakeupMonthlyLimit {
				jsonError(c, http.StatusBadRequest, "本月补签次数已用完")
				return
			}
		}
	} else {
		source = "normal"
		status := service.CheckinStatus(it, dateCount, now)
		if status != "available" && status != "completed_can_continue" {
			jsonError(c, http.StatusBadRequest, service.StatusMessage(status))
			return
		}
	}

	if req.Note != nil {
		_, err = h.DB.Exec(c.Request.Context(), `
			UPDATE checkins
			SET note = $1
			WHERE user_id = $2 AND item_id = $3 AND checkin_date = $4
		`, note, userID, id, checkinDate)
		if err != nil {
			jsonError(c, http.StatusInternalServerError, "更新签到备注失败")
			return
		}
	}

	_, err = h.DB.Exec(c.Request.Context(), `
		INSERT INTO checkins (user_id, item_id, checkin_date, checkin_time, count, note, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, id, checkinDate, now.Format(util.ClockLayout), count, note, source)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "签到失败")
		return
	}

	di, err := h.Service.BuildDashboardItem(c.Request.Context(), userID, it)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "刷新签到状态失败")
		return
	}
	c.JSON(http.StatusCreated, di)
}

func (h *Handler) DeleteCheckin(c *gin.Context) {
	itemID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	checkinID, ok := parseIDParam(c, "checkinId")
	if !ok {
		return
	}
	tag, err := h.DB.Exec(c.Request.Context(), `
		DELETE FROM checkins
		WHERE id = $1 AND item_id = $2 AND user_id = $3
	`, checkinID, itemID, currentUserID(c))
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "删除签到记录失败")
		return
	}
	if tag.RowsAffected() == 0 {
		jsonError(c, http.StatusNotFound, "签到记录不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
