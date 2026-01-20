import axios, { AxiosInstance, AxiosError } from 'axios';
import { API_CONFIG, ENDPOINTS } from '../config/api';
import { LeaderboardResponse, UserSearchResponse } from '../types/leaderboard';

/**
 * API Service for communicating with the Go backend.
 * 
 * DESIGN:
 * - Singleton instance for unified state management
 * - Axios interceptors for global error handling and logging
 * - Type-safe response handling with TypeScript interfaces
 */
class ApiService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_CONFIG.BASE_URL,
      timeout: API_CONFIG.TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor for logging
    this.client.interceptors.request.use(
      (config) => {
        console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => {
        console.error('[API] Request error:', error);
        return Promise.reject(error);
      }
    );

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => {
        console.log(`[API] Response:`, response.status);
        return response;
      },
      (error: AxiosError) => {
        console.error('[API] Response error:', error.message);
        return Promise.reject(this.handleError(error));
      }
    );
  }

  private handleError(error: AxiosError): Error {
    if (error.response) {
      // Server responded with error status
      const data = error.response.data as any;
      return new Error(data?.error || 'Server error occurred');
    } else if (error.request) {
      // Request made but no response
      return new Error('Network error - please check your connection');
    } else {
      // Something else happened
      return new Error(error.message || 'An unexpected error occurred');
    }
  }

  /**
   * Get leaderboard with pagination
   */
  async getLeaderboard(page: number = 1, limit: number = 50): Promise<LeaderboardResponse> {
    const response = await this.client.get<LeaderboardResponse>(ENDPOINTS.LEADERBOARD, {
      params: { page, limit },
    });
    return response.data;
  }

  /**
   * Get top N users
   */
  async getTopUsers(n: number = 10): Promise<LeaderboardResponse> {
    const response = await this.client.get<LeaderboardResponse>(
      ENDPOINTS.LEADERBOARD_TOP(n)
    );
    return response.data;
  }

  /**
   * Search for a user by exact username
   */
  async searchUser(username: string): Promise<UserSearchResponse> {
    const response = await this.client.get<UserSearchResponse>(ENDPOINTS.USER_SEARCH, {
      params: { username },
    });
    return response.data;
  }

  /**
   * Search for users by prefix match (case-insensitive).
   * Returns up to 500 results to show all matching users.
   */
  async searchUsersPartial(query: string): Promise<any> {
    const response = await this.client.get(ENDPOINTS.USER_SEARCH, {
      params: { prefix: query, limit: 500 },
    });
    return response.data;
  }

  /**
   * Update user rating (admin operation)
   */
  async updateUserRating(userId: string, rating: number): Promise<any> {
    const response = await this.client.put(ENDPOINTS.USER_BY_ID(userId) + '/score', {
      rating,
    });
    return response.data;
  }

  /**
   * Bulk update X users with random ratings
   */
  async bulkUpdateRandom(count: number): Promise<any> {
    const response = await this.client.post('/bulk-update/random', { count });
    return response.data;
  }

  /**
   * Bulk update X users to a specific rating
   */
  async bulkUpdateToValue(count: number, rating: number): Promise<any> {
    const response = await this.client.post('/bulk-update/value', { count, rating });
    return response.data;
  }
}

export const apiService = new ApiService();
