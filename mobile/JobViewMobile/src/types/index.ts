// Base Types
export interface BaseEntity {
  id: string;
  createdAt: Date;
  updatedAt: Date;
}

// User Types
export interface User extends BaseEntity {
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  avatar?: string;
  preferences: UserPreferences;
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'system';
  language: 'zh' | 'en';
  notifications: NotificationSettings;
  biometricEnabled: boolean;
}

export interface NotificationSettings {
  enabled: boolean;
  interviewReminders: boolean;
  followUpReminders: boolean;
  statusUpdates: boolean;
  sound: boolean;
  vibration: boolean;
}

// Notes and Reminders
export interface ApplicationNote extends BaseEntity {
  applicationId: string;
  title: string;
  content: string;
  type: 'general' | 'interview' | 'offer' | 'followup';
}

export interface ApplicationReminder extends BaseEntity {
  applicationId: string;
  title: string;
  description?: string;
  reminderDate: Date;
  isCompleted: boolean;
  notificationType: 'local' | 'push';
}

// Timeline Events
export interface TimelineEvent extends BaseEntity {
  applicationId: string;
  eventType: TimelineEventType;
  title: string;
  description?: string;
  eventDate: Date;
  metadata?: Record<string, any>;
}

export type TimelineEventType =
  | 'application_created' | 'application_submitted' | 'status_changed'
  | 'interview_scheduled' | 'interview_completed' | 'offer_received'
  | 'offer_accepted' | 'offer_rejected' | 'application_withdrawn'
  | 'note_added' | 'reminder_set' | 'custom';

// Job Application Types
export interface JobApplication extends BaseEntity {
  userId: string;
  companyName: string;
  position: string;
  location: string;
  salaryRange?: string;
  workType: 'remote' | 'onsite' | 'hybrid';
  description?: string;
  requirements?: string;
  status: ApplicationStatus;
  priority: ApplicationPriority;
  source: string;
  jobUrl?: string;
  companyUrl?: string;
  department?: string;
  experienceLevel?: string;
  contactPerson?: string;
  contactEmail?: string;
  contactPhone?: string;
  applicationDate: Date;
  notes?: string;
  notesArray?: ApplicationNote[];
  reminders: ApplicationReminder[];
  timeline: TimelineEvent[];
  tags: string[];
}

export type ApplicationStatus =
  | '已保存' | '已投递' | '简历筛选中' | '笔试中' | '一面中' | '二面中'
  | '三面中' | 'HR面中' | '等待offer' | '已收到offer' | '已拒绝offer'
  | '简历挂' | '笔试挂' | '一面挂' | '二面挂' | '三面挂' | 'HR面挂'
  | '被拒绝' | '已入职' | '已放弃';

export type ApplicationPriority = 'low' | 'medium' | 'high';

// Navigation Types
export type RootStackParamList = {
  Splash: undefined;
  Auth: undefined;
  Main: undefined;
};

export type AuthStackParamList = {
  Login: undefined;
  Register: undefined;
  ForgotPassword: undefined;
};

export type MainTabParamList = {
  Dashboard: undefined;
  Applications: undefined;
  Kanban: undefined;
  Statistics: undefined;
  Profile: undefined;
};

export type ApplicationsStackParamList = {
  ApplicationsList: undefined;
  AddApplication: undefined;
  EditApplication: { applicationId: string };
  ApplicationDetails: { applicationId: string };
};

export type ProfileStackParamList = {
  ProfileMain: undefined;
  DataManagement: undefined;
  Reminders: undefined;
};
