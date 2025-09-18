import { apiClient } from './authService';
import { JobApplication, ApplicationStatus, ApplicationPriority, ApplicationNote, ApplicationReminder } from '../types';
import { PAGINATION } from '../config';

// API Response Types
interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
  pagination?: {
    currentPage: number;
    totalPages: number;
    totalItems: number;
    pageSize: number;
  };
}

interface CreateApplicationRequest {
  companyName: string;
  position: string;
  location: string;
  salaryRange?: string;
  workType: 'remote' | 'onsite' | 'hybrid';
  description?: string;
  requirements?: string;
  priority: ApplicationPriority;
  source: string;
  jobUrl?: string;
  companyUrl?: string;
  department?: string;
  experienceLevel?: string;
  contactPerson?: string;
  contactEmail?: string;
  contactPhone?: string;
  applicationDate: string; // ISO date string
  notes?: string;
  tags?: string[];
}

interface UpdateApplicationRequest extends Partial<CreateApplicationRequest> {
  id?: string;
  status?: ApplicationStatus;
}

interface GetApplicationsParams {
  page?: number;
  pageSize?: number;
  status?: ApplicationStatus;
  search?: string;
  sortBy?: 'applicationDate' | 'companyName' | 'position' | 'status' | 'createdAt';
  sortOrder?: 'asc' | 'desc';
  priority?: ApplicationPriority;
  workType?: 'remote' | 'onsite' | 'hybrid';
  dateFrom?: string;
  dateTo?: string;
}

// Applications API
export const applicationsAPI = {
  // Get all applications with pagination and filters
  getApplications: async (params: GetApplicationsParams = {}): Promise<ApiResponse<JobApplication[]>> => {
    try {
      const queryParams = new URLSearchParams();

      // Add pagination params
      queryParams.append('page', String(params.page || 1));
      queryParams.append('pageSize', String(params.pageSize || PAGINATION.DEFAULT_PAGE_SIZE));

      // Add filter params
      if (params.status) queryParams.append('status', params.status);
      if (params.search) queryParams.append('search', params.search);
      if (params.sortBy) queryParams.append('sortBy', params.sortBy);
      if (params.sortOrder) queryParams.append('sortOrder', params.sortOrder);
      if (params.priority) queryParams.append('priority', params.priority);
      if (params.workType) queryParams.append('workType', params.workType);
      if (params.dateFrom) queryParams.append('dateFrom', params.dateFrom);
      if (params.dateTo) queryParams.append('dateTo', params.dateTo);

      const response = await apiClient.get(`/api/v1/applications?${queryParams.toString()}`);
      return response.data;
    } catch (error: any) {
      console.error('Get Applications API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '获取申请记录失败',
        data: [],
      };
    }
  },

  // Get single application by ID
  getApplication: async (id: string): Promise<ApiResponse<JobApplication>> => {
    try {
      const response = await apiClient.get(`/api/v1/applications/${id}`);
      return response.data;
    } catch (error: any) {
      console.error('Get Application API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '获取申请详情失败',
      };
    }
  },

  // Create new application
  createApplication: async (applicationData: CreateApplicationRequest): Promise<ApiResponse<JobApplication>> => {
    try {
      const response = await apiClient.post('/api/v1/applications', applicationData);
      return response.data;
    } catch (error: any) {
      console.error('Create Application API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '创建申请记录失败',
      };
    }
  },

  // Update application
  updateApplication: async (id: string, updates: UpdateApplicationRequest): Promise<ApiResponse<JobApplication>> => {
    try {
      const response = await apiClient.patch(`/api/v1/applications/${id}`, updates);
      return response.data;
    } catch (error: any) {
      console.error('Update Application API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '更新申请记录失败',
      };
    }
  },

  // Delete application
  deleteApplication: async (id: string): Promise<ApiResponse<void>> => {
    try {
      const response = await apiClient.delete(`/api/v1/applications/${id}`);
      return response.data;
    } catch (error: any) {
      console.error('Delete Application API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '删除申请记录失败',
      };
    }
  },

  // Batch operations
  batchCreateApplications: async (applications: CreateApplicationRequest[]): Promise<ApiResponse<JobApplication[]>> => {
    try {
      const response = await apiClient.post('/api/v1/applications/batch', { applications });
      return response.data;
    } catch (error: any) {
      console.error('Batch Create Applications API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '批量创建申请记录失败',
      };
    }
  },

  batchUpdateApplications: async (updates: Array<{ id: string; updates: UpdateApplicationRequest }>): Promise<ApiResponse<JobApplication[]>> => {
    try {
      const response = await apiClient.patch('/api/v1/applications/batch', { updates });
      return response.data;
    } catch (error: any) {
      console.error('Batch Update Applications API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '批量更新申请记录失败',
      };
    }
  },

  batchDeleteApplications: async (ids: string[]): Promise<ApiResponse<void>> => {
    try {
      const response = await apiClient.delete('/api/v1/applications/batch', { data: { ids } });
      return response.data;
    } catch (error: any) {
      console.error('Batch Delete Applications API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '批量删除申请记录失败',
      };
    }
  },

  // Search applications
  searchApplications: async (query: string, filters?: Partial<GetApplicationsParams>): Promise<ApiResponse<JobApplication[]>> => {
    try {
      const params: GetApplicationsParams = {
        search: query,
        ...filters,
      };
      return await applicationsAPI.getApplications(params);
    } catch (error: any) {
      console.error('Search Applications API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '搜索申请记录失败',
        data: [],
      };
    }
  },

  // Get applications statistics
  getStatistics: async (params?: { dateFrom?: string; dateTo?: string }): Promise<ApiResponse<{
    totalApplications: number;
    statusCounts: Record<ApplicationStatus, number>;
    recentActivity: Array<{
      date: string;
      count: number;
    }>;
    priorityCounts: Record<ApplicationPriority, number>;
    workTypeCounts: Record<string, number>;
    topCompanies: Array<{
      company: string;
      count: number;
    }>;
    conversionRates: {
      interviewRate: number;
      offerRate: number;
      acceptanceRate: number;
    };
  }>> => {
    try {
      const queryParams = new URLSearchParams();
      if (params?.dateFrom) queryParams.append('dateFrom', params.dateFrom);
      if (params?.dateTo) queryParams.append('dateTo', params.dateTo);

      const response = await apiClient.get(`/api/v1/applications/statistics?${queryParams.toString()}`);
      return response.data;
    } catch (error: any) {
      console.error('Get Statistics API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '获取统计数据失败',
      };
    }
  },

  // Export applications
  exportApplications: async (format: 'json' | 'csv' | 'excel', filters?: GetApplicationsParams): Promise<ApiResponse<{
    downloadUrl: string;
    fileName: string;
  }>> => {
    try {
      const queryParams = new URLSearchParams();
      queryParams.append('format', format);

      if (filters) {
        Object.entries(filters).forEach(([key, value]) => {
          if (value !== undefined) {
            queryParams.append(key, String(value));
          }
        });
      }

      const response = await apiClient.post(`/api/v1/applications/export?${queryParams.toString()}`);
      return response.data;
    } catch (error: any) {
      console.error('Export Applications API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '导出申请记录失败',
      };
    }
  },
};

