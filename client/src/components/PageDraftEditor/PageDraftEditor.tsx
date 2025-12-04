import { useState } from 'react';
import { Edit3, Check, X, Sparkles, Calendar } from 'lucide-react';
import type { PhotoCluster, PageDraft, Theme } from '../../types/photo';
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

export function PageDraftEditor({ cluster, onApprove, onDiscard }: PageDraftEditorProps) {
  const [title, setTitle] = useState(cluster.suggestedTitle || cluster.title);
  const [description, setDescription] = useState(cluster.suggestedDescription || cluster.description);
  const [theme, setTheme] = useState<Theme>(cluster.suggestedTheme || cluster.theme || 'family');
  const [isEditing, setIsEditing] = useState(false);

  const themeStyle = themeColors[theme] || themeColors.family;
  const photos = cluster.photos || [];
  const dateDisplay = cluster.dateRange || cluster.date || '';
  const ageDisplay = cluster.ageString || '';

  const handleApprove = () => {
    const draft: PageDraft = {
      id: cluster.draftId || crypto.randomUUID(), // Use server's draft ID if available
      clusterId: cluster.id,
      photoIds: cluster.photoIds || photos.map(p => p.id),
      title,
      description,
      theme,
      photos: photos,
      dateRange: dateDisplay,
      ageString: ageDisplay,
      status: 'approved',
      createdAt: new Date().toISOString(),
      approvedAt: new Date().toISOString(),
    };
    onApprove(draft);
  };

  return (
    <div className={`rounded-3xl overflow-hidden shadow-xl ${themeStyle.bg} border border-white/50`}>
      {/* Header with photos preview */}
      <div className="relative h-64">
        <div className="absolute inset-0 grid grid-cols-3 gap-1">
          {photos.slice(0, 3).map((photo, idx) => (
            <div key={photo.id} className={`relative ${idx === 0 ? 'col-span-2 row-span-2' : ''}`}>
              <img
                src={getPhotoUrl(photo.path)}
                alt=""
                className="w-full h-full object-cover"
              />
            </div>
          ))}
        </div>
        <div className="absolute inset-0 bg-linear-to-t from-black/60 via-transparent to-transparent" />
        
        {/* Age badge */}
        {ageDisplay && (
          <div className="absolute top-4 left-4">
            <span className="px-3 py-1.5 bg-white/90 backdrop-blur-sm rounded-full text-sm font-medium text-gray-700 shadow-lg">
              {ageDisplay}
            </span>
          </div>
        )}

        {/* AI badge */}
        <div className="absolute top-4 right-4">
          <span className="px-3 py-1.5 bg-linear-to-r from-pink-500 to-purple-500 text-white rounded-full text-sm font-medium flex items-center gap-1.5 shadow-lg">
            <Sparkles className="w-4 h-4" />
            AI Draft
          </span>
        </div>

        {/* Date overlay */}
        {dateDisplay && (
          <div className="absolute bottom-4 left-4 flex items-center gap-2 text-white">
            <Calendar className="w-4 h-4" />
            <span className="text-sm font-medium">{dateDisplay}</span>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-6">
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
            
            <div className="flex items-center gap-2 mb-6">
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${themeStyle.bg} ${themeStyle.accent}`}>
                {themeStyle.label}
              </span>
              <span className="text-sm text-gray-500">
                {photos.length} photo{photos.length !== 1 ? 's' : ''}
              </span>
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
                className="flex-1 py-3 px-4 bg-linear-to-r from-pink-500 to-purple-500 text-white rounded-xl hover:from-pink-600 hover:to-purple-600 transition-all font-medium flex items-center justify-center gap-2 shadow-lg shadow-pink-500/25"
              >
                <Check className="w-5 h-5" />
                Add to Book
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
