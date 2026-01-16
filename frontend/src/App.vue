<!-- 应用根组件与全局布局 -->
<template>
  <div class="app">
    <!-- 顶部栏包含品牌与主题/语言切换 -->
    <header class="topbar">
      <div class="brand">DBlive</div>
      <div class="topbar-actions">
        <LanguageToggle />
        <ThemeToggle />
      </div>
    </header>
    <main :class="['content', isDashboard ? 'content--dashboard' : 'content--center']">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";
import ThemeToggle from "./components/ThemeToggle.vue";
import LanguageToggle from "./components/LanguageToggle.vue";

const route = useRoute();
const isDashboard = computed(() => route.path.startsWith("/app"));
</script>

