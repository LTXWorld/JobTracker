# JobView认证系统测试套件

## 测试概述

**测试范围**: JobView注册登录系统  
**测试类型**: 单元测试、集成测试、E2E测试  
**测试工具**: Jest, Supertest, Playwright  
**覆盖目标**: 核心认证功能和用户体验

## 测试用例设计

### 1. 单元测试用例

#### 1.1 用户名可用性检查

```typescript
// tests/unit/auth/username-availability.test.ts

describe('Username Availability Checker', () => {
  describe('Valid Username Scenarios', () => {
    test('should return available for new username', async () => {
      const username = 'newuser123'
      const result = await AuthService.checkUsernameAvailability(username)
      expect(result.available).toBe(true)
      expect(result.message).toBe('用户名可用')
    })

    test('should handle valid username formats', async () => {
      const validUsernames = [
        'user123', 'User_Name', 'test_user', 'a1b2c3'
      ]
      for (const username of validUsernames) {
        const result = await AuthService.checkUsernameAvailability(username)
        expect(result).toHaveProperty('available')
      }
    })
  })

  describe('Invalid Username Scenarios', () => {
    test('should return unavailable for existing username', async () => {
      const existingUsername = 'testuser'
      const result = await AuthService.checkUsernameAvailability(existingUsername)
      expect(result.available).toBe(false)
      expect(result.message).toBe('用户名已被使用')
    })

    test('should reject invalid username formats', async () => {
      const invalidUsernames = [
        'ab',           // too short
        'a'.repeat(21), // too long
        'user@name',    // invalid characters
        '123-user',     // invalid characters
        'user name'     // spaces not allowed
      ]
      for (const username of invalidUsernames) {
        await expect(
          AuthService.checkUsernameAvailability(username)
        ).rejects.toThrow()
      }
    })
  })

  describe('Edge Cases', () => {
    test('should handle case insensitive check', async () => {
      const result1 = await AuthService.checkUsernameAvailability('TestUser')
      const result2 = await AuthService.checkUsernameAvailability('testuser')
      expect(result1.available).toBe(result2.available)
    })

    test('should handle database errors gracefully', async () => {
      // Mock database error
      jest.spyOn(database, 'query').mockRejectedValueOnce(new Error('DB Error'))
      await expect(
        AuthService.checkUsernameAvailability('testuser')
      ).rejects.toThrow('Database connection failed')
    })
  })
})
```

#### 1.2 邮箱可用性检查

```typescript
// tests/unit/auth/email-availability.test.ts

describe('Email Availability Checker', () => {
  describe('Valid Email Scenarios', () => {
    test('should return available for new email', async () => {
      const email = 'newuser@example.com'
      const result = await AuthService.checkEmailAvailability(email)
      expect(result.available).toBe(true)
    })

    test('should handle various valid email formats', async () => {
      const validEmails = [
        'user@example.com',
        'user.name@example.com',
        'user+tag@example.co.uk',
        'user_name@sub.example.com'
      ]
      for (const email of validEmails) {
        const result = await AuthService.checkEmailAvailability(email)
        expect(result).toHaveProperty('available')
      }
    })
  })

  describe('Invalid Email Scenarios', () => {
    test('should return unavailable for existing email', async () => {
      const existingEmail = 'test@example.com'
      const result = await AuthService.checkEmailAvailability(existingEmail)
      expect(result.available).toBe(false)
      expect(result.message).toBe('邮箱已被注册')
    })

    test('should reject invalid email formats', async () => {
      const invalidEmails = [
        'invalid-email',
        'user@',
        '@example.com',
        'user..name@example.com',
        'user@.example.com'
      ]
      for (const email of invalidEmails) {
        await expect(
          AuthService.checkEmailAvailability(email)
        ).rejects.toThrow('Invalid email format')
      }
    })
  })
})
```

#### 1.3 密码强度计算

```typescript
// tests/unit/auth/password-strength.test.ts

describe('Password Strength Calculator', () => {
  test('should calculate weak password strength', () => {
    const weakPasswords = ['12345678', 'password', 'abcdefgh']
    weakPasswords.forEach(password => {
      const strength = calculatePasswordStrength(password)
      expect(strength.level).toBe(1)
      expect(strength.text).toBe('弱')
    })
  })

  test('should calculate strong password strength', () => {
    const strongPasswords = ['Password123!', 'MyStr0ng@Pass', 'Secure#2024']
    strongPasswords.forEach(password => {
      const strength = calculatePasswordStrength(password)
      expect(strength.level).toBeGreaterThanOrEqual(3)
    })
  })

  test('should require minimum length', () => {
    const shortPassword = 'Pass1!'
    const strength = calculatePasswordStrength(shortPassword)
    expect(strength.level).toBeLessThan(3)
  })
})
```

### 2. 集成测试用例

#### 2.1 API端点测试

