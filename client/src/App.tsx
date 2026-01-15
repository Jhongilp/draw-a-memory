import { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import { Baby, Upload, BookOpen, Sparkles, Loader2 } from 'lucide-react';
import { PhotoUpload } from './components/PhotoUpload';
import { PageDraftEditor } from './components/PageDraftEditor';
import { BookLayout } from './components/BookLayout';
import { BookOverview } from './components/BookOverview';
import { SinglePageView } from './components/SinglePageView';
import type { Photo, PhotoCluster, PageDraft, Theme } from './types/photo';
import { getPhotos, analyzePhotos, getPages, savePageDraft } from './api/photoApi';

function AppContent() {
  const location = useLocation();
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
    
    setIsAnalyzing(true);
    try {
      const photoIds = newPhotos.map((p) => p.id);
      const response = await analyzePhotos(photoIds);
      if (response.clusters && response.clusters.length > 0) {
        const clustersWithPhotos = response.clusters.map(cluster => {
          const serverDraft = response.drafts?.find(d => d.clusterId === cluster.id);
          return {
            ...cluster,
            draftId: serverDraft?.id,
            backgroundPath: serverDraft?.backgroundPath || cluster.backgroundPath,
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
      }
    } catch (error) {
      console.error('Failed to analyze photos:', error);
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
  };

  const handleDiscardDraft = (clusterId: string) => {
    setClusters((prev) => prev.filter((c) => c.id !== clusterId));
  };

  const handleReorderPages = (reorderedPages: PageDraft[]) => {
    setPages(reorderedPages);
    // TODO: Persist order to server
  };

  const isUploadView = location.pathname === '/' || location.pathname === '/upload';
  const isDraftsView = location.pathname === '/drafts';
  const isBookView = location.pathname.startsWith('/book');

  return (
    <div className="min-h-screen flex flex-col bg-gradient-to-br from-pink-50 via-purple-50 to-blue-50">
      {/* Header */}
      <header className="bg-white/80 backdrop-blur-sm border-b border-pink-100 sticky top-0 z-50">
        <div className="max-w-full px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-pink-400 to-purple-500 flex items-center justify-center">
                <Baby className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold bg-gradient-to-r from-pink-600 to-purple-600 bg-clip-text text-transparent">
                  BabySteps AI Journal
                </h1>
                <p className="text-xs text-gray-500">Turn photos into magical memories</p>
              </div>
            </div>

            {/* Navigation */}
            <nav className="flex gap-2">
              <Link
                to="/"
                className={`px-4 py-2 rounded-xl font-medium text-sm flex items-center gap-2 transition-all
                  ${isUploadView 
                    ? 'bg-gradient-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
                    : 'text-gray-600 hover:bg-pink-50'
                  }`}
              >
                <Upload className="w-4 h-4" />
                Upload
              </Link>
              {clusters.length > 0 && (
                <Link
                  to="/drafts"
                  className={`px-4 py-2 rounded-xl font-medium text-sm flex items-center gap-2 transition-all
                    ${isDraftsView 
                      ? 'bg-gradient-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
                      : 'text-gray-600 hover:bg-pink-50'
                    }`}
                >
                  <Sparkles className="w-4 h-4" />
                  Drafts
                  <span className="bg-white/20 px-1.5 py-0.5 rounded-full text-xs">{clusters.length}</span>
                </Link>
              )}
              <Link
                to="/book"
                className={`px-4 py-2 rounded-xl font-medium text-sm flex items-center gap-2 transition-all
                  ${isBookView 
                    ? 'bg-gradient-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
                    : 'text-gray-600 hover:bg-pink-50'
                  }`}
              >
                <BookOpen className="w-4 h-4" />
                My Book
                {pages.length > 0 && (
                  <span className={`px-1.5 py-0.5 rounded-full text-xs ${isBookView ? 'bg-white/20' : 'bg-pink-100 text-pink-600'}`}>
                    {pages.length}
                  </span>
                )}
              </Link>
            </nav>
          </div>
        </div>
      </header>

      {/* Main Content */}
      {isLoading ? (
        <div className="flex-1 flex items-center justify-center">
          <Loader2 className="w-8 h-8 text-pink-500 animate-spin" />
        </div>
      ) : (
        <Routes>
          {/* Upload View */}
          <Route
            path="/"
            element={
              <main className="flex-1 flex items-center justify-center p-10">
                <div className="max-w-2xl w-full">
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
                      <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-gradient-to-br from-pink-100 to-purple-100 flex items-center justify-center">
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
              </main>
            }
          />

          {/* Drafts View */}
          <Route
            path="/drafts"
            element={
              <main className="flex-1 overflow-auto p-10">
                <div className="max-w-xl mx-auto space-y-8">
                  <div className="text-center mb-10">
                    <h2 className="text-3xl font-bold text-gray-800 mb-3">
                      Review Your Drafts
                    </h2>
                    <p className="text-gray-500">
                      Edit the AI suggestions and add them to your memory book
                    </p>
                  </div>
                  
                  {clusters.length === 0 ? (
                    <div className="text-center py-16">
                      <Sparkles className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                      <p className="text-gray-500">No drafts to review</p>
                      <Link to="/" className="text-pink-500 hover:text-pink-600 underline text-sm mt-2 inline-block">
                        Upload some photos
                      </Link>
                    </div>
                  ) : (
                    clusters.map((cluster) => (
                      <PageDraftEditor
                        key={cluster.id}
                        cluster={cluster}
                        onApprove={handleApproveDraft}
                        onDiscard={() => handleDiscardDraft(cluster.id)}
                      />
                    ))
                  )}
                </div>
              </main>
            }
          />

          {/* Book View with Sidebar */}
          <Route
            path="/book"
            element={<BookLayout pages={pages} onReorderPages={handleReorderPages} />}
          >
            <Route index element={<BookOverview pages={pages} />} />
            <Route path="page/:pageId" element={<SinglePageView pages={pages} />} />
          </Route>
        </Routes>
      )}

      {/* Footer - only show on non-book views */}
      {!isBookView && (
        <footer className="text-center py-8 text-gray-400 text-sm">
          Made with 💕 for your little ones
        </footer>
      )}
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  );
}

export default App;
