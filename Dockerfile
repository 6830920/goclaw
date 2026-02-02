# Multi-stage Dockerfile for Goclaw
# 使用多阶段构建以减小最终镜像大小

# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装依赖
RUN apk add --no-cache git ca-certificates

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o goclaw-server -ldflags="-w -s" ./cmd/server

# 最终阶段
FROM alpine:latest

# 安装ca-certificates以支持HTTPS
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN adduser -D -s /bin/sh goclaw

# 设置工作目录
WORKDIR /app

# 从builder阶段复制二进制文件
COPY --from=builder /app/goclaw-server .

# 更改文件所有权
RUN chown goclaw:goclaw goclaw-server

# 切换到非root用户
USER goclaw

# 暴露端口
EXPOSE 55789

# 启动命令
CMD ["./goclaw-server"]