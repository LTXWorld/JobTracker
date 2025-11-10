import { pinyin } from 'pinyin-pro'

/**
 * 针对中文内容构建可用于搜索的关键字串，包含原文及拼音形式
 */
const CHINESE_REGEX = /[\u4e00-\u9fff]/
const cache = new Map<string, string>()

export const createSearchTokens = (text?: string | null): string => {
  const raw = (text ?? '').trim()
  if (!raw) return ''

  const cached = cache.get(raw)
  if (cached) return cached

  const lower = raw.toLowerCase()
  if (!CHINESE_REGEX.test(raw)) {
    cache.set(raw, lower)
    return lower
  }

  let tokens = lower
  try {
    const syllables = pinyin(raw, { toneType: 'none', type: 'array' }) as string[]
    if (Array.isArray(syllables) && syllables.length > 0) {
      const continuous = syllables.join('')
      const spaced = syllables.join(' ')
      const initials = syllables.map(s => (s?.[0] ?? '')).join('')
      tokens = [lower, continuous, spaced, initials]
        .filter(Boolean)
        .map(str => str.toLowerCase())
        .join(' ')
    }
  } catch {
    // 忽略拼音库异常，直接退回原始小写内容
  }

  cache.set(raw, tokens)
  return tokens
}
