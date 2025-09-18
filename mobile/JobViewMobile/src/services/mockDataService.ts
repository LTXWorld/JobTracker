// Mock data service for development
import { JobApplication, ApplicationNote, ApplicationReminder, TimelineEvent, ApplicationStatus, ApplicationPriority } from '@types';
import { generateId } from '@utils';

// Mock job applications data
const mockJobApplications: JobApplication[] = [
  {
    id: '1',
    userId: '1',
    companyName: '腾讯',
    position: '高级前端开发工程师',
    location: '深圳',
    salaryRange: '20-35K',
    workType: 'hybrid',
    description: '负责微信小程序和Web前端开发，使用Vue.js和React技术栈',
    requirements: '3年以上前端开发经验，熟悉Vue.js/React框架',
    status: '一面中',
    priority: 'high',
    source: 'BOSS直聘',
    jobUrl: 'https://example.com/job1',
    contactPerson: '张经理',
    contactEmail: 'zhang@tencent.com',
    contactPhone: '13800138001',
    applicationDate: new Date('2024-01-15'),
    createdAt: new Date('2024-01-15'),
    updatedAt: new Date('2024-01-20'),
    tags: ['前端', 'Vue.js', 'React', '大厂'],
    notes: '需要准备Vue.js相关的技术问题，特别是组件通信和状态管理。记得复习React hooks的使用场景。',
    notesArray: [
      {
        id: 'note1',
        applicationId: '1',
        title: '面试准备',
        content: '需要准备Vue.js相关的技术问题，特别是组件通信和状态管理',
        type: 'interview',
        createdAt: new Date('2024-01-16'),
        updatedAt: new Date('2024-01-16'),
      },
    ],
    reminders: [
      {
        id: 'reminder1',
        applicationId: '1',
        title: '一面提醒',
        description: '腾讯前端开发岗位一面，地点：腾讯大厦',
        reminderDate: new Date('2024-01-25'),
        isCompleted: false,
        notificationType: 'local',
        createdAt: new Date('2024-01-16'),
        updatedAt: new Date('2024-01-16'),
      },
    ],
    timeline: [
      {
        id: 'event1',
        applicationId: '1',
        eventType: 'application_created',
        title: '创建投递记录',
        description: '在系统中创建了腾讯前端开发岗位的投递记录',
        eventDate: new Date('2024-01-15'),
        createdAt: new Date('2024-01-15'),
        updatedAt: new Date('2024-01-15'),
      },
      {
        id: 'event2',
        applicationId: '1',
        eventType: 'application_submitted',
        title: '简历投递',
        description: '通过BOSS直聘投递简历',
        eventDate: new Date('2024-01-15'),
        createdAt: new Date('2024-01-15'),
        updatedAt: new Date('2024-01-15'),
      },
      {
        id: 'event3',
        applicationId: '1',
        eventType: 'status_changed',
        title: '状态更新',
        description: '状态从"已投递"更新为"一面中"',
        eventDate: new Date('2024-01-20'),
        createdAt: new Date('2024-01-20'),
        updatedAt: new Date('2024-01-20'),
      },
    ],
  },
  {
    id: '2',
    userId: '1',
    companyName: '阿里巴巴',
    position: '全栈开发工程师',
    location: '杭州',
    salaryRange: '25-40K',
    workType: 'onsite',
    description: '负责电商平台的全栈开发，包括前端React和后端Node.js',
    requirements: '5年以上开发经验，熟悉React、Node.js、数据库设计',
    status: '等待offer',
    priority: 'high',
    source: '拉勾网',
    jobUrl: 'https://example.com/job2',
    contactPerson: '李HR',
    contactEmail: 'li@alibaba.com',
    applicationDate: new Date('2024-01-10'),
    createdAt: new Date('2024-01-10'),
    updatedAt: new Date('2024-01-22'),
    tags: ['全栈', 'React', 'Node.js', '阿里'],
    notes: '',
    notesArray: [],
    reminders: [],
    timeline: [
      {
        id: 'event4',
        applicationId: '2',
        eventType: 'application_created',
        title: '创建投递记录',
        eventDate: new Date('2024-01-10'),
        createdAt: new Date('2024-01-10'),
        updatedAt: new Date('2024-01-10'),
      },
      {
        id: 'event5',
        applicationId: '2',
        eventType: 'interview_completed',
        title: '终面完成',
        description: '完成了技术终面，等待HR通知结果',
        eventDate: new Date('2024-01-22'),
        createdAt: new Date('2024-01-22'),
        updatedAt: new Date('2024-01-22'),
      },
    ],
  },
  {
    id: '3',
    userId: '1',
    companyName: '字节跳动',
    position: 'React Native开发工程师',
    location: '北京',
    salaryRange: '22-32K',
    workType: 'remote',
    description: '负责移动端APP开发，使用React Native框架',
    requirements: '熟悉React Native开发，有iOS/Android开发经验优先',
    status: '简历挂',
    priority: 'medium',
    source: '内推',
    applicationDate: new Date('2024-01-12'),
    createdAt: new Date('2024-01-12'),
    updatedAt: new Date('2024-01-18'),
    tags: ['React Native', '移动开发', '字节跳动'],
    notes: '',
    notesArray: [],
    reminders: [],
    timeline: [
      {
        id: 'event6',
        applicationId: '3',
        eventType: 'application_created',
        title: '创建投递记录',
        eventDate: new Date('2024-01-12'),
        createdAt: new Date('2024-01-12'),
        updatedAt: new Date('2024-01-12'),
      },
      {
        id: 'event7',
        applicationId: '3',
        eventType: 'status_changed',
        title: '简历被拒',
        description: '简历筛选未通过，可能是经验不够匹配',
        eventDate: new Date('2024-01-18'),
        createdAt: new Date('2024-01-18'),
        updatedAt: new Date('2024-01-18'),
      },
    ],
  },
  {
    id: '4',
    userId: '1',
    companyName: '美团',
    position: '前端开发工程师',
    location: '北京',
    salaryRange: '18-28K',
    workType: 'hybrid',
    description: '负责美团外卖前端页面开发和优化',
    status: '已投递',
    priority: 'medium',
    source: '智联招聘',
    applicationDate: new Date('2024-01-20'),
    createdAt: new Date('2024-01-20'),
    updatedAt: new Date('2024-01-20'),
    tags: ['前端', 'Vue.js', '美团'],
    notes: '',
    notesArray: [],
    reminders: [],
    timeline: [
      {
        id: 'event8',
        applicationId: '4',
        eventType: 'application_created',
        title: '创建投递记录',
        eventDate: new Date('2024-01-20'),
        createdAt: new Date('2024-01-20'),
        updatedAt: new Date('2024-01-20'),
      },
    ],
  },
  {
    id: '5',
    userId: '1',
    companyName: '滴滴出行',
    position: '高级前端开发工程师',
    location: '北京',
    salaryRange: '25-35K',
    workType: 'onsite',
    status: '二面中',
    priority: 'high',
    source: '猎头推荐',
    applicationDate: new Date('2024-01-08'),
    createdAt: new Date('2024-01-08'),
    updatedAt: new Date('2024-01-23'),
    tags: ['前端', '高级', '滴滴'],
    notes: '',
    notesArray: [],
    reminders: [
      {
        id: 'reminder2',
        applicationId: '5',
        title: '二面准备',
        description: '准备系统设计和项目经验分享',
        reminderDate: new Date('2024-01-26'),
        isCompleted: false,
        notificationType: 'local',
        createdAt: new Date('2024-01-23'),
        updatedAt: new Date('2024-01-23'),
      },
    ],
    timeline: [],
  },
];

