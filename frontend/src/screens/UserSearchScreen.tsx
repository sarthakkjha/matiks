import React, { useState } from 'react';
import { View, Text, StyleSheet, ScrollView, FlatList } from 'react-native';
import { apiService } from '../services/apiService';
import { SearchBar } from '../components/SearchBar';
import { ErrorMessage } from '../components/ErrorMessage';

interface SearchResultUser {
  rank: number;
  username: string;
  rating: number;
  userId: string;
}

/**
 * User Search Screen
 * Allows searching for users by partial username match
 * Features:
 * - Search by partial username (e.g., "rahul" finds all users with "rahul" in name)
 * - Displays: Global Rank, Username, Rating for all matches
 * - Shows multiple results
 */
export const UserSearchScreen: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchResults, setSearchResults] = useState<SearchResultUser[]>([]);
  const [searchTerm, setSearchTerm] = useState<string>('');

  const handleSearch = async (username: string) => {
    try {
      setLoading(true);
      setError(null);
      setSearchResults([]);
      setSearchTerm(username);

      const response = await apiService.searchUsersPartial(username);

      if (response.success) {
        if (response.data.users.length === 0) {
          setError(`No users found matching "${username}"`);
        } else {
          setSearchResults(response.data.users);
        }
      } else {
        setError(response.error || 'Failed to search users');
      }
    } catch (err: any) {
      setError(err.message || 'Failed to search users');
    } finally {
      setLoading(false);
    }
  };

  const renderUserItem = ({ item, index }: { item: SearchResultUser; index: number }) => {
    const isTopThree = item.rank <= 3;
    const backgroundColor = index % 2 === 0 ? '#fff' : '#f8f9fa';

    return (
      <View style={[styles.userCard, { backgroundColor }]}>
        <View style={styles.rankBadge}>
          <Text style={[styles.rankText, isTopThree && styles.topRankText]}>
            #{item.rank}
          </Text>
          {isTopThree && (
            <Text style={styles.medal}>
              {item.rank === 1 ? '🥇' : item.rank === 2 ? '🥈' : '🥉'}
            </Text>
          )}
        </View>
        <View style={styles.userInfo}>
          <Text style={styles.usernameText}>{item.username}</Text>
        </View>
        <View style={styles.ratingInfo}>
          <Text style={styles.ratingValue}>{item.rating}</Text>
          <Text style={styles.ratingLabel}>rating</Text>
        </View>
      </View>
    );
  };

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.title}>🔍 User Search</Text>
        <Text style={styles.subtitle}>
          Search by username to see global ranks (e.g., "Alpha", "Beta")
        </Text>
      </View>

      {/* Search Bar */}
      <SearchBar
        onSearch={handleSearch}
        loading={loading}
        placeholder="Enter username to search..."
      />

      {/* Results */}
      {searchResults.length > 0 && (
        <View style={styles.resultsHeader}>
          <Text style={styles.resultsCount}>
            Found {searchResults.length} user{searchResults.length !== 1 ? 's' : ''} matching "{searchTerm}"
          </Text>
        </View>
      )}

      {/* Content */}
      {searchResults.length > 0 ? (
        <FlatList
          data={searchResults}
          keyExtractor={(item) => item.userId}
          renderItem={renderUserItem}
          contentContainerStyle={styles.listContent}
        />
      ) : (
        <ScrollView contentContainerStyle={styles.content}>
          {error && (
            <View style={styles.errorCard}>
              <Text style={styles.errorEmoji}>❌</Text>
              <Text style={styles.errorText}>{error}</Text>
              <Text style={styles.errorHint}>
                Try different search terms (e.g., "Alpha", "Beta", "Gamma")
              </Text>
            </View>
          )}

          {!error && !loading && (
            <View style={styles.placeholderCard}>
              <Text style={styles.placeholderEmoji}>👆</Text>
              <Text style={styles.placeholderText}>
                Search for users by username
              </Text>
              <Text style={styles.examplesText}>
                Try: "player", "shadow", "alpha", "ninja"
              </Text>
            </View>
          )}
        </ScrollView>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    backgroundColor: '#fff',
    padding: 20,
    borderBottomWidth: 2,
    borderBottomColor: '#4a90e2',
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    color: '#333',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 13,
    color: '#666',
    lineHeight: 18,
  },
  resultsHeader: {
    backgroundColor: '#e3f2fd',
    padding: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#90caf9',
  },
  resultsCount: {
    fontSize: 14,
    fontWeight: '600',
    color: '#1976d2',
    textAlign: 'center',
  },
  listContent: {
    flexGrow: 1,
  },
  userCard: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  rankBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    minWidth: 80,
  },
  rankText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },
  topRankText: {
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
  usernameText: {
    fontSize: 16,
    fontWeight: '500',
    color: '#333',
  },
  ratingInfo: {
    alignItems: 'flex-end',
    minWidth: 70,
  },
  ratingValue: {
    fontSize: 18,
    fontWeight: '700',
    color: '#4a90e2',
  },
  ratingLabel: {
    fontSize: 12,
    color: '#666',
  },
  content: {
    padding: 16,
    flexGrow: 1,
  },
  errorCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 30,
    alignItems: 'center',
  },
  errorEmoji: {
    fontSize: 48,
    marginBottom: 12,
  },
  errorText: {
    fontSize: 16,
    color: '#d32f2f',
    textAlign: 'center',
    marginBottom: 8,
  },
  errorHint: {
    fontSize: 14,
    color: '#999',
    textAlign: 'center',
  },
  placeholderCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 40,
    alignItems: 'center',
  },
  placeholderEmoji: {
    fontSize: 48,
    marginBottom: 12,
  },
  placeholderText: {
    fontSize: 16,
    color: '#999',
    textAlign: 'center',
    marginBottom: 8,
  },
  examplesText: {
    fontSize: 14,
    color: '#4a90e2',
    textAlign: 'center',
    fontStyle: 'italic',
  },
});
