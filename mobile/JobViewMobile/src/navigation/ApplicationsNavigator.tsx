import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { TouchableOpacity, Text } from 'react-native';
import { useNavigation } from '@react-navigation/native';

// Import navigation types
import { ApplicationsStackParamList } from '@types';

// Import screens
import ApplicationsScreen from '@screens/main/ApplicationsScreen';
import AddApplicationScreen from '@screens/main/AddApplicationScreen';
import ApplicationDetailsScreen from '@screens/main/ApplicationDetailsScreen';
import EditApplicationScreen from '@screens/main/EditApplicationScreen';

const Stack = createNativeStackNavigator<ApplicationsStackParamList>();

const ApplicationsNavigator: React.FC = () => {
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
        name="ApplicationsList"
        component={ApplicationsScreen}
        options={({ navigation }) => ({
          title: '投递记录',
          headerRight: () => (
            <TouchableOpacity
              onPress={() => navigation.navigate('AddApplication')}
              style={{
                backgroundColor: '#6750A4',
                paddingHorizontal: 12,
                paddingVertical: 6,
                borderRadius: 16,
              }}
            >
              <Text style={{
                color: '#FFFFFF',
                fontSize: 14,
                fontWeight: '600',
              }}>
                + 新增
              </Text>
            </TouchableOpacity>
          ),
        })}
      />
      <Stack.Screen
        name="AddApplication"
        component={AddApplicationScreen}
        options={{
          title: '新增投递',
          presentation: 'modal',
          headerShown: false, // We handle the header in the component
        }}
      />
      <Stack.Screen
        name="ApplicationDetails"
        component={ApplicationDetailsScreen}
        options={{
          title: '投递详情',
          headerShown: false, // We handle the header in the component
        }}
      />
      <Stack.Screen
        name="EditApplication"
        component={EditApplicationScreen}
        options={{
          title: '编辑记录',
          headerShown: false, // We handle the header in the component
        }}
      />
    </Stack.Navigator>
  );
};

export default ApplicationsNavigator;