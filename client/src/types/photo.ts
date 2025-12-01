export interface Photo {
  id: string;
  filename: string;
  path: string;
  size: number;
  uploadedAt: string;
}

export interface UploadResponse {
  success: boolean;
  message: string;
  photos?: Photo[];
}

export interface ErrorResponse {
  success: boolean;
  error: string;
}
