// 环境配置管理
// 注意：后端 auth 路由为 /api/auth，业务 v1 路由为 /api/v1。
// 为避免路径拼接错误，分别暴露 AUTH_BASE_URL 与 API_V1_BASE_URL。
const ENV_CONFIG = {
  // 本地开发环境
  local: {
    AUTH_BASE_URL: 'http://localhost:8010/api/auth',
    API_V1_BASE_URL: 'http://localhost:8010/api/v1',
    FRONTEND_URL: 'http://localhost:3000'
  },
  // 生产环境（云服务器）
  production: {
    AUTH_BASE_URL: 'https://jobview.bfsmlt.top/api/auth',
    API_V1_BASE_URL: 'https://jobview.bfsmlt.top/api/v1',
    FRONTEND_URL: 'https://jobview.bfsmlt.top'
  }
};

// 自动检测环境（用于初始默认值）
function detectEnvironment() {
  if (location.hostname === 'localhost' || location.hostname === '127.0.0.1') {
    return 'local';
  }
  return 'production';
}

// 根据环境名获取配置
function getConfig(env) {
  return ENV_CONFIG[env] || ENV_CONFIG.production;
}

if (typeof module !== 'undefined' && module.exports) {
  module.exports = { ENV_CONFIG, detectEnvironment, getConfig };
}
