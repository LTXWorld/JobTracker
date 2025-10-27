<template>
  <div class="interview-experience-page">
    <a-page-header
      title="记录面试体验"
      sub-title="请完善本轮面试的反馈信息，帮助后续复盘"
      @back="handleCancel"
    />

    <div v-if="!currentRequest" class="empty-state">
      <a-result
        status="warning"
        title="未找到待记录的面试体验"
        sub-title="请从岗位状态流转入口重新进入此页面。"
      >
        <template #extra>
          <a-button type="primary" @click="handleCancel">返回</a-button>
        </template>
      </a-result>
    </div>

    <a-card v-else class="experience-card" :bordered="false">
      <a-alert
        type="info"
        show-icon
        :message="`状态将从「${fromStatus}」推进至「${toStatus}」，请如实记录本轮面试体验。`"
      />

      <div class="status-overview">
        <div class="status-bridge">
          <span class="status-tag">{{ fromStatus }}</span>
          <span class="status-arrow">→</span>
          <span class="status-tag highlight">{{ toStatus }}</span>
        </div>
      </div>

      <div class="mode-block">
        <div class="mode-label">反馈方式</div>
        <a-radio-group v-model:value="mode" button-style="solid" class="mode-switcher">
          <a-radio-button value="record">填写反馈</a-radio-button>
          <a-radio-button value="skip">跳过反馈</a-radio-button>
        </a-radio-group>
      </div>

      <div v-if="mode === 'record'" class="record-section">
        <h3>体验评分</h3>
        <div class="rating-grid">
          <a-card
            v-for="option in ratingOptions"
            :key="option.value"
            :hoverable="true"
            class="rating-card"
            :class="{ selected: rating === option.value }"
            @click="selectRating(option.value)"
          >
            <div class="rating-content">
              <span class="emoji">{{ option.emoji }}</span>
              <div class="text">
                <div class="title">{{ option.label }}</div>
                <div class="desc">{{ option.description }}</div>
              </div>
            </div>
          </a-card>
        </div>
        <div class="rating-hint">请选择最符合实际的整体体验，辅助后续复盘。</div>

        <h3>补充备注</h3>
        <a-textarea
          v-model:value="note"
          :rows="4"
          placeholder="可记录关键对话、亮点或后续跟进计划"
          allow-clear
        />
      </div>

      <div v-else class="skip-section">
        <h3>跳过说明</h3>
        <a-input
          v-model:value="skipReason"
          placeholder="例如：暂无时间整理、体验信息不足、待同伴同步等"
          allow-clear
        />

        <h3>补充备注</h3>
        <a-textarea
          v-model:value="note"
          :rows="3"
          placeholder="如需后续补填，可在此记录提醒或背景信息"
          allow-clear
        />
      </div>

      <div class="actions">
        <a-button @click="handleCancel">取消</a-button>
        <a-button
          v-if="mode === 'skip'"
          type="default"
          @click="handleSkip"
        >
          跳过并继续
        </a-button>
        <a-button
          v-else
          type="primary"
          :disabled="!rating"
          @click="handleSubmit"
        >
          提交反馈
        </a-button>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, onBeforeRouteLeave } from 'vue-router'
import { message } from 'ant-design-vue'
import { useInterviewExperienceFlowStore } from '../stores/interviewExperienceFlow'
import type { InterviewExperienceSubmission, InterviewExperienceRating } from '../types'

type Mode = 'record' | 'skip'

const router = useRouter()
const flowStore = useInterviewExperienceFlowStore()

const currentRequest = computed(() => flowStore.pendingRequest)

const mode = ref<Mode>('record')
const rating = ref<InterviewExperienceRating | ''>('')
const note = ref('')
const skipReason = ref('')

const ratingOptions: Array<{ value: InterviewExperienceRating; label: string; description: string; emoji: string }> = [
  { value: 'good', label: '好', description: '体验顺畅、交流积极，值得复盘', emoji: '😊' },
  { value: 'average', label: '一般', description: '中规中矩，可结合后续面试评估', emoji: '🙂' },
  { value: 'bad', label: '差', description: '反馈欠佳或存在问题，需要调整策略', emoji: '😞' }
]

const fromStatus = computed(() => currentRequest.value?.context.fromStatus ?? '')
const toStatus = computed(() => currentRequest.value?.context.toStatus ?? '')

const resetForm = (preset?: InterviewExperienceSubmission | null) => {
  if (preset) {
    mode.value = preset.skip ? 'skip' : 'record'
    rating.value = !preset.skip && preset.rating ? preset.rating : ''
    note.value = preset.note ?? ''
    skipReason.value = preset.skip_reason ?? ''
  } else {
    mode.value = 'record'
    rating.value = ''
    note.value = ''
    skipReason.value = ''
  }
}

watch(currentRequest, (req) => {
  if (req) {
    resetForm(req.defaultValue)
  } else {
    resetForm()
  }
}, { immediate: true })

const normalizeText = (value: string) => {
  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : ''
}

const selectRating = (value: InterviewExperienceRating) => {
  rating.value = value
}

