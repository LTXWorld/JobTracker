/**
 * JobView AutoFill Extension - Content Script
 * 负责页面表单识别、自动填充和用户交互
 */

// 当用户在 JobView 网站完成登录后，通过桥接把 token 传给扩展后台
function initAuthBridgeIfOnJobViewSite() {
  try {
    const host = location.hostname || '';
    const isLocal = host === 'localhost' || host === '127.0.0.1';
    const isJobView = host.includes('jobview.bfsmlt.top') || isLocal;
    if (!isJobView) return false;

    // 不再注入内联脚本，直接在内容脚本中读取 Web Storage
    const trySendFromContentScript = () => {
      try {
        const accessToken = window.sessionStorage && window.sessionStorage.getItem('access_token');
        const refreshToken = window.localStorage && window.localStorage.getItem('refresh_token');
        if (accessToken && refreshToken) {
          chrome.runtime.sendMessage(
            { type: 'SAVE_TOKENS', accessToken, refreshToken },
            () => {
              try {
                const note = document.createElement('div');
                note.textContent = 'JobView 已连接扩展，数据同步中...';
                Object.assign(note.style, {
                  position: 'fixed', right: '20px', bottom: '20px',
                  background: 'rgba(24,144,255,0.95)', color: '#fff',
                  padding: '10px 12px', borderRadius: '6px', zIndex: 999999
                });
                document.body.appendChild(note);
                setTimeout(() => note.remove(), 2500);
              } catch {}
            }
          );
          return true;
        }
      } catch {}
      return false;
    };

    // 立即尝试一次
    if (!trySendFromContentScript()) {
      // 轮询等待登录完成（最多10秒）
      let attempts = 0;
      const timer = setInterval(() => {
        attempts++;
        if (trySendFromContentScript() || attempts > 40) {
          clearInterval(timer);
        }
      }, 250);
    }

    return true;
  } catch {
    return false;
  }
}

class FormFieldDetector {
  constructor() {
    this.fieldPatterns = {
      // 基础信息字段模式
      name: [
        'name', '姓名', 'fullname', 'realname', 'username',
        '真实姓名', '用户姓名', '联系人', 'contact_name'
      ],
      email: [
        'email', 'mail', '邮箱', '电子邮件', 'e-mail',
        '邮件地址', '电邮', 'email_address'
      ],
      phone: [
        'phone', 'mobile', 'tel', '电话', '手机', '联系电话',
        '手机号', '电话号码', 'phone_number', 'mobile_number'
      ],
      city: [
        'city', 'location', '城市', '所在地', '居住地',
        '工作城市', '现居城市', 'work_city', 'live_city'
      ],

      // 教育信息字段模式
      school: [
        'school', 'university', 'college', '学校', '院校',
        '毕业院校', '就读学校', 'education_school'
      ],
      major: [
        'major', 'specialty', 'subject', '专业', '学科',
        '所学专业', '专业名称', 'study_major'
      ],
      degree: [
        'degree', 'education', '学历', '学位', '教育程度',
        '最高学历', 'education_level'
      ],

      // 工作信息字段模式
      company: [
        'company', 'corporation', 'employer', '公司', '单位',
        '工作单位', '雇主', 'work_company', 'company_name'
      ],
      position: [
        'position', 'job', 'title', '职位', '岗位',
        '工作职位', '职务', 'job_title', 'work_position'
      ],
      salary: [
        'salary', 'wage', 'pay', '薪资', '工资', '薪水',
        '期望薪资', '薪酬', 'expected_salary'
      ]
    };
  }

  // 检测页面中的所有表单字段
  detectAllFields() {
    const fields = [];
    const inputs = document.querySelectorAll('input, select, textarea');

    inputs.forEach(element => {
      const field = this.analyzeField(element);
      if (field) {
        fields.push(field);
      }
    });

    return fields;
  }

  // 分析单个字段
  analyzeField(element) {
    // 跳过隐藏字段和特殊类型
    if (element.type === 'hidden' || element.type === 'submit' ||
        element.type === 'button' || element.style.display === 'none') {
      return null;
    }

    const fieldInfo = this.extractFieldInfo(element);
    const fieldType = this.identifyFieldType(fieldInfo);

    if (!fieldType) {
      return null;
    }

    return {
      element: element,
      type: fieldType,
      info: fieldInfo,
      confidence: this.calculateConfidence(fieldInfo, fieldType),
      xpath: this.getXPath(element)
    };
  }

  // 提取字段信息
  extractFieldInfo(element) {
    const info = {
      id: element.id || '',
      name: element.name || '',
      className: element.className || '',
      placeholder: element.placeholder || '',
      type: element.type || element.tagName.toLowerCase(),
      label: '',
      parentText: ''
    };

    // 查找关联的label
    const label = this.findLabel(element);
    if (label) {
      info.label = label.textContent.trim();
    }

    // 获取父元素文本
    const parent = element.closest('div, td, li, span');
    if (parent) {
      info.parentText = parent.textContent.replace(element.value || '', '').trim();
    }

    return info;
  }

  // 查找字段的标签
  findLabel(element) {
    // 方式1: 通过for属性关联
    if (element.id) {
      const label = document.querySelector(`label[for="${element.id}"]`);
      if (label) return label;
    }

    // 方式2: 查找父级label
    const parentLabel = element.closest('label');
    if (parentLabel) return parentLabel;

    // 方式3: 查找前面的兄弟元素
    let sibling = element.previousElementSibling;
    while (sibling) {
      if (sibling.tagName === 'LABEL') {
        return sibling;
      }
      if (sibling.tagName === 'SPAN' || sibling.tagName === 'DIV') {
        const text = sibling.textContent.trim();
        if (text && text.length < 50) {
          return sibling;
        }
      }
      sibling = sibling.previousElementSibling;
    }

    return null;
  }

  // 识别字段类型
  identifyFieldType(fieldInfo) {
    const searchText = [
      fieldInfo.id,
      fieldInfo.name,
      fieldInfo.className,
      fieldInfo.placeholder,
      fieldInfo.label,
      fieldInfo.parentText
    ].join(' ').toLowerCase();

    let bestMatch = null;
    let highestScore = 0;

    for (const [type, patterns] of Object.entries(this.fieldPatterns)) {
      let score = 0;

      for (const pattern of patterns) {
        const regex = new RegExp(pattern.toLowerCase(), 'gi');
        const matches = searchText.match(regex);
        if (matches) {
          score += matches.length * pattern.length;
        }
      }

      if (score > highestScore) {
        highestScore = score;
        bestMatch = type;
      }
    }

    return highestScore > 0 ? bestMatch : null;
  }

  // 计算匹配置信度
  calculateConfidence(fieldInfo, fieldType) {
    const patterns = this.fieldPatterns[fieldType] || [];
    const searchText = [
      fieldInfo.id,
      fieldInfo.name,
      fieldInfo.label,
      fieldInfo.placeholder
    ].join(' ').toLowerCase();

    let maxScore = 0;
    for (const pattern of patterns) {
      if (searchText === pattern.toLowerCase()) {
        maxScore = 100;
        break;
      } else if (searchText.includes(pattern.toLowerCase())) {
        maxScore = Math.max(maxScore, 80);
      } else if (new RegExp(pattern.toLowerCase()).test(searchText)) {
        maxScore = Math.max(maxScore, 60);
      }
    }

    return maxScore;
  }

