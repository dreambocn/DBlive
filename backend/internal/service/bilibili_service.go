// B站接口与Cookie管理
package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dblive/internal/model"
	"dblive/internal/repo"
)

const bilibiliUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type BilibiliService struct {
	cookies *repo.BilibiliCookieRepo
	client  *http.Client
}

type BiliUser struct {
	Mid   int64  `json:"mid"`
	Uname string `json:"uname"`
	Face  string `json:"face"`
}

type CookieStatus struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	User    *BiliUser `json:"user,omitempty"`
}

type UIDInfo struct {
	UID    string
	RoomID int64
	Uname  string
}

type RoomInfo struct {
	RoomID     int64
	Title      string
	LiveStatus int
}

func NewBilibiliService(cookieRepo *repo.BilibiliCookieRepo) *BilibiliService {
	return &BilibiliService{
		cookies: cookieRepo,
		client:  &http.Client{Timeout: 12 * time.Second},
	}
}

func (s *BilibiliService) GenerateQRCode(ctx context.Context) (string, string, error) {
	// 请求B站生成二维码
	req, err := s.newRequest(ctx, http.MethodGet, "https://passport.bilibili.com/x/passport-login/web/qrcode/generate?source=main-fe-header", "")
	if err != nil {
		return "", "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			URL       string `json:"url"`
			QrcodeKey string `json:"qrcode_key"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", err
	}
	if payload.Code != 0 || payload.Data.URL == "" || payload.Data.QrcodeKey == "" {
		return "", "", errors.New("failed to generate qrcode")
	}

	return payload.Data.URL, payload.Data.QrcodeKey, nil
}

func (s *BilibiliService) PollQRCode(ctx context.Context, userID int64, qrcodeKey string) (*CookieStatus, error) {
	// 轮询二维码状态，使用独立CookieJar承接登录跳转
	pollURL := "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=" + url.QueryEscape(qrcodeKey) + "&source=main-fe-header"
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 12 * time.Second, Jar: jar}

	req, err := s.newRequest(ctx, http.MethodGet, pollURL, "https://passport.bilibili.com/")
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			URL     string `json:"url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 {
		return nil, errors.New("qrcode poll failed")
	}

	// 映射二维码状态码
	status := "pending"
	switch payload.Data.Code {
	case 0:
		status = "success"
	case 86038:
		status = "expired"
	default:
		status = "pending"
	}

	if status != "success" {
		return &CookieStatus{Status: status, Message: payload.Data.Message}, nil
	}
	// 访问回调URL以获取登录态Cookie
	loginReq, err := s.newRequest(ctx, http.MethodGet, payload.Data.URL, "https://passport.bilibili.com/")
	if err != nil {
		return nil, err
	}
	if _, err := client.Do(loginReq); err != nil {
		return nil, err
	}

	// 从CookieJar提取关键字段并持久化
	cookieText, sessdata, biliJct := cookiesFromJar(jar)
	if cookieText == "" {
		return nil, errors.New("login succeeded but cookies missing")
	}

	user, err := s.fetchUser(ctx, client)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.cookies.Upsert(&model.BilibiliCookie{
		UserID:    userID,
		Cookie:    cookieText,
		Sessdata:  sessdata,
		BiliJct:   biliJct,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return nil, err
	}

	return &CookieStatus{
		Status:  "success",
		Message: payload.Data.Message,
		User:    user,
	}, nil
}

func (s *BilibiliService) CookieStatus(ctx context.Context, userID int64) (*CookieStatus, error) {
	// 读取本地保存的Cookie
	stored, err := s.cookies.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if stored == nil || stored.Cookie == "" {
		return &CookieStatus{Status: "missing", Message: "no cookie"}, nil
	}

	// 构造CookieJar用于校验登录态
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	cookieURL, _ := url.Parse("https://bilibili.com/")
	wwwURL, _ := url.Parse("https://www.bilibili.com/")
	apiURL, _ := url.Parse("https://api.bilibili.com/")
	parsed := parseCookieString(stored.Cookie)
	jar.SetCookies(cookieURL, parsed)
	jar.SetCookies(wwwURL, parsed)
	jar.SetCookies(apiURL, parsed)

	client := &http.Client{Timeout: 12 * time.Second, Jar: jar}
	// 调用导航接口验证登录状态
	user, err := s.fetchUser(ctx, client)
	if err != nil {
		return &CookieStatus{Status: "invalid", Message: "cookie invalid"}, nil
	}

	return &CookieStatus{Status: "active", Message: "ok", User: user}, nil
}

func (s *BilibiliService) GetUIDInfo(ctx context.Context, uid string) (*UIDInfo, error) {
	// 通过UID查询直播间信息
	target := "https://api.live.bilibili.com/live_user/v1/Master/info?uid=" + url.QueryEscape(uid)
	req, err := s.newRequest(ctx, http.MethodGet, target, "https://live.bilibili.com/")
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			RoomID  int64 `json:"roomid"`
			RoomAlt int64 `json:"room_id"`
			Info    struct {
				Uname string `json:"uname"`
			} `json:"info"`
			Uname string `json:"uname"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 {
		return nil, errors.New(payload.Message)
	}

	roomID := payload.Data.RoomID
	if roomID == 0 {
		roomID = payload.Data.RoomAlt
	}
	uname := payload.Data.Info.Uname
	if uname == "" {
		uname = payload.Data.Uname
	}

	return &UIDInfo{
		UID:    uid,
		RoomID: roomID,
		Uname:  uname,
	}, nil
}

func (s *BilibiliService) GetRoomInfo(ctx context.Context, roomID int64) (*RoomInfo, error) {
	// 查询直播间标题与直播状态
	target := "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=" + url.QueryEscape(strconv.FormatInt(roomID, 10))
	req, err := s.newRequest(ctx, http.MethodGet, target, "https://live.bilibili.com/")
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Title      string `json:"title"`
			LiveStatus int    `json:"live_status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 {
		return nil, errors.New(payload.Message)
	}

	return &RoomInfo{
		RoomID:     roomID,
		Title:      payload.Data.Title,
		LiveStatus: payload.Data.LiveStatus,
	}, nil
}

