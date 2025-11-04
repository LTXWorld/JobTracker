import { createApp, type App as VueApp } from 'vue'
import OfferCelebration from '../components/OfferCelebration.vue'

export const OFFER_CELEBRATION_EVENT = 'offer-celebration'

type CelebrationInstance = {
  app: VueApp<Element>
  container: HTMLElement
}

type CelebrationOptions = {
  title?: string
  subtitle?: string
  withConfetti?: boolean
  tone?: 'success' | 'info'
}

let currentInstance: CelebrationInstance | null = null

const cleanUp = () => {
  if (!currentInstance) return
  currentInstance.app.unmount()
  if (currentInstance.container.parentNode) {
    currentInstance.container.parentNode.removeChild(currentInstance.container)
  }
  currentInstance = null
}

const mountCelebration = (options: CelebrationOptions, hooks?: { onClosed?: () => void }) => {
  if (typeof window === 'undefined' || typeof document === 'undefined') return

  cleanUp()

  const container = document.createElement('div')
  container.setAttribute('data-offer-celebration', 'true')
  document.body.appendChild(container)

  const app = createApp(OfferCelebration, {
    ...options,
    onClosed: () => {
      cleanUp()
      hooks?.onClosed?.()
    }
  })

  currentInstance = { app, container }
  app.mount(container)
}

export const triggerOfferCelebration = () => {
  mountCelebration({
    title: '恭喜拿下 offer！',
    subtitle: '这段时间辛苦啦',
    withConfetti: true,
    tone: 'success'
  })

  window.dispatchEvent(new CustomEvent(OFFER_CELEBRATION_EVENT, {
    detail: {
      songTitle: '你离开了南京 从此没有人和我说话'
    }
  }))
}

export const triggerEncouragement = () => {
  mountCelebration({
    title: '别灰心，这是一场持久战！',
    subtitle: '相信自己，坚持就是胜利！',
    withConfetti: false,
    tone: 'info'
  })
}

export const destroyOfferCelebration = () => {
  cleanUp()
}
