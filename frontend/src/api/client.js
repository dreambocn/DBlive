import { authState, refresh } from "../stores/auth";

const baseUrl = import.meta.env.VITE_API_URL || "http://localhost:8080";

async function request(path, options = {}, retry = true) {
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

  if (res.status === 401 && retry && authState.refreshToken) {
    await refresh();
    return request(path, options, false);
  }

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
    })
};

export default api;
