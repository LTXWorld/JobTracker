// API Services
export { authAPI, apiClient } from './authService';
export { applicationsAPI, notesAPI, remindersAPI } from './apiService';

// Mock Services (for development and testing)
export { mockAuthAPI } from './mockAuthService';
export { mockDataService } from './mockDataService';

// Other Services
export { NotificationService } from './notificationService';
export { DataImportExportService } from './dataImportExportService';

// Service configuration
export const serviceConfig = {
  useRealAPI: __DEV__ ? false : true,
  apiTimeout: 10000,
  maxRetries: 3,
  retryDelay: 1000,
};

// Service factory to get the appropriate service based on configuration
export const getAuthService = (useRealAPI: boolean = serviceConfig.useRealAPI) => {
  return useRealAPI ? authAPI : mockAuthAPI;
};

export const getApplicationsService = (useRealAPI: boolean = serviceConfig.useRealAPI) => {
  return useRealAPI ? applicationsAPI : mockDataService;
};