  // 获取元素的XPath
  getXPath(element) {
    const parts = [];
    let current = element;

    while (current && current.nodeType === Node.ELEMENT_NODE) {
      let index = 1;
      let sibling = current.previousElementSibling;

      while (sibling) {
        if (sibling.tagName === current.tagName) {
          index++;
        }
        sibling = sibling.previousElementSibling;
      }

      const tagName = current.tagName.toLowerCase();
      const part = `${tagName}[${index}]`;
      parts.unshift(part);

      current = current.parentElement;
    }

    return '/' + parts.join('/');
  }
}

class AutoFiller {
  constructor() {
    this.fillStrategies = new Map([
      ['input', this.fillInput.bind(this)],
      ['select', this.fillSelect.bind(this)],
      ['textarea', this.fillTextarea.bind(this)]
    ]);

    // 站点特定策略（初版：Moka 下拉）
    this.siteStrategies = {
      selectClick: async (element, value) => {
        try {
          // 打开下拉
          element.click();
          element.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
          await new Promise(r => setTimeout(r, 120));

          // 在常见的下拉容器中查找匹配项（AntD、Moka 等）
          const containers = Array.from(document.querySelectorAll('[role="listbox"], .ant-select-dropdown, .select-dropdown, .dropdown, .ant-cascader-menus, ul'));
          const norm = (t) => String(t || '').replace(/\s+/g, '').toLowerCase();
          const target = norm(value);

          let candidate = null;
          for (const c of containers) {
            const items = c.querySelectorAll('li, .ant-select-item, .ant-cascader-menu-item, [role="option"], div');
            for (const it of items) {
              const txt = norm(it.textContent);
              if (!txt) continue;
              if (txt.includes(target) || target.includes(txt)) {
                candidate = it;
                break;
              }
            }
            if (candidate) break;
          }

          if (candidate) {
            candidate.dispatchEvent(new MouseEvent('mouseenter', { bubbles: true }));
            candidate.click();
            await new Promise(r => setTimeout(r, 80));
            return true;
          }
        } catch (e) {
          console.warn('selectClick strategy failed:', e);
        }
        return false;
      }
    };
  }

  // 执行自动填充
  async fill(fields, resumeData) {
    const results = [];

    for (const field of fields) {
      try {
        const result = await this.fillField(field, resumeData);
        results.push({
          field: field.type,
          success: result.success,
          value: result.value,
          error: result.error
        });
      } catch (error) {
        results.push({
          field: field.type,
          success: false,
          error: error.message
        });
      }
    }

    return results;
  }

  // 填充单个字段
  async fillField(field, resumeData) {
    const value = this.getValueForField(field.type, resumeData);

    if (!value) {
      return { success: false, error: 'No data available' };
    }

    const tagName = field.element.tagName.toLowerCase();
    const strategy = this.fillStrategies.get(tagName);

    if (!strategy) {
      return { success: false, error: 'Unsupported element type' };
    }

    try {
      // 如果字段被标记为下拉点击型（来自学习映射），优先使用站点策略
      if (field.fillStrategy === 'selectClick' && this.siteStrategies.selectClick) {
        const ok = await this.siteStrategies.selectClick(field.element, value);
        if (!ok) {
          // 回退尝试常规策略
          await (strategy ? strategy(field.element, value) : Promise.resolve());
        }
      } else {
        await strategy(field.element, value);
      }
      this.triggerEvents(field.element);
      return { success: true, value: value };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  // 根据字段类型获取对应的数据值
  getValueForField(fieldType, resumeData) {
    const mapping = {
      name: () => resumeData.basic?.name,
      email: () => resumeData.basic?.email,
      phone: () => resumeData.basic?.phone,
      city: () => resumeData.basic?.city || resumeData.intent?.city,
      school: () => resumeData.education?.[0]?.school,
      major: () => resumeData.education?.[0]?.major,
      degree: () => resumeData.education?.[0]?.degree,
      company: () => resumeData.experience?.[0]?.company,
      position: () => resumeData.intent?.position || resumeData.experience?.[0]?.position,
      salary: () => resumeData.intent?.salary
    };

    const getter = mapping[fieldType];
    return getter ? getter() : null;
  }

  // 填充输入框
  async fillInput(element, value) {
    // 清空现有内容
    element.focus();
    element.select();

    // 设置值
    element.value = value;

    // 模拟输入延迟
    await new Promise(resolve => setTimeout(resolve, 100));
  }

  // 填充选择框
  async fillSelect(element, value) {
    // 尝试精确匹配
    let option = Array.from(element.options).find(opt =>
      opt.value === value || opt.text === value
    );

    // 尝试模糊匹配
    if (!option) {
      option = Array.from(element.options).find(opt =>
        opt.text.includes(value) || value.includes(opt.text)
      );
    }

    if (option) {
      element.value = option.value;
      option.selected = true;
    }
  }

  // 填充文本区域
  async fillTextarea(element, value) {
    element.focus();
    element.value = value;
  }

  // 触发必要的事件
  triggerEvents(element) {
    const events = ['input', 'change', 'blur', 'keyup'];

    events.forEach(eventType => {
      const event = new Event(eventType, {
        bubbles: true,
        cancelable: true
      });
      element.dispatchEvent(event);
    });

    // 特殊处理React/Vue框架
    this.triggerFrameworkEvents(element);
  }

  // 触发框架特定事件
  triggerFrameworkEvents(element) {
    // React事件处理
    if (window.React || document.querySelector('[data-reactroot]')) {
      const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
        window.HTMLInputElement.prototype, 'value'
      )?.set;

      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(element, element.value);
        const event = new Event('input', { bubbles: true });
        element.dispatchEvent(event);
      }
    }

    // Vue事件处理
    if (window.Vue || document.querySelector('[v-model]')) {
      element.dispatchEvent(new Event('input', { bubbles: true }));
    }
  }
}

class FloatingButton {
  constructor() {
    this.button = null;
    this.tooltip = null;
    this.isVisible = false;
  }

