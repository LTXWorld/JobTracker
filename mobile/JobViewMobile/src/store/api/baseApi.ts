import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { API_CONFIG } from '@config';

// Base query with auth headers
const baseQuery = fetchBaseQuery({
  baseUrl: `${API_CONFIG.BASE_URL}/api/${API_CONFIG.VERSION}`,
  timeout: API_CONFIG.TIMEOUT,
  prepareHeaders: (headers, { getState }: any) => {
    const token = getState().auth.tokens?.accessToken;
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }
    headers.set('Content-Type', 'application/json');
    return headers;
  },
});

// Create the base API slice
export const baseApi = createApi({
  reducerPath: 'api',
  baseQuery,
  tagTypes: ['User', 'Application', 'Note', 'Reminder', 'Timeline', 'Stats'],
  endpoints: () => ({}),
});

// Export hooks for usage in functional components
export const {
  // Will be populated by individual API slices
} = baseApi;