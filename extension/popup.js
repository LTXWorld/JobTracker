/**
 * JobView AutoFill Extension - Popup Script
 * 负责插件弹窗的交互逻辑
 */

// 环境配置
const ENV_CONFIG = {
  // 本地开发环境
  local: {
    API_BASE_URL: 'http://localhost:8010/api/v1',
    FRONTEND_URL: 'http://localhost:3000'
  },
  // 生产环境（云服务器）
  production: {
    API_BASE_URL: 'https://jobview.bfsmlt.top/api/v1',
    FRONTEND_URL: 'https://jobview.bfsmlt.top'
  }
};

// 自动检测环境
function detectEnvironment() {
  // 如果是本地开发环境
  if (location.hostname === 'localhost' || location.hostname === '127.0.0.1') {
    return 'local';
  }
  // 生产环境
  return 'production';
}

// 获取当前环境配置
const CONFIG = ENV_CONFIG[detectEnvironment()];

class PopupManager {
  constructor() {
    this.currentTab = null;
    this.isLoggedIn = false;
    this.resumeData = null;
    this.siteInfo = null;

    this.init();
  }

  async init() {
    // 获取当前标签页
    await this.getCurrentTab();

    // 绑定事件监听
    this.bindEvents();

    // 检查登录状态
    await this.checkLoginStatus();

    // 如果已登录，加载数据
    if (this.isLoggedIn) {
      await this.loadData();
    }

    // 隐藏加载状态
    this.hideElement('loading');
  }

