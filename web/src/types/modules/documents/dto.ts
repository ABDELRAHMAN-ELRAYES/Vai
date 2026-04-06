export interface UploadResponse {
  success: boolean;
  message: string;
  data: {
    id: string;
    original_name: string;
    name: string;
    size: number;
    mime_type: string;
  };
}
