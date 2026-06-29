package handler

import (
	"net/http"
	"strings"

	"lumalog-backend/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "请求格式不正确")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}
	if !strings.Contains(email, "@") || len(password) < 6 {
		jsonError(c, http.StatusBadRequest, "请输入有效邮箱和至少 6 位密码")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "密码处理失败")
		return
	}

	ctx := c.Request.Context()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "创建用户失败")
		return
	}
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id
	`, email, string(hash), displayName).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			jsonError(c, http.StatusConflict, "邮箱已注册")
			return
		}
		jsonError(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	if err := h.Repo.InsertDefaultCategoriesTx(ctx, tx, userID); err != nil {
		jsonError(c, http.StatusInternalServerError, "初始化分类失败")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		jsonError(c, http.StatusInternalServerError, "提交用户失败")
		return
	}

	u, err := h.Repo.GetUser(ctx, userID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取用户失败")
		return
	}
	cats, _ := h.Repo.ListCategories(ctx, userID, false)
	token, _ := h.signToken(userID)
	c.JSON(http.StatusCreated, model.AuthResponse{Token: token, User: u, Categories: cats})
}

func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "请求格式不正确")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	var userID int64
	var hash string
	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT id, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&userID, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		jsonError(c, http.StatusUnauthorized, "邮箱或密码不正确")
		return
	}

	_ = h.Repo.EnsureDefaultCategories(c.Request.Context(), userID)
	u, err := h.Repo.GetUser(c.Request.Context(), userID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "读取用户失败")
		return
	}
	cats, _ := h.Repo.ListCategories(c.Request.Context(), userID, false)
	token, _ := h.signToken(userID)
	c.JSON(http.StatusOK, model.AuthResponse{Token: token, User: u, Categories: cats})
}

func (h *Handler) Me(c *gin.Context) {
	u, err := h.Repo.GetUser(c.Request.Context(), currentUserID(c))
	if err != nil {
		jsonError(c, http.StatusNotFound, "用户不存在")
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	var req model.PreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "请求格式不正确")
		return
	}

	u, err := h.Repo.GetUser(c.Request.Context(), currentUserID(c))
	if err != nil {
		jsonError(c, http.StatusNotFound, "用户不存在")
		return
	}

	if req.ThemePreference != nil {
		value := strings.TrimSpace(*req.ThemePreference)
		if value != "system" && value != "light" && value != "dark" {
			jsonError(c, http.StatusBadRequest, "主题偏好不正确")
			return
		}
		u.ThemePreference = value
	}
	if req.LanguagePreference != nil {
		value := strings.TrimSpace(*req.LanguagePreference)
		if value != "zh" && value != "en" {
			jsonError(c, http.StatusBadRequest, "语言偏好不正确")
			return
		}
		u.LanguagePreference = value
	}
	if req.DashboardViewMode != nil {
		value := strings.TrimSpace(*req.DashboardViewMode)
		if value != "all" && value != "category" {
			jsonError(c, http.StatusBadRequest, "首页显示模式不正确")
			return
		}
		u.DashboardViewMode = value
	}
	if req.ShowTodayStatus != nil {
		u.ShowTodayStatus = *req.ShowTodayStatus
	}
	if req.ShowCurrentStreak != nil {
		u.ShowCurrentStreak = *req.ShowCurrentStreak
	}
	if req.ShowLongestStreak != nil {
		u.ShowLongestStreak = *req.ShowLongestStreak
	}
	if req.ShowCompletionRate != nil {
		u.ShowCompletionRate = *req.ShowCompletionRate
	}
	if req.ShowTotalCheckins != nil {
		u.ShowTotalCheckins = *req.ShowTotalCheckins
	}

	_, err = h.DB.Exec(c.Request.Context(), `
		UPDATE users
		SET theme_preference = $1,
			language_preference = $2,
			dashboard_view_mode = $3,
			show_today_status = $4,
			show_current_streak = $5,
			show_longest_streak = $6,
			show_completion_rate = $7,
			show_total_checkins = $8,
			updated_at = NOW()
		WHERE id = $9
	`, u.ThemePreference, u.LanguagePreference, u.DashboardViewMode, u.ShowTodayStatus, u.ShowCurrentStreak, u.ShowLongestStreak, u.ShowCompletionRate, u.ShowTotalCheckins, u.ID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "保存偏好失败")
		return
	}

	u, _ = h.Repo.GetUser(c.Request.Context(), u.ID)
	c.JSON(http.StatusOK, u)
}
