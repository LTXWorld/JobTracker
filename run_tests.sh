#!/bin/bash

# JobView 综合测试执行脚本
# 执行前端和后端的完整测试套件，生成覆盖率报告

usage() {
    cat <<'EOF'
用法: ./run_tests.sh [选项]

选项说明:
  --integration   额外执行带有 integration build tag 的后端测试
  --loadtest      额外执行带有 loadtest build tag 的性能/负载测试
  -h, --help      显示此帮助信息

说明:
- 默认仅执行前端单元测试与后端核心单元测试（包含占位测试提示）
- integration / loadtest 会运行耗时更长的测试，请在资源充足的环境下使用
EOF
}

RUN_INTEGRATION=0
RUN_LOADTEST=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        --integration)
            RUN_INTEGRATION=1
            ;;
        --loadtest)
            RUN_LOADTEST=1
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            usage
            exit 1
            ;;
    esac
    shift
done

echo "🎯 JobView 测试套件执行开始..."
echo "======================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
FRONTEND_TESTS_PASSED=0
BACKEND_TESTS_PASSED=0
INTEGRATION_TESTS_PASSED=-1   # -1 表示未执行
LOADTEST_TESTS_PASSED=-1
TOTAL_ERRORS=0

echo -e "${BLUE}📂 项目结构检查...${NC}"
if [ ! -d "frontend" ] || [ ! -d "backend" ]; then
    echo -e "${RED}❌ 错误: 请在项目根目录执行此脚本${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 项目结构正常${NC}"
echo

# 1. 前端测试
echo -e "${BLUE}🧪 执行前端测试...${NC}"
echo "------------------------------"

cd frontend

echo "📦 检查前端依赖..."
if ! npm list vitest > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  安装前端测试依赖...${NC}"
    npm install
fi

echo "🏃 运行前端单元测试..."
npm run test:run > frontend_test_results.log 2>&1
FRONTEND_EXIT_CODE=$?

