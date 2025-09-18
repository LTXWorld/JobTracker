# JobView Mobile Android 故障排除指南

## 🔧 常见问题与解决方案

本文档包含 JobView Mobile Android 部署和运行过程中常见问题的解决方案。

## 📱 设备连接问题

### 问题 1: 设备未显示在 adb devices 中

**症状：**
```bash
$ adb devices
List of devices attached
```

**可能原因：**
- USB 调试未开启
- 设备驱动问题
- USB 线缆问题
- ADB 服务问题

**解决方案：**

1. **检查 USB 调试设置**
   ```
   设置 → 开发者选项 → USB 调试 (确保开启)
   设置 → 开发者选项 → 允许 USB 安装 (如果有，请开启)
   ```

2. **重启 ADB 服务**
   ```bash
   adb kill-server
   adb start-server
   adb devices
   ```

3. **检查 USB 连接**
   - 尝试不同的 USB 端口
   - 使用原装或质量好的数据线
   - 确保数据线支持数据传输（不仅仅是充电）

4. **重置 USB 授权**
   ```
   设置 → 开发者选项 → 撤销 USB 调试授权
   重新连接设备并授权
   ```

### 问题 2: 设备显示 "unauthorized"

**症状：**
```bash
$ adb devices
List of devices attached
12345678    unauthorized
```

**解决方案：**
1. 查看手机屏幕是否有授权弹窗
2. 选择"允许"并勾选"始终允许来自此计算机"
3. 如果没有弹窗，断开重连 USB 线
4. 尝试重启 ADB：`adb kill-server && adb start-server`

### 问题 3: 多个设备连接时的冲突

**症状：**
```bash
$ adb devices
List of devices attached
device1     device
device2     device
emulator-5554   device
```

**解决方案：**
```bash
# 指定特定设备进行操作
adb -s device1 install app.apk
adb -s device1 shell

# 或者断开其他设备，只保留目标设备
```

## 🏗️ 构建问题

### 问题 4: Gradle 构建失败

**症状：**
```
FAILURE: Build failed with an exception.
```

**常见解决方案：**

1. **清理构建缓存**
   ```bash
   cd android
   ./gradlew clean
   ./gradlew --refresh-dependencies
   cd ..
   npm run clean
   ```

2. **删除 .gradle 缓存**
   ```bash
   rm -rf ~/.gradle/caches
   cd android
   ./gradlew clean
   ```

3. **检查 Java 版本**
   ```bash
   java -version
   # 确保使用 JDK 11 或 17
   ```

4. **检查 Android SDK 配置**
   ```bash
   echo $ANDROID_HOME
   # 应该指向正确的 SDK 目录
   ```

### 问题 5: Metro bundler 启动失败

**症状：**
```
error Failed to start the packager.
```

**解决方案：**

1. **重置 Metro 缓存**
   ```bash
   npx react-native start --reset-cache
   ```

2. **检查端口占用**
   ```bash
   # 查看 8081 端口是否被占用
   lsof -i :8081

   # 如果被占用，终止进程
   kill -9 <PID>
   ```

3. **清理 npm 缓存**
   ```bash
   npm start -- --reset-cache
   rm -rf node_modules
   npm install
   ```

### 问题 6: SDK 版本不匹配

**症状：**
```
Android SDK Build-Tools 33.0.0 is missing
```

**解决方案：**

1. **通过 Android Studio 安装**
   - 打开 Android Studio
   - SDK Manager → SDK Tools
   - 安装所需的 Build Tools 版本

2. **通过命令行安装**
   ```bash
   sdkmanager "build-tools;33.0.0"
   sdkmanager "platforms;android-33"
   ```

## 🔗 网络和 API 问题

### 问题 7: Metro bundler 连接失败

**症状：**
- 应用启动后显示红屏
- "Could not connect to Metro bundler"

**解决方案：**

1. **设置端口转发**
   ```bash
   adb reverse tcp:8081 tcp:8081
   ```

2. **检查防火墙设置**
   - 确保防火墙允许 8081 端口
   - 临时关闭防火墙测试

3. **使用设备 IP**
   ```bash
   # 在 Bundle 配置中指定设备 IP
   npx react-native start --host <YOUR_IP>
   ```

### 问题 8: API 请求失败

**症状：**
- 网络请求超时
- "Network request failed"

**解决方案：**

1. **检查网络安全配置**
   创建 `android/app/src/main/res/xml/network_security_config.xml`：
   ```xml
   <?xml version="1.0" encoding="utf-8"?>
   <network-security-config>
       <domain-config cleartextTrafficPermitted="true">
           <domain includeSubdomains="true">localhost</domain>
           <domain includeSubdomains="true">10.0.2.2</domain>
       </domain-config>
   </network-security-config>
   ```

2. **在 AndroidManifest.xml 中应用配置**
   ```xml
   <application
       android:networkSecurityConfig="@xml/network_security_config"
       ... >
   ```

3. **检查 API 地址配置**
   - 确保 API 地址正确
   - 模拟器使用 `10.0.2.2` 代替 `localhost`
   - 真机使用电脑的实际 IP 地址

## 📦 依赖和版本问题

### 问题 9: React Native 版本冲突

