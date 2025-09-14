# JobView 简历自动填充功能 - 技术方案

> 文档版本：v1.0
> 创建日期：2024年9月14日
> 技术负责人：JobView 技术团队

## 🏗️ 技术架构概览

### 整体架构
```
┌─────────────────────────────────────────────────────┐
│                   用户浏览器                         │
├─────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────┐   │
│  │          Chrome Extension                    │   │
│  ├───────────────┬──────────────┬──────────────┤   │
│  │  Background   │  Content     │   Popup      │   │
│  │   Service     │  Scripts     │   UI         │   │
│  └───────┬───────┴──────┬───────┴──────┬───────┘   │
│          │              │              │           │
│          ▼              ▼              ▼           │
│  ┌──────────────────────────────────────────────┐  │
│  │           Extension Storage API              │  │
│  └──────────────────────────────────────────────┘  │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
         ┌─────────────────────────┐
         │   JobView Backend API   │
         └─────────────────────────┘
```

## 🔧 技术选型

### 核心技术栈

#### 浏览器插件开发
- **Manifest V3**: Chrome Extension 最新规范
- **TypeScript**: 类型安全的开发语言
- **React**: Popup UI 开发框架
- **Webpack**: 模块打包工具
- **Chrome APIs**:
  - Storage API: 数据存储
  - Tabs API: 标签页管理
  - Runtime API: 消息通信
  - Identity API: 用户认证

#### 数据处理
- **JSON Schema**: 数据格式验证
- **CryptoJS**: 客户端加密
- **DOMPurify**: XSS 防护
- **Ajv**: JSON Schema 验证器

#### 开发工具
- **Vite**: 快速构建工具
- **ESLint**: 代码规范检查
- **Prettier**: 代码格式化
- **Jest**: 单元测试框架
- **Puppeteer**: E2E 测试

## 📦 插件架构设计

### 1. Background Service Worker
```typescript
// background.ts
class BackgroundService {
  private jobViewAPI: JobViewAPIClient;
  private dataCache: DataCache;
  private syncManager: SyncManager;

  constructor() {
    this.initializeServices();
    this.registerListeners();
  }

  // 初始化服务
  private initializeServices() {
    this.jobViewAPI = new JobViewAPIClient();
    this.dataCache = new DataCache();
    this.syncManager = new SyncManager();
  }

  // 注册消息监听
  private registerListeners() {
    chrome.runtime.onMessage.addListener(this.handleMessage);
    chrome.runtime.onInstalled.addListener(this.handleInstall);
    chrome.alarms.onAlarm.addListener(this.handleAlarm);
  }

  // 处理消息
  private handleMessage = (request, sender, sendResponse) => {
    switch (request.type) {
      case 'FETCH_RESUME_DATA':
        return this.fetchResumeData(sendResponse);
      case 'SYNC_DATA':
        return this.syncData(sendResponse);
      case 'ANALYZE_PAGE':
        return this.analyzePage(sender.tab, sendResponse);
    }
  };

  // 获取简历数据
  private async fetchResumeData(sendResponse) {
    try {
      const token = await this.getAuthToken();
      const data = await this.jobViewAPI.getResumeData(token);
      await this.dataCache.store(data);
      sendResponse({ success: true, data });
    } catch (error) {
      sendResponse({ success: false, error: error.message });
    }
  }
}
```

### 2. Content Scripts
```typescript
// content-script.ts
class ContentScriptManager {
  private formAnalyzer: FormAnalyzer;
  private autoFiller: AutoFiller;
  private uiController: UIController;

  constructor() {
    this.initialize();
  }

  private async initialize() {
    // 初始化组件
    this.formAnalyzer = new FormAnalyzer();
    this.autoFiller = new AutoFiller();
    this.uiController = new UIController();

    // 检测当前页面
    const pageInfo = await this.detectPage();
    if (pageInfo.isSupported) {
      this.activateAutoFill(pageInfo);
    }
  }

  // 页面检测
  private async detectPage(): Promise<PageInfo> {
    const url = window.location.href;
    const domain = new URL(url).hostname;

    // 匹配支持的网站
    const siteConfig = SUPPORTED_SITES[domain];
    if (!siteConfig) {
      return { isSupported: false };
    }

    // 检测表单
    const forms = this.formAnalyzer.detectForms();
    return {
      isSupported: true,
      site: siteConfig,
      forms: forms
    };
  }

  // 激活自动填充
  private activateAutoFill(pageInfo: PageInfo) {
    // 显示填充按钮
    this.uiController.showFloatingButton();

    // 监听用户操作
    this.uiController.onFillRequest(() => {
      this.performAutoFill(pageInfo);
    });
  }

  // 执行自动填充
  private async performAutoFill(pageInfo: PageInfo) {
    // 获取用户数据
    const userData = await this.getUserData();

    // 映射字段
    const mappedData = this.mapFields(userData, pageInfo.site);

    // 填充表单
    const results = await this.autoFiller.fill(
      pageInfo.forms,
      mappedData
    );

    // 显示结果
    this.uiController.showResults(results);
  }
}
```

