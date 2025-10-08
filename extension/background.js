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
  LAST_SYNC: 'jobview_last_sync',
  PENDING_APPLICATIONS: 'jobview_pending_applications',
  RECENT_KEYS: 'jobview_recent_app_keys',
  SETTINGS: 'jobview_settings',
  ENABLED_SITES: 'jobview_enabled_sites',
  ENVIRONMENT: 'jobview_environment'
};

// JobView API 基础配置
const API_TIMEOUT = 10000;

class JobViewAPI {
  constructor(envName) {
    this.timeout = API_TIMEOUT;
    this.setEnvironment(envName || detectEnvironment());
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

  async logout() {
    try {
      const { refreshToken, accessToken } = await this.getTokens();
      if (!accessToken && !refreshToken) {
        return true;
      }
      const headers = { 'Content-Type': 'application/json' };
      if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
      } else if (refreshToken) {
        headers['Authorization'] = `Bearer ${refreshToken}`;
      }
      await fetch(`${this.authBase}/logout`, {
        method: 'POST',
        headers,
        body: JSON.stringify({}),
        signal: AbortSignal.timeout(this.timeout)
      });
    } catch (error) {
      console.warn('Logout request failed:', error?.message || error);
    }
    return true;
  }

  setEnvironment(envName) {
    const config = getConfig(envName);
    this.envName = envName;
    this.authBase = config.AUTH_BASE_URL;
    this.v1Base = config.API_V1_BASE_URL;
    this.frontendBase = config.FRONTEND_URL;
  }

  getEnvironment() {
    return {
      name: this.envName,
      config: getConfig(this.envName)
    };
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
    this.envName = detectEnvironment();
    this.api = new JobViewAPI(this.envName);
    this.dataManager = new DataManager(this.api);
    this.init();
  }

  async getEnabledSites() {
    const result = await chrome.storage.local.get([STORAGE_KEYS.ENABLED_SITES]);
    return result[STORAGE_KEYS.ENABLED_SITES] || {};
  }

  async setEnabledSites(sites) {
    await chrome.storage.local.set({ [STORAGE_KEYS.ENABLED_SITES]: sites });
    return sites;
  }

  async toggleSite(host, enabled) {
    if (!host) return await this.getEnabledSites();
    const cleanHost = sanitizeHost(host);
    const sites = await this.getEnabledSites();
    if (enabled) {
      sites[cleanHost] = { enabled: true, updatedAt: Date.now() };
    } else {
      delete sites[cleanHost];
    }
    await this.setEnabledSites(sites);
    return sites;
  }

  async isSiteEnabled(host) {
    if (!host) return false;
    const sites = await this.getEnabledSites();
    const cleanHost = sanitizeHost(host);
    if (sites[cleanHost]?.enabled) {
      return true;
    }
    // 支持通配符 *.domain.com
    const parts = cleanHost.split('.');
    for (let i = 0; i < parts.length - 1; i++) {
      const suffix = parts.slice(i).join('.');
      const wildcardKey = `*.${suffix}`;
      if (sites[wildcardKey]?.enabled) {
        return true;
      }
    }
    return false;
  }

