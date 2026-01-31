import { configureStore } from '@reduxjs/toolkit';
import { pagesReducer, clustersReducer } from './slices';

export const store = configureStore({
  reducer: {
    pages: pagesReducer,
    clusters: clustersReducer,
  },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
