# JobView Mobile Android 文档中心

## 📚 文档导航

欢迎来到 JobView Mobile Android 部署文档中心！这里包含了将应用部署到 Android 设备所需的所有信息。

## 🚀 开始使用

### 新手推荐路径
1. **[快速开始指南](quick-start.md)** - 适合想要快速部署应用的用户
2. **[环境检查脚本](../../scripts/check-environment.sh)** - 检查开发环境是否正确配置
3. **[故障排除指南](troubleshooting.md)** - 遇到问题时的解决方案

### 深入了解
- **[完整部署指南](deployment-guide.md)** - 详细的环境配置和部署流程

## 📋 文档列表

| 文档 | 描述 | 适用人群 |
|------|------|----------|
| [🚀 快速开始](quick-start.md) | 一键部署指南，最简单的部署方式 | 新手用户 |
| [📖 完整部署指南](deployment-guide.md) | 详细的环境设置和部署流程说明 | 开发者 |
| [🔧 故障排除](troubleshooting.md) | 常见问题和解决方案汇总 | 遇到问题的用户 |

## 🛠️ 部署脚本

项目提供了自动化部署脚本，位于 `scripts/` 目录：

| 脚本 | 功能 | 使用场景 |
|------|------|----------|
| `check-environment.sh` | 检查开发环境配置 | 部署前环境验证 |
| `deploy-debug.sh` | 部署开发版本 | 开发和调试 |
| `deploy-release.sh` | 部署生产版本 | 日常使用 |

### 使用方法
```bash
# 给脚本执行权限
chmod +x scripts/*.sh

# 环境检查
./scripts/check-environment.sh

# 开发版本部署
./scripts/deploy-debug.sh

# 生产版本部署
./scripts/deploy-release.sh
```

## 📱 快速部署流程

### 第一次部署
1. **准备手机**
   - 开启开发者选项和 USB 调试
   - 使用 USB 线连接电脑

2. **环境检查**
   ```bash
   ./scripts/check-environment.sh
   ```

3. **一键部署**
   ```bash
   ./scripts/deploy-debug.sh
   ```

### 日常使用
如果您已经配置好环境，可以直接使用：
```bash
# 开发版本（支持热重载）
./scripts/deploy-debug.sh

# 生产版本（性能更好）
./scripts/deploy-release.sh
```

## 🎯 选择合适的部署方式

### 开发版本 (Debug)
- ✅ 支持热重载和实时调试
- ✅ 错误信息详细
- ⚠️ 需要保持电脑连接
- ⚠️ 性能较慢，APK 较大

**适用场景：**
- 开发和调试
- 功能测试
- 学习 React Native

### 生产版本 (Release)
- ✅ 性能优化，运行流畅
- ✅ APK 体积小
- ✅ 可以断开电脑独立使用
- ❌ 不支持调试功能

**适用场景：**
- 日常使用
- 性能测试
- 应用分发

## 🔍 常见问题快速解答

### Q: 找不到设备怎么办？
A: 确保 USB 调试已开启，运行 `adb devices` 检查连接。详见 [故障排除指南](troubleshooting.md#设备连接问题)

### Q: 构建失败怎么办？
A: 尝试清理缓存：`npm run clean`，然后重新部署。详见 [故障排除指南](troubleshooting.md#构建问题)

### Q: 应用无法连接网络？
A: 检查网络安全配置和 API 地址设置。详见 [故障排除指南](troubleshooting.md#网络和api问题)

### Q: 权限被拒绝？
A: 检查 AndroidManifest.xml 权限声明，手动在设置中授权。详见 [故障排除指南](troubleshooting.md#权限问题)

## 📚 相关资源

### 官方文档
- [React Native 官方文档](https://reactnative.dev/docs/getting-started)
- [Android 开发者指南](https://developer.android.com/guide)
- [React Native 环境配置](https://reactnative.dev/docs/environment-setup)

### 开发工具
- [Android Studio](https://developer.android.com/studio) - Android 开发 IDE
- [Flipper](https://fbflipper.com/) - React Native 调试工具
- [Reactotron](https://github.com/infinitered/reactotron) - React Native 调试器

### 社区资源
- [React Native 中文网](https://reactnative.cn/)
- [React Native 官方社区](https://github.com/facebook/react-native)

## 📞 获取帮助

### 自助解决
1. 查看 [故障排除指南](troubleshooting.md)
2. 运行环境检查：`./scripts/check-environment.sh`
3. 查看应用日志：`adb logcat | grep ReactNativeJS`

### 联系支持
如果问题仍未解决：
1. 收集错误信息和环境信息
2. 查看项目 Issues
3. 提交新的 Issue 并附带详细信息

## 🔄 文档更新

本文档会随着项目的发展持续更新。如果您发现：
- 文档内容过时
- 步骤不准确
- 有改进建议

欢迎提交 Issue 或 Pull Request！

---

**祝您使用愉快！** 🎉

> 💡 **提示**: 建议先从 [快速开始指南](quick-start.md) 开始，它会引导您完成整个部署流程。