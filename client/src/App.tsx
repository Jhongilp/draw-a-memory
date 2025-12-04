import { useState, useEffect } from 'react';
import { Baby, Upload, BookOpen, Sparkles, Loader2 } from 'lucide-react';
import { PhotoUpload } from './components/PhotoUpload';
import { PageDraftEditor } from './components/PageDraftEditor';
import { BookView } from './components/BookView';
import type { Photo, PhotoCluster, PageDraft, Theme } from './types/photo';
import { getPhotos, analyzePhotos, getPages, savePageDraft } from './api/photoApi';

type View = 'upload' | 'drafts' | 'book';

function App() {
  const [currentView, setCurrentView] = useState<View>('upload');
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [clusters, setClusters] = useState<PhotoCluster[]>([]);
  const [pages, setPages] = useState<PageDraft[]>([]);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [fetchedPhotos, fetchedPages] = await Promise.all([
        getPhotos(),
        getPages().catch(() => []),
      ]);
      setPhotos(fetchedPhotos || []);
      setPages(fetchedPages || []);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleUploadComplete = async (newPhotos: Photo[]) => {
    const allPhotos = [...newPhotos, ...photos];
    setPhotos(allPhotos);
    
    // Automatically analyze the uploaded photos
    setIsAnalyzing(true);
    try {
      const photoIds = newPhotos.map((p) => p.id);
      const response = await analyzePhotos(photoIds);
      if (response.clusters && response.clusters.length > 0) {
        // Populate photos on each cluster from photo IDs
        // Match clusters with their corresponding drafts from server
        const clustersWithPhotos = response.clusters.map(cluster => {
          // Find the draft that corresponds to this cluster
          const serverDraft = response.drafts?.find(d => d.clusterId === cluster.id);
          return {
            ...cluster,
            draftId: serverDraft?.id, // Store the server's draft ID
            photos: cluster.photoIds.map(id => 
              allPhotos.find(p => p.id === id) || newPhotos.find(p => p.id === id)
            ).filter(Boolean) as Photo[],
            suggestedTitle: cluster.title,
            suggestedDescription: cluster.description,
            suggestedTheme: cluster.theme,
            dateRange: cluster.date,
            ageString: '',
            status: 'draft' as const,
          };
        });
        setClusters(clustersWithPhotos);
        setCurrentView('drafts');
      }
    } catch (error) {
      console.error('Failed to analyze photos:', error);
      // Create a mock cluster for demo purposes
      const mockCluster: PhotoCluster = {
        id: crypto.randomUUID(),
        photoIds: newPhotos.map(p => p.id),
        photos: newPhotos,
        title: 'A Special Day',
        description: 'A wonderful day filled with precious moments and happy memories.',
        theme: 'family' as Theme,
        date: new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' }),
        suggestedTitle: 'A Special Day',
        suggestedDescription: 'A wonderful day filled with precious moments and happy memories.',
        suggestedTheme: 'family' as Theme,
        dateRange: new Date().toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' }),
        ageString: '',
        status: 'draft',
      };
      setClusters([mockCluster]);
      setCurrentView('drafts');
    } finally {
      setIsAnalyzing(false);
    }
  };

  const handleApproveDraft = async (draft: PageDraft) => {
    try {
      await savePageDraft(draft);
    } catch (error) {
      console.error('Failed to save page:', error);
    }
    setPages((prev) => [...prev, draft]);
    setClusters((prev) => prev.filter((c) => c.id !== draft.clusterId));
    
    if (clusters.length <= 1) {
      setCurrentView('book');
    }
  };

  const handleDiscardDraft = (clusterId: string) => {
    setClusters((prev) => prev.filter((c) => c.id !== clusterId));
    if (clusters.length <= 1) {
      setCurrentView('upload');
    }
  };

  return (
    <div className="min-h-screen bg-linear-to-br from-pink-50 via-purple-50 to-blue-50">
      {/* Header */}
      <header className="bg-white/80 backdrop-blur-sm border-b border-pink-100 sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-linear-to-br from-pink-400 to-purple-500 flex items-center justify-center">
                <Baby className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold bg-linear-to-r from-pink-600 to-purple-600 bg-clip-text text-transparent">
                  BabySteps AI Journal
                </h1>
                <p className="text-xs text-gray-500">Turn photos into magical memories</p>
              </div>
            </div>

            {/* Navigation */}
            <nav className="flex gap-2">
              <button
                onClick={() => setCurrentView('upload')}
                className={`px-4 py-2 rounded-xl font-medium text-sm flex items-center gap-2 transition-all
                  ${currentView === 'upload' 
                    ? 'bg-linear-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
                    : 'text-gray-600 hover:bg-pink-50'
                  }`}
              >
                <Upload className="w-4 h-4" />
                Upload
              </button>
              {clusters.length > 0 && (
                <button
                  onClick={() => setCurrentView('drafts')}
                  className={`px-4 py-2 rounded-xl font-medium text-sm flex items-center gap-2 transition-all
                    ${currentView === 'drafts' 
                      ? 'bg-linear-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
                      : 'text-gray-600 hover:bg-pink-50'
                    }`}
                >
                  <Sparkles className="w-4 h-4" />
                  Drafts
                  <span className="bg-white/20 px-1.5 py-0.5 rounded-full text-xs">{clusters.length}</span>
                </button>
              )}
              <button
                onClick={() => setCurrentView('book')}
                className={`px-4 py-2 rounded-xl font-medium text-sm flex items-center gap-2 transition-all
                  ${currentView === 'book' 
                    ? 'bg-linear-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
                    : 'text-gray-600 hover:bg-pink-50'
                  }`}
              >
                <BookOpen className="w-4 h-4" />
                My Book
                {pages.length > 0 && (
                  <span className={`px-1.5 py-0.5 rounded-full text-xs ${currentView === 'book' ? 'bg-white/20' : 'bg-pink-100 text-pink-600'}`}>
                    {pages.length}
                  </span>
                )}
              </button>
            </nav>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-6xl mx-auto px-6 py-10">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-8 h-8 text-pink-500 animate-spin" />
          </div>
        ) : (
          <>
            {/* Upload View */}
            {currentView === 'upload' && (
              <div className="max-w-2xl mx-auto">
                <div className="text-center mb-10">
                  <h2 className="text-3xl font-bold text-gray-800 mb-3">
                    Add New Memories
                  </h2>
                  <p className="text-gray-500">
                    Upload photos and let AI create beautiful memory book pages
                  </p>
                </div>
                
                {isAnalyzing ? (
                  <div className="text-center py-16 bg-white rounded-3xl shadow-xl">
                    <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-linear-to-br from-pink-100 to-purple-100 flex items-center justify-center">
                      <Sparkles className="w-10 h-10 text-pink-500 animate-pulse" />
                    </div>
                    <h3 className="text-xl font-semibold text-gray-700 mb-2">
                      Creating Magic...
                    </h3>
                    <p className="text-gray-500">
                      AI is analyzing your photos and crafting the perfect memory
                    </p>
                  </div>
                ) : (
                  <PhotoUpload onUploadComplete={handleUploadComplete} />
                )}
              </div>
            )}

            {/* Drafts View */}
            {currentView === 'drafts' && (
              <div className="max-w-xl mx-auto space-y-8">
                <div className="text-center mb-10">
                  <h2 className="text-3xl font-bold text-gray-800 mb-3">
                    Review Your Drafts
                  </h2>
                  <p className="text-gray-500">
                    Edit the AI suggestions and add them to your memory book
                  </p>
                </div>
                
                {clusters.map((cluster) => (
                  <PageDraftEditor
                    key={cluster.id}
                    cluster={cluster}
                    onApprove={handleApproveDraft}
                    onDiscard={() => handleDiscardDraft(cluster.id)}
                  />
                ))}
              </div>
            )}

            {/* Book View */}
            {currentView === 'book' && (
              <BookView pages={pages} />
            )}
          </>
        )}
      </main>

      {/* Footer */}
      <footer className="text-center py-8 text-gray-400 text-sm">
        Made with 💕 for your little ones
      </footer>
    </div>
  );
}

export default App;
