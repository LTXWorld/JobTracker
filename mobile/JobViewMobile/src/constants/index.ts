// Application Status Constants
export const APPLICATION_STATUSES = [
  '已保存', '已投递', '简历筛选中', '笔试中', '一面中', '二面中',
  '三面中', 'HR面中', '等待offer', '已收到offer', '已拒绝offer',
  '简历挂', '笔试挂', '一面挂', '二面挂', '三面挂', 'HR面挂',
  '被拒绝', '已入职', '已放弃'
] as const;

export const APPLICATION_PRIORITIES = ['low', 'medium', 'high'] as const;

export const WORK_TYPES = ['remote', 'onsite', 'hybrid'] as const;

// UI Constants
export const THEME_MODES = ['light', 'dark', 'system'] as const;

export const SCREEN_NAMES = {
  SPLASH: 'Splash',
  AUTH: 'Auth',
  MAIN: 'Main',
  LOGIN: 'Login',
  REGISTER: 'Register',
  FORGOT_PASSWORD: 'ForgotPassword',
  DASHBOARD: 'Dashboard',
  APPLICATIONS: 'Applications',
  KANBAN: 'Kanban',
  STATISTICS: 'Statistics',
  PROFILE: 'Profile',
} as const;
