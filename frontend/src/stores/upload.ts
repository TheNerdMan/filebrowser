import { defineStore } from "pinia";
import { useFileStore } from "./file";
import { files as api } from "@/api";
import * as scannerApi from "@/api/scanner";
import buttons from "@/utils/buttons";
import { computed, inject, markRaw, ref } from "vue";
import * as tus from "@/api/tus";

// TODO: make this into a user setting
const UPLOADS_LIMIT = 5;

const beforeUnload = (event: Event) => {
  event.preventDefault();
  // To remove >> is deprecated
  // event.returnValue = "";
};

export const useUploadStore = defineStore("upload", () => {
  const $showError = inject<IToastError>("$showError")!;

  let progressInterval: number | null = null;

  //
  // STATE
  //

  const allUploads = ref<Upload[]>([]);
  const activeUploads = ref<Set<Upload>>(new Set());
  const lastUpload = ref<number>(-1);
  const totalBytes = ref<number>(0);
  const sentBytes = ref<number>(0);

  //
  // ACTIONS
  //

  const upload = (
    path: string,
    name: string,
    file: File | null,
    overwrite: boolean,
    type: ResourceType
  ) => {
    if (!hasActiveUploads() && !hasPendingUploads()) {
      window.addEventListener("beforeunload", beforeUnload);
      buttons.loading("upload");
    }

    const upload: Upload = {
      path,
      name,
      file,
      overwrite,
      type,
      totalBytes: file?.size || 1,
      sentBytes: 0,
      // Stores rapidly changing sent bytes value without causing component re-renders
      rawProgress: markRaw({
        sentBytes: 0,
      }),
      scanStatus: "uploading",
    };

    totalBytes.value += upload.totalBytes;
    allUploads.value.push(upload);

    processUploads();
  };

  const abort = () => {
    // Resets the state by preventing the processing of the remaning uploads
    lastUpload.value = Infinity;
    tus.abortAllUploads();
  };

  //
  // GETTERS
  //

  const pendingUploadCount = computed(
    () =>
      allUploads.value.length -
      (lastUpload.value + 1) +
      activeUploads.value.size
  );

  //
  // PRIVATE FUNCTIONS
  //

  const hasActiveUploads = () => activeUploads.value.size > 0;

  const hasPendingUploads = () =>
    allUploads.value.length > lastUpload.value + 1;

  const isActiveUploadsOnLimit = () => activeUploads.value.size < UPLOADS_LIMIT;

  const processUploads = async () => {
    if (!hasActiveUploads() && !hasPendingUploads()) {
      const fileStore = useFileStore();
      window.removeEventListener("beforeunload", beforeUnload);
      buttons.success("upload");
      reset();
      fileStore.reload = true;
    }

    if (isActiveUploadsOnLimit() && hasPendingUploads()) {
      if (!hasActiveUploads()) {
        // Update the state in a fixed time interval
        progressInterval = window.setInterval(syncState, 1000);
      }

      const upload = nextUpload();

      if (upload.type === "dir") {
        await api.post(upload.path).catch($showError);
      } else {
        const onUpload = (event: ProgressEvent) => {
          upload.rawProgress.sentBytes = event.loaded;
        };

        await api
          .post(upload.path, upload.file!, upload.overwrite, onUpload)
          .catch((err) => err.message !== "Upload aborted" && $showError(err));
      }

      finishUpload(upload);
    }
  };

  const nextUpload = (): Upload => {
    lastUpload.value++;

    const upload = allUploads.value[lastUpload.value];
    activeUploads.value.add(upload);

    return upload;
  };

  const finishUpload = (upload: Upload) => {
    sentBytes.value += upload.totalBytes - upload.sentBytes;
    upload.sentBytes = upload.totalBytes;
    upload.file = null;

    // Start polling for scan status if not a directory
    if (upload.type !== "dir") {
      upload.scanStatus = "scanning";
      pollScanStatus(upload);
    }

    activeUploads.value.delete(upload);
    processUploads();
  };

  const pollScanStatus = async (upload: Upload) => {
    // Poll for scan status up to 10 times (20 seconds)
    let attempts = 0;
    const maxAttempts = 10;
    const pollInterval = 2000; // 2 seconds

    const poll = async () => {
      try {
        const status = await scannerApi.getScanStatus(upload.path);
        upload.scanStatus = status.status;

        // If still scanning and haven't exceeded max attempts, continue polling
        if (status.status === "scanning" && attempts < maxAttempts) {
          attempts++;
          setTimeout(poll, pollInterval);
        } else if (status.status === "clean" || status.status === "security_risk" || status.status === "scan_error") {
          // Scan complete
          if (status.status === "security_risk") {
            $showError(new Error(`File ${upload.name} flagged as security risk: ${status.signature || "unknown threat"}`));
          }
        }
      } catch (error) {
        // If scan status check fails, mark as error instead of assuming clean
        upload.scanStatus = "scan_error";
        console.error("Failed to check scan status:", error);
      }
    };

    // Start polling after a short delay to give the backend time to start scanning
    setTimeout(poll, 1000);
  };

  const syncState = () => {
    for (const upload of activeUploads.value) {
      sentBytes.value += upload.rawProgress.sentBytes - upload.sentBytes;
      upload.sentBytes = upload.rawProgress.sentBytes;
    }
  };

  const reset = () => {
    if (progressInterval !== null) {
      clearInterval(progressInterval);
      progressInterval = null;
    }

    allUploads.value = [];
    activeUploads.value = new Set();
    lastUpload.value = -1;
    totalBytes.value = 0;
    sentBytes.value = 0;
  };

  return {
    // STATE
    activeUploads,
    totalBytes,
    sentBytes,

    // ACTIONS
    upload,
    abort,

    // GETTERS
    pendingUploadCount,
  };
});
