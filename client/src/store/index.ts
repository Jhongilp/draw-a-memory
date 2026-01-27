import { configureStore } from '@reduxjs/toolkit';
import { pagesReducer } from './slices';

export const store = configureStore({
  reducer: {
    pages: pagesReducer,
  },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
