// 设置服务
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"dblive/internal/model"
	"dblive/internal/repo"
)

const (
	defaultOutputDir      = "recordings"
	defaultSegmentTimeMin = 30
	defaultQuality        = 10000
	defaultSegmentSizeMB  = 20
)

type UpdateSettingsRequest struct {
	OutputDir             *string `json:"output_dir"`
	DefaultSegmentTimeMin *int    `json:"default_segment_time_min"`
	DefaultQuality        *int    `json:"default_quality"`
	DefaultSegmentSizeMB  *int    `json:"default_segment_size_mb"`
}

type SettingsService struct {
	settings *repo.SettingsRepo
}

func NewSettingsService(settings *repo.SettingsRepo) *SettingsService {
	return &SettingsService{settings: settings}
}

func (s *SettingsService) Get(ctx context.Context, userID int64) (*model.UserSettings, error) {
	_ = ctx
	// 优先读取已保存的用户设置
	existing, err := s.settings.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	// 未设置时返回默认值
	return &model.UserSettings{
		UserID:                userID,
		OutputDir:             defaultOutputDir,
		DefaultSegmentTimeMin: defaultSegmentTimeMin,
		DefaultQuality:        defaultQuality,
		DefaultSegmentSizeMB:  defaultSegmentSizeMB,
	}, nil
}

func (s *SettingsService) Update(ctx context.Context, userID int64, req UpdateSettingsRequest) (*model.UserSettings, error) {
	_ = ctx
	// 读取现有设置或初始化默认值
	settings, err := s.settings.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		settings = &model.UserSettings{
			UserID:                userID,
			OutputDir:             defaultOutputDir,
			DefaultSegmentTimeMin: defaultSegmentTimeMin,
			DefaultQuality:        defaultQuality,
			DefaultSegmentSizeMB:  defaultSegmentSizeMB,
		}
	}

	// 按字段增量更新
	updated := false
	if req.OutputDir != nil {
		outputDir := strings.TrimSpace(*req.OutputDir)
		if outputDir == "" {
			return nil, errors.New("output_dir cannot be empty")
		}
		settings.OutputDir = outputDir
		updated = true
	}
	if req.DefaultSegmentTimeMin != nil {
		if *req.DefaultSegmentTimeMin <= 0 {
			return nil, errors.New("default_segment_time_min must be positive")
		}
		settings.DefaultSegmentTimeMin = *req.DefaultSegmentTimeMin
		updated = true
	}
	if req.DefaultQuality != nil {
		if *req.DefaultQuality <= 0 {
			return nil, errors.New("default_quality must be positive")
		}
		settings.DefaultQuality = *req.DefaultQuality
		updated = true
	}
	if req.DefaultSegmentSizeMB != nil {
		if *req.DefaultSegmentSizeMB <= 0 {
			return nil, errors.New("default_segment_size_mb must be positive")
		}
		settings.DefaultSegmentSizeMB = *req.DefaultSegmentSizeMB
		updated = true
	}

	if !updated {
		return nil, errors.New("no fields to update")
	}

	now := time.Now().UTC()
	if settings.CreatedAt.IsZero() {
		settings.CreatedAt = now
	}
	settings.UpdatedAt = now

	// 持久化更新结果
	if err := s.settings.Upsert(settings); err != nil {
		return nil, err
	}

	return settings, nil
}
