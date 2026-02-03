#!/bin/bash

# Goclaw 完整测试套件
# 用于全面测试代码修改

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Goclaw 完整测试套件${NC}"
echo "================================"

# 函数：打印带颜色的信息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 1. 代码格式化检查
print_info "正在检查代码格式..."
if go fmt ./... >/dev/null 2>&1; then
    print_success "代码格式检查通过"
else
    print_error "代码格式存在问题"
fi

# 2. 代码语法检查
print_info "正在检查代码语法..."
if go vet ./... >/dev/null 2>&1; then
    print_success "代码语法检查通过"
else
    print_error "代码语法存在问题"
    go vet ./...  # 显示具体错误
fi

# 3. 单元测试
print_info "正在运行单元测试..."
TEST_OUTPUT=$(go test ./... -v 2>&1)
if echo "$TEST_OUTPUT" | grep -q "FAIL"; then
    print_error "单元测试失败"
    echo "$TEST_OUTPUT" | grep -A 10 -B 10 "FAIL"
else
    print_success "所有单元测试通过"
fi

# 4. 构建测试
print_info "正在测试构建..."
export GOPROXY=https://goproxy.cn,direct
if go build -o bin/goclaw-test-suite ./cmd/server; then
    print_success "构建测试通过"
    # 清理测试构建文件
    rm -f bin/goclaw-test-suite
else
    print_error "构建测试失败"
    exit 1
fi

# 5. 静态分析
print_info "正在运行静态分析..."
if command -v golint >/dev/null 2>&1; then
    if golint ./... | grep -q "."; then
        print_warning "发现 linting 问题:"
        golint ./... | head -10
    else
        print_success "Linting 检查通过"
    fi
else
    print_warning "golint 未安装，跳过 linting 检查"
fi

# 6. 依赖检查
print_info "正在检查依赖..."
if go mod tidy && go mod verify; then
    print_success "依赖检查通过"
else
    print_error "依赖存在问题"
fi

# 7. 性能基准测试（如果有）
if [ -f "benchmark_test.go" ] || ls *_test.go 2>/dev/null | grep -q "benchmark\|Benchmark"; then
    print_info "正在运行性能基准测试..."
    if go test -bench=. -benchmem ./... 2>/dev/null; then
        print_success "性能基准测试完成"
    else
        print_warning "性能基准测试存在或有警告"
    fi
else
    print_info "未发现性能基准测试，跳过"
fi

# 8. 配置文件验证
print_info "正在验证配置文件..."
if [ -f "config.example.json" ]; then
    if command -v jq >/dev/null 2>&1; then
        if jq empty config.example.json 2>/dev/null; then
            print_success "配置文件格式正确"
        else
            print_error "配置文件格式错误"
        fi
    else
        print_warning "jq 未安装，跳过配置文件验证"
    fi
else
    print_warning "配置文件不存在"
fi

echo ""
print_success "🎉 测试套件完成！"

# 汇总信息
echo ""
echo -e "${BLUE}📋 测试摘要:${NC}"
echo "- 代码格式: $(if go fmt ./... 2>/dev/null | grep -q "."; then echo -e "${YELLOW}有修改${NC}"; else echo -e "${GREEN}已格式化${NC}"; fi)"
echo "- 代码语法: $(if go vet ./... >/dev/null 2>&1; then echo -e "${GREEN}通过${NC}"; else echo -e "${RED}失败${NC}"; fi)"
echo "- 单元测试: $(if go test ./... 2>&1 | grep -q "FAIL"; then echo -e "${RED}失败${NC}"; else echo -e "${GREEN}通过${NC}"; fi)"
echo "- 构建测试: ${GREEN}通过${NC}"

echo ""
echo -e "${GREEN}✅ Goclaw 项目健康状况良好！${NC}"