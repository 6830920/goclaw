#!/bin/bash
# Goclaw项目实时错误检测和构建验证脚本

PROJECT_DIR="${PROJECT_DIR:-/home/daniel/projects/goclaw}"

echo "🔍 检查Goclaw项目: $PROJECT_DIR"

# 切换到项目目录
cd $PROJECT_DIR || { echo "❌ 无法进入项目目录: $PROJECT_DIR"; exit 1; }

# 设置国内代理以解决网络问题
export GOPROXY=https://goproxy.cn,direct

# 1. 语法检查
echo "📝 运行 go vet..."
go vet ./... 2>&1
VET_RESULT=$?
if [ $VET_RESULT -ne 0 ]; then
    echo "❌ go vet 发现问题:"
    go vet ./... 2>&1
    exit 1
else
    echo "✅ go vet 通过"
fi

# 2. 导入检查
echo "📦 检查未使用导入..."
go mod tidy
UNUSED_IMPORTS=$(go vet -printfuncs="fmt.Printf,fmt.Sprintf" ./... 2>&1 | grep -i "imported and not used")
if [ -n "$UNUSED_IMPORTS" ]; then
    echo "❌ 发现未使用的导入:"
    echo "$UNUSED_IMPORTS"
    exit 1
else
    echo "✅ 未发现未使用的导入"
fi

# 3. 构建测试
echo "🔨 尝试构建服务器..."
go build -o bin/test-build-$$.tmp ./cmd/server 2>&1
BUILD_RESULT=$?
if [ $BUILD_RESULT -ne 0 ]; then
    echo "❌ 构建失败:"
    go build -o bin/test-build-$$.tmp ./cmd/server 2>&1
    exit 1
else
    echo "✅ 构建成功"
    rm -f bin/test-build-$$.tmp
fi

# 4. 依赖检查
echo "🔗 检查依赖..."
go list -m all > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ 依赖检查失败"
    go list -m all
    exit 1
else
    echo "✅ 依赖检查通过"
fi

# 5. 基础测试
echo "🧪 运行基础测试..."
go test ./internal/vector -v 2>&1
TEST_RESULT=$?
if [ $TEST_RESULT -ne 0 ]; then
    echo "❌ 测试失败"
    exit 1
else
    echo "✅ 测试通过"
fi

echo "🎉 所有检查通过！项目状态良好。"