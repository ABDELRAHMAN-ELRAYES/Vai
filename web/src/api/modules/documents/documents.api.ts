import { apiClient } from "@/api/client";
import type { UploadResponse } from "@/types/modules/documents/dto";

const ROUTER_PREFIX = "/documents";

export const documentsApi = {
  upload(file: File) {
    const formData = new FormData();
    formData.append("file", file);

    return apiClient.post<UploadResponse>(`${ROUTER_PREFIX}/upload`, formData);
  },
};
