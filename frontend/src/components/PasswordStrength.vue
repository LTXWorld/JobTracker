<template>
  <div class="password-strength" v-if="password">
    <div class="strength-bar">
      <div 
        class="strength-fill"
        :class="strengthClass"
        :style="{ width: `${(strength.score + 1) * 20}%` }"
      ></div>
    </div>
    <div class="strength-info">
      <span class="strength-text" :class="strengthClass">
        {{ strengthText }}
      </span>
      <div class="strength-requirements" v-if="showRequirements">
        <div 
          v-for="(req, index) in requirements" 
          :key="index"
          class="requirement-item"
          :class="{ 'met': req.met }"
        >
          <check-outlined v-if="req.met" />
          <close-outlined v-else />
          <span>{{ req.text }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CheckOutlined, CloseOutlined } from '@ant-design/icons-vue'
import type { PasswordStrength } from '../types/auth'

interface Props {
  password: string
  showRequirements?: boolean
}

interface Requirement {
  text: string
  met: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showRequirements: true
})

// 密码强度检查
const checkPasswordStrength = (password: string): PasswordStrength => {
  let score = 0
  const feedback: string[] = []
  const suggestions: string[] = []

  if (!password) {
    return { score: 0, feedback: ['密码不能为空'], suggestions: ['请输入密码'] }
  }

  // 长度检查
  if (password.length >= 8) {
    score++
  } else {
    feedback.push('密码长度不足')
    suggestions.push('密码至少需要8个字符')
  }

  // 包含小写字母
  if (/[a-z]/.test(password)) {
    score++
  } else {
    feedback.push('缺少小写字母')
    suggestions.push('请添加小写字母')
  }

  // 包含大写字母
  if (/[A-Z]/.test(password)) {
    score++
  } else {
    feedback.push('缺少大写字母')
    suggestions.push('请添加大写字母')
  }

  // 包含数字
  if (/\d/.test(password)) {
    score++
  } else {
    feedback.push('缺少数字')
    suggestions.push('请添加数字')
  }

  // 包含特殊字符
  if (/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
    score++
  } else {
    feedback.push('缺少特殊字符')
    suggestions.push('请添加特殊字符 (!@#$%^&* 等)')
  }

  return { score: Math.min(score, 4), feedback, suggestions }
}

// 计算密码强度
const strength = computed(() => checkPasswordStrength(props.password))

// 强度等级文本
const strengthText = computed(() => {
  const texts = ['很弱', '较弱', '一般', '较强', '很强']
  return texts[strength.value.score] || '很弱'
})

// 强度等级样式类
const strengthClass = computed(() => {
  const classes = ['very-weak', 'weak', 'fair', 'good', 'strong']
  return classes[strength.value.score] || 'very-weak'
})

// 密码要求检查
const requirements = computed((): Requirement[] => [
  {
    text: '至少8个字符',
    met: props.password.length >= 8
  },
  {
    text: '包含小写字母',
    met: /[a-z]/.test(props.password)
  },
  {
    text: '包含大写字母',
    met: /[A-Z]/.test(props.password)
  },
  {
    text: '包含数字',
    met: /\d/.test(props.password)
  },
  {
    text: '包含特殊字符',
    met: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(props.password)
  }
])
</script>

<style scoped>
.password-strength {
  margin-top: 8px;
}

.strength-bar {
  width: 100%;
  height: 4px;
  background-color: #f0f0f0;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 8px;
}

.strength-fill {
  height: 100%;
  transition: all 0.3s ease;
  border-radius: 2px;
}

.strength-fill.very-weak {
  background-color: #ff4d4f;
}

.strength-fill.weak {
  background-color: #ff7a45;
}

.strength-fill.fair {
  background-color: #ffa940;
}

.strength-fill.good {
  background-color: #73d13d;
}

.strength-fill.strong {
  background-color: #52c41a;
}

.strength-info {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: 12px;
}

.strength-text {
  font-weight: 500;
}

.strength-text.very-weak {
  color: #ff4d4f;
}

.strength-text.weak {
  color: #ff7a45;
}

.strength-text.fair {
  color: #ffa940;
}

.strength-text.good {
  color: #73d13d;
}

.strength-text.strong {
  color: #52c41a;
}

.strength-requirements {
  flex: 1;
  max-width: 200px;
  margin-left: 16px;
}

.requirement-item {
  display: flex;
  align-items: center;
  margin: 2px 0;
  font-size: 11px;
  color: #8c8c8c;
  transition: color 0.2s ease;
}

.requirement-item.met {
  color: #52c41a;
}

.requirement-item.met .anticon {
  color: #52c41a;
}

.requirement-item:not(.met) .anticon {
  color: #ff4d4f;
}

.requirement-item .anticon {
  margin-right: 4px;
  font-size: 10px;
}

.requirement-item span {
  flex: 1;
}
</style>