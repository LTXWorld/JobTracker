<template>
  <transition name="offer-celebration-fade" @after-leave="handleAfterLeave">
    <div v-if="visible" class="offer-celebration-overlay" @click="closeNow">
      <div class="offer-celebration-card" :class="variantClass">
        <div class="offer-celebration-title">{{ titleText }}</div>
        <div class="offer-celebration-subtitle" v-if="subtitleText">{{ subtitleText }}</div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import confetti from 'canvas-confetti'

const props = withDefaults(defineProps<{
  title?: string
  subtitle?: string
  withConfetti?: boolean
  tone?: 'success' | 'info'
}>(), {
  title: '恭喜拿下 offer！',
  subtitle: '这段时间辛苦啦',
  withConfetti: true,
  tone: 'success'
})

const emit = defineEmits<{
  (e: 'closed'): void
}>()

const variantClass = computed(() => props.tone === 'info' ? 'offer-celebration-card--info' : '')
const titleText = computed(() => props.title)
const subtitleText = computed(() => props.subtitle)

const visible = ref(false)
let hideTimer: number | null = null
let confettiTimer: number | null = null

const launchFireworks = () => {
  if (!props.withConfetti) return

  const duration = 4000
  const animationEnd = Date.now() + duration
  const defaults = {
    startVelocity: 45,
    spread: 360,
    ticks: 90,
    gravity: 0.9,
    scalar: 1.2,
    zIndex: 9999
  }

  const randomInRange = (min: number, max: number) => Math.random() * (max - min) + min

  confettiTimer = window.setInterval(() => {
    const timeLeft = animationEnd - Date.now()
    if (timeLeft <= 0) {
      if (confettiTimer !== null) {
        window.clearInterval(confettiTimer)
        confettiTimer = null
      }
      return
    }

    const particleCount = Math.round(80 * (timeLeft / duration))
    void confetti({
      ...defaults,
      particleCount,
      origin: { x: randomInRange(0.2, 0.4), y: randomInRange(0, 0.2) }
    })
    void confetti({
      ...defaults,
      particleCount,
      origin: { x: randomInRange(0.6, 0.8), y: randomInRange(0, 0.2) }
    })
  }, 250)
}

const handleAfterLeave = () => {
  emit('closed')
}

const closeNow = () => {
  if (!visible.value) return
  visible.value = false
  if (hideTimer !== null) {
    window.clearTimeout(hideTimer)
    hideTimer = null
  }
  if (confettiTimer !== null) {
    window.clearInterval(confettiTimer)
    confettiTimer = null
  }
}

onMounted(() => {
  visible.value = true
  launchFireworks()
  hideTimer = window.setTimeout(() => {
    closeNow()
  }, 4800)
})

onBeforeUnmount(() => {
  if (hideTimer !== null) {
    window.clearTimeout(hideTimer)
    hideTimer = null
  }

  if (confettiTimer !== null) {
    window.clearInterval(confettiTimer)
    confettiTimer = null
  }
})
</script>

<style scoped>
.offer-celebration-overlay {
  position: fixed;
  inset: 0;
  z-index: 9998;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(17, 24, 39, 0.55);
  backdrop-filter: blur(4px);
  animation: overlay-fade-in 0.6s ease;
}

.offer-celebration-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 48px;
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.94), rgba(236, 248, 255, 0.9));
  box-shadow: 0 25px 50px -12px rgba(59, 130, 246, 0.35), inset 0 0 0 2px rgba(59, 130, 246, 0.12);
  text-align: center;
  transform: scale(0.92);
  animation: card-bounce-in 0.6s ease forwards;
}

.offer-celebration-card--info {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(240, 249, 255, 0.9));
  box-shadow: 0 25px 50px -12px rgba(59, 130, 246, 0.25), inset 0 0 0 2px rgba(59, 130, 246, 0.12);
}

.offer-celebration-title {
  font-size: 28px;
  font-weight: 700;
  color: #1e3a8a;
  letter-spacing: 2px;
}

.offer-celebration-card--info .offer-celebration-title {
  color: #1d4ed8;
}

.offer-celebration-subtitle {
  margin-top: 12px;
  font-size: 18px;
  color: #2563eb;
  letter-spacing: 1px;
}

.offer-celebration-card--info .offer-celebration-subtitle {
  color: #1e40af;
}

.offer-celebration-fade-enter-active,
.offer-celebration-fade-leave-active {
  transition: opacity 0.4s ease;
}

.offer-celebration-fade-enter-from,
.offer-celebration-fade-leave-to {
  opacity: 0;
}

@keyframes card-bounce-in {
  0% {
    transform: scale(0.92);
    opacity: 0;
  }
  60% {
    transform: scale(1.05);
    opacity: 1;
  }
  100% {
    transform: scale(1);
  }
}

@keyframes overlay-fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
