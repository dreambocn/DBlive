<!-- 录制设置页面 -->
<template>
  <section class="panel panel--wide">
    <header class="panel-header">
      <div>
        <h1>{{ t("settings") }}</h1>
        <p>{{ t("settingsSubtitle") }}</p>
      </div>
    </header>

    <form class="record-form" @submit.prevent="saveSettings">
      <div class="form-grid">
        <div class="form-group">
          <label>{{ t("outputDir") }}</label>
          <input v-model="form.outputDir" type="text" :placeholder="t('outputDirPlaceholder')" />
        </div>
        <div class="form-group">
          <label>{{ t("defaultSegmentTime") }}</label>
          <input v-model.number="form.defaultSegmentTimeMin" type="number" min="1" />
        </div>
        <div class="form-group">
          <label>{{ t("defaultQuality") }}</label>
          <input v-model.number="form.defaultQuality" type="number" min="1" />
        </div>
        <div class="form-group">
          <label>{{ t("defaultSegmentSize") }}</label>
          <input v-model.number="form.defaultSegmentSizeMB" type="number" min="1" />
        </div>
      </div>
      <button type="submit" :disabled="saving">
        {{ saving ? t("loading") : t("saveSettings") }}
      </button>
      <span v-if="formError" class="form-error">{{ formError }}</span>
      <span v-else-if="saved" class="form-success">{{ t("saved") }}</span>
    </form>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import api from "../api/client";

const { t } = useI18n();
const saving = ref(false);
const saved = ref(false);
const formError = ref("");

const form = reactive({
  outputDir: "",
  defaultSegmentTimeMin: 30,
  defaultQuality: 10000,
  defaultSegmentSizeMB: 20
});

const loadSettings = async () => {
  try {
    // 拉取当前用户设置
    const res = await api.get("/api/v1/settings");
    form.outputDir = res.output_dir || "";
    form.defaultSegmentTimeMin = res.default_segment_time_min || 30;
    form.defaultQuality = res.default_quality || 10000;
    form.defaultSegmentSizeMB = res.default_segment_size_mb || 20;
  } catch (err) {
    formError.value = err.message || t("updateFailed");
  }
};

const saveSettings = async () => {
  saving.value = true;
  formError.value = "";
  saved.value = false;
  try {
    // 提交设置更新
    const payload = {
      output_dir: form.outputDir,
      default_segment_time_min: Number(form.defaultSegmentTimeMin),
      default_quality: Number(form.defaultQuality),
      default_segment_size_mb: Number(form.defaultSegmentSizeMB)
    };
    await api.put("/api/v1/settings", payload);
    saved.value = true;
  } catch (err) {
    formError.value = err.message || t("updateFailed");
  } finally {
    saving.value = false;
  }
};

onMounted(() => {
  loadSettings();
});
</script>

