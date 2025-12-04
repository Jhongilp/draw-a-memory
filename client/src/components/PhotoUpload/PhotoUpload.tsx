import { useState, useCallback, useRef } from 'react';
import { Upload, ImagePlus, X, Loader2 } from 'lucide-react';
import { uploadPhotos } from '../../api/photoApi';
import type { Photo } from '../../types/photo';

const MAX_FILES = 10;
const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5 MB

interface PhotoUploadProps {
  onUploadComplete: (photos: Photo[]) => void;
}

export function PhotoUpload({ onUploadComplete }: PhotoUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [previewFiles, setPreviewFiles] = useState<{ file: File; preview: string }[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFiles = useCallback((files: FileList | null) => {
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
      setError(`Some files exceed the 5MB limit`);
      return;
    }

    setError(null);
    
    // Create previews
    const previews = imageFiles.map((file) => ({
      file,
      preview: URL.createObjectURL(file),
    }));
    setPreviewFiles(previews);
  }, []);

  const handleUpload = async () => {
    if (previewFiles.length === 0) return;

    setIsUploading(true);
    setError(null);

    try {
      const files = previewFiles.map((p) => p.file);
      const response = await uploadPhotos(files);
      
      if (response.success && response.photos) {
        onUploadComplete(response.photos);
        // Cleanup previews
        previewFiles.forEach((p) => URL.revokeObjectURL(p.preview));
        setPreviewFiles([]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setIsUploading(false);
    }
  };

  const removePreview = (index: number) => {
    URL.revokeObjectURL(previewFiles[index].preview);
    setPreviewFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const clearAll = () => {
    previewFiles.forEach((p) => URL.revokeObjectURL(p.preview));
    setPreviewFiles([]);
    setError(null);
  };

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
    <div className="w-full max-w-2xl mx-auto">
      {/* Drop Zone */}
      <div
        className={`
          relative border-3 border-dashed rounded-3xl p-8 text-center cursor-pointer
          transition-all duration-300 ease-out
          ${isDragging 
            ? 'border-pink-400 bg-pink-50 scale-[1.02]' 
            : 'border-pink-200 bg-white hover:border-pink-300 hover:bg-pink-50/50'
          }
          ${isUploading ? 'pointer-events-none opacity-60' : ''}
        `}
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
          className="hidden"
        />
        
        <div className="flex flex-col items-center gap-4">
          <div className="w-16 h-16 rounded-full bg-linear-to-br from-pink-100 to-purple-100 flex items-center justify-center">
            <ImagePlus className="w-8 h-8 text-pink-500" />
          </div>
          <div>
            <h3 className="text-xl font-semibold text-gray-700 mb-1">
              Drop your memories here
            </h3>
            <p className="text-gray-500">
              or click to browse your photos
            </p>
          </div>
          <p className="text-sm text-gray-400">
            Up to {MAX_FILES} photos, max 5MB each • JPG, PNG, HEIC
          </p>
        </div>
      </div>

      {/* Preview Grid */}
      {previewFiles.length > 0 && (
        <div className="mt-6">
          <div className="flex items-center justify-between mb-4">
            <span className="text-sm font-medium text-gray-600">
              {previewFiles.length} photo{previewFiles.length !== 1 ? 's' : ''} selected
            </span>
            <button
              onClick={(e) => { e.stopPropagation(); clearAll(); }}
              className="text-sm text-gray-500 hover:text-red-500 transition-colors"
            >
              Clear all
            </button>
          </div>
          
          <div className="grid grid-cols-4 gap-3">
            {previewFiles.map((item, index) => (
              <div key={index} className="relative group aspect-square">
                <img
                  src={item.preview}
                  alt={`Preview ${index + 1}`}
                  className="w-full h-full object-cover rounded-xl"
                />
                <button
                  onClick={(e) => { e.stopPropagation(); removePreview(index); }}
                  className="absolute -top-2 -right-2 w-6 h-6 bg-red-500 text-white rounded-full 
                           flex items-center justify-center opacity-0 group-hover:opacity-100 
                           transition-opacity shadow-lg"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>

          <button
            onClick={(e) => { e.stopPropagation(); handleUpload(); }}
            disabled={isUploading}
            className="mt-6 w-full py-3 px-6 bg-linear-to-r from-pink-500 to-purple-500 
                     text-white font-semibold rounded-xl hover:from-pink-600 hover:to-purple-600 
                     transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed
                     flex items-center justify-center gap-2 shadow-lg shadow-pink-500/25"
          >
            {isUploading ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                Uploading...
              </>
            ) : (
              <>
                <Upload className="w-5 h-5" />
                Upload & Analyze
              </>
            )}
          </button>
        </div>
      )}

      {/* Error Message */}
      {error && (
        <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-xl text-red-600 text-sm">
          {error}
        </div>
      )}
    </div>
  );
}
