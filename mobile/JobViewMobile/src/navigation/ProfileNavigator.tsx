import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';

// Import navigation types
import { ProfileStackParamList } from '@types';

// Import screens
import ProfileScreen from '@screens/main/ProfileScreen';
import DataManagementScreen from '@screens/main/DataManagementScreen';
import RemindersScreen from '@screens/main/RemindersScreen';

const Stack = createNativeStackNavigator<ProfileStackParamList>();

const ProfileNavigator: React.FC = () => {
  return (
    <Stack.Navigator
      screenOptions={{
        headerStyle: {
          backgroundColor: '#FEF7FF',
        },
        headerTintColor: '#1C1B1F',
        headerTitleStyle: {
          fontWeight: 'bold',
          fontSize: 18,
        },
      }}
    >
      <Stack.Screen
        name="ProfileMain"
        component={ProfileScreen}
        options={{
          title: '个人中心',
          headerShown: false, // We handle the header in the component
        }}
      />
      <Stack.Screen
        name="DataManagement"
        component={DataManagementScreen}
        options={{
          title: '数据管理',
          headerShown: false, // We handle the header in the component
        }}
      />
      <Stack.Screen
        name="Reminders"
        component={RemindersScreen}
        options={{
          title: '提醒中心',
          headerShown: false, // We handle the header in the component
        }}
      />
    </Stack.Navigator>
  );
};

export default ProfileNavigator;