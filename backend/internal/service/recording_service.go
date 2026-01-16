// 录制任务服务
package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"dblive/internal/model"
	"dblive/internal/repo"
)

type CreateRecordingRequest struct {
	UID string `json:"uid"`
}

type UpdateRecordingRequest struct {
	UID string `json:"uid"`
}

type RecordingService struct {
	recordings *repo.RecordingRepo
	bili       *BilibiliService
	settings   *SettingsService
	mu         sync.Mutex
	jobs       map[int64]*recordingJob
}

func NewRecordingService(recordings *repo.RecordingRepo, bili *BilibiliService, settings *SettingsService) *RecordingService {
	return &RecordingService{
		recordings: recordings,
		bili:       bili,
		settings:   settings,
		jobs:       make(map[int64]*recordingJob),
	}
}

func (s *RecordingService) List(ctx context.Context, userID int64) ([]model.Recording, error) {
	return s.recordings.ListByUserID(userID)
}

func (s *RecordingService) Create(ctx context.Context, userID int64, req CreateRecordingRequest) (*model.Recording, error) {
	// 录制目标必须提供UID
	uid := strings.TrimSpace(req.UID)
	uname := ""

	if uid == "" {
		return nil, errors.New("uid required")
	}

	// 通过UID换取房间号
	uidInfo, err := s.bili.GetUIDInfo(ctx, uid)
	if err != nil {
		return nil, err
	}
	roomID := uidInfo.RoomID
	uname = uidInfo.Uname
	if roomID == 0 {
		return nil, errors.New("invalid uid")
	}

	// 拉取直播间基础信息
	roomInfo, err := s.bili.GetRoomInfo(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// 获取用户全局默认配置
	defaults, err := s.settings.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 应用默认分片大小
	segmentSize := defaults.DefaultSegmentSizeMB
	if segmentSize <= 0 {
		segmentSize = defaultSegmentSizeMB
	}
	// 应用默认分片时长
	segmentTime := defaults.DefaultSegmentTimeMin
	if segmentTime <= 0 {
		segmentTime = defaultSegmentTimeMin
	}
	// 应用默认画质
	quality := defaults.DefaultQuality
	if quality <= 0 {
		quality = defaultQuality
	}
	fileExt := ".flv"
	filenameTemplate := "{{title}}-{{live_time}}"

	now := time.Now().UTC()
	rec := &model.Recording{
		UserID:           userID,
		Platform:         "bilibili",
		UID:              uid,
		RoomID:           roomID,
		Uname:            uname,
		RoomTitle:        roomInfo.Title,
		LiveStatus:       roomInfo.LiveStatus,
		Status:           "idle",
		SegmentSizeMB:    segmentSize,
		SegmentTimeMin:   segmentTime,
		Quality:          quality,
		FileExt:          fileExt,
		FilenameTemplate: filenameTemplate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	id, err := s.recordings.Create(rec)
	if err != nil {
		return nil, err
	}
	rec.ID = id
	return rec, nil
}

func (s *RecordingService) Update(ctx context.Context, userID, id int64, req UpdateRecordingRequest) (*model.Recording, error) {
	// 读取目标任务并更新参数
	rec, err := s.recordings.GetByID(userID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("recording not found")
	}

	newUID := strings.TrimSpace(req.UID)
	if newUID == "" {
		return nil, errors.New("uid required")
	}

	if newUID != rec.UID {
		// UID变更时同步更新房间信息
		uidInfo, err := s.bili.GetUIDInfo(ctx, newUID)
		if err != nil {
			return nil, err
		}
		if uidInfo.RoomID == 0 {
			return nil, errors.New("invalid uid")
		}
		roomInfo, err := s.bili.GetRoomInfo(ctx, uidInfo.RoomID)
		if err != nil {
			return nil, err
		}
		rec.UID = newUID
		rec.RoomID = uidInfo.RoomID
		rec.Uname = uidInfo.Uname
		rec.RoomTitle = roomInfo.Title
		rec.LiveStatus = roomInfo.LiveStatus
	}

	now := time.Now().UTC()
	if err := s.recordings.Update(rec, now); err != nil {
		return nil, err
	}
	rec.UpdatedAt = now
	return rec, nil
}

func (s *RecordingService) Start(ctx context.Context, userID, id int64) (*model.Recording, error) {
	// 切换任务状态为录制中
	rec, err := s.recordings.GetByID(userID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("recording not found")
	}
	if rec.Status == "recording" {
		return nil, errors.New("recording already started")
	}

	if err := s.startJob(ctx, rec); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.recordings.UpdateStatus(userID, id, "recording", now); err != nil {
		s.stopJob(id)
		return nil, err
	}
	rec.Status = "recording"
	rec.UpdatedAt = now
	return rec, nil
}

func (s *RecordingService) Stop(ctx context.Context, userID, id int64) (*model.Recording, error) {
	// 切换任务状态为已停止
	rec, err := s.recordings.GetByID(userID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("recording not found")
	}
	s.stopJob(id)

	now := time.Now().UTC()
	if err := s.recordings.UpdateStatus(userID, id, "stopped", now); err != nil {
		return nil, err
	}
	rec.Status = "stopped"
	rec.UpdatedAt = now
	return rec, nil
}

func (s *RecordingService) Delete(ctx context.Context, userID, id int64) error {
	// 删除任务前先停止
	s.stopJob(id)
	return s.recordings.Delete(userID, id)
}
