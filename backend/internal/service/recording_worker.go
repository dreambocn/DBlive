// 录制任务后台执行器
package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dblive/internal/model"
)

const (
	// recordingExt 录制文件的扩展名
	recordingExt = ".flv"
	// defaultCheckInterval 默认的录制任务检查间隔
	defaultCheckInterval = 15 * time.Second
)

// recordingJob 代表一个正在运行的录制任务
type recordingJob struct {
	// cancel 用于取消该录制任务的上下文
	cancel context.CancelFunc
	// done 在任务完成时关闭的通道
	done chan struct{}
}

// playInfo 包含从B站API获取的直播流播放信息
type playInfo struct {
	// Host 服务器主机地址
	Host string
	// BaseURL 基础URL路径
	BaseURL string
	// Extra 额外的查询参数
	Extra string
	// AcceptQN 支持的画质列表
	AcceptQN []int
	// QN 选定的画质值
	QN int
	// FullURL 完整的播放URL
	FullURL string
}

// m3u8Segment 代表M3U8文件中的一个视频分片
type m3u8Segment struct {
	// Duration 分片的时长（秒）
	Duration float64
	// FileName 分片文件名
	FileName string
}

// m3u8Info 包含M3U8播放列表的信息
type m3u8Info struct {
	// HeaderFile FLV文件头的URI
	HeaderFile string
	// Segments 视频分片列表
	Segments []m3u8Segment
}

// StartScheduler 启动后台调度器，定时检查和管理录制任务
// 根据直播间的在线状态自动启动或停止录制
func (s *RecordingService) StartScheduler(ctx context.Context) {
	// 定时巡检录制任务并自动启动/停止
	ticker := time.NewTicker(defaultCheckInterval)
	defer ticker.Stop()

	// 启动时先执行一次
	s.tickRecordings(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tickRecordings(ctx)
		}
	}
}

// tickRecordings 单次检查所有录制任务的状态
// 根据直播间的在线状态自动启动或停止对应的录制
func (s *RecordingService) tickRecordings(ctx context.Context) {
	records, err := s.recordings.ListAll()
	if err != nil {
		return
	}

	for _, rec := range records {
		roomInfo, err := s.bili.GetRoomInfo(ctx, rec.RoomID)
		if err != nil {
			continue
		}
		rec.RoomTitle = roomInfo.Title
		rec.LiveStatus = roomInfo.LiveStatus
		_ = s.recordings.Update(&rec, time.Now().UTC())

		if roomInfo.LiveStatus == 1 || roomInfo.LiveStatus == 2 {
			if rec.Status != "recording" {
				_, _ = s.Start(ctx, rec.UserID, rec.ID)
			}
			continue
		}
		if rec.Status == "recording" {
			_, _ = s.Stop(ctx, rec.UserID, rec.ID)
		}
	}
}

// startJob 启动一个新的录制任务
// 返回error如果该录制已经在运行中
func (s *RecordingService) startJob(ctx context.Context, rec *model.Recording) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.jobs == nil {
		s.jobs = make(map[int64]*recordingJob)
	}
	if _, exists := s.jobs[rec.ID]; exists {
		return errors.New("recording already running")
	}

	jobCtx, cancel := context.WithCancel(ctx)
	job := &recordingJob{cancel: cancel, done: make(chan struct{})}
	s.jobs[rec.ID] = job

	go func() {
		defer close(job.done)
		defer s.clearJob(rec.ID)
		s.runRecording(jobCtx, rec)
	}()

	return nil
}

// stopJob 停止指定ID的录制任务
func (s *RecordingService) stopJob(id int64) {
	s.mu.Lock()
	job := s.jobs[id]
	s.mu.Unlock()
	if job == nil {
		return
	}
	job.cancel()
	<-job.done
}

// clearJob 清理已完成的录制任务信息
func (s *RecordingService) clearJob(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
}