func (s *BilibiliService) fetchUser(ctx context.Context, client *http.Client) (*BiliUser, error) {
	// 拉取用户导航信息以确认登录
	req, err := s.newRequest(ctx, http.MethodGet, "https://api.bilibili.com/x/web-interface/nav", "https://www.bilibili.com/")
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			IsLogin bool   `json:"isLogin"`
			Mid     int64  `json:"mid"`
			Uname   string `json:"uname"`
			Face    string `json:"face"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 || !payload.Data.IsLogin {
		return nil, errors.New("not logged in")
	}

	return &BiliUser{
		Mid:   payload.Data.Mid,
		Uname: payload.Data.Uname,
		Face:  payload.Data.Face,
	}, nil
}

func (s *BilibiliService) newRequest(ctx context.Context, method, target, referer string) (*http.Request, error) {
	// 统一设置B站请求头，避免风控
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", bilibiliUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	if referer != "" {
		req.Header.Set("Referer", referer)
		if parsed, err := url.Parse(referer); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			req.Header.Set("Origin", parsed.Scheme+"://"+parsed.Host)
		}
	}
	return req, nil
}

func cookiesFromJar(jar *cookiejar.Jar) (string, string, string) {
	// 提取完整Cookie字符串与关键字段
	cookieURL, _ := url.Parse("https://bilibili.com/")
	cookies := jar.Cookies(cookieURL)
	if len(cookies) == 0 {
		return "", "", ""
	}

	pairs := make([]string, 0, len(cookies))
	var sessdata string
	var biliJct string
	for _, c := range cookies {
		pairs = append(pairs, c.Name+"="+c.Value)
		switch c.Name {
		case "SESSDATA":
			sessdata = c.Value
		case "bili_jct":
			biliJct = c.Value
		}
	}
	return strings.Join(pairs, "; "), sessdata, biliJct
}

func parseCookieString(raw string) []*http.Cookie {
	// 将Cookie文本解析为http.Cookie列表
	parts := strings.Split(raw, ";")
	cookies := make([]*http.Cookie, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		kv := strings.SplitN(trimmed, "=", 2)
		if len(kv) != 2 {
			continue
		}
		cookies = append(cookies, &http.Cookie{
			Name:  strings.TrimSpace(kv[0]),
			Value: strings.TrimSpace(kv[1]),
		})
	}
	return cookies
}