### 3. Form Analyzer（表单分析器）
```typescript
// form-analyzer.ts
class FormAnalyzer {
  // 表单检测策略
  private strategies: DetectionStrategy[] = [
    new InputDetectionStrategy(),
    new SelectDetectionStrategy(),
    new TextareaDetectionStrategy(),
    new RadioDetectionStrategy(),
    new CheckboxDetectionStrategy()
  ];

  // 检测所有表单
  detectForms(): FormField[] {
    const fields: FormField[] = [];

    // 遍历所有策略
    this.strategies.forEach(strategy => {
      const detectedFields = strategy.detect();
      fields.push(...detectedFields);
    });

    // 智能识别字段类型
    return fields.map(field => this.identifyField(field));
  }

  // 识别字段类型
  private identifyField(field: FormField): FormField {
    const identifiers = {
      // 基础信息
      name: ['name', '姓名', 'fullname', 'username'],
      email: ['email', '邮箱', 'mail'],
      phone: ['phone', 'mobile', '电话', '手机'],

      // 教育信息
      school: ['school', 'university', '学校', '院校'],
      major: ['major', 'subject', '专业'],
      degree: ['degree', 'education', '学历'],

      // 工作信息
      company: ['company', 'employer', '公司', '单位'],
      position: ['position', 'title', 'job', '职位'],
      salary: ['salary', 'wage', '薪资', '工资']
    };

    // 匹配字段
    for (const [type, keywords] of Object.entries(identifiers)) {
      if (this.matchesKeywords(field, keywords)) {
        field.type = type;
        field.confidence = this.calculateConfidence(field, keywords);
        break;
      }
    }

    return field;
  }

  // 关键词匹配
  private matchesKeywords(field: FormField, keywords: string[]): boolean {
    const fieldText = this.getFieldText(field).toLowerCase();
    return keywords.some(keyword =>
      fieldText.includes(keyword.toLowerCase())
    );
  }

  // 获取字段文本
  private getFieldText(field: FormField): string {
    const texts = [
      field.element.id,
      field.element.name,
      field.element.className,
      field.element.placeholder,
      field.label?.textContent
    ].filter(Boolean);

    return texts.join(' ');
  }

  // 计算置信度
  private calculateConfidence(field: FormField, keywords: string[]): number {
    const fieldText = this.getFieldText(field).toLowerCase();
    let score = 0;

    keywords.forEach(keyword => {
      if (fieldText === keyword.toLowerCase()) score += 10;
      else if (fieldText.includes(keyword.toLowerCase())) score += 5;
    });

    return Math.min(score, 100);
  }
}
```