// runRecording 执行录制任务的主循环
// 不断从M3U8播放列表中获取新的视频分片并保存到文件
// 根据设置的时间或文件大小限制自动分割录制文件
func (s *RecordingService) runRecording(ctx context.Context, rec *model.Recording) {
	cookie, err := s.cookiesForUser(rec.UserID)
	if err != nil || cookie == "" {
		_ = s.recordings.UpdateStatus(rec.UserID, rec.ID, "stopped", time.Now().UTC())
		return
	}

	info, err := s.fetchPlayInfo(ctx, rec.RoomID, rec.Quality, cookie)
	if err != nil || info.FullURL == "" {
		_ = s.recordings.UpdateStatus(rec.UserID, rec.ID, "stopped", time.Now().UTC())
		return
	}

	baseDir := defaultOutputDir
	settings, err := s.settings.Get(ctx, rec.UserID)
	if err == nil && settings.OutputDir != "" {
		baseDir = settings.OutputDir
	}

	uname := sanitizeFilename(rec.Uname)
	if uname == "" {
		uname = "unknown_user"
	}
	outputDir := filepath.Join(baseDir, uname)
	_ = os.MkdirAll(outputDir, 0o755)

	headerBytes, _ := s.downloadHeader(ctx, info, cookie)
	downloaded := make(map[string]struct{})
	partIndex := 1
	partStart := time.Now()
	outputPath := s.buildOutputPath(outputDir, rec, partIndex)

	for {
		select {
		case <-ctx.Done():
			_ = s.recordings.UpdateStatus(rec.UserID, rec.ID, "stopped", time.Now().UTC())
			return
		default:
		}

		m3u8, err := s.fetchM3U8(ctx, info.FullURL, cookie)
		if err != nil || len(m3u8.Segments) == 0 {
			// 直播可能结束，稍后再确认
			time.Sleep(3 * time.Second)
			if !s.isRoomLive(ctx, rec.RoomID) {
				_ = s.recordings.UpdateStatus(rec.UserID, rec.ID, "stopped", time.Now().UTC())
				return
			}
			continue
		}

		for _, seg := range m3u8.Segments {
			if _, exists := downloaded[seg.FileName]; exists {
				continue
			}
			if err := s.appendSegment(ctx, info, seg.FileName, cookie, outputPath, headerBytes); err != nil {
				continue
			}
			downloaded[seg.FileName] = struct{}{}

			// 分片切割（时间或大小）
			if s.shouldRotate(outputPath, partStart, rec) {
				partIndex++
				partStart = time.Now()
				outputPath = s.buildOutputPath(outputDir, rec, partIndex)
			}
		}

		time.Sleep(2 * time.Second)
	}
}

// cookiesForUser 获取指定用户的B站Cookie
func (s *RecordingService) cookiesForUser(userID int64) (string, error) {
	stored, err := s.bili.cookies.GetByUserID(userID)
	if err != nil || stored == nil {
		return "", err
	}
	return stored.Cookie, nil
}

// isRoomLive 检查直播间是否在线
func (s *RecordingService) isRoomLive(ctx context.Context, roomID int64) bool {
	info, err := s.bili.GetRoomInfo(ctx, roomID)
	if err != nil {
		return false
	}
	return info.LiveStatus == 1 || info.LiveStatus == 2
}

