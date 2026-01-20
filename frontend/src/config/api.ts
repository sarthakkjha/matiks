/**
 * API Configuration
 * Uses environment variable for production, falls back to localhost for development
 */

// Get API URL from environment or use localhost for development
const getApiUrl = (): string => {
  // Check for Expo environment variable (works in Expo/React Native)
  if (typeof process !== 'undefined' && process.env?.EXPO_PUBLIC_API_URL) {
    return process.env.EXPO_PUBLIC_API_URL;
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
