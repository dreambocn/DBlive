<!-- 主题切换组件 -->
<template>
  <button class="toggle" @click="toggleTheme">
    {{ isDark ? "🌙" : "☀️" }}
  </button>
</template>

<script setup>
import { ref, onMounted } from "vue";

const isDark = ref(false);

const applyTheme = (value) => {
  // 切换html根节点主题类并保存
  isDark.value = value;
  document.documentElement.classList.toggle("dark", value);
  localStorage.setItem("dblive-theme", value ? "dark" : "light");
};

const toggleTheme = () => {
  applyTheme(!isDark.value);
};

onMounted(() => {
  // 读取本地主题设置
  const saved = localStorage.getItem("dblive-theme") || "light";
  applyTheme(saved === "dark");
});
</script>