**症状：**
```
error React Native version mismatch
```

**解决方案：**

1. **检查版本一致性**
   ```bash
   npx react-native --version
   cat package.json | grep react-native
   ```

2. **重新安装依赖**
   ```bash
   rm -rf node_modules
   rm package-lock.json
   npm install
   ```

3. **更新到兼容版本**
   ```bash
   npm update react-native
   npx react-native upgrade
   ```

### 问题 10: 原生模块链接问题

**症状：**
- 某些第三方库功能不工作
- 编译时找不到原生模块

**解决方案：**

1. **重新链接原生模块**
   ```bash
   cd android
   ./gradlew clean
   cd ..
   npx react-native unlink <library-name>
   npx react-native link <library-name>
   ```

2. **手动配置 (如果自动链接失败)**
   - 查看库的官方文档
   - 手动修改 `android/settings.gradle`
   - 手动修改 `android/app/build.gradle`

## 🔐 权限问题

### 问题 11: 应用权限被拒绝

**症状：**
- 相机、存储等功能不工作
- 权限请求失败

**解决方案：**

1. **检查权限声明**
   在 `android/app/src/main/AndroidManifest.xml` 中：
   ```xml
   <uses-permission android:name="android.permission.CAMERA" />
   <uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
   <uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
   ```

2. **手动授权权限**
   ```
   设置 → 应用 → JobView Mobile → 权限
   开启所需的权限
   ```

3. **在代码中请求权限**
   ```javascript
   import { PermissionsAndroid } from 'react-native';

   const requestPermission = async () => {
     const granted = await PermissionsAndroid.request(
       PermissionsAndroid.PERMISSIONS.CAMERA
     );
     return granted === PermissionsAndroid.RESULTS.GRANTED;
   };
   ```

## 🚀 性能问题

### 问题 12: 应用启动缓慢

**症状：**
- 应用启动时间过长
- 白屏时间较长

**解决方案：**

1. **启用 Hermes 引擎**
   在 `android/app/build.gradle` 中：
   ```gradle
   project.ext.react = [
       enableHermes: true
   ]
   ```

2. **优化 Bundle 大小**
   ```bash
   # 生成分析报告
   npx react-native bundle --platform android --dev false --entry-file index.js --bundle-output android/app/src/main/assets/index.android.bundle --sourcemap-output sourcemap.js
   ```

3. **使用 ProGuard 压缩**
   在 `android/app/build.gradle` 中：
   ```gradle
   buildTypes {
       release {
           minifyEnabled true
           shrinkResources true
           proguardFiles getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro"
       }
   }
   ```

### 问题 13: 内存泄漏

**症状：**
- 应用使用过程中变卡顿
- 内存使用持续增长

**解决方案：**

1. **使用 Flipper 调试**
   ```bash
   # 安装 Flipper
   # 连接设备查看内存使用情况
   ```

2. **检查组件卸载**
   ```javascript
   useEffect(() => {
       // 设置定时器或监听器
       return () => {
           // 清理资源
       };
   }, []);
   ```

## 🔍 调试技巧

### 查看日志

1. **React Native 日志**
   ```bash
   adb logcat | grep ReactNativeJS
   ```

2. **Android 系统日志**
   ```bash
   adb logcat
   ```

3. **特定标签日志**
   ```bash
   adb logcat -s ReactNative
   adb logcat -s ReactNativeJS
   ```

### 远程调试

1. **Chrome 开发者工具**
   - 摇晃设备
   - 选择 "Debug"
   - 在 Chrome 中打开 `http://localhost:8081/debugger-ui`

2. **Flipper 调试**
   - 安装 Flipper 桌面应用
   - 连接设备进行网络、布局、性能调试

### 性能分析

1. **FPS 监控**
   ```javascript
   // 在开发者菜单中开启 FPS Monitor
   ```

2. **Bundle 分析**
   ```bash
   npx react-native-bundle-visualizer
   ```

## 📚 有用的命令

### 快速诊断
```bash
# 环境检查
npx react-native doctor

# 设备信息
adb shell getprop | grep -E "(model|brand|version)"

# 应用信息
adb shell pm list packages | grep jobview
adb shell pm path com.jobviewmobile

# 清理所有缓存
npm run clean && rm -rf node_modules && npm install
```

### 日志收集
```bash
# 收集完整日志
adb logcat > debug.log &
# 重现问题后停止日志收集
kill %1
```

### 应用调试
```bash
# 启动应用
adb shell am start -n com.jobviewmobile/.MainActivity

# 停止应用
adb shell am force-stop com.jobviewmobile

# 清除应用数据
adb shell pm clear com.jobviewmobile
```

---

## 🆘 寻求帮助

如果上述解决方案都无法解决您的问题：

1. **收集诊断信息**
   ```bash
   ./scripts/check-environment.sh > environment-info.txt
   adb logcat > app-logs.txt
   ```

2. **提供详细信息**
   - 操作系统版本
   - Android 设备型号和版本
   - 具体错误信息
   - 重现步骤

3. **联系支持**
   - 查看项目 Issues
   - 提交新的 Issue 包含诊断信息

记住：大多数问题都有解决方案，保持耐心并系统性地排查问题！ 🚀