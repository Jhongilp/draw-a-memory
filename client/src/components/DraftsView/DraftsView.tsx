import { Link } from 'react-router-dom';
import { Sparkles } from 'lucide-react';
import { PageDraftEditor } from '../PageDraftEditor';
import type { PageDraft } from '../../types/photo';
import { useAppSelector } from '../../store/hooks';

interface DraftsViewProps {
  onApprove: (draft: PageDraft) => void;
  onDiscard: (clusterId: string) => void;
}

export function DraftsView({ onApprove, onDiscard }: DraftsViewProps) {
  const clusters = useAppSelector((state) => state.clusters.clusters);
  return (
    <main className="flex-1 overflow-auto p-10">
      <div className="max-w-xl mx-auto space-y-8">
        <div className="text-center mb-10">
          <h2 className="text-3xl font-bold text-gray-800 mb-3">
            Review Your Drafts
          </h2>
          <p className="text-gray-500">
            Edit the AI suggestions and add them to your memory book
          </p>
        </div>
        
        {clusters.length === 0 ? (
          <div className="text-center py-16">
            <Sparkles className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <p className="text-gray-500">No drafts to review</p>
            <Link to="/upload" className="text-pink-500 hover:text-pink-600 underline text-sm mt-2 inline-block">
              Upload some photos
            </Link>
          </div>
        ) : (
          clusters.map((cluster) => (
            <PageDraftEditor
              key={cluster.id}
              cluster={cluster}
              onApprove={onApprove}
              onDiscard={() => onDiscard(cluster.id)}
            />
          ))
        )}
      </div>
    </main>
  );
}
