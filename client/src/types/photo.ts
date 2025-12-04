export interface Photo {
  id: string;
  filename: string;
  path: string;
  size: number;
  uploadedAt: string;
  takenAt?: string;
  metadata?: PhotoMetadata;
}

export interface PhotoMetadata {
  dateTaken?: string;
  location?: string;
  camera?: string;
}

export interface PhotoCluster {
  id: string;
  photoIds: string[];
  photos?: Photo[]; // Populated on client side
  title: string;
  description: string;
  theme: Theme;
  date: string;
  suggestedTitle?: string;
  suggestedDescription?: string;
  suggestedTheme?: Theme;
  dateRange?: string;
  ageString?: string;
  status?: 'draft' | 'approved';
}

export interface PageDraft {
  id: string;
  clusterId: string;
  photoIds: string[];
  title: string;
  description: string;
  theme: Theme;
  photos?: Photo[]; // Populated on client side
  dateRange?: string;
  ageString?: string;
  status: 'draft' | 'approved' | 'rejected';
  createdAt: string;
  approvedAt?: string;
}

export type Theme = 
  | 'adventure'
  | 'cozy'
  | 'celebration'
  | 'nature'
  | 'family'
  | 'milestone'
  | 'playful'
  | 'love'
  | 'growth'
  | 'serene';

export interface UploadResponse {
  success: boolean;
  message: string;
  photos?: Photo[];
}

export interface ClusterResponse {
  success?: boolean;
  clusters: PhotoCluster[];
  drafts: PageDraft[];
}

export interface ErrorResponse {
  success: boolean;
  error: string;
}