### 4. Auto Filler（自动填充器）
```typescript
// auto-filler.ts
class AutoFiller {
  // 填充策略映射
  private fillStrategies: Map<string, FillStrategy> = new Map([
    ['input', new InputFillStrategy()],
    ['select', new SelectFillStrategy()],
    ['textarea', new TextareaFillStrategy()],
    ['radio', new RadioFillStrategy()],
    ['checkbox', new CheckboxFillStrategy()],
    ['date', new DateFillStrategy()],
    ['file', new FileFillStrategy()]
  ]);

  // 执行填充
  async fill(
    fields: FormField[],
    data: MappedData
  ): Promise<FillResult[]> {
    const results: FillResult[] = [];

    for (const field of fields) {
      try {
        const result = await this.fillField(field, data);
        results.push(result);
      } catch (error) {
        results.push({
          field: field,
          success: false,
          error: error.message
        });
      }
    }

    return results;
  }

  // 填充单个字段
  private async fillField(
    field: FormField,
    data: MappedData
  ): Promise<FillResult> {
    // 获取对应数据
    const value = data[field.type];
    if (!value) {
      return {
        field: field,
        success: false,
        error: 'No data for field'
      };
    }

    // 获取填充策略
    const strategy = this.fillStrategies.get(field.element.tagName.toLowerCase());
    if (!strategy) {
      return {
        field: field,
        success: false,
        error: 'No strategy for element type'
      };
    }

    // 执行填充
    await strategy.fill(field.element, value);

    // 触发事件
    this.triggerEvents(field.element);

    return {
      field: field,
      success: true,
      value: value
    };
  }

  // 触发必要的事件
  private triggerEvents(element: HTMLElement) {
    // 触发 input 事件
    element.dispatchEvent(new Event('input', { bubbles: true }));

    // 触发 change 事件
    element.dispatchEvent(new Event('change', { bubbles: true }));

    // 触发 blur 事件（某些网站需要）
    element.dispatchEvent(new Event('blur', { bubbles: true }));

    // React 网站特殊处理
    if (this.isReactSite()) {
      this.triggerReactChange(element);
    }

    // Vue 网站特殊处理
    if (this.isVueSite()) {
      this.triggerVueChange(element);
    }
  }

  // React 事件触发
  private triggerReactChange(element: HTMLElement) {
    const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype,
      'value'
    )?.set;

    if (nativeInputValueSetter) {
      nativeInputValueSetter.call(element, (element as HTMLInputElement).value);
    }

    const event = new Event('input', { bubbles: true });
    element.dispatchEvent(event);
  }
}
```

### 5. Popup UI
```tsx
// popup.tsx
import React, { useState, useEffect } from 'react';

const PopupApp: React.FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [resumeData, setResumeData] = useState(null);
  const [autoFillEnabled, setAutoFillEnabled] = useState(true);
  const [stats, setStats] = useState({
    totalFills: 0,
    sitesSupported: 10,
    timeSaved: '2.5小时'
  });

  useEffect(() => {
    checkLoginStatus();
    loadResumeData();
    loadStats();
  }, []);

  // 检查登录状态
  const checkLoginStatus = async () => {
    const response = await chrome.runtime.sendMessage({
      type: 'CHECK_LOGIN'
    });
    setIsLoggedIn(response.isLoggedIn);
  };

  // 加载简历数据
  const loadResumeData = async () => {
    const response = await chrome.runtime.sendMessage({
      type: 'GET_RESUME_DATA'
    });
    setResumeData(response.data);
  };

  // 同步数据
  const handleSync = async () => {
    const response = await chrome.runtime.sendMessage({
      type: 'SYNC_DATA'
    });
    if (response.success) {
      message.success('数据同步成功');
      loadResumeData();
    }
  };

  return (
    <div className="popup-container">
      <header className="popup-header">
        <img src="/logo.png" alt="JobView" />
        <h1>JobView 自动填充</h1>
      </header>

      {!isLoggedIn ? (
        <LoginPrompt />
      ) : (
        <>
          <StatusCard
            enabled={autoFillEnabled}
            onToggle={setAutoFillEnabled}
          />

          <ResumeInfo data={resumeData} />

          <StatsDisplay stats={stats} />

          <ActionButtons onSync={handleSync} />
        </>
      )}
    </div>
  );
};
```

## 🔐 数据安全方案

### 1. 数据加密存储
```typescript
// encryption.ts
class DataEncryption {
  private algorithm = 'AES-GCM';
  private keyLength = 256;

  // 生成密钥
  async generateKey(): Promise<CryptoKey> {
    return await crypto.subtle.generateKey(
      {
        name: this.algorithm,
        length: this.keyLength
      },
      true,
      ['encrypt', 'decrypt']
    );
  }

  // 加密数据
  async encrypt(data: any, key: CryptoKey): Promise<EncryptedData> {
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encoded = new TextEncoder().encode(JSON.stringify(data));

    const encrypted = await crypto.subtle.encrypt(
      {
        name: this.algorithm,
        iv: iv
      },
      key,
      encoded
    );

    return {
      data: Array.from(new Uint8Array(encrypted)),
      iv: Array.from(iv)
    };
  }

  // 解密数据
  async decrypt(encryptedData: EncryptedData, key: CryptoKey): Promise<any> {
    const decrypted = await crypto.subtle.decrypt(
      {
        name: this.algorithm,
        iv: new Uint8Array(encryptedData.iv)
      },
      key,
      new Uint8Array(encryptedData.data)
    );

    const decoded = new TextDecoder().decode(decrypted);
    return JSON.parse(decoded);
  }
}
```