  init() {
    chrome.storage.local.get([STORAGE_KEYS.ENVIRONMENT], (res) => {
      if (res && res[STORAGE_KEYS.ENVIRONMENT]) {
        this.envName = res[STORAGE_KEYS.ENVIRONMENT];
        this.api.setEnvironment(this.envName);
      } else {
        chrome.storage.local.set({ [STORAGE_KEYS.ENVIRONMENT]: this.envName });
      }
    });

    chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
      this.handleMessage(request, sender, sendResponse);
      return true;
    });

    // 插件安装/更新时的初始化
    chrome.runtime.onInstalled.addListener(() => {
      console.log('JobView AutoFill Extension installed');
    });

    // 定期同步数据
    this.setupPeriodicSync();

    // 定期重试待提交的投递记录
    this.setupPendingRetry();
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

        case 'SAVE_TOKENS':
          try {
            const { accessToken, refreshToken } = request;
            if (!accessToken || !refreshToken) {
              sendResponse({ success: false, error: 'Missing tokens' });
              break;
            }
            await this.api.saveTokens(accessToken, refreshToken);
            // 尝试校验并预同步数据（不阻塞太久）
            let synced = false;
            try {
              const status = await this.api.checkLoginStatus();
              if (status.isLoggedIn) {
                await this.dataManager.syncResumeData();
                synced = true;
              }
            } catch (e) {
              // 同步失败不应影响存储结果
              console.warn('Post-login sync failed:', e?.message || e);
            }
            // 通知可能打开的popup更新UI（非强依赖）
            try { chrome.runtime.sendMessage({ type: 'UPDATE_UI' }); } catch {}
            sendResponse({ success: true, synced });
          } catch (e) {
            console.error('SAVE_TOKENS error:', e);
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'CREATE_APPLICATION':
          try {
            const result = await this.createApplication(request.payload || {}, { force: !!request.force });
            sendResponse(result);
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'GET_SETTINGS':
          try {
            const res = await chrome.storage.local.get([STORAGE_KEYS.SETTINGS]);
            sendResponse({ success: true, data: res[STORAGE_KEYS.SETTINGS] || {} });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'SAVE_SETTINGS':
          try {
            const current = (await chrome.storage.local.get([STORAGE_KEYS.SETTINGS]))[STORAGE_KEYS.SETTINGS] || {};
            const next = { ...current, ...(request.data || {}) };
            await chrome.storage.local.set({ [STORAGE_KEYS.SETTINGS]: next });
            sendResponse({ success: true, data: next });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'GET_ENABLED_SITES':
          try {
            const data = await this.getEnabledSites();
            sendResponse({ success: true, data });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'SET_SITE_ENABLED':
          try {
            const host = request.host || request.domain;
            const enabled = Boolean(request.enabled);
            const sites = await this.toggleSite(host, enabled);
            // 通知当前 tab 更新监听状态
            if (sender?.tab?.id != null) {
              try {
                chrome.tabs.sendMessage(sender.tab.id, {
                  type: 'SITE_STATUS_CHANGED',
                  host: sanitizeHost(host),
                  enabled
                });
              } catch (err) {
                console.warn('Notify content script failed:', err?.message || err);
              }
            } else if (request.tabId != null) {
              try {
                chrome.tabs.sendMessage(request.tabId, {
                  type: 'SITE_STATUS_CHANGED',
                  host: sanitizeHost(host),
                  enabled
                });
              } catch (err) {
                console.warn('Notify content script failed:', err?.message || err);
              }
            }
            sendResponse({ success: true, data: sites });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'IS_SITE_ENABLED':
          try {
            const enabled = await this.isSiteEnabled(request.host || request.domain);
            sendResponse({ success: true, enabled });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'GET_ENVIRONMENT':
          try {
            const env = this.api.getEnvironment();
            sendResponse({ success: true, environment: env.name, config: env.config });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'SET_ENVIRONMENT':
          try {
            const target = request?.environment;
            if (!target || !ENV_CONFIG[target]) {
              throw new Error('无效的环境名称');
            }
            await chrome.storage.local.set({ [STORAGE_KEYS.ENVIRONMENT]: target });
            this.envName = target;
            this.api.setEnvironment(target);
            await this.handleLogout();
            sendResponse({ success: true, environment: target, config: getConfig(target) });
            try { chrome.runtime.sendMessage({ type: 'UPDATE_UI' }); } catch {}
            chrome.tabs.query({}, (tabs) => {
              tabs?.forEach(tab => {
                if (tab?.id != null) {
                  try {
                    chrome.tabs.sendMessage(tab.id, { type: 'AUTH_STATE_CHANGED', isLoggedIn: false });
                  } catch (err) {
                    /* ignore */
                  }
                }
              });
            });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'LOGOUT':
          try {
            await this.handleLogout();
            sendResponse({ success: true });
            try { chrome.runtime.sendMessage({ type: 'UPDATE_UI' }); } catch {}
            chrome.tabs.query({}, (tabs) => {
              tabs?.forEach(tab => {
                if (tab?.id != null) {
                  try {
                    chrome.tabs.sendMessage(tab.id, { type: 'AUTH_STATE_CHANGED', isLoggedIn: false });
                  } catch (err) {
                    /* ignore */
                  }
                }
              });
            });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
          break;

        case 'PROCESS_PENDING_APPLICATIONS':
          try {
            const processed = await this.processPendingApplications();
            sendResponse({ success: true, processed });
          } catch (e) {
            sendResponse({ success: false, error: e?.message || String(e) });
          }
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
      },
      'mokahr.com': {
        name: 'Moka 招聘',
        patterns: ['/apply', '/campus-recruitment/', '/job/']
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

  // ============== 投递记录创建与重试 ==============
  async createApplication(payload, options = {}) {
    // 读取设置，补齐默认值
    const settings = (await chrome.storage.local.get([STORAGE_KEYS.SETTINGS]))[STORAGE_KEYS.SETTINGS] || {};

    const today = new Date();
    const yyyy = today.getFullYear();
    const mm = String(today.getMonth() + 1).padStart(2, '0');
    const dd = String(today.getDate()).padStart(2, '0');
    const defaultDate = `${yyyy}-${mm}-${dd}`;

    const companyName = (payload.company_name || '').trim();
    const positionTitle = (payload.position_title || '').trim();
    const companyAttribute = (payload.company_attribute || settings.default_company_attribute || '').trim();
    const applicationDate = (payload.application_date || defaultDate).trim();

    if (!companyName) throw new Error('缺少公司名称');
    if (!positionTitle) throw new Error('缺少职位名称');
    if (!companyAttribute) throw new Error('缺少企业属性（央国企/私企）');

    // 简单去重：同一日、同一域名、同一公司+职位
    const domain = (() => { try { return new URL(payload.job_url || '').hostname || (payload.source_domain || ''); } catch { return payload.source_domain || ''; } })();
    const key = `${companyName}||${positionTitle}||${applicationDate}||${domain}`;
    const recent = (await chrome.storage.local.get([STORAGE_KEYS.RECENT_KEYS]))[STORAGE_KEYS.RECENT_KEYS] || [];
    if (!options.force && recent.includes(key)) {
      return { success: false, duplicate: true, message: '疑似重复记录（公司/职位/日期相同）' };
    }

    const notesPrefix = (payload.source_domain || payload.job_url)
      ? `来源: ${payload.source_domain || ''} ${payload.job_url || ''}`.trim()
      : '';
    const notesText = [notesPrefix, payload.notes || ''].filter(Boolean).join(' | ');

    const body = {
      company_name: companyName,
      position_title: positionTitle,
      application_date: applicationDate,
      status: '已投递',
      company_attribute: companyAttribute,
      job_description: payload.job_description || undefined,
      salary_range: payload.salary_range || (payload.salary_text || undefined),
      work_location: payload.work_location || (payload.location_text || undefined),
      contact_info: payload.contact_info || undefined,
      notes: notesText || undefined
    };

    try {
      const resp = await this.api.request('/applications', { base: 'v1', method: 'POST', body });
      // 写入最近键
      const nextRecent = [key, ...recent.filter(k => k !== key)].slice(0, 100);
      await chrome.storage.local.set({ [STORAGE_KEYS.RECENT_KEYS]: nextRecent });
      return { success: true, data: resp?.data };
    } catch (e) {
      // 网络/鉴权问题：进入队列
      const isNetwork = (e?.name === 'TypeError') || /Failed to fetch|NetworkError|timeout/i.test(e?.message || '');
      if (isNetwork || !navigator.onLine) {
        await this.enqueuePending({ body, key, createdAt: Date.now() });
        return { success: false, queued: true, message: '网络不可用，已加入稍后提交队列' };
      }
      throw e;
    }
  }

  async enqueuePending(item) {
    const list = (await chrome.storage.local.get([STORAGE_KEYS.PENDING_APPLICATIONS]))[STORAGE_KEYS.PENDING_APPLICATIONS] || [];
    list.push(item);
    await chrome.storage.local.set({ [STORAGE_KEYS.PENDING_APPLICATIONS]: list });
  }

  async processPendingApplications() {
    const result = await chrome.storage.local.get([STORAGE_KEYS.PENDING_APPLICATIONS, STORAGE_KEYS.RECENT_KEYS]);
    let list = result[STORAGE_KEYS.PENDING_APPLICATIONS] || [];
    let recent = result[STORAGE_KEYS.RECENT_KEYS] || [];
    const processed = { success: 0, failed: 0 };

    const remain = [];
    for (const item of list) {
      try {
        const resp = await this.api.request('/applications', { base: 'v1', method: 'POST', body: item.body });
        const key = item.key;
        recent = [key, ...recent.filter(k => k !== key)].slice(0, 100);
        processed.success++;
      } catch (e) {
        // 保留到下次
        remain.push(item);
        processed.failed++;
      }
    }
    await chrome.storage.local.set({ [STORAGE_KEYS.PENDING_APPLICATIONS]: remain, [STORAGE_KEYS.RECENT_KEYS]: recent });
    return processed;
  }

  setupPendingRetry() {
    chrome.alarms.create('syncPendingApplications', { delayInMinutes: 5, periodInMinutes: 10 });
    chrome.alarms.onAlarm.addListener(async (alarm) => {
      if (alarm.name === 'syncPendingApplications') {
        try {
          await this.processPendingApplications();
        } catch (e) {
          console.warn('Retry pending applications failed:', e);
        }
      }
    });
  }

  async handleLogout() {
    await this.api.logout().catch(() => {});
    await chrome.storage.local.remove([
      STORAGE_KEYS.ACCESS_TOKEN,
      STORAGE_KEYS.REFRESH_TOKEN,
      STORAGE_KEYS.USER_DATA,
      STORAGE_KEYS.RESUME_DATA,
      STORAGE_KEYS.LAST_SYNC,
      STORAGE_KEYS.PENDING_APPLICATIONS,
      STORAGE_KEYS.RECENT_KEYS
    ]);
  }
}

// 初始化后台服务
const backgroundService = new BackgroundService();

// 导出给测试使用
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { BackgroundService, JobViewAPI, DataManager };
}

function sanitizeHost(host) {
  if (!host) return '';
  return String(host).trim().toLowerCase().replace(/^https?:\/\//, '').split('/')[0];
}
