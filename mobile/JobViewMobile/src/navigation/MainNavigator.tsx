import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Text } from 'react-native';

// Import navigation types
import { MainTabParamList } from '@types';

// Import screens and navigators
import DashboardScreen from '@screens/main/DashboardScreen';
import ApplicationsNavigator from '@navigation/ApplicationsNavigator';
import KanbanScreen from '@screens/main/KanbanScreen';
import StatisticsScreen from '@screens/main/StatisticsScreen';
import ProfileNavigator from '@navigation/ProfileNavigator';

// Simple icon placeholders
const TabIcon = ({ label, focused }: { label: string; focused: boolean }) => (
  <Text style={{
    color: focused ? '#6750A4' : '#79747E',
    fontSize: 12,
    fontWeight: focused ? '600' : '400'
  }}>
    {label.charAt(0)}
  </Text>
);

const Tab = createBottomTabNavigator<MainTabParamList>();

const MainNavigator: React.FC = () => {
  return (
    <Tab.Navigator
      initialRouteName="Dashboard"
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: '#6750A4',
        tabBarInactiveTintColor: '#79747E',
        tabBarStyle: {
          backgroundColor: '#FEF7FF',
          borderTopColor: '#CAC4D0',
          paddingBottom: 8,
          paddingTop: 8,
          height: 80,
        },
        tabBarLabelStyle: {
          fontSize: 12,
          fontWeight: '500',
        },
      }}
    >
      <Tab.Screen
        name="Dashboard"
        component={DashboardScreen}
        options={{
          title: '首页',
          tabBarIcon: ({ focused }) => <TabIcon label="首页" focused={focused} />,
        }}
      />
      <Tab.Screen
        name="Applications"
        component={ApplicationsNavigator}
        options={{
          title: '投递记录',
          tabBarIcon: ({ focused }) => <TabIcon label="记录" focused={focused} />,
        }}
      />
      <Tab.Screen
        name="Kanban"
        component={KanbanScreen}
        options={{
          title: '看板',
          tabBarIcon: ({ focused }) => <TabIcon label="看板" focused={focused} />,
        }}
      />
      <Tab.Screen
        name="Statistics"
        component={StatisticsScreen}
        options={{
          title: '统计',
          tabBarIcon: ({ focused }) => <TabIcon label="统计" focused={focused} />,
        }}
      />
      <Tab.Screen
        name="Profile"
        component={ProfileNavigator}
        options={{
          title: '我的',
          tabBarIcon: ({ focused }) => <TabIcon label="我的" focused={focused} />,
        }}
      />
    </Tab.Navigator>
  );
};

export default MainNavigator;