// Mock API delay
const delay = (ms: number) => new Promise<void>(resolve => setTimeout(() => resolve(), ms));

export const mockDataService = {
  // Get all applications for a user
  getApplications: async (userId: string): Promise<JobApplication[]> => {
    await delay(500);
    return mockJobApplications.filter(app => app.userId === userId);
  },

  // Get application by ID
  getApplicationById: async (id: string): Promise<JobApplication | null> => {
    await delay(300);
    return mockJobApplications.find(app => app.id === id) || null;
  },

  // Create new application
  createApplication: async (application: Omit<JobApplication, 'id' | 'createdAt' | 'updatedAt'>): Promise<JobApplication> => {
    await delay(800);
    const newApplication: JobApplication = {
      ...application,
      id: generateId(),
      createdAt: new Date(),
      updatedAt: new Date(),
      notes: application.notes || '',
      notesArray: application.notesArray || [],
      reminders: application.reminders || [],
      timeline: application.timeline || [
        {
          id: generateId(),
          applicationId: '',
          eventType: 'application_created',
          title: '创建投递记录',
          description: `创建了${application.companyName}的投递记录`,
          eventDate: new Date(),
          createdAt: new Date(),
          updatedAt: new Date(),
        },
      ],
    };

    // Update timeline with correct application ID
    newApplication.timeline = newApplication.timeline.map(event => ({
      ...event,
      applicationId: newApplication.id,
    }));

    mockJobApplications.push(newApplication);
    return newApplication;
  },

  // Update application
  updateApplication: async (id: string, updates: Partial<JobApplication>): Promise<JobApplication | null> => {
    await delay(600);
    const index = mockJobApplications.findIndex(app => app.id === id);
    if (index === -1) return null;

    const oldStatus = mockJobApplications[index].status;
    mockJobApplications[index] = {
      ...mockJobApplications[index],
      ...updates,
      updatedAt: new Date(),
    };

    // Add timeline event if status changed
    if (updates.status && updates.status !== oldStatus) {
      const timelineEvent: TimelineEvent = {
        id: generateId(),
        applicationId: id,
        eventType: 'status_changed',
        title: '状态更新',
        description: `状态从"${oldStatus}"更新为"${updates.status}"`,
        eventDate: new Date(),
        createdAt: new Date(),
        updatedAt: new Date(),
      };
      mockJobApplications[index].timeline.push(timelineEvent);
    }

    return mockJobApplications[index];
  },

  // Delete application
  deleteApplication: async (id: string): Promise<boolean> => {
    await delay(400);
    const index = mockJobApplications.findIndex(app => app.id === id);
    if (index === -1) return false;

    mockJobApplications.splice(index, 1);
    return true;
  },

  // Get applications by status for Kanban view
  getApplicationsByStatus: async (userId: string): Promise<Record<ApplicationStatus, JobApplication[]>> => {
    await delay(400);
    const userApplications = mockJobApplications.filter(app => app.userId === userId);

    const statusGroups: Record<ApplicationStatus, JobApplication[]> = {
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

    userApplications.forEach(app => {
      statusGroups[app.status].push(app);
    });

    return statusGroups;
  },

  // Get statistics
  getStatistics: async (userId: string): Promise<{
    total: number;
    byStatus: Record<ApplicationStatus, number>;
    byPriority: Record<ApplicationPriority, number>;
    byWorkType: Record<string, number>;
    recentActivity: number;
  }> => {
    await delay(600);
    const userApplications = mockJobApplications.filter(app => app.userId === userId);

    const byStatus: Record<ApplicationStatus, number> = {
      '已保存': 0,
      '已投递': 0,
      '简历筛选中': 0,
      '笔试中': 0,
      '一面中': 0,
      '二面中': 0,
      '三面中': 0,
      'HR面中': 0,
      '等待offer': 0,
      '已收到offer': 0,
      '已拒绝offer': 0,
      '简历挂': 0,
      '笔试挂': 0,
      '一面挂': 0,
      '二面挂': 0,
      '三面挂': 0,
      'HR面挂': 0,
      '被拒绝': 0,
      '已入职': 0,
      '已放弃': 0,
    };

    const byPriority: Record<ApplicationPriority, number> = {
      low: 0,
      medium: 0,
      high: 0,
    };

    const byWorkType: Record<string, number> = {
      remote: 0,
      onsite: 0,
      hybrid: 0,
    };

    userApplications.forEach(app => {
      byStatus[app.status]++;
      byPriority[app.priority]++;
      byWorkType[app.workType]++;
    });

    // Count recent activity (last 7 days)
    const sevenDaysAgo = new Date();
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
    const recentActivity = userApplications.filter(app =>
      app.updatedAt > sevenDaysAgo
    ).length;

    return {
      total: userApplications.length,
      byStatus,
      byPriority,
      byWorkType,
      recentActivity,
    };
  },

  // Search applications
  searchApplications: async (userId: string, query: string): Promise<JobApplication[]> => {
    await delay(300);
    const userApplications = mockJobApplications.filter(app => app.userId === userId);

    if (!query.trim()) return userApplications;

    const searchQuery = query.toLowerCase();
    return userApplications.filter(app =>
      app.companyName.toLowerCase().includes(searchQuery) ||
      app.position.toLowerCase().includes(searchQuery) ||
      app.location.toLowerCase().includes(searchQuery) ||
      app.tags.some(tag => tag.toLowerCase().includes(searchQuery)) ||
      app.source.toLowerCase().includes(searchQuery)
    );
  },
};

export default mockDataService;