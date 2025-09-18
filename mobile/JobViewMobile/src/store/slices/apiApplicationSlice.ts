import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { applicationsAPI, notesAPI, remindersAPI } from '../services/apiService';
import { JobApplication, ApplicationStatus, ApplicationPriority, ApplicationNote, ApplicationReminder } from '../types';

interface ApplicationState {
  applications: JobApplication[];
  currentApplication: JobApplication | null;
  kanbanData: Record<ApplicationStatus, JobApplication[]>;
  statistics: {
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
  } | null;
  notes: ApplicationNote[];
  reminders: ApplicationReminder[];
  isLoading: boolean;
  error: string | null;
  searchQuery: string;
  pagination: {
    currentPage: number;
    totalPages: number;
    totalItems: number;
    pageSize: number;
  } | null;
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
  notes: [],
  reminders: [],
  isLoading: false,
  error: null,
  searchQuery: '',
  pagination: null,
};

// Applications Async Thunks
export const loadApplications = createAsyncThunk(
  'applications/loadApplications',
  async (params: {
    page?: number;
    pageSize?: number;
    status?: ApplicationStatus;
    search?: string;
    sortBy?: string;
    sortOrder?: 'asc' | 'desc';
  } = {}, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.getApplications(params);
      if (!response.success) {
        return rejectWithValue(response.message || '加载申请记录失败');
      }
      return {
        applications: response.data,
        pagination: response.pagination,
      };
    } catch (error: any) {
      return rejectWithValue(error.message || '加载申请记录失败');
    }
  }
);