  // 创建浮动按钮
  create() {
    if (this.button) return;

    this.button = document.createElement('div');
    this.button.id = 'jobview-autofill-button';
    this.button.innerHTML = `
      <svg width="24" height="24" viewBox="0 0 24 24" fill="white">
        <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zM9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z"/>
      </svg>
    `;

    // 设置样式
    Object.assign(this.button.style, {
      position: 'fixed',
      top: '20px',
      right: '20px',
      width: '50px',
      height: '50px',
      backgroundColor: '#1890ff',
      borderRadius: '25px',
      cursor: 'pointer',
      zIndex: '10000',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
      transition: 'all 0.3s ease',
      opacity: '0.9'
    });

    // 悬停效果
    this.button.addEventListener('mouseenter', () => {
      this.button.style.opacity = '1';
      this.button.style.transform = 'scale(1.1)';
      this.showTooltip();
    });

    this.button.addEventListener('mouseleave', () => {
      this.button.style.opacity = '0.9';
      this.button.style.transform = 'scale(1)';
      this.hideTooltip();
    });

    // 点击事件
    this.button.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      this.handleClick();
    });

    document.body.appendChild(this.button);
  }

  // 显示提示
  showTooltip() {
    if (this.tooltip) return;

    this.tooltip = document.createElement('div');
    this.tooltip.textContent = 'JobView 自动填充';

    Object.assign(this.tooltip.style, {
      position: 'fixed',
      top: '25px',
      right: '80px',
      backgroundColor: 'rgba(0,0,0,0.8)',
      color: 'white',
      padding: '8px 12px',
      borderRadius: '4px',
      fontSize: '14px',
      zIndex: '10001',
      whiteSpace: 'nowrap'
    });

    document.body.appendChild(this.tooltip);
  }

  // 隐藏提示
  hideTooltip() {
    if (this.tooltip) {
      document.body.removeChild(this.tooltip);
      this.tooltip = null;
    }
  }

  // 处理按钮点击
  async handleClick() {
    this.button.style.pointerEvents = 'none';
    this.button.innerHTML = `
      <svg width="24" height="24" viewBox="0 0 24 24" fill="white">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
      </svg>
    `;

    try {
      await window.jobViewAutoFill.performAutoFill();
    } catch (error) {
      console.error('Auto fill failed:', error);
    }

    setTimeout(() => {
      this.button.style.pointerEvents = 'auto';
      this.button.innerHTML = `
        <svg width="24" height="24" viewBox="0 0 24 24" fill="white">
          <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zM9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z"/>
        </svg>
      `;
    }, 2000);
  }

  // 显示按钮
  show() {
    if (!this.button) this.create();
    this.button.style.display = 'flex';
    this.isVisible = true;
  }

  // 隐藏按钮
  hide() {
    if (this.button) {
      this.button.style.display = 'none';
      this.isVisible = false;
    }
  }
}

class ContentScriptManager {
  constructor() {
    this.detector = new FormFieldDetector();
    this.autoFiller = new AutoFiller();
    this.floatingButton = new FloatingButton();
    this.trainingButton = null;
    this.trainingMode = false;
    this.mappingManager = new MappingManager();
    this.resumeData = null;
    this.detectedFields = [];

    // 将实例暴露到全局，便于消息监听器访问
    try { window.__jobViewCSM = this; } catch {}
    this.init();
  }

  async init() {
    // 检查当前页面是否支持
    const pageInfo = await this.checkPageSupport();

    if (pageInfo.supported && pageInfo.site.relevant) {
      await this.initAutoFill();
    }

    // 监听页面变化
    this.observePageChanges();

    // 无论是否为已支持站点，均提供学习模式入口（便于长尾自建站点手动映射）
    this.showTrainingToggle();
  }

  // 检查页面支持
  async checkPageSupport() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(
        { type: 'ANALYZE_PAGE' },
        (response) => resolve(response || { supported: false })
      );
    });
  }

  // 初始化自动填充功能
  async initAutoFill() {
    // 获取简历数据
    await this.loadResumeData();

    // 检测表单字段
    this.detectFormFields();

    // 如果有可填充的字段，显示浮动按钮
    if (this.detectedFields.length > 0) {
      this.floatingButton.show();
    }

    // 显示学习模式开关按钮
    this.showTrainingToggle();

    // 暴露全局方法供按钮调用
    window.jobViewAutoFill = {
      performAutoFill: this.performAutoFill.bind(this)
    };
  }

  // 加载简历数据
  async loadResumeData() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(
        { type: 'GET_RESUME_DATA' },
        (response) => {
          if (response.success) {
            this.resumeData = response.data;
          }
          resolve();
        }
      );
    });
  }

  // 检测表单字段
  detectFormFields() {
    this.detectedFields = this.detector.detectAllFields();
    console.log('检测到的字段:', this.detectedFields);
  }

  // 执行自动填充
  async performAutoFill() {
    if (!this.resumeData) {
      alert('简历数据未加载，请先登录JobView账号');
      return;
    }

    if (this.detectedFields.length === 0) {
      alert('未检测到可填充的字段');
      return;
    }

    try {
      const allFields = [...this.detectedFields];

      // 先应用学习映射
      const learned = await this.applyLearnedMappings();
      if (learned && learned.length) {
        allFields.push(...learned);
      }

      const results = await this.autoFiller.fill(allFields, this.resumeData);

      const successCount = results.filter(r => r.success).length;
      const totalCount = results.length;

      console.log('填充结果:', results);

      if (successCount > 0) {
        alert(`成功填充 ${successCount}/${totalCount} 个字段`);
      } else {
        alert('未能填充任何字段，请检查页面是否支持');
      }
    } catch (error) {
      console.error('自动填充失败:', error);
      alert('自动填充失败: ' + error.message);
    }
  }

  // 监听页面变化
  observePageChanges() {
    const observer = new MutationObserver((mutations) => {
      let shouldRedetect = false;

      mutations.forEach((mutation) => {
        if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
          // 检查是否有新的表单元素
          const hasFormElements = Array.from(mutation.addedNodes).some(node =>
            node.nodeType === Node.ELEMENT_NODE &&
            (node.querySelector('input, select, textarea') ||
             ['INPUT', 'SELECT', 'TEXTAREA'].includes(node.tagName))
          );

          if (hasFormElements) {
            shouldRedetect = true;
          }
        }
      });

      if (shouldRedetect) {
        setTimeout(() => {
          this.detectFormFields();
          if (this.detectedFields.length > 0 && !this.floatingButton.isVisible) {
            this.floatingButton.show();
          }
        }, 1000);
      }
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true
    });
  }

  // ================= 学习模式与映射 =================
  showTrainingToggle() {
    if (this.trainingButton) return;
    const btn = document.createElement('div');
    btn.id = 'jobview-train-button';
    btn.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="white"><path d="M19.14,12.94a7.14,7.14,0,0,0,.05-.94,7.14,7.14,0,0,0-.05-.94l2.11-1.65a.5.5,0,0,0,.12-.64l-2-3.46a.5.5,0,0,0-.6-.22l-2.49,1a7,7,0,0,0-1.63-.94l-.38-2.65A.5.5,0,0,0,13.23,2H10.77a.5.5,0,0,0-.5.42L9.89,5.07a7,7,0,0,0-1.63.94l-2.49-1a.5.5,0,0,0-.6.22l-2,3.46a.5.5,0,0,0,.12.64L5.4,11.06a7.14,7.14,0,0,0-.05.94,7.14,7.14,0,0,0,.05.94L3.29,14.59a.5.5,0,0,0-.12.64l2,3.46a.5.5,0,0,0,.6.22l2.49-1a7,7,0,0,0,1.63.94l.38,2.65a.5.5,0,0,0,.5.42h2.46a.5.5,0,0,0,.5-.42l.38-2.65a7,7,0,0,0,1.63-.94l2.49,1a.5.5,0,0,0,.6-.22l2-3.46a.5.5,0,0,0-.12-.64Zm-7.14,2.56A3.5,3.5,0,1,1,15.5,12,3.5,3.5,0,0,1,12,15.5Z"/></svg>';
    Object.assign(btn.style, {
      position: 'fixed', top: '80px', right: '20px', width: '44px', height: '44px',
      backgroundColor: '#722ed1', borderRadius: '22px', cursor: 'pointer', zIndex: 10000,
      display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
      opacity: '0.9'
    });
    btn.title = 'JobView 字段选择模式';
    btn.addEventListener('mouseenter', () => btn.style.opacity = '1');
    btn.addEventListener('mouseleave', () => btn.style.opacity = '0.9');
    btn.addEventListener('click', () => this.toggleTraining());
    document.body.appendChild(btn);
    this.trainingButton = btn;
  }

  toggleTraining() {
    this.trainingMode = !this.trainingMode;
    if (this.trainingMode) {
      this.enableTrainingListeners();
      this.toast('字段选择模式已开启：点击页面字段以建立映射');
      if (this.trainingButton) this.trainingButton.style.backgroundColor = '#52c41a';
    } else {
      this.disableTrainingListeners();
      this.toast('字段选择模式已关闭');
      if (this.trainingButton) this.trainingButton.style.backgroundColor = '#722ed1';
    }
  }

  enableTrainingListeners() {
    this._onMouseOver = (e) => {
      const el = e.target;
      if (!el || !(el instanceof HTMLElement)) return;
      // 忽略扩展自身UI元素
      if (el.closest('#jobview-train-button') || el.closest('#jobview-autofill-button')) return;
      this._hoverPrevOutline = el.style.outline;
      el.style.outline = '2px dashed #faad14';
    };
    this._onMouseOut = (e) => {
      const el = e.target;
      if (!el || !(el instanceof HTMLElement)) return;
      if (el.closest('#jobview-train-button') || el.closest('#jobview-autofill-button')) return;
      el.style.outline = this._hoverPrevOutline || '';
    };
    this._onClick = async (e) => {
      const el = e.target;
      if (!el || !(el instanceof HTMLElement)) return;
      // 若点击的是扩展自身按钮，允许事件继续，避免拦截导致无法关闭
      if (el.closest('#jobview-train-button') || el.closest('#jobview-autofill-button')) return;

      e.preventDefault();
      e.stopPropagation();

      const field = prompt('映射到哪个字段? 可选示例: name,email,phone,city,school,major,degree,company,position,salary 或 company_name,position_title,salary_text,location_text,job_description', 'company_name');
      if (!field) return;
      const strategy = this.guessStrategy(el);
      const xpath = this.detector.getXPath(el);
      await this.mappingManager.saveFieldMapping(location.hostname, field.trim(), { type: 'xpath', path: xpath, strategy });
      this.toast('映射已保存: ' + field + ' → ' + xpath + '（' + strategy + '）');
    };
    document.addEventListener('mouseover', this._onMouseOver, true);
    document.addEventListener('mouseout', this._onMouseOut, true);
    document.addEventListener('click', this._onClick, true);
  }

  disableTrainingListeners() {
    document.removeEventListener('mouseover', this._onMouseOver, true);
    document.removeEventListener('mouseout', this._onMouseOut, true);
    document.removeEventListener('click', this._onClick, true);
  }

  guessStrategy(el) {
    const tag = el.tagName.toLowerCase();
    if (tag === 'select') return 'selectClick';
    const cls = (el.className || '').toString().toLowerCase();
    const role = (el.getAttribute('role') || '').toLowerCase();
    if (role.includes('combobox') || cls.includes('select') || cls.includes('picker') || cls.includes('cascader')) {
      return 'selectClick';
    }
    return 'text';
  }

  async applyLearnedMappings() {
    const map = await this.mappingManager.loadSiteMapping(location.hostname);
    if (!map) return [];
    const fields = [];
    for (const [jobField, rule] of Object.entries(map.fields || {})) {
      const el = this.mappingManager.queryByRule(rule);
      if (el) {
        fields.push({ element: el, type: jobField, fillStrategy: rule.strategy || 'text', info: {}, confidence: 100, xpath: rule.path });
      }
    }
    return fields;
  }

  toast(text) {
    try {
      const n = document.createElement('div');
      n.textContent = text;
      Object.assign(n.style, { position: 'fixed', right: '20px', bottom: '20px', background: 'rgba(0,0,0,0.75)', color: '#fff', padding: '8px 12px', borderRadius: '6px', zIndex: 999999, fontSize: '13px' });
      document.body.appendChild(n);
      setTimeout(() => n.remove(), 2200);
    } catch {}
  }
}

