<template>
  <base-modal>
    <template #title>
      <h2>{{ $t("prompts.requestGlobalShare") }}</h2>
    </template>

    <template #content>
      <p>{{ $t("prompts.requestGlobalShareMessage") }}</p>
      <p>
        <strong>{{ $t("prompts.path") }}:</strong> {{ req?.path || "" }}
      </p>
      <textarea
        v-model="message"
        :placeholder="$t('prompts.messageOptional')"
        rows="4"
        style="width: 100%"
      ></textarea>
    </template>

    <template #action>
      <button
        class="button button--flat"
        @click="closeHovers"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat"
        @click="requestShare"
        :aria-label="$t('buttons.request')"
        :title="$t('buttons.request')"
      >
        {{ $t("buttons.request") }}
      </button>
    </template>
  </base-modal>
</template>

<script setup lang="ts">
import BaseModal from "./BaseModal.vue";
import { useLayoutStore } from "@/stores/layout";
import { globalshare as api } from "@/api";
import { inject, ref } from "vue";
import { useI18n } from "vue-i18n";

const layoutStore = useLayoutStore();
const { t } = useI18n();

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const req = layoutStore.currentPrompt;
const message = ref("");

const closeHovers = () => {
  layoutStore.closeHovers();
};

const requestShare = async () => {
  if (!req || !req.path) {
    $showError(new Error("No path specified"));
    return;
  }

  try {
    await api.requestGlobalShare(req.path, message.value);
    $showSuccess(t("success.globalShareRequested"));
    layoutStore.closeHovers();
  } catch (err) {
    if (err instanceof Error) {
      $showError(err);
    }
  }
};
</script>
