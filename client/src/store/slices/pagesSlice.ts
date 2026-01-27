import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit';
import type { PageDraft, Photo } from '../../types/photo';
import { getPhotos, getPages } from '../../api/photoApi';

interface PagesState {
  pages: PageDraft[];
  photos: Photo[];
  isLoading: boolean;
  error: string | null;
}

const initialState: PagesState = {
  pages: [],
  photos: [],
  isLoading: false,
  error: null,
};

// Async thunk to fetch photos and pages
export const fetchPagesData = createAsyncThunk(
  'pages/fetchPagesData',
  async (_, { rejectWithValue }) => {
    try {
      const [fetchedPhotos, fetchedPages] = await Promise.all([
        getPhotos(),
        getPages().catch(() => []),
      ]);

      const allPhotos = fetchedPhotos || [];

      // Populate photos in each page based on photoIds
      const pagesWithPhotos = (fetchedPages || []).map((page) => ({
        ...page,
        photos: page.photoIds
          ?.map((id) => allPhotos.find((p) => p.id === id))
          .filter((p): p is Photo => p !== undefined) || [],
      }));

      return { photos: allPhotos, pages: pagesWithPhotos };
    } catch (error) {
      return rejectWithValue(
        error instanceof Error ? error.message : 'Failed to load data'
      );
    }
  }
);

const pagesSlice = createSlice({
  name: 'pages',
  initialState,
  reducers: {
    setPages: (state, action: PayloadAction<PageDraft[]>) => {
      state.pages = action.payload;
      state.error = null;
    },
    addPage: (state, action: PayloadAction<PageDraft>) => {
      state.pages.push(action.payload);
    },
    updatePage: (state, action: PayloadAction<PageDraft>) => {
      const index = state.pages.findIndex((p) => p.id === action.payload.id);
      if (index !== -1) {
        state.pages[index] = action.payload;
      }
    },
    removePage: (state, action: PayloadAction<string>) => {
      state.pages = state.pages.filter((p) => p.id !== action.payload);
    },
    reorderPages: (state, action: PayloadAction<PageDraft[]>) => {
      state.pages = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    setPhotos: (state, action: PayloadAction<Photo[]>) => {
      state.photos = action.payload;
    },
    // Populate photos for pages based on a photos array
    populatePhotos: (state, action: PayloadAction<Photo[]>) => {
      const allPhotos = action.payload;
      state.pages = state.pages.map((page) => ({
        ...page,
        photos: page.photoIds
          ?.map((id) => allPhotos.find((p) => p.id === id))
          .filter((p): p is Photo => p !== undefined) || [],
      }));
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchPagesData.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchPagesData.fulfilled, (state, action) => {
        state.isLoading = false;
        state.photos = action.payload.photos;
        state.pages = action.payload.pages;
      })
      .addCase(fetchPagesData.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });
  },
});

export const {
  setPages,
  addPage,
  updatePage,
  removePage,
  reorderPages,
  setLoading,
  setError,
  setPhotos,
  populatePhotos,
} = pagesSlice.actions;

export default pagesSlice.reducer;
