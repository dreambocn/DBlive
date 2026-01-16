// 录制任务接口
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"dblive/internal/service"
	"dblive/internal/util"
)

type RecordingHandler struct {
	recordings *service.RecordingService
}

func NewRecordingHandler(recordings *service.RecordingService) *RecordingHandler {
	return &RecordingHandler{recordings: recordings}
}

func (h *RecordingHandler) List(c *gin.Context) {
	// 仅返回当前用户的录制任务
	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	records, err := h.recordings.List(c.Request.Context(), userID)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, records)
}

func (h *RecordingHandler) Create(c *gin.Context) {
	var req service.CreateRecordingRequest
	// 解析创建参数
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 创建录制任务并返回
	record, err := h.recordings.Create(c.Request.Context(), userID, req)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordingHandler) Update(c *gin.Context) {
	var req service.UpdateRecordingRequest
	// 解析更新参数
	if err := c.ShouldBindJSON(&req); err != nil {
		util.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// 解析录制任务ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		util.JSONError(c, http.StatusBadRequest, "invalid id")
		return
	}

	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 更新录制参数
	record, err := h.recordings.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordingHandler) Start(c *gin.Context) {
	// 启动录制任务
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		util.JSONError(c, http.StatusBadRequest, "invalid id")
		return
	}

	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	record, err := h.recordings.Start(c.Request.Context(), userID, id)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordingHandler) Stop(c *gin.Context) {
	// 停止录制任务
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		util.JSONError(c, http.StatusBadRequest, "invalid id")
		return
	}

	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	record, err := h.recordings.Stop(c.Request.Context(), userID, id)
	if err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordingHandler) Delete(c *gin.Context) {
	// 删除录制任务
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		util.JSONError(c, http.StatusBadRequest, "invalid id")
		return
	}

	userID := util.UserIDFromContext(c)
	if userID == 0 {
		util.JSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.recordings.Delete(c.Request.Context(), userID, id); err != nil {
		util.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