```typescript
// tests/integration/auth-api.test.ts

describe('Authentication API Integration', () => {
  let server: Server
  let request: supertest.SuperTest<supertest.Test>

  beforeAll(async () => {
    server = await createTestServer()
    request = supertest(server)
  })

  afterAll(async () => {
    await server.close()
  })

  describe('GET /api/auth/check-username', () => {
    test('should return 200 for available username', async () => {
      const response = await request
        .get('/api/auth/check-username?username=newuser123')
        .expect(200)

      expect(response.body).toEqual({
        code: 200,
        message: '检查完成',
        data: {
          available: true,
          message: '用户名可用'
        }
      })
    })

    test('should return 200 for unavailable username', async () => {
      await createTestUser({ username: 'existinguser' })
      
      const response = await request
        .get('/api/auth/check-username?username=existinguser')
        .expect(200)

      expect(response.body.data.available).toBe(false)
    })

    test('should return 400 for invalid username', async () => {
      const response = await request
        .get('/api/auth/check-username?username=ab')
        .expect(400)

      expect(response.body.message).toContain('用户名格式不正确')
    })

    test('should handle missing username parameter', async () => {
      await request
        .get('/api/auth/check-username')
        .expect(400)
    })
  })

  describe('GET /api/auth/check-email', () => {
    test('should return 200 for available email', async () => {
      const response = await request
        .get('/api/auth/check-email?email=new@example.com')
        .expect(200)

      expect(response.body.data.available).toBe(true)
    })

    test('should return 200 for unavailable email', async () => {
      await createTestUser({ email: 'existing@example.com' })
      
      const response = await request
        .get('/api/auth/check-email?email=existing@example.com')
        .expect(200)

      expect(response.body.data.available).toBe(false)
    })
  })

  describe('CORS Configuration', () => {
    test('should allow requests from localhost:3000', async () => {
      const response = await request
        .options('/api/auth/check-username')
        .set('Origin', 'http://localhost:3000')
        .set('Access-Control-Request-Method', 'GET')
        .expect(200)

      expect(response.headers['access-control-allow-origin']).toBe('http://localhost:3000')
    })

    test('should reject requests from unknown origins', async () => {
      await request
        .options('/api/auth/check-username')
        .set('Origin', 'http://malicious-site.com')
        .expect(403)
    })
  })
})
```

#### 2.2 数据库集成测试

```typescript
// tests/integration/auth-database.test.ts

describe('Authentication Database Integration', () => {
  beforeEach(async () => {
    await cleanDatabase()
    await runMigrations()
  })

  describe('User Registration Flow', () => {
    test('should create user with unique constraints', async () => {
      const userData = {
        username: 'testuser',
        email: 'test@example.com',
        password: 'hashedPassword123'
      }

      const userId = await AuthService.createUser(userData)
      expect(userId).toBeDefined()

      // Verify user was created
      const user = await AuthService.getUserById(userId)
      expect(user.username).toBe(userData.username)
      expect(user.email).toBe(userData.email)
    })

    test('should prevent duplicate usernames', async () => {
      const userData = {
        username: 'duplicateuser',
        email: 'user1@example.com',
        password: 'password123'
      }

      await AuthService.createUser(userData)

      // Attempt to create user with same username
      await expect(
        AuthService.createUser({
          ...userData,
          email: 'user2@example.com'
        })
      ).rejects.toThrow('Username already exists')
    })

    test('should prevent duplicate emails', async () => {
      const userData = {
        username: 'user1',
        email: 'duplicate@example.com',
        password: 'password123'
      }

      await AuthService.createUser(userData)

      // Attempt to create user with same email
      await expect(
        AuthService.createUser({
          ...userData,
          username: 'user2'
        })
      ).rejects.toThrow('Email already registered')
    })
  })
})
```

### 3. E2E测试用例

#### 3.1 完整注册流程

```typescript
// tests/e2e/registration.test.ts

describe('User Registration E2E', () => {
  test('complete registration workflow', async () => {
    const { page } = await createBrowser()

    // Navigate to registration page
    await page.goto('http://localhost:3000/register')

    // Fill username field
    await page.fill('[data-testid="username-input"]', 'newe2euser')
    await page.blur('[data-testid="username-input"]')

    // Wait for username availability check
    await page.waitForSelector('[data-testid="username-success-icon"]')

    // Fill email field
    await page.fill('[data-testid="email-input"]', 'newe2e@example.com')
    await page.blur('[data-testid="email-input"]')

    // Wait for email availability check
    await page.waitForSelector('[data-testid="email-success-icon"]')

    // Fill password fields
    await page.fill('[data-testid="password-input"]', 'SecurePass123!')
    await page.fill('[data-testid="confirm-password-input"]', 'SecurePass123!')

    // Check agreement
    await page.check('[data-testid="agreement-checkbox"]')

    // Submit form
    await page.click('[data-testid="register-button"]')

    // Verify success
    await page.waitForSelector('[data-testid="success-message"]')
    await expect(page.locator('[data-testid="success-message"]')).toContainText('注册成功')

    // Verify redirect to dashboard
    await page.waitForURL('http://localhost:3000/')
  })

  test('should show real-time validation errors', async () => {
    const { page } = await createBrowser()
    await page.goto('http://localhost:3000/register')

    // Test username conflict
    await page.fill('[data-testid="username-input"]', 'testuser') // existing user
    await page.blur('[data-testid="username-input"]')
    
    await page.waitForSelector('[data-testid="username-error-icon"]')
    await expect(page.locator('[data-testid="username-error"]')).toContainText('用户名已被使用')

    // Test email conflict
    await page.fill('[data-testid="email-input"]', 'test@example.com') // existing email
    await page.blur('[data-testid="email-input"]')
    
    await page.waitForSelector('[data-testid="email-error-icon"]')
  })

  test('should handle network errors gracefully', async () => {
    const { page } = await createBrowser()
    
    // Block API requests
    await page.route('/api/auth/check-username*', route => route.abort())
    
    await page.goto('http://localhost:3000/register')
    await page.fill('[data-testid="username-input"]', 'testuser')
    await page.blur('[data-testid="username-input"]')
    
    // Should show error state
    await page.waitForSelector('[data-testid="username-error-icon"]')
  })
})
```

