<template>
  <!-- 悬浮机器人头像按钮 -->
  <div class="robot-fab" @click="toggleChat" :class="{ active: chatOpen }">
    <div class="robot-avatar">
      <img :src="robotAvatar" alt="银月" />
      <div v-if="hasUnread" class="notification-dot"></div>
    </div>
  </div>

  <!-- 聊天窗口 -->
  <transition name="chat-slide">
    <div v-if="chatOpen" class="robot-chat-panel">
      <div class="chat-header">
        <div class="robot-info">
          <img :src="robotAvatar" alt="银月" class="header-avatar" />
          <div class="robot-name">
            <div class="name">银月</div>
            <div class="status">主人的贴心小助手</div>
          </div>
        </div>
        <div class="header-actions">
          <a-button v-if="showBackButton" type="text" size="small" @click="backToWelcome" title="返回首页">
            <HomeOutlined />
          </a-button>
          <a-button type="text" size="small" @click="minimizeChat" title="最小化">
            <MinusOutlined />
          </a-button>
          <a-button type="text" size="small" @click="toggleChat" title="关闭">
            <CloseOutlined />
          </a-button>
        </div>
      </div>

      <div class="chat-body" ref="chatBodyRef">
        <!-- 欢迎消息 -->
        <div v-if="showWelcomeScreen" class="welcome-section">
          <div class="welcome-text">
            <h3>你好主人！我是银月 ✨</h3>
            <p>很高兴为您服务，我可以帮助您：</p>
            <div class="feature-grid">
              <div class="feature-card" @click="sendQuickQuestion('如何使用看板功能？')">
                <div class="feature-icon">📋</div>
                <div class="feature-title">看板指导</div>
                <div class="feature-desc">学习拖拽管理求职状态</div>
              </div>
              <div class="feature-card" @click="sendQuickQuestion('我的求职进展如何？')">
                <div class="feature-icon">📊</div>
                <div class="feature-title">数据分析</div>
                <div class="feature-desc">查看投递统计和进展</div>
              </div>
              <div class="feature-card" @click="sendQuickQuestion('如何设置面试提醒？')">
                <div class="feature-icon">⏰</div>
                <div class="feature-title">提醒设置</div>
                <div class="feature-desc">管理面试时间提醒</div>
              </div>
              <div class="feature-card" @click="sendQuickQuestion('如何导出数据？')">
                <div class="feature-icon">📤</div>
                <div class="feature-title">数据导出</div>
                <div class="feature-desc">导出Excel报表</div>
              </div>
              <div class="feature-card" @click="sendQuickQuestion('有什么使用技巧？')">
                <div class="feature-icon">💡</div>
                <div class="feature-title">使用技巧</div>
                <div class="feature-desc">高效使用系统功能</div>
              </div>
              <div class="feature-card" @click="sendQuickQuestion('如何优化求职效果？')">
                <div class="feature-icon">🎯</div>
                <div class="feature-title">求职建议</div>
                <div class="feature-desc">提升投递成功率</div>
              </div>
            </div>
            <div class="welcome-footer">
              <p>💭 您也可以直接输入问题，银月会尽力为您解答</p>
              <!-- 欢迎页面底部输入框 -->
              <div class="welcome-input-container">
                <a-input
                  v-model:value="welcomeInput"
                  placeholder="主人，直接在这里问银月任何问题..."
                  @pressEnter="sendWelcomeMessage"
                  class="welcome-input"
                  size="large"
                />
                <a-button
                  type="primary"
                  @click="sendWelcomeMessage"
                  :disabled="!welcomeInput.trim()"
                  class="welcome-send-btn"
                  size="large"
                >
                  开始对话
                </a-button>
              </div>
            </div>
          </div>
        </div>

        <!-- 聊天消息 -->
        <div v-if="!showWelcomeScreen" class="messages-container">
          <div v-for="(msg, index) in messages" :key="index" class="message-wrapper">
            <div :class="['message', { 'user-message': msg.isUser, 'robot-message': !msg.isUser }]">
              <div v-if="!msg.isUser" class="message-avatar">
                <img :src="robotAvatar" alt="银月" />
              </div>
              <div class="message-content">
                <div class="message-bubble">
                  <div v-if="msg.isTyping" class="typing-indicator">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                  <div v-else class="message-text" v-html="msg.content"></div>
                </div>
                <div class="message-time">{{ msg.timestamp }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 快速回复建议 -->
      <div v-if="!showWelcomeScreen && quickReplies.length > 0" class="quick-replies">
        <div class="quick-replies-title">💡 快速提问：</div>
        <div class="quick-replies-list">
          <a-tag
            v-for="(reply, index) in quickReplies"
            :key="index"
            @click="sendQuickQuestion(reply)"
            class="quick-reply-tag"
          >
            {{ reply }}
          </a-tag>
        </div>
      </div>

      <!-- 输入区域 -->
      <div v-if="!showWelcomeScreen" class="chat-input">
        <div class="input-container">
          <a-input
            v-model:value="userInput"
            placeholder="主人，有什么问题想问银月吗？"
            @pressEnter="sendMessage"
            :disabled="isTyping"
            class="message-input"
          />
          <a-button
            type="primary"
            @click="sendMessage"
            :disabled="!userInput.trim() || isTyping"
            class="send-button"
          >
            <SendOutlined />
          </a-button>
        </div>
      </div>
    </div>
  </transition>

  <!-- 最小化状态 -->
  <transition name="mini-slide">
    <div v-if="isMinimized" class="robot-mini-panel" @click="restoreChat">
      <div class="mini-content">
        <img :src="robotAvatar" alt="银月" class="mini-avatar" />
        <div class="mini-text">
          <div class="mini-name">银月</div>
          <div class="mini-status">点击恢复对话</div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { CloseOutlined, MinusOutlined, SendOutlined, HomeOutlined } from '@ant-design/icons-vue'
import { useJobApplicationStore } from '@/stores/jobApplication'
import { getRobotResponse, getQuickReplies, type ChatMessage } from '@/utils/robotKnowledge'
import { getEnhancedRobotResponse, checkLLMAvailability } from '@/utils/llmService'
import robotAvatar1 from '@/assets/robot-avatar-1.jpg'

const route = useRoute()
const applicationStore = useJobApplicationStore()

// 状态管理
const chatOpen = ref(false)
const isMinimized = ref(false)
const userInput = ref('')
const welcomeInput = ref('') // 新增：欢迎页面输入框
const isTyping = ref(false)
const hasUnread = ref(true)
const chatBodyRef = ref<HTMLElement>()
const llmAvailable = ref(false) // 新增：LLM服务可用性

// 新增：控制欢迎页面和对话页面的切换
const showWelcomeScreen = computed(() => messages.length === 0)

// 新增：显示返回按钮的条件
const showBackButton = computed(() => !showWelcomeScreen.value)

// 机器人头像（随机选择）
const robotAvatar = computed(() => Math.random() > 0.5 ? robotAvatar1 : robotAvatar1)

// 消息列表
const messages = reactive<ChatMessage[]>([])

// 快速回复建议
const quickReplies = computed(() => {
  if (showWelcomeScreen.value) return []
  return getQuickReplies(route.name as string, messages.length)
})

// 切换聊天窗口
const toggleChat = () => {
  chatOpen.value = !chatOpen.value
  isMinimized.value = false
  if (chatOpen.value) {
    hasUnread.value = false
    // 不自动滚动到底部，让用户从顶部开始浏览
    // nextTick(() => {
    //   scrollToBottom()
    // })
  }
}

// 最小化聊天
const minimizeChat = () => {
  chatOpen.value = false
  isMinimized.value = true
}

// 恢复聊天
const restoreChat = () => {
  isMinimized.value = false
  chatOpen.value = true
  hasUnread.value = false
}

// 新增：返回欢迎页面
const backToWelcome = () => {
  messages.length = 0 // 清空消息
  userInput.value = '' // 清空输入
  welcomeInput.value = '' // 清空欢迎页面输入
  nextTick(() => {
    // 滚动到顶部
    if (chatBodyRef.value) {
      chatBodyRef.value.scrollTop = 0
    }
  })
}

// 新增：从欢迎页面发送消息
const sendWelcomeMessage = () => {
  const content = welcomeInput.value.trim()
  if (!content) return

  // 将欢迎页面的输入转移到普通输入框
  userInput.value = content
  welcomeInput.value = ''

  // 发送消息
  sendMessage()
}

// 发送消息
const sendMessage = async () => {
  const content = userInput.value.trim()
  if (!content || isTyping.value) return

  // 添加用户消息
  const userMessage: ChatMessage = {
    content,
    isUser: true,
    timestamp: new Date().toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    })
  }
  messages.push(userMessage)
  userInput.value = ''

  // 添加机器人打字状态
  const typingMessage: ChatMessage = {
    content: '',
    isUser: false,
    timestamp: new Date().toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    }),
    isTyping: true
  }
  messages.push(typingMessage)
  isTyping.value = true

  await nextTick()
  scrollToBottom()

  // 模拟思考时间
  await new Promise(resolve => setTimeout(resolve, 1000 + Math.random() * 2000))

  // 获取机器人回复
  try {
    const response = await getEnhancedRobotResponse(content, {
      currentRoute: route.name as string,
      applications: applicationStore.applications,
      userContext: {
        totalApplications: applicationStore.applications.length,
        pendingInterviews: applicationStore.applications.filter(app =>
          ['笔试中', '一面中', '二面中', '三面中', 'HR面中'].includes(app.status)
        ).length
      }
    })

    // 替换打字消息为实际回复
    const lastMessage = messages[messages.length - 1]
    lastMessage.content = response
    lastMessage.isTyping = false
  } catch (error) {
    // 错误处理
    const lastMessage = messages[messages.length - 1]
    lastMessage.content = '抱歉，银月暂时无法回答这个问题。主人可以稍后再试或联系技术支持。'
    lastMessage.isTyping = false
  }

  isTyping.value = false
  await nextTick()
  scrollToBottom()
}

