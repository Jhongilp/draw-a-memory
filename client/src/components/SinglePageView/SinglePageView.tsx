import { useParams, Link } from 'react-router-dom';
import { ChevronLeft, ChevronRight, Calendar, Book } from 'lucide-react';
import type { PageDraft, Theme } from '../../types/photo';
import { getPhotoUrl } from '../../api/photoApi';

interface SinglePageViewProps {
  pages: PageDraft[];
}

const themeColors: Record<Theme, { bg: string; accent: string; border: string }> = {
  adventure: { bg: 'bg-amber-50', accent: 'text-amber-600', border: 'border-amber-200' },
  cozy: { bg: 'bg-orange-50', accent: 'text-orange-600', border: 'border-orange-200' },
  celebration: { bg: 'bg-pink-50', accent: 'text-pink-600', border: 'border-pink-200' },
  nature: { bg: 'bg-green-50', accent: 'text-green-600', border: 'border-green-200' },
  family: { bg: 'bg-purple-50', accent: 'text-purple-600', border: 'border-purple-200' },
  milestone: { bg: 'bg-blue-50', accent: 'text-blue-600', border: 'border-blue-200' },
  playful: { bg: 'bg-yellow-50', accent: 'text-yellow-600', border: 'border-yellow-200' },
  love: { bg: 'bg-rose-50', accent: 'text-rose-600', border: 'border-rose-200' },
  growth: { bg: 'bg-emerald-50', accent: 'text-emerald-600', border: 'border-emerald-200' },
  serene: { bg: 'bg-sky-50', accent: 'text-sky-600', border: 'border-sky-200' },
};

export function SinglePageView({ pages }: SinglePageViewProps) {
  const { pageId } = useParams<{ pageId: string }>();

  const currentIndex = pages.findIndex((p) => p.id === pageId);
  const page = pages[currentIndex];

  if (!page) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <Book className="w-16 h-16 mx-auto mb-4 text-gray-300" />
          <h2 className="text-xl font-semibold text-gray-600 mb-2">Page not found</h2>
          <Link to="/book" className="text-pink-500 hover:text-pink-600 underline">
            Go to book overview
          </Link>
        </div>
      </div>
    );
  }

  const themeStyle = themeColors[page.theme] || themeColors.family;
  const photos = page.photos || [];
  const dateDisplay = page.dateRange || '';
  const ageDisplay = page.ageString || '';

  const prevPage = currentIndex > 0 ? pages[currentIndex - 1] : null;
  const nextPage = currentIndex < pages.length - 1 ? pages[currentIndex + 1] : null;

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

        <div className="text-center">
          <span className="text-sm font-medium text-gray-600">
            Page {currentIndex + 1} of {pages.length}
          </span>
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
      <div className="flex-1 overflow-auto p-8">
        <div className="max-w-4xl mx-auto">
          {/* Page card */}
          <div 
            className={`${themeStyle.border} border-2 rounded-3xl overflow-hidden shadow-xl relative`}
          >
            {/* Background image with overlay */}
            {page.backgroundPath ? (
              <>
                <div 
                  className="absolute inset-0 bg-cover bg-center"
                  style={{ backgroundImage: `url(${getPhotoUrl(page.backgroundPath)})` }}
                />
                <div className="absolute inset-0 bg-white/70" />
              </>
            ) : (
              <div className={`absolute inset-0 ${themeStyle.bg}`} />
            )}

            {/* Content wrapper */}
            <div className="relative z-10">
              {/* Photo grid */}
              {photos.length > 0 && (
                <div className="p-4">
                  <div className={`grid gap-3 ${
                    photos.length === 1 ? 'grid-cols-1' :
                    photos.length === 2 ? 'grid-cols-2' :
                    photos.length <= 4 ? 'grid-cols-2' :
                    'grid-cols-3'
                  }`}>
                    {photos.map((photo, idx) => (
                      <div 
                        key={photo.id} 
                        className={`overflow-hidden rounded-2xl shadow-lg ${
                          photos.length === 3 && idx === 0 ? 'col-span-2 row-span-2' : ''
                        } ${
                          photos.length >= 5 && idx === 0 ? 'col-span-2 row-span-2' : ''
                        }`}
                      >
                        <img
                          src={getPhotoUrl(photo.path)}
                          alt=""
                          className="w-full h-full object-cover aspect-square hover:scale-105 transition-transform duration-300"
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Text content */}
              <div className="p-8">
                {/* Age badge */}
                {ageDisplay && (
                  <div className="inline-block mb-4">
                    <span className="px-4 py-1.5 bg-white/80 backdrop-blur-sm rounded-full text-sm font-semibold text-pink-600 shadow-sm">
                      {ageDisplay}
                    </span>
                  </div>
                )}

                {dateDisplay && (
                  <div className="flex items-center gap-2 mb-3">
                    <Calendar className={`w-4 h-4 ${themeStyle.accent}`} />
                    <span className="text-sm text-gray-500">{dateDisplay}</span>
                  </div>
                )}

                <h2 className="text-3xl font-bold text-gray-800 mb-4">{page.title}</h2>
                <p className="text-lg text-gray-600 leading-relaxed">{page.description}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