### 4. 性能测试用例

```typescript
// tests/performance/auth-performance.test.ts

describe('Authentication Performance Tests', () => {
  test('username availability check should respond within 200ms', async () => {
    const startTime = Date.now()
    
    await request
      .get('/api/auth/check-username?username=perftest')
      .expect(200)
    
    const responseTime = Date.now() - startTime
    expect(responseTime).toBeLessThan(200)
  })

  test('should handle concurrent availability checks', async () => {
    const concurrentRequests = 50
    const promises = Array.from({ length: concurrentRequests }, (_, i) =>
      request.get(`/api/auth/check-username?username=user${i}`)
    )

    const startTime = Date.now()
    const responses = await Promise.all(promises)
    const totalTime = Date.now() - startTime

    expect(responses.every(r => r.status === 200)).toBe(true)
    expect(totalTime / concurrentRequests).toBeLessThan(100) // Average < 100ms
  })
})
```

### 5. 安全测试用例

```typescript
// tests/security/auth-security.test.ts

describe('Authentication Security Tests', () => {
  test('should prevent SQL injection in username check', async () => {
    const maliciousUsernames = [
      "admin'; DROP TABLE users; --",
      "admin' OR '1'='1",
      "admin' UNION SELECT * FROM users --"
    ]

    for (const username of maliciousUsernames) {
      const response = await request
        .get(`/api/auth/check-username?username=${encodeURIComponent(username)}`)
        
      expect(response.status).not.toBe(500)
      expect(response.body).toHaveProperty('data')
    }
  })

  test('should rate limit availability checks', async () => {
    // Make multiple requests rapidly
    const requests = Array.from({ length: 15 }, () =>
      request.get('/api/auth/check-username?username=ratetest')
    )

    const responses = await Promise.all(requests.map(r => r.catch(e => e.response)))
    const rateLimitedResponses = responses.filter(r => r.status === 429)
    
    expect(rateLimitedResponses.length).toBeGreaterThan(0)
  })

  test('should not leak sensitive information in errors', async () => {
    // Simulate database error
    jest.spyOn(database, 'query').mockRejectedValueOnce(new Error('Connection failed'))

    const response = await request
      .get('/api/auth/check-username?username=testuser')
      .expect(500)

    expect(response.body.message).not.toContain('Connection failed')
    expect(response.body.message).toBe('Internal server error')
  })
})
```

## 测试配置和工具

### Jest配置 (jest.config.js)
```javascript
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/tests'],
  testMatch: ['**/__tests__/**/*.ts', '**/?(*.)+(spec|test).ts'],
  transform: {
    '^.+\\.ts$': 'ts-jest',
  },
  collectCoverageFrom: [
    'src/**/*.{ts,js}',
    '!src/**/*.d.ts',
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
}
```

### Playwright配置 (playwright.config.ts)
```typescript
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    port: 3000,
  },
})
```

## 测试执行策略

### 本地开发环境
```bash
# 运行所有测试
npm test

# 运行特定测试套件
npm run test:unit
npm run test:integration
npm run test:e2e

# 生成覆盖率报告
npm run test:coverage

# 监听模式
npm run test:watch
```

### CI/CD流水线
```bash
# 预测试环境检查
npm run lint
npm run type-check

# 单元和集成测试
npm run test:unit
npm run test:integration

# E2E测试
npm run build
npm run test:e2e

# 性能测试
npm run test:performance

# 安全测试
npm run test:security
```

## 验收标准

### 功能验收
- [ ] 所有API端点返回正确响应
- [ ] 前端实时验证正常工作
- [ ] 错误处理和用户反馈完善
- [ ] CORS配置正确

### 性能验收
- [ ] API响应时间 < 200ms (P95)
- [ ] 前端渲染时间 < 100ms
- [ ] 并发处理能力 > 50 RPS
- [ ] 内存使用稳定

### 安全验收
- [ ] 输入验证完整
- [ ] SQL注入防护有效
- [ ] 速率限制正常工作
- [ ] 错误信息不泄露敏感数据

### 用户体验验收
- [ ] 实时反馈响应及时
- [ ] 错误提示清晰友好
- [ ] Loading状态明确
- [ ] 表单验证逻辑合理