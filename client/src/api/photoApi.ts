import type { Photo, UploadResponse, ClusterResponse, PageDraft } from '../types/photo';

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

export async function analyzePhotos(photoIds: string[]): Promise<ClusterResponse> {
  const response = await fetch(`${API_BASE_URL}/photos/cluster`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ photoIds }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to analyze photos');
  }

  return response.json();
}

export async function savePageDraft(draft: PageDraft): Promise<PageDraft> {
  const response = await fetch(`${API_BASE_URL}/drafts/${draft.id}/approve`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(draft),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to save page');
  }

  return response.json();
}

export async function getPages(): Promise<PageDraft[]> {
  const response = await fetch(`${API_BASE_URL}/drafts/`);

  if (!response.ok) {
    throw new Error('Failed to fetch pages');
  }

  const drafts: PageDraft[] = await response.json();
  // Filter to only return approved drafts as "pages"
  return drafts.filter(d => d.status === 'approved');
}

export function getPhotoUrl(path: string): string {
  return `http://localhost:8080${path}`;
}
