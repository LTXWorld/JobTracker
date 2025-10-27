import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { InterviewExperienceCaptureContext, InterviewExperienceSubmission } from '../types'

export interface InterviewExperienceFlowResult {
  cancelled: boolean
  submission?: InterviewExperienceSubmission
}

type Resolver = (result: InterviewExperienceFlowResult) => void

interface PendingRequest {
  context: InterviewExperienceCaptureContext
  defaultValue: InterviewExperienceSubmission | null
  originRoute: string
  resolve: Resolver
}

export const useInterviewExperienceFlowStore = defineStore('interviewExperienceFlow', () => {
  const pendingRequest = ref<PendingRequest | null>(null)

  const startRequest = (payload: Omit<PendingRequest, 'resolve'>, resolve: Resolver) => {
    pendingRequest.value = {
      ...payload,
      resolve
    }
  }

  const completeRequest = (submission: InterviewExperienceSubmission) => {
    if (!pendingRequest.value) return
    const resolver = pendingRequest.value.resolve
    pendingRequest.value.resolve = () => {}
    pendingRequest.value = null
    resolver({ cancelled: false, submission })
  }

  const cancelRequest = () => {
    if (!pendingRequest.value) return
    const resolver = pendingRequest.value.resolve
    pendingRequest.value.resolve = () => {}
    pendingRequest.value = null
    resolver({ cancelled: true })
  }

  const clearRequest = () => {
    pendingRequest.value = null
  }

  return {
    pendingRequest,
    startRequest,
    completeRequest,
    cancelRequest,
    clearRequest
  }
})
