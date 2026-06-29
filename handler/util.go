package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func currentUserID(c *gin.Context) int64 {
	value, exists := c.Get("userID")
	if !exists {
		return 0
	}
	id, _ := value.(int64)
	return id
}

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		jsonError(c, http.StatusBadRequest, "参数不正确")
		return 0, false
	}
	return id, true
}

func jsonError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
