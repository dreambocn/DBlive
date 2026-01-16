import { reactive } from "vue";
import api from "../api/client";

export const authState = reactive({
  accessToken: localStorage.getItem("dblive-access") || "",
  refreshToken: localStorage.getItem("dblive-refresh") || "",
  user: null
});

export async function login(username, password) {
  const res = await api.post("/api/v1/auth/login", { username, password });
  setTokens(res);
  await fetchMe();
}

export async function refresh() {
  if (!authState.refreshToken) {
    throw new Error("missing refresh token");
  }
  const res = await api.post("/api/v1/auth/refresh", {
    refresh_token: authState.refreshToken
  });
  setTokens(res);
}

export async function logout() {
  if (authState.refreshToken) {
    await api.post("/api/v1/auth/logout", {
      refresh_token: authState.refreshToken
    });
  }
  authState.accessToken = "";
  authState.refreshToken = "";
  authState.user = null;
  localStorage.removeItem("dblive-access");
  localStorage.removeItem("dblive-refresh");
}

export async function fetchMe() {
  authState.user = await api.get("/api/v1/me");
}

function setTokens(res) {
  authState.accessToken = res.access_token;
  authState.refreshToken = res.refresh_token;
  localStorage.setItem("dblive-access", authState.accessToken);
  localStorage.setItem("dblive-refresh", authState.refreshToken);
}
