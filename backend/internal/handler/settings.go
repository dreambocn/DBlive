// 设置接口
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dblive/internal/service"
	"dblive/internal/util"
)

type SettingsHandler struct {
	settings *service.SettingsService
}

func NewSettingsHandler(settings *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	// 获取当前用户设置
	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	settings, err := h.settings.Get(c.Request.Context(), userID)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var req service.UpdateSettingsRequest
	// 解析设置更新参数
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 保存设置并返回
	settings, err := h.settings.Update(c.Request.Context(), userID, req)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, settings)
}
