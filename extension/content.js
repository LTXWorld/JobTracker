/**
 * JobView AutoFill Extension - Content Script
 * 负责页面表单识别、自动填充和用户交互
 */

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
      await strategy(field.element, value);
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
    this.resumeData = null;
    this.detectedFields = [];

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
      const results = await this.autoFiller.fill(this.detectedFields, this.resumeData);

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
}

// 页面加载完成后初始化
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    new ContentScriptManager();
  });
} else {
  new ContentScriptManager();
}