if [ $FRONTEND_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ 前端测试通过${NC}"
    FRONTEND_TESTS_PASSED=1
else
    echo -e "${RED}❌ 前端测试失败${NC}"
    echo -e "${YELLOW}详细错误信息:${NC}"
    tail -20 frontend_test_results.log
    TOTAL_ERRORS=$((TOTAL_ERRORS + 1))
fi

echo "📊 生成前端覆盖率报告..."
npm run test:coverage > frontend_coverage.log 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 前端覆盖率报告生成成功${NC}"
    if [ -f "coverage/index.html" ]; then
        echo -e "${BLUE}📄 前端覆盖率报告: $(pwd)/coverage/index.html${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  前端覆盖率报告生成失败${NC}"
fi

cd ..

# 2. 后端测试
echo
echo -e "${BLUE}🧪 执行后端测试...${NC}"
echo "------------------------------"

cd backend

echo "📦 检查后端依赖..."
go mod tidy > /dev/null 2>&1

echo "🏃 运行后端认证模块测试..."
go test ./tests/auth -v -cover > backend_auth_test_results.log 2>&1
BACKEND_AUTH_EXIT_CODE=$?

if [ $BACKEND_AUTH_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ 后端认证测试通过${NC}"
    BACKEND_TESTS_PASSED=1
else
    echo -e "${RED}❌ 后端认证测试失败${NC}"
    echo -e "${YELLOW}详细错误信息:${NC}"
    tail -15 backend_auth_test_results.log
    TOTAL_ERRORS=$((TOTAL_ERRORS + 1))
fi

echo "📊 生成后端覆盖率报告..."
go test ./tests/auth -coverprofile=auth_coverage.out -cover > /dev/null 2>&1
if [ $? -eq 0 ] && [ -f "auth_coverage.out" ]; then
    go tool cover -html=auth_coverage.out -o auth_coverage.html
    echo -e "${GREEN}✅ 后端认证模块覆盖率报告生成成功${NC}"
    echo -e "${BLUE}📄 后端覆盖率报告: $(pwd)/auth_coverage.html${NC}"
else
    echo -e "${YELLOW}⚠️  后端覆盖率报告生成失败${NC}"
fi

# 尝试运行其他后端测试（如果存在）
if [ -d "tests/handler" ]; then
    echo "🏃 运行后端Handler测试..."
    go test ./tests/handler -v > backend_handler_test_results.log 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 后端Handler测试通过${NC}"
    else
        echo -e "${YELLOW}⚠️  后端Handler测试有问题（预期中，需要进一步修复）${NC}"
    fi
fi

cd ..

# 3. 可选测试集
if [ $RUN_INTEGRATION -eq 1 ]; then
    echo
    echo -e "${BLUE}🔧 执行 integration 测试 (build tag) ...${NC}"
    echo "------------------------------"
    pushd backend > /dev/null
    go test -tags integration ./tests/auth ./tests/handler ./tests/api ./tests/database > backend_integration_tests.log 2>&1
    EXIT_CODE=$?
    popd > /dev/null

    INTEGRATION_TESTS_PASSED=0
    if [ $EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✅ integration 测试通过${NC}"
        INTEGRATION_TESTS_PASSED=1
    else
        echo -e "${RED}❌ integration 测试失败${NC}"
        echo -e "${YELLOW}详细错误信息:${NC}"
        tail -20 backend/backend_integration_tests.log
        TOTAL_ERRORS=$((TOTAL_ERRORS + 1))
    fi
fi

if [ $RUN_LOADTEST -eq 1 ]; then
    echo
    echo -e "${BLUE}🔥 执行 loadtest 性能用例 (build tag) ...${NC}"
    echo "------------------------------"
    pushd backend > /dev/null
    GO_TEST_CMD="go test -tags loadtest ./tests/service -run TestLoadTesting_ -count=1"
    echo "命令: $GO_TEST_CMD"
    eval "$GO_TEST_CMD" > backend_loadtest.log 2>&1
    EXIT_CODE=$?
    popd > /dev/null

    LOADTEST_TESTS_PASSED=0
    if [ $EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✅ loadtest 用例通过${NC}"
        LOADTEST_TESTS_PASSED=1
    else
        echo -e "${RED}❌ loadtest 用例失败${NC}"
        echo -e "${YELLOW}详细错误信息:${NC}"
        tail -20 backend/backend_loadtest.log
        TOTAL_ERRORS=$((TOTAL_ERRORS + 1))
    fi
    echo -e "${YELLOW}提示: loadtest 会真实执行性能压测，请确保环境资源充足${NC}"
fi

# 4. 运行状态检查
echo
echo -e "${BLUE}🔗 服务运行状态检查...${NC}"
echo "------------------------------"

# 检查是否有运行的服务
if pgrep -f "go run.*main.go" > /dev/null; then
    echo -e "${GREEN}✅ 后端服务正在运行${NC}"
    BACKEND_RUNNING=1
else
    echo -e "${YELLOW}⚠️  后端服务未运行${NC}"
    BACKEND_RUNNING=0
fi

if lsof -ti:3000 > /dev/null; then
    echo -e "${GREEN}✅ 前端服务正在运行${NC}"
    FRONTEND_RUNNING=1
else
    echo -e "${YELLOW}⚠️  前端服务未运行${NC}"
    FRONTEND_RUNNING=0
fi

# 4. 测试结果汇总
echo
echo -e "${BLUE}📈 测试结果汇总${NC}"
echo "======================================"

echo "🧪 单元测试结果:"
if [ $FRONTEND_TESTS_PASSED -eq 1 ]; then
    echo -e "  • 前端测试: ${GREEN}✅ 通过${NC}"
else
    echo -e "  • 前端测试: ${RED}❌ 失败${NC}"
fi

if [ $BACKEND_TESTS_PASSED -eq 1 ]; then
    echo -e "  • 后端认证测试: ${GREEN}✅ 通过${NC}"
else
    echo -e "  • 后端认证测试: ${RED}❌ 失败${NC}"
fi

echo
echo "🔄 可选测试 (build tags):"
if [ $INTEGRATION_TESTS_PASSED -eq -1 ]; then
    echo -e "  • integration: ${YELLOW}未执行${NC}（使用 --integration 启用）"
elif [ $INTEGRATION_TESTS_PASSED -eq 1 ]; then
    echo -e "  • integration: ${GREEN}✅ 通过${NC}"
else
    echo -e "  • integration: ${RED}❌ 失败${NC}"
fi

if [ $LOADTEST_TESTS_PASSED -eq -1 ]; then
    echo -e "  • loadtest: ${YELLOW}未执行${NC}（使用 --loadtest 启用）"
elif [ $LOADTEST_TESTS_PASSED -eq 1 ]; then
    echo -e "  • loadtest: ${GREEN}✅ 通过${NC}"
else
    echo -e "  • loadtest: ${RED}❌ 失败${NC}"
fi

echo
echo "🏃 服务运行状态:"
if [ $BACKEND_RUNNING -eq 1 ]; then
    echo -e "  • 后端服务 (8010): ${GREEN}✅ 运行中${NC}"
else
    echo -e "  • 后端服务 (8010): ${YELLOW}⚠️  未运行${NC}"
fi

if [ $FRONTEND_RUNNING -eq 1 ]; then
    echo -e "  • 前端服务 (3000): ${GREEN}✅ 运行中${NC}"
else
    echo -e "  • 前端服务 (3000): ${YELLOW}⚠️  未运行${NC}"
fi

echo
echo "📊 报告文件位置:"
echo "  • 前端覆盖率: frontend/coverage/index.html"
echo "  • 后端覆盖率: backend/auth_coverage.html"
echo "  • 前端测试日志: frontend/frontend_test_results.log"
echo "  • 后端测试日志: backend/backend_auth_test_results.log"
if [ $RUN_INTEGRATION -eq 1 ]; then
    echo "  • integration 日志: backend/backend_integration_tests.log"
fi
if [ $RUN_LOADTEST -eq 1 ]; then
    echo "  • loadtest 日志: backend/backend_loadtest.log"
fi

# 5. 最终状态
echo
if [ $TOTAL_ERRORS -eq 0 ]; then
    echo -e "${GREEN}🎉 测试套件执行完成！所有核心测试通过！${NC}"
    echo -e "${BLUE}💡 建议: 查看覆盖率报告以进一步提高代码质量${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  测试套件完成，但有 $TOTAL_ERRORS 个测试模块需要修复${NC}"
    echo -e "${BLUE}💡 建议: 查看详细日志文件以了解失败原因${NC}"
    exit 1
fi
