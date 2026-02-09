import { fetchURL } from "./utils";

export interface ScanInfo {
  path: string;
  userId: number;
  status: "uploading" | "scanning" | "clean" | "security_risk" | "scan_error";
  signature?: string;
  scannedAt?: string;
  uploadedAt: string;
}

export async function getScanStatus(path: string): Promise<ScanInfo> {
  const url = `/api/scan/status?path=${encodeURIComponent(path)}`;
  return fetchURL(url, {});
}

export async function getSecurityRisks(): Promise<ScanInfo[]> {
  return fetchURL("/api/scan/risks", {});
}

export async function deleteSecurityRisk(path: string): Promise<void> {
  const url = `/api/scan/risks?path=${encodeURIComponent(path)}`;
  await fetchURL(url, {
    method: "DELETE",
  });
}
