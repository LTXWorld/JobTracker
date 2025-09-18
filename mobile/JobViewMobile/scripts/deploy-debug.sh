#!/bin/bash

# JobView Mobile Debug 部署脚本
# 用于快速部署开发版本到 Android 设备

set -e  # 遇到错误立即退出

echo "🚀 JobView Mobile Debug 部署开始..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数：打印彩色消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 函数：检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_message $RED "❌ $1 命令未找到，请先安装"
        exit 1
    fi
}

# 函数：检查 Node.js 版本
check_node_version() {
    local node_version=$(node --version | sed 's/v//')
    local major_version=$(echo $node_version | cut -d. -f1)

    if [ $major_version -lt 18 ]; then
        print_message $RED "❌ Node.js 版本过低，需要 18.0.0 或更高版本"
        print_message $YELLOW "💡 当前版本: v$node_version"
        exit 1
    fi

    print_message $GREEN "✅ Node.js 版本检查通过: v$node_version"
}

# 检查必需的命令
print_message $BLUE "🔍 检查环境依赖..."
check_command "node"
check_command "npm"
check_command "adb"
check_command "npx"

# 检查版本
check_node_version

# 检查是否在项目根目录
if [ ! -f "package.json" ]; then
    print_message $RED "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 检查是否是 React Native 项目
if ! grep -q "react-native" package.json; then
    print_message $RED "❌ 这不是一个 React Native 项目"
    exit 1
fi

print_message $GREEN "✅ 项目检查通过"

# 检查设备连接
print_message $BLUE "📱 检查 Android 设备连接..."
if ! adb devices | grep -q "device$"; then
    print_message $RED "❌ 未检测到 Android 设备"
    print_message $YELLOW "💡 请确保："
    print_message $YELLOW "   - 设备已通过 USB 连接"
    print_message $YELLOW "   - 已开启 USB 调试"
    print_message $YELLOW "   - 已在设备上授权此电脑"
    exit 1
fi

print_message $GREEN "✅ 设备连接正常"

# 显示连接的设备
print_message $BLUE "📋 已连接的设备:"
adb devices | grep "device$" | while read line; do
    device_id=$(echo $line | cut -f1)
    print_message $YELLOW "   📱 $device_id"
done

# 检查依赖是否已安装
if [ ! -d "node_modules" ]; then
    print_message $YELLOW "⚠️  node_modules 不存在，正在安装依赖..."
    npm install
    if [ $? -ne 0 ]; then
        print_message $RED "❌ 依赖安装失败"
        exit 1
    fi
    print_message $GREEN "✅ 依赖安装完成"
else
    print_message $GREEN "✅ 依赖已存在"
fi

# 询问是否清理缓存
read -p "$(echo -e ${YELLOW}🧹 是否清理缓存？这可能需要一些时间 [y/N]: ${NC})" -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_message $BLUE "🧹 清理缓存..."

    # 清理 React Native 缓存
    npx react-native start --reset-cache &
    METRO_PID=$!
    sleep 2
    kill $METRO_PID 2>/dev/null || true

    # 清理 Gradle 缓存
    if [ -d "android" ]; then
        cd android
        ./gradlew clean
        cd ..
    fi

    print_message $GREEN "✅ 缓存清理完成"
fi

# 启动 Metro bundler
print_message $BLUE "📦 启动 Metro bundler..."
npx react-native start &
METRO_PID=$!

# 等待 Metro 启动
print_message $YELLOW "⏳ 等待 Metro bundler 启动..."
sleep 10

# 检查 Metro 是否成功启动
if ! ps -p $METRO_PID > /dev/null; then
    print_message $RED "❌ Metro bundler 启动失败"
    exit 1
fi

print_message $GREEN "✅ Metro bundler 启动成功 (PID: $METRO_PID)"

# 设置端口转发
print_message $BLUE "🔗 设置端口转发..."
adb reverse tcp:8081 tcp:8081

# 构建并安装应用
print_message $BLUE "🏗️ 构建并安装应用..."
npx react-native run-android

if [ $? -eq 0 ]; then
    print_message $GREEN "✅ Debug 版本部署完成!"
    print_message $BLUE "📱 应用已安装到设备"
    print_message $YELLOW "📝 注意事项:"
    print_message $YELLOW "   - Metro bundler 正在运行 (PID: $METRO_PID)"
    print_message $YELLOW "   - 支持热重载功能"
    print_message $YELLOW "   - 可以通过摇晃设备打开开发者菜单"
    print_message $YELLOW "   - 使用 Ctrl+C 停止 Metro bundler"

    # 等待用户输入停止
    echo
    read -p "$(echo -e ${BLUE}按 Enter 键停止 Metro bundler 并退出...${NC})"

    # 停止 Metro bundler
    kill $METRO_PID 2>/dev/null || true
    print_message $GREEN "👋 部署脚本已退出"
else
    print_message $RED "❌ 应用安装失败"
    # 停止 Metro bundler
    kill $METRO_PID 2>/dev/null || true
    exit 1
fi