### 2. 敏感信息处理
```typescript
// privacy.ts
class PrivacyManager {
  // 敏感字段定义
  private sensitiveFields = [
    'idCard',
    'bankAccount',
    'socialSecurity',
    'passport'
  ];

  // 脱敏处理
  maskSensitiveData(data: any): any {
    const masked = { ...data };

    this.sensitiveFields.forEach(field => {
      if (masked[field]) {
        masked[field] = this.mask(masked[field]);
      }
    });

    return masked;
  }

  // 脱敏算法
  private mask(value: string): string {
    if (value.length <= 4) return '****';

    const showLength = Math.min(4, Math.floor(value.length / 3));
    const hideLength = value.length - showLength * 2;

    return value.substring(0, showLength) +
           '*'.repeat(hideLength) +
           value.substring(value.length - showLength);
  }

  // 用户授权检查
  async checkPermission(field: string): Promise<boolean> {
    if (!this.sensitiveFields.includes(field)) {
      return true;
    }

    // 请求用户授权
    return await this.requestUserPermission(field);
  }
}
```

## 🌐 网站适配方案

### 1. 网站配置管理
```typescript
// site-config.ts
interface SiteConfig {
  domain: string;
  name: string;
  selectors: {
    forms: string[];
    fields: Record<string, string>;
  };
  mappings: Record<string, string>;
  special: {
    requiresDelay?: number;
    usesReact?: boolean;
    usesVue?: boolean;
    dynamicLoad?: boolean;
  };
}

const SITE_CONFIGS: Record<string, SiteConfig> = {
  'zhaopin.com': {
    domain: 'zhaopin.com',
    name: '智联招聘',
    selectors: {
      forms: ['#resumeForm', '.resume-form'],
      fields: {
        name: 'input[name="name"], #userName',
        email: 'input[type="email"], #email',
        phone: 'input[name="mobile"], #mobile'
      }
    },
    mappings: {
      'full_name': 'name',
      'email': 'email',
      'phone': 'mobile'
    },
    special: {
      usesVue: true,
      requiresDelay: 500
    }
  },

  '51job.com': {
    domain: '51job.com',
    name: '前程无忧',
    selectors: {
      forms: ['#BaseInfo', '.baseInfo'],
      fields: {
        name: '#Name',
        email: '#Email',
        phone: '#Mobile'
      }
    },
    mappings: {
      'full_name': 'name',
      'email': 'email',
      'phone': 'phone'
    },
    special: {
      requiresDelay: 300
    }
  }
};
```

### 2. 动态适配器
```typescript
// adapter.ts
class SiteAdapter {
  private config: SiteConfig;
  private observer: MutationObserver;

  constructor(config: SiteConfig) {
    this.config = config;
    this.initObserver();
  }

  // 初始化观察器
  private initObserver() {
    this.observer = new MutationObserver((mutations) => {
      this.handleDOMChanges(mutations);
    });

    this.observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true
    });
  }

  // 处理 DOM 变化
  private handleDOMChanges(mutations: MutationRecord[]) {
    // 检测新增表单
    for (const mutation of mutations) {
      if (mutation.type === 'childList') {
        this.detectNewForms(mutation.addedNodes);
      }
    }
  }

  // 查找表单元素
  findFormElements(): FormElement[] {
    const elements: FormElement[] = [];

    // 使用配置的选择器
    for (const selector of this.config.selectors.forms) {
      const forms = document.querySelectorAll(selector);
      forms.forEach(form => {
        const fields = this.findFields(form);
        elements.push({ form, fields });
      });
    }

    return elements;
  }

  // 字段映射
  mapField(jobViewField: string): string {
    return this.config.mappings[jobViewField] || jobViewField;
  }
}
```

## 🧪 测试策略

