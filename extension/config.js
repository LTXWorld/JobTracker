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

// 自动检测环境
function detectEnvironment() {
  // 如果是本地开发环境
  if (location.hostname === 'localhost' || location.hostname === '127.0.0.1') {
    return 'local';
  }
  // 生产环境
  return 'production';
}

// 获取当前环境配置
function getConfig() {
  const env = detectEnvironment();
  return ENV_CONFIG[env];
}

// 导出配置
const CONFIG = getConfig();

// 为了方便调试，在控制台输出当前配置
console.log('[JobView Extension] Current environment:', detectEnvironment());
console.log('[JobView Extension] AUTH URL:', CONFIG.AUTH_BASE_URL);
console.log('[JobView Extension] API v1 URL:', CONFIG.API_V1_BASE_URL);
console.log('[JobView Extension] Frontend URL:', CONFIG.FRONTEND_URL);
