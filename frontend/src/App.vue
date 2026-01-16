<template>
  <div class="app">
    <header class="topbar">
      <div class="brand">DBlive</div>
      <div class="topbar-actions">
        <LanguageToggle />
        <ThemeToggle />
      </div>
    </header>
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import ThemeToggle from "./components/ThemeToggle.vue";
import LanguageToggle from "./components/LanguageToggle.vue";
</script>
