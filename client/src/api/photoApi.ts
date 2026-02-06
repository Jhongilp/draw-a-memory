import type { Photo, UploadResponse, ClusterResponse, PageDraft } from '../types/photo';

// API configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

// Token getter function - will be set by the hook
let getAuthToken: (() => Promise<string | null>) | null = null;

/**
 * Initialize the API with a token getter function from Clerk
 * Call this from a component that has access to useAuth()
 */
export function initializeApi(tokenGetter: () => Promise<string | null>) {
  getAuthToken = tokenGetter;
}

/**
 * Get authorization headers for API requests
 */
async function getAuthHeaders(): Promise<HeadersInit> {
  if (!getAuthToken) {
    console.warn('API not initialized with auth token getter');
    return {};
  }

  const token = await getAuthToken();
  if (!token) {
    throw new Error('Not authenticated');
  }

  return {
    'Authorization': `Bearer ${token}`,
  };
}

/**
 * Make an authenticated fetch request
 */
async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const authHeaders = await getAuthHeaders();
  
  const response = await fetch(url, {
    ...options,
    headers: {
      ...authHeaders,
      ...options.headers,
    },
  });

  if (response.status === 401) {
    throw new Error('Session expired. Please sign in again.');
  }

  return response;
}

export async function uploadPhotos(files: File[]): Promise<UploadResponse> {
  const formData = new FormData();
  
  files.forEach((file) => {
    formData.append('photos', file);
  });

  const response = await authFetch(`${API_BASE_URL}/photos/upload`, {
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
  const response = await authFetch(`${API_BASE_URL}/photos`);

  if (!response.ok) {
    throw new Error('Failed to fetch photos');
  }

  const photos = await response.json();
  return photos || [];
}

export async function deletePhoto(photoId: string): Promise<void> {
  const response = await authFetch(`${API_BASE_URL}/photos/${photoId}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete photo');
  }
}

export async function analyzePhotos(photoIds: string[]): Promise<ClusterResponse> {
  const response = await authFetch(`${API_BASE_URL}/photos/cluster`, {
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
  const response = await authFetch(`${API_BASE_URL}/drafts/${draft.id}/approve`, { 
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

export async function updateDraft(draft: PageDraft): Promise<PageDraft> {
  const response = await authFetch(`${API_BASE_URL}/drafts/${draft.id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(draft),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update draft');
  }

  return response.json();
}

export async function deleteDraft(draftId: string): Promise<void> {
  const response = await authFetch(`${API_BASE_URL}/drafts/${draftId}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete draft');
  }
}

export async function getPages(): Promise<PageDraft[]> {
  const response = await authFetch(`${API_BASE_URL}/drafts/`);

  if (!response.ok) {
    throw new Error('Failed to fetch pages');
  }

  const drafts: PageDraft[] = await response.json();
  // Filter to only return approved drafts as "pages"
  return (drafts || []).filter(d => d.status === 'approved');
}

export async function getDrafts(): Promise<PageDraft[]> {
  const response = await authFetch(`${API_BASE_URL}/drafts/`);

  if (!response.ok) {
    throw new Error('Failed to fetch drafts');
  }

  return (await response.json()) || [];
}

/**
 * Get photo URL - now photos come with signed URLs from the server
 * The path is already a full signed URL, so we return it directly
 * 
 * Note: Signed URLs expire after 15 minutes. If you get a 403,
 * refresh the photo list to get new signed URLs.
 */
export function getPhotoUrl(path: string, _thumb: boolean = true): string {
  // If it's already a full URL (signed URL from GCS), return as-is
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  
  // Fallback for local development without GCS
  const baseUrl = `http://localhost:8080${path}`;
  return _thumb ? `${baseUrl}?thumb=1` : baseUrl;
}

/**
 * Check if a signed URL might be expired (rough estimate)
 * Call refreshPhotos() if this returns true
 */
export function isUrlLikelyExpired(_url: string, fetchedAt: Date): boolean {
  const SIGNED_URL_LIFETIME_MS = 15 * 60 * 1000; // 15 minutes
  const elapsed = Date.now() - fetchedAt.getTime();
  return elapsed > SIGNED_URL_LIFETIME_MS * 0.8; // Refresh at 80% of lifetime
}

// User settings types
export interface UserSettings {
  childName?: string;
  childBirthday?: string; // ISO date string YYYY-MM-DD
}

/**
 * Get user settings
 */
export async function getSettings(): Promise<UserSettings> {
  const response = await authFetch(`${API_BASE_URL}/settings`);

  if (!response.ok) {
    throw new Error('Failed to fetch settings');
  }

  return response.json();
}

/**
 * Update user settings
 */
export async function updateSettings(settings: UserSettings): Promise<UserSettings> {
  const response = await authFetch(`${API_BASE_URL}/settings`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(settings),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update settings');
  }

  return response.json();
}
