import React, { useState } from 'react';
import { SafeAreaView, View, TouchableOpacity, Text, StyleSheet, StatusBar, useWindowDimensions } from 'react-native';
import { LeaderboardScreen } from './src/screens/LeaderboardScreen';
import { UserSearchScreen } from './src/screens/UserSearchScreen';
import { AdminScreen } from './src/screens/AdminScreen';

/**
 * Main App Component
 * Simple tab navigation between Leaderboard, User Search, and Admin screens
 * Responsive: max-width 800px on web, full-width on mobile
 */
export default function App() {
  const [activeTab, setActiveTab] = useState<'leaderboard' | 'search' | 'admin'>('leaderboard');
  const { width } = useWindowDimensions();
  const isWideScreen = width > 1200;

  const renderScreen = () => {
    switch (activeTab) {
      case 'leaderboard':
        return <LeaderboardScreen />;
      case 'search':
        return <UserSearchScreen />;
      case 'admin':
        return <AdminScreen />;
    }
  };

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar barStyle="dark-content" backgroundColor="#fff" />
      <View style={[styles.container, isWideScreen && styles.containerWide]}>
        {/* Tab Content */}
        <View style={styles.content}>
          {renderScreen()}
        </View>

        {/* Bottom Tab Navigation */}
        <View style={styles.tabBar}>
          <TouchableOpacity
            style={[styles.tab, activeTab === 'leaderboard' && styles.activeTab]}
            onPress={() => setActiveTab('leaderboard')}
          >
            <Text style={styles.tabIcon}>🏆</Text>
            <Text
              style={[styles.tabLabel, activeTab === 'leaderboard' && styles.activeTabLabel]}
            >
              Leaderboard
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.tab, activeTab === 'search' && styles.activeTab]}
            onPress={() => setActiveTab('search')}
          >
            <Text style={styles.tabIcon}>🔍</Text>
            <Text style={[styles.tabLabel, activeTab === 'search' && styles.activeTabLabel]}>
              Search
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.tab, activeTab === 'admin' && styles.activeTab]}
            onPress={() => setActiveTab('admin')}
          >
            <Text style={styles.tabIcon}>⚙️</Text>
            <Text style={[styles.tabLabel, activeTab === 'admin' && styles.activeTabLabel]}>
              Admin
            </Text>
          </TouchableOpacity>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  container: {
    flex: 1,
    backgroundColor: '#fff',
    width: '100%',
  },
  containerWide: {
    maxWidth: 1200,
    alignSelf: 'center',
    borderLeftWidth: 1,
    borderRightWidth: 1,
    borderColor: '#e0e0e0',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
  },
  content: {
    flex: 1,
  },
  tabBar: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    borderTopWidth: 1,
    borderTopColor: '#e0e0e0',
    paddingBottom: 0,
  },
  tab: {
    flex: 1,
    alignItems: 'center',
    paddingVertical: 12,
  },
  activeTab: {
    backgroundColor: '#f0f7ff',
  },
  tabIcon: {
    fontSize: 24,
    marginBottom: 4,
  },
  tabLabel: {
    fontSize: 12,
    color: '#666',
    fontWeight: '500',
  },
  activeTabLabel: {
    color: '#4a90e2',
    fontWeight: '700',
  },
});
