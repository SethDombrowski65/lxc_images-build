# LXC容器构建器 - GitHub Actions专用

专门为GitHub Actions设计的LXC容器自动构建工具，能够自动检测镜像、下载文件、创建容器、安装SSH服务，并发布构建的SSH镜像。

## 功能特性

- 🔍 **自动镜像检测** - 自动从Linux Containers镜像仓库检测可用镜像
- 🐳 **容器创建** - 自动创建和配置LXC容器
- 🔑 **SSH自动配置** - 自动安装和配置SSH服务器（端口22，root用户）
- 🚀 **GitHub Actions集成** - 专门为CI/CD流水线设计
- 📦 **镜像发布** - 支持将构建的容器打包发布

## 快速开始

### 1. 在GitHub仓库中启用Actions

将本仓库的代码复制到您的GitHub仓库中，确保包含以下文件：
- `main.go` - 主程序
- `go.mod` - Go模块定义
- `.github/workflows/build-lxc.yml` - GitHub Actions工作流

### 2. 手动触发构建

在GitHub仓库的Actions页面：
1. 选择"Build LXC Containers"工作流
2. 点击"Run workflow"
3. 填写参数：
   - **发行版**: centos
   - **版本**: 10-Stream, 9-Stream
   - **架构**: amd64, arm64
   - **是否发布镜像**: 可选，勾选后会将容器打包上传

### 3. 查看构建结果

构建完成后，您将获得：
- 一个配置好SSH的LXC容器
- SSH连接信息（IP地址、端口22、root用户、密码）
- 可选的容器打包文件

## SSH连接信息

构建完成后，容器的SSH配置如下：
- **主机**: 容器IP地址（在Actions日志中显示）
- **端口**: 22
- **用户**: root
- **密码**: github-actions-123

## 支持的镜像

### CentOS 10-Stream
- amd64架构: https://images.linuxcontainers.org/images/centos/10-Stream/amd64/default/
- arm64架构: https://images.linuxcontainers.org/images/centos/10-Stream/arm64/default/

### CentOS 9-Stream
- amd64架构: https://images.linuxcontainers.org/images/centos/9-Stream/amd64/default/
- arm64架构: https://images.linuxcontainers.org/images/centos/9-Stream/arm64/default/

## 项目结构

```
lxc-builder/
├── main.go                 # 主程序 - 容器构建逻辑
├── go.mod                  # Go模块定义
└── .github/workflows/
    └── build-lxc.yml       # GitHub Actions工作流
```

## 工作原理

1. **镜像检测**: 自动从images.linuxcontainers.org检测可用镜像
2. **容器创建**: 下载镜像并创建LXC容器
3. **SSH配置**: 安装openssh-server，配置root密码和SSH服务
4. **容器启动**: 启动容器并等待SSH服务就绪
5. **镜像发布**: 可选地将容器打包为tar.gz文件

## 使用示例

### 在GitHub Actions中构建CentOS 10 Stream amd64容器

工作流配置：
```yaml
- name: Build CentOS 10 Stream amd64 container
  run: ./lxc-builder build centos 10-Stream amd64
```

### 构建CentOS 9 Stream arm64容器并发布

工作流配置：
```yaml
- name: Build and publish CentOS 9 Stream arm64 container
  run: |
    ./lxc-builder build centos 9-Stream arm64
    ./lxc-builder publish <容器名>
```

## 注意事项

- 需要GitHub Actions运行在Ubuntu环境中
- 需要LXC和bridge-utils依赖
- SSH密码固定为`github-actions-123`（可在代码中修改）
- 构建的容器仅在Actions运行期间存在，除非发布

## 许可证

MIT License