class JobCaptureManager {
  constructor() {
    this.host = location.hostname || '';
    this.enabled = false;
    this.settings = {};
    this.observer = null;
    this.lastPromptAt = 0;
    this.modal = null;
    this.form = null;
    this.alert = null;
    this.rememberCheckbox = null;
    this.companyInput = null;
    this.positionInput = null;
    this.dateInput = null;
    this.attrSelect = null;
    this.locationInput = null;
    this.salaryInput = null;
    this.notesInput = null;
    this.boundButtons = new WeakSet();
    this.triggerKeywords = [
      '投递', '申请', '提交', '报名', '发送',
      'apply', 'submit', 'send', 'postuler', 'bewerben'
    ];
    this.init();
  }

  async init() {
    if (!this.host || isJobViewHost(this.host)) {
      return;
    }
    await this.loadSettings();
    await this.refreshSiteStatus();
    chrome.runtime.onMessage.addListener((request) => {
      if (request?.type === 'SITE_STATUS_CHANGED') {
        const changedHost = sanitizeHost(request.host || '');
        const currentHost = sanitizeHost(this.host);
        if (changedHost === currentHost) {
          this.setEnabled(Boolean(request.enabled));
        }
      }
      if (request?.type === 'AUTH_STATE_CHANGED' && request.isLoggedIn === false) {
        this.closeModal();
        showToast('您已退出登录，请先登录 JobView');
      }
    });
  }

  async loadSettings() {
    try {
      const res = await sendMessageSafe({ type: 'GET_SETTINGS' });
      if (res?.success) {
        this.settings = res.data || {};
      }
    } catch (e) {
      console.warn('Load settings failed:', e);
    }
  }

  async refreshSiteStatus() {
    try {
      const res = await sendMessageSafe({ type: 'IS_SITE_ENABLED', host: this.host });
      const enabled = Boolean(res?.enabled);
      this.setEnabled(enabled);
    } catch (e) {
      console.warn('Check site status failed:', e);
    }
  }

  setEnabled(enabled) {
    if (this.enabled === enabled) return;
    this.enabled = enabled;
    if (enabled) {
      this.start();
    } else {
      this.stop();
    }
  }

  start() {
    this.scanAndBind();
    if (!this.observer) {
      this.observer = new MutationObserver(() => {
        this.scanAndBind();
      });
      try {
        this.observer.observe(document.body, { childList: true, subtree: true });
      } catch (e) {
        console.warn('Observer start failed:', e);
      }
    }
  }

  stop() {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
    this.boundButtons = new WeakSet();
    this.closeModal();
  }

  scanAndBind() {
    if (!this.enabled) return;
    const candidates = document.querySelectorAll('button, a, input[type="button"], input[type="submit"], div[role="button"], span[role="button"]');
    candidates.forEach((el) => {
      if (this.boundButtons.has(el)) return;
      if (!this.isVisible(el)) return;
      if (!this.isApplyButton(el)) return;
      this.bindButton(el);
    });
  }