// 快速提问
const sendQuickQuestion = (question: string) => {
  userInput.value = question
  sendMessage()
}

// 滚动到底部
const scrollToBottom = () => {
  if (chatBodyRef.value) {
    chatBodyRef.value.scrollTop = chatBodyRef.value.scrollHeight
  }
}

// 监听路由变化，显示未读提示
watch(() => route.name, () => {
  if (!chatOpen.value && !isMinimized.value) {
    hasUnread.value = true
  }
})

onMounted(async () => {
  // 初始化应用数据
  if (applicationStore.applications.length === 0) {
    applicationStore.fetchApplications()
  }

  // 检查LLM服务可用性
  llmAvailable.value = await checkLLMAvailability()
  if (llmAvailable.value) {
    console.log('✅ LLM服务可用，银月现在可以回答更多问题了！')
  } else {
    console.log('ℹ️ LLM服务不可用，银月将使用内置知识库回答问题')
  }
})
</script>

<style scoped>
.robot-fab {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 2000;
  cursor: pointer;
  transition: all 0.3s ease;
}

.robot-fab:hover {
  transform: scale(1.05);
}

.robot-fab.active {
  transform: scale(0.95);
}

.robot-avatar {
  position: relative;
  width: 60px;
  height: 60px;
  border-radius: 50%;
  overflow: hidden;
  border: 3px solid #1890ff;
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
  transition: all 0.3s ease;
}