  async getCurrentTab() {
    return new Promise((resolve) => {
      chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
        this.currentTab = tabs[0];
        resolve();
      });
    });
  }

  bindEvents() {
    // 登录按钮
    const loginBtn = document.getElementById('loginBtn');
    loginBtn?.addEventListener('click', this.handleLogin.bind(this));

    // 填充按钮
    const fillBtn = document.getElementById('fillBtn');
    fillBtn?.addEventListener('click', this.handleFill.bind(this));

    // 同步按钮
    const syncBtn = document.getElementById('syncBtn');
    syncBtn?.addEventListener('click', this.handleSync.bind(this));
  }

  async checkLoginStatus() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({ type: 'CHECK_LOGIN' }, (response) => {
        this.isLoggedIn = response?.isLoggedIn || false;

        if (this.isLoggedIn) {
          this.showElement('mainContent');
          this.updateConnectionStatus('已连接', true);
        } else {
          this.showElement('loginPrompt');
          this.updateConnectionStatus('未登录', false);
        }

        resolve();
      });
    });
  }

  async loadData() {
    // 并行加载数据
    await Promise.all([
      this.loadResumeData(),
      this.analyzePage(),
      this.loadStats()
    ]);

    this.updateUI();
  }

  async loadResumeData() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({ type: 'GET_RESUME_DATA' }, (response) => {
        if (response?.success) {
          this.resumeData = response.data;
        }
        resolve();
      });
    });
  }

  async analyzePage() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({ type: 'ANALYZE_PAGE' }, (response) => {
        this.siteInfo = response;
        resolve();
      });
    });
  }

  async loadStats() {
    // 从本地存储加载统计数据
    const result = await chrome.storage.local.get([
      'totalFills',
      'timeSaved',
      'lastFillDate'
    ]);

    const today = new Date().toDateString();
    const isToday = result.lastFillDate === today;

    this.stats = {
      totalFills: isToday ? (result.totalFills || 0) : 0,
      timeSaved: isToday ? (result.timeSaved || 0) : 0
    };
  }

  updateUI() {
    // 更新网站信息
    if (this.siteInfo?.supported) {
      document.getElementById('currentSite').textContent = this.siteInfo.site.name;
      this.enableButton('fillBtn');
    } else {
      document.getElementById('currentSite').textContent = '不支持';
      this.disableButton('fillBtn');
    }

    // 更新简历信息
    if (this.resumeData) {
      const completeness = this.resumeData.meta?.completeness || 0;
      document.getElementById('resumeCompleteness').textContent = completeness + '%';
      document.getElementById('resumeProgress').style.width = completeness + '%';

      const lastUpdated = this.resumeData.meta?.lastUpdated;
      if (lastUpdated) {
        const date = new Date(lastUpdated);
        document.getElementById('lastSync').textContent =
          `${date.getMonth() + 1}/${date.getDate()} 同步`;
      }
    }

    // 更新统计信息
    document.getElementById('totalFills').textContent = this.stats.totalFills;
    document.getElementById('timeSaved').textContent = this.stats.timeSaved + '分钟';

    // 模拟检测字段数量（实际应该从content script获取）
    const fieldCount = this.estimateFieldCount();
    document.getElementById('fieldCount').textContent = fieldCount + '个';
  }

  estimateFieldCount() {
    if (!this.siteInfo?.supported || !this.resumeData) {
      return 0;
    }

    // 基于简历数据估算可填充字段数量
    let count = 0;

    if (this.resumeData.basic?.name) count++;
    if (this.resumeData.basic?.email) count++;
    if (this.resumeData.basic?.phone) count++;
    if (this.resumeData.basic?.city) count++;
    if (this.resumeData.intent?.position) count++;
    if (this.resumeData.intent?.salary) count++;
    if (this.resumeData.education?.length > 0) count += 2; // 学校、专业
    if (this.resumeData.experience?.length > 0) count += 2; // 公司、职位

    return Math.min(count, 12); // 限制最大显示数量
  }

  async handleLogin() {
    // 打开登录页面
    chrome.tabs.create({ url: CONFIG.FRONTEND_URL + '/login' });
    window.close();
  }

  async handleFill() {
    if (!this.resumeData) {
      this.showError('简历数据未加载');
      return;
    }

    this.disableButton('fillBtn');
    this.setButtonText('fillBtn', '填充中...');

    try {
      // 向content script发送填充请求
      await this.sendMessageToTab({ type: 'PERFORM_AUTOFILL' });

      // 更新统计数据
      await this.updateStats();

      this.showSuccess('自动填充完成');

      // 延迟关闭弹窗
      setTimeout(() => {
        window.close();
      }, 1500);

    } catch (error) {
      this.showError('填充失败: ' + error.message);
    } finally {
      this.enableButton('fillBtn');
      this.setButtonText('fillBtn', '开始自动填充');
    }
  }

  async handleSync() {
    this.disableButton('syncBtn');
    this.setButtonText('syncBtn', '同步中...');

    try {
      const response = await this.sendMessage({ type: 'SYNC_RESUME_DATA' });

      if (response.success) {
        this.resumeData = response.data;
        this.updateUI();
        this.showSuccess('数据同步成功');
      } else {
        throw new Error(response.error || '同步失败');
      }
    } catch (error) {
      this.showError('同步失败: ' + error.message);
    } finally {
      this.enableButton('syncBtn');
      this.setButtonText('syncBtn', '同步简历数据');
    }
  }

  async sendMessage(message) {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(message, resolve);
    });
  }

  async sendMessageToTab(message) {
    return new Promise((resolve, reject) => {
      if (!this.currentTab?.id) {
        reject(new Error('No active tab'));
        return;
      }

      chrome.tabs.sendMessage(this.currentTab.id, message, (response) => {
        if (chrome.runtime.lastError) {
          reject(new Error(chrome.runtime.lastError.message));
        } else {
          resolve(response);
        }
      });
    });
  }

  async updateStats() {
    this.stats.totalFills += 1;
    this.stats.timeSaved += 5; // 假设每次节省5分钟

    const today = new Date().toDateString();

    await chrome.storage.local.set({
      totalFills: this.stats.totalFills,
      timeSaved: this.stats.timeSaved,
      lastFillDate: today
    });

    // 更新UI
    document.getElementById('totalFills').textContent = this.stats.totalFills;
    document.getElementById('timeSaved').textContent = this.stats.timeSaved + '分钟';
  }

  showElement(elementId) {
    const element = document.getElementById(elementId);
    if (element) {
      element.style.display = 'block';
    }
  }

  hideElement(elementId) {
    const element = document.getElementById(elementId);
    if (element) {
      element.style.display = 'none';
    }
  }

  enableButton(buttonId) {
    const button = document.getElementById(buttonId);
    if (button) {
      button.disabled = false;
      button.style.opacity = '1';
    }
  }

  disableButton(buttonId) {
    const button = document.getElementById(buttonId);
    if (button) {
      button.disabled = true;
      button.style.opacity = '0.5';
    }
  }

  setButtonText(buttonId, text) {
    const button = document.getElementById(buttonId);
    if (button) {
      // 保留图标，只更改文本
      const svg = button.querySelector('svg');
      button.innerHTML = '';
      if (svg) {
        button.appendChild(svg);
      }
      button.appendChild(document.createTextNode(text));
    }
  }

  updateConnectionStatus(status, isOnline) {
    document.getElementById('connectionStatus').textContent = status;

    const dot = document.getElementById('connectionDot');
    dot.className = 'status-dot ' + (isOnline ? 'status-online' : 'status-offline');
  }

  showError(message) {
    const errorDiv = document.getElementById('errorMessage');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';

    // 3秒后自动隐藏
    setTimeout(() => {
      errorDiv.style.display = 'none';
    }, 3000);
  }

  showSuccess(message) {
    // 创建临时成功提示
    const successDiv = document.createElement('div');
    successDiv.className = 'success-message';
    successDiv.textContent = message;
    successDiv.style.cssText = `
      background: #f6ffed;
      border: 1px solid #b7eb8f;
      border-radius: 6px;
      padding: 12px;
      color: #52c41a;
      font-size: 14px;
      margin-bottom: 16px;
    `;

    const content = document.querySelector('.content');
    content.insertBefore(successDiv, content.firstChild);

    // 3秒后移除
    setTimeout(() => {
      if (successDiv.parentNode) {
        successDiv.parentNode.removeChild(successDiv);
      }
    }, 3000);
  }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
  new PopupManager();
});

// 监听来自background的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'UPDATE_UI') {
    // 可以在这里处理UI更新
  }
});