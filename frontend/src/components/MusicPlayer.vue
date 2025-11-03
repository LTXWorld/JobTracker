<template>
  <div class="music-player" :class="{ expanded: isExpanded }">
    <!-- 折叠状态 - 只显示播放按钮 -->
    <div v-if="!isExpanded" class="player-collapsed" @click="toggleExpanded">
      <div class="vinyl-mini" :class="{ spinning: isPlaying }">
        <div class="vinyl-center"></div>
      </div>
    </div>

    <!-- 展开状态 - 完整播放器 -->
    <div v-else class="player-expanded">
      <div class="player-header">
        <span class="player-title">🎵 音乐播放器</span>
        <a-button type="text" size="small" @click="toggleExpanded">
          <CloseOutlined />
        </a-button>
      </div>

      <!-- 黑胶唱片 -->
      <div class="vinyl-container">
        <div class="vinyl-record" :class="{ spinning: isPlaying }">
          <div class="vinyl-surface">
            <div class="vinyl-rings"></div>
            <div class="vinyl-center">
              <div class="center-dot"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- 歌曲信息 -->
      <div class="song-info">
        <div class="song-title">{{ currentSong.title }}</div>
        <div class="song-artist">{{ currentSong.artist }}</div>
        <div v-if="hasError" class="error-message">
          ⚠️ {{ errorMessage }}
        </div>
      </div>

      <!-- 进度条 -->
      <div class="progress-container">
        <span class="time-current">{{ formatTime(currentTime) }}</span>
        <div class="progress-bar" @click="seekTo">
          <div class="progress-bg"></div>
          <div class="progress-fill" :style="{ width: progressPercentage + '%' }"></div>
          <div class="progress-thumb" :style="{ left: progressPercentage + '%' }"></div>
        </div>
        <span class="time-duration">{{ formatTime(duration) }}</span>
      </div>

      <!-- 控制按钮 -->
      <div class="player-controls">
        <a-button type="text" @click="previousSong" class="control-btn">
          <StepBackwardOutlined />
        </a-button>
        <a-button type="primary" @click="togglePlay" class="play-btn" size="large">
          <PauseOutlined v-if="isPlaying" />
          <CaretRightOutlined v-else />
        </a-button>
        <a-button type="text" @click="nextSong" class="control-btn">
          <StepForwardOutlined />
        </a-button>
      </div>

      <!-- 音量控制 -->
      <div class="volume-container">
        <SoundOutlined />
        <div class="volume-slider">
          <a-slider
            v-model:value="volume"
            :min="0"
            :max="100"
            @change="setVolume"
            :tooltip-formatter="(v) => `${v}%`"
          />
        </div>
      </div>
    </div>

    <!-- 音频元素 -->
    <audio
      ref="audioRef"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
      @ended="onSongEnded"
      @error="onAudioError"
      @canplay="onCanPlay"
      preload="metadata"
    ></audio>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import {
  CaretRightOutlined,
  PauseOutlined,
  StepBackwardOutlined,
  StepForwardOutlined,
  SoundOutlined,
  CloseOutlined
} from '@ant-design/icons-vue'
import { OFFER_CELEBRATION_EVENT } from '../utils/offerCelebration'

// 音乐数据结构
interface Song {
  id: number;
  title: string;
  artist: string;
  src: string;
  cover?: string;
}

// 状态管理
const isExpanded = ref(false)
const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const volume = ref(70)
const currentIndex = ref(0)
const audioRef = ref<HTMLAudioElement>()
const isCelebrationMode = ref(false)
const celebrationRestoreState = ref<{ index: number; time: number; wasPlaying: boolean } | null>(null)
const pendingCelebration = ref<{ songId?: number; songTitle?: string } | null>(null)
const pendingRestoreTime = ref<number | null>(null)
const CELEBRATION_SONG_ID = 1
const celebrationExpandState = ref<boolean | null>(null)

