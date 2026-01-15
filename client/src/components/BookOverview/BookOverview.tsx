import { Link } from 'react-router-dom';
import { Book, Heart, ArrowRight } from 'lucide-react';
import type { PageDraft } from '../../types/photo';
import { getPhotoUrl } from '../../api/photoApi';

interface BookOverviewProps {
  pages: PageDraft[];
}

export function BookOverview({ pages }: BookOverviewProps) {
  if (pages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center py-16">
          <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-gradient-to-br from-pink-100 to-purple-100 flex items-center justify-center">
            <Book className="w-10 h-10 text-pink-400" />
          </div>
          <h3 className="text-xl font-semibold text-gray-700 mb-2">
            Your memory book is empty
          </h3>
          <p className="text-gray-500 max-w-md mx-auto mb-6">
            Upload some photos and let the AI create beautiful pages for your baby's memory book!
          </p>
          <Link
            to="/"
            className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-pink-500 to-purple-500 text-white rounded-xl font-medium hover:from-pink-600 hover:to-purple-600 transition-all shadow-lg shadow-pink-500/25"
          >
            Upload Photos
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    );
  }

  const firstPage = pages[0];
  const thumbnail = firstPage.photos?.[0]?.path 
    ? getPhotoUrl(firstPage.photos[0].path) 
    : null;

  return (
    <div className="flex-1 flex items-center justify-center p-8">
      <div className="text-center max-w-lg">
        {/* Book cover preview */}
        <div className="relative inline-block mb-8">
          <div className="w-64 h-80 bg-gradient-to-br from-pink-200 to-purple-200 rounded-2xl shadow-2xl overflow-hidden transform rotate-2 hover:rotate-0 transition-transform duration-300">
            {thumbnail ? (
              <img src={thumbnail} alt="" className="w-full h-full object-cover" />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <Book className="w-16 h-16 text-white/50" />
              </div>
            )}
            <div className="absolute inset-0 bg-gradient-to-t from-black/50 to-transparent" />
            <div className="absolute bottom-4 left-4 right-4 text-white text-left">
              <h3 className="font-bold text-lg">Baby's First Year</h3>
              <p className="text-sm text-white/80">{pages.length} memories</p>
            </div>
          </div>
          
          {/* Decorative badge */}
          <div className="absolute -top-3 -right-3 bg-gradient-to-r from-pink-500 to-purple-500 text-white rounded-full p-3 shadow-lg">
            <Heart className="w-5 h-5" />
          </div>
        </div>

        <h2 className="text-2xl font-bold text-gray-800 mb-3">
          Welcome to Your Memory Book
        </h2>
        <p className="text-gray-500 mb-8">
          Select a page from the sidebar to view, or start from the beginning.
        </p>

        <Link
          to={`/book/page/${firstPage.id}`}
          className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-pink-500 to-purple-500 text-white rounded-xl font-medium hover:from-pink-600 hover:to-purple-600 transition-all shadow-lg shadow-pink-500/25"
        >
          Start Reading
          <ArrowRight className="w-4 h-4" />
        </Link>
      </div>
    </div>
  );
}