// Notes API
export const notesAPI = {
  // Get notes for an application
  getNotes: async (applicationId: string): Promise<ApiResponse<ApplicationNote[]>> => {
    try {
      const response = await apiClient.get(`/api/v1/applications/${applicationId}/notes`);
      return response.data;
    } catch (error: any) {
      console.error('Get Notes API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '获取备注失败',
        data: [],
      };
    }
  },

  // Create note
  createNote: async (applicationId: string, noteData: {
    title: string;
    content: string;
    type: 'general' | 'interview' | 'offer' | 'followup';
  }): Promise<ApiResponse<ApplicationNote>> => {
    try {
      const response = await apiClient.post(`/api/v1/applications/${applicationId}/notes`, noteData);
      return response.data;
    } catch (error: any) {
      console.error('Create Note API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '创建备注失败',
      };
    }
  },

  // Update note
  updateNote: async (noteId: string, updates: {
    title?: string;
    content?: string;
    type?: 'general' | 'interview' | 'offer' | 'followup';
  }): Promise<ApiResponse<ApplicationNote>> => {
    try {
      const response = await apiClient.patch(`/api/v1/notes/${noteId}`, updates);
      return response.data;
    } catch (error: any) {
      console.error('Update Note API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '更新备注失败',
      };
    }
  },

  // Delete note
  deleteNote: async (noteId: string): Promise<ApiResponse<void>> => {
    try {
      const response = await apiClient.delete(`/api/v1/notes/${noteId}`);
      return response.data;
    } catch (error: any) {
      console.error('Delete Note API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '删除备注失败',
      };
    }
  },
};

// Reminders API
export const remindersAPI = {
  // Get reminders for an application
  getReminders: async (applicationId: string): Promise<ApiResponse<ApplicationReminder[]>> => {
    try {
      const response = await apiClient.get(`/api/v1/applications/${applicationId}/reminders`);
      return response.data;
    } catch (error: any) {
      console.error('Get Reminders API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '获取提醒失败',
        data: [],
      };
    }
  },

  // Get all user reminders
  getAllReminders: async (): Promise<ApiResponse<ApplicationReminder[]>> => {
    try {
      const response = await apiClient.get('/api/v1/reminders');
      return response.data;
    } catch (error: any) {
      console.error('Get All Reminders API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '获取所有提醒失败',
        data: [],
      };
    }
  },

  // Create reminder
  createReminder: async (applicationId: string, reminderData: {
    title: string;
    description?: string;
    reminderDate: string; // ISO date string
    notificationType: 'local' | 'push';
  }): Promise<ApiResponse<ApplicationReminder>> => {
    try {
      const response = await apiClient.post(`/api/v1/applications/${applicationId}/reminders`, reminderData);
      return response.data;
    } catch (error: any) {
      console.error('Create Reminder API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '创建提醒失败',
      };
    }
  },

  // Update reminder
  updateReminder: async (reminderId: string, updates: {
    title?: string;
    description?: string;
    reminderDate?: string;
    isCompleted?: boolean;
    notificationType?: 'local' | 'push';
  }): Promise<ApiResponse<ApplicationReminder>> => {
    try {
      const response = await apiClient.patch(`/api/v1/reminders/${reminderId}`, updates);
      return response.data;
    } catch (error: any) {
      console.error('Update Reminder API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '更新提醒失败',
      };
    }
  },

  // Delete reminder
  deleteReminder: async (reminderId: string): Promise<ApiResponse<void>> => {
    try {
      const response = await apiClient.delete(`/api/v1/reminders/${reminderId}`);
      return response.data;
    } catch (error: any) {
      console.error('Delete Reminder API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '删除提醒失败',
      };
    }
  },

  // Mark reminder as completed
  completeReminder: async (reminderId: string): Promise<ApiResponse<ApplicationReminder>> => {
    try {
      const response = await apiClient.patch(`/api/v1/reminders/${reminderId}/complete`);
      return response.data;
    } catch (error: any) {
      console.error('Complete Reminder API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '完成提醒失败',
      };
    }
  },
};

export default {
  applications: applicationsAPI,
  notes: notesAPI,
  reminders: remindersAPI,
};