.robot-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.notification-dot {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 12px;
  height: 12px;
  background: #ff4d4f;
  border-radius: 50%;
  border: 2px solid white;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

.robot-chat-panel {
  position: fixed;
  right: 24px;
  bottom: 100px;
  width: 380px;
  height: 650px; /* 增加高度以显示更多内容 */
  background: white;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  border: 1px solid #f0f0f0;
  display: flex;
  flex-direction: column;
  z-index: 2000;
  overflow: hidden;
}

.chat-header {
  padding: 16px;
  background: linear-gradient(135deg, #1890ff, #69c0ff);
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.robot-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.8);
}

.robot-name .name {
  font-weight: 600;
  font-size: 16px;
}

.robot-name .status {
  font-size: 12px;
  opacity: 0.9;
}

.header-actions {
  display: flex;
  gap: 4px;
}

.header-actions .ant-btn {
  color: white;
  border: none;
}

.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px; /* 减少内边距 */
  background: #fafafa;
}

.welcome-section {
  text-align: center;
  padding: 16px 0; /* 减少上下内边距 */
}

.welcome-text h3 {
  color: #262626;
  margin-bottom: 8px;
}

.welcome-text p {
  color: #666;
  margin-bottom: 16px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px; /* 减少卡片间距 */
  margin-bottom: 12px; /* 减少底部间距 */
}

.feature-card {
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 8px; /* 减少圆角 */
  padding: 12px; /* 减少内边距 */
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
}

.feature-card:hover {
  border-color: #1890ff;
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.15);
  transform: translateY(-2px);
}

.feature-icon {
  font-size: 24px; /* 减小图标大小 */
  margin-bottom: 6px; /* 减少底部间距 */
  height: 32px; /* 减少高度 */
  display: flex;
  align-items: center;
  justify-content: center;
}

