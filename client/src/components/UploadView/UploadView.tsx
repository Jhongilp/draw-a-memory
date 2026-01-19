import { Sparkles } from 'lucide-react';
import { PhotoUpload } from '../PhotoUpload';
import type { Photo } from '../../types/photo';

interface UploadViewProps {
  isAnalyzing: boolean;
  onUploadComplete: (photos: Photo[]) => void;
}

export function UploadView({ isAnalyzing, onUploadComplete }: UploadViewProps) {


  
  return (
    <main className="flex-1 flex items-center justify-center p-10">
      <div className="max-w-2xl w-full">
        <div className="text-center mb-10">
          <h2 className="text-3xl font-bold text-gray-800 mb-3">
            Add New Memories
          </h2>
          <p className="text-gray-500">
            Upload photos and let AI create beautiful memory book pages
          </p>
        </div>
        
        {isAnalyzing ? (
          <div className="text-center py-16 bg-white rounded-3xl shadow-xl">
            <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-gradient-to-br from-pink-100 to-purple-100 flex items-center justify-center">
              <Sparkles className="w-10 h-10 text-pink-500 animate-pulse" />
            </div>
            <h3 className="text-xl font-semibold text-gray-700 mb-2">
              Creating Magic...
            </h3>
            <p className="text-gray-500">
              AI is analyzing your photos and crafting the perfect memory
            </p>
          </div>
        ) : (
          <PhotoUpload onUploadComplete={onUploadComplete} />
        )}
      </div>
    </main>
  );
}
