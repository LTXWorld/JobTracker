export interface ChatMessage {
  content: string;
  isUser: boolean;
  timestamp: string;
  isTyping?: boolean;
}

export interface UserContext {
  totalApplications: number;
  pendingInterviews: number;
}

export interface ChatContext {
  currentRoute: string;
  applications: any[];
  userContext: UserContext;
}

// 知识库类型定义
interface KnowledgeItem {
  keywords: string[];
  response: string;
  category: 'usage' | 'analysis' | 'feature' | 'general';
}

// 机器人知识库
const knowledgeBase: KnowledgeItem[] = [
  // 使用指导类
  {
    keywords: ['看板', '拖拽', '状态更新', '如何使用看板'],
    response: `🎯 <strong>看板功能使用指南</strong><br><br>
    <strong>1. 拖拽更新状态</strong><br>
    • 直接拖拽岗位卡片到目标列即可更新状态<br>
    • 回退到早期阶段时需要确认操作<br><br>
    <strong>2. 查看详细信息</strong><br>
    • 点击卡片查看完整的状态变化时间线<br>
    • 使用右上角更多按钮进行编辑删除<br><br>
    <strong>3. 快速搜索</strong><br>
    • 在搜索框输入公司或职位名称快速定位`,
    category: 'usage'
  },
  {
    keywords: ['添加投递', '新建记录', '如何添加', '创建投递'],
    response: `📝 <strong>添加投递记录</strong><br><br>
    <strong>方式一：手动添加</strong><br>
    • 点击"添加投递"按钮<br>
    • 填写公司名称、职位、企业属性等信息<br>
    • 保存即可创建新记录<br><br>
    <strong>方式二：批量导入</strong><br>
    • 使用"批量导入"功能<br>
    • 支持从Excel表格一次性导入多条记录<br>
    • 提高录入效率`,
    category: 'usage'
  },
  {
    keywords: ['面试提醒', '设置提醒', '提醒功能', '如何设置提醒'],
    response: `⏰ <strong>面试提醒设置</strong><br><br>
    <strong>设置步骤：</strong><br>
    • 进入编辑投递记录界面<br>
    • 设置"面试时间"<br>
    • 开启"启用提醒"开关<br>
    • 系统将在指定时间提醒您<br><br>
    <strong>提醒中心：</strong><br>
    • 可在提醒中心查看所有启用提醒的记录<br>
    • 支持批量管理和个性化设置`,
    category: 'usage'
  },
  {
    keywords: ['导出', 'excel', '下载', '如何导出'],
    response: `📊 <strong>数据导出功能</strong><br><br>
    <strong>导出步骤：</strong><br>
    • 点击"导出Excel"按钮<br>
    • 选择需要的字段和筛选条件<br>
    • 生成下载任务<br><br>
    <strong>下载管理：</strong><br>
    • 在导出历史中查看生成进度<br>
    • 完成后点击下载链接获取文件<br>
    • 支持多种格式和自定义字段`,
    category: 'usage'
  },
  {
    keywords: ['时间线', '历史记录', '状态变化'],
    response: `📈 <strong>时间线功能</strong><br><br>
    <strong>功能特点：</strong><br>
    • 显示所有投递记录的完整时间线<br>
    • 支持多维度筛选和排序<br>
    • 分页显示，性能优化<br><br>
    <strong>使用技巧：</strong><br>
    • 使用筛选条件快速找到特定记录<br>
    • 点击记录查看详细状态变化历史<br>
    • 支持批量操作和状态更新`,
    category: 'usage'
  },

  // 数据分析类
  {
    keywords: ['求职进展', '投递情况', '数据统计', '我的情况'],
    response: `📊 <strong>您的求职数据分析</strong><br><br>
    根据当前数据：<br>
    • 总投递数：{totalApplications} 个岗位<br>
    • 待面试数：{pendingInterviews} 个<br><br>
    <strong>建议：</strong><br>
    • 保持投递频率，建议每周投递5-10个岗位<br>
    • 及时跟进面试邀请<br>
    • 利用数据统计页面分析投递效果`,
    category: 'analysis'
  },
  {
    keywords: ['成功率', '通过率', '效果分析'],
    response: `📈 <strong>投递效果分析</strong><br><br>
    您可以在统计页面查看：<br>
    • 各阶段通过率统计<br>
    • 不同公司类型的成功率<br>
    • 投递时间分布分析<br><br>
    <strong>优化建议：</strong><br>
    • 关注简历通过率较低的原因<br>
    • 分析面试失败的共同点<br>
    • 调整投递策略和目标岗位`,
    category: 'analysis'
  },

  // 功能介绍类
  {
    keywords: ['功能', '特性', '能做什么', '有什么功能'],
    response: `✨ <strong>JobView 核心功能</strong><br><br>
    <strong>🎯 看板管理</strong><br>
    • 拖拽式状态管理，直观易用<br>
    • 实时状态跟踪和历史记录<br><br>
    <strong>📅 提醒中心</strong><br>
    • 智能面试提醒<br>
    • 个性化提醒设置<br><br>
    <strong>📊 数据统计</strong><br>
    • 多维度数据分析<br>
    • 可视化图表展示<br><br>
    <strong>📤 数据导出</strong><br>
    • 支持Excel导出<br>
    • 自定义字段和筛选`,
    category: 'feature'
  },

  // 通用回复
  {
    keywords: ['你好', '您好', 'hello', 'hi'],
    response: `👋 主人您好！银月很高兴为您服务<br><br>
    我可以帮助您：<br>
    • 🎯 学习如何使用各项功能<br>
    • 📊 分析您的求职进展情况<br>
    • 💡 提供使用建议和求职技巧<br><br>
    主人有什么问题尽管问银月吧！`,
    category: 'general'
  },
  {
    keywords: ['谢谢', '感谢', 'thanks', '谢了'],
    response: `😊 主人客气了！银月很开心能帮助到您<br><br>
    如果还有其他问题，随时可以找银月哦！<br>
    祝主人求职顺利！🌟`,
    category: 'general'
  }
];

