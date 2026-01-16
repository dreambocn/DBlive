<!-- 录制任务页面 -->
<template>
  <section class="panel panel--wide">
    <header class="panel-header">
      <div>
        <h1>{{ t("recordings") }}</h1>
        <p>{{ t("recordingsSubtitle") }}</p>
      </div>
    </header>

    <form class="record-form" @submit.prevent="createRecording">
      <div class="form-grid">
        <div class="form-group">
          <label>{{ t("uid") }}</label>
          <input v-model="form.uid" type="text" :placeholder="t('uidRequiredPlaceholder')" required />
        </div>
      </div>
      <button type="submit" :disabled="saving">
        {{ saving ? t("loading") : t("createRecording") }}
      </button>
      <span v-if="formError" class="form-error">{{ formError }}</span>
    </form>

    <div v-if="records.length" class="record-list">
      <article v-for="record in records" :key="record.id" class="record-card">
        <div class="record-head">
          <div>
            <h3>{{ record.room_title || t("unknownRoom") }}</h3>
            <p>{{ record.uname || t("unknownUser") }}</p>
          </div>
          <span class="badge">{{ record.status }}</span>
        </div>
        <div class="record-meta">
          <div>{{ t("roomId") }}: {{ record.room_id }}</div>
          <div>{{ t("uid") }}: {{ record.uid || "-" }}</div>
          <div>{{ t("liveStatus") }}: {{ record.live_status }}</div>
        </div>

        <div v-if="editingId === record.id" class="record-edit">
          <div class="form-grid">
            <div class="form-group">
              <label>{{ t("uid") }}</label>
              <input v-model="editForm.uid" type="text" :placeholder="t('uidRequiredPlaceholder')" required />
            </div>
          </div>
          <div class="record-actions">
            <button type="button" @click="saveEdit(record.id)">
              {{ t("save") }}
            </button>
            <button type="button" class="secondary" @click="cancelEdit">
              {{ t("cancel") }}
            </button>
          </div>
        </div>
        <div v-else class="record-actions">
          <button
            type="button"
            :class="record.status === 'recording' ? 'danger' : 'secondary'"
            @click="record.status === 'recording' ? stopRecording(record.id) : startRecording(record.id)"
          >
            {{ record.status === "recording" ? t("stop") : t("start") }}
          </button>
          <button type="button" class="secondary" @click="startEdit(record)">
            {{ t("edit") }}
          </button>
          <button type="button" class="secondary" @click="deleteRecording(record.id)">
            {{ t("delete") }}
          </button>
        </div>
      </article>
    </div>
    <div v-else class="empty-state">{{ t("noRecordings") }}</div>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import api from "../api/client";

const { t } = useI18n();
const records = ref([]);
const saving = ref(false);
const formError = ref("");
const editingId = ref(null);
const editForm = reactive({
  uid: ""
});

const form = reactive({
  uid: ""
});

const loadRecords = async () => {
  // 加载录制任务列表
  records.value = await api.get("/api/v1/recordings");
};

const createRecording = async () => {
  formError.value = "";
  saving.value = true;
  try {
    // 提交创建请求
    const payload = {
      uid: form.uid
    };
    const res = await api.post("/api/v1/recordings", payload);
    // 根据后端返回判断是否创建成功，再刷新列表
    if (res && res.id) {
      await loadRecords();
      form.uid = "";
    } else {
      formError.value = t("createFailed");
    }
  } catch (err) {
    formError.value = err.message || t("createFailed");
  } finally {
    saving.value = false;
  }
};

const startEdit = (record) => {
  editingId.value = record.id;
  editForm.uid = record.uid;
};

const cancelEdit = () => {
  editingId.value = null;
};

const saveEdit = async (id) => {
  saving.value = true;
  try {
    // 提交参数更新
    const payload = {
      uid: editForm.uid
    };
    const res = await api.put(`/api/v1/recordings/${id}`, payload);
    const idx = records.value.findIndex((record) => record.id === id);
    if (idx !== -1) {
      records.value[idx] = res;
    }
    editingId.value = null;
  } catch (err) {
    formError.value = err.message || t("updateFailed");
  } finally {
    saving.value = false;
  }
};

const startRecording = async (id) => {
  saving.value = true;
  try {
    // 调用启动接口
    const res = await api.post(`/api/v1/recordings/${id}/start`, {});
    const idx = records.value.findIndex((record) => record.id === id);
    if (idx !== -1) {
      records.value[idx] = res;
    }
  } catch (err) {
    formError.value = err.message || t("updateFailed");
  } finally {
    saving.value = false;
  }
};

const stopRecording = async (id) => {
  saving.value = true;
  try {
    // 调用停止接口
    const res = await api.post(`/api/v1/recordings/${id}/stop`, {});
    const idx = records.value.findIndex((record) => record.id === id);
    if (idx !== -1) {
      records.value[idx] = res;
    }
  } catch (err) {
    formError.value = err.message || t("updateFailed");
  } finally {
    saving.value = false;
  }
};

const deleteRecording = async (id) => {
  saving.value = true;
  try {
    await api.delete(`/api/v1/recordings/${id}`);
    records.value = records.value.filter((record) => record.id !== id);
  } catch (err) {
    formError.value = err.message || t("updateFailed");
  } finally {
    saving.value = false;
  }
};

onMounted(() => {
  loadRecords();
});
</script>

