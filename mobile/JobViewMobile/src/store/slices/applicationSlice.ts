import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { mockDataService } from '@services/mockDataService';
import { JobApplication, ApplicationStatus, ApplicationPriority } from '@types';

interface ApplicationState {
  applications: JobApplication[];
  currentApplication: JobApplication | null;
  kanbanData: Record<ApplicationStatus, JobApplication[]>;
  statistics: {
    total: number;
    byStatus: Record<ApplicationStatus, number>;
    byPriority: Record<ApplicationPriority, number>;
    byWorkType: Record<string, number>;
    recentActivity: number;
  } | null;
  isLoading: boolean;
  error: string | null;
  searchQuery: string;
  filteredApplications: JobApplication[];
}

const initialState: ApplicationState = {
  applications: [],
  currentApplication: null,
  kanbanData: {
    '已保存': [],
    '已投递': [],
    '简历筛选中': [],
    '笔试中': [],
    '一面中': [],
    '二面中': [],
    '三面中': [],
    'HR面中': [],
    '等待offer': [],
    '已收到offer': [],
    '已拒绝offer': [],
    '简历挂': [],
    '笔试挂': [],
    '一面挂': [],
    '二面挂': [],
    '三面挂': [],
    'HR面挂': [],
    '被拒绝': [],
    '已入职': [],
    '已放弃': [],
  },
  statistics: null,
  isLoading: false,
  error: null,
  searchQuery: '',
  filteredApplications: [],
};

// Async thunks
export const loadApplications = createAsyncThunk(
  'applications/loadApplications',
  async (userId: string, { rejectWithValue }) => {
    try {
      const applications = await mockDataService.getApplications(userId);
      return applications;
    } catch (error: any) {
      return rejectWithValue(error.message || '加载申请记录失败');
    }
  }
);

export const createApplication = createAsyncThunk(
  'applications/createApplication',
  async (applicationData: Omit<JobApplication, 'id' | 'createdAt' | 'updatedAt'>, { rejectWithValue }) => {
    try {
      const newApplication = await mockDataService.createApplication(applicationData);
      return newApplication;
    } catch (error: any) {
      return rejectWithValue(error.message || '创建申请记录失败');
    }
  }
);

export const updateApplication = createAsyncThunk(
  'applications/updateApplication',
  async ({ id, updates }: { id: string; updates: Partial<JobApplication> }, { rejectWithValue }) => {
    try {
      const updatedApplication = await mockDataService.updateApplication(id, updates);
      if (!updatedApplication) {
        return rejectWithValue('申请记录不存在');
      }
      return updatedApplication;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新申请记录失败');
    }
  }
);

export const deleteApplication = createAsyncThunk(
  'applications/deleteApplication',
  async (id: string, { rejectWithValue }) => {
    try {
      const success = await mockDataService.deleteApplication(id);
      if (!success) {
        return rejectWithValue('申请记录不存在');
      }
      return id;
    } catch (error: any) {
      return rejectWithValue(error.message || '删除申请记录失败');
    }
  }
);

export const loadKanbanData = createAsyncThunk(
  'applications/loadKanbanData',
  async (userId: string, { rejectWithValue }) => {
    try {
      const kanbanData = await mockDataService.getApplicationsByStatus(userId);
      return kanbanData;
    } catch (error: any) {
      return rejectWithValue(error.message || '加载看板数据失败');
    }
  }
);

export const loadStatistics = createAsyncThunk(
  'applications/loadStatistics',
  async (userId: string, { rejectWithValue }) => {
    try {
      const statistics = await mockDataService.getStatistics(userId);
      return statistics;
    } catch (error: any) {
      return rejectWithValue(error.message || '加载统计数据失败');
    }
  }
);

export const searchApplications = createAsyncThunk(
  'applications/searchApplications',
  async ({ userId, query }: { userId: string; query: string }, { rejectWithValue }) => {
    try {
      const results = await mockDataService.searchApplications(userId, query);
      return { query, results };
    } catch (error: any) {
      return rejectWithValue(error.message || '搜索失败');
    }
  }
);

export const createApplicationSlice = () => createSlice({
  name: 'applications',
  initialState,
  reducers: {
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    clearError: (state) => {
      state.error = null;
    },
    setCurrentApplication: (state, action: PayloadAction<JobApplication | null>) => {
      state.currentApplication = action.payload;
    },
    clearSearchQuery: (state) => {
      state.searchQuery = '';
      state.filteredApplications = [];
    },
    // Optimistic updates for real-time UI
    optimisticUpdateStatus: (state, action: PayloadAction<{ id: string; status: ApplicationStatus }>) => {
      const { id, status } = action.payload;
      const application = state.applications.find(app => app.id === id);
      if (application) {
        application.status = status;
        application.updatedAt = new Date();
      }

      // Update kanban data
      Object.keys(state.kanbanData).forEach(statusKey => {
        state.kanbanData[statusKey as ApplicationStatus] = state.kanbanData[statusKey as ApplicationStatus]
          .filter(app => app.id !== id);
      });

      if (application) {
        state.kanbanData[status].push(application);
      }
    },
  },
  extraReducers: (builder) => {
    // Load applications
    builder
      .addCase(loadApplications.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loadApplications.fulfilled, (state, action) => {
        state.isLoading = false;
        state.applications = action.payload;
        state.error = null;
      })
      .addCase(loadApplications.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Create application
    builder
      .addCase(createApplication.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(createApplication.fulfilled, (state, action) => {
        state.isLoading = false;
        state.applications.push(action.payload);
        state.kanbanData[action.payload.status].push(action.payload);
        state.error = null;
      })
      .addCase(createApplication.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Update application
    builder
      .addCase(updateApplication.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(updateApplication.fulfilled, (state, action) => {
        state.isLoading = false;
        const updatedApp = action.payload;
        const index = state.applications.findIndex(app => app.id === updatedApp.id);
        if (index !== -1) {
          state.applications[index] = updatedApp;
        }

        // Update kanban data
        Object.keys(state.kanbanData).forEach(statusKey => {
          state.kanbanData[statusKey as ApplicationStatus] = state.kanbanData[statusKey as ApplicationStatus]
            .filter(app => app.id !== updatedApp.id);
        });
        state.kanbanData[updatedApp.status].push(updatedApp);

        state.error = null;
      })
      .addCase(updateApplication.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Delete application
    builder
      .addCase(deleteApplication.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(deleteApplication.fulfilled, (state, action) => {
        state.isLoading = false;
        const deletedId = action.payload;
        state.applications = state.applications.filter(app => app.id !== deletedId);

        // Update kanban data
        Object.keys(state.kanbanData).forEach(statusKey => {
          state.kanbanData[statusKey as ApplicationStatus] = state.kanbanData[statusKey as ApplicationStatus]
            .filter(app => app.id !== deletedId);
        });

        state.error = null;
      })
      .addCase(deleteApplication.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Load kanban data
    builder
      .addCase(loadKanbanData.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loadKanbanData.fulfilled, (state, action) => {
        state.isLoading = false;
        state.kanbanData = action.payload;
        state.error = null;
      })
      .addCase(loadKanbanData.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Load statistics
    builder
      .addCase(loadStatistics.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loadStatistics.fulfilled, (state, action) => {
        state.isLoading = false;
        state.statistics = action.payload;
        state.error = null;
      })
      .addCase(loadStatistics.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Search applications
    builder
      .addCase(searchApplications.fulfilled, (state, action) => {
        state.searchQuery = action.payload.query;
        state.filteredApplications = action.payload.results;
      });
  },
});