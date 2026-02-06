import { useState, useMemo } from 'react';
import { Edit3, Check, X, Sparkles, Calendar, Trash2 } from 'lucide-react';
import type { PhotoCluster, PageDraft, Theme, Photo } from '../../types/photo';
import { getPhotoUrl } from '../../api/photoApi';

interface PageDraftEditorProps {
  cluster: PhotoCluster;
  onApprove: (draft: PageDraft) => void;
  onDiscard: () => void;
}

const themeColors: Record<Theme, { bg: string; accent: string; label: string }> = {
  adventure: { bg: 'bg-amber-50', accent: 'text-amber-600', label: 'Adventure' },
  cozy: { bg: 'bg-orange-50', accent: 'text-orange-600', label: 'Cozy' },
  celebration: { bg: 'bg-pink-50', accent: 'text-pink-600', label: 'Celebration' },
  nature: { bg: 'bg-green-50', accent: 'text-green-600', label: 'Nature' },
  family: { bg: 'bg-purple-50', accent: 'text-purple-600', label: 'Family' },
  milestone: { bg: 'bg-blue-50', accent: 'text-blue-600', label: 'Milestone' },
  playful: { bg: 'bg-yellow-50', accent: 'text-yellow-600', label: 'Playful' },
  love: { bg: 'bg-rose-50', accent: 'text-rose-600', label: 'Love' },
  growth: { bg: 'bg-emerald-50', accent: 'text-emerald-600', label: 'Growth' },
  serene: { bg: 'bg-sky-50', accent: 'text-sky-600', label: 'Serene' },
};

const themes: Theme[] = ['adventure', 'cozy', 'celebration', 'nature', 'family', 'milestone', 'playful', 'love', 'growth', 'serene'];

// Dynamic layout component based on number of photos
function PhotoLayout({ photos, onRemove }: { photos: Photo[]; onRemove: (id: string) => void }) {
  const count = photos.length;

  if (count === 0) {
    return (
      <div className="h-64 flex items-center justify-center bg-gray-100 text-gray-400">
        No photos selected
      </div>
    );
  }

  if (count === 1) {
    return (
      <div className="relative h-80">
        <PhotoItem photo={photos[0]} onRemove={onRemove} className="w-full h-full" />
      </div>
    );
  }

  if (count === 2) {
    return (
      <div className="grid grid-cols-2 gap-1 h-64">
        {photos.map((photo) => (
          <PhotoItem key={photo.id} photo={photo} onRemove={onRemove} className="h-full" />
        ))}
      </div>
    );
  }

  if (count === 3) {
    return (
      <div className="grid grid-cols-3 grid-rows-2 gap-1 h-64">
        <div className="col-span-2 row-span-2">
          <PhotoItem photo={photos[0]} onRemove={onRemove} className="h-full" />
        </div>
        <PhotoItem photo={photos[1]} onRemove={onRemove} className="h-full" />
        <PhotoItem photo={photos[2]} onRemove={onRemove} className="h-full" />
      </div>
    );
  }

  if (count === 4) {
    return (
      <div className="grid grid-cols-2 grid-rows-2 gap-1 h-72">
        {photos.map((photo) => (
          <PhotoItem key={photo.id} photo={photo} onRemove={onRemove} className="h-full" />
        ))}
      </div>
    );
  }

  if (count === 5) {
    return (
      <div className="grid grid-cols-6 grid-rows-2 gap-1 h-72">
        <div className="col-span-3 row-span-2">
          <PhotoItem photo={photos[0]} onRemove={onRemove} className="h-full" />
        </div>
        <div className="col-span-3">
          <PhotoItem photo={photos[1]} onRemove={onRemove} className="h-full" />
        </div>
        {photos.slice(2, 5).map((photo) => (
          <div key={photo.id} className="col-span-1">
            <PhotoItem photo={photo} onRemove={onRemove} className="h-full" />
          </div>
        ))}
      </div>
    );
  }

  if (count === 6) {
    return (
      <div className="grid grid-cols-3 grid-rows-2 gap-1 h-72">
        {photos.map((photo) => (
          <PhotoItem key={photo.id} photo={photo} onRemove={onRemove} className="h-full" />
        ))}
      </div>
    );
  }

  // 7+ photos: masonry-like grid
  return (
    <div className="grid grid-cols-4 gap-1 auto-rows-[120px]">
      {photos.map((photo, idx) => {
        // Make first photo larger
        const isLarge = idx === 0;
        return (
          <div key={photo.id} className={isLarge ? 'col-span-2 row-span-2' : ''}>
            <PhotoItem photo={photo} onRemove={onRemove} className="h-full" />
          </div>
        );
      })}
    </div>
  );
}