// fetchPlayInfo 从B站API获取直播流的播放信息
// 返回包含播放URL及相关参数的playInfo结构
func (s *RecordingService) fetchPlayInfo(ctx context.Context, roomID int64, quality int, cookie string) (*playInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("room_id", strconv.FormatInt(roomID, 10))
	q.Set("qn", strconv.Itoa(quality))
	q.Set("platform", "web")
	q.Set("protocol", "0,1")
	q.Set("format", "0,1,2")
	q.Set("codec", "0,1,2")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", bilibiliUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Origin", "https://live.bilibili.com")
	req.Header.Set("Referer", "https://live.bilibili.com/"+strconv.FormatInt(roomID, 10))
	req.Header.Set("Cookie", cookie)

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Code int `json:"code"`
		Data struct {
			PlayURLInfo struct {
				PlayURL struct {
					Stream []struct {
						Format []struct {
							Codec []struct {
								BaseURL  string `json:"base_url"`
								AcceptQN []int  `json:"accept_qn"`
								URLInfo  []struct {
									Host  string `json:"host"`
									Extra string `json:"extra"`
								} `json:"url_info"`
							} `json:"codec"`
						} `json:"format"`
					} `json:"stream"`
				} `json:"playurl"`
			} `json:"playurl_info"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 {
		return nil, errors.New("play info failed")
	}

	var selected *playInfo
	for _, stream := range payload.Data.PlayURLInfo.PlayURL.Stream {
		for _, format := range stream.Format {
			for _, codec := range format.Codec {
				if len(codec.URLInfo) == 0 {
					continue
				}
				info := playInfo{
					Host:     codec.URLInfo[0].Host,
					BaseURL:  codec.BaseURL,
					Extra:    codec.URLInfo[0].Extra,
					AcceptQN: codec.AcceptQN,
				}
				info.QN = chooseQuality(quality, codec.AcceptQN)
				info.Extra = replaceQN(info.Extra, info.QN)
				info.FullURL = info.Host + info.BaseURL + info.Extra

				// 优先选择HLS(m3u8)播放地址
				if strings.Contains(info.BaseURL, ".m3u8") || strings.Contains(info.FullURL, ".m3u8") {
					selected = &info
					break
				}
				// 记录第一个可用地址作为兜底
				if selected == nil {
					selected = &info
				}
			}
			if selected != nil && strings.Contains(selected.FullURL, ".m3u8") {
				break
			}
		}
		if selected != nil && strings.Contains(selected.FullURL, ".m3u8") {
			break
		}
	}
	if selected == nil {
		return nil, errors.New("no play url")
	}

	return selected, nil
}

// fetchM3U8 获取M3U8播放列表
// 解析列表中的视频分片信息
func (s *RecordingService) fetchM3U8(ctx context.Context, url, cookie string) (*m3u8Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", bilibiliUserAgent)
	req.Header.Set("Cookie", cookie)
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	info := &m3u8Info{}
	for i, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-MAP:URI=") {
			parts := strings.Split(line, "\"")
			if len(parts) >= 2 {
				info.HeaderFile = parts[1]
			}
			continue
		}
		if strings.HasPrefix(line, "#EXTINF") && i+1 < len(lines) {
			durationStr := strings.TrimSuffix(strings.SplitN(strings.TrimPrefix(line, "#EXTINF:"), ",", 2)[0], "\r")
			duration, _ := strconv.ParseFloat(durationStr, 64)
			filename := strings.TrimSpace(lines[i+1])
			if filename != "" && !strings.HasPrefix(filename, "#") {
				info.Segments = append(info.Segments, m3u8Segment{Duration: duration, FileName: filename})
			}
		}
	}
	return info, nil
}

// downloadHeader 下载FLV文件头
// FLV文件头需要写在视频分片数据之前
func (s *RecordingService) downloadHeader(ctx context.Context, info *playInfo, cookie string) ([]byte, error) {
	m3u8, err := s.fetchM3U8(ctx, info.FullURL, cookie)
	if err != nil || m3u8.HeaderFile == "" {
		return nil, err
	}

	headerURL := s.buildSegmentURL(info, m3u8.HeaderFile)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, headerURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", bilibiliUserAgent)
	req.Header.Set("Cookie", cookie)
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// appendSegment 下载单个M3U8分片并追加写入输出文件
func (s *RecordingService) appendSegment(ctx context.Context, info *playInfo, filename, cookie, outputPath string, header []byte) error {
	segmentURL := s.buildSegmentURL(info, filename)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, segmentURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", bilibiliUserAgent)
	req.Header.Set("Cookie", cookie)

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New("segment download failed")
	}

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, _ := file.Stat()
	if stat.Size() == 0 && len(header) > 0 {
		if _, err := file.Write(header); err != nil {
			return err
		}
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

// buildSegmentURL 根据playInfo和分片文件名构建完整的分片URL
func (s *RecordingService) buildSegmentURL(info *playInfo, filename string) string {
	basePath := path.Dir(info.BaseURL)
	extra := strings.TrimPrefix(info.Extra, "?")
	return info.Host + basePath + "/" + filename + "?" + extra
}

// buildOutputPath 根据配置的模板构建输出文件路径
// 支持的模板变量: {{title}} {{live_time}} {{room_id}} {{uname}}
func (s *RecordingService) buildOutputPath(dir string, rec *model.Recording, part int) string {
	title := sanitizeFilename(rec.RoomTitle)
	if title == "" {
		title = "live"
	}
	liveTime := time.Now().Format("20060102_150405")
	filename := strings.ReplaceAll(rec.FilenameTemplate, "{{title}}", title)
	filename = strings.ReplaceAll(filename, "{{live_time}}", liveTime)
	filename = strings.ReplaceAll(filename, "{{room_id}}", strconv.FormatInt(rec.RoomID, 10))
	filename = strings.ReplaceAll(filename, "{{uname}}", sanitizeFilename(rec.Uname))

	filename = sanitizeFilename(filename)
	if filename == "" {
		filename = "live"
	}
	filename = filename + "_" + strconv.Itoa(part) + recordingExt
	return filepath.Join(dir, filename)
}

// shouldRotate 判断是否应该进行文件分割
// 根据设置的时间间隔或文件大小限制进行判断
func (s *RecordingService) shouldRotate(outputPath string, start time.Time, rec *model.Recording) bool {
	if rec.SegmentTimeMin > 0 && time.Since(start) >= time.Duration(rec.SegmentTimeMin)*time.Minute {
		return true
	}
	if rec.SegmentSizeMB <= 0 {
		return false
	}
	stat, err := os.Stat(outputPath)
	if err != nil {
		return false
	}
	return stat.Size() >= int64(rec.SegmentSizeMB)*1024*1024
}

// invalidFileChars 匹配文件名中的非法字符
var invalidFileChars = regexp.MustCompile(`[\\/:*?"<>|]+`)

// chooseQuality 从支持的画质列表中选择目标画质
// 如果目标画质被支持则使用目标，否则使用最高画质
func chooseQuality(target int, accepts []int) int {
	if len(accepts) == 0 {
		if target > 0 {
			return target
		}
		return 10000
	}
	max := accepts[0]
	found := false
	for _, qn := range accepts {
		if qn > max {
			max = qn
		}
		if qn == target {
			found = true
		}
	}
	if target > 0 && found {
		return target
	}
	return max
}

// replaceQN 替换URL查询参数中的画质值
func replaceQN(extra string, qn int) string {
	if qn <= 0 {
		return extra
	}
	re := regexp.MustCompile(`qn=\\d+`)
	if re.MatchString(extra) {
		return re.ReplaceAllString(extra, "qn="+strconv.Itoa(qn))
	}
	sep := "?"
	if strings.Contains(extra, "?") {
		sep = "&"
	}
	return extra + sep + "qn=" + strconv.Itoa(qn)
}

// sanitizeFilename 清理文件名中的非法字符
// 替换为下划线并合并多个连续下划线
func sanitizeFilename(name string) string {
	if name == "" {
		return ""
	}
	cleaned := invalidFileChars.ReplaceAllString(name, "_")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.Trim(cleaned, ".")
	cleaned = regexp.MustCompile(`_+`).ReplaceAllString(cleaned, "_")
	return cleaned
}
