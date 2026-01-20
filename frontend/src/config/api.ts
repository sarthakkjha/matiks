/**
 * API Configuration
 * Uses environment variable for production, falls back to localhost for development
 */

// Get API URL from environment or use localhost for development
const getApiUrl = (): string => {
  // Expo replaces this with the string value at build time.
  // We access it directly to ensure the replacement works even if 'process' is not polyfilled.
  const vercelUrl = process.env.EXPO_PUBLIC_API_URL;
  if (vercelUrl) {
    return vercelUrl;
  }

  // Fallback for local development
  return 'http://localhost:3000/api';
};

export const API_CONFIG = {
  BASE_URL: getApiUrl(),
  TIMEOUT: 10000,
};

// API Endpoints
export const ENDPOINTS = {
  LEADERBOARD: '/leaderboard',
  LEADERBOARD_TOP: (n: number) => `/leaderboard/top/${n}`,
  USER_SEARCH: '/users/search',
  USER_SEARCH_PARTIAL: '/users/search-partial',
  USER_CREATE: '/users',
  USER_BY_ID: (id: string) => `/users/${id}`,
};
