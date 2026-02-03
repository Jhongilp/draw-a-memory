import { useParams, Link, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, Calendar, Book, Trash2 } from "lucide-react";
import type { PageDraft, Theme } from "../../types/photo";
import { getPhotoUrl, deleteDraft } from "../../api/photoApi";
import { useAppSelector, useAppDispatch } from "../../store/hooks";
import { removePage } from "../../store/slices";

// Cache for preloaded images to keep them in memory
const imageCache = new Map<string, HTMLImageElement>();

// Preload images for a page and store in cache
function preloadPageImages(page: PageDraft | null): Promise<void>[] {
  if (!page) return [];

  const promises: Promise<void>[] = [];

  // Preload background
  if (page.backgroundPath) {
    const url = getPhotoUrl(page.backgroundPath);
    if (!imageCache.has(url)) {
      const bgImg = new Image();
      const promise = new Promise<void>((resolve) => {
        bgImg.onload = () => resolve();
        bgImg.onerror = () => resolve();
      });
      bgImg.src = url;
      imageCache.set(url, bgImg);
      promises.push(promise);
    }
  }

  // Preload photos
  page.photos?.forEach((photo) => {
    const url = getPhotoUrl(photo.path);
    if (!imageCache.has(url)) {
      const img = new Image();
      const promise = new Promise<void>((resolve) => {
        img.onload = () => resolve();
        img.onerror = () => resolve();
      });
      img.src = url;
      imageCache.set(url, img);
      promises.push(promise);
    }
  });

  return promises;
}

const themeColors: Record<
  Theme,
  { bg: string; accent: string; border: string }
> = {
  adventure: {
    bg: "bg-amber-50",
    accent: "text-amber-600",
    border: "border-amber-200",
  },
  cozy: {
    bg: "bg-orange-50",
    accent: "text-orange-600",
    border: "border-orange-200",
  },
  celebration: {
    bg: "bg-pink-50",
    accent: "text-pink-600",
    border: "border-pink-200",
  },
  nature: {
    bg: "bg-green-50",
    accent: "text-green-600",
    border: "border-green-200",
  },
  family: {
    bg: "bg-purple-50",
    accent: "text-purple-600",
    border: "border-purple-200",
  },
  milestone: {
    bg: "bg-blue-50",
    accent: "text-blue-600",
    border: "border-blue-200",
  },
  playful: {
    bg: "bg-yellow-50",
    accent: "text-yellow-600",
    border: "border-yellow-200",
  },
  love: {
    bg: "bg-rose-50",
    accent: "text-rose-600",
    border: "border-rose-200",
  },
  growth: {
    bg: "bg-emerald-50",
    accent: "text-emerald-600",
    border: "border-emerald-200",
  },
  serene: { bg: "bg-sky-50", accent: "text-sky-600", border: "border-sky-200" },
};

