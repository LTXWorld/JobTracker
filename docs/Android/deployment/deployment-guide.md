# JobView Mobile Android 部署指南

## 📱 概述

本文档将指导您将 JobView Mobile 应用程序部署到 Android 设备上。包括开发环境设置、构建配置、签名设置和安装流程。

## 🛠️ 环境要求

### 系统要求
- **操作系统**: Windows 10+, macOS 10.15+, 或 Ubuntu 18.04+
- **内存**: 至少 8GB RAM (推荐 16GB)
- **存储**: 至少 30GB 可用空间
- **网络**: 稳定的互联网连接

### 必需软件

#### 1. Node.js 和 npm
```bash
# 检查版本
node --version  # 应该 >= 18.0.0
npm --version   # 应该 >= 8.0.0

# 如果没有安装，从 https://nodejs.org 下载安装
```

#### 2. Java Development Kit (JDK)
```bash
# 检查版本
java -version   # 应该是 JDK 11 或 17

# macOS 安装
brew install openjdk@11

# Windows 安装
# 从 Oracle 或 OpenJDK 官网下载安装
```

#### 3. Android Studio 和 Android SDK
1. 下载并安装 [Android Studio](https://developer.android.com/studio)
2. 打开 Android Studio，进入 SDK Manager
3. 安装以下组件：
   - Android SDK Platform 33 (或更高)
   - Android SDK Build-Tools 33.0.0 (或更高)
   - Android SDK Platform-Tools
   - Android Emulator (可选，用于测试)

#### 4. 环境变量配置

**macOS/Linux (`~/.bashrc` 或 `~/.zshrc`):**
```bash
export ANDROID_HOME=$HOME/Library/Android/sdk
export PATH=$PATH:$ANDROID_HOME/emulator
export PATH=$PATH:$ANDROID_HOME/platform-tools
export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin
```

**Windows (环境变量):**
```
ANDROID_HOME = C:\Users\YourUsername\AppData\Local\Android\Sdk
PATH += %ANDROID_HOME%\platform-tools
PATH += %ANDROID_HOME%\emulator
PATH += %ANDROID_HOME%\cmdline-tools\latest\bin
```

## 📦 项目准备

### 1. 克隆和安装项目
```bash
# 进入项目目录
cd /path/to/JobViewMobile

# 安装依赖
npm install

# 清理缓存 (如果之前构建过)
npm run clean
npx react-native start --reset-cache
```

### 2. 验证环境
```bash
# 检查 React Native 环境
npx react-native doctor

# 检查 Android 设备/模拟器
adb devices
```

## 📱 设备准备

### 方法一：使用真实 Android 设备 (推荐)

#### 1. 启用开发者选项
1. 打开 **设置** > **关于手机**
2. 连续点击 **版本号** 7 次
3. 返回设置，找到 **开发者选项**

#### 2. 启用 USB 调试
1. 进入 **开发者选项**
2. 开启 **USB 调试**
3. 开启 **允许 USB 安装** (如果有)
4. 开启 **USB 调试 (安全设置)** (如果有)

#### 3. 连接设备
1. 使用 USB 数据线连接手机和电脑
2. 手机上会弹出授权提示，选择 **允许**
3. 运行以下命令验证连接:
```bash
adb devices
# 应该显示你的设备 ID
```

### 方法二：使用 Android 模拟器

#### 1. 创建 AVD (Android Virtual Device)
```bash
# 列出可用的系统镜像
avdmanager list

# 创建 AVD (示例)
avdmanager create avd -n JobViewTest -k "system-images;android-33;google_apis;x86_64"

# 启动模拟器
emulator -avd JobViewTest
```

#### 2. 或者通过 Android Studio
1. 打开 Android Studio
2. 选择 **Tools** > **AVD Manager**
3. 点击 **Create Virtual Device**
4. 选择设备型号和系统镜像
5. 完成创建并启动

## 🔨 构建和部署

### 1. 开发版本部署 (Debug Build)

这是最简单的部署方式，适合开发和测试：

```bash
# 确保设备已连接
adb devices

# 启动 Metro bundler (新终端窗口)
npm start

# 构建并安装到设备 (另一个终端窗口)
npm run android

# 或者手动运行
npx react-native run-android
```

**注意事项：**
- Debug 版本会连接到开发服务器
- 需要保持 Metro bundler 运行
- 支持热重载和调试功能
- APK 体积较大，性能较慢

### 2. 生产版本部署 (Release Build)

适合正式使用的优化版本：

#### Step 1: 构建 Release APK
```bash
# 清理之前的构建
npm run clean

# 构建 Release APK
cd android
./gradlew assembleRelease
cd ..

# APK 位置
# android/app/build/outputs/apk/release/app-release.apk
```

#### Step 2: 安装到设备
```bash
# 直接安装
adb install android/app/build/outputs/apk/release/app-release.apk

# 如果已安装，需要卸载后重装
adb uninstall com.jobviewmobile
adb install android/app/build/outputs/apk/release/app-release.apk
```

### 3. 签名 APK (可选，用于发布)

#### Step 1: 生成签名密钥
```bash
# 进入 android/app 目录
cd android/app

# 生成密钥库文件
keytool -genkeypair -v -keystore jobview-release.keystore -alias jobview-key-alias -keyalg RSA -keysize 2048 -validity 10000

# 按提示输入信息，记住密码和别名
```

#### Step 2: 配置签名
在 `android/app/build.gradle` 中添加：

```gradle
android {
    ...
    signingConfigs {
        release {
            if (project.hasProperty('JOBVIEW_RELEASE_STORE_FILE')) {
                storeFile file(JOBVIEW_RELEASE_STORE_FILE)
                storePassword JOBVIEW_RELEASE_STORE_PASSWORD
                keyAlias JOBVIEW_RELEASE_KEY_ALIAS
                keyPassword JOBVIEW_RELEASE_KEY_PASSWORD
            }
        }
    }
    buildTypes {
        release {
            ...
            signingConfig signingConfigs.release
        }
    }
}
```

#### Step 3: 设置签名属性
在 `android/gradle.properties` 中添加：

```properties
JOBVIEW_RELEASE_STORE_FILE=jobview-release.keystore
JOBVIEW_RELEASE_KEY_ALIAS=jobview-key-alias
JOBVIEW_RELEASE_STORE_PASSWORD=your_store_password
JOBVIEW_RELEASE_KEY_PASSWORD=your_key_password
```

#### Step 4: 构建签名 APK
```bash
cd android
./gradlew assembleRelease
```

## 🚀 部署脚本

为了简化部署流程，您可以使用以下脚本：

### 开发部署脚本 (`scripts/deploy-debug.sh`)
```bash
#!/bin/bash

echo "🚀 JobView Mobile Debug 部署开始..."

# 检查设备连接
if ! adb devices | grep -q "device$"; then
    echo "❌ 未检测到 Android 设备，请检查连接"
    exit 1
fi

echo "✅ 设备连接正常"

# 清理缓存
echo "🧹 清理缓存..."
npm run clean
npx react-native start --reset-cache &
METRO_PID=$!

# 等待 Metro 启动
sleep 5

# 部署应用
echo "📱 部署到设备..."
npx react-native run-android

echo "✅ Debug 版本部署完成!"
echo "📝 注意：请保持 Metro bundler 运行以支持热重载"
```

### 生产部署脚本 (`scripts/deploy-release.sh`)
```bash
#!/bin/bash

echo "🚀 JobView Mobile Release 部署开始..."

# 检查设备连接
if ! adb devices | grep -q "device$"; then
    echo "❌ 未检测到 Android 设备，请检查连接"
    exit 1
fi

echo "✅ 设备连接正常"

# 清理和构建
echo "🏗️ 构建 Release 版本..."
npm run clean
cd android
./gradlew assembleRelease

if [ $? -eq 0 ]; then
    echo "✅ 构建成功"

    # 卸载旧版本
    echo "🗑️ 卸载旧版本..."
    adb uninstall com.jobviewmobile 2>/dev/null

    # 安装新版本
    echo "📱 安装新版本..."
    adb install app/build/outputs/apk/release/app-release.apk

    if [ $? -eq 0 ]; then
        echo "✅ Release 版本部署完成!"
        echo "📱 应用已安装到设备，可以断开调试连接"
    else
        echo "❌ 安装失败"
        exit 1
    fi
else
    echo "❌ 构建失败"
    exit 1
fi

cd ..
```

### 使用脚本
```bash
# 给脚本执行权限
chmod +x scripts/deploy-debug.sh
chmod +x scripts/deploy-release.sh

# 运行部署
./scripts/deploy-debug.sh    # 开发版本
./scripts/deploy-release.sh  # 生产版本
```

## 🐛 常见问题和解决方案

### 1. 设备连接问题

**问题**: `adb devices` 显示 "unauthorized"
**解决方案**:
```bash
# 重启 adb 服务
adb kill-server
adb start-server

# 重新连接设备，在手机上重新授权
```

**问题**: 设备无法识别
**解决方案**:
- 检查 USB 调试是否开启
- 尝试不同的 USB 数据线
- 重启设备和电脑
- 检查驱动程序安装

### 2. 构建问题

**问题**: "SDK location not found"
**解决方案**:
```bash
# 创建 local.properties 文件
echo "sdk.dir=$ANDROID_HOME" > android/local.properties
```

**问题**: Gradle 构建失败
**解决方案**:
```bash
# 清理 Gradle 缓存
cd android
./gradlew clean
./gradlew --refresh-dependencies

# 如果还有问题，删除 .gradle 文件夹
rm -rf ~/.gradle/caches
```

**问题**: Metro bundler 错误
**解决方案**:
```bash
# 重置 Metro 缓存
npx react-native start --reset-cache

# 清理 npm 缓存
npm start -- --reset-cache

# 重新安装 node_modules
rm -rf node_modules
npm install
```

### 3. 运行时问题

**问题**: 应用崩溃或白屏
**解决方案**:
```bash
# 查看日志
adb logcat | grep ReactNativeJS

# 检查 Metro 连接
adb reverse tcp:8081 tcp:8081
```

**问题**: 网络请求失败
**解决方案**:
- 检查网络权限
- 配置网络安全策略（Android 9+）
- 确保 API 地址正确

### 4. 权限问题

**问题**: 某些功能不工作（相机、存储等）
**解决方案**:
- 检查 `AndroidManifest.xml` 中的权限声明
- 在设备设置中手动授权应用权限

## 📋 部署检查清单

### 部署前检查
- [ ] Node.js 和 npm 版本正确
- [ ] Android SDK 和工具已安装
- [ ] 环境变量配置正确
- [ ] 设备已连接并授权
- [ ] 项目依赖已安装

### 开发版本部署
- [ ] Metro bundler 正常启动
- [ ] 应用成功安装到设备
- [ ] 热重载功能正常
- [ ] 调试功能可用

### 生产版本部署
- [ ] Release APK 构建成功
- [ ] APK 文件大小合理（通常 < 50MB）
- [ ] 应用安装并启动正常
- [ ] 所有功能测试通过
- [ ] 性能表现良好

### 发布前准备
- [ ] 签名密钥已生成并安全保存
- [ ] 版本号和版本名已更新
- [ ] 应用图标和元数据正确
- [ ] 隐私政策和用户协议完善

## 🔧 高级配置

### 1. 构建优化

在 `android/app/build.gradle` 中：

```gradle
android {
    ...
    buildTypes {
        release {
            minifyEnabled true
            shrinkResources true
            proguardFiles getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro"
        }
    }
}
```

### 2. 多架构支持

```gradle
android {
    ...
    splits {
        abi {
            enable true
            reset()
            include "armeabi-v7a", "arm64-v8a", "x86", "x86_64"
            universalApk true
        }
    }
}
```

### 3. 自定义应用图标

1. 准备不同尺寸的图标文件
2. 放置在 `android/app/src/main/res/` 对应目录
3. 更新 `AndroidManifest.xml`

### 4. 网络安全配置

创建 `android/app/src/main/res/xml/network_security_config.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
    <domain-config cleartextTrafficPermitted="true">
        <domain includeSubdomains="true">localhost</domain>
        <domain includeSubdomains="true">10.0.2.2</domain>
        <domain includeSubdomains="true">your-api-domain.com</domain>
    </domain-config>
</network-security-config>
```

## 📚 相关资源

- [React Native 官方文档](https://reactnative.dev/docs/getting-started)
- [Android 开发者指南](https://developer.android.com/guide)
- [React Native Android 设置](https://reactnative.dev/docs/environment-setup)
- [Android APK 签名指南](https://developer.android.com/studio/publish/app-signing)

## 📞 技术支持

如果在部署过程中遇到问题，请：

1. 查看本文档的常见问题部分
2. 查看应用日志：`adb logcat`
3. 查看 Metro bundler 日志
4. 提交 Issue 到项目仓库

---

**祝您部署成功！** 🎉