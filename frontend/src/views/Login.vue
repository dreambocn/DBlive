<!-- 登录页面 -->
<template>
  <section class="panel">
    <h1>{{ t("loginTitle") }}</h1>
    <p>{{ t("loginSubtitle") }}</p>
    <form @submit.prevent="handleLogin">
      <div class="form-group">
        <label>{{ t("username") }}</label>
        <input v-model="username" type="text" autocomplete="username" required />
      </div>
      <div class="form-group">
        <label>{{ t("password") }}</label>
        <input v-model="password" type="password" autocomplete="current-password" required />
      </div>
      <button type="submit" :disabled="loading">
        {{ loading ? "..." : t("login") }}
      </button>
      <p v-if="error" style="color: var(--danger); margin-top: 1rem;">
        {{ error }}
      </p>
    </form>
  </section>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { login } from "../stores/auth";

const { t } = useI18n();
const router = useRouter();
const username = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");

const handleLogin = async () => {
  error.value = "";
  loading.value = true;
  try {
    // 登录成功后跳转到控制台
    await login(username.value, password.value);
    router.push("/app");
  } catch (err) {
    error.value = err.message || "Login failed";
  } finally {
    loading.value = false;
  }
};
</script>