// 根据路由获取快速回复建议
export function getQuickReplies(routeName: string, messageCount: number): string[] {
  if (messageCount === 0) {
    // 首次对话根据页面提供建议
    switch (routeName) {
      case 'kanban':
        return ['如何使用看板功能？', '如何拖拽更新状态？', '我的求职进展如何？'];
      case 'timeline':
        return ['如何筛选时间线数据？', '如何导出Excel？', '我的投递效果怎样？'];
      case 'reminders':
        return ['如何设置面试提醒？', '提醒功能有哪些特点？'];
      case 'statistics':
        return ['如何分析我的数据？', '投递成功率如何提升？'];
      default:
        return ['如何添加投递记录？', '有什么功能？', '我的求职情况如何？'];
    }
  } else {
    // 对话中提供通用建议
    return [
      '如何提高投递成功率？',
      '还有什么功能吗？',
      '数据统计怎么看？'
    ];
  }
}

// 智能问答匹配
export async function getRobotResponse(
  userInput: string,
  context: ChatContext
): Promise<string> {
  const input = userInput.toLowerCase().trim();

  // 关键词匹配
  for (const item of knowledgeBase) {
    for (const keyword of item.keywords) {
      if (input.includes(keyword.toLowerCase())) {
        let response = item.response;

        // 替换上下文变量
        if (response.includes('{totalApplications}')) {
          response = response.replace('{totalApplications}', context.userContext.totalApplications.toString());
        }
        if (response.includes('{pendingInterviews}')) {
          response = response.replace('{pendingInterviews}', context.userContext.pendingInterviews.toString());
        }

        return response;
      }
    }
  }

  // 特殊问题处理
  if (input.includes('多少') || input.includes('数量')) {
    return `📊 <strong>您的投递数据</strong><br><br>
    • 总投递数：${context.userContext.totalApplications} 个岗位<br>
    • 面试进行中：${context.userContext.pendingInterviews} 个<br>
    • 完成投递：${context.userContext.totalApplications - context.userContext.pendingInterviews} 个<br><br>
    继续加油！每一次投递都是成功的开始 🚀`;
  }

  if (input.includes('建议') || input.includes('怎么办') || input.includes('如何')) {
    return `💡 <strong>求职建议</strong><br><br>
    基于您当前的投递情况，我建议：<br><br>
    <strong>1. 保持投递节奏</strong><br>
    • 每周投递5-10个合适的岗位<br>
    • 重质量胜过数量<br><br>
    <strong>2. 及时跟进</strong><br>
    • 使用提醒功能，不错过任何面试机会<br>
    • 主动联系HR了解进展<br><br>
    <strong>3. 总结经验</strong><br>
    • 分析投递数据，优化简历和策略<br>
    • 记录面试经验，持续改进`;
  }

  // 默认回复
  const defaultResponses = [
    `🤔 这个问题很有趣呢，主人！不过银月暂时还没有完美的答案。<br><br>
    您可以试试问银月：<br>
    • 如何使用某个功能<br>
    • 求职数据分析<br>
    • 使用技巧和建议<br><br>
    银月会继续学习，为主人提供更好的帮助！`,

    `😊 银月正在努力理解主人的问题！<br><br>
    目前银月比较擅长回答：<br>
    • 功能使用指导<br>
    • 数据分析和建议<br>
    • 求职技巧分享<br><br>
    主人可以换个方式问问银月哦～`,

    `🌟 虽然银月暂时无法完美回答这个问题，但银月很乐意帮助主人！<br><br>
    试试这些问题：<br>
    • "我的求职进展如何？"<br>
    • "如何使用看板功能？"<br>
    • "怎么设置面试提醒？"<br><br>
    让银月和主人一起让求职变得更高效！`
  ];

  return defaultResponses[Math.floor(Math.random() * defaultResponses.length)];
}

// 工具函数：分析用户投递数据
export function analyzeApplicationData(applications: any[]) {
  const total = applications.length;
  const byStatus = applications.reduce((acc, app) => {
    acc[app.status] = (acc[app.status] || 0) + 1;
    return acc;
  }, {});

  const pending = ['笔试中', '一面中', '二面中', '三面中', 'HR面中']
    .reduce((sum, status) => sum + (byStatus[status] || 0), 0);

  const rejected = ['简历挂', '笔试挂', '一面挂', '二面挂', '三面挂']
    .reduce((sum, status) => sum + (byStatus[status] || 0), 0);

  const success = byStatus['已入职'] || 0;

  return {
    total,
    pending,
    rejected,
    success,
    byStatus,
    successRate: total > 0 ? ((success / total) * 100).toFixed(1) : '0'
  };
}