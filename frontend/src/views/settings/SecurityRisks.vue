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
          <th>{{ t("settings.status") }}</th>
          <th>{{ t("buttons.actions") }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="risk in risks" :key="risk.path">
          <td>{{ getFileName(risk.path) }}</td>
          <td>{{ getUserName(risk.userId) }}</td>
          <td class="threat-signature">{{ risk.signature || "Unknown" }}</td>
          <td>{{ formatDate(risk.scannedAt) }}</td>
          <td>
            <span v-if="risk.status === 'overridden'" class="status-overridden">
              <i class="material-icons">check_circle</i>
              {{ t("settings.overridden") }}
            </span>
            <span v-else class="status-risk">
              <i class="material-icons">warning</i>
              {{ t("settings.securityRisk") }}
            </span>
          </td>
          <td class="actions-cell">
            <button
              v-if="risk.status !== 'overridden'"
              @click="overrideRisk(risk)"
              class="button button--flat button--primary"
              :aria-label="t('settings.markAsFalsePositive')"
              :title="t('settings.markAsFalsePositive')"
            >
              <i class="material-icons">verified_user</i>
            </button>
            <button
              @click="quarantineRisk(risk)"
              class="button button--flat button--warning"
              :aria-label="t('settings.quarantine')"
              :title="t('settings.quarantine')"
            >
              <i class="material-icons">archive</i>
            </button>
            <button
              @click="deleteRisk(risk)"
              class="button button--flat button--danger"
              :aria-label="t('buttons.delete')"
              :title="t('buttons.delete')"
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
    await scannerApi.deleteSecurityRisk(risk.path, "delete");
    risks.value = risks.value.filter((r) => r.path !== risk.path);
  } catch (error) {
    console.error("Failed to delete security risk:", error);
    alert(t("prompts.error"));
  }
};

const quarantineRisk = async (risk: ScanInfo) => {
  if (!confirm(t("settings.quarantineConfirm", { name: getFileName(risk.path) }))) {
    return;
  }

  try {
    await scannerApi.deleteSecurityRisk(risk.path, "quarantine");
    // Reload the list after quarantine
    await loadRisks();
  } catch (error) {
    console.error("Failed to quarantine file:", error);
    alert(t("prompts.error"));
  }
};

const overrideRisk = async (risk: ScanInfo) => {
  if (!confirm(t("settings.overrideConfirm", { name: getFileName(risk.path) }))) {
    return;
  }

  try {
    await scannerApi.overrideSecurityRisk(risk.path);
    // Reload the list to show updated status
    await loadRisks();
  } catch (error) {
    console.error("Failed to override security risk:", error);
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

.actions-cell {
  display: flex;
  gap: 0.5rem;
}

.button--danger {
  color: #f44336;
}

.button--danger:hover {
  background: rgba(244, 67, 54, 0.1);
}

.button--warning {
  color: #ff9800;
}

.button--warning:hover {
  background: rgba(255, 152, 0, 0.1);
}

.button--primary {
  color: #2196f3;
}

.button--primary:hover {
  background: rgba(33, 150, 243, 0.1);
}

.status-overridden {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: #4caf50;
}

.status-risk {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: #f44336;
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
