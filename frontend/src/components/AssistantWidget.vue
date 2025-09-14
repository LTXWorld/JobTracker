<template>
  <!-- 悬浮启动按钮 -->
  <div class="assistant-fab" @click="toggle" :title="open ? '关闭助手' : '打开助手'">
    <a-badge :count="unreadCount" :offset="[-2,10]">
      <a-button type="primary" shape="circle" size="large">
        <template #icon>
          <QuestionCircleOutlined />
        </template>
      </a-button>
    </a-badge>
  </div>

  <!-- 右下角助手面板 -->
  <transition name="assistant-slide">
    <div v-if="open" class="assistant-panel">
      <a-card :bordered="true" class="assistant-card" :bodyStyle="{ padding: '12px' }">
        <template #title>
          <span>
            <RobotOutlined /> 使用助手
          </span>
        </template>
        <template #extra>
          <a-space>
            <a-tooltip title="收起">
              <a-button type="text" size="small" @click.stop="toggle">
                <CloseOutlined />
              </a-button>
            </a-tooltip>
          </a-space>
        </template>

        <!-- 建议问题（根据路由上下文） -->
        <div class="assistant-section">
          <div class="section-title">
            <BulbOutlined /> 快速指引
          </div>
          <a-space wrap>
            <a-tag v-for="q in suggestedQuestions" :key="q.key" @click="select(q)" class="q-chip" color="blue">
              {{ q.title }}
            </a-tag>
          </a-space>
        </div>

        <a-divider style="margin: 10px 0" />

        <!-- 常见问题 -->
        <div class="assistant-section">
          <div class="section-title">
            <QuestionCircleOutlined /> 常见问题
          </div>
          <a-space wrap>
            <a-tag v-for="q in commonQuestions" :key="q.key" @click="select(q)" class="q-chip">
              {{ q.title }}
            </a-tag>
          </a-space>
        </div>

        <a-divider style="margin: 10px 0" />

        <!-- 答案展示 -->
        <div class="assistant-answer" v-if="active">
          <div class="answer-title">
            <span class="dot"></span>
            {{ active.title }}
          </div>
          <div class="answer-content">
            <ul>
              <li v-for="(line, idx) in active.answer" :key="idx">{{ line }}</li>
            </ul>
          </div>
        </div>
        <div class="assistant-answer" v-else>
          <div class="answer-title">
            <span class="dot"></span>
            欢迎使用 JobView 助手
          </div>
          <div class="answer-content muted">
            选择上方的问题快速了解如何使用当前页面功能。
          </div>
        </div>
      </a-card>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { 
  QuestionCircleOutlined,
  RobotOutlined,
  CloseOutlined,
  BulbOutlined
} from '@ant-design/icons-vue'

type QA = { key: string; title: string; answer: string[] }

const route = useRoute()
const open = ref(false)
const unreadCount = ref<number>(0)
const active = ref<QA | null>(null)

const baseQuestions: Record<string, QA> = reactive({
  view_status: {
    key: 'view_status',
    title: '点击岗位卡片查看状态变化',
    answer: [
      '在看板视图中，点击岗位卡片可打开“状态详情”，查看该岗位的历史状态变化时间线。',
      '也可使用卡片右上角的更多按钮，进入“状态详情”或“编辑/删除”。'
    ]
  },
  drag_update: {
    key: 'drag_update',
    title: '拖拽岗位卡片以快速更新状态',
    answer: [
      '按住岗位卡片并拖拽至目标列即可快速更新该岗位的状态。',
      '回退到更早阶段时需要确认，部分终态回退需要填写备注。'
    ]
  },
  create_job: {
    key: 'create_job',
    title: '如何添加投递记录',
    answer: [
      '点击“添加投递”按钮，填写公司、职位、企业属性等信息并保存。',
      '也可以使用“批量导入”从表格一次性导入多条记录。'
    ]
  },
  reminders: {
    key: 'reminders',
    title: '如何设置面试提醒',
    answer: [
      '在编辑弹窗中设置“面试时间”并开启“启用提醒”，系统将在指定时间提醒你。',
      '提醒中心可集中查看所有启用提醒的记录。'
    ]
  },
  export: {
    key: 'export',
    title: '如何导出 Excel',
    answer: [
      '点击“导出Excel”按钮，可选择字段与筛选条件，生成下载任务。',
      '下载完成后可在导出历史中查看并下载文件。'
    ]
  },
  search: {
    key: 'search',
    title: '如何快速定位公司/职位',
    answer: [
      '在搜索框输入公司或职位关键字并回车，页面会定位到匹配的卡片并闪烁提示。',
      '再次回车可定位到下一个匹配项。'
    ]
  }
})

// 根据当前页面提供首推指引
const suggestedQuestions = computed<QA[]>(() => {
  const name = route.name as string
  if (name === 'kanban') {
    return [baseQuestions.drag_update, baseQuestions.view_status, baseQuestions.create_job, baseQuestions.search]
  }
  if (name === 'timeline') {
    return [baseQuestions.create_job, baseQuestions.search, baseQuestions.reminders, baseQuestions.export]
  }
  return [baseQuestions.create_job, baseQuestions.reminders, baseQuestions.export]
})

const commonQuestions = computed<QA[]>(() => [
  baseQuestions.view_status,
  baseQuestions.drag_update,
  baseQuestions.reminders,
  baseQuestions.export
])

const toggle = () => {
  open.value = !open.value
  if (open.value) unreadCount.value = 0
}

const select = (q: QA) => {
  active.value = q
}

// 首次进入某些页面时给一个未读提示小红点
watch(
  () => route.name,
  () => {
    if (!open.value) unreadCount.value = 1
    active.value = null
  }
)
</script>

<style scoped>
.assistant-fab {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 2000;
}

.assistant-panel {
  position: fixed;
  right: 24px;
  bottom: 88px;
  width: 340px;
  z-index: 2000;
}

.assistant-card :deep(.ant-card-head-title) {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-color);
}

.assistant-section { margin-bottom: 6px; }
.section-title { font-weight: 600; margin-bottom: 6px; display:flex; align-items:center; gap:6px; color: var(--text-color); }
.q-chip { cursor: pointer; user-select: none; }
.q-chip:not(.ant-tag-blue) { background: var(--bg-muted) !important; color: var(--text-color) !important; border-color: var(--border-color) !important; }
.q-chip:hover { opacity: 0.9; }

.assistant-answer { margin-top: 6px; }
.answer-title { font-weight: 600; display:flex; align-items:center; gap:8px; color: var(--text-color); }
.answer-title .dot { width:6px; height:6px; border-radius: 50%; background:#1890ff; display:inline-block; }
.answer-content { margin-top: 6px; font-size: 13px; line-height: 1.6; color: var(--text-color); }
.answer-content.muted { color: var(--text-secondary); }

.assistant-slide-enter-active, .assistant-slide-leave-active {
  transition: all .2s ease;
}
.assistant-slide-enter-from, .assistant-slide-leave-to {
  transform: translateY(10px);
  opacity: 0;
}
</style>
