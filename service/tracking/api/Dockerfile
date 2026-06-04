# Go Zero Dockerfile
# 用途：构建 tracking-api 镜像

# ========== 构建阶段 ==========
FROM golang:1.21-alpine AS builder

# 设置 Go 代理（加速国内下载）
ENV GOPROXY=https://goproxy.cn,direct

# 安装必要工具
RUN apk update && apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用（关闭 CGO，生成静态链接）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/tracking-api tracking-api.go

# ========== 运行阶段 ==========
FROM alpine:latest

# 安装必要运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录
WORKDIR /app

# 从构建阶段复制应用
COPY --from=builder /app/tracking-api .
COPY --from=builder /build/etc ./etc

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 启动应用
ENTRYPOINT ["./tracking-api", "-f", "etc/tracking-api-api.yaml"]