# 多阶段构建 Dockerfile for DevOps管理平台

# 第一阶段：构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# 复制前端依赖文件
COPY web/package*.json ./

# 安装前端依赖
RUN npm ci --only=production

# 复制前端源码
COPY web/ ./

# 构建前端
RUN npm run build

# 第二阶段：构建后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY . .

# 构建后端应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o devops-server main.go

# 第三阶段：运行时镜像
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata sqlite

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 创建必要的目录
RUN mkdir -p /app/logs /app/uploads /app/data /app/config

# 从构建阶段复制文件
COPY --from=backend-builder /app/devops-server .
COPY --from=frontend-builder /app/web/dist ./web/dist
COPY --from=backend-builder /app/templates ./templates
COPY --from=backend-builder /app/config ./config

# 复制配置文件
COPY config/config.yaml ./config/

# 设置权限
RUN chmod +x devops-server

# 创建非root用户
RUN addgroup -g 1001 -S devops && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G devops -g devops devops

# 修改目录权限
RUN chown -R devops:devops /app

# 切换到非root用户
USER devops

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 设置环境变量
ENV ENVIRONMENT=production
ENV LOG_TO_FILE=true
ENV DATABASE_URL=/app/data/devops.db

# 启动应用
CMD ["./devops-server"]
