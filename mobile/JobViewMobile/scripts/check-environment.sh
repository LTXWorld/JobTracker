#!/bin/bash

# JobView Mobile 环境检查脚本
# 用于检查 Android 开发环境是否正确配置

echo "🔍 JobView Mobile 环境检查..."

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
    local cmd=$1
    local name=$2
    if command -v $cmd &> /dev/null; then
        local version=$($cmd --version 2>/dev/null | head -n1 || echo "未知版本")
        print_message $GREEN "✅ $name: $version"
        return 0
    else
        print_message $RED "❌ $name: 未安装"
        return 1
    fi
}

# 函数：检查环境变量
check_env_var() {
    local var_name=$1
    local var_value=${!var_name}
    if [ -n "$var_value" ]; then
        print_message $GREEN "✅ $var_name: $var_value"
        return 0
    else
        print_message $RED "❌ $var_name: 未设置"
        return 1
    fi
}

# 函数：检查文件或目录
check_path() {
    local path=$1
    local name=$2
    if [ -e "$path" ]; then
        print_message $GREEN "✅ $name: $path"
        return 0
    else
        print_message $RED "❌ $name: $path (不存在)"
        return 1
    fi
}

# 开始检查
print_message $BLUE "📋 基础工具检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 检查基础命令
check_command "node" "Node.js"
check_command "npm" "npm"
check_command "npx" "npx"
check_command "java" "Java"
check_command "adb" "Android Debug Bridge"

echo

# 检查 Node.js 版本
print_message $BLUE "📋 Node.js 版本检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v node &> /dev/null; then
    node_version=$(node --version | sed 's/v//')
    major_version=$(echo $node_version | cut -d. -f1)

    if [ $major_version -ge 18 ]; then
        print_message $GREEN "✅ Node.js 版本符合要求: v$node_version (>= 18.0.0)"
    else
        print_message $RED "❌ Node.js 版本过低: v$node_version (需要 >= 18.0.0)"
    fi
else
    print_message $RED "❌ Node.js 未安装"
fi

echo

# 检查 Java 版本
print_message $BLUE "📋 Java 版本检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v java &> /dev/null; then
    java_version=$(java -version 2>&1 | head -n1 | cut -d'"' -f2)
    print_message $GREEN "✅ Java 版本: $java_version"

    # 检查是否是 JDK
    if command -v javac &> /dev/null; then
        print_message $GREEN "✅ JDK 已安装"
    else
        print_message $YELLOW "⚠️  只检测到 JRE，建议安装完整的 JDK"
    fi
else
    print_message $RED "❌ Java 未安装"
fi

echo

# 检查环境变量
print_message $BLUE "📋 环境变量检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_env_var "ANDROID_HOME"
check_env_var "JAVA_HOME"

# 检查 PATH 中的 Android 工具
if [[ ":$PATH:" == *":$ANDROID_HOME/platform-tools:"* ]]; then
    print_message $GREEN "✅ Android platform-tools 在 PATH 中"
else
    print_message $RED "❌ Android platform-tools 不在 PATH 中"
fi

if [[ ":$PATH:" == *":$ANDROID_HOME/emulator:"* ]]; then
    print_message $GREEN "✅ Android emulator 在 PATH 中"
else
    print_message $YELLOW "⚠️  Android emulator 不在 PATH 中 (可选)"
fi

echo

