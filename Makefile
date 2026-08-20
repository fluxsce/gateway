.PHONY: run build test clean debug

# 运行程序
run:
	go run cmd/app/main.go --config ./configs

# 构建程序
build:
	go build -o bin/gateway cmd/app/main.go

# 运行测试
test:
	go test ./...

# 清理构建文件
clean:
	rm -rf bin/

# 调试程序
debug:
	dlv debug cmd/app/main.go -- --config ./configs

# 安装依赖
deps:
	go mod tidy

# 生成 API 文档
docs:
	swag init -g cmd/app/main.go 