  isVisible(el) {
    const rect = el.getBoundingClientRect();
    const style = window.getComputedStyle(el);
    return style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
  }

  isApplyButton(el) {
    const text = (el.innerText || el.value || '').trim().toLowerCase();
    if (!text) return false;
    return this.triggerKeywords.some(keyword => text.includes(keyword.toLowerCase()));
  }

  bindButton(el) {
    const handler = (event) => {
      if (!this.enabled) return;
      const now = Date.now();
      if (now - this.lastPromptAt < 1000) return;
      this.lastPromptAt = now;
      setTimeout(() => {
        this.openModal();
      }, 180);
    };
    el.addEventListener('click', handler, { capture: false, passive: true });
    this.boundButtons.add(el);
  }

  openModal() {
    if (this.modal) {
      this.modal.style.display = 'flex';
      this.resetForm();
      return;
    }
    this.modal = document.createElement('div');
    this.modal.className = 'jobview-capture-mask';
    this.modal.innerHTML = `
      <div class="jobview-capture-modal">
        <div class="jobview-capture-header">
          <div>
            <h3>保存投递信息到 JobView</h3>
            <p>填写岗位信息，便于后续追踪</p>
          </div>
          <button type="button" class="jobview-capture-close" aria-label="close">×</button>
        </div>
        <form class="jobview-capture-form">
          <div class="jobview-capture-alert" style="display:none;"></div>
          <label class="jobview-capture-field">
            <span>公司名称 <span class="required">*</span></span>
            <input id="jobview-company" type="text" placeholder="请输入公司名称" required />
          </label>
          <label class="jobview-capture-field">
            <span>职位名称 <span class="required">*</span></span>
            <input id="jobview-position" type="text" placeholder="请输入职位名称" required />
          </label>
          <div class="jobview-capture-row">
            <label class="jobview-capture-field">
              <span>投递日期</span>
              <input id="jobview-date" type="date" />
            </label>
            <label class="jobview-capture-field">
              <span>企业属性 <span class="required">*</span></span>
              <select id="jobview-attr" required>
                <option value="">请选择</option>
                <option value="私企">私企</option>
                <option value="央国企">央国企</option>
              </select>
            </label>
          </div>
          <div class="jobview-capture-row">
            <label class="jobview-capture-field">
              <span>地区（可选）</span>
              <input id="jobview-location" type="text" placeholder="工作地点" />
            </label>
            <label class="jobview-capture-field">
              <span>薪资（可选）</span>
              <input id="jobview-salary" type="text" placeholder="如：20-30K/月" />
            </label>
          </div>
          <label class="jobview-capture-field">
            <span>备注</span>
            <textarea id="jobview-notes" rows="3" placeholder="可选，记录面试官、渠道等信息"></textarea>
          </label>
          <label class="jobview-capture-remember">
            <input id="jobview-remember-attr" type="checkbox" />
            下次默认使用当前企业属性
          </label>
          <div class="jobview-capture-actions">
            <button type="button" class="jobview-btn jobview-btn-secondary" data-action="cancel">暂不保存</button>
            <button type="submit" class="jobview-btn jobview-btn-primary" data-action="submit">保存到 JobView</button>
          </div>
        </form>
      </div>
    `;
    document.body.appendChild(this.modal);

    this.form = this.modal.querySelector('form');
    this.alert = this.modal.querySelector('.jobview-capture-alert');
    this.companyInput = this.modal.querySelector('#jobview-company');
    this.positionInput = this.modal.querySelector('#jobview-position');
    this.dateInput = this.modal.querySelector('#jobview-date');
    this.attrSelect = this.modal.querySelector('#jobview-attr');
    this.locationInput = this.modal.querySelector('#jobview-location');
    this.salaryInput = this.modal.querySelector('#jobview-salary');
    this.notesInput = this.modal.querySelector('#jobview-notes');
    this.rememberCheckbox = this.modal.querySelector('#jobview-remember-attr');

    this.modal.querySelector('.jobview-capture-close').addEventListener('click', () => this.closeModal());
    this.form.querySelector('[data-action="cancel"]').addEventListener('click', () => this.closeModal());
    this.form.addEventListener('submit', (e) => {
      e.preventDefault();
      this.submit();
    });

    this.modal.addEventListener('click', (e) => {
      if (e.target === this.modal) {
        this.closeModal();
      }
    });

    this.resetForm();
  }

  resetForm() {
    if (!this.form) return;
    this.alert.style.display = 'none';
    this.alert.textContent = '';
    this.companyInput.value = '';
    this.positionInput.value = '';
    this.locationInput.value = '';
    this.salaryInput.value = '';
    this.notesInput.value = '';
    this.rememberCheckbox.checked = false;
    const defaultAttr = this.settings?.default_company_attribute || '';
    if (defaultAttr && this.attrSelect) {
      this.attrSelect.value = defaultAttr;
    } else if (this.attrSelect) {
      this.attrSelect.value = '';
    }
    if (this.dateInput) {
      this.dateInput.value = getTodayDateString();
    }
    if (this.companyInput) {
      this.companyInput.focus();
    }
  }

  closeModal() {
    if (!this.modal) return;
    this.modal.style.display = 'none';
  }

  async submit() {
    const loginStatus = await sendMessageSafe({ type: 'CHECK_LOGIN' });
    if (!loginStatus?.isLoggedIn) {
      this.closeModal();
      showToast('请先登录 JobView 后再保存');
      return;
    }

    const companyName = this.companyInput.value.trim();
    const positionTitle = this.positionInput.value.trim();
    const companyAttr = this.attrSelect.value.trim();
    if (!companyName || !positionTitle || !companyAttr) {
      this.showAlert('请完整填写带 * 的必填信息');
      return;
    }

    const payload = {
      company_name: companyName,
      position_title: positionTitle,
      application_date: this.dateInput.value,
      company_attribute: companyAttr,
      work_location: this.locationInput.value.trim() || undefined,
      salary_range: this.salaryInput.value.trim() || undefined,
      notes: this.notesInput.value.trim() || undefined
    };

    const remember = this.rememberCheckbox.checked;

    try {
      const res = await sendMessageSafe({ type: 'CREATE_APPLICATION', payload });
      if (res?.duplicate) {
        const force = window.confirm('检测到疑似重复投递，仍然保存吗？');
        if (force) {
          const retry = await sendMessageSafe({ type: 'CREATE_APPLICATION', payload, force: true });
          if (!retry?.success) throw new Error(retry?.message || retry?.error || '提交失败');
          this.handleSuccess();
        }
        return;
      }
      if (res?.queued) {
        this.handleSuccess('网络不可用，已加入稍后提交队列');
      } else if (res?.success) {
        this.handleSuccess();
      } else {
        throw new Error(res?.message || res?.error || '提交失败');
      }

      if (remember) {
        await sendMessageSafe({ type: 'SAVE_SETTINGS', data: { default_company_attribute: companyAttr } });
        this.settings.default_company_attribute = companyAttr;
      }
    } catch (e) {
      this.showAlert(e?.message || '保存失败，请稍后重试');
    }
  }

  handleSuccess(customMessage) {
    this.closeModal();
    showToast(customMessage || '已保存到 JobView！');
  }

  showAlert(message) {
    if (!this.alert) return;
    this.alert.textContent = message;
    this.alert.style.display = 'block';
  }
}

