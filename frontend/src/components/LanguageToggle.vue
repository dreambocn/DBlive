<template>
  <button class="toggle" @click="toggleLang">
    {{ locale === "zh-CN" ? "中文" : "EN" }}
  </button>
</template>

<script setup>
import { useI18n } from "vue-i18n";

const { locale } = useI18n();

const toggleLang = () => {
  locale.value = locale.value === "zh-CN" ? "en-US" : "zh-CN";
  localStorage.setItem("dblive-lang", locale.value);
};
</script>
