<!-- 旧版仪表盘页面 -->
<template>
  <div class="card">
    <h2>{{ t("dashboardTitle") }}</h2>
    <p>{{ t("dashboardSubtitle") }}</p>

    <div class="stat">
      <span>{{ t("user") }}</span>
      <strong>{{ user?.username || "-" }}</strong>
    </div>
    <div class="stat">
      <span>{{ t("accessToken") }}</span>
      <strong>{{ accessToken ? "Active" : "Missing" }}</strong>
    </div>
    <div class="stat">
      <span>{{ t("refreshToken") }}</span>
      <strong>{{ refreshToken ? "Active" : "Missing" }}</strong>
    </div>

    <button class="secondary" @click="handleLogout">
      {{ t("logout") }}
    </button>
  </div>
</template>

<script setup>
import { computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { authState, fetchMe, logout } from "../stores/auth";

const { t } = useI18n();
const router = useRouter();
const accessToken = computed(() => authState.accessToken);
const refreshToken = computed(() => authState.refreshToken);
const user = computed(() => authState.user);

onMounted(async () => {
  // 进入页面时补齐用户信息
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

