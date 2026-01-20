export interface LeaderboardEntry {
  rank: number;
  username: string;
  rating: number;
  userId: string;
}

export interface LeaderboardResponse {
  success: boolean;
  data: {
    entries: LeaderboardEntry[];
    totalUsers: number;
    totalPages: number;
    currentPage: number;
  };
}

export interface UserSearchResponse {
  success: boolean;
  data: {
    rank: number;
    username: string;
    rating: number;
  };
  error?: string;
}

export interface ApiError {
  success: false;
  error: string;
}
