import { fetchURL, fetchJSON } from "./utils";

export interface GlobalShareRequest {
  id: number;
  path: string;
  userID: number;
  username?: string;
  status: string;
  createdAt: number;
  message?: string;
}

export interface GlobalShare {
  id: number;
  path: string;
  originalPath: string;
  userID: number;
  username?: string;
  addedAt: number;
  requestID?: number;
}

// Request to add a file/folder to global share
export async function requestGlobalShare(path: string, message?: string) {
  return fetchJSON<GlobalShareRequest>("/api/globalshare/request", {
    method: "POST",
    body: JSON.stringify({ path, message }),
  });
}

// Get all global share requests (admin only)
export async function listGlobalShareRequests() {
  return fetchJSON<GlobalShareRequest[]>("/api/globalshare/requests");
}

// Get user's own global share requests
export async function getMyRequests() {
  return fetchJSON<GlobalShareRequest[]>("/api/globalshare/myrequests");
}

// Approve a global share request (admin only)
export async function approveGlobalShareRequest(id: number) {
  return fetchJSON<GlobalShareRequest>(
    `/api/globalshare/requests/${id}/approve`,
    {
      method: "POST",
    }
  );
}

// Reject a global share request (admin only)
export async function rejectGlobalShareRequest(id: number) {
  return fetchJSON<GlobalShareRequest>(
    `/api/globalshare/requests/${id}/reject`,
    {
      method: "POST",
    }
  );
}

// Delete a global share request
export async function deleteGlobalShareRequest(id: number) {
  await fetchURL(`/api/globalshare/requests/${id}`, {
    method: "DELETE",
  });
}

// Get all global shares
export async function listGlobalShares() {
  return fetchJSON<GlobalShare[]>("/api/globalshare");
}

// Delete a global share (admin only)
export async function deleteGlobalShare(id: number) {
  await fetchURL(`/api/globalshare/${id}`, {
    method: "DELETE",
  });
}
