// API请求封装
import { authState, refresh } from "../stores/auth";

const baseUrl = import.meta.env.VITE_API_URL || "http://localhost:8080";

async function request(path, options = {}, retry = true) {
  // 自动附带访问令牌
  const headers = {
    "Content-Type": "application/json",
    ...options.headers
  };

  if (authState.accessToken) {
    headers.Authorization = `Bearer ${authState.accessToken}`;
  }

  const res = await fetch(`${baseUrl}${path}`, {
    ...options,
    headers
  });

  // 访问令牌过期时尝试刷新
  if (res.status === 401 && retry && authState.refreshToken) {
    await refresh();
    return request(path, options, false);
  }

  // 统一错误处理
  if (!res.ok) {
    const payload = await res.json().catch(() => ({}));
    throw new Error(payload.error || "Request failed");
  }

  if (res.status === 204) {
    return null;
  }

  return res.json();
}

const api = {
  get: (path) => request(path, { method: "GET" }),
  post: (path, body) =>
    request(path, {
      method: "POST",
      body: JSON.stringify(body)
    }),
  put: (path, body) =>
    request(path, {
      method: "PUT",
      body: JSON.stringify(body)
    }),
  delete: (path) => request(path, { method: "DELETE" })
};

export default api;