.feature-title {
  font-weight: 600;
  color: #262626;
  margin-bottom: 3px; /* 减少底部间距 */
  font-size: 13px; /* 减小字体 */
}

.feature-desc {
  color: #8c8c8c;
  font-size: 11px; /* 减小字体 */
  line-height: 1.3; /* 减少行高 */
}

.welcome-footer {
  margin-top: 12px; /* 减少顶部间距 */
  padding-top: 12px; /* 减少顶部内边距 */
  border-top: 1px solid #f0f0f0;
}

.welcome-footer p {
  color: #8c8c8c;
  font-size: 12px; /* 减小字体 */
  margin: 0 0 12px 0; /* 减少底部间距 */
  text-align: center;
}

.welcome-input-container {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 8px; /* 减少顶部间距 */
}

.welcome-input {
  flex: 1;
  border-radius: 24px;
  border: 2px solid #f0f0f0;
  transition: all 0.3s ease;
}

.welcome-input:focus {
  border-color: #1890ff;
  box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}

.welcome-input:hover {
  border-color: #69c0ff;
}

.welcome-send-btn {
  border-radius: 24px;
  height: 40px;
  padding: 0 20px;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.3);
  transition: all 0.3s ease;
}

.welcome-send-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.4);
}

.feature-tags {
  display: flex;
  justify-content: center;
  gap: 8px;
  flex-wrap: wrap;
}

.feature-tags .ant-tag {
  cursor: pointer;
  transition: all 0.2s;
}

.feature-tags .ant-tag:hover {
  transform: translateY(-1px);
}

.messages-container {
  margin-top: 16px;
}

.message-wrapper {
  margin-bottom: 16px;
}

.message {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.user-message {
  flex-direction: row-reverse;
}

.message-avatar img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
}

.message-content {
  max-width: 70%;
}

.user-message .message-content {
  text-align: right;
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 18px;
  margin-bottom: 4px;
  word-break: break-word;
}

.robot-message .message-bubble {
  background: white;
  border: 1px solid #f0f0f0;
  border-top-left-radius: 6px;
}

.user-message .message-bubble {
  background: #1890ff;
  color: white;
  border-top-right-radius: 6px;
}

.message-time {
  font-size: 11px;
  color: #999;
  padding: 0 4px;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  align-items: center;
}

.typing-indicator span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ccc;
  animation: typing 1.4s infinite;
}

.typing-indicator span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-indicator span:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); }
  30% { transform: translateY(-10px); }
}

.quick-replies {
  padding: 12px 16px;
  background: white;
  border-top: 1px solid #f0f0f0;
}

.quick-replies-title {
  font-size: 12px;
  color: #666;
  margin-bottom: 8px;
}

.quick-replies-list {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.quick-reply-tag {
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #d9d9d9;
  background: #fafafa;
}

.quick-reply-tag:hover {
  border-color: #1890ff;
  color: #1890ff;
}

.chat-input {
  padding: 16px;
  background: white;
  border-top: 1px solid #f0f0f0;
}

.input-container {
  display: flex;
  gap: 8px;
  align-items: center;
}

.message-input {
  flex: 1;
  border-radius: 20px;
}

.send-button {
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.robot-mini-panel {
  position: fixed;
  right: 24px;
  bottom: 100px;
  background: white;
  border-radius: 12px;
  padding: 12px 16px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  border: 1px solid #f0f0f0;
  cursor: pointer;
  z-index: 2000;
  transition: all 0.2s;
}

.robot-mini-panel:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
}

.mini-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.mini-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 2px solid #1890ff;
}

.mini-name {
  font-weight: 600;
  color: #262626;
}

.mini-status {
  font-size: 12px;
  color: #666;
}

.chat-slide-enter-active, .chat-slide-leave-active {
  transition: all 0.3s ease;
}

.chat-slide-enter-from, .chat-slide-leave-to {
  transform: translateY(20px) scale(0.95);
  opacity: 0;
}

.mini-slide-enter-active, .mini-slide-leave-active {
  transition: all 0.2s ease;
}

.mini-slide-enter-from, .mini-slide-leave-to {
  transform: translateX(20px);
  opacity: 0;
}

/* 滚动条样式 */
.chat-body::-webkit-scrollbar {
  width: 4px;
}

.chat-body::-webkit-scrollbar-track {
  background: transparent;
}

.chat-body::-webkit-scrollbar-thumb {
  background: #d9d9d9;
  border-radius: 2px;
}

.chat-body::-webkit-scrollbar-thumb:hover {
  background: #bfbfbf;
}
</style>