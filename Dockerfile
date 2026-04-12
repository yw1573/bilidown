# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置 Alpine 镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置 Go 模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 安装构建依赖
RUN apk add --no-cache git nodejs npm ffmpeg

# 安装 pnpm
RUN npm install -g pnpm

WORKDIR /app

# 复制前端代码
COPY ui ./ui
WORKDIR /app/ui
RUN pnpm install && pnpm build

# 复制后端代码
COPY server ./server
WORKDIR /app/server

# 下载依赖并构建
RUN go mod download && go build -ldflags="-s -w" -o /bilidown ./cmd/bilidown

# 运行阶段
FROM alpine:3.19

# 设置 Alpine 镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖
RUN apk add --no-cache ffmpeg ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /bilidown .

# 创建数据目录
RUN mkdir -p /data

# 设置环境变量
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8098

# 运行
CMD ["./bilidown"]
