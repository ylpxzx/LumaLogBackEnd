package handler

import (
	"net/http"
	"strings"

	"lumalog-backend/model"
	"lumalog-backend/util"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.Repo.ListCategories(c.Request.Context(), currentUserID(c), c.Query("include_hidden") == "true")
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取分类失败")
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req model.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "请求格式不正确")
		return
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		jsonError(c, http.StatusBadRequest, "分类名称不能为空")
		return
	}

	name := strings.TrimSpace(*req.Name)
	colorTheme := util.StringValue(req.ColorTheme, "green")
	sortOrder := util.IntValue(req.SortOrder, 100)
	isHidden := util.BoolValue(req.IsHidden, false)
	slug := util.MakeSlug(name)
	userID := currentUserID(c)

	var id int64
	err := h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO categories (user_id, name, slug, color_theme, sort_order, is_default, is_hidden)
		VALUES ($1, $2, $3, $4, $5, FALSE, $6)
		RETURNING id
	`, userID, name, slug, colorTheme, sortOrder, isHidden).Scan(&id)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "创建分类失败")
		return
	}

	cat, _ := h.Repo.GetCategory(c.Request.Context(), userID, id)
	c.JSON(http.StatusCreated, cat)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := currentUserID(c)
	cat, err := h.Repo.GetCategory(c.Request.Context(), userID, id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "分类不存在")
		return
	}

	var req model.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "请求格式不正确")
		return
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		cat.Name = strings.TrimSpace(*req.Name)
	}
	if req.ColorTheme != nil {
		cat.ColorTheme = strings.TrimSpace(*req.ColorTheme)
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}
	if req.IsHidden != nil {
		cat.IsHidden = *req.IsHidden
	}

	_, err = h.DB.Exec(c.Request.Context(), `
		UPDATE categories
		SET name = $1, color_theme = $2, sort_order = $3, is_hidden = $4, updated_at = NOW()
		WHERE id = $5 AND user_id = $6
	`, cat.Name, cat.ColorTheme, cat.SortOrder, cat.IsHidden, cat.ID, userID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "更新分类失败")
		return
	}

	cat, _ = h.Repo.GetCategory(c.Request.Context(), userID, id)
	c.JSON(http.StatusOK, cat)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	userID := currentUserID(c)
	cat, err := h.Repo.GetCategory(c.Request.Context(), userID, id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "分类不存在")
		return
	}
	if cat.IsDefault {
		jsonError(c, http.StatusBadRequest, "默认分类不能删除，可以隐藏")
		return
	}

	var count int
	err = h.DB.QueryRow(c.Request.Context(), `
		SELECT COUNT(*)
		FROM items
		WHERE user_id = $1 AND category_id = $2 AND deleted_at IS NULL
	`, userID, id).Scan(&count)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "检查分类失败")
		return
	}
	if count > 0 {
		jsonError(c, http.StatusBadRequest, "该分类下还有 item，请先移动或删除 item")
		return
	}

	_, err = h.DB.Exec(c.Request.Context(), `DELETE FROM categories WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "删除分类失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
