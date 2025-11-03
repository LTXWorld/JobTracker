import { createApp, type App as VueApp } from 'vue'
import OfferCelebration from '../components/OfferCelebration.vue'

export const OFFER_CELEBRATION_EVENT = 'offer-celebration'

type CelebrationInstance = {
  app: VueApp<Element>
  container: HTMLElement
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

export const triggerOfferCelebration = () => {
  if (typeof window === 'undefined' || typeof document === 'undefined') return

  // 若正在展示中，先清理，避免多个实例叠加
  cleanUp()

  const container = document.createElement('div')
  container.setAttribute('data-offer-celebration', 'true')
  document.body.appendChild(container)

  const app = createApp(OfferCelebration, {
    onClosed: () => cleanUp()
  })

  currentInstance = { app, container }
  app.mount(container)

  window.dispatchEvent(new CustomEvent(OFFER_CELEBRATION_EVENT, {
    detail: {
      songTitle: '你离开了南京 从此没有人和我说话'
    }
  }))
}

export const destroyOfferCelebration = () => {
  cleanUp()
}
