import type { Photo, UploadResponse } from '../types/photo';

const API_BASE_URL = 'http://localhost:8080/api';

export async function uploadPhotos(files: File[]): Promise<UploadResponse> {
  const formData = new FormData();
  
  files.forEach((file) => {
    formData.append('photos', file);
  });

  const response = await fetch(`${API_BASE_URL}/photos/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to upload photos');
  }

  return response.json();
}

export async function getPhotos(): Promise<Photo[]> {
  const response = await fetch(`${API_BASE_URL}/photos`);

  if (!response.ok) {
    throw new Error('Failed to fetch photos');
  }

  return response.json();
}

export function getPhotoUrl(path: string): string {
  return `http://localhost:8080${path}`;
}