// 示例歌曲列表（使用实际的音乐文件）
const playlist = reactive<Song[]>([
  {
    id: 1,
    title: "你离开了南京 从此没有人和我说话",
    artist: "李志",
    src: "/music/05 - 你离开了南京 从此没有人和我说话 (2015 Live).mp3"
  },
  {
    id: 2,
    title: "热河",
    artist: "李志",
    src: "/music/07 - 热河 (2015 Live).mp3"
  },
  {
    id: 3,
    title: "关于郑州的记忆",
    artist: "李志",
    src: "/music/李志 - 关于郑州的记忆 (2016 unplugged).mp3"
  },
  {
    id: 4,
    title: "这个世界会好吗",
    artist: "李志",
    src: "/music/李志 - 这个世界会好吗 (2016 unplugged).mp3"
  },
  {
    id: 5,
    title: "山阴路的夏天",
    artist: "李志,张怡然",
    src: "/music/李志,张怡然 - 山阴路的夏天 (2016 unplugged).mp3"
  }
])

const findSongIndexById = (id: number) => playlist.findIndex(song => song.id === id)

// 错误处理状态
const hasError = ref(false)
const errorMessage = ref('')

const playCelebrationSong = (options?: { songId?: number; songTitle?: string }) => {
  if (celebrationExpandState.value === null) {
    celebrationExpandState.value = isExpanded.value
  }
  isExpanded.value = true

  const audio = audioRef.value
  if (!audio) {
    pendingCelebration.value = options ?? {}
    return
  }

  pendingCelebration.value = null

  const desiredSongId = options?.songId ?? CELEBRATION_SONG_ID
  let celebrationIndex = findSongIndexById(desiredSongId)
  if (celebrationIndex === -1 && options?.songTitle) {
    celebrationIndex = playlist.findIndex(song => song.title === options.songTitle)
  }
  if (celebrationIndex === -1) {
    celebrationIndex = findSongIndexById(CELEBRATION_SONG_ID)
  }
  if (celebrationIndex === -1) return

  if (isCelebrationMode.value) {
    audio.currentTime = 0
    audio
      .play()
      .then(() => {
        isPlaying.value = true
        hasError.value = false
        errorMessage.value = ''
      })
      .catch((error) => {
        console.error('音频播放失败:', error)
        isPlaying.value = false
      })
    return
  }

  if (!celebrationRestoreState.value) {
    celebrationRestoreState.value = {
      index: currentIndex.value,
      time: currentTime.value,
      wasPlaying: isPlaying.value
    }
  }

  pendingRestoreTime.value = null
  isCelebrationMode.value = true

  audio.pause()
  currentIndex.value = celebrationIndex
  currentTime.value = 0
  isPlaying.value = true
  loadCurrentSong()
}

// 计算属性
const currentSong = computed(() => playlist[currentIndex.value] || playlist[0])
const progressPercentage = computed(() => {
  return duration.value > 0 ? (currentTime.value / duration.value) * 100 : 0
})

// 展开/折叠播放器
const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value
}

// 播放/暂停
const togglePlay = async () => {
  if (!audioRef.value) return

  if (isPlaying.value) {
    audioRef.value.pause()
    isPlaying.value = false
    return
  }

  try {
    await audioRef.value.play()
    isPlaying.value = true
    hasError.value = false
    errorMessage.value = ''
  } catch (error) {
    console.error('音频播放失败:', error)
    hasError.value = true
    errorMessage.value = '音频文件加载失败，请检查文件是否存在'
    isPlaying.value = false
  }
}

// 上一首
const previousSong = () => {
  isCelebrationMode.value = false
  celebrationRestoreState.value = null
  pendingRestoreTime.value = null
  celebrationExpandState.value = null
  currentIndex.value = currentIndex.value > 0 ? currentIndex.value - 1 : playlist.length - 1
  loadCurrentSong()
}

// 下一首
const nextSong = () => {
  isCelebrationMode.value = false
  celebrationRestoreState.value = null
  pendingRestoreTime.value = null
  celebrationExpandState.value = null
  currentIndex.value = (currentIndex.value + 1) % playlist.length
  loadCurrentSong()
}