export const loadApplication = createAsyncThunk(
  'applications/loadApplication',
  async (id: string, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.getApplication(id);
      if (!response.success) {
        return rejectWithValue(response.message || '加载申请详情失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '加载申请详情失败');
    }
  }
);

export const createApplication = createAsyncThunk(
  'applications/createApplication',
  async (applicationData: {
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
    applicationDate: string;
    notes?: string;
    tags?: string[];
  }, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.createApplication(applicationData);
      if (!response.success) {
        return rejectWithValue(response.message || '创建申请记录失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '创建申请记录失败');
    }
  }
);

export const updateApplication = createAsyncThunk(
  'applications/updateApplication',
  async ({ id, updates }: { id: string; updates: Partial<JobApplication> }, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.updateApplication(id, updates);
      if (!response.success) {
        return rejectWithValue(response.message || '更新申请记录失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新申请记录失败');
    }
  }
);

export const deleteApplication = createAsyncThunk(
  'applications/deleteApplication',
  async (id: string, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.deleteApplication(id);
      if (!response.success) {
        return rejectWithValue(response.message || '删除申请记录失败');
      }
      return id;
    } catch (error: any) {
      return rejectWithValue(error.message || '删除申请记录失败');
    }
  }
);

export const batchCreateApplications = createAsyncThunk(
  'applications/batchCreateApplications',
  async (applications: any[], { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.batchCreateApplications(applications);
      if (!response.success) {
        return rejectWithValue(response.message || '批量创建申请记录失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '批量创建申请记录失败');
    }
  }
);

export const batchUpdateApplications = createAsyncThunk(
  'applications/batchUpdateApplications',
  async (updates: Array<{ id: string; updates: any }>, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.batchUpdateApplications(updates);
      if (!response.success) {
        return rejectWithValue(response.message || '批量更新申请记录失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '批量更新申请记录失败');
    }
  }
);

export const batchDeleteApplications = createAsyncThunk(
  'applications/batchDeleteApplications',
  async (ids: string[], { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.batchDeleteApplications(ids);
      if (!response.success) {
        return rejectWithValue(response.message || '批量删除申请记录失败');
      }
      return ids;
    } catch (error: any) {
      return rejectWithValue(error.message || '批量删除申请记录失败');
    }
  }
);

export const searchApplications = createAsyncThunk(
  'applications/searchApplications',
  async ({ query, filters }: { query: string; filters?: any }, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.searchApplications(query, filters);
      if (!response.success) {
        return rejectWithValue(response.message || '搜索申请记录失败');
      }
      return {
        applications: response.data,
        pagination: response.pagination,
      };
    } catch (error: any) {
      return rejectWithValue(error.message || '搜索申请记录失败');
    }
  }
);

export const loadStatistics = createAsyncThunk(
  'applications/loadStatistics',
  async (params?: { dateFrom?: string; dateTo?: string }, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.getStatistics(params);
      if (!response.success) {
        return rejectWithValue(response.message || '获取统计数据失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取统计数据失败');
    }
  }
);

export const exportApplications = createAsyncThunk(
  'applications/exportApplications',
  async ({ format, filters }: { format: 'json' | 'csv' | 'excel'; filters?: any }, { rejectWithValue }) => {
    try {
      const response = await applicationsAPI.exportApplications(format, filters);
      if (!response.success) {
        return rejectWithValue(response.message || '导出申请记录失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '导出申请记录失败');
    }
  }
);

// Notes Async Thunks
export const loadNotes = createAsyncThunk(
  'applications/loadNotes',
  async (applicationId: string, { rejectWithValue }) => {
    try {
      const response = await notesAPI.getNotes(applicationId);
      if (!response.success) {
        return rejectWithValue(response.message || '获取备注失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取备注失败');
    }
  }
);

export const createNote = createAsyncThunk(
  'applications/createNote',
  async ({ applicationId, noteData }: {
    applicationId: string;
    noteData: {
      title: string;
      content: string;
      type: 'general' | 'interview' | 'offer' | 'followup';
    };
  }, { rejectWithValue }) => {
    try {
      const response = await notesAPI.createNote(applicationId, noteData);
      if (!response.success) {
        return rejectWithValue(response.message || '创建备注失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '创建备注失败');
    }
  }
);

export const updateNote = createAsyncThunk(
  'applications/updateNote',
  async ({ noteId, updates }: { noteId: string; updates: any }, { rejectWithValue }) => {
    try {
      const response = await notesAPI.updateNote(noteId, updates);
      if (!response.success) {
        return rejectWithValue(response.message || '更新备注失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新备注失败');
    }
  }
);

export const deleteNote = createAsyncThunk(
  'applications/deleteNote',
  async (noteId: string, { rejectWithValue }) => {
    try {
      const response = await notesAPI.deleteNote(noteId);
      if (!response.success) {
        return rejectWithValue(response.message || '删除备注失败');
      }
      return noteId;
    } catch (error: any) {
      return rejectWithValue(error.message || '删除备注失败');
    }
  }
);

// Reminders Async Thunks
export const loadReminders = createAsyncThunk(
  'applications/loadReminders',
  async (applicationId: string, { rejectWithValue }) => {
    try {
      const response = await remindersAPI.getReminders(applicationId);
      if (!response.success) {
        return rejectWithValue(response.message || '获取提醒失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取提醒失败');
    }
  }
);

export const loadAllReminders = createAsyncThunk(
  'applications/loadAllReminders',
  async (_, { rejectWithValue }) => {
    try {
      const response = await remindersAPI.getAllReminders();
      if (!response.success) {
        return rejectWithValue(response.message || '获取所有提醒失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取所有提醒失败');
    }
  }
);

export const createReminder = createAsyncThunk(
  'applications/createReminder',
  async ({ applicationId, reminderData }: {
    applicationId: string;
    reminderData: {
      title: string;
      description?: string;
      reminderDate: string;
      notificationType: 'local' | 'push';
    };
  }, { rejectWithValue }) => {
    try {
      const response = await remindersAPI.createReminder(applicationId, reminderData);
      if (!response.success) {
        return rejectWithValue(response.message || '创建提醒失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '创建提醒失败');
    }
  }
);

export const updateReminder = createAsyncThunk(
  'applications/updateReminder',
  async ({ reminderId, updates }: { reminderId: string; updates: any }, { rejectWithValue }) => {
    try {
      const response = await remindersAPI.updateReminder(reminderId, updates);
      if (!response.success) {
        return rejectWithValue(response.message || '更新提醒失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新提醒失败');
    }
  }
);

export const deleteReminder = createAsyncThunk(
  'applications/deleteReminder',
  async (reminderId: string, { rejectWithValue }) => {
    try {
      const response = await remindersAPI.deleteReminder(reminderId);
      if (!response.success) {
        return rejectWithValue(response.message || '删除提醒失败');
      }
      return reminderId;
    } catch (error: any) {
      return rejectWithValue(error.message || '删除提醒失败');
    }
  }
);

export const completeReminder = createAsyncThunk(
  'applications/completeReminder',
  async (reminderId: string, { rejectWithValue }) => {
    try {
      const response = await remindersAPI.completeReminder(reminderId);
      if (!response.success) {
        return rejectWithValue(response.message || '完成提醒失败');
      }
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '完成提醒失败');
    }
  }
);

// Helper function to organize applications into kanban columns
const organizeApplicationsForKanban = (applications: JobApplication[]) => {
  const kanbanData: Record<ApplicationStatus, JobApplication[]> = {
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
  };

  applications.forEach(app => {
    if (kanbanData[app.status]) {
      kanbanData[app.status].push(app);
    }
  });

  return kanbanData;
};

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
    setSearchQuery: (state, action: PayloadAction<string>) => {
      state.searchQuery = action.payload;
    },
    clearSearchQuery: (state) => {
      state.searchQuery = '';
    },
    optimisticUpdateStatus: (state, action: PayloadAction<{ id: string; status: ApplicationStatus }>) => {
      const { id, status } = action.payload;
      const applicationIndex = state.applications.findIndex(app => app.id === id);
      if (applicationIndex !== -1) {
        state.applications[applicationIndex].status = status;
        state.kanbanData = organizeApplicationsForKanban(state.applications);
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
        state.applications = action.payload.applications;
        state.pagination = action.payload.pagination;
        state.kanbanData = organizeApplicationsForKanban(action.payload.applications);
      })
      .addCase(loadApplications.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Load single application
    builder
      .addCase(loadApplication.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loadApplication.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentApplication = action.payload;
      })
      .addCase(loadApplication.rejected, (state, action) => {
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
        state.kanbanData = organizeApplicationsForKanban(state.applications);
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
        const index = state.applications.findIndex(app => app.id === action.payload.id);
        if (index !== -1) {
          state.applications[index] = action.payload;
        }
        if (state.currentApplication?.id === action.payload.id) {
          state.currentApplication = action.payload;
        }
        state.kanbanData = organizeApplicationsForKanban(state.applications);
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
        state.applications = state.applications.filter(app => app.id !== action.payload);
        if (state.currentApplication?.id === action.payload) {
          state.currentApplication = null;
        }
        state.kanbanData = organizeApplicationsForKanban(state.applications);
      })
      .addCase(deleteApplication.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Batch operations
    builder
      .addCase(batchCreateApplications.fulfilled, (state, action) => {
        state.applications.push(...action.payload);
        state.kanbanData = organizeApplicationsForKanban(state.applications);
      })
      .addCase(batchUpdateApplications.fulfilled, (state, action) => {
        action.payload.forEach(updatedApp => {
          const index = state.applications.findIndex(app => app.id === updatedApp.id);
          if (index !== -1) {
            state.applications[index] = updatedApp;
          }
        });
        state.kanbanData = organizeApplicationsForKanban(state.applications);
      })
      .addCase(batchDeleteApplications.fulfilled, (state, action) => {
        state.applications = state.applications.filter(app => !action.payload.includes(app.id));
        state.kanbanData = organizeApplicationsForKanban(state.applications);
      });

    // Search applications
    builder
      .addCase(searchApplications.fulfilled, (state, action) => {
        state.applications = action.payload.applications;
        state.pagination = action.payload.pagination;
        state.kanbanData = organizeApplicationsForKanban(action.payload.applications);
      });

    // Load statistics
    builder
      .addCase(loadStatistics.fulfilled, (state, action) => {
        state.statistics = action.payload;
      });

    // Notes
    builder
      .addCase(loadNotes.fulfilled, (state, action) => {
        state.notes = action.payload;
      })
      .addCase(createNote.fulfilled, (state, action) => {
        state.notes.push(action.payload);
      })
      .addCase(updateNote.fulfilled, (state, action) => {
        const index = state.notes.findIndex(note => note.id === action.payload.id);
        if (index !== -1) {
          state.notes[index] = action.payload;
        }
      })
      .addCase(deleteNote.fulfilled, (state, action) => {
        state.notes = state.notes.filter(note => note.id !== action.payload);
      });

    // Reminders
    builder
      .addCase(loadReminders.fulfilled, (state, action) => {
        state.reminders = action.payload;
      })
      .addCase(loadAllReminders.fulfilled, (state, action) => {
        state.reminders = action.payload;
      })
      .addCase(createReminder.fulfilled, (state, action) => {
        state.reminders.push(action.payload);
      })
      .addCase(updateReminder.fulfilled, (state, action) => {
        const index = state.reminders.findIndex(reminder => reminder.id === action.payload.id);
        if (index !== -1) {
          state.reminders[index] = action.payload;
        }
      })
      .addCase(deleteReminder.fulfilled, (state, action) => {
        state.reminders = state.reminders.filter(reminder => reminder.id !== action.payload);
      })
      .addCase(completeReminder.fulfilled, (state, action) => {
        const index = state.reminders.findIndex(reminder => reminder.id === action.payload.id);
        if (index !== -1) {
          state.reminders[index] = action.payload;
        }
      });
  },
});