import { StatusHelper, type ApplicationStatus, type InterviewExperienceCaptureContext } from '../types'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { useInterviewExperienceFlowStore, type InterviewExperienceFlowResult } from '../stores/interviewExperienceFlow'
import { useRouter, useRoute } from 'vue-router'

export const useInterviewExperienceCapture = () => {
  const statusTrackingStore = useStatusTrackingStore()
  const flowStore = useInterviewExperienceFlowStore()
  const router = useRouter()
  const route = useRoute()

  const shouldCapture = (status: ApplicationStatus): boolean => {
    return StatusHelper.shouldCaptureInterviewExperience(status)
  }

  const requestInterviewExperience = async (
    context: InterviewExperienceCaptureContext
  ): Promise<InterviewExperienceFlowResult> => {
    if (!shouldCapture(context.fromStatus)) {
      return { cancelled: false }
    }

    const defaultValue = statusTrackingStore.getCachedInterviewExperience()
    const originRoute = route.fullPath

    return await new Promise((resolve) => {
      flowStore.startRequest(
        {
          context,
          defaultValue: defaultValue ?? null,
          originRoute
        },
        (result) => {
          if (!result.cancelled && result.submission) {
            statusTrackingStore.cacheInterviewExperience(result.submission)
          }
          resolve(result)
        }
      )

      router.push({ name: 'interview-experience' }).catch((error) => {
        console.error('跳转面试体验页面失败:', error)
        flowStore.cancelRequest()
      })
    })
  }

  return {
    shouldCaptureInterviewExperience: shouldCapture,
    requestInterviewExperience
  }
}
