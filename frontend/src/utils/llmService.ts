// LLM模型配置
export interface LLMConfig {
  provider: 'ollama' | 'openai' | 'qwen' | 'local';
  apiUrl: string;
  apiKey?: string;
  model: string;
  temperature?: number;
  maxTokens?: number;
}

// LLM响应类型
export interface LLMResponse {
  content: string;
  success: boolean;
  error?: string;
  source: 'knowledge_base' | 'llm_model';
}

// 默认配置
const defaultLLMConfig: LLMConfig = {
  provider: 'ollama', // 使用本地Ollama
  apiUrl: 'http://localhost:11434/api/generate',
  model: 'qwen2.5-coder:latest', // 使用您的qwen2.5-coder模型
  temperature: 0.7,
  maxTokens: 500
};

// 检测问题类型
export function classifyQuestion(question: string): 'job_related' | 'general' | 'personal' {
  const jobKeywords = [
    '求职', '投递', '面试', '简历', '工作', '职位', '公司', '看板', '提醒', '导出',
    '应聘', '招聘', 'hr', '薪资', '职场', '跳槽', '入职', '离职'
  ];

  const personalKeywords = [
    '心情', '情感', '压力', '焦虑', '烦恼', '开心', '难过', '困扰', '建议',
    '怎么办', '感觉', '情绪', '心理', '生活', '健康'
  ];

  const generalKeywords = [
    '天气', '时间', '日期', '新闻', '科技', '电影', '音乐', '美食', '旅游',
    '学习', '知识', '历史', '地理', '数学', '物理'
  ];

  const lowerQuestion = question.toLowerCase();

  // 检查求职相关关键词
  if (jobKeywords.some(keyword => lowerQuestion.includes(keyword))) {
    return 'job_related';
  }

  // 检查个人情感关键词
  if (personalKeywords.some(keyword => lowerQuestion.includes(keyword))) {
    return 'personal';
  }

  // 检查通用知识关键词
  if (generalKeywords.some(keyword => lowerQuestion.includes(keyword))) {
    return 'general';
  }

  // 默认分类为通用问题
  return 'general';
}

// 调用Ollama本地模型
async function callOllamaAPI(question: string, config: LLMConfig): Promise<LLMResponse> {
  try {
    const response = await fetch(config.apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        model: config.model,
        prompt: buildPrompt(question),
        stream: false,
        options: {
          temperature: config.temperature || 0.7,
          num_predict: config.maxTokens || 500
        }
      })
    });

    if (!response.ok) {
      throw new Error(`Ollama API error: ${response.status}`);
    }

    const data = await response.json();
    return {
      content: formatLLMResponse(data.response),
      success: true,
      source: 'llm_model'
    };
  } catch (error) {
    console.error('Ollama API调用失败:', error);
    return {
      content: '抱歉，银月的AI大脑暂时无法连接。主人可以稍后再试，或者问银月一些求职相关的问题。',
      success: false,
      error: error instanceof Error ? error.message : '未知错误',
      source: 'llm_model'
    };
  }
}

// 调用OpenAI API
async function callOpenAIAPI(question: string, config: LLMConfig): Promise<LLMResponse> {
  try {
    const response = await fetch('https://api.openai.com/v1/chat/completions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${config.apiKey}`
      },
      body: JSON.stringify({
        model: config.model || 'gpt-3.5-turbo',
        messages: [
          {
            role: 'system',
            content: '你是银月，用户的贴心小助手。请用亲切、可爱的语气回答问题，称呼用户为"主人"。回答要简洁、有用。'
          },
          {
            role: 'user',
            content: question
          }
        ],
        temperature: config.temperature || 0.7,
        max_tokens: config.maxTokens || 500
      })
    });

    if (!response.ok) {
      throw new Error(`OpenAI API error: ${response.status}`);
    }

    const data = await response.json();
    return {
      content: formatLLMResponse(data.choices[0].message.content),
      success: true,
      source: 'llm_model'
    };
  } catch (error) {
    console.error('OpenAI API调用失败:', error);
    return {
      content: '抱歉，银月的AI大脑暂时无法连接。主人可以稍后再试，或者问银月一些求职相关的问题。',
      success: false,
      error: error instanceof Error ? error.message : '未知错误',
      source: 'llm_model'
    };
  }
}

// 构建提示词
function buildPrompt(question: string): string {
  return `你是银月，一个可爱贴心的AI助手。用户是你的主人。请用温暖、可爱的语气回答以下问题，称呼用户为"主人"。

重要规则：
1. 保持亲切可爱的语气，像个贴心的小助手
2. 称呼用户为"主人"
3. 回答要简洁有用，不要过于啰嗦
4. 可以使用一些可爱的语气词，但不要过度
5. 如果涉及专业建议，要提醒主人寻求专业人士帮助
6. 对于心情问题，给予温暖的安慰和积极的建议

主人的问题：${question}

银月：`;
}

// 格式化LLM响应
function formatLLMResponse(content: string): string {
  // 移除不必要的前缀
  let formatted = content.trim();

  // 移除常见的AI助手前缀
  const prefixes = ['银月的回答：', '银月：', '回答：', '答：', 'A:', 'Answer:'];
  for (const prefix of prefixes) {
    if (formatted.startsWith(prefix)) {
      formatted = formatted.substring(prefix.length).trim();
    }
  }

  // 确保回答语气合适
  if (!formatted.includes('主人') && !formatted.startsWith('哎呀') && !formatted.startsWith('呀')) {
    // 如果回答比较正式，加上银月的语气
    if (formatted.length > 0) {
      formatted = `主人，${formatted}`;
    }
  }

  // 将换行转换为HTML格式
  formatted = formatted.replace(/\n/g, '<br>');

  return formatted;
}

// 增强的机器人回答函数
export async function getEnhancedRobotResponse(
  userInput: string,
  context: any,
  llmConfig?: Partial<LLMConfig>
): Promise<string> {
  const questionType = classifyQuestion(userInput);

  // 如果是求职相关问题，优先使用知识库
  if (questionType === 'job_related') {
    try {
      const knowledgeResponse = await getRobotResponse(userInput, context);
      // 如果知识库有满意的回答，直接返回
      if (!knowledgeResponse.includes('暂时无法回答') && !knowledgeResponse.includes('换个方式问')) {
        return knowledgeResponse;
      }
    } catch (error) {
      console.error('知识库查询失败:', error);
    }
  }

  // 对于通用问题或个人问题，尝试调用LLM
  const config = { ...defaultLLMConfig, ...llmConfig };

  let llmResponse: LLMResponse;

  // 优先尝试本地模型
  if (config.provider === 'ollama') {
    llmResponse = await callOllamaAPI(userInput, config);
  } else if (config.provider === 'openai') {
    llmResponse = await callOpenAIAPI(userInput, config);
  } else {
    // 降级到知识库回答
    return await getRobotResponse(userInput, context);
  }

  if (llmResponse.success) {
    return llmResponse.content;
  } else {
    // LLM失败时降级到知识库
    return await getRobotResponse(userInput, context);
  }
}

// 检查LLM服务可用性
export async function checkLLMAvailability(config?: Partial<LLMConfig>): Promise<boolean> {
  const testConfig = { ...defaultLLMConfig, ...config };

  try {
    if (testConfig.provider === 'ollama') {
      const response = await fetch(`${testConfig.apiUrl.replace('/api/generate', '/api/tags')}`, {
        method: 'GET',
        signal: AbortSignal.timeout(3000) // 3秒超时
      });
      return response.ok;
    }
    return false;
  } catch (error) {
    console.log('LLM服务不可用，将使用内置知识库');
    return false;
  }
}