// 加载当前歌曲
const loadCurrentSong = () => {
  if (!audioRef.value) return

  audioRef.value.src = currentSong.value.src
  audioRef.value.load()

  if (isPlaying.value) {
    audioRef.value
      .play()
      .then(() => {
        hasError.value = false
        errorMessage.value = ''
      })
      .catch((error) => {
        console.error('音频播放失败:', error)
        isPlaying.value = false
      })
  }
}

// 设置音量
const setVolume = (val: number) => {
  if (audioRef.value) {
    audioRef.value.volume = val / 100
  }
}

// 跳转到指定时间
const seekTo = (event: MouseEvent) => {
  if (!audioRef.value || duration.value === 0) return

  const progressBar = event.currentTarget as HTMLElement
  const rect = progressBar.getBoundingClientRect()
  const percentage = (event.clientX - rect.left) / rect.width
  const newTime = percentage * duration.value

  audioRef.value.currentTime = newTime
  currentTime.value = newTime
}

// 格式化时间
const formatTime = (time: number): string => {
  const minutes = Math.floor(time / 60)
  const seconds = Math.floor(time % 60)
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
}

// 音频事件处理
const onLoadedMetadata = () => {
  if (audioRef.value) {
    duration.value = audioRef.value.duration
  }
}

const onTimeUpdate = () => {
  if (audioRef.value) {
    currentTime.value = audioRef.value.currentTime
  }
}

const onSongEnded = () => {
  if (isCelebrationMode.value) {
    const restoreState = celebrationRestoreState.value
    isCelebrationMode.value = false
    celebrationRestoreState.value = null
    isPlaying.value = false
    if (audioRef.value) {
      audioRef.value.pause()
    }
    if (restoreState) {
      currentIndex.value = restoreState.index
      pendingRestoreTime.value = restoreState.time
      currentTime.value = restoreState.time
      loadCurrentSong()
    }
    if (celebrationExpandState.value === false) {
      isExpanded.value = false
    }
    celebrationExpandState.value = null
    return
  }
  nextSong()
}

// 新增：音频错误处理
const onAudioError = (event: Event) => {
  console.error('音频加载错误:', event)
  hasError.value = true
  errorMessage.value = '音频文件无法加载，请检查文件路径'
  isPlaying.value = false
}

// 新增：音频可以播放时
const onCanPlay = () => {
  hasError.value = false
  errorMessage.value = ''
  if (!isCelebrationMode.value && pendingRestoreTime.value !== null && audioRef.value) {
    const durationSafe = Number.isFinite(audioRef.value.duration) ? audioRef.value.duration : pendingRestoreTime.value
    const targetTime = Math.min(pendingRestoreTime.value, durationSafe || pendingRestoreTime.value)
    audioRef.value.currentTime = targetTime
    currentTime.value = targetTime
    pendingRestoreTime.value = null
  }
}

// 组件挂载
const handleCelebrationTrigger = (event: Event) => {
  const detail = (event as CustomEvent<{ songId?: number; songTitle?: string }>).detail
  playCelebrationSong(detail)
}

onMounted(() => {
  window.addEventListener(OFFER_CELEBRATION_EVENT, handleCelebrationTrigger)
  if (audioRef.value) {
    audioRef.value.volume = volume.value / 100
    loadCurrentSong()
  }
  if (pendingCelebration.value) {
    playCelebrationSong(pendingCelebration.value)
  }
})

// 组件卸载时清理
onUnmounted(() => {
  window.removeEventListener(OFFER_CELEBRATION_EVENT, handleCelebrationTrigger)
  if (audioRef.value) {
    audioRef.value.pause()
  }
})
</script>

<style scoped>
.music-player {
  position: fixed;
  bottom: 24px;
  left: 24px; /* 改为左下角 */
  z-index: 1500; /* 调整层级，低于银月助手(2000)但高于其他元素 */
  transition: all 0.3s ease;
}

/* 折叠状态样式 */
.player-collapsed {
  cursor: pointer;
  transition: all 0.3s ease;
}

.player-collapsed:hover {
  transform: scale(1.1);
}

