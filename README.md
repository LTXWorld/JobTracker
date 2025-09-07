# CLAUDE.md

这是一个功能完善的基于 Vue 3 + Go 的求职跟踪看板系统，帮助求职者全方位管理和跟踪投递状态。系统采用前后端分离架构，提供看板视图、时间线、提醒中心、数据统计等多个功能模块。

## 技术栈

### 前端

  - 框架: Vue 3 + TypeScript + Composition API
  - 构建工具: Vite
  - UI组件库: Ant Design Vue
  - 图表库: Vue ECharts (支持饼图、折线图、柱状图)
  - 状态管理: Pinia
  - 路由: Vue Router
  - 拖拽功能: Vuedraggable
  - 日期处理: Day.js

###  后端

  - 语言: Go
  - Web框架: Gin
  - 数据库: Postgresql
  - API设计: RESTful

##  核心功能模块

### 看板管理 (KanbanBoard.vue)

  主要功能:
  - 投递状态可视化展示
  - 拖拽式状态更新
  - 分页切换不同状态组（进行中、失败状态）
  - 卡片CRUD操作
  - 紧凑布局设计

  状态流转:
  - 进行中状态: 已投递 → 简历筛选中 → 笔试中 → 一面中 → 二面中 → 三面中 → HR面中
  - 失败状态: 简历挂、笔试挂、一面挂、二面挂、三面挂

  布局特点:
  - 头部：标题 + 状态切换标签 + 操作按钮水平布局
  - 主体：充分利用垂直空间的状态列
  - 紧凑卡片：仅显示公司名、职位、申请时间

### 时间线视图 (Timeline.vue) 

  功能特性:
  - 时间线形式展示所有投递记录
  - 多维度筛选:
    - 关键词搜索（公司名、职位名）
    - 状态筛选
    - 日期范围选择
    - 薪资范围筛选
    - 工作地点筛选
  - 高级功能:
    - 搜索关键词高亮显示
    - 多种排序方式（时间、薪资）
    - 分页展示
    - 快速统计（面试中、已拒绝、收到offer数量）
  - 交互操作:
    - 状态快速更新
    - 就地编辑功能

### 提醒中心 (Reminders.vue + ReminderManager.vue) 

  核心功能:
  - 智能提醒管理:
    - 面试时间提醒
    - 跟进任务提醒
    - 自定义提醒消息
  - 统计面板:
    - 今日待办、本周待办数量
    - 面试提醒、跟进提醒分类统计
  - 快速操作:
    - 选择投递记录快速设置提醒
    - 提醒类型筛选（全部/面试/跟进/今日/即将到来）
  - 个性化设置:
    - 默认提前提醒时间（15分钟-1天）
    - 提醒方式选择（浏览器通知、声音、邮件）
    - 面试状态自动创建提醒开关

### 数据统计 (Statistics.vue) 

  统计概览:
  - 总投递数、进行中、已通过、已失败数量
  - 总通过率、本月投递、本周投递统计

  可视化图表:
  - 状态分布饼图: 各状态投递数量分布
  - 投递趋势折线图: 最近30天投递趋势分析
  - 各阶段通过率柱状图: 笔试、一面、二面等各阶段通过率
  - 薪资分布柱状图: 不同薪资范围的岗位分布

  详细数据表格:
  - 按公司分组的投递详情统计
  - 包含投递数、面试中、收到offer、被拒绝等维度

### 高级筛选组件 (FilterBar.vue) 

  基础筛选:
  - 关键词搜索（支持公司名、职位搜索）
  - 状态下拉选择
  - 日期范围选择器
  - 薪资范围筛选

  高级筛选:
  - 工作地点筛选
  - 多种排序方式（最新投递、薪资高低等）
  - 可展开/收起的高级筛选面板

##  数据模型扩展

  interface JobApplication {
    id: number
    company_name: string       // 公司名称
    position_title: string     // 职位名称
    status: ApplicationStatus  // 当前状态
    application_date: string   // 申请日期
    salary_range?: string      // 薪资范围
    work_location?: string     // 工作地点
    interview_time?: string    // 面试时间
    notes?: string            // 备注信息

    // 提醒相关字段
    reminder_enabled?: boolean // 是否启用提醒
    reminder_time?: string     // 提醒时间
    follow_up_date?: string    // 跟进日期
    
    created_at: string        // 创建时间
    updated_at: string        // 更新时间
  }

  // 提醒数据模型
  interface Reminder {
    id: number
    application_id: number
    company_name: string
    position_title: string
    type: 'interview' | 'follow_up'
    reminder_time: string
    interview_time?: string
    message?: string
    is_dismissed: boolean
  }

  API接口扩展

  class JobApplicationAPI {
    // 基础CRUD
    static healthCheck(): Promise<boolean>
    static getAll(): Promise<JobApplication[]>
    static create(data: Partial<JobApplication>): Promise<JobApplication>
    static update(id: number, data: Partial<JobApplication>): Promise<JobApplication>
    static delete(id: number): Promise<void>

    // 统计数据
    static getStatistics(): Promise<StatisticsData>
    
    // 提醒相关
    static getReminders(): Promise<Reminder[]>
    static createReminder(data: Partial<Reminder>): Promise<Reminder>
    static dismissReminder(id: number): Promise<void>
  }

  状态管理增强

  Pinia Store 包含：
  - applications: 投递记录列表
  - statistics: 统计数据
  - reminders: 提醒列表
  - loading: 各种加载状态
  - 完善的异步操作方法

##  关键设计特性

  1. 响应式设计

  - 完整的移动端适配
  - 不同屏幕尺寸下的布局自动调整
  - 触屏友好的交互设计

  2. 用户体验优化

  - 搜索关键词高亮
  - 操作反馈提示
  - 加载状态显示
  - 空状态友好提示

  3. 数据可视化

  - 多种图表类型（饼图、折线图、柱状图）
  - 渐变色彩设计
  - 交互式图表展示

  4. 智能提醒系统

  - 基于时间的智能提醒
  - 多种提醒方式支持
  - 个性化提醒设置

##  功能完成度

###   ✅ 已完成功能

  - 看板视图: 拖拽式状态管理，紧凑布局
  - 时间线视图: 全功能筛选、排序、分页
  - 提醒中心: 智能提醒、统计面板、个性化设置
  - 数据统计: 多维度统计、可视化图表、详细数据表
  - 搜索筛选: 多字段筛选、高级筛选选项
  - 数据导出: 统计数据展示和分析
  - 批量操作: 批量导入功能
  - 响应式设计: 全端设备支持

###   🔄 可扩展功能

  - 数据导出到Excel/PDF
  - 邮件提醒功能集成
  - 面试日历集成
  - 求职进度报告生成
  - 多用户系统支持

## 总结

这是一个功能完善、设计精良的现代化求职管理工具。相比初期版本，当前系统已经实现了：

  1. 完整的数据管理流程 - 从录入到分析的全链路支持
  2. 强大的筛选和搜索 - 多维度数据筛选，快速定位目标信息
  3. 智能提醒系统 - 避免错过重要面试和跟进时间
  4. 数据可视化分析 - 通过图表直观了解求职进展
  5. 优秀的用户体验 - 响应式设计，操作便捷流畅

  技术实现上采用现代化的前端技术栈，代码结构清晰，可维护性强，为后续功能扩展奠定了良好基础。
