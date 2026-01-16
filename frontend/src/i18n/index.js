// i18n初始化
import { createI18n } from "vue-i18n";
import enUS from "./locales/en-US.json";
import zhCN from "./locales/zh-CN.json";

const messages = {
  "zh-CN": zhCN,
  "en-US": enUS
};

const i18n = createI18n({
  legacy: false,
  // 优先使用本地缓存语言
  locale: localStorage.getItem("dblive-lang") || "zh-CN",
  fallbackLocale: "en-US",
  messages
});

export default i18n;

