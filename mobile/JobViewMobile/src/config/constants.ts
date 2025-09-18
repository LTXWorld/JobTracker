// API Configuration
export const API_CONFIG = {
  BASE_URL: __DEV__ ? 'http://10.0.2.2:8080' : 'https://api.jobview.app',
  TIMEOUT: 10000,
  VERSION: 'v1',
} as const;

// Database Configuration
export const DB_CONFIG = {
  DB_NAME: 'JobViewDB',
  SCHEMA_VERSION: 1,
  ADAPTER_NAME: 'sqlite',
} as const;

// Application Configuration
export const APP_CONFIG = {
  APP_NAME: 'JobView Mobile',
  VERSION: '1.0.0',
  BUNDLE_ID: 'com.jobview.mobile',
  MIN_ANDROID_VERSION: 21,
  TARGET_ANDROID_VERSION: 34,
} as const;

// Storage Keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: '@jobview/auth_token',
  REFRESH_TOKEN: '@jobview/refresh_token',
  USER_INFO: '@jobview/user_info',
  THEME_MODE: '@jobview/theme_mode',
  FIRST_LAUNCH: '@jobview/first_launch',
  BIOMETRIC_ENABLED: '@jobview/biometric_enabled',
  NOTIFICATION_ENABLED: '@jobview/notification_enabled',
  SYNC_ENABLED: '@jobview/sync_enabled',
  LAST_SYNC: '@jobview/last_sync',
} as const;

// Feature Flags
export const FEATURE_FLAGS = {
  BIOMETRIC_AUTH: true,
  PUSH_NOTIFICATIONS: true,
  OFFLINE_MODE: true,
  DARK_MODE: true,
  ANALYTICS: false,
} as const;

// Pagination
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
} as const;

// Date Formats
export const DATE_FORMATS = {
  DISPLAY: 'yyyy年MM月dd日',
  API: 'yyyy-MM-dd',
  DATETIME: 'yyyy-MM-dd HH:mm:ss',
  TIME: 'HH:mm',
} as const;