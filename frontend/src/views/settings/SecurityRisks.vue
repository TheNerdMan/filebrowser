<template>
  <div>
    <h2 class="title">{{ t("settings.securityRisks") }}</h2>
    <p class="small">{{ t("settings.securityRisksHelp") }}</p>

    <div v-if="loading" class="loading-spinner">
      <i class="material-icons spinning">refresh</i>
      {{ t("files.loading") }}
    </div>

    <div v-else-if="risks.length === 0" class="message">
      <i class="material-icons">check_circle</i>
      <p>{{ t("settings.noSecurityRisks") }}</p>
    </div>

    <table v-else>
      <thead>
        <tr>
          <th>{{ t("files.name") }}</th>
          <th>{{ t("settings.user") }}</th>
          <th>{{ t("settings.threat") }}</th>
          <th>{{ t("prompts.lastModified") }}</th>
          <th>{{ t("buttons.delete") }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="risk in risks" :key="risk.path">
          <td>{{ getFileName(risk.path) }}</td>
          <td>{{ getUserName(risk.userId) }}</td>
          <td class="threat-signature">{{ risk.signature || "Unknown" }}</td>
          <td>{{ formatDate(risk.scannedAt) }}</td>
          <td>
            <button
              @click="deleteRisk(risk)"
              class="button button--flat button--danger"
              :aria-label="t('buttons.delete')"
            >
              <i class="material-icons">delete</i>
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import * as scannerApi from "@/api/scanner";
import type { ScanInfo } from "@/api/scanner";
import { useAuthStore } from "@/stores/auth";

const { t } = useI18n();
const authStore = useAuthStore();

const loading = ref(true);
const risks = ref<ScanInfo[]>([]);

const loadRisks = async () => {
  loading.value = true;
  try {
    risks.value = await scannerApi.getSecurityRisks();
  } catch (error) {
    console.error("Failed to load security risks:", error);
  } finally {
    loading.value = false;
  }
};

const deleteRisk = async (risk: ScanInfo) => {
  if (!confirm(t("settings.deleteSecurityRiskConfirm", { name: getFileName(risk.path) }))) {
    return;
  }

  try {
    await scannerApi.deleteSecurityRisk(risk.path);
    risks.value = risks.value.filter((r) => r.path !== risk.path);
  } catch (error) {
    console.error("Failed to delete security risk:", error);
    alert(t("prompts.error"));
  }
};

const getFileName = (path: string): string => {
  const parts = path.split("/");
  return parts[parts.length - 1] || path;
};

const getUserName = (userId: number): string => {
  // TODO: Fetch user name from user store
  return `User ${userId}`;
};

const formatDate = (dateString: string | undefined): string => {
  if (!dateString) return "-";
  const date = new Date(dateString);
  return date.toLocaleString();
};

onMounted(() => {
  loadRisks();
});
</script>

<style scoped>
.title {
  margin-bottom: 1rem;
}

.small {
  color: #666;
  margin-bottom: 2rem;
}

.loading-spinner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem;
  justify-content: center;
}

.message {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 3rem;
  color: #4caf50;
}

.message i {
  font-size: 4rem;
}

table {
  width: 100%;
  border-collapse: collapse;
}

thead {
  background: #f5f5f5;
}

th,
td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #e0e0e0;
}

.threat-signature {
  font-family: monospace;
  color: #f44336;
  font-weight: 500;
}

.button--danger {
  color: #f44336;
}

.button--danger:hover {
  background: rgba(244, 67, 54, 0.1);
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.spinning {
  animation: spin 2s linear infinite;
}
</style>