const navigateBack = async (fallback?: string) => {
  const target = currentRequest.value?.originRoute || fallback
  if (target) {
    try {
      await router.push(target)
      return
    } catch (error) {
      console.warn('跳转原页面失败，将尝试返回上一页', error)
    }
  }
  router.back()
}

const handleSubmit = () => {
  const request = currentRequest.value
  if (!request) {
    message.warning('当前没有待记录的面试体验')
    return
  }

  if (!rating.value) {
    message.warning('请选择本轮面试体验的评分')
    return
  }

  const payload: InterviewExperienceSubmission = {
    skip: false,
    rating: rating.value,
    note: normalizeText(note.value) || undefined
  }

  flowStore.completeRequest(payload)
  navigateBack(request.originRoute)
}

const handleSkip = () => {
  const request = currentRequest.value
  if (!request) {
    message.warning('当前没有待记录的面试体验')
    return
  }

  const payload: InterviewExperienceSubmission = {
    skip: true,
    skip_reason: normalizeText(skipReason.value) || undefined,
    note: normalizeText(note.value) || undefined
  }

  flowStore.completeRequest(payload)
  navigateBack(request.originRoute)
}

const handleCancel = () => {
  if (currentRequest.value) {
    flowStore.cancelRequest()
  }
  navigateBack('/kanban')
}

onBeforeRouteLeave(() => {
  if (flowStore.pendingRequest) {
    flowStore.cancelRequest()
  }
})
</script>

<style scoped>
.interview-experience-page {
  max-width: 780px;
  margin: 0 auto;
  padding: 24px 0 80px;
}

.empty-state {
  margin-top: 48px;
}

.experience-card {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.status-overview {
  padding: 12px 16px;
  background: #f5f9ff;
  border-radius: 12px;
  display: inline-block;
  box-shadow: inset 0 0 0 1px rgba(22, 119, 255, 0.12);
  margin-bottom: 20px;
}

.status-bridge {
  display: flex;
  align-items: center;
  gap: 14px;
  font-weight: 600;
  color: #1f2937;
}

.status-tag {
  padding: 4px 10px;
  background: #ffffff;
  border-radius: 999px;
  min-width: 80px;
  text-align: center;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.06);
}

.status-tag.highlight {
  background: linear-gradient(135deg, rgba(22, 119, 255, 0.22), rgba(22, 119, 255, 0.08));
  color: #1056d6;
}

.status-arrow {
  color: #7990b3;
  font-size: 18px;
}

.mode-block {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  margin: 28px 0 32px;
}

.mode-label {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
}

.mode-switcher {
  display: inline-flex;
  gap: 0;
  border-radius: 999px;
  background: #f2f6ff;
  padding: 6px;
}

.mode-switcher :deep(.ant-radio-button-wrapper) {
  min-width: 140px;
  text-align: center;
  padding: 0 24px;
  font-weight: 600;
  height: 44px;
  line-height: 42px;
  border-radius: 999px !important;
  border: none;
  color: #4b5563;
  background: transparent;
  transition: all 0.2s ease;
}

.mode-switcher :deep(.ant-radio-button-wrapper:not(:first-child)) {
  margin-left: 6px;
}

.mode-switcher :deep(.ant-radio-button-wrapper::before) {
  display: none;
}

.mode-switcher :deep(.ant-radio-button-wrapper-checked) {
  background: linear-gradient(135deg, #1677ff, #1a8fff);
  color: #ffffff;
  box-shadow: 0 10px 20px rgba(22, 119, 255, 0.25);
}

.mode-switcher :deep(.ant-radio-button-wrapper:hover) {
  color: #1677ff;
}

.record-section,
.skip-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.rating-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.rating-card {
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
  background: #ffffff;
}

.rating-card :deep(.ant-card-body) {
  padding: 16px;
}

.rating-card .rating-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.rating-card .emoji {
  font-size: 28px;
}

.rating-card .text .title {
  font-weight: 600;
  font-size: 16px;
  margin-bottom: 4px;
}

.rating-card .text .desc {
  font-size: 12px;
  color: #6b7280;
}

.rating-card:hover {
  border-color: #1677ff;
  box-shadow: 0 8px 18px rgba(22, 119, 255, 0.12);
  transform: translateY(-3px);
}

.rating-card.selected {
  border: 2px solid #1677ff;
  box-shadow: 0 10px 22px rgba(22, 119, 255, 0.18);
  background: linear-gradient(135deg, rgba(22, 119, 255, 0.16), rgba(22, 119, 255, 0.04));
  transform: translateY(-6px);
}

.rating-card.selected .text .title {
  color: #1677ff;
}

.rating-hint {
  color: #6b7280;
  font-size: 12px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

@media (max-width: 768px) {
  .interview-experience-page {
    padding: 16px 12px 60px;
  }

  .mode-switcher {
    width: 100%;
    justify-content: space-between;
  }

  .mode-switcher :deep(.ant-radio-button-wrapper) {
    flex: 1;
    min-width: auto;
  }
}
</style>
