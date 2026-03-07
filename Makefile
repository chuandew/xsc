.PHONY: build build-xftp build-all clean install run run-xftp test

BINARY_NAME=xsc
BINARY_XFTP=xftp
BUILD_DIR=./build
INSTALL_DIR=/usr/local/bin

# 构建 xsc
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/xsc

# 构建 xftp
build-xftp:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_XFTP) ./cmd/xftp

# 构建所有二进制
build-all: build build-xftp

# 清理构建文件
clean:
	rm -rf $(BUILD_DIR)

# 安装到系统
install: build-all
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	cp $(BUILD_DIR)/$(BINARY_XFTP) $(INSTALL_DIR)/

# 卸载
uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -f $(INSTALL_DIR)/$(BINARY_XFTP)

# 运行 xsc
run:
	go run ./cmd/xsc

# 运行 xftp
run-xftp:
	go run ./cmd/xftp

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
