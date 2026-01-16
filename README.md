# DBlive
个人+AI 开发的直播录制系统。前后端分离基础框架：Go + Gin + SQLite + JWT/Refresh Token + Vue3 + Vite + i18n + 日夜切换。

## 目录结构
- backend: Go 后端服务
- frontend: Vue 前端应用

## 后端启动

```bash
cd backend
go mod tidy
go run ./cmd/server
```

默认管理账号：`admin` / `admin123`

## 前端启动

```bash
cd frontend
npm install
npm run dev
```

前端默认请求 `http://localhost:8080`，可以通过 `VITE_API_URL` 指定。

## Docker Compose

```bash
docker compose up --build
```

服务地址：
- 后端: `http://localhost:8080`
- 前端: `http://localhost:5173`

## 单容器 Docker

```bash
docker build -t dblive:latest .
docker run -p 8080:8080 dblive:latest
```
