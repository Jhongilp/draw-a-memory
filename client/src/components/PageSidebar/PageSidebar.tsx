import { useRef, useEffect, useState, useCallback } from 'react';
import { NavLink } from 'react-router-dom';
import { GripVertical, BookOpen } from 'lucide-react';
import type { PageDraft } from '../../types/photo';
import { getPhotoUrl } from '../../api/photoApi';

interface PageSidebarProps {
  pages: PageDraft[];
  onReorder?: (pages: PageDraft[]) => void;
}

const ITEMS_PER_PAGE = 20;

export function PageSidebar({ pages, onReorder: _onReorder }: PageSidebarProps) {
  const [visibleCount, setVisibleCount] = useState(ITEMS_PER_PAGE);
  const loaderRef = useRef<HTMLDivElement>(null);

  // Infinite scroll observer
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && visibleCount < pages.length) {
          setVisibleCount((prev) => Math.min(prev + ITEMS_PER_PAGE, pages.length));
        }
      },
      { threshold: 0.1 }
    );

    if (loaderRef.current) {
      observer.observe(loaderRef.current);
    }

    return () => observer.disconnect();
  }, [visibleCount, pages.length]);

  // Reset visible count when pages change significantly
  useEffect(() => {
    if (pages.length < visibleCount) {
      setVisibleCount(Math.max(ITEMS_PER_PAGE, pages.length));
    }
  }, [pages.length, visibleCount]);

  const visiblePages = pages.slice(0, visibleCount);

  // Get the first photo from a page's photos array for thumbnail
  const getThumbnail = useCallback((page: PageDraft): string | null => {
    if (page.photos && page.photos.length > 0) {
      return getPhotoUrl(page.photos[0].path);
    }
    return null;
  }, []);

  console.log('PageSidebar rendering with pages:', pages);

  if (pages.length === 0) {
    return (
      <aside className="w-20 bg-white/80 backdrop-blur-sm border-r border-pink-100 flex flex-col items-center py-6">
        <div className="text-center text-gray-400">
          <BookOpen className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <span className="text-xs">No pages</span>
        </div>
      </aside>
    );
  }

  return (
    <aside className="w-24 bg-white/80 backdrop-blur-sm border-r border-pink-100 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="p-3 border-b border-pink-100 text-center">
        <span className="text-xs font-medium text-gray-500">{pages.length} Pages</span>
      </div>

      {/* Scrollable page list */}
      <div className="flex-1 overflow-y-auto overflow-x-hidden py-2 px-2 space-y-2">
        {visiblePages.map((page, index) => {
          const thumbnail = getThumbnail(page);
          
          return (
            <NavLink
              key={page.id}
              to={`/book/page/${page.id}`}
              className={({ isActive }) =>
                `group relative block rounded-lg overflow-hidden transition-all duration-200 ${
                  isActive
                    ? 'ring-2 ring-pink-500 ring-offset-2 shadow-lg scale-105'
                    : 'hover:ring-2 hover:ring-pink-300 hover:ring-offset-1 opacity-80 hover:opacity-100'
                }`
              }
            >
              {/* Drag handle placeholder for future sorting */}
              <div className="absolute top-0.5 left-0.5 z-10 opacity-0 group-hover:opacity-60 transition-opacity cursor-grab">
                <GripVertical className="w-3 h-3 text-white drop-shadow-md" />
              </div>

              {/* Page number badge */}
              <div className="absolute top-0.5 right-0.5 z-10 bg-black/50 backdrop-blur-sm text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full">
                {index + 1}
              </div>

              {/* Thumbnail */}
              <div className="aspect-[3/4] bg-gradient-to-br from-pink-100 to-purple-100">
                {thumbnail ? (
                  <img
                    src={thumbnail}
                    alt={page.title}
                    className="w-full h-full object-cover"
                    loading="lazy"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <BookOpen className="w-6 h-6 text-pink-300" />
                  </div>
                )}
              </div>
            </NavLink>
          );
        })}

        {/* Infinite scroll loader */}
        {visibleCount < pages.length && (
          <div ref={loaderRef} className="py-4 text-center">
            <div className="w-6 h-6 mx-auto border-2 border-pink-300 border-t-transparent rounded-full animate-spin" />
          </div>
        )}
      </div>
    </aside>
  );
}
