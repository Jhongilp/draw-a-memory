import { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, useLocation, Navigate } from 'react-router-dom';
import { SignedIn, SignedOut, useAuth } from '@clerk/clerk-react';
import { Loader2 } from 'lucide-react';
import { Header } from './components/Header';
import { ProtectedRoute } from './components/ProtectedRoute';
import { UploadView } from './components/UploadView';
import { DraftsView } from './components/DraftsView';
import { BookLayout } from './components/BookLayout';
import { BookOverview } from './components/BookOverview';
import { SinglePageView } from './components/SinglePageView';
import { SignInPage } from './components/SignInPage';
import { SignUpPage } from './components/SignUpPage';
import { LandingPage } from './components/LandingPage';
import type { Photo, PhotoCluster, PageDraft, Theme } from './types/photo';
import { getPhotos, analyzePhotos, getPages, savePageDraft, initializeApi } from './api/photoApi';

function AppContent() {
  const location = useLocation();
  const { getToken, isSignedIn } = useAuth();
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [clusters, setClusters] = useState<PhotoCluster[]>([]);
  const [pages, setPages] = useState<PageDraft[]>([]);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isApiInitialized, setIsApiInitialized] = useState(false);

  const isBookView = location.pathname.startsWith('/book');

  // Initialize API with Clerk token getter
  useEffect(() => {
    if (isSignedIn && getToken) {
      initializeApi(getToken);
      setIsApiInitialized(true);
    }
  }, [isSignedIn, getToken]);

  // Load data only after API is initialized and user is signed in
  useEffect(() => {
    if (isApiInitialized && isSignedIn) {
      loadData();
    } else if (!isSignedIn) {
      setIsLoading(false);
    }
  }, [isApiInitialized, isSignedIn]);

  const loadData = async () => {
    try {
      const [fetchedPhotos, fetchedPages] = await Promise.all([
        getPhotos(), 
        getPages().catch(() => []),
      ]);
      const allPhotos = fetchedPhotos || [];
      setPhotos(allPhotos);
      console.log('Fetched photos:', allPhotos);
      
      // Populate photos in each page based on photoIds
      const pagesWithPhotos = (fetchedPages || []).map(page => ({
        ...page,
        photos: page.photoIds
          ?.map(id => allPhotos.find(p => p.id === id))
          .filter((p): p is Photo => p !== undefined) || [],
      }));
      console.log('Fetched pages:', pagesWithPhotos);
      setPages(pagesWithPhotos);
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

  return (
    <div className="min-h-screen flex flex-col bg-linear-to-br from-pink-50 via-purple-50 to-blue-50">
      <Header 
        clusters={clusters} 
        pages={pages} 
        currentPath={location.pathname} 
      />

      {/* Main Content */}
      {isLoading ? (
        <div className="flex-1 flex items-center justify-center">
          <Loader2 className="w-8 h-8 text-pink-500 animate-spin" />
        </div>
      ) : (
        <>
          {location.pathname.startsWith('/upload') && (
            <UploadView 
              isAnalyzing={isAnalyzing} 
              onUploadComplete={handleUploadComplete} 
            />
          )}
          {location.pathname.startsWith('/drafts') && (
            <DraftsView 
              clusters={clusters} 
              onApprove={handleApproveDraft} 
              onDiscard={handleDiscardDraft} 
            />
          )}
          {location.pathname.startsWith('/book') && (
            <Routes>
              <Route element={<BookLayout pages={pages} onReorderPages={handleReorderPages} />}>
                <Route index element={<BookOverview pages={pages} />} />
                <Route path="page/:pageId" element={<SinglePageView pages={pages} />} />
              </Route>
            </Routes>
          )}
        </>
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
      <Routes>
        {/* Public routes */}
        <Route
          path="/"
          element={
            <>
              <SignedIn>
                <Navigate to="/upload" replace />
              </SignedIn>
              <SignedOut>
                <LandingPage />
              </SignedOut>
            </>
          }
        />
        <Route path="/sign-in/*" element={<SignInPage />} />
        <Route path="/sign-up/*" element={<SignUpPage />} />

        {/* Protected routes - using ProtectedRoute wrapper */}
        <Route
          path="/upload/*"
          element={
            <ProtectedRoute>
              <AppContent />
            </ProtectedRoute>
          }
        />
        <Route
          path="/drafts/*"
          element={
            <ProtectedRoute>
              <AppContent />
            </ProtectedRoute>
          }
        />
        <Route
          path="/book/*"
          element={
            <ProtectedRoute>
              <AppContent />
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