### 1. 单元测试
```typescript
// tests/form-analyzer.test.ts
describe('FormAnalyzer', () => {
  let analyzer: FormAnalyzer;

  beforeEach(() => {
    analyzer = new FormAnalyzer();
  });

  test('should detect input fields', () => {
    document.body.innerHTML = `
      <input type="text" name="username" />
      <input type="email" name="email" />
    `;

    const fields = analyzer.detectForms();
    expect(fields).toHaveLength(2);
    expect(fields[0].type).toBe('name');
    expect(fields[1].type).toBe('email');
  });

  test('should calculate confidence correctly', () => {
    const field = {
      element: document.createElement('input'),
      label: { textContent: '姓名' }
    };

    const identified = analyzer.identifyField(field);
    expect(identified.confidence).toBeGreaterThan(80);
  });
});
```

### 2. 集成测试
```typescript
// tests/integration.test.ts
describe('AutoFill Integration', () => {
  let page: Page;

  beforeAll(async () => {
    const browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  test('should fill form on zhaopin.com', async () => {
    await page.goto('https://zhaopin.com/resume');

    // 加载插件
    await page.evaluate(() => {
      // 模拟插件注入
    });

    // 触发自动填充
    await page.click('#autofill-button');

    // 验证填充结果
    const name = await page.$eval('#userName', el => el.value);
    expect(name).toBe('张三');
  });
});
```

## 📊 性能优化

### 1. 缓存策略
```typescript
class CacheManager {
  private cache: Map<string, CacheEntry> = new Map();
  private maxAge = 3600000; // 1小时

  async get(key: string): Promise<any> {
    const entry = this.cache.get(key);

    if (!entry) return null;

    if (Date.now() - entry.timestamp > this.maxAge) {
      this.cache.delete(key);
      return null;
    }

    return entry.data;
  }

  async set(key: string, data: any): Promise<void> {
    this.cache.set(key, {
      data: data,
      timestamp: Date.now()
    });
  }
}
```

### 2. 懒加载
```typescript
class LazyLoader {
  private modules: Map<string, any> = new Map();

  async load(moduleName: string): Promise<any> {
    if (this.modules.has(moduleName)) {
      return this.modules.get(moduleName);
    }

    const module = await import(`./modules/${moduleName}`);
    this.modules.set(moduleName, module);
    return module;
  }
}
```

## 🚀 部署方案

### 1. Chrome Web Store 发布
```json
// manifest.json
{
  "manifest_version": 3,
  "name": "JobView 简历自动填充助手",
  "version": "1.0.0",
  "description": "一键填充简历信息到各大招聘网站",
  "permissions": [
    "storage",
    "tabs",
    "activeTab"
  ],
  "host_permissions": [
    "https://*.zhaopin.com/*",
    "https://*.51job.com/*",
    "https://*.zhipin.com/*",
    "https://*.lagou.com/*",
    "https://*.liepin.com/*"
  ],
  "background": {
    "service_worker": "background.js"
  },
  "content_scripts": [
    {
      "matches": ["<all_urls>"],
      "js": ["content.js"],
      "css": ["content.css"],
      "run_at": "document_idle"
    }
  ],
  "action": {
    "default_popup": "popup.html",
    "default_icon": {
      "16": "icons/icon16.png",
      "48": "icons/icon48.png",
      "128": "icons/icon128.png"
    }
  }
}
```

### 2. 自动更新机制
```typescript
class AutoUpdater {
  private updateUrl = 'https://api.jobview.com/extension/update';

  async checkForUpdates(): Promise<UpdateInfo> {
    const currentVersion = chrome.runtime.getManifest().version;

    const response = await fetch(`${this.updateUrl}?v=${currentVersion}`);
    const updateInfo = await response.json();

    if (updateInfo.hasUpdate) {
      this.notifyUser(updateInfo);
    }

    return updateInfo;
  }

  private notifyUser(updateInfo: UpdateInfo) {
    chrome.notifications.create({
      type: 'basic',
      iconUrl: 'icons/icon128.png',
      title: '插件更新可用',
      message: `新版本 ${updateInfo.version} 已发布，点击更新`,
      buttons: [{ title: '立即更新' }]
    });
  }
}
```

## 📝 技术总结

通过 Chrome Extension 技术方案，我们可以：

1. **安全可靠**：数据本地加密存储，用户完全控制
2. **兼容性强**：支持各种技术栈的网站
3. **用户友好**：非侵入式设计，操作简单
4. **易于维护**：模块化架构，便于扩展
5. **性能优秀**：缓存优化，响应快速

该方案能够有效解决用户重复填写简历的痛点，显著提升求职效率。