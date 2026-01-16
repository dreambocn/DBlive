<!-- 控制台布局 -->
<template>
  <section class="dashboard-shell">
    <aside class="dashboard-nav">
      <div class="nav-brand">
        <div class="nav-title">{{ t("dashboardTitle") }}</div>
        <div class="nav-subtitle">{{ user?.username || "-" }}</div>
      </div>
      <nav class="nav-links">
        <router-link to="/app/recordings" class="nav-link">
          {{ t("recordings") }}
        </router-link>
        <router-link to="/app/cookie" class="nav-link">
          {{ t("cookieAuth") }}
        </router-link>
        <router-link to="/app/settings" class="nav-link">
          {{ t("settings") }}
        </router-link>
      </nav>
      <button class="secondary nav-logout" @click="handleLogout">
        {{ t("logout") }}
      </button>
    </aside>
    <div class="dashboard-main">
      <router-view />
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { authState, fetchMe, logout } from "../stores/auth";

const { t } = useI18n();
const router = useRouter();
const user = computed(() => authState.user);

onMounted(async () => {
  // 未加载用户信息时先拉取
  if (!authState.user) {
    try {
      await fetchMe();
    } catch {
      await logout();
      router.push("/login");
    }
  }
});

const handleLogout = async () => {
  await logout();
  router.push("/login");
};
</script>

