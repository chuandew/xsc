.PHONY: build clean install run test

BINARY_NAME=xsc
BUILD_DIR=./build
INSTALL_DIR=/usr/local/bin

# 构建目标
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/xsc

# 清理构建文件
clean:
	rm -rf $(BUILD_DIR)

# 安装到系统
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/

# 卸载
uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)

# 运行
run:
	go run ./cmd/xsc

# 带参数运行
tui:
	go run ./cmd/xsc

list:
	go run ./cmd/xsc list

# 测试
test:
	go test -v ./...

# 格式化代码
fmt:
	go fmt ./...

# 检查代码
vet:
	go vet ./...

# 下载依赖
deps:
	go mod download
	go mod tidy

# 开发模式（自动重载）
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air
