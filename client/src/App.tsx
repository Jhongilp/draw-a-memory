import { useState, useEffect } from 'react';
import { PhotoUpload } from './components/PhotoUpload';
import { PhotoGallery } from './components/PhotoGallery';
import type { Photo } from './types/photo';
import { getPhotos } from './api/photoApi';
import './App.css';

function App() {
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadPhotos();
  }, []);

  const loadPhotos = async () => {
    try {
      const fetchedPhotos = await getPhotos();
      setPhotos(fetchedPhotos || []);
    } catch (error) {
      console.error('Failed to load photos:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleUploadComplete = (newPhotos: Photo[]) => {
    setPhotos((prev) => [...newPhotos, ...prev]);
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>🍼 Draw a Memory</h1>
        <p>Capture and preserve your precious family moments</p>
      </header>

      <main className="app-main">
        <section className="upload-section">
          <PhotoUpload onUploadComplete={handleUploadComplete} />
        </section>

        <section className="gallery-section">
          {isLoading ? (
            <div className="loading">Loading your memories...</div>
          ) : (
            <PhotoGallery photos={photos} />
          )}
        </section>
      </main>

      <footer className="app-footer">
        <p>Made with ❤️ for your little ones</p>
      </footer>
    </div>
  );
}

export default App;
