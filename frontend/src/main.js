// 前端入口
import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import i18n from "./i18n";
// 引入全局样式
import "./assets/styles.css";

// 注册路由与国际化
createApp(App).use(router).use(i18n).mount("#app");