function sendMessageSafe(payload) {
  return new Promise((resolve) => {
    try {
      chrome.runtime.sendMessage(payload, (response) => {
        resolve(response);
      });
    } catch (e) {
      console.warn('sendMessageSafe error:', e);
      resolve(null);
    }
  });
}

function getTodayDateString() {
  const today = new Date();
  const yyyy = today.getFullYear();
  const mm = String(today.getMonth() + 1).padStart(2, '0');
  const dd = String(today.getDate()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

function showToast(text) {
  try {
    const toast = document.createElement('div');
    toast.className = 'jobview-success-toast';
    toast.textContent = text;
    document.body.appendChild(toast);
    setTimeout(() => {
      toast.style.opacity = '0';
      setTimeout(() => toast.remove(), 300);
    }, 2000);
  } catch (e) {
    console.warn('showToast error:', e);
  }
}

function sanitizeHost(host) {
  if (!host) return '';
  return String(host).trim().toLowerCase().replace(/^https?:\/\//, '').split('/')[0];
}

function isJobViewHost(host) {
  const clean = sanitizeHost(host);
  return clean.includes('jobview.bfsmlt.top') || clean === 'localhost' || clean.startsWith('127.0.0.1');
}

// 简单的站点映射存储管理器（MVP）
class MappingManager {
  constructor() {
    this.STORAGE_KEY = 'jobview_site_mappings';
  }

  async loadSiteMapping(host) {
    try {
      const res = await chrome.storage.local.get([this.STORAGE_KEY]);
      const all = res[this.STORAGE_KEY] || {};
      return all[host] || null;
    } catch { return null; }
  }

  async saveFieldMapping(host, jobField, rule) {
    const res = await chrome.storage.local.get([this.STORAGE_KEY]);
    const all = res[this.STORAGE_KEY] || {};
    if (!all[host]) all[host] = { fields: {} };
    all[host].fields[jobField] = { type: rule.type, path: rule.path, strategy: rule.strategy };
    await chrome.storage.local.set({ [this.STORAGE_KEY]: all });
  }

  queryByRule(rule) {
    if (!rule || rule.type !== 'xpath' || !rule.path) return null;
    try {
      const xpath = rule.path;
      const node = document.evaluate(xpath, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue;
      if (node && node instanceof HTMLElement) return node;
    } catch {}
    return null;
  }
}

// 页面加载完成后初始化
function initJobViewFeatures() {
  if (initAuthBridgeIfOnJobViewSite()) {
    return;
  }
  if (!window.__jobViewCSM) {
    try { window.__jobViewCSM = new ContentScriptManager(); } catch (e) { console.warn('Init CSM failed:', e); }
  }
  if (!window.__jobViewJobCapture) {
    try { window.__jobViewJobCapture = new JobCaptureManager(); } catch (e) { console.warn('Init JobCapture failed:', e); }
  }
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initJobViewFeatures);
} else {
  initJobViewFeatures();
}

// 接收来自 popup 的消息（用于按需触发自动填充、连通性检测）
try {
  chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request && request.type === 'PING') {
      sendResponse({ ok: true });
      return; // 同步响应
    }
    if (request && request.type === 'EXTRACT_JOB_POSTING') {
      (async () => {
        try {
          const data = await extractJobPosting();
          sendResponse({ success: true, data });
        } catch (e) {
          sendResponse({ success: false, error: e?.message || String(e) });
        }
      })();
      return true; // 异步响应
    }
    if (request && request.type === 'START_TRAINING') {
      try {
        const mgr = window.__jobViewCSM;
        if (mgr) {
          // 显示训练按钮并确保进入训练模式
          try { mgr.showTrainingToggle(); } catch {}
          if (!mgr.trainingMode && typeof mgr.toggleTraining === 'function') {
            mgr.toggleTraining();
          }
          sendResponse({ success: true });
        } else {
          sendResponse({ success: false, error: 'Content manager not ready' });
        }
      } catch (e) {
        sendResponse({ success: false, error: e?.message || String(e) });
      }
      return true;
    }
    if (request && request.type === 'PERFORM_AUTOFILL') {
      try {
        if (window.jobViewAutoFill && typeof window.jobViewAutoFill.performAutoFill === 'function') {
          window.jobViewAutoFill.performAutoFill().then(() => sendResponse({ success: true })).catch(e => sendResponse({ success: false, error: e?.message || String(e) }));
          return true; // 异步响应
        } else if (window.__jobViewCSM) {
          // 在未启用通用检测的站点，仅使用学习映射执行填充
          const mgr = window.__jobViewCSM;
          (async () => {
            try {
              await mgr.loadResumeData();
              const learned = await mgr.applyLearnedMappings();
              if (!mgr.autoFiller || !learned || learned.length === 0) {
                sendResponse({ success: false, error: '未找到可用映射' });
                return;
              }
              const results = await mgr.autoFiller.fill(learned, mgr.resumeData);
              const ok = (results || []).some(r => r.success);
              sendResponse({ success: ok, results });
            } catch (err) {
              sendResponse({ success: false, error: err?.message || String(err) });
            }
          })();
          return true; // 异步响应
        } else {
          sendResponse({ success: false, error: 'AutoFill not ready' });
        }
      } catch (e) {
        sendResponse({ success: false, error: e?.message || String(e) });
      }
    }
  });
} catch {}

// ================= 职位解析（MVP） =================
async function extractJobPosting() {
  // 先进行常规解析
  const base = extractJobPostingCore();
  // 应用训练模式标注的映射覆盖（让标注即刻生效）
  try {
    const withMapped = await applyMappingOverrides(base);
    return withMapped;
  } catch {
    return base;
  }
}

function extractJobPostingCore() {
  const source_domain = location.hostname;
  const job_url = location.href;

  // 1) JSON-LD 提取
  const fromJsonLd = extractFromJsonLd();
  if (fromJsonLd && (fromJsonLd.company_name || fromJsonLd.position_title)) {
    return { ...fromJsonLd, source_domain, job_url, captured_at: Date.now() };
  }

  // 2) 站点适配器
  const adapter = pickAdapter(source_domain);
  if (adapter) {
    try {
      const data = adapter();
      if (data && (data.company_name || data.position_title)) {
        return { ...data, source_domain, job_url, captured_at: Date.now() };
      }
    } catch {}
  }

  // 3) Meta & 通用 DOM 兜底
  const generic = extractGeneric();
  return { ...generic, source_domain, job_url, captured_at: Date.now() };
}

async function applyMappingOverrides(base) {
  const mgr = new MappingManager();
  const host = location.hostname;
  const map = await mgr.loadSiteMapping(host);
  if (!map || !map.fields) return base;

  // 支持的目标字段与同义键（用户可能输入简写）
  const targets = {
    company_name: ['company_name', 'company'],
    position_title: ['position_title', 'position', 'job_title', 'title'],
    salary_text: ['salary_text', 'salary', 'pay'],
    location_text: ['location_text', 'location', 'city', 'work_location'],
    job_description: ['job_description', 'description', 'desc', 'jd']
  };

  const result = { ...(base || {}) };

  for (const [key, aliases] of Object.entries(targets)) {
    let rule = null;
    for (const a of aliases) {
      if (map.fields[a]) { rule = map.fields[a]; break; }
    }
    if (!rule) continue;
    const el = mgr.queryByRule(rule);
    if (!el) continue;
    let v = '';
    const tag = el.tagName ? el.tagName.toLowerCase() : '';
    if (tag === 'input' || tag === 'textarea' || (el instanceof HTMLInputElement) || (el instanceof HTMLTextAreaElement)) {
      v = el.value || '';
    } else {
      v = textify(el.textContent || el.innerText || '');
    }
    if (!v) continue;
    if (key === 'job_description') v = truncate(v, 6000);
    result[key] = v;
  }

  return result;
}

