// B站Cookie扫码授权接口
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dblive/internal/service"
	"dblive/internal/util"
)

type BilibiliCookieHandler struct {
	bili *service.BilibiliService
}

func NewBilibiliCookieHandler(bili *service.BilibiliService) *BilibiliCookieHandler {
	return &BilibiliCookieHandler{bili: bili}
}

func (h *BilibiliCookieHandler) GenerateQRCode(c *gin.Context) {
	// 向B站申请二维码
	qrURL, qrKey, err := h.bili.GenerateQRCode(c.Request.Context())
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, "failed to generate qrcode")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qr_url":     qrURL,
		"qrcode_key": qrKey,
	})
}

type qrcodePollRequest struct {
	QrcodeKey string `json:"qrcode_key"`
}

func (h *BilibiliCookieHandler) PollQRCode(c *gin.Context) {
	var req qrcodePollRequest
	// 校验二维码Key
	if err := c.ShouldBindJSON(&req); err != nil || req.QrcodeKey == "" {
		util.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// 登录用户才允许绑定Cookie
	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 轮询二维码状态并保存Cookie
	status, err := h.bili.PollQRCode(c.Request.Context(), userID, req.QrcodeKey)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *BilibiliCookieHandler) Status(c *gin.Context) {
	// 读取当前用户授权状态
	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	status, err := h.bili.CookieStatus(c.Request.Context(), userID)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, status)
}