export function SinglePageView() {
  const pages = useAppSelector((state) => state.pages.pages);
  const { pageId } = useParams<{ pageId: string }>();
  const [imagesReady, setImagesReady] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const currentIndex = pages.findIndex((p) => p.id === pageId);
  const page = pages[currentIndex];
  const prevPage = currentIndex > 0 ? pages[currentIndex - 1] : null;
  const nextPage =
    currentIndex < pages.length - 1 ? pages[currentIndex + 1] : null;

  // Preload current page images and wait for them
  useEffect(() => {
    setImagesReady(false);
    if (page) {
      const promises = preloadPageImages(page);
      if (promises.length === 0) {
        setImagesReady(true);
      } else {
        Promise.all(promises).then(() => setImagesReady(true));
      }
    }
  }, [page]);

  // Preload adjacent pages' images for smoother navigation (in background)
  useEffect(() => {
    preloadPageImages(prevPage);
    preloadPageImages(nextPage);
  }, [prevPage, nextPage]);

  const handleDeletePage = async () => {
    if (!page || isDeleting) return;
    
    setIsDeleting(true);
    try {
      await deleteDraft(page.id);
      dispatch(removePage(page.id));
      setShowDeleteConfirm(false);
      // Navigate to next page, previous page, or book overview
      if (nextPage) {
        navigate(`/book/page/${nextPage.id}`);
      } else if (prevPage) {
        navigate(`/book/page/${prevPage.id}`);
      } else {
        navigate('/book');
      }
    } catch (error) {
      console.error('Failed to delete page:', error);
      setIsDeleting(false);
      setShowDeleteConfirm(false);
    }
  };

  if (!page) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <Book className="w-16 h-16 mx-auto mb-4 text-gray-300" />
          <h2 className="text-xl font-semibold text-gray-600 mb-2">
            Page not found
          </h2>
          <Link
            to="/book"
            className="text-pink-500 hover:text-pink-600 underline"
          >
            Go to book overview
          </Link>
        </div>
      </div>
    );
  }

  const themeStyle = themeColors[page.theme] || themeColors.family;
  const photos = page.photos || [];
  const dateDisplay = page.dateRange || "";
  const ageDisplay = page.ageString || "";

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Page navigation header */}
      <div className="bg-white/60 backdrop-blur-sm border-b border-pink-100 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          {prevPage ? (
            <Link
              to={`/book/page/${prevPage.id}`}
              className="flex items-center gap-1 text-gray-500 hover:text-pink-600 transition-colors"
            >
              <ChevronLeft className="w-5 h-5" />
              <span className="text-sm">Previous</span>
            </Link>
          ) : (
            <div className="w-20" />
          )}
        </div>

        <div className="text-center flex items-center gap-3">
          <span className="text-sm font-medium text-gray-600">
            Page {currentIndex + 1} of {pages.length}
          </span>
          <button
            onClick={() => setShowDeleteConfirm(true)}
            className="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
            title="Delete page"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>

        <div className="flex items-center gap-4">
          {nextPage ? (
            <Link
              to={`/book/page/${nextPage.id}`}
              className="flex items-center gap-1 text-gray-500 hover:text-pink-600 transition-colors"
            >
              <span className="text-sm">Next</span>
              <ChevronRight className="w-5 h-5" />
            </Link>
          ) : (
            <div className="w-20" />
          )}
        </div>
      </div>

      {/* Page content */}
      <div className="flex-1 overflow-auto p-8 flex items-center justify-center">
        <div className="w-full max-w-5xl">
          {/* Page card with fixed 16:9 aspect ratio */}
          <div
            className={`${themeStyle.border} border-2 rounded-3xl overflow-hidden shadow-xl relative aspect-video transition-opacity duration-150 ${imagesReady ? 'opacity-100' : 'opacity-0'}`}
          >
            {/* Background image with overlay */}
            {page.backgroundPath ? (
              <>
                <div
                  className="absolute inset-0 bg-cover bg-center"
                  style={{
                    backgroundImage: `url("${getPhotoUrl(page.backgroundPath)}")`,
                  }}
                />
                <div className="absolute inset-0 bg-white/40" />
              </>
            ) : (
              <div className={`absolute inset-0 ${themeStyle.bg}`} />
            )}

            {/* Content wrapper - absolute positioning to fill the fixed aspect ratio container */}
            <div className="absolute inset-0 z-10 flex">
              {/* Photo section - takes left portion */}
              <div className="flex-1 p-4 flex items-center justify-center">
                {photos.length > 0 ? (
                  <div
                    className={`w-full h-full grid gap-2 ${
                      photos.length === 1
                        ? "grid-cols-1 grid-rows-1"
                        : photos.length === 2
                        ? "grid-cols-2 grid-rows-1"
                        : photos.length <= 4
                        ? "grid-cols-2 grid-rows-2"
                        : "grid-cols-3 grid-rows-2"
                    }`}
                  >
                    {photos.map((photo, idx) => (
                      <div
                        key={photo.id}
                        className={`overflow-hidden rounded-2xl shadow-lg ${
                          photos.length === 3 && idx === 0
                            ? "col-span-1 row-span-2"
                            : ""
                        } ${
                          photos.length >= 5 && idx === 0
                            ? "col-span-2 row-span-2"
                            : ""
                        }`}
                      >
                        <img
                          src={getPhotoUrl(photo.path)}
                          alt=""
                          loading="eager"
                          decoding="sync"
                          className="w-full h-full object-cover"
                        />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-gray-400">
                    <span className="text-lg">No photos</span>
                  </div>
                )}
              </div>

              {/* Text content - takes right portion */}
              <div className="w-80 p-6 flex flex-col justify-center bg-white/40 backdrop-blur-sm">
                {/* Age badge */}
                {ageDisplay && (
                  <div className="mb-3">
                    <span className="px-3 py-1 bg-white/80 backdrop-blur-sm rounded-full text-sm font-semibold text-pink-600 shadow-sm">
                      {ageDisplay}
                    </span>
                  </div>
                )}

                {dateDisplay && (
                  <div className="flex items-center gap-2 mb-2">
                    <Calendar className={`w-4 h-4 ${themeStyle.accent}`} />
                    <span className="text-sm text-gray-500">{dateDisplay}</span>
                  </div>
                )}

                <h2 className="text-2xl font-bold text-gray-800 mb-3 line-clamp-2">
                  {page.title}
                </h2>
                <p className="text-base text-gray-600 leading-relaxed line-clamp-6">
                  {page.description}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Delete confirmation modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
          <div className="bg-white rounded-2xl p-6 max-w-sm mx-4 shadow-2xl">
            <h3 className="text-lg font-semibold text-gray-800 mb-2">
              Delete this page?
            </h3>
            <p className="text-gray-600 text-sm mb-4">
              This will permanently delete "{page.title}" and all its photos. This action cannot be undone.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                disabled={isDeleting}
                className="flex-1 py-2 px-4 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50 transition-colors font-medium disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDeletePage}
                disabled={isDeleting}
                className="flex-1 py-2 px-4 bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors font-medium disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {isDeleting ? (
                  <>
                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Deleting...
                  </>
                ) : (
                  <>
                    <Trash2 className="w-4 h-4" />
                    Delete
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