function extractFromJsonLd() {
  try {
    const scripts = Array.from(document.querySelectorAll('script[type="application/ld+json"]'));
    for (const s of scripts) {
      let json;
      try {
        json = JSON.parse(s.textContent.trim());
      } catch { continue; }

      const candidates = [];
      if (Array.isArray(json)) candidates.push(...json);
      else if (json['@graph']) candidates.push(...[].concat(json['@graph']));
      else candidates.push(json);

      for (const obj of candidates) {
        const type = (obj['@type'] || obj['type'] || '').toString();
        if (/JobPosting/i.test(type)) {
          const position_title = obj.title || obj.name || '';
          const company_name = obj.hiringOrganization?.name || obj.hiringOrganization || '';
          const salary_text = normalizeSalary(obj.baseSalary?.value?.value, obj.baseSalary?.value?.currency, obj.baseSalary?.unitText);
          const location_text = obj.jobLocation?.address?.addressLocality || obj.jobLocation?.address?.addressRegion || '';
          let desc = obj.description || '';
          try { desc = stripHtml(String(desc)); } catch {}
          return {
            company_name: textify(company_name),
            position_title: textify(position_title),
            salary_text: textify(salary_text),
            location_text: textify(location_text),
            job_description: truncate(desc, 6000)
          };
        }
      }
    }
  } catch {}
  return null;
}

function normalizeSalary(value, currency, unit) {
  if (!value) return '';
  let v = Number(value);
  if (!isFinite(v)) return '';
  // 仅给出粗略表达
  let unitTxt = unit || 'MONTH';
  if (/YEAR/i.test(unitTxt)) unitTxt = '年';
  else unitTxt = '月';
  return `${v} ${currency || ''}/${unitTxt}`.trim();
}

function pickAdapter(host) {
  const h = host || '';
  if (h.includes('zhipin.com')) return extractZhipin;
  if (h.includes('51job.com')) return extract51job;
  if (h.includes('zhaopin.com')) return extractZhaopin;
  if (h.includes('mokahr.com')) return extractMokahr;
  if (h.includes('didiglobal.com')) return extractDidiCampus;
  if (h.includes('lagou.com')) return extractLagou;
  if (h.includes('liepin.com')) return extractLiepin;
  if (h.includes('tencent.com')) return extractTencent;
  if (h.includes('bytedance.com')) return extractBytedance;
  if (h.includes('alibaba.com') || h.includes('alibaba-inc.com')) return extractAlibaba;
  if (h.includes('baidu.com')) return extractBaidu;
  if (h.includes('meituan.com')) return extractMeituan;
  return null;
}

function extractGeneric() {
  // Meta 标题
  const ogTitle = getMeta(['og:title', 'twitter:title']);
  const titleEl = document.querySelector('h1, h2, .job-title, [class*="job"][class*="title"]');
  let position_title = textify(ogTitle || titleEl?.textContent || '');

  // 可能包含“公司-职位”，尝试拆分
  let company_name = '';
  if (position_title.includes('-')) {
    const parts = position_title.split('-').map(s => s.trim());
    if (parts.length >= 2) {
      // 猜测“职位 - 公司”或“公司 - 职位”，选择较合理的
      const a = parts[0], b = parts[1];
      if (/[公司|集团|有限|股份]/.test(b)) { company_name = b; position_title = a; }
      else if (/[公司|集团|有限|股份]/.test(a)) { company_name = a; position_title = b; }
    }
  }

  // 公司元素通用选择器
  const companyEl = document.querySelector('.company-name a, .company-name, [class*="company"] a, [class*="company"]');
  company_name = textify(company_name || companyEl?.textContent || '');

  const salaryEl = document.querySelector('.salary, .pay, .job-money, [class*="salary"]');
  const locationEl = document.querySelector('.job-area, .location, .job-address, [class*="address"], [class*="location"]');
  const descEl = document.querySelector('.job-sec, .job-description, .jd, [class*="description"], article');

  return {
    company_name,
    position_title,
    salary_text: textify(salaryEl?.textContent || ''),
    location_text: textify(locationEl?.textContent || ''),
    job_description: truncate(textify(descEl?.innerText || ''), 6000)
  };
}

// ======= 站点提取器（增强稳定性：多选择器优先队列） =======
function pickFirstText(selectors) {
  for (const sel of selectors) {
    const el = document.querySelector(sel);
    if (el) {
      const t = textify(el.textContent || el.innerText || '');
      if (t) return t;
    }
  }
  return '';
}

function extractZhipin() {
  const position_title = pickFirstText([
    'h1', '.name', '.job-title', '.title-text', '.job-name'
  ]);
  const company_name = pickFirstText([
    '.company-info .name a', '.company-info .title', '.job-company .name a', '.job-company .name', '.sider-company .company-name', '.company-detail .company-name'
  ]);
  const salary_text = pickFirstText([
    '.salary', '.job-money', '.job-sec .salary', '.job-primary .salary'
  ]);
  const location_text = pickFirstText([
    '.info-primary .job-area', '.job-address', '.location', '.job-area', '.job-detail .location'
  ]);
  const descText = pickFirstText([
    '.job-sec', '.job-sec-text', '.job-detail', '.job-description', '.detail-content'
  ]);
  return {
    company_name,
    position_title,
    salary_text,
    location_text,
    job_description: truncate(descText, 6000)
  };
}

function extract51job() {
  const position_title = pickFirstText([
    'div.cn > h1', 'h1', '.tit h1', '.jtop h1'
  ]);
  const company_name = pickFirstText([
    'p.cname a', '.cname a', '.com_name a', '.com_name', '.company-name a', '.company-name'
  ]);
  const salary_text = pickFirstText([
    'div.cn strong', '.cn .lname', '.jtag .sp4', '.job-item-title .sal', '.salary'
  ]);
  const location_text = pickFirstText([
    '.ltype', '.msg', '.work_addr', '.fp', '.job-msg .fp', '.job-area'
  ]);
  const descText = pickFirstText([
    '#job-detail', '#jobMsg', '.bmsg.job_msg', '.job-sec', '.job-content', '.job-detail'
  ]);
  return {
    company_name,
    position_title,
    salary_text,
    location_text,
    job_description: truncate(descText, 6000)
  };
}

