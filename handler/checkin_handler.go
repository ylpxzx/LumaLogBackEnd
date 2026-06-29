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
	todayCount, err := h.Repo.CountForDate(c.Request.Context(), userID, id, today)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取今日进度失败")
		return
	}
	status := service.CheckinStatus(it, todayCount, now)
	if status != "available" && status != "completed_can_continue" {
		jsonError(c, http.StatusBadRequest, service.StatusMessage(status))
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

	_, err = h.DB.Exec(c.Request.Context(), `
		INSERT INTO checkins (user_id, item_id, checkin_date, checkin_time, count, note, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, id, today, now.Format(util.ClockLayout), count, note, source)
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
