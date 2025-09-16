// 用户指导相关的状态管理和本地存储
export interface UserGuideState {
  hasSeenGuide: boolean
  guideCompletedAt: string | null
  skipCount: number
}

export class UserGuideManager {
  private static readonly STORAGE_KEY = 'jobview_user_guide'

  // 获取用户指导状态
  static getGuideState(): UserGuideState {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY)
      if (stored) {
        return JSON.parse(stored)
      }
    } catch (error) {
      console.warn('获取用户指导状态失败:', error)
    }

    return {
      hasSeenGuide: false,
      guideCompletedAt: null,
      skipCount: 0
    }
  }

  // 保存用户指导状态
  static saveGuideState(state: UserGuideState): void {
    try {
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(state))
    } catch (error) {
      console.error('保存用户指导状态失败:', error)
    }
  }

  // 标记指导为已完成
  static markGuideCompleted(): void {
    const state = this.getGuideState()
    state.hasSeenGuide = true
    state.guideCompletedAt = new Date().toISOString()
    this.saveGuideState(state)
  }

  // 增加跳过次数
  static incrementSkipCount(): void {
    const state = this.getGuideState()
    state.skipCount += 1
    this.saveGuideState(state)
  }

  // 检查是否应该显示指导
  static shouldShowGuide(): boolean {
    const state = this.getGuideState()

    // 如果已经看过指导，则不显示
    if (state.hasSeenGuide) {
      return false
    }

    // 如果跳过次数超过3次，则不再显示
    if (state.skipCount >= 3) {
      return false
    }

    return true
  }

  // 重置指导状态（用于开发测试）
  static resetGuideState(): void {
    localStorage.removeItem(this.STORAGE_KEY)
    console.log('✅ 用户指导状态已重置')
  }

  // 手动触发显示指导（用于帮助按钮）
  static forceShowGuide(): boolean {
    return true
  }
}

// 在开发模式下将重置方法暴露到全局
if (typeof window !== 'undefined') {
  (window as any).resetUserGuide = UserGuideManager.resetGuideState
  (window as any).checkGuideState = UserGuideManager.getGuideState
}