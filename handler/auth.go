package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"lumalog-backend/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			jsonError(c, http.StatusUnauthorized, "请先登录")
			c.Abort()
			return
		}
		userID, err := h.validateToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			jsonError(c, http.StatusUnauthorized, "登录已失效")
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

func (h *Handler) signToken(userID int64) (string, error) {
	payload, err := json.Marshal(model.TokenPayload{
		UserID: userID,
		Exp:    time.Now().Add(14 * 24 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	sig := signHMAC([]byte(body), h.JWTSecret)
	return body + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (h *Handler) validateToken(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, errors.New("bad token")
	}
	expected := signHMAC([]byte(parts[0]), h.JWTSecret)
	actual, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(expected, actual) {
		return 0, errors.New("bad signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, err
	}
	var payload model.TokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, err
	}
	if payload.Exp < time.Now().Unix() {
		return 0, errors.New("expired")
	}
	return payload.UserID, nil
}

func signHMAC(payload, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return mac.Sum(nil)
}