function extractZhaopin() {
  const position_title = pickFirstText([
    'h1', 'h3[class*="job"]', '.summary-plane__title__jobname', '.job-title', '.job-name'
  ]);
  const company_name = pickFirstText([
    '.company-name', 'a.company__title', '.summary-plane__title__company a', '.company__title', '.company__name', '.job-company a'
  ]);
  const salary_text = pickFirstText([
    '[class*="salary"]', '.summary-plane__salary', '.job-salary', '.job-money'
  ]);
  const location_text = pickFirstText([
    '[class*="location"]', '[class*="address"]', '.job-address', '.job-location', '.job-area'
  ]);
  const descText = pickFirstText([
    '.describtion', '.job-description', '.pos-ul', 'article', '.detail-content', '.job-detail'
  ]);
  return {
    company_name,
    position_title,
    salary_text,
    location_text,
    job_description: truncate(descText, 6000)
  };
}

function extractMokahr() {
  const position_title = pickFirstText([
    '.job-title', '.job-head .title', 'h1', '.job-name'
  ]);
  const company_name = pickFirstText([
    '.company-name', '.company-info a', '[class*="company"] a', '.topbar .logo-title', '.company a'
  ]);
  const salary_text = pickFirstText([
    '[class*="salary"]', '.salary', '.job-salary', '.job-money'
  ]);
  const location_text = pickFirstText([
    '[class*="address"]', '[class*="location"]', '.job-address', '.job-location', '.job-area'
  ]);
  const descText = pickFirstText([
    '.job-description', '.job-content', '.content', '.detail-content', '.description'
  ]);
  return {
    company_name,
    position_title,
    salary_text,
    location_text,
    job_description: truncate(descText, 6000)
  };
}

function extractDidiCampus() {
  // 滴滴校招站点（SPA），常见结构：标题/公司/地点/薪资/描述
  const position_title = pickFirstText([
    '.job-title', '.job-name', '.position-title', '.job-header h1', 'h1'
  ]);
  // 公司名在头部或侧边，找不到则回退为“滴滴出行”
  let company_name = pickFirstText([
    '.company-name', '.header .title', '.topbar .logo-title', '.company a', '.header .company', '.logo + span'
  ]);
  if (!company_name) {
    // 页面元素里如果包含“滴滴”字样则使用“滴滴出行”，否则使用“DiDi Global”
    const hasZhName = /滴滴/.test(document.body.innerText || '');
    company_name = hasZhName ? '滴滴出行' : 'DiDi Global';
  }
  const salary_text = pickFirstText([
    '.salary', '.compensation', '.job-basic .salary', '.job-info .salary', '.info .salary'
  ]);
  const location_text = pickFirstText([
    '.job-location', '.location', '.job-basic .city', '.job-info .city', '.info .city', '[data-label*="地点"] + *', '[class*="地点"] + *'
  ]);
  const descText = pickFirstText([
    '.job-description', '.job-content', '.job-desc', '.description', '#job-detail', '.detail-content', '.content'
  ]);
  return {
    company_name,
    position_title,
    salary_text,
    location_text,
    job_description: truncate(descText, 6000)
  };
}

function extractLagou() {
  const position_title = pickFirstText(['.position-head .name', '.job-name', 'h1', '.position-title']);
  const company_name = pickFirstText(['.position-head .company a', '.company', '.company-name a', '.job-company a']);
  const salary_text = pickFirstText(['.position-content .salary', '.job-request .salary', '.salary']);
  const location_text = pickFirstText(['.work_addr', '.job-addr', '.job-location', '.job-area']);
  const descText = pickFirstText(['.job_bt', '.job-detail', '.job-description', '.content']);
  return { company_name, position_title, salary_text, location_text, job_description: truncate(descText, 6000) };
}

function extractLiepin() {
  const position_title = pickFirstText(['.job-title', '.job-detail-title', 'h1']);
  const company_name = pickFirstText(['.company-logo-txt', '.company-name', '.company a', '.info-company a']);
  const salary_text = pickFirstText(['.job-salary', '.salary', '.compensation']);
  const location_text = pickFirstText(['.basic-infor .location', '.job-address', '.job-area', '.address']);
  const descText = pickFirstText(['.content-word', '.job-description', '.job-detail', '.job-content']);
  return { company_name, position_title, salary_text, location_text, job_description: truncate(descText, 6000) };
}

function extractTencent() {
  const position_title = pickFirstText(['.job-title', 'h1', '.job-name']);
  let company_name = pickFirstText(['.company-name', '.header .title']);
  if (!company_name) company_name = '腾讯';
  const salary_text = pickFirstText(['.salary', '.job-salary']);
  const location_text = pickFirstText(['.job-location', '.address', '.city', '.job-area']);
  const descText = pickFirstText(['.job-description', '.job-content', '.description', '.detail-content']);
  return { company_name, position_title, salary_text, location_text, job_description: truncate(descText, 6000) };
}

function extractBytedance() {
  const position_title = pickFirstText(['.job-title', 'h1', '.title']);
  let company_name = pickFirstText(['.company-name', '.header .company', '.logo-title']);
  if (!company_name) company_name = '字节跳动';
  const salary_text = pickFirstText(['.salary', '.job-salary']);
  const location_text = pickFirstText(['.job-location', '.location', '.city']);
  const descText = pickFirstText(['.job-description', '.description', '.job-content']);
  return { company_name, position_title, salary_text, location_text, job_description: truncate(descText, 6000) };
}

function extractAlibaba() {
  const position_title = pickFirstText(['.job-title', 'h1', '.title']);
  let company_name = pickFirstText(['.company-name', '.header .company', '.logo-title']);
  if (!company_name) company_name = '阿里巴巴';
  const salary_text = pickFirstText(['.salary', '.job-salary']);
  const location_text = pickFirstText(['.job-location', '.location', '.city']);
  const descText = pickFirstText(['.job-description', '.description', '.job-content']);
  return { company_name, position_title, salary_text, location_text, job_description: truncate(descText, 6000) };
}

function extractBaidu() {
  const position_title = pickFirstText(['.job-title', 'h1', '.title']);
  let company_name = pickFirstText(['.company-name', '.header .company']);
  if (!company_name) company_name = '百度';
  const salary_text = pickFirstText(['.salary', '.job-salary']);
  const location_text = pickFirstText(['.job-location', '.location', '.city']);
  const descText = pickFirstText(['.job-description', '.description', '.job-content']);
  return { company_name, position_title, salary_text, location_text, job_description: truncate(descText, 6000) };
}

function extractMeituan() {
  const position_title = pickFirstText(['.job-title', 'h1', '.title']);
  let company_name = pickFirstText(['.company-name', '.header .company']);
  if (!company_name) company_name = '美团';
  const salary_text = pickFirstText(['.salary', '.job-salary']);
  const location_text = pickFirstText(['.job-location', '.location', '.city']);
  const descText = pickFirstText(['.job-description', '.description', '.job-content']);
  return { company_name, position_title, salary_text, location_text, job_description: truncate(descText, 6000) };
}

function getMeta(props) {
  for (const p of props) {
    const el = document.querySelector(`meta[property="${p}"]`) || document.querySelector(`meta[name="${p}"]`);
    if (el && el.content) return el.content;
  }
  return '';
}

function textify(s) {
  if (!s) return '';
  return String(s).replace(/\s+/g, ' ').trim();
}

function stripHtml(html) {
  const tmp = document.createElement('div');
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || '';
}

function truncate(s, n) { s = String(s || ''); return s.length > n ? s.slice(0, n) : s; }
