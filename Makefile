# Cloud Storage 项目 Makefile
# 用于一键启动各个微服务

# 定义服务名称
SERVICES = user_service file_service share_service gateway
SERVICE_DIRS = services/user_service services/file_service services/share_service gateway

# 默认目标
.PHONY: all
all: help

# 帮助信息
.PHONY: help
help:
	@echo "Cloud Storage 项目 Makefile"
	@echo "使用方法:"
	@echo "  make all            - 显示帮助信息"
	@echo "  make start          - 启动所有基础设施和微服务"
	@echo "  make start-services - 启动所有微服务（不包括基础设施）"
	@echo "  make start-infra    - 启动基础设施(MinIO, MySQL, ETCD)"
	@echo "  make start-user     - 启动用户服务"
	@echo "  make start-file     - 启动文件服务"
	@echo "  make start-share    - 启动分享服务"
	@echo "  make start-gateway  - 启动网关服务"
	@echo "  make stop           - 停止所有服务"
	@echo "  make stop-services  - 停止所有微服务（不包括基础设施）"
	@echo "  make stop-infra     - 停止基础设施"
	@echo "  make build          - 构建所有服务"
	@echo "  make clean          - 清理生成的二进制文件"
	@echo "  make logs           - 查看基础设施日志"
	@echo "  make logs-user      - 查看用户服务日志"
	@echo "  make logs-file      - 查看文件服务日志"
	@echo "  make logs-share     - 查看分享服务日志"
	@echo "  make logs-gateway   - 查看网关服务日志"

# 启动所有服务
.PHONY: start
start: start-infra
	@echo "等待基础设施启动完成..."
	sleep 10
	@make start-services

# 启动所有微服务（不包括基础设施）
.PHONY: start-services
start-services: start-user start-file start-share start-gateway

# 启动基础设施
.PHONY: start-infra
start-infra:
	@echo "正在启动基础设施(MinIO, MySQL, ETCD)..."
	docker-compose up -d
	@echo "基础设施启动命令已执行，请使用 'make logs' 查看状态"

# 启动用户服务
.PHONY: start-user
start-user:
	@echo "正在启动用户服务..."
	cd services/user_service && nohup go run main.go > user_service.log 2>&1 &
	@echo "用户服务已在后台启动"

# 启动文件服务
.PHONY: start-file
start-file:
	@echo "正在启动文件服务..."
	cd services/file_service && nohup go run main.go > file_service.log 2>&1 &
	@echo "文件服务已在后台启动"

# 启动分享服务
.PHONY: start-share
start-share:
	@echo "正在启动分享服务..."
	cd services/share_service && nohup go run main.go > share_service.log 2>&1 &
	@echo "分享服务已在后台启动"

# 启动网关服务
.PHONY: start-gateway
start-gateway:
	@echo "正在启动网关服务..."
	cd gateway && nohup go run main.go > gateway.log 2>&1 &
	@echo "网关服务已在后台启动"

# 构建所有服务
.PHONY: build
build:
	@echo "正在构建所有服务..."
	@for dir in $(SERVICE_DIRS); do \
		echo "构建 $$dir..."; \
		cd $$dir && go build -o bin/service .; \
	done
	@echo "所有服务构建完成"

# 清理生成的二进制文件
.PHONY: clean
clean:
	@echo "正在清理生成的二进制文件..."
	@for dir in $(SERVICE_DIRS); do \
		echo "清理 $$dir..."; \
		cd $$dir && rm -rf bin/; \
	done
	@echo "清理完成"

# 停止所有服务
.PHONY: stop
stop: stop-services stop-infra

# 停止所有微服务（不包括基础设施）
.PHONY: stop-services
stop-services:
	@echo "正在停止所有微服务..."
	@pkill -f "go run main.go" || true
	@pkill -f "user_service" || true
	@pkill -f "file_service" || true
	@pkill -f "share_service" || true
	@pkill -f "gateway" || true
	@pkill -f "cloud-storage" || true
	@sleep 3
	@echo "检查是否还有残留进程..."
	@pkill -9 -f "go run main.go" 2>/dev/null || true
	@pkill -9 -f "user_service" 2>/dev/null || true
	@pkill -9 -f "file_service" 2>/dev/null || true
	@pkill -9 -f "share_service" 2>/dev/null || true
	@pkill -9 -f "gateway" 2>/dev/null || true
	@pkill -9 -f "cloud-storage" 2>/dev/null || true
	@sleep 2
	@echo "查找并终止可能遗漏的端口占用进程..."
	@lsof -i :8080 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@lsof -i :35001 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@lsof -i :35002 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@lsof -i :35003 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@lsof -i :35004 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@lsof -i :35005 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@lsof -i :35006 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@lsof -i :35007 | grep LISTEN | awk '{print $$2}' | xargs -r kill -9 2>/dev/null || true
	@sleep 2
	@echo "微服务已停止"

# 停止基础设施
.PHONY: stop-infra
stop-infra:
	@echo "正在停止基础设施..."
	docker-compose down
	@echo "基础设施已停止"

# 查看基础设施日志
.PHONY: logs
logs:
	@echo "查看基础设施日志:"
	docker-compose logs -f

# 查看特定服务日志
.PHONY: logs-user
logs-user:
	@echo "查看用户服务日志:"
	@tail -f services/user_service/user_service.log

.PHONY: logs-file
logs-file:
	@echo "查看文件服务日志:"
	@tail -f services/file_service/file_service.log

.PHONY: logs-share
logs-share:
	@echo "查看分享服务日志:"
	@tail -f services/share_service/share_service.log

.PHONY: logs-gateway
logs-gateway:
	@echo "查看网关服务日志:"
	@tail -f gateway/gateway.log