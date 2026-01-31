import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import type { PhotoCluster } from '../../types/photo';

interface ClustersState {
  clusters: PhotoCluster[];
  isLoading: boolean;
  error: string | null;
}

const initialState: ClustersState = {
  clusters: [],
  isLoading: false,
  error: null,
};

const clustersSlice = createSlice({
  name: 'clusters',
  initialState,
  reducers: {
    setClusters: (state, action: PayloadAction<PhotoCluster[]>) => {
      state.clusters = action.payload;
      state.error = null;
    },
    addCluster: (state, action: PayloadAction<PhotoCluster>) => {
      state.clusters.push(action.payload);
    },
    addClusters: (state, action: PayloadAction<PhotoCluster[]>) => {
      state.clusters.push(...action.payload);
    },
    updateCluster: (state, action: PayloadAction<PhotoCluster>) => {
      const index = state.clusters.findIndex((c) => c.id === action.payload.id);
      if (index !== -1) {
        state.clusters[index] = action.payload;
      }
    },
    removeCluster: (state, action: PayloadAction<string>) => {
      state.clusters = state.clusters.filter((c) => c.id !== action.payload);
    },
    removeClusterByDraftId: (state, action: PayloadAction<string>) => {
      state.clusters = state.clusters.filter((c) => c.draftId !== action.payload);
    },
    clearClusters: (state) => {
      state.clusters = [];
    },
    setClustersLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setClustersError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

export const {
  setClusters,
  addCluster,
  addClusters,
  updateCluster,
  removeCluster,
  removeClusterByDraftId,
  clearClusters,
  setClustersLoading,
  setClustersError,
} = clustersSlice.actions;

export default clustersSlice.reducer;
