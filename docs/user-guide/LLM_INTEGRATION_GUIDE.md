# 银月智能助手 - LLM大模型集成指南

## ✨ 功能概述

银月现在支持三种回答模式：
1. **内置知识库** - 回答求职相关问题（快速、准确）
2. **LLM大模型** - 回答通用问题和情感咨询（智能、全面）
3. **混合模式** - 自动选择最适合的回答方式

## 🤖 支持的问题类型

### 求职相关（内置知识库）
- "如何使用看板功能？"
- "我的求职进展如何？"
- "如何设置面试提醒？"
- "怎么导出数据？"

### 通用知识（LLM模型）
- "今天天气如何？"
- "北京的历史有多久？"
- "如何学习编程？"
- "推荐几部好电影"

### 情感咨询（LLM模型）
- "我最近心情不好，我该怎么办？"
- "面试失败了，感觉很沮丧"
- "工作压力很大，怎么缓解？"
- "如何保持积极心态？"

## 🔧 配置LLM服务

### 方案一：本地Ollama（推荐）

1. **安装Ollama**
```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows
# 从 https://ollama.ai 下载安装包
```

2. **启动Ollama服务**
```bash
ollama serve
```

3. **下载轻量级模型**
```bash
# 推荐使用Qwen2.5 3B模型（体积小，速度快）
ollama pull qwen2.5:3b

# 或者使用Llama3 8B模型（效果更好，需要更多内存）
ollama pull llama3:8b
```

4. **测试服务**
```bash
curl http://localhost:11434/api/tags
```

### 方案二：云端API

修改 `llmService.ts` 中的配置：

```typescript
const defaultLLMConfig: LLMConfig = {
  provider: 'openai',
  apiUrl: 'https://api.openai.com/v1/chat/completions',
  apiKey: 'your-api-key-here',
  model: 'gpt-3.5-turbo',
  temperature: 0.7,
  maxTokens: 500
};
```

## 📊 智能问题分类

银月会自动分析问题类型：

- **求职相关** → 使用内置知识库（快速准确）
- **通用知识** → 调用LLM模型（知识丰富）
- **情感咨询** → 调用LLM模型（温暖贴心）

## 🎯 使用体验

1. **欢迎页面** - 点击银月头像进入
2. **功能卡片** - 6个常用功能快速入口
3. **底部输入框** - 直接输入任何问题
4. **对话界面** - 支持连续对话和快速回复
5. **返回首页** - 随时通过🏠按钮返回

## 🔍 故障排除

### LLM服务不可用
- 检查Ollama是否启动：`curl http://localhost:11434/api/tags`
- 检查模型是否下载：`ollama list`
- 查看控制台日志获取详细错误信息

### 回答质量不佳
- 调整 `temperature` 参数（0.1-1.0）
- 尝试不同的模型（qwen2.5, llama3等）
- 优化提示词内容

### 响应速度慢
- 使用更小的模型（如qwen2.5:3b）
- 减少 `maxTokens` 限制
- 确保有足够的GPU/CPU资源

## 🌟 最佳实践

1. **本地部署** - 使用Ollama + Qwen2.5获得最佳平衡
2. **混合使用** - 让银月自动选择最适合的回答方式
3. **个性化** - 根据需要调整模型参数
4. **隐私保护** - 本地模型确保数据安全

## 💡 扩展建议

- 支持语音输入输出
- 添加更多专业领域知识库
- 集成图片识别能力
- 支持多轮对话记忆
- 个性化学习用户偏好

现在银月不仅是您的求职助手，更是您的全能AI伙伴！🚀