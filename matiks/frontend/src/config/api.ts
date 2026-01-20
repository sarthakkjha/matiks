/**
 * API Configuration
 * Update the BASE_URL to match your backend server
 */

// For iOS Simulator: use localhost
// For Android Emulator: use 10.0.2.2
// For physical device: use your computer's IP address
export const API_CONFIG = {
  BASE_URL: 'http://localhost:3000/api',
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
