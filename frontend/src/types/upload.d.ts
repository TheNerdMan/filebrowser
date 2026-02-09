type Upload = {
  path: string;
  name: string;
  file: File | null;
  type: ResourceType;
  overwrite: boolean;
  totalBytes: number;
  sentBytes: number;
  rawProgress: {
    sentBytes: number;
  };
  scanStatus?: "uploading" | "scanning" | "clean" | "security_risk" | "scan_error";
};

interface UploadEntry {
  name: string;
  size: number;
  isDir: boolean;
  fullPath?: string;
  file?: File;
}

type UploadList = UploadEntry[];