function PhotoItem({ 
  photo, 
  onRemove, 
  className = '' 
}: { 
  photo: Photo; 
  onRemove: (id: string) => void; 
  className?: string;
}) {
  return (
    <div className={`relative group ${className}`}>
      <img
        src={getPhotoUrl(photo.path)}
        alt=""
        className="w-full h-full object-cover"
      />
      <button
        onClick={() => onRemove(photo.id)}
        className="absolute top-2 right-2 p-1.5 bg-red-500/80 hover:bg-red-600 text-white rounded-full opacity-0 group-hover:opacity-100 transition-opacity shadow-lg"
        title="Remove photo"
      >
        <Trash2 className="w-4 h-4" />
      </button>
    </div>
  );
}

export function PageDraftEditor({ cluster, onApprove, onDiscard }: PageDraftEditorProps) {
  const [title, setTitle] = useState(cluster.suggestedTitle || cluster.title);
  const [description, setDescription] = useState(cluster.suggestedDescription || cluster.description);
  const [theme, setTheme] = useState<Theme>(cluster.suggestedTheme || cluster.theme || 'family');
  const [isEditing, setIsEditing] = useState(false);
  const [discardedPhotoIds, setDiscardedPhotoIds] = useState<Set<string>>(new Set());

  const themeStyle = themeColors[theme] || themeColors.family;
  
  // Filter out discarded photos
  const photos = useMemo(() => 
    (cluster.photos || []).filter(p => !discardedPhotoIds.has(p.id)), 
    [cluster.photos, discardedPhotoIds]
  );
  
  const dateDisplay = cluster.dateRange || cluster.date || '';
  const ageDisplay = cluster.ageString || '';

  const handleRemovePhoto = (photoId: string) => {
    setDiscardedPhotoIds(prev => new Set([...prev, photoId]));
  };

  const handleRestoreAll = () => {
    setDiscardedPhotoIds(new Set());
  };

  const handleApprove = () => {
    if (photos.length === 0) return;
    
    const draft: PageDraft = {
      id: cluster.draftId || crypto.randomUUID(), // Use server's draft ID if available
      clusterId: cluster.id,
      photoIds: photos.map(p => p.id), // Use only the non-discarded photos
      title,
      description,
      theme,
      photos: photos,
      dateRange: dateDisplay,
      ageString: ageDisplay,
      backgroundPath: cluster.backgroundPath,
      status: 'approved',
      createdAt: new Date().toISOString(),
      approvedAt: new Date().toISOString(),
    };
    onApprove(draft);
  };

  const discardedCount = discardedPhotoIds.size;
  const backgroundUrl = cluster.backgroundPath ? getPhotoUrl(cluster.backgroundPath) : null;

  return (
    <div className="rounded-3xl overflow-hidden shadow-xl border border-white/50 relative">
      {/* Background image with overlay */}
      {backgroundUrl ? (
        <>
          <div 
            className="absolute inset-0 bg-cover bg-center"
            style={{ backgroundImage: `url("${backgroundUrl}")` }}
          />
          <div className="absolute inset-0 bg-white/70" />
        </>
      ) : (
        <div className={`absolute inset-0 ${themeStyle.bg}`} />
      )}

      {/* Content wrapper (above background) */}
      <div className="relative z-10">
        {/* Dynamic photo layout */}
        <div className="relative">
          <PhotoLayout photos={photos} onRemove={handleRemovePhoto} />
        <div className="absolute inset-0 bg-linear-to-t from-black/60 via-transparent to-transparent pointer-events-none" />
        
        {/* Age badge */}
        {ageDisplay && (
          <div className="absolute top-4 left-4 z-10">
            <span className="px-3 py-1.5 bg-white/90 backdrop-blur-sm rounded-full text-sm font-medium text-gray-700 shadow-lg">
              {ageDisplay}
            </span>
          </div>
        )}

        {/* AI badge */}
        <div className="absolute top-4 right-4 z-10">
          <span className="px-3 py-1.5 bg-linear-to-r from-pink-500 to-purple-500 text-white rounded-full text-sm font-medium flex items-center gap-1.5 shadow-lg">
            <Sparkles className="w-4 h-4" />
            AI Draft
          </span>
        </div>

        {/* Date overlay */}
        {dateDisplay && (
          <div className="absolute bottom-4 left-4 flex items-center gap-2 text-white z-10">
            <Calendar className="w-4 h-4" />
            <span className="text-sm font-medium">{dateDisplay}</span>
          </div>
        )}

        {/* Hover hint */}
        <div className="absolute bottom-4 right-4 z-10">
          <span className="px-2 py-1 bg-black/50 backdrop-blur-sm text-white/80 rounded text-xs">
            Hover to remove photos
          </span>
        </div>
      </div>

      {/* Content */}
      <div className="p-6 relative z-20 bg-white/90 backdrop-blur-sm">
        {isEditing ? (
          <div className="space-y-4">
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full text-2xl font-bold bg-white rounded-xl px-4 py-2 border border-gray-200 focus:outline-none focus:ring-2 focus:ring-pink-300"
              placeholder="Page title..."
            />
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="w-full bg-white rounded-xl px-4 py-3 border border-gray-200 focus:outline-none focus:ring-2 focus:ring-pink-300 resize-none"
              placeholder="Describe this memory..."
            />
            
            {/* Theme selector */}
            <div>
              <label className="text-sm font-medium text-gray-600 mb-2 block">Theme</label>
              <div className="flex flex-wrap gap-2">
                {themes.map((t) => (
                  <button
                    key={t}
                    onClick={() => setTheme(t)}
                    className={`px-3 py-1.5 rounded-full text-sm font-medium transition-all
                      ${theme === t 
                        ? `${themeColors[t].bg} ${themeColors[t].accent} ring-2 ring-offset-1 ring-current` 
                        : 'bg-white text-gray-600 hover:bg-gray-50'
                      }`}
                  >
                    {themeColors[t].label}
                  </button>
                ))}
              </div>
            </div>

            <div className="flex gap-2 pt-2">
              <button
                onClick={() => setIsEditing(false)}
                className="flex-1 py-2 px-4 bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 transition-colors font-medium"
              >
                Done Editing
              </button>
            </div>
          </div>
        ) : (
          <>
            <div className="flex items-start justify-between mb-3">
              <h3 className="text-2xl font-bold text-gray-800">{title}</h3>
              <button
                onClick={() => setIsEditing(true)}
                className="p-2 text-gray-400 hover:text-gray-600 hover:bg-white rounded-lg transition-colors"
              >
                <Edit3 className="w-5 h-5" />
              </button>
            </div>
            
            <p className="text-gray-600 leading-relaxed mb-4">{description}</p>
            
            <div className="flex items-center flex-wrap gap-2 mb-6">
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${themeStyle.bg} ${themeStyle.accent}`}>
                {themeStyle.label}
              </span>
              <span className="text-sm text-gray-500">
                {photos.length} photo{photos.length !== 1 ? 's' : ''}
              </span>
              {discardedCount > 0 && (
                <button
                  onClick={handleRestoreAll}
                  className="text-sm text-pink-500 hover:text-pink-600 underline"
                >
                  Restore {discardedCount} removed
                </button>
              )}
            </div>

            {/* Action buttons */}
            <div className="flex gap-3">
              <button
                onClick={onDiscard}
                className="flex-1 py-3 px-4 border border-gray-200 text-gray-600 rounded-xl hover:bg-white transition-colors font-medium flex items-center justify-center gap-2"
              >
                <X className="w-5 h-5" />
                Discard
              </button>
              <button
                onClick={handleApprove}
                disabled={photos.length === 0}
                className="flex-1 py-3 px-4 bg-linear-to-r from-pink-500 to-purple-500 text-white rounded-xl hover:from-pink-600 hover:to-purple-600 transition-all font-medium flex items-center justify-center gap-2 shadow-lg shadow-pink-500/25 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:from-pink-500 disabled:hover:to-purple-500"
              >
                <Check className="w-5 h-5" />
                Add to Book
              </button>
            </div>
          </>
        )}
      </div>
      </div>
    </div>
  );
}
