package handler

import (
	"net/http"

	"lumalog-backend/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Dashboard(c *gin.Context) {
	ctx := c.Request.Context()
	userID := currentUserID(c)

	u, err := h.Repo.GetUser(ctx, userID)
	if err != nil {
		jsonError(c, http.StatusNotFound, "用户不存在")
		return
	}
	cats, err := h.Repo.ListCategories(ctx, userID, false)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取分类失败")
		return
	}
	items, err := h.Repo.ListItems(ctx, userID, true, false)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取 item 失败")
		return
	}

	result := make([]model.DashboardItem, 0, len(items))
	for _, it := range items {
		di, err := h.Service.BuildDashboardItem(ctx, userID, it)
		if err != nil {
			jsonError(c, http.StatusInternalServerError, "读取热力图失败")
			return
		}
		result = append(result, di)
	}

	c.JSON(http.StatusOK, model.DashboardResponse{
		User:       u,
		Categories: cats,
		Items:      result,
	})
}

func (h *Handler) Badges(c *gin.Context) {
	badges, err := h.Service.UserBadges(c.Request.Context(), currentUserID(c))
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取成就徽章失败")
		return
	}
	c.JSON(http.StatusOK, badges)
}
