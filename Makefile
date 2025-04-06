.PHONY: run build test clean debug

# 运行程序
run:
	go run cmd/gateway/main.go

# 构建程序
build:
	go build -o bin/gateway cmd/gateway/main.go

# 运行测试
test:
	go test ./...

# 清理构建文件
clean:
	rm -rf bin/

# 调试程序
debug:
	dlv debug cmd/gateway/main.go

# 安装依赖
deps:
	go mod tidy

# 生成 API 文档
docs:
	swag init -g cmd/gateway/main.go 