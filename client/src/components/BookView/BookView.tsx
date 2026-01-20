import { Book, Calendar, Heart } from 'lucide-react';
import type { PageDraft, Theme } from '../../types/photo';
import { getPhotoUrl } from '../../api/photoApi';

interface BookViewProps {
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

export function BookView({ pages }: BookViewProps) {
  if (pages.length === 0) {
    return (
      <div className="text-center py-16">
        <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-linear-to-br from-pink-100 to-purple-100 flex items-center justify-center">
          <Book className="w-10 h-10 text-pink-400" />
        </div>
        <h3 className="text-xl font-semibold text-gray-700 mb-2">
          Your memory book is empty
        </h3>
        <p className="text-gray-500 max-w-md mx-auto">
          Upload some photos and let the AI create beautiful pages for your baby's memory book!
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      {/* Book cover */}
      <div className="text-center mb-12">
        <div className="inline-flex items-center gap-2 px-4 py-2 bg-linear-to-r from-pink-500 to-purple-500 text-white rounded-full text-sm font-medium mb-4">
          <Heart className="w-4 h-4" />
          {pages.length} {pages.length === 1 ? 'Memory' : 'Memories'}
        </div>
        <h2 className="text-3xl font-bold text-gray-800">Baby's First Year</h2>
      </div>

      {/* Timeline */}
      <div className="relative">
        {/* Timeline line */}
        <div className="absolute left-8 top-0 bottom-0 w-0.5 bg-linear-to-b from-pink-300 via-purple-300 to-pink-300" />

        {/* Pages */}
        <div className="space-y-8">
          {pages.map((page) => {
            const themeStyle = themeColors[page.theme] || themeColors.family;
            const photos = page.photos || [];
            const dateDisplay = page.dateRange || '';
            const ageDisplay = page.ageString || '';

            return (
              <div key={page.id} className="relative pl-20">
                {/* Timeline dot */}
                <div className="absolute left-6 top-8 w-5 h-5 rounded-full bg-white border-4 border-pink-400 shadow-lg" />

                {/* Age label */}
                {ageDisplay && (
                  <div className="absolute left-0 top-7 text-xs font-bold text-pink-500 w-12 text-right">
                    {ageDisplay}
                  </div>
                )}

                {/* Page card */}
                {page.backgroundPath && console.log('Background URL:', getPhotoUrl(page.backgroundPath))}
                <div 
                  className={`${themeStyle.border} border-2 rounded-3xl overflow-hidden shadow-lg hover:shadow-xl transition-shadow relative`}
                >
                  {/* Background image with overlay */}
                  {page.backgroundPath ? (
                    <>
                      <img 
                        src={getPhotoUrl(page.backgroundPath)} 
                        alt="" 
                        className="absolute inset-0 w-full h-full object-cover"
                        onError={(e) => console.error('Background image failed to load:', e)}
                        onLoad={() => console.log('Background image loaded successfully')}
                      />
                      {/* Semi-transparent overlay for readability */}
                      <div className="absolute inset-0 bg-white/70" />
                    </>
                  ) : (
                    <div className={`absolute inset-0 ${themeStyle.bg}`} />
                  )}

                  {/* Content wrapper (above background) */}
                  <div className="relative z-10">
                    {/* Photo grid */}
                    {photos.length > 0 && (
                      <div className="grid grid-cols-4 gap-1 p-1">
                        {photos.slice(0, 4).map((photo, idx) => (
                          <div 
                            key={photo.id} 
                            className={`aspect-square overflow-hidden rounded-lg ${idx === 0 ? 'col-span-2 row-span-2' : ''}`}
                          >
                            <img
                              src={getPhotoUrl(photo.path)}
                              alt=""
                              className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
                            />
                          </div>
                        ))}
                      </div>
                    )}

                    {/* Content */}
                    <div className="p-6">
                      {dateDisplay && (
                        <div className="flex items-center gap-2 mb-2">
                          <Calendar className={`w-4 h-4 ${themeStyle.accent}`} />
                          <span className="text-sm text-gray-500">{dateDisplay}</span>
                        </div>
                      )}
                      <h3 className="text-xl font-bold text-gray-800 mb-2">{page.title}</h3>
                      <p className="text-gray-600 leading-relaxed">{page.description}</p>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* End decoration */}
        <div className="relative pl-20 pt-8">
          <div className="absolute left-5 top-8 w-7 h-7 rounded-full bg-linear-to-br from-pink-400 to-purple-400 flex items-center justify-center">
            <Heart className="w-4 h-4 text-white" />
          </div>
          <p className="text-gray-400 italic">More memories to come...</p>
        </div>
      </div>
    </div>
  );
}
