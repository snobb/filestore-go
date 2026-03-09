const API_BASE = "";

function getStoredToken() {
  return localStorage.getItem("auth_token");
}

export interface Document {
  id: string;
  user_id: string;
  file_name: string;
  file_path: string;
  file_size: number;
  content_type: string;
  checksum: string;
  status: string;
  uploaded_at: string;
  updated_at: string;
}

export interface UploadPendingResponse {
  id: string;
  upload_url: string;
  status_url: string;
}

export interface FileInfo {
  path: string;
  check_sum: string;
  file_size: number;
}

export async function requestRaw<T>(
  url: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getStoredToken();

  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token && { "Authorization": `Bearer ${token}` }),
      ...options.headers,
    },
  });
  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: "Request failed" }));
    throw new Error(error.error || "Request failed");
  }
  return response.json();
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getStoredToken();

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token && { "Authorization": `Bearer ${token}` }),
      ...options.headers,
    },
  });
  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: "Request failed" }));
    throw new Error(error.error || "Request failed");
  }
  return response.json();
}

export async function listDocuments(): Promise<Document[]> {
  return request<Document[]>("/api/documents");
}

export async function getDocument(id: string): Promise<Document> {
  return request<Document>(`/api/documents/${id}`);
}

export async function uploadPending(
  fileName: string,
  contentType: string,
): Promise<UploadPendingResponse> {
  return request<UploadPendingResponse>("/api/documents", {
    method: "POST",
    body: JSON.stringify({ file_name: fileName, content_type: contentType }),
  });
}

export async function updateDocumentStatus(
  status_url: string,
  status: string,
  fileSize: number,
  checksum: string,
): Promise<Document> {
  return request<Document>(status_url, {
    method: "PATCH",
    body: JSON.stringify({ status, file_size: fileSize, checksum }),
  });
}

export async function uploadFile(
  uploadUrl: string,
  file: File,
): Promise<FileInfo> {
  const token = getStoredToken();

  const response = await fetch(`${API_BASE}${uploadUrl}`, {
    method: "POST",
    headers: {
      ...(token && { "Authorization": `Bearer ${token}` }),
    },
    body: file,
  });
  if (!response.ok) {
    throw new Error("File upload failed");
  }
  return response.json();
}

export function getDownloadUrl(path: string): string {
  return `${API_BASE}/file_store/downloads/${encodeURIComponent(path)}`;
}
