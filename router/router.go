package router

import (
	"net/http"

	"lumalog-backend/handler"

	"github.com/gin-gonic/gin"
)

func New(h *handler.Handler) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/logout", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	protected := api.Group("")
	protected.Use(h.AuthMiddleware())
	protected.GET("/me", h.Me)
	protected.PATCH("/me/preferences", h.UpdatePreferences)
	protected.GET("/dashboard", h.Dashboard)
	protected.GET("/categories", h.ListCategories)
	protected.POST("/categories", h.CreateCategory)
	protected.PATCH("/categories/:id", h.UpdateCategory)
	protected.DELETE("/categories/:id", h.DeleteCategory)
	protected.GET("/items", h.ListItems)
	protected.POST("/items", h.CreateItem)
	protected.GET("/items/:id", h.GetItem)
	protected.PATCH("/items/:id", h.UpdateItem)
	protected.DELETE("/items/:id", h.DeleteItem)
	protected.GET("/items/:id/stats", h.ItemStats)
	protected.GET("/items/:id/checkins", h.ListCheckins)
	protected.POST("/items/:id/checkins", h.CreateCheckin)
	protected.DELETE("/items/:id/checkins/:checkinId", h.DeleteCheckin)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
