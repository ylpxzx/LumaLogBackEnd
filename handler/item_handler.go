package handler

import (
	"net/http"
	"strings"
	"time"

	"lumalog-backend/model"
	"lumalog-backend/util"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListItems(c *gin.Context) {
	items, err := h.Repo.ListItems(c.Request.Context(), currentUserID(c), c.Query("dashboard") == "true")
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取 item 失败")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req model.ItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "请求格式不正确")
		return
	}
	userID := currentUserID(c)

	name := strings.TrimSpace(util.StringValue(req.Name, ""))
	if name == "" {
		jsonError(c, http.StatusBadRequest, "item 名称不能为空")
		return
	}

	categoryID := util.Int64Value(req.CategoryID, 0)
	if categoryID == 0 {
		first, err := h.Repo.FirstVisibleCategory(c.Request.Context(), userID)
		if err != nil {
			jsonError(c, http.StatusBadRequest, "请先创建分类")
			return
		}
		categoryID = first.ID
	}
	cat, err := h.Repo.GetCategory(c.Request.Context(), userID, categoryID)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "分类不存在")
		return
	}

	today := time.Now().Format(util.DateLayout)
	startDate := util.NormalizeDate(util.StringValue(req.StartDate, today), today)
	isUnlimited := util.BoolValue(req.IsUnlimited, true)
	endDate := util.NormalizeDate(util.StringValue(req.EndDate, ""), "")
	if isUnlimited {
		endDate = ""
	}
	target := util.IntValue(req.DailyTargetCount, 1)
	if target < 1 {
		target = 1
	}
	timeMode := util.NormalizeTimeMode(util.StringValue(req.TimeMode, "all_day"))
	validStart := util.NormalizeClock(util.StringValue(req.ValidStartTime, ""))
	validEnd := util.NormalizeClock(util.StringValue(req.ValidEndTime, ""))
	if timeMode == "time_range" {
		if validStart == "" {
			validStart = "09:00"
		}
		if validEnd == "" {
			validEnd = "23:59"
		}
	}

	var id int64
	err = h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO items (
			user_id, category_id, name, description, color_theme, start_date, end_date,
			is_unlimited, daily_target_count, time_mode, valid_start_time, valid_end_time,
			allow_makeup, makeup_limit_days, allow_extra_checkins, show_on_dashboard, sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id
	`, userID, categoryID, name, util.StringValue(req.Description, ""), util.StringValue(req.ColorTheme, cat.ColorTheme),
		startDate, endDate, isUnlimited, target, timeMode, validStart, validEnd,
		util.BoolValue(req.AllowMakeup, false), util.IntValue(req.MakeupLimitDays, 0),
		util.BoolValue(req.AllowExtraCheckins, false), util.BoolValue(req.ShowOnDashboard, true), util.IntValue(req.SortOrder, 0)).Scan(&id)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "创建 item 失败")
		return
	}

	it, _ := h.Repo.GetItem(c.Request.Context(), userID, id)
	c.JSON(http.StatusCreated, it)
}

func (h *Handler) GetItem(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	it, err := h.Repo.GetItem(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "item 不存在")
		return
	}
	di, err := h.Service.BuildDashboardItem(c.Request.Context(), currentUserID(c), it)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取 item 失败")
		return
	}
	c.JSON(http.StatusOK, di)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	userID := currentUserID(c)
	it, err := h.Repo.GetItem(c.Request.Context(), userID, id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "item 不存在")
		return
	}

	var req model.ItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "请求格式不正确")
		return
	}

	if req.CategoryID != nil {
		if _, err := h.Repo.GetCategory(c.Request.Context(), userID, *req.CategoryID); err != nil {
			jsonError(c, http.StatusBadRequest, "分类不存在")
			return
		}
		it.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		it.Name = strings.TrimSpace(*req.Name)
	}
	if it.Name == "" {
		jsonError(c, http.StatusBadRequest, "item 名称不能为空")
		return
	}
	if req.Description != nil {
		it.Description = strings.TrimSpace(*req.Description)
	}
	if req.ColorTheme != nil {
		it.ColorTheme = strings.TrimSpace(*req.ColorTheme)
	}
	if req.StartDate != nil {
		it.StartDate = util.NormalizeDate(*req.StartDate, it.StartDate)
	}
	if req.EndDate != nil {
		it.EndDate = util.NormalizeDate(*req.EndDate, "")
	}
	if req.IsUnlimited != nil {
		it.IsUnlimited = *req.IsUnlimited
		if it.IsUnlimited {
			it.EndDate = ""
		}
	}
	if req.DailyTargetCount != nil {
		it.DailyTargetCount = *req.DailyTargetCount
		if it.DailyTargetCount < 1 {
			it.DailyTargetCount = 1
		}
	}
	if req.TimeMode != nil {
		it.TimeMode = util.NormalizeTimeMode(*req.TimeMode)
	}
	if req.ValidStartTime != nil {
		it.ValidStartTime = util.NormalizeClock(*req.ValidStartTime)
	}
	if req.ValidEndTime != nil {
		it.ValidEndTime = util.NormalizeClock(*req.ValidEndTime)
	}
	if it.TimeMode == "all_day" {
		it.ValidStartTime = ""
		it.ValidEndTime = ""
	}
	if req.AllowMakeup != nil {
		it.AllowMakeup = *req.AllowMakeup
	}
	if req.MakeupLimitDays != nil {
		it.MakeupLimitDays = *req.MakeupLimitDays
	}
	if req.AllowExtraCheckins != nil {
		it.AllowExtraCheckins = *req.AllowExtraCheckins
	}
	if req.ShowOnDashboard != nil {
		it.ShowOnDashboard = *req.ShowOnDashboard
	}
	if req.SortOrder != nil {
		it.SortOrder = *req.SortOrder
	}

	_, err = h.DB.Exec(c.Request.Context(), `
		UPDATE items
		SET category_id = $1, name = $2, description = $3, color_theme = $4,
			start_date = $5, end_date = $6, is_unlimited = $7, daily_target_count = $8,
			time_mode = $9, valid_start_time = $10, valid_end_time = $11,
			allow_makeup = $12, makeup_limit_days = $13, allow_extra_checkins = $14,
			show_on_dashboard = $15, sort_order = $16, updated_at = NOW()
		WHERE id = $17 AND user_id = $18 AND deleted_at IS NULL
	`, it.CategoryID, it.Name, it.Description, it.ColorTheme, it.StartDate, it.EndDate,
		it.IsUnlimited, it.DailyTargetCount, it.TimeMode, it.ValidStartTime, it.ValidEndTime,
		it.AllowMakeup, it.MakeupLimitDays, it.AllowExtraCheckins, it.ShowOnDashboard,
		it.SortOrder, it.ID, userID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "更新 item 失败")
		return
	}

	it, _ = h.Repo.GetItem(c.Request.Context(), userID, id)
	c.JSON(http.StatusOK, it)
}

func (h *Handler) DeleteItem(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	tag, err := h.DB.Exec(c.Request.Context(), `
		UPDATE items
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, currentUserID(c))
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "删除 item 失败")
		return
	}
	if tag.RowsAffected() == 0 {
		jsonError(c, http.StatusNotFound, "item 不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ItemStats(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	userID := currentUserID(c)
	it, err := h.Repo.GetItem(c.Request.Context(), userID, id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "item 不存在")
		return
	}
	di, err := h.Service.BuildDashboardItem(c.Request.Context(), userID, it)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取统计失败")
		return
	}
	c.JSON(http.StatusOK, di.Stats)
}
