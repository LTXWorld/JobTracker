import { combineReducers, configureStore } from '@reduxjs/toolkit';
import { persistReducer, persistStore } from 'redux-persist';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useDispatch, useSelector, TypedUseSelectorHook } from 'react-redux';

// Import slices (placeholder for now)
import { createAuthSlice } from './slices/authSlice';
import { createApplicationSlice } from './slices/applicationSlice';
import { createUiSlice } from './slices/uiSlice';
import { createRemindersSlice } from './slices/remindersSlice';

// Import RTK Query API (placeholder for now)
import { baseApi } from './api/baseApi';

// Create slices
const authSlice = createAuthSlice();
const applicationSlice = createApplicationSlice();
const uiSlice = createUiSlice();
const remindersSlice = createRemindersSlice();

// Persist config
const persistConfig = {
  key: 'root',
  storage: AsyncStorage,
  whitelist: ['auth', 'ui'], // Only persist auth and ui state
  blacklist: ['api'], // Don't persist RTK Query cache
};

const rootReducer = combineReducers({
  auth: authSlice.reducer,
  applications: applicationSlice.reducer,
  ui: uiSlice.reducer,
  reminders: remindersSlice.reducer,
  [baseApi.reducerPath]: baseApi.reducer,
});

const persistedReducer = persistReducer(persistConfig, rootReducer);

export const store = configureStore({
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: [
          'persist/PERSIST',
          'persist/REHYDRATE',
          'persist/PAUSE',
          'persist/PURGE',
          'persist/REGISTER',
          'persist/FLUSH',
        ],
      },
    }).concat(baseApi.middleware),
  devTools: __DEV__,
});

export const persistor = persistStore(store);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// Typed hooks
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;

// Export action creators and async thunks
export const {
  setLoading,
  setError,
  clearError,
  setUser,
  setTokens,
  setBiometricAvailable,
  setBiometricEnabled,
  clearAuth
} = authSlice.actions;

// Export async thunks from auth
export { loginUser, registerUser, logoutUser, loadStoredAuth, forgotPassword } from './slices/authSlice';

export const {
  setLoading: setApplicationsLoading,
  setError: setApplicationsError,
  clearError: clearApplicationsError,
  setCurrentApplication,
  clearSearchQuery,
  optimisticUpdateStatus
} = applicationSlice.actions;

// Export async thunks from applications
export {
  loadApplications,
  createApplication,
  updateApplication,
  deleteApplication,
  loadKanbanData,
  loadStatistics,
  searchApplications
} from './slices/applicationSlice';

// Export reminders actions and async thunks
export const {
  setLoading: setRemindersLoading,
  setError: setRemindersError,
  clearError: clearRemindersError,
  addReminder,
  removeReminder,
  updateReminder,
  setNotificationPermission,
  updateSettings: updateReminderSettings,
  resetReminders
} = remindersSlice.actions;

export {
  initializeNotifications,
  requestNotificationPermission,
  scheduleReminder,
  cancelReminder,
  updateNotificationSettings,
  loadScheduledNotifications,
  cleanupExpiredNotifications
} from './slices/remindersSlice';

export const {
  setTheme,
  setSystemTheme,
  setLoading: setUiLoading,
  setError: setUiError,
  clearError: clearUiError,
  setNetworkStatus,
  setActiveScreen,
  showModal,
  hideModal,
  showToast,
  hideToast,
  resetUI
} = uiSlice.actions;