# 检查 Android SDK
print_message $BLUE "📋 Android SDK 检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -n "$ANDROID_HOME" ]; then
    check_path "$ANDROID_HOME" "Android SDK 根目录"
    check_path "$ANDROID_HOME/platform-tools" "Platform Tools"
    check_path "$ANDROID_HOME/build-tools" "Build Tools"
    check_path "$ANDROID_HOME/platforms" "Android Platforms"

    # 检查具体的构建工具版本
    if [ -d "$ANDROID_HOME/build-tools" ]; then
        build_tools_versions=$(ls "$ANDROID_HOME/build-tools" 2>/dev/null | sort -V | tail -n3)
        if [ -n "$build_tools_versions" ]; then
            print_message $GREEN "✅ 可用的 Build Tools 版本:"
            echo "$build_tools_versions" | while read version; do
                print_message $YELLOW "   📦 $version"
            done
        fi
    fi

    # 检查可用的 Android 平台
    if [ -d "$ANDROID_HOME/platforms" ]; then
        platforms=$(ls "$ANDROID_HOME/platforms" 2>/dev/null | sort -V | tail -n3)
        if [ -n "$platforms" ]; then
            print_message $GREEN "✅ 可用的 Android 平台:"
            echo "$platforms" | while read platform; do
                print_message $YELLOW "   📱 $platform"
            done
        fi
    fi
else
    print_message $RED "❌ ANDROID_HOME 未设置，无法检查 SDK"
fi

echo

# 检查设备连接
print_message $BLUE "📋 Android 设备检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v adb &> /dev/null; then
    # 启动 adb 服务
    adb start-server > /dev/null 2>&1

    devices=$(adb devices | grep -v "List of devices" | grep "device$")
    if [ -n "$devices" ]; then
        print_message $GREEN "✅ 检测到连接的设备:"
        echo "$devices" | while read line; do
            device_id=$(echo $line | cut -f1)
            print_message $YELLOW "   📱 $device_id"

            # 获取设备信息
            brand=$(adb -s $device_id shell getprop ro.product.brand 2>/dev/null)
            model=$(adb -s $device_id shell getprop ro.product.model 2>/dev/null)
            android_version=$(adb -s $device_id shell getprop ro.build.version.release 2>/dev/null)

            if [ -n "$brand" ] && [ -n "$model" ]; then
                print_message $YELLOW "      🏷️  $brand $model (Android $android_version)"
            fi
        done
    else
        print_message $YELLOW "⚠️  未检测到连接的设备"
        print_message $BLUE "💡 请确保："
        print_message $BLUE "   - 设备已通过 USB 连接"
        print_message $BLUE "   - 已开启开发者选项和 USB 调试"
        print_message $BLUE "   - 已在设备上授权此电脑"
    fi
else
    print_message $RED "❌ adb 命令不可用"
fi

echo

# 检查项目配置
print_message $BLUE "📋 项目配置检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f "package.json" ]; then
    print_message $GREEN "✅ package.json 存在"

    # 检查项目类型
    if grep -q "react-native" package.json; then
        print_message $GREEN "✅ React Native 项目"
    else
        print_message $YELLOW "⚠️  可能不是 React Native 项目"
    fi

    # 检查项目版本
    if grep -q '"version"' package.json; then
        version=$(grep '"version"' package.json | sed 's/.*"version": *"\([^"]*\)".*/\1/')
        print_message $GREEN "✅ 项目版本: $version"
    fi
else
    print_message $RED "❌ package.json 不存在 (请在项目根目录运行)"
fi

check_path "android" "Android 目录"
check_path "android/app" "Android App 目录"
check_path "android/build.gradle" "Android Build 配置"
check_path "android/app/build.gradle" "App Build 配置"

if [ -d "node_modules" ]; then
    print_message $GREEN "✅ node_modules 存在"
else
    print_message $YELLOW "⚠️  node_modules 不存在，请运行 npm install"
fi

echo

# 运行 React Native Doctor
print_message $BLUE "📋 React Native Doctor 检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v npx &> /dev/null; then
    print_message $BLUE "🔍 运行 React Native Doctor..."
    npx react-native doctor
else
    print_message $RED "❌ npx 不可用，无法运行 React Native Doctor"
fi

echo

# 总结
print_message $BLUE "📋 检查完成"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

print_message $GREEN "🎉 环境检查完成！"
print_message $BLUE "💡 如果发现问题，请参考部署文档进行修复："
print_message $BLUE "   docs/Android/deployment-guide.md"

echo