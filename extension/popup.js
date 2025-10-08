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

function getEnvironmentConfig(env) {
  return ENV_CONFIG[env] || ENV_CONFIG.production;
}

class PopupManager {
  constructor() {
    this.currentTab = null;
    this.isLoggedIn = false;
    this.resumeData = null;
    this.siteInfo = null;
    this.settings = { default_company_attribute: '' };
    this.recordData = null;
    this.enabledSites = {};
    this.currentHost = '';
    this.siteEnabled = false;
    this.stats = { totalFills: 0, timeSaved: 0 };
    this.environment = detectEnvironment();
    this.envConfig = getEnvironmentConfig(this.environment);

    this.init();
  }

  async init() {
    // 获取当前标签页
    await this.getCurrentTab();

    // 预加载站点配置
    await this.loadEnabledSites();

    // 加载环境配置
    await this.loadEnvironment();

    // 绑定事件监听
    this.bindEvents();

    // 检查登录状态
    await this.checkLoginStatus();

    // 如果已登录，加载数据
    if (this.isLoggedIn) {
      await Promise.all([this.loadData(), this.loadSettings()]);
    }

    // 隐藏加载状态
    this.hideElement('loading');
  }

  async getCurrentTab() {
    return new Promise((resolve) => {
      chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
        this.currentTab = tabs[0];
        try {
          if (this.currentTab?.url) {
            const url = new URL(this.currentTab.url);
            this.currentHost = url.hostname;
          }
        } catch (e) {
          console.warn('解析当前标签页地址失败:', e);
          this.currentHost = '';
        }
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

    // 记录投递入口
    const recordBtn = document.getElementById('recordBtn');
    recordBtn?.addEventListener('click', this.handleRecordOpen.bind(this));

    const submitRecordBtn = document.getElementById('submitRecordBtn');
    submitRecordBtn?.addEventListener('click', this.handleSubmitRecord.bind(this));

    const trainBtn = document.getElementById('trainBtn');
    trainBtn?.addEventListener('click', this.handleStartTraining.bind(this));

    const toggleSiteBtn = document.getElementById('toggleSiteBtn');
    toggleSiteBtn?.addEventListener('click', this.handleToggleSite.bind(this));

    const logoutBtn = document.getElementById('logoutBtn');
    logoutBtn?.addEventListener('click', this.handleLogout.bind(this));

    const envSelect = document.getElementById('envSelect');
    envSelect?.addEventListener('change', this.handleEnvironmentChange.bind(this));
  }

  async checkLoginStatus() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({ type: 'CHECK_LOGIN' }, (response) => {
        this.isLoggedIn = response?.isLoggedIn || false;

        if (this.isLoggedIn) {
          this.showElement('mainContent');
          this.hideElement('loginPrompt');
          this.showLogoutButton();
          this.updateConnectionStatus('已连接', true);
        } else {
          this.hideElement('mainContent');
          this.showElement('loginPrompt');
          this.hideLogoutButton();
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
      // 未在白名单中：开启“实验性填充”（按需注入）
      document.getElementById('currentSite').textContent = '实验性填充';
      this.enableButton('fillBtn');
    }

    this.siteEnabled = this.isSiteEnabled(this.currentHost);
    this.updateSiteControl();

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

    // 设置默认企业属性与日期
    const attrSel = document.getElementById('recordAttr');
    if (attrSel && this.settings.default_company_attribute) {
      attrSel.value = this.settings.default_company_attribute;
    }
    const dateInput = document.getElementById('recordDate');
    if (dateInput) {
      const today = new Date();
      const yyyy = today.getFullYear();
      const mm = String(today.getMonth() + 1).padStart(2, '0');
      const dd = String(today.getDate()).padStart(2, '0');
      dateInput.value = `${yyyy}-${mm}-${dd}`;
    }
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
    chrome.tabs.create({ url: this.envConfig.FRONTEND_URL + '/login' });
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

  async loadSettings() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({ type: 'GET_SETTINGS' }, (res) => {
        if (res?.success) {
          this.settings = { default_company_attribute: '', ...(res.data || {}) };
        }
        resolve();
      });
    });
  }

  async loadEnabledSites() {
    try {
      const res = await this.sendMessage({ type: 'GET_ENABLED_SITES' });
      if (res?.success) {
        this.enabledSites = res.data || {};
      }
    } catch (e) {
      console.warn('加载站点配置失败:', e);
    }
    this.siteEnabled = this.isSiteEnabled(this.currentHost);
    this.updateSiteControl();
  }

  async loadEnvironment() {
    try {
      const res = await this.sendMessage({ type: 'GET_ENVIRONMENT' });
      if (res?.success) {
        this.environment = res.environment || this.environment;
        this.envConfig = res.config || getEnvironmentConfig(this.environment);
      }
    } catch (e) {
      console.warn('读取环境配置失败:', e);
      this.environment = this.environment || 'production';
      this.envConfig = getEnvironmentConfig(this.environment);
    }

    const select = document.getElementById('envSelect');
    if (select) {
      select.value = this.environment;
    }
  }

  isSiteEnabled(host) {
    if (!host) return false;
    const cleanHost = sanitizeHost(host);
    if (!cleanHost) return false;
    if (this.enabledSites[cleanHost]?.enabled) return true;
    const parts = cleanHost.split('.');
    for (let i = 0; i < parts.length - 1; i++) {
      const suffix = parts.slice(i).join('.');
      const wildcard = `*.${suffix}`;
      if (this.enabledSites[wildcard]?.enabled) {
        return true;
      }
    }
    return false;
  }

  updateSiteControl() {
    const hostEl = document.getElementById('siteHost');
    if (hostEl) {
      hostEl.textContent = this.currentHost || '未知';
    }
    const statusEl = document.getElementById('siteStatus');
    const toggleBtn = document.getElementById('toggleSiteBtn');
    const hintEl = document.getElementById('siteHint');
    if (!statusEl || !toggleBtn) return;

    if (this.siteEnabled) {
      statusEl.textContent = '已启用';
      statusEl.style.color = '#52c41a';
      toggleBtn.textContent = '停用此站点监听';
      toggleBtn.classList.remove('btn-secondary');
      toggleBtn.classList.add('btn-primary');
      if (hintEl) hintEl.textContent = '已启用。点击“投递/申请”按钮后，会提示您保存岗位信息。';
    } else {
      statusEl.textContent = '未启用';
      statusEl.style.color = '#ff4d4f';
      toggleBtn.textContent = '启用此站点监听';
      toggleBtn.classList.add('btn-secondary');
      toggleBtn.classList.remove('btn-primary');
      if (hintEl) hintEl.textContent = '启用后，当您点击该网站的“投递/申请”按钮时，插件会提示保存岗位信息。';
    }

    toggleBtn.disabled = !this.currentHost;
  }

  async handleToggleSite() {
    if (!this.currentHost) return;
    const target = !this.siteEnabled;
    try {
      const res = await this.sendMessage({
        type: 'SET_SITE_ENABLED',
        host: this.currentHost,
        enabled: target,
        tabId: this.currentTab?.id
      });
      if (res?.success) {
        this.enabledSites = res.data || {};
        this.siteEnabled = target;
        this.updateSiteControl();
        this.showSuccess(target ? '已启用当前站点监听' : '已停用当前站点监听');
      } else {
        throw new Error(res?.error || '操作失败');
      }
    } catch (e) {
      this.showError(e?.message || '无法更新站点状态');
    }
  }

  async handleEnvironmentChange(event) {
    const targetEnv = event?.target?.value;
    if (!targetEnv || targetEnv === this.environment) {
      return;
    }

    const select = document.getElementById('envSelect');
    if (select) select.disabled = true;

    try {
      const res = await this.sendMessage({ type: 'SET_ENVIRONMENT', environment: targetEnv });
      if (!res?.success) {
        throw new Error(res?.error || '环境切换失败');
      }
      this.environment = res.environment;
      this.envConfig = res.config || getEnvironmentConfig(this.environment);
      if (select) select.value = this.environment;
      this.performLocalLogout();
      this.showSuccess(`已切换到${this.environment === 'local' ? '本地环境' : '线上环境'}，请重新登录`);
    } catch (err) {
      this.showError(err?.message || '环境切换失败');
      if (select) select.value = this.environment;
    } finally {
      if (select) select.disabled = false;
    }
  }

  async handleLogout() {
    if (!this.isLoggedIn) return;
    if (!window.confirm('确定要退出登录吗？')) {
      return;
    }

    const btn = document.getElementById('logoutBtn');
    if (btn) {
      btn.disabled = true;
      btn.textContent = '退出中...';
    }

    try {
      const res = await this.sendMessage({ type: 'LOGOUT' });
      if (!res?.success) {
        throw new Error(res?.error || '退出登录失败');
      }
      this.performLocalLogout();
      this.showSuccess('已退出登录');
    } catch (e) {
      this.showError(e?.message || '退出登录失败');
    } finally {
      if (btn) {
        btn.disabled = false;
        btn.textContent = '退出登录';
      }
    }
  }

  async handleRecordOpen() {
    const form = document.getElementById('recordForm');
    if (form) form.style.display = 'block';
    // 每次打开尝试自动提取
    try { await this.handleExtract(); } catch {}
  }

  async handleExtract(e) {
    if (e) e.preventDefault();
    const tryOnce = async () => {
      const resp = await this.sendMessageToTab({ type: 'EXTRACT_JOB_POSTING' });
      if (!resp?.success) throw new Error(resp?.error || '无法提取职位信息');
      return resp.data || {};
    };

    try {
      let data = await tryOnce();
      // 若首次未拿到关键字段，等待渲染后重试最多2次
      let attempts = 0;
      while (!(data.company_name || data.position_title) && attempts < 2) {
        await new Promise(r => setTimeout(r, 600));
        data = await tryOnce();
        attempts++;
      }
      this.recordData = data;
      this.fillRecordForm(data);
    } catch (err) {
      this.showRecordError(err?.message || String(err));
    }
  }

  fillRecordForm(data) {
    const set = (id, v) => { const el = document.getElementById(id); if (el) el.value = v || ''; };
    set('recordCompany', data.company_name);
    set('recordTitle', data.position_title);
    set('recordLocation', data.location_text);
    set('recordSalary', data.salary_text);
    // 日期默认值已在 updateUI 设置
    if (this.settings.default_company_attribute) {
      const attrSel = document.getElementById('recordAttr');
      if (attrSel && !attrSel.value) attrSel.value = this.settings.default_company_attribute;
    }
  }

  async handleSubmitRecord(e) {
    if (e) e.preventDefault();
    const get = (id) => { const el = document.getElementById(id); return el ? el.value.trim() : ''; };
    const companyName = get('recordCompany');
    const positionTitle = get('recordTitle');
    const companyAttr = get('recordAttr');

    // 必填校验
    const missing = [];
    if (!companyName) missing.push('公司名称');
    if (!positionTitle) missing.push('职位名称');
    if (!companyAttr) missing.push('企业属性');
    if (missing.length) {
      this.showRecordError('请完善必填项：' + missing.join('、'));
      return;
    }

    // 先合并解析结果，再用用户输入覆盖，确保手填优先生效
    const payload = {
      ...(this.recordData || {}),
      company_name: companyName,
      position_title: positionTitle,
      application_date: get('recordDate'),
      company_attribute: companyAttr,
      work_location: get('recordLocation') || undefined,
      salary_range: get('recordSalary') || undefined,
      notes: get('recordNotes') || undefined
    };

    // 记住默认企业属性
    const remember = document.getElementById('rememberAttr');
    if (remember && remember.checked && payload.company_attribute) {
      await this.sendMessage({ type: 'SAVE_SETTINGS', data: { default_company_attribute: payload.company_attribute } });
      this.settings.default_company_attribute = payload.company_attribute;
    }

    try {
      const res = await this.sendMessage({ type: 'CREATE_APPLICATION', payload });
      if (res?.duplicate) {
        // 再次确认
        const ok = confirm('检测到疑似重复记录。是否仍要创建？');
        if (ok) {
          const res2 = await this.sendMessage({ type: 'CREATE_APPLICATION', payload, force: true });
          if (!res2?.success) throw new Error(res2?.message || res2?.error || '提交失败');
          this.showSuccess('岗位信息已成功保存！请到官网查看！');
        } else {
          return;
        }
      } else if (res?.queued) {
        this.showSuccess('网络不可用，已加入稍后提交队列');
      } else if (res?.success) {
        this.showSuccess('岗位信息已成功保存！请到官网查看！');
      } else {
        throw new Error(res?.message || res?.error || '提交失败');
      }

      // 展示成功提示后保持窗口，便于用户查看
    } catch (err) {
      this.showRecordError(err?.message || String(err));
    }
  }

  showRecordError(msg) {
    const box = document.getElementById('recordError');
    if (box) { box.textContent = msg; box.style.display = 'block'; setTimeout(() => box.style.display = 'none', 3000); }
  }

  async handleStartTraining(e) {
    if (e) e.preventDefault();
    try {
      await this.sendMessageToTab({ type: 'START_TRAINING' });
      this.showSuccess('字段选择模式已开启：在页面点击字段完成标注');
    } catch (err) {
      this.showRecordError('无法开启字段选择模式：' + (err?.message || err));
    }
  }

  async sendMessage(message) {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(message, resolve);
    });
  }

  async sendMessageToTab(message) {
    return new Promise(async (resolve, reject) => {
      if (!this.currentTab?.id) {
        reject(new Error('No active tab'));
        return;
      }

      const trySend = () => {
        chrome.tabs.sendMessage(this.currentTab.id, message, (response) => {
          const err = chrome.runtime.lastError;
          if (err) {
            reject(new Error(err.message));
          } else {
            resolve(response);
          }
        });
      };

      // 先尝试发送；若失败且提示没有接收端，则按需注入
      chrome.tabs.sendMessage(this.currentTab.id, { type: 'PING' }, async () => {
        const err = chrome.runtime.lastError;
        if (err && /Receiving end does not exist|Could not establish connection/i.test(err.message)) {
          try {
            await this.injectContentScripts();
            // 注入后稍等再发
            setTimeout(() => trySend(), 200);
          } catch (e) {
            reject(new Error('无法注入内容脚本: ' + (e?.message || e)));
          }
        } else {
          trySend();
        }
      });
    });
  }

  async injectContentScripts() {
    if (!this.currentTab?.id) return;
    try {
      // 需要 scripting 权限
      await chrome.scripting.insertCSS({ target: { tabId: this.currentTab.id }, files: ['content.css'] });
      await chrome.scripting.executeScript({ target: { tabId: this.currentTab.id }, files: ['content.js'] });
    } catch (e) {
      throw e;
    }
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

  performLocalLogout() {
    this.isLoggedIn = false;
    this.resumeData = null;
    this.recordData = null;
    this.stats = { totalFills: 0, timeSaved: 0 };
    this.hideElement('mainContent');
    this.hideElement('recordForm');
    this.showElement('loginPrompt');
    this.hideLogoutButton();
    this.updateConnectionStatus('未登录', false);
    const fieldEl = document.getElementById('fieldCount');
    if (fieldEl) fieldEl.textContent = '-';
    const totalEl = document.getElementById('totalFills');
    if (totalEl) totalEl.textContent = '0';
    const timeEl = document.getElementById('timeSaved');
    if (timeEl) timeEl.textContent = '0分钟';
    const completeness = document.getElementById('resumeCompleteness');
    if (completeness) completeness.textContent = '0%';
    const progress = document.getElementById('resumeProgress');
    if (progress) progress.style.width = '0%';
    const lastSync = document.getElementById('lastSync');
    if (lastSync) lastSync.textContent = '未同步';
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

  showLogoutButton() {
    const btn = document.getElementById('logoutBtn');
    if (btn) btn.style.display = 'block';
  }

  hideLogoutButton() {
    const btn = document.getElementById('logoutBtn');
    if (btn) btn.style.display = 'none';
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
    // 若已有提示，先移除
    const old = document.getElementById('jobview-success-toast');
    if (old && old.parentNode) old.parentNode.removeChild(old);

    const toast = document.createElement('div');
    toast.id = 'jobview-success-toast';
    toast.style.cssText = `
      position: fixed; left: 12px; right: 12px; bottom: 12px; z-index: 2147483647;
      background: #f6ffed; border: 1px solid #b7eb8f; border-left: 4px solid #52c41a;
      border-radius: 8px; padding: 10px 12px; box-shadow: 0 6px 16px rgba(0,0,0,0.15);
      display: flex; align-items: center; justify-content: space-between; gap: 8px;
      color: #135200; font-size: 13px;
    `;

    const text = document.createElement('div');
    text.textContent = message;
    text.style.cssText = 'flex:1; padding-right:8px;';

    const btn = document.createElement('button');
    btn.textContent = '查看官网';
    btn.style.cssText = `
      background:#1890ff; color:#fff; border:none; border-radius:6px; padding:6px 10px; cursor:pointer;
      font-size:12px; white-space:nowrap;
    `;
    btn.addEventListener('click', (e) => {
      e.preventDefault();
      try { chrome.tabs.create({ url: this.envConfig.FRONTEND_URL }); } catch {}
    });

    toast.appendChild(text);
    toast.appendChild(btn);
    document.body.appendChild(toast);

    // 5秒后自动移除
    setTimeout(() => { if (toast && toast.parentNode) toast.parentNode.removeChild(toast); }, 5000);
  }
}

function sanitizeHost(host) {
  if (!host) return '';
  return String(host).trim().toLowerCase().replace(/^https?:\/\//, '').split('/')[0];
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
  // 暴露实例，便于接收 background 的 UI 更新通知时主动刷新
  const instance = new PopupManager();
  window.__jobviewPopup = instance;
});

// 监听来自background的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'UPDATE_UI') {
    try {
      const inst = window.__jobviewPopup;
      if (inst && typeof inst.checkLoginStatus === 'function') {
        inst.checkLoginStatus().then(() => {
          if (inst.isLoggedIn) {
            inst.loadData();
          }
        });
      }
    } catch (e) {
      // 忽略
    }
  }
});
