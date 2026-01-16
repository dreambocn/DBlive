// 路由配置与守卫
import { createRouter, createWebHistory } from "vue-router";
import Login from "../views/Login.vue";
import DashboardLayout from "../views/DashboardLayout.vue";
import CookieAuth from "../views/CookieAuth.vue";
import Recordings from "../views/Recordings.vue";
import Settings from "../views/Settings.vue";
import { authState } from "../stores/auth";

const routes = [
  { path: "/", redirect: "/login" },
  { path: "/login", component: Login },
  {
    path: "/app",
    component: DashboardLayout,
    redirect: "/app/recordings",
    children: [
      { path: "cookie", component: CookieAuth },
      { path: "recordings", component: Recordings },
      { path: "settings", component: Settings }
    ]
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach((to, from, next) => {
  // 未登录禁止访问控制台页面
  if (to.path.startsWith("/app") && !authState.accessToken) {
    next("/login");
    return;
  }
  next();
});

export default router;

