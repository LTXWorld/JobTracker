/**
 * JobView AutoFill Extension - Background Service Worker
 * 负责与JobView API通信，管理数据存储和同步
 */

// 导入配置
self.importScripts('config.js');

// 存储键定义
const STORAGE_KEYS = {
  ACCESS_TOKEN: 'jobview_access_token',
  REFRESH_TOKEN: 'jobview_refresh_token',
  USER_DATA: 'jobview_user_data',
  RESUME_DATA: 'jobview_resume_data',
  LAST_SYNC: 'jobview_last_sync'
};

// JobView API 基础配置
const API_CONFIG = {
  AUTH_BASE_URL: CONFIG.AUTH_BASE_URL,       // 认证基址 /api/auth
  API_V1_BASE_URL: CONFIG.API_V1_BASE_URL,   // 业务基址 /api/v1
  TIMEOUT: 10000
};

class JobViewAPI {
  constructor() {
    this.authBase = API_CONFIG.AUTH_BASE_URL;
    this.v1Base = API_CONFIG.API_V1_BASE_URL;
    this.timeout = API_CONFIG.TIMEOUT;
  }

  // 获取存储的令牌
  async getTokens() {
    const result = await chrome.storage.local.get([
      STORAGE_KEYS.ACCESS_TOKEN,
      STORAGE_KEYS.REFRESH_TOKEN
    ]);
    return {
      accessToken: result[STORAGE_KEYS.ACCESS_TOKEN],
      refreshToken: result[STORAGE_KEYS.REFRESH_TOKEN]
    };
  }

  // 保存令牌
  async saveTokens(accessToken, refreshToken) {
    await chrome.storage.local.set({
      [STORAGE_KEYS.ACCESS_TOKEN]: accessToken,
      [STORAGE_KEYS.REFRESH_TOKEN]: refreshToken
    });
  }

  // 发起API请求
  // base: 'auth' | 'v1'（默认v1）
  async request(endpoint, options = {}) {
    const { accessToken } = await this.getTokens();
    const base = options.base === 'auth' ? this.authBase : this.v1Base;
    const url = `${base}${endpoint}`;

    const headers = {
      'Content-Type': 'application/json',
      ...options.headers
    };

    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    try {
      const response = await fetch(url, {
        method: options.method || 'GET',
        headers,
        body: options.body ? JSON.stringify(options.body) : undefined,
        signal: AbortSignal.timeout(this.timeout)
      });

      if (response.status === 401) {
        // Token 过期，尝试刷新
        const refreshed = await this.refreshToken();
        if (refreshed) {
          // 重新发起请求
          headers['Authorization'] = `Bearer ${refreshed}`;
          const retryResponse = await fetch(url, {
            method: options.method || 'GET',
            headers,
            body: options.body ? JSON.stringify(options.body) : undefined,
            signal: AbortSignal.timeout(this.timeout)
          });
          return await retryResponse.json();
        }
        throw new Error('Authentication failed');
      }

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('API Request failed:', error);
      throw error;
    }
  }

  // 刷新令牌
  async refreshToken() {
    const { refreshToken } = await this.getTokens();
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    try {
      const response = await fetch(`${this.authBase}/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${refreshToken}`
        }
      });

      if (response.ok) {
        const data = await response.json();
        const newAccessToken = data.data?.access_token;
        if (newAccessToken) {
          await this.saveTokens(newAccessToken, refreshToken);
          return newAccessToken;
        }
      }
      return null;
    } catch (error) {
      console.error('Token refresh failed:', error);
      return null;
    }
  }

  // 获取用户信息
  async getCurrentUser() {
    // 后端为 /api/auth/profile
    return await this.request('/profile', { base: 'auth' });
  }

  // 获取简历数据
  async getResumeData() {
    try {
      // 并行获取简历基本信息和各个部分
      const [resumeInfo, sections] = await Promise.all([
        this.request('/resumes/me', { base: 'v1' }),
        this.getResumeSections()
      ]);

      return {
        resume: resumeInfo.data,
        sections: sections
      };
    } catch (error) {
      console.error('Failed to get resume data:', error);
      throw error;
    }
  }

  // 获取简历各部分内容
  async getResumeSections() {
    try {
      const resumeInfo = await this.request('/resumes/me', { base: 'v1' });
      const resumeId = resumeInfo.data?.resume?.id;

      if (!resumeId) {
        return {};
      }

      const response = await this.request(`/resumes/${resumeId}/sections`, { base: 'v1' });
      const sections = response.data || [];

      // 转换为对象格式
      const sectionsMap = {};
      sections.forEach(section => {
        sectionsMap[section.type] = section.content;
      });

      return sectionsMap;
    } catch (error) {
      console.error('Failed to get resume sections:', error);
      return {};
    }
  }

  // 检查登录状态
  async checkLoginStatus() {
    try {
      const user = await this.getCurrentUser();
      return { isLoggedIn: true, user: user.data };
    } catch (error) {
      return { isLoggedIn: false, user: null };
    }
  }
}

class DataManager {
  constructor(api) {
    this.api = api;
  }

  // 同步简历数据
  async syncResumeData() {
    try {
      const resumeData = await this.api.getResumeData();

      // 转换为标准化格式
      const normalizedData = this.normalizeResumeData(resumeData);

      // 保存到本地存储
      await chrome.storage.local.set({
        [STORAGE_KEYS.RESUME_DATA]: normalizedData,
        [STORAGE_KEYS.LAST_SYNC]: Date.now()
      });

      return normalizedData;
    } catch (error) {
      console.error('Failed to sync resume data:', error);
      throw error;
    }
  }

