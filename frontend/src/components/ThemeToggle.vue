<template>
  <button class="toggle" @click="toggleTheme">
    {{ isDark ? "🌙" : "☀️" }}
  </button>
</template>

<script setup>
import { ref, onMounted } from "vue";

const isDark = ref(false);

const applyTheme = (value) => {
  isDark.value = value;
  document.documentElement.classList.toggle("dark", value);
  localStorage.setItem("dblive-theme", value ? "dark" : "light");
};

const toggleTheme = () => {
  applyTheme(!isDark.value);
};

onMounted(() => {
  const saved = localStorage.getItem("dblive-theme") || "light";
  applyTheme(saved === "dark");
});
</script>
