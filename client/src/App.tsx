import { useEffect } from "react";
import {
  BrowserRouter,
  Routes,
  Route,
  useLocation,
  Navigate,
} from "react-router-dom";
import { SignedIn, SignedOut, useAuth } from "@clerk/clerk-react";
import { Loader2 } from "lucide-react";
import { Header } from "./components/Header";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { UploadView } from "./components/UploadView";
import { DraftsView } from "./components/DraftsView";
import { BookLayout } from "./components/BookLayout";
import { BookOverview } from "./components/BookOverview";
import { SinglePageView } from "./components/SinglePageView";
import { SignInPage } from "./components/SignInPage";
import { SignUpPage } from "./components/SignUpPage";
import { LandingPage } from "./components/LandingPage";
import { initializeApi } from "./api/photoApi";
import { useAppDispatch, useAppSelector } from "./store/hooks";
import { fetchPagesData } from "./store/slices";

function AppContent() {
  const location = useLocation();
  const { getToken, isSignedIn } = useAuth();
  const dispatch = useAppDispatch();

  // Get loading state from Redux store
  const isLoading = useAppSelector((state) => state.pages.isLoading);

  const isBookView = location.pathname.startsWith("/book");

  // Initialize API with Clerk token and load data when signed in
  useEffect(() => {
    if (isSignedIn && getToken) {
      initializeApi(getToken);
      dispatch(fetchPagesData());
    }
  }, [isSignedIn, getToken, dispatch]);

  return (
    <div className="min-h-screen flex flex-col bg-linear-to-br from-pink-50 via-purple-50 to-blue-50">
      <Header currentPath={location.pathname} />

      {/* Main Content */}
      {isLoading ? (
        <div className="flex-1 flex items-center justify-center">
          <Loader2 className="w-8 h-8 text-pink-500 animate-spin" />
        </div>
      ) : (
        <>
          {location.pathname.startsWith("/upload") && <UploadView />}
          {location.pathname.startsWith("/drafts") && <DraftsView />}
          {location.pathname.startsWith("/book") && (
            <Routes>
              <Route element={<BookLayout />}>
                <Route index element={<BookOverview />} />
                <Route path="page/:pageId" element={<SinglePageView />} />
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
