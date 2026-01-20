import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { LeaderboardEntry } from '../types/leaderboard';

interface LeaderboardItemProps {
  entry: LeaderboardEntry;
  index: number;
}

/**
 * Individual leaderboard item component
 * Displays rank, username, and rating
 */
export const LeaderboardItem: React.FC<LeaderboardItemProps> = ({ entry, index }) => {
  const isTopThree = entry.rank <= 3;
  const backgroundColor = index % 2 === 0 ? '#ffffff' : '#f8f9fa';

  return (
    <View style={[styles.container, { backgroundColor }]}>
      <View style={styles.rankContainer}>
        <Text style={[styles.rank, isTopThree && styles.topRank]}>
          #{entry.rank}
        </Text>
        {isTopThree && <Text style={styles.medal}>{getMedal(entry.rank)}</Text>}
      </View>
      
      <View style={styles.userInfo}>
        <Text style={styles.username} numberOfLines={1}>
          {entry.username}
        </Text>
      </View>
      
      <View style={styles.ratingContainer}>
        <Text style={styles.rating}>{entry.rating}</Text>
        <Text style={styles.ratingLabel}>rating</Text>
      </View>
    </View>
  );
};

function getMedal(rank: number): string {
  switch (rank) {
    case 1:
      return '🥇';
    case 2:
      return '🥈';
    case 3:
      return '🥉';
    default:
      return '';
  }
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  rankContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    width: 80,
  },
  rank: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },
  topRank: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ff6b35',
  },
  medal: {
    fontSize: 20,
    marginLeft: 4,
  },
  userInfo: {
    flex: 1,
    marginHorizontal: 12,
  },
  username: {
    fontSize: 16,
    fontWeight: '500',
    color: '#333',
  },
  ratingContainer: {
    alignItems: 'flex-end',
    minWidth: 70,
  },
  rating: {
    fontSize: 18,
    fontWeight: '700',
    color: '#4a90e2',
  },
  ratingLabel: {
    fontSize: 12,
    color: '#666',
  },
});
