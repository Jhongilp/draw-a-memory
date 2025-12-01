import type { Photo } from '../../types/photo';
import { getPhotoUrl } from '../../api/photoApi';
import './PhotoGallery.css';

interface PhotoGalleryProps {
  photos: Photo[];
}

export function PhotoGallery({ photos }: PhotoGalleryProps) {
  if (photos.length === 0) {
    return (
      <div className="photo-gallery-empty">
        <div className="empty-icon">🖼️</div>
        <h3>No memories yet</h3>
        <p>Upload some photos to start creating your memory album</p>
      </div>
    );
  }

  return (
    <div className="photo-gallery">
      <h2 className="gallery-title">Your Memories ({photos.length})</h2>
      <div className="photo-grid">
        {photos.map((photo) => (
          <div key={photo.id} className="photo-card">
            <div className="photo-wrapper">
              <img
                src={getPhotoUrl(photo.path)}
                alt={photo.filename}
                loading="lazy"
              />
            </div>
            <div className="photo-info">
              <p className="photo-name" title={photo.filename}>
                {photo.filename}
              </p>
              <p className="photo-date">
                {new Date(photo.uploadedAt).toLocaleDateString()}
              </p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
