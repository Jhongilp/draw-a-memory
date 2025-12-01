import { useState, useCallback, useRef } from 'react';
import { uploadPhotos } from '../../api/photoApi';
import type { Photo } from '../../types/photo';
import './PhotoUpload.css';

const MAX_FILES = 10;
const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5 MB

interface PhotoUploadProps {
  onUploadComplete: (photos: Photo[]) => void;
}

export function PhotoUpload({ onUploadComplete }: PhotoUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState<string>('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFiles = useCallback(async (files: FileList | null) => {
    if (!files || files.length === 0) return;

    const imageFiles = Array.from(files).filter((file) =>
      file.type.startsWith('image/')
    );

    if (imageFiles.length === 0) {
      setError('Please select valid image files');
      return;
    }

    if (imageFiles.length > MAX_FILES) {
      setError(`Maximum ${MAX_FILES} photos per upload`);
      return;
    }

    const oversizedFiles = imageFiles.filter((file) => file.size > MAX_FILE_SIZE);
    if (oversizedFiles.length > 0) {
      setError(`Some files exceed the 5MB limit: ${oversizedFiles.map(f => f.name).join(', ')}`);
      return;
    }

    const validFiles = imageFiles;

    setError(null);
    setIsUploading(true);
    setUploadProgress(`Uploading ${validFiles.length} photo(s)...`);

    try {
      const response = await uploadPhotos(validFiles);
      
      if (response.success && response.photos) {
        onUploadComplete(response.photos);
        setUploadProgress(`Successfully uploaded ${response.photos.length} photo(s)!`);
        setTimeout(() => setUploadProgress(''), 3000);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setIsUploading(false);
    }
  }, [onUploadComplete]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    handleFiles(e.dataTransfer.files);
  }, [handleFiles]);

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleFiles(e.target.files);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  return (
    <div className="photo-upload-container">
      <div
        className={`photo-upload-dropzone ${isDragging ? 'dragging' : ''} ${isUploading ? 'uploading' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={handleClick}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          multiple
          onChange={handleInputChange}
          className="photo-upload-input"
        />
        
        <div className="photo-upload-content">
          {isUploading ? (
            <>
              <div className="upload-spinner"></div>
              <p>{uploadProgress}</p>
            </>
          ) : (
            <>
              <div className="upload-icon">📸</div>
              <h3>Upload Your Memories</h3>
              <p>Drag and drop photos here, or click to select</p>
              <p className="upload-hint">Up to 10 photos, max 5MB each</p>
              <p className="upload-hint">Supports JPG, PNG, GIF, WebP, HEIC</p>
            </>
          )}
        </div>
      </div>

      {error && (
        <div className="upload-error">
          <span>⚠️</span> {error}
        </div>
      )}

      {uploadProgress && !isUploading && (
        <div className="upload-success">
          <span>✅</span> {uploadProgress}
        </div>
      )}
    </div>
  );
}