  // 规范化简历数据格式
  normalizeResumeData(rawData) {
    const { resume, sections } = rawData;

    return {
      // 基础信息
      basic: {
        name: sections.base?.name || '',
        email: sections.base?.email || '',
        phone: sections.base?.phone || '',
        city: sections.base?.city || '',
        gender: sections.base?.gender || '',
        birthDate: sections.base?.birth_date || '',
        address: sections.base?.address || ''
      },

      // 求职意向
      intent: {
        position: sections.intent?.position || '',
        city: sections.intent?.city || '',
        salary: sections.intent?.salary || '',
        jobType: sections.intent?.job_type || ''
      },

      // 教育经历
      education: sections.edu || [],

      // 工作经历
      experience: sections.exp || [],

      // 项目经历
      projects: sections.project || [],

      // 技能
      skills: sections.skill || [],

      // 证书
      certificates: sections.cert || [],

      // 自我评价
      summary: sections.summary?.content || '',

      // 元数据
      meta: {
        completeness: resume?.completeness || 0,
        lastUpdated: resume?.updated_at || new Date().toISOString(),
        syncTime: Date.now()
      }
    };
  }

  // 获取本地简历数据
  async getLocalResumeData() {
    const result = await chrome.storage.local.get([STORAGE_KEYS.RESUME_DATA]);
    return result[STORAGE_KEYS.RESUME_DATA] || null;
  }

  // 检查是否需要同步
  async shouldSync() {
    const result = await chrome.storage.local.get([STORAGE_KEYS.LAST_SYNC]);
    const lastSync = result[STORAGE_KEYS.LAST_SYNC] || 0;
    const now = Date.now();
    const SYNC_INTERVAL = 5 * 60 * 1000; // 5分钟

    return (now - lastSync) > SYNC_INTERVAL;
  }
}

class BackgroundService {
  constructor() {
    this.api = new JobViewAPI();
    this.dataManager = new DataManager(this.api);
    this.init();
  }

  init() {
    // 监听来自content script和popup的消息
    chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
      this.handleMessage(request, sender, sendResponse);
      return true; // 保持消息通道开放
    });

    // 插件安装/更新时的初始化
    chrome.runtime.onInstalled.addListener(() => {
      console.log('JobView AutoFill Extension installed');
    });

    // 定期同步数据
    this.setupPeriodicSync();
  }

  async handleMessage(request, sender, sendResponse) {
    try {
      switch (request.type) {
        case 'CHECK_LOGIN':
          const loginStatus = await this.api.checkLoginStatus();
          sendResponse(loginStatus);
          break;

        case 'SYNC_RESUME_DATA':
          const resumeData = await this.dataManager.syncResumeData();
          sendResponse({ success: true, data: resumeData });
          break;

        case 'GET_RESUME_DATA':
          const localData = await this.dataManager.getLocalResumeData();
          if (!localData || await this.dataManager.shouldSync()) {
            const freshData = await this.dataManager.syncResumeData();
            sendResponse({ success: true, data: freshData });
          } else {
            sendResponse({ success: true, data: localData });
          }
          break;

        case 'ANALYZE_PAGE':
          const pageInfo = await this.analyzePage(sender.tab);
          sendResponse(pageInfo);
          break;

        case 'GET_AUTH_TOKEN':
          const tokens = await this.api.getTokens();
          sendResponse({ accessToken: tokens.accessToken });
          break;

        default:
          sendResponse({ success: false, error: 'Unknown message type' });
      }
    } catch (error) {
      console.error('Background service error:', error);
      sendResponse({ success: false, error: error.message });
    }
  }

  // 分析页面类型
  async analyzePage(tab) {
    if (!tab || !tab.url) {
      return { supported: false };
    }

    const url = new URL(tab.url);
    const hostname = url.hostname;

    // 支持的网站配置
    const supportedSites = {
      'zhaopin.com': {
        name: '智联招聘',
        patterns: ['/resume/', '/jobs/apply/']
      },
      '51job.com': {
        name: '前程无忧',
        patterns: ['/resume/', '/jobs/apply/']
      },
      'zhipin.com': {
        name: 'BOSS直聘',
        patterns: ['/web/geek/', '/job_detail/']
      }
    };

    for (const [domain, config] of Object.entries(supportedSites)) {
      if (hostname.includes(domain)) {
        const isRelevantPage = config.patterns.some(pattern =>
          url.pathname.includes(pattern)
        );

        return {
          supported: true,
          site: {
            domain: domain,
            name: config.name,
            url: tab.url,
            relevant: isRelevantPage
          }
        };
      }
    }

    return { supported: false };
  }

  // 设置定期同步
  setupPeriodicSync() {
    // 每30分钟检查一次是否需要同步
    chrome.alarms.create('syncResumeData', {
      delayInMinutes: 30,
      periodInMinutes: 30
    });

    chrome.alarms.onAlarm.addListener(async (alarm) => {
      if (alarm.name === 'syncResumeData') {
        try {
          const loginStatus = await this.api.checkLoginStatus();
          if (loginStatus.isLoggedIn && await this.dataManager.shouldSync()) {
            await this.dataManager.syncResumeData();
            console.log('Auto-sync completed');
          }
        } catch (error) {
          console.error('Auto-sync failed:', error);
        }
      }
    });
  }
}

// 初始化后台服务
const backgroundService = new BackgroundService();

// 导出给测试使用
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { BackgroundService, JobViewAPI, DataManager };
}
