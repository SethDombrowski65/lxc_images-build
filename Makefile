.PHONY: build test clean install

# 构建目标
build:
	go build -o lxc-builder

# 测试
test:
	go test ./...

# 清理构建文件
clean:
	rm -f lxc-builder

# 安装依赖
deps:
	go mod tidy

# 开发模式（构建并运行测试）
dev: deps build test

# 创建默认配置
config:
	@echo "创建默认配置文件..."
	@if [ ! -f config.yaml ]; then \
		echo "# LXC构建器配置文件" > config.yaml; \
		echo "" >> config.yaml; \
		echo "lxc:" >> config.yaml; \
		echo "  storage_path: \"/var/lib/lxc\"" >> config.yaml; \
		echo "  network_type: \"veth\"" >> config.yaml; \
		echo "  bridge: \"lxcbr0\"" >> config.yaml; \
		echo "" >> config.yaml; \
		echo "ssh:" >> config.yaml; \
		echo "  port: 22" >> config.yaml; \
		echo "  user: \"root\"" >> config.yaml; \
		echo "  password: \"change_this_password\"" >> config.yaml; \
		echo "创建完成，请编辑 config.yaml 文件配置SSH密钥"; \
	else \
		echo "配置文件已存在"; \
	fi

# 显示帮助信息
help:
	@echo "可用命令:"
	@echo "  build    - 构建程序"
	@echo "  test     - 运行测试"
	@echo "  clean    - 清理构建文件"
	@echo "  deps     - 安装依赖"
	@echo "  dev      - 开发模式（构建+测试）"
	@echo "  config   - 创建默认配置文件"
	@echo "  help     - 显示此帮助信息"