<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading">
    <div class="column">
      <div class="card">
        <div class="card-title">
          <h2>{{ t("settings.globalShareManagement") }}</h2>
        </div>

        <div class="card-content">
          <h3>{{ t("settings.pendingRequests") }}</h3>
          <div v-if="pendingRequests.length > 0">
            <table>
              <tr>
                <th>{{ t("settings.username") }}</th>
                <th>{{ t("settings.path") }}</th>
                <th>{{ t("settings.message") }}</th>
                <th>{{ t("settings.requestedAt") }}</th>
                <th></th>
              </tr>
              <tr v-for="req in pendingRequests" :key="req.id">
                <td>{{ req.username }}</td>
                <td>{{ req.path }}</td>
                <td>{{ req.message || "-" }}</td>
                <td>{{ humanTime(req.createdAt) }}</td>
                <td>
                  <button
                    class="button button--flat"
                    @click="approveRequest(req.id)"
                    :title="t('buttons.approve')"
                  >
                    <i class="material-icons">check</i>
                    {{ t("buttons.approve") }}
                  </button>
                  <button
                    class="button button--flat"
                    @click="rejectRequest(req.id)"
                    :title="t('buttons.reject')"
                  >
                    <i class="material-icons">close</i>
                    {{ t("buttons.reject") }}
                  </button>
                </td>
              </tr>
            </table>
          </div>
          <h2 class="message" v-else>
            <i class="material-icons">inbox</i>
            <span>{{ t("settings.noPendingRequests") }}</span>
          </h2>

          <h3 style="margin-top: 2em">
            {{ t("settings.activeGlobalShares") }}
          </h3>
          <div v-if="globalShares.length > 0">
            <table>
              <tr>
                <th>{{ t("settings.username") }}</th>
                <th>{{ t("settings.path") }}</th>
                <th>{{ t("settings.addedAt") }}</th>
                <th></th>
              </tr>
              <tr v-for="share in globalShares" :key="share.id">
                <td>{{ share.username }}</td>
                <td>{{ share.path }}</td>
                <td>{{ humanTime(share.addedAt) }}</td>
                <td class="small">
                  <button
                    class="action"
                    @click="deleteGlobalShare($event, share.id)"
                    :aria-label="t('buttons.delete')"
                    :title="t('buttons.delete')"
                  >
                    <i class="material-icons">delete</i>
                  </button>
                </td>
              </tr>
            </table>
          </div>
          <h2 class="message" v-else>
            <i class="material-icons">sentiment_dissatisfied</i>
            <span>{{ t("settings.noGlobalShares") }}</span>
          </h2>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { globalshare as api } from "@/api";
import dayjs from "dayjs";
import Errors from "@/views/Errors.vue";
import { inject, ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import type { GlobalShareRequest, GlobalShare } from "@/api/globalshare";

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const { t } = useI18n();

const layoutStore = useLayoutStore();

const error = ref<StatusError | null>(null);
const pendingRequests = ref<GlobalShareRequest[]>([]);
const globalShares = ref<GlobalShare[]>([]);

const loadData = async () => {
  layoutStore.loading = true;

  try {
    const requests = await api.listGlobalShareRequests();
    pendingRequests.value = requests.filter((r) => r.status === "pending");

    const shares = await api.listGlobalShares();
    globalShares.value = shares;
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
};

onMounted(loadData);

const humanTime = (time: number) => {
  return dayjs(time * 1000).fromNow();
};

const approveRequest = async (id: number) => {
  try {
    await api.approveGlobalShareRequest(id);
    $showSuccess(t("success.requestApproved"));
    await loadData();
  } catch (err) {
    if (err instanceof Error) {
      $showError(err);
    }
  }
};

const rejectRequest = async (id: number) => {
  try {
    await api.rejectGlobalShareRequest(id);
    $showSuccess(t("success.requestRejected"));
    await loadData();
  } catch (err) {
    if (err instanceof Error) {
      $showError(err);
    }
  }
};

const deleteGlobalShare = async (event: Event, id: number) => {
  event.preventDefault();

  layoutStore.showHover({
    prompt: "global-share-delete",
    confirm: async () => {
      layoutStore.closeHovers();

      try {
        await api.deleteGlobalShare(id);
        globalShares.value = globalShares.value.filter((s) => s.id !== id);
        $showSuccess(t("success.globalShareDeleted"));
      } catch (err) {
        if (err instanceof Error) {
          $showError(err);
        }
      }
    },
  });
};
</script>
