import React, { useState } from 'react';
import {
    View,
    Text,
    TextInput,
    StyleSheet,
    ScrollView,
    TouchableOpacity,
    ActivityIndicator,
    Alert,
} from 'react-native';
import { apiService } from '../services/apiService';

interface UserResult {
    userId: string;
    username: string;
    rating: number;
    rank: number;
}

/**
 * Admin Screen
 * Allows administrators to search for users and update their ratings
 */
export const AdminScreen: React.FC = () => {
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState<UserResult[]>([]);
    const [selectedUser, setSelectedUser] = useState<UserResult | null>(null);
    const [newRating, setNewRating] = useState('');
    const [loading, setLoading] = useState(false);
    const [updating, setUpdating] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    // Mass update state
    const [bulkCount, setBulkCount] = useState('100');
    const [bulkRating, setBulkRating] = useState('2500');
    const [bulkLoading, setBulkLoading] = useState(false);
    const [bulkResult, setBulkResult] = useState<{
        updated: number;
        durationMs: number;
        updatesPerSec: number;
    } | null>(null);

    const handleSearch = async () => {
        const trimmed = searchQuery.trim();
        if (!trimmed) return;

        try {
            setLoading(true);
            setMessage(null);
            setSelectedUser(null);

            const response = await apiService.searchUsersPartial(trimmed);

            if (response.success && response.data.users.length > 0) {
                setSearchResults(response.data.users);
            } else {
                setSearchResults([]);
                setMessage({ type: 'error', text: `No users found matching "${trimmed}"` });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err.message || 'Search failed' });
            setSearchResults([]);
        } finally {
            setLoading(false);
        }
    };

    const handleSelectUser = (user: UserResult) => {
        setSelectedUser(user);
        setNewRating(user.rating.toString());
        setMessage(null);
    };

    const handleUpdateRating = async () => {
        if (!selectedUser) return;

        const rating = parseInt(newRating, 10);
        if (isNaN(rating) || rating < 100 || rating > 5000) {
            setMessage({ type: 'error', text: 'Rating must be between 100 and 5000' });
            return;
        }

        try {
            setUpdating(true);
            setMessage(null);

            const response = await apiService.updateUserRating(selectedUser.userId, rating);

            if (response.success) {
                setMessage({
                    type: 'success',
                    text: `Updated ${selectedUser.username}'s rating to ${rating}. New rank: #${response.data.user.rank}`,
                });

                // Update the selected user with new data
                setSelectedUser({
                    ...selectedUser,
                    rating: response.data.user.rating,
                    rank: response.data.user.rank,
                });

                // Update in search results too
                setSearchResults((prev) =>
                    prev.map((u) =>
                        u.userId === selectedUser.userId
                            ? { ...u, rating: response.data.user.rating, rank: response.data.user.rank }
                            : u
                    )
                );
            } else {
                setMessage({ type: 'error', text: response.error || 'Update failed' });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err.message || 'Update failed' });
        } finally {
            setUpdating(false);
        }
    };

    // Mass update handlers
    const handleBulkUpdateRandom = async () => {
        const count = parseInt(bulkCount, 10);
        if (isNaN(count) || count < 1) {
            setMessage({ type: 'error', text: 'Enter a valid number of users' });
            return;
        }

        try {
            setBulkLoading(true);
            setMessage(null);
            setBulkResult(null);

            const response = await apiService.bulkUpdateRandom(count);
            if (response.success) {
                setBulkResult(response.data);
                setMessage({
                    type: 'success',
                    text: `Updated ${response.data.updated} users in ${response.data.durationMs}ms (${Math.round(response.data.updatesPerSec)} updates/sec)`,
                });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err.message || 'Bulk update failed' });
        } finally {
            setBulkLoading(false);
        }
    };

    const handleBulkUpdateToValue = async () => {
        const count = parseInt(bulkCount, 10);
        const rating = parseInt(bulkRating, 10);

        if (isNaN(count) || count < 1) {
            setMessage({ type: 'error', text: 'Enter a valid number of users' });
            return;
        }
        if (isNaN(rating) || rating < 100 || rating > 5000) {
            setMessage({ type: 'error', text: 'Rating must be between 100 and 5000' });
            return;
        }

        try {
            setBulkLoading(true);
            setMessage(null);
            setBulkResult(null);

            const response = await apiService.bulkUpdateToValue(count, rating);
            if (response.success) {
                setBulkResult(response.data);
                setMessage({
                    type: 'success',
                    text: `Updated ${response.data.updated} users to ${rating} in ${response.data.durationMs}ms (${Math.round(response.data.updatesPerSec)} updates/sec)`,
                });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err.message || 'Bulk update failed' });
        } finally {
            setBulkLoading(false);
        }
    };

    return (
        <View style={styles.container}>
            {/* Header */}
            <View style={styles.header}>
                <Text style={styles.title}>⚙️ Admin Panel</Text>
                <Text style={styles.subtitle}>Search users and update their ratings</Text>
            </View>

            {/* Search Section */}
            <View style={styles.searchSection}>
                <View style={styles.searchRow}>
                    <TextInput
                        style={styles.searchInput}
                        value={searchQuery}
                        onChangeText={setSearchQuery}
                        placeholder="Enter username prefix..."
                        placeholderTextColor="#999"
                        autoCapitalize="none"
                        returnKeyType="search"
                        onSubmitEditing={handleSearch}
                    />
                    <TouchableOpacity
                        style={[styles.searchButton, loading && styles.buttonDisabled]}
                        onPress={handleSearch}
                        disabled={loading || !searchQuery.trim()}
                    >
                        {loading ? (
                            <ActivityIndicator color="#fff" size="small" />
                        ) : (
                            <Text style={styles.buttonText}>Search</Text>
                        )}
                    </TouchableOpacity>
                </View>
            </View>

            {/* Message */}
            {message && (
                <View
                    style={[
                        styles.messageBox,
                        message.type === 'success' ? styles.successBox : styles.errorBox,
                    ]}
                >
                    <Text
                        style={[
                            styles.messageText,
                            message.type === 'success' ? styles.successText : styles.errorText,
                        ]}
                    >
                        {message.type === 'success' ? '✅ ' : '❌ '}
                        {message.text}
                    </Text>
                </View>
            )}

            <ScrollView style={styles.content}>
                {/* Search Results */}
                {searchResults.length > 0 && (
                    <View style={styles.resultsSection}>
                        <Text style={styles.sectionTitle}>
                            Search Results ({searchResults.length})
                        </Text>
                        {searchResults.map((user) => (
                            <TouchableOpacity
                                key={user.userId}
                                style={[
                                    styles.userCard,
                                    selectedUser?.userId === user.userId && styles.selectedCard,
                                ]}
                                onPress={() => handleSelectUser(user)}
                            >
                                <View style={styles.userCardLeft}>
                                    <Text style={styles.userRank}>#{user.rank}</Text>
                                    <Text style={styles.userName}>{user.username}</Text>
                                </View>
                                <Text style={styles.userRating}>{user.rating}</Text>
                            </TouchableOpacity>
                        ))}
                    </View>
                )}

                {/* Edit Section */}
                {selectedUser && (
                    <View style={styles.editSection}>
                        <Text style={styles.sectionTitle}>Edit Rating</Text>
                        <View style={styles.editCard}>
                            <Text style={styles.editUserName}>{selectedUser.username}</Text>
                            <Text style={styles.editUserInfo}>
                                Current Rank: #{selectedUser.rank} | Current Rating: {selectedUser.rating}
                            </Text>

                            <Text style={styles.inputLabel}>New Rating (100-5000):</Text>
                            <TextInput
                                style={styles.ratingInput}
                                value={newRating}
                                onChangeText={setNewRating}
                                keyboardType="numeric"
                                placeholder="Enter new rating"
                                placeholderTextColor="#999"
                            />

                            <TouchableOpacity
                                style={[styles.updateButton, updating && styles.buttonDisabled]}
                                onPress={handleUpdateRating}
                                disabled={updating}
                            >
                                {updating ? (
                                    <ActivityIndicator color="#fff" size="small" />
                                ) : (
                                    <Text style={styles.buttonText}>Update Rating</Text>
                                )}
                            </TouchableOpacity>
                        </View>
                    </View>
                )}

                {/* Instructions */}
                {!selectedUser && searchResults.length === 0 && !loading && (
                    <View style={styles.instructions}>
                        <Text style={styles.instructionEmoji}>📝</Text>
                        <Text style={styles.instructionText}>
                            Search for a user by their username prefix, then select them to update their rating.
                        </Text>
                        <Text style={styles.instructionHint}>
                            Try: "player", "shadow", "ninja", "alpha"
                        </Text>
                    </View>
                )}

                {/* Mass Update Section */}
                <View style={styles.massUpdateSection}>
                    <Text style={styles.sectionTitle}>⚡ Mass Update (Demo High-Throughput)</Text>

                    <View style={styles.massUpdateCard}>
                        <Text style={styles.inputLabel}>Number of Users to Update:</Text>
                        <TextInput
                            style={styles.massInput}
                            value={bulkCount}
                            onChangeText={setBulkCount}
                            keyboardType="numeric"
                            placeholder="e.g., 100"
                            placeholderTextColor="#999"
                        />

                        <Text style={styles.inputLabel}>Target Rating (for value update):</Text>
                        <TextInput
                            style={styles.massInput}
                            value={bulkRating}
                            onChangeText={setBulkRating}
                            keyboardType="numeric"
                            placeholder="100-5000"
                            placeholderTextColor="#999"
                        />

                        <View style={styles.massButtonRow}>
                            <TouchableOpacity
                                style={[styles.massButton, styles.randomButton, bulkLoading && styles.buttonDisabled]}
                                onPress={handleBulkUpdateRandom}
                                disabled={bulkLoading}
                            >
                                {bulkLoading ? (
                                    <ActivityIndicator color="#fff" size="small" />
                                ) : (
                                    <Text style={styles.buttonText}>🎲 Random Ratings</Text>
                                )}
                            </TouchableOpacity>

                            <TouchableOpacity
                                style={[styles.massButton, styles.valueButton, bulkLoading && styles.buttonDisabled]}
                                onPress={handleBulkUpdateToValue}
                                disabled={bulkLoading}
                            >
                                {bulkLoading ? (
                                    <ActivityIndicator color="#fff" size="small" />
                                ) : (
                                    <Text style={styles.buttonText}>🎯 Set to Value</Text>
                                )}
                            </TouchableOpacity>
                        </View>

                        {bulkResult && (
                            <View style={styles.resultBox}>
                                <Text style={styles.resultTitle}>Performance Results:</Text>
                                <Text style={styles.resultText}>✅ Updated: {bulkResult.updated} users</Text>
                                <Text style={styles.resultText}>⏱️ Duration: {bulkResult.durationMs}ms</Text>
                                <Text style={styles.resultHighlight}>
                                    🚀 Speed: {Math.round(bulkResult.updatesPerSec).toLocaleString()} updates/sec
                                </Text>
                            </View>
                        )}
                    </View>
                </View>
            </ScrollView>
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
        borderBottomColor: '#ff9800',
    },
    title: {
        fontSize: 28,
        fontWeight: '700',
        color: '#333',
        marginBottom: 4,
    },
    subtitle: {
        fontSize: 14,
        color: '#666',
    },
    searchSection: {
        backgroundColor: '#fff',
        padding: 16,
        borderBottomWidth: 1,
        borderBottomColor: '#e0e0e0',
    },
    searchRow: {
        flexDirection: 'row',
    },
    searchInput: {
        flex: 1,
        height: 44,
        backgroundColor: '#f5f5f5',
        borderRadius: 8,
        paddingHorizontal: 12,
        fontSize: 16,
        marginRight: 8,
    },
    searchButton: {
        backgroundColor: '#ff9800',
        borderRadius: 8,
        paddingHorizontal: 20,
        justifyContent: 'center',
        alignItems: 'center',
        minWidth: 80,
        height: 44,
    },
    buttonDisabled: {
        backgroundColor: '#ccc',
    },
    buttonText: {
        color: '#fff',
        fontSize: 16,
        fontWeight: '600',
    },
    messageBox: {
        margin: 16,
        marginBottom: 0,
        padding: 12,
        borderRadius: 8,
    },
    successBox: {
        backgroundColor: '#e8f5e9',
        borderColor: '#4caf50',
        borderWidth: 1,
    },
    errorBox: {
        backgroundColor: '#ffebee',
        borderColor: '#f44336',
        borderWidth: 1,
    },
    messageText: {
        fontSize: 14,
        textAlign: 'center',
    },
    successText: {
        color: '#2e7d32',
    },
    errorText: {
        color: '#c62828',
    },
    content: {
        flex: 1,
        padding: 16,
    },
    resultsSection: {
        marginBottom: 20,
    },
    sectionTitle: {
        fontSize: 16,
        fontWeight: '600',
        color: '#333',
        marginBottom: 12,
    },
    userCard: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        backgroundColor: '#fff',
        padding: 14,
        borderRadius: 8,
        marginBottom: 8,
        borderWidth: 2,
        borderColor: 'transparent',
    },
    selectedCard: {
        borderColor: '#ff9800',
        backgroundColor: '#fff8e1',
    },
    userCardLeft: {
        flexDirection: 'row',
        alignItems: 'center',
    },
    userRank: {
        fontSize: 14,
        fontWeight: '600',
        color: '#666',
        minWidth: 50,
    },
    userName: {
        fontSize: 16,
        fontWeight: '500',
        color: '#333',
    },
    userRating: {
        fontSize: 16,
        fontWeight: '700',
        color: '#4a90e2',
    },
    editSection: {
        marginTop: 8,
    },
    editCard: {
        backgroundColor: '#fff',
        padding: 20,
        borderRadius: 12,
        borderWidth: 2,
        borderColor: '#ff9800',
    },
    editUserName: {
        fontSize: 20,
        fontWeight: '700',
        color: '#333',
        marginBottom: 4,
    },
    editUserInfo: {
        fontSize: 14,
        color: '#666',
        marginBottom: 16,
    },
    inputLabel: {
        fontSize: 14,
        fontWeight: '600',
        color: '#333',
        marginBottom: 8,
    },
    ratingInput: {
        height: 50,
        backgroundColor: '#f5f5f5',
        borderRadius: 8,
        paddingHorizontal: 16,
        fontSize: 18,
        marginBottom: 16,
        textAlign: 'center',
    },
    updateButton: {
        backgroundColor: '#4caf50',
        borderRadius: 8,
        paddingVertical: 14,
        alignItems: 'center',
    },
    instructions: {
        backgroundColor: '#fff',
        borderRadius: 12,
        padding: 40,
        alignItems: 'center',
    },
    instructionEmoji: {
        fontSize: 48,
        marginBottom: 12,
    },
    instructionText: {
        fontSize: 16,
        color: '#666',
        textAlign: 'center',
        marginBottom: 8,
    },
    instructionHint: {
        fontSize: 14,
        color: '#ff9800',
        fontStyle: 'italic',
    },
    massUpdateSection: {
        marginTop: 24,
        paddingTop: 20,
        borderTopWidth: 2,
        borderTopColor: '#e0e0e0',
    },
    massUpdateCard: {
        backgroundColor: '#fff',
        padding: 20,
        borderRadius: 12,
        borderWidth: 2,
        borderColor: '#9c27b0',
        marginTop: 12,
    },
    massInput: {
        height: 50,
        backgroundColor: '#f5f5f5',
        borderRadius: 8,
        paddingHorizontal: 16,
        fontSize: 18,
        marginBottom: 16,
        textAlign: 'center',
    },
    massButtonRow: {
        flexDirection: 'row',
        gap: 12,
    },
    massButton: {
        flex: 1,
        borderRadius: 8,
        paddingVertical: 14,
        alignItems: 'center',
    },
    randomButton: {
        backgroundColor: '#9c27b0',
    },
    valueButton: {
        backgroundColor: '#2196f3',
    },
    resultBox: {
        backgroundColor: '#e8f5e9',
        padding: 16,
        borderRadius: 8,
        marginTop: 16,
        borderWidth: 1,
        borderColor: '#4caf50',
    },
    resultTitle: {
        fontSize: 16,
        fontWeight: '700',
        color: '#2e7d32',
        marginBottom: 8,
    },
    resultText: {
        fontSize: 14,
        color: '#333',
        marginBottom: 4,
    },
    resultHighlight: {
        fontSize: 18,
        fontWeight: '700',
        color: '#4caf50',
        marginTop: 8,
    },
});
