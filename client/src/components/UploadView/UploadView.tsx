import { useState } from "react";
import { Sparkles } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { PhotoUpload } from "../PhotoUpload";
import { useAppDispatch } from "../../store/hooks";
import type { Photo } from "../../types/photo";
import { setPhotos, setClusters } from "../../store/slices";
import { analyzePhotos } from "../../api/photoApi";

export function UploadView() {
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const handleUploadComplete = async (newPhotos: Photo[]) => {
    const allPhotos = [...newPhotos];
    dispatch(setPhotos(allPhotos));

    setIsAnalyzing(true);
    try {
      const photoIds = newPhotos.map((p) => p.id);
      const response = await analyzePhotos(photoIds);
      if (response.clusters && response.clusters.length > 0) {
        const clustersWithPhotos = response.clusters.map((cluster) => {
          const serverDraft = response.drafts?.find(
            (d) => d.clusterId === cluster.id,
          );
          return {
            ...cluster,
            draftId: serverDraft?.id,
            backgroundPath:
              serverDraft?.backgroundPath || cluster.backgroundPath,
            photos: cluster.photoIds
              .map(
                (id) =>
                  allPhotos.find((p) => p.id === id) ||
                  newPhotos.find((p) => p.id === id),
              )
              .filter(Boolean) as Photo[],
            suggestedTitle: cluster.title,
            suggestedDescription: cluster.description,
            suggestedTheme: cluster.theme,
            dateRange: cluster.date,
            ageString: "",
            status: "draft" as const,
          };
        });
        dispatch(setClusters(clustersWithPhotos));
        navigate("/drafts");
      }
    } catch (error) {
      console.error("Failed to analyze photos:", error);
    } finally {
      setIsAnalyzing(false);
    }
  };

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
            <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-linear-to-br from-pink-100 to-purple-100 flex items-center justify-center">
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
          <PhotoUpload onUploadComplete={handleUploadComplete} />
        )}
      </div>
    </main>
  );
}
