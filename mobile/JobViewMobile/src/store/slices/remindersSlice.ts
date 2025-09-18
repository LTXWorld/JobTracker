import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { ApplicationReminder } from '@types';
import { NotificationService } from '@services/notificationService';

interface RemindersState {
  reminders: ApplicationReminder[];
  isLoading: boolean;
  error: string | null;
  notificationPermission: 'granted' | 'denied' | 'unknown';
  settings: {
    enabled: boolean;
    interviewReminders: boolean;
    followUpReminders: boolean;
    statusUpdates: boolean;
    sound: boolean;
    vibration: boolean;
  };
}

const initialState: RemindersState = {
  reminders: [],
  isLoading: false,
  error: null,
  notificationPermission: 'unknown',
  settings: {
    enabled: true,
    interviewReminders: true,
    followUpReminders: true,
    statusUpdates: true,
    sound: true,
    vibration: true,
  },
};

// Async thunks
export const initializeNotifications = createAsyncThunk(
  'reminders/initializeNotifications',
  async (_, { rejectWithValue }) => {
    try {
      const initialized = await NotificationService.initialize();
      if (!initialized) {
        return rejectWithValue('Failed to initialize notifications');
      }

      const permission = await NotificationService.getPermissionStatus();
      const settings = await NotificationService.getSettings();

      return { permission, settings };
    } catch (error: any) {
      return rejectWithValue(error.message || 'Failed to initialize notifications');
    }
  }
);

export const requestNotificationPermission = createAsyncThunk(
  'reminders/requestPermission',
  async (_, { rejectWithValue }) => {
    try {
      const granted = await NotificationService.requestPermissions();
      return granted ? 'granted' : 'denied';
    } catch (error: any) {
      return rejectWithValue(error.message || 'Failed to request permission');
    }
  }
);

export const scheduleReminder = createAsyncThunk(
  'reminders/scheduleReminder',
  async (reminder: ApplicationReminder, { rejectWithValue }) => {
    try {
      const scheduled = await NotificationService.scheduleReminder(reminder);
      if (!scheduled) {
        return rejectWithValue('Failed to schedule reminder');
      }
      return reminder;
    } catch (error: any) {
      return rejectWithValue(error.message || 'Failed to schedule reminder');
    }
  }
);

export const cancelReminder = createAsyncThunk(
  'reminders/cancelReminder',
  async (reminderId: string, { rejectWithValue }) => {
    try {
      await NotificationService.cancelReminder(reminderId);
      return reminderId;
    } catch (error: any) {
      return rejectWithValue(error.message || 'Failed to cancel reminder');
    }
  }
);

export const updateNotificationSettings = createAsyncThunk(
  'reminders/updateSettings',
  async (settings: any, { rejectWithValue }) => {
    try {
      await NotificationService.saveSettings(settings);
      return settings;
    } catch (error: any) {
      return rejectWithValue(error.message || 'Failed to update settings');
    }
  }
);

export const loadScheduledNotifications = createAsyncThunk(
  'reminders/loadScheduled',
  async (_, { rejectWithValue }) => {
    try {
      const notifications = await NotificationService.getScheduledNotifications();
      return notifications;
    } catch (error: any) {
      return rejectWithValue(error.message || 'Failed to load notifications');
    }
  }
);

export const cleanupExpiredNotifications = createAsyncThunk(
  'reminders/cleanupExpired',
  async (_, { rejectWithValue }) => {
    try {
      await NotificationService.cleanupExpiredNotifications();
      return true;
    } catch (error: any) {
      return rejectWithValue(error.message || 'Failed to cleanup notifications');
    }
  }
);

export const createRemindersSlice = () => createSlice({
  name: 'reminders',
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
    addReminder: (state, action: PayloadAction<ApplicationReminder>) => {
      state.reminders.push(action.payload);
    },
    removeReminder: (state, action: PayloadAction<string>) => {
      state.reminders = state.reminders.filter(reminder => reminder.id !== action.payload);
    },
    updateReminder: (state, action: PayloadAction<{ id: string; updates: Partial<ApplicationReminder> }>) => {
      const { id, updates } = action.payload;
      const index = state.reminders.findIndex(reminder => reminder.id === id);
      if (index !== -1) {
        state.reminders[index] = { ...state.reminders[index], ...updates };
      }
    },
    setNotificationPermission: (state, action: PayloadAction<'granted' | 'denied' | 'unknown'>) => {
      state.notificationPermission = action.payload;
    },
    updateSettings: (state, action: PayloadAction<Partial<RemindersState['settings']>>) => {
      state.settings = { ...state.settings, ...action.payload };
    },
    resetReminders: () => initialState,
  },
  extraReducers: (builder) => {
    // Initialize notifications
    builder
      .addCase(initializeNotifications.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(initializeNotifications.fulfilled, (state, action) => {
        state.isLoading = false;
        state.notificationPermission = action.payload.permission;
        state.settings = {
          ...state.settings,
          ...action.payload.settings,
        };
      })
      .addCase(initializeNotifications.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Request permission
    builder
      .addCase(requestNotificationPermission.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(requestNotificationPermission.fulfilled, (state, action) => {
        state.isLoading = false;
        state.notificationPermission = action.payload;
      })
      .addCase(requestNotificationPermission.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Schedule reminder
    builder
      .addCase(scheduleReminder.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(scheduleReminder.fulfilled, (state, action) => {
        state.isLoading = false;
        state.reminders.push(action.payload);
      })
      .addCase(scheduleReminder.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Cancel reminder
    builder
      .addCase(cancelReminder.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(cancelReminder.fulfilled, (state, action) => {
        state.isLoading = false;
        state.reminders = state.reminders.filter(reminder => reminder.id !== action.payload);
      })
      .addCase(cancelReminder.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Update settings
    builder
      .addCase(updateNotificationSettings.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(updateNotificationSettings.fulfilled, (state, action) => {
        state.isLoading = false;
        state.settings = { ...state.settings, ...action.payload };
      })
      .addCase(updateNotificationSettings.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Load scheduled notifications
    builder
      .addCase(loadScheduledNotifications.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loadScheduledNotifications.fulfilled, (state) => {
        state.isLoading = false;
        // Note: scheduled notifications are handled separately from reminders
      })
      .addCase(loadScheduledNotifications.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Cleanup expired
    builder
      .addCase(cleanupExpiredNotifications.pending, (state) => {
        state.error = null;
      })
      .addCase(cleanupExpiredNotifications.fulfilled, (state) => {
        // Cleanup completed
      })
      .addCase(cleanupExpiredNotifications.rejected, (state, action) => {
        state.error = action.payload as string;
      });
  },
});