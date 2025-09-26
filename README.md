# LXC容器构建器

## 功能说明
自动获取Linux容器镜像并构建LXC容器。

## 使用方法

```bash
# 基本用法
go run main.go <发行版> <版本> <架构>

# 示例
go run main.go centos 9-Stream arm64
go run main.go centos 10-Stream amd64
```

## 支持的参数

- **发行版**: centos, ubuntu, debian 等
- **版本**: 9-Stream, 10-Stream, 22.04 等  
- **架构**: amd64, arm64, armhf 等

## 输出
程序会输出最新的镜像下载链接，格式为：
```
镜像URL: https://images.linuxcontainers.org/images/发行版/版本/架构/default/日期_时间/rootfs.tar.xz