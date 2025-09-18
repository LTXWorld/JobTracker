// Mock API service for development
interface MockUser {
  id: string;
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
}

interface MockLoginResponse {
  success: boolean;
  data?: {
    user: MockUser;
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  };
  message?: string;
}

// Mock users for testing
const mockUsers = [
  {
    id: '1',
    username: 'demo@jobview.com',
    email: 'demo@jobview.com',
    firstName: '演示',
    lastName: '用户',
    password: 'demo123',
  },
  {
    id: '2',
    username: 'testuser',
    email: 'test@jobview.com',
    firstName: '测试',
    lastName: '用户',
    password: 'test123',
  },
];

// Mock delay to simulate network request
const delay = (ms: number) => new Promise<void>(resolve => setTimeout(() => resolve(), ms));

export const mockAuthAPI = {
  login: async (credentials: { username: string; password: string }): Promise<MockLoginResponse> => {
    await delay(1000); // Simulate network delay

    const user = mockUsers.find(
      u => (u.username === credentials.username || u.email === credentials.username)
           && u.password === credentials.password
    );

    if (user) {
      const { password, ...userWithoutPassword } = user;
      return {
        success: true,
        data: {
          user: userWithoutPassword,
          accessToken: 'mock_access_token_' + Date.now(),
          refreshToken: 'mock_refresh_token_' + Date.now(),
          expiresIn: 3600, // 1 hour
        },
      };
    } else {
      return {
        success: false,
        message: '用户名或密码错误',
      };
    }
  },

  register: async (userData: {
    username: string;
    email: string;
    password: string;
    firstName?: string;
    lastName?: string;
  }) => {
    await delay(1000);

    // Check if user already exists
    const existingUser = mockUsers.find(
      u => u.username === userData.username || u.email === userData.email
    );

    if (existingUser) {
      return {
        success: false,
        message: '用户名或邮箱已存在',
      };
    }

    // In real app, would save to database
    return {
      success: true,
      message: '注册成功，请登录',
    };
  },

  forgotPassword: async (email: string) => {
    await delay(1000);

    const user = mockUsers.find(u => u.email === email);

    if (user) {
      return {
        success: true,
        message: '重置密码邮件已发送',
      };
    } else {
      return {
        success: false,
        message: '邮箱不存在',
      };
    }
  },
};

export default mockAuthAPI;