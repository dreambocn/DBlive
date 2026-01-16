// 认证相关HTTP处理
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dblive/internal/service"
	"dblive/internal/util"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	// 解析登录请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// 校验账号密码并生成令牌
	tokens, err := h.auth.Login(c, req.Username, req.Password)
	if err != nil {
		util.JSONError(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	// 解析刷新令牌
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	tokens, err := h.auth.Refresh(c, req.RefreshToken)
	if err != nil {
		util.JSONError(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	// 解析注销请求
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := h.auth.Logout(c, req.RefreshToken); err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	// 从上下文读取当前用户
	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.auth.UserByID(c, userID)
	if err != nil {
		util.JSONError(c, http.StatusNotFound, "user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}
