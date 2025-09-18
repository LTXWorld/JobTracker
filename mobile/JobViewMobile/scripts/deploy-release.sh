#!/bin/bash

# JobView Mobile Release 部署脚本
# 用于构建并部署生产版本到 Android 设备

set -e  # 遇到错误立即退出

echo "🚀 JobView Mobile Release 部署开始..."

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

# 函数：获取应用版本信息
get_app_version() {
    if [ -f "package.json" ]; then
        local version=$(grep '"version"' package.json | sed 's/.*"version": *"\([^"]*\)".*/\1/')
        echo $version
    else
        echo "unknown"
    fi
}

# 函数：获取 APK 大小
get_apk_size() {
    local apk_path=$1
    if [ -f "$apk_path" ]; then
        local size=$(ls -lh "$apk_path" | awk '{print $5}')
        echo $size
    else
        echo "unknown"
    fi
}

# 检查必需的命令
print_message $BLUE "🔍 检查环境依赖..."
check_command "node"
check_command "npm"
check_command "adb"

# 检查是否在项目根目录
if [ ! -f "package.json" ]; then
    print_message $RED "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 检查 Android 目录
if [ ! -d "android" ]; then
    print_message $RED "❌ 未找到 android 目录"
    exit 1
fi

print_message $GREEN "✅ 项目检查通过"

# 获取应用信息
APP_VERSION=$(get_app_version)
print_message $BLUE "📋 应用信息:"
print_message $YELLOW "   版本: $APP_VERSION"

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

# 询问是否继续
echo
read -p "$(echo -e ${YELLOW}⚠️  即将构建生产版本，这可能需要几分钟时间，是否继续？ [y/N]: ${NC})" -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_message $BLUE "👋 部署已取消"
    exit 0
fi

# 检查依赖
if [ ! -d "node_modules" ]; then
    print_message $YELLOW "⚠️  node_modules 不存在，正在安装依赖..."
    npm install
    if [ $? -ne 0 ]; then
        print_message $RED "❌ 依赖安装失败"
        exit 1
    fi
    print_message $GREEN "✅ 依赖安装完成"
fi

# 清理之前的构建
print_message $BLUE "🧹 清理之前的构建..."
npm run clean

# 进入 Android 目录
cd android

# 检查 Gradle wrapper
if [ ! -f "gradlew" ]; then
    print_message $RED "❌ 未找到 Gradle wrapper"
    exit 1
fi

# 给 gradlew 执行权限
chmod +x gradlew

# 构建 Release APK
print_message $BLUE "🏗️ 构建 Release APK..."
print_message $YELLOW "⏳ 这可能需要几分钟时间，请耐心等待..."

./gradlew assembleRelease

if [ $? -eq 0 ]; then
    print_message $GREEN "✅ 构建成功"

    # 检查 APK 文件
    APK_PATH="app/build/outputs/apk/release/app-release.apk"
    if [ ! -f "$APK_PATH" ]; then
        print_message $RED "❌ 未找到构建的 APK 文件"
        exit 1
    fi

    # 显示 APK 信息
    APK_SIZE=$(get_apk_size "$APK_PATH")
    print_message $BLUE "📱 APK 信息:"
    print_message $YELLOW "   文件路径: android/$APK_PATH"
    print_message $YELLOW "   文件大小: $APK_SIZE"

    # 回到项目根目录
    cd ..

    # 询问是否安装到设备
    echo
    read -p "$(echo -e ${YELLOW}📱 是否安装到连接的设备？ [Y/n]: ${NC})" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        print_message $BLUE "📱 准备安装到设备..."

        # 获取应用包名
        PACKAGE_NAME="com.jobviewmobile"

        # 检查是否已安装
        if adb shell pm list packages | grep -q "$PACKAGE_NAME"; then
            print_message $YELLOW "⚠️  检测到应用已安装"
            read -p "$(echo -e ${YELLOW}是否卸载旧版本？ [Y/n]: ${NC})" -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Nn]$ ]]; then
                print_message $BLUE "🗑️ 卸载旧版本..."
                adb uninstall $PACKAGE_NAME
                if [ $? -eq 0 ]; then
                    print_message $GREEN "✅ 旧版本卸载成功"
                else
                    print_message $YELLOW "⚠️  卸载失败，继续安装新版本"
                fi
            fi
        fi

        # 安装新版本
        print_message $BLUE "📱 安装新版本..."
        adb install "android/$APK_PATH"

        if [ $? -eq 0 ]; then
            print_message $GREEN "✅ Release 版本部署完成!"
            print_message $BLUE "🎉 应用已成功安装到设备"
            print_message $YELLOW "📝 注意事项:"
            print_message $YELLOW "   - 这是生产版本，无需连接开发服务器"
            print_message $YELLOW "   - 应用已优化，性能更好"
            print_message $YELLOW "   - 可以断开 USB 连接正常使用"

            # 询问是否启动应用
            echo
            read -p "$(echo -e ${YELLOW}🚀 是否立即启动应用？ [Y/n]: ${NC})" -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Nn]$ ]]; then
                print_message $BLUE "🚀 启动应用..."
                adb shell am start -n "$PACKAGE_NAME/.MainActivity"
                if [ $? -eq 0 ]; then
                    print_message $GREEN "✅ 应用启动成功"
                else
                    print_message $YELLOW "⚠️  自动启动失败，请手动启动应用"
                fi
            fi
        else
            print_message $RED "❌ 安装失败"
            exit 1
        fi
    else
        print_message $BLUE "📱 APK 已构建完成，可以手动安装"
        print_message $YELLOW "💡 手动安装命令:"
        print_message $YELLOW "   adb install android/$APK_PATH"
    fi

    # 显示总结
    echo
    print_message $GREEN "🎉 部署流程完成!"
    print_message $BLUE "📋 总结:"
    print_message $YELLOW "   ✅ 应用版本: $APP_VERSION"
    print_message $YELLOW "   ✅ APK 大小: $APK_SIZE"
    print_message $YELLOW "   ✅ 文件位置: android/$APK_PATH"

else
    print_message $RED "❌ 构建失败"
    cd ..
    exit 1
fi