.vinyl-mini {
  width: 40px;
  height: 40px;
  background: radial-gradient(circle, #1a1a1a 30%, #333 40%, #1a1a1a 50%);
  border-radius: 50%;
  position: relative;
  border: 2px solid #333;
  transition: all 0.3s ease;
}

.vinyl-mini.spinning {
  animation: spin 3s linear infinite;
}

.vinyl-mini .vinyl-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 8px;
  height: 8px;
  background: #666;
  border-radius: 50%;
}

/* 展开状态样式 */
.player-expanded {
  background: white;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid #f0f0f0;
  padding: 20px;
  width: 280px;
  min-height: 400px;
  /* 展开面板位置调整 */
  position: relative;
}

.player-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.player-title {
  font-weight: 600;
  color: #262626;
}

/* 黑胶唱片样式 */
.vinyl-container {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.vinyl-record {
  width: 120px;
  height: 120px;
  position: relative;
  transition: all 0.3s ease;
}

.vinyl-record.spinning {
  animation: spin 3s linear infinite;
}

.vinyl-surface {
  width: 100%;
  height: 100%;
  background: radial-gradient(
    circle,
    #1a1a1a 25%,
    #333 30%,
    #1a1a1a 35%,
    #333 40%,
    #1a1a1a 45%,
    #333 50%,
    #1a1a1a 55%
  );
  border-radius: 50%;
  position: relative;
  border: 3px solid #333;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3);
}

.vinyl-rings {
  position: absolute;
  top: 10%;
  left: 10%;
  right: 10%;
  bottom: 10%;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}

.vinyl-rings::before,
.vinyl-rings::after {
  content: '';
  position: absolute;
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 50%;
}

.vinyl-rings::before {
  top: 15%;
  left: 15%;
  right: 15%;
  bottom: 15%;
}

.vinyl-rings::after {
  top: 30%;
  left: 30%;
  right: 30%;
  bottom: 30%;
}

.vinyl-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 24px;
  height: 24px;
  background: #666;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.center-dot {
  width: 6px;
  height: 6px;
  background: #999;
  border-radius: 50%;
}

/* 歌曲信息 */
.song-info {
  text-align: center;
  margin-bottom: 20px;
}

.song-title {
  font-weight: 600;
  color: #262626;
  margin-bottom: 4px;
  font-size: 16px;
}

.song-artist {
  color: #8c8c8c;
  font-size: 14px;
}

.error-message {
  color: #ff4d4f;
  font-size: 12px;
  margin-top: 4px;
  text-align: center;
}

/* 进度条 */
.progress-container {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
}

.time-current,
.time-duration {
  font-size: 12px;
  color: #8c8c8c;
  width: 35px;
  text-align: center;
}

.progress-bar {
  flex: 1;
  height: 4px;
  position: relative;
  cursor: pointer;
}

.progress-bg {
  width: 100%;
  height: 100%;
  background: #f0f0f0;
  border-radius: 2px;
}

.progress-fill {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background: #1890ff;
  border-radius: 2px;
  transition: width 0.1s ease;
}

.progress-thumb {
  position: absolute;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 12px;
  height: 12px;
  background: #1890ff;
  border: 2px solid white;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: left 0.1s ease;
}

/* 控制按钮 */
.player-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.control-btn {
  color: #666 !important;
  border: none !important;
  font-size: 18px;
}

.control-btn:hover {
  color: #1890ff !important;
}

.play-btn {
  width: 50px !important;
  height: 50px !important;
  border-radius: 50% !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  font-size: 20px !important;
}

/* 音量控制 */
.volume-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.volume-container .anticon {
  color: #8c8c8c;
}

.volume-slider {
  flex: 1;
}

.volume-slider :deep(.ant-slider) {
  margin: 0;
}

.volume-slider :deep(.ant-slider-track) {
  background: #1890ff;
}

.volume-slider :deep(.ant-slider-handle) {
  border-color: #1890ff;
}

/* 旋转动画 */
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .music-player {
    left: 16px; /* 移动端减少边距 */
    bottom: 16px;
  }

  .player-expanded {
    width: 240px;
    padding: 16px;
  }

  .vinyl-record {
    width: 100px;
    height: 100px;
  }
}
</style>
