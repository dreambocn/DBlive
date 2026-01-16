import { createI18n } from "vue-i18n";

const messages = {
  "zh-CN": {
    loginTitle: "欢迎回来",
    loginSubtitle: "使用你的账号登录 DBlive",
    username: "用户名",
    password: "密码",
    login: "登录",
    logout: "退出登录",
    dashboardTitle: "控制台",
    dashboardSubtitle: "认证信息概览",
    accessToken: "访问令牌",
    refreshToken: "刷新令牌",
    user: "当前用户"
  },
  "en-US": {
    loginTitle: "Welcome Back",
    loginSubtitle: "Sign in to DBlive",
    username: "Username",
    password: "Password",
    login: "Sign In",
    logout: "Sign Out",
    dashboardTitle: "Dashboard",
    dashboardSubtitle: "Auth status overview",
    accessToken: "Access Token",
    refreshToken: "Refresh Token",
    user: "User"
  }
};

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem("dblive-lang") || "zh-CN",
  fallbackLocale: "en-US",
  messages
});

export default i18n;
