# 第一阶段：构建可执行文件
FROM golang:1.23.6-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置 Go 镜像源
ENV GOPROXY=https://goproxy.cn,direct

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# 第二阶段：创建轻量级镜像
FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 从第一阶段复制可执行文件
COPY --from=builder /app/main .

# 暴露应用程序使用的端口（根据实际情况修改）
EXPOSE 8080

# 定义启动命令
CMD ["./main"]