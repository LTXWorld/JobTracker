import { createRouter, createWebHistory } from 'vue-router'
import { message } from 'ant-design-vue'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // 重定向规则
    {
      path: '/',
      name: 'home',
      redirect: '/kanban'
    },
    
    // 认证相关路由
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/auth/Login.vue'),
      meta: {
        title: '用户登录',
        requiresGuest: true // 只有未登录用户可以访问
      }
    },
    {
      path: '/register', 
      name: 'register',
      component: () => import('../views/auth/Register.vue'),
      meta: {
        title: '用户注册',
        requiresGuest: true
      }
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('../views/auth/Profile.vue'),
      meta: {
        title: '个人资料',
        requiresAuth: true
      }
    },
    
    // 业务功能路由 - 全部需要认证
    {
      path: '/kanban',
      name: 'kanban',
      component: () => import('../views/KanbanBoard.vue'),
      meta: {
        title: '看板视图',
        requiresAuth: true
      }
    },
    {
      path: '/timeline',
      name: 'timeline',
      component: () => import('../views/Timeline.vue'),
      meta: {
        title: '投递记录',
        requiresAuth: true
      }
    },
    {
      path: '/reminders',
      name: 'reminders',
      component: () => import('../views/Reminders.vue'),
      meta: {
        title: '提醒中心',
        requiresAuth: true
      }
    },
    {
      path: '/statistics',
      name: 'statistics', 
      component: () => import('../views/Statistics.vue'),
      meta: {
        title: '数据统计',
        requiresAuth: true
      }
    },
    {
      path: '/application/:id',
      name: 'application-detail',
      component: () => import('../views/ApplicationDetail.vue'),
      meta: {
        title: '投递详情',
        requiresAuth: true
      }
    },
    
    // 404 页面
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../views/NotFound.vue'),
      meta: {
        title: '页面不存在'
      }
    }
  ]
})

// 全局路由守卫
router.beforeEach(async (to, from, next) => {
  // 获取认证状态
  const authStore = useAuthStore()
  
  // 初始化认证状态（如果尚未初始化）
  if (!authStore.isAuthenticated && sessionStorage.getItem('access_token')) {
    authStore.initAuth()
  }

  // 设置页面标题
  if (to.meta?.title) {
    document.title = `${to.meta.title} - JobView`
  } else {
    document.title = 'JobView - 求职记录管理'
  }

  // 检查路由权限
  const requiresAuth = to.meta.requiresAuth
  const requiresGuest = to.meta.requiresGuest
  const isAuthenticated = authStore.isAuthenticated

  if (requiresAuth && !isAuthenticated) {
    // 需要认证但未登录，跳转到登录页
    message.warning('请先登录后访问')
    next({
      name: 'login',
      query: { 
        redirect: to.fullPath,
        error: 'unauthorized' 
      }
    })
  } else if (requiresGuest && isAuthenticated) {
    // 只允许游客访问但已登录，跳转到首页
    message.info('您已登录，自动跳转到首页')
    next({ name: 'kanban' })
  } else {
    // 对于需要认证的路由，使用智能验证策略
    if (requiresAuth && isAuthenticated) {
      // 如果是刚登录后的重定向，跳过token验证避免时序问题
      if (from.name === 'login' || from.name === 'register') {
        next()
        return
      }
      
      // 检查token是否需要验证（避免过度验证）
      const needsValidation = authStore.shouldValidateToken()
      if (!needsValidation) {
        // token仍在有效期内，直接放行
        next()
        return
      }
      
      // 需要验证token时才进行服务端验证
      try {
        const isValid = await authStore.validateToken()
        if (!isValid) {
          // Token无效，清除认证信息并跳转登录
          authStore.clearAuth()
          message.error('登录已过期，请重新登录')
          next({
            name: 'login',
            query: { 
              redirect: to.fullPath,
              error: 'token_expired'
            }
          })
          return
        }
      } catch (error) {
        // Token验证失败，但要区分网络错误和真实的认证错误
        console.error('Token validation failed:', error)
        
        // 如果是网络错误，给予一次宽限，避免因网络波动导致误判
        const isNetworkError = (error as Error).message?.includes('网络') || 
                              (error as Error).message?.includes('Network') ||
                              (error as Error).message?.includes('timeout')
        
        if (isNetworkError && authStore.isTokenRecentlyValid()) {
          console.warn('网络错误导致token验证失败，允许本次访问')
          next()
          return
        }
        
        // 真实的认证错误或token确实过期
        authStore.clearAuth()
        message.error('登录验证失败，请重新登录')
        next({
          name: 'login',
          query: { 
            redirect: to.fullPath,
            error: 'validation_failed'
          }
        })
        return
      }
    }
    
    next()
  }
})

// 路由跳转完成后的处理
router.afterEach((to) => {
  // 滚动到页面顶部
  window.scrollTo(0, 0)
  
  // 这里可以添加页面访问统计等逻辑
  console.log(`Route changed to: ${String(to.name)}`)
})

export default router