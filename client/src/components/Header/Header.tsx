import { Link } from 'react-router-dom';
import { UserButton } from '@clerk/clerk-react';
import { Baby, Upload, BookOpen, Sparkles } from 'lucide-react';
import { useAppSelector } from '../../store/hooks';

interface HeaderProps {
  currentPath: string;
}

export function Header({ currentPath }: HeaderProps) {
  const clusters = useAppSelector((state) => state.clusters.clusters);
  const pages = useAppSelector((state) => state.pages.pages);
  const isUploadView = currentPath === '/' || currentPath === '/upload';
  const isDraftsView = currentPath === '/drafts';
  const isBookView = currentPath.startsWith('/book');

  return (
    <header className="bg-white/80 backdrop-blur-sm border-b border-pink-100 sticky top-0 z-50">
      <div className="max-w-full px-6 py-4">
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
            <Link
              to="/upload"
              className={`px-4 py-2 rounded-xl font-medium text-sm flex items-center gap-2 transition-all
                ${isUploadView 
                  ? 'bg-linear-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
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
                    ? 'bg-linear-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
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
                  ? 'bg-linear-to-r from-pink-500 to-purple-500 text-white shadow-lg shadow-pink-500/25' 
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
            
            {/* User Button */}
            <div className="ml-4 pl-4 border-l border-pink-200">
              <UserButton 
                afterSignOutUrl="/"
                appearance={{
                  elements: {
                    avatarBox: 'w-9 h-9',
                  },
                }}
              />
            </div>
          </nav>
        </div>
      </div>
    </header>
  );
}
