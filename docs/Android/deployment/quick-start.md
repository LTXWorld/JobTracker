# JobView Mobile 快速部署指南

## 🚀 一键部署到您的 Android 手机

本指南将帮助您快速将 JobView Mobile 应用部署到您的 Android 设备上。

## 📱 准备工作

### 1. 手机设置
1. **开启开发者选项**
   - 进入 **设置** > **关于手机**
   - 连续点击 **版本号** 7 次
   - 返回设置，找到 **开发者选项**

2. **启用 USB 调试**
   - 进入 **开发者选项**
   - 开启 **USB 调试**
   - 开启 **允许 USB 安装** (如果有此选项)

3. **连接手机**
   - 使用 USB 数据线连接手机和电脑
   - 手机上会弹出授权提示，选择 **允许**

### 2. 环境检查
运行环境检查脚本确保开发环境正确配置：

```bash
./scripts/check-environment.sh
```

如果检查发现问题，请参考详细部署文档进行修复：[deployment-guide.md](deployment-guide.md)

## 🎯 快速部署

### 方式一：开发版本 (推荐新手)
开发版本支持热重载，方便调试：

```bash
./scripts/deploy-debug.sh
```

**特点：**
- ✅ 支持热重载和调试
- ✅ 可以实时查看代码更改
- ⚠️ 需要保持电脑连接
- ⚠️ APK 体积较大

### 方式二：生产版本 (推荐日常使用)
生产版本经过优化，性能更好：

```bash
./scripts/deploy-release.sh
```

**特点：**
- ✅ 性能优化，运行流畅
- ✅ APK 体积小
- ✅ 可以断开电脑连接
- ❌ 不支持热重载调试

## 📋 部署步骤详解

### Step 1: 检查连接
```bash
# 检查设备连接
adb devices

# 应该显示你的设备，例如：
# 12345678    device
```

### Step 2: 选择部署方式
```bash
# 开发版本 (第一次推荐)
./scripts/deploy-debug.sh

# 或者生产版本
./scripts/deploy-release.sh
```

### Step 3: 等待完成
脚本会自动：
- 检查环境和设备连接
- 安装依赖包
- 构建应用
- 安装到手机
- 启动应用

## 🔧 手动命令

如果您更喜欢手动操作：

### 开发版本
```bash
# 1. 安装依赖
npm install

# 2. 启动 Metro bundler (保持运行)
npm start

# 3. 构建并安装 (新终端窗口)
npm run android
```

### 生产版本
```bash
# 1. 安装依赖
npm install

# 2. 清理缓存
npm run clean

# 3. 构建 Release APK
cd android
./gradlew assembleRelease
cd ..

# 4. 安装到设备
adb install android/app/build/outputs/apk/release/app-release.apk
```

## 🐛 常见问题解决

### 问题 1: 设备未授权
```
List of devices attached
12345678    unauthorized
```

**解决方案：**
- 检查手机屏幕是否有授权弹窗
- 重新连接 USB 线
- 重启 adb：`adb kill-server && adb start-server`

### 问题 2: 找不到设备
```
List of devices attached
```

**解决方案：**
- 确认 USB 调试已开启
- 尝试不同的 USB 端口
- 检查数据线是否支持数据传输
- 在开发者选项中撤销 USB 调试授权，重新连接

### 问题 3: 构建失败
```
BUILD FAILED
```

**解决方案：**
```bash
# 清理并重新构建
npm run clean
./scripts/deploy-debug.sh
```

### 问题 4: Metro bundler 连接失败
**解决方案：**
```bash
# 设置端口转发
adb reverse tcp:8081 tcp:8081

# 重启 Metro
npm start -- --reset-cache
```

## 📱 应用使用说明

### 首次启动
1. 应用安装完成后会自动启动
2. 首次使用需要创建账户或使用测试账户
3. 设置通知权限（推荐开启）

### 测试账户
- 用户名：`testuser`
- 密码：`TestPass123!`

### 主要功能
- 📝 求职申请记录管理
- 📊 数据统计和可视化
- 🔔 智能提醒系统
- 📤 数据导入导出
- 🎨 美观的界面设计

## 🔄 更新应用

### 开发版本更新
如果使用开发版本，直接保存代码文件即可热重载更新。

### 生产版本更新
```bash
# 重新构建并安装
./scripts/deploy-release.sh
```

## 📚 更多资源

- **详细部署文档**: [deployment-guide.md](deployment-guide.md)
- **React Native 官方文档**: https://reactnative.dev
- **Android 开发者指南**: https://developer.android.com

## 🆘 需要帮助？

如果遇到问题：

1. **查看日志**：
   ```bash
   # Android 系统日志
   adb logcat | grep ReactNativeJS

   # Metro bundler 日志
   npm start
   ```

2. **环境检查**：
   ```bash
   ./scripts/check-environment.sh
   npx react-native doctor
   ```

3. **重置环境**：
   ```bash
   npm run clean
   rm -rf node_modules
   npm install
   ```

---

**祝您部署成功，享受使用 JobView Mobile！** 🎉