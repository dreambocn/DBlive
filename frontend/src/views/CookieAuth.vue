<!-- Cookie授权页面 -->
<template>
  <section class="panel panel--wide">
    <header class="panel-header">
      <div>
        <h1>{{ t("cookieAuth") }}</h1>
        <p>{{ t("cookieAuthSubtitle") }}</p>
      </div>
      <button class="secondary" :disabled="loading" @click="generateQRCode">
        {{ loading ? t("loading") : t("refreshQRCode") }}
      </button>
    </header>

    <div class="cookie-grid">
      <div class="cookie-status">
        <div class="status-item">
          <span>{{ t("cookieStatus") }}</span>
          <strong>{{ statusLabel }}</strong>
        </div>
        <div class="status-item">
          <span>{{ t("cookieUser") }}</span>
          <strong>{{ statusUser }}</strong>
        </div>
        <div class="status-item">
          <span>{{ t("cookieMessage") }}</span>
          <strong>{{ statusMessage || "-" }}</strong>
        </div>
        <div class="status-avatar" v-if="statusUserFace">
          <img :src="statusUserFace" :alt="statusUser || t('cookieUser')" />
        </div>
      </div>

      <div class="cookie-qr">
        <div class="qr-frame">
          <img v-if="qrImage" :src="qrImage" :alt="t('cookieQR')" />
          <div v-else class="qr-placeholder">{{ t("cookieQRPlaceholder") }}</div>
        </div>
        <p class="qr-hint">{{ t("cookieQRHint") }}</p>
        <button :disabled="!qrcodeKey || polling" @click="pollOnce">
          {{ polling ? t("polling") : t("checkStatus") }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import api from "../api/client";

const { t } = useI18n();
const loading = ref(false);
const polling = ref(false);
const pollInFlight = ref(false);
const qrcodeKey = ref("");
const qrImage = ref("");
const status = ref("missing");
const statusMessage = ref("");
const statusUser = ref("");
const statusUserFace = ref("");
let pollTimer = null;

const statusLabel = computed(() => {
  switch (status.value) {
    case "active":
      return t("cookieActive");
    case "success":
      return t("cookieActive");
    case "invalid":
      return t("cookieInvalid");
    case "expired":
      return t("cookieExpired");
    case "pending":
      return t("cookiePending");
    default:
      return t("cookieMissing");
  }
});

const fetchStatus = async () => {
  // 页面初始化时拉取Cookie状态
  const res = await api.get("/api/v1/bilibili/cookie/status");
  status.value = res.status;
  statusMessage.value = res.message || "";
  statusUser.value = res.user?.uname || "";
  statusUserFace.value = res.user?.face || "";
};

const generateQRCode = async () => {
  loading.value = true;
  try {
    // 请求后端生成二维码
    const res = await api.post("/api/v1/bilibili/cookie/qrcode", {});
    qrcodeKey.value = res.qrcode_key;
    const encoded = encodeURIComponent(res.qr_url);
    qrImage.value = `https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=${encoded}`;
    status.value = "pending";
    statusMessage.value = t("cookieWaiting");
    statusUser.value = "";
    startPolling();
  } catch (err) {
    statusMessage.value = err.message || t("cookieFailed");
  } finally {
    loading.value = false;
  }
};

const pollOnce = async () => {
  if (!qrcodeKey.value) return;
  if (pollInFlight.value) return;
  pollInFlight.value = true;
  try {
    // 轮询二维码状态并更新UI
    const res = await api.post("/api/v1/bilibili/cookie/poll", {
      qrcode_key: qrcodeKey.value
    });
    status.value = res.status;
    statusMessage.value = res.message || "";
    statusUser.value = res.user?.uname || "";
    statusUserFace.value = res.user?.face || "";
    if (res.status === "success" || res.status === "expired") {
      stopPolling();
    }
  } catch (err) {
    statusMessage.value = err.message || t("cookieFailed");
    stopPolling();
  } finally {
    pollInFlight.value = false;
  }
};

const startPolling = () => {
  stopPolling();
  polling.value = true;
  // 进入轮询循环
  pollTimer = setInterval(pollOnce, 2000);
  pollOnce();
};

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
  polling.value = false;
};

onMounted(async () => {
  try {
    // 进入页面时检查授权状态
    await fetchStatus();
  } catch {
    status.value = "missing";
  }
});

onUnmounted(() => {
  stopPolling();
});
</script>

