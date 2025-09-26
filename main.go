package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	action := os.Args[1]
	
	switch action {
	case "build":
		if len(os.Args) < 5 {
			log.Fatal("用法: lxc-builder build <发行版> <版本> <架构>")
		}
		distro := os.Args[2]
		version := os.Args[3]
		arch := os.Args[4]
		buildContainer(distro, version, arch)
	case "publish":
		if len(os.Args) < 3 {
			log.Fatal("用法: lxc-builder publish <容器名>")
		}
		containerName := os.Args[2]
		publishContainer(containerName)
	default:
		log.Fatalf("未知操作: %s", action)
	}
}

func printUsage() {
	fmt.Println("LXC容器构建器 - GitHub Actions专用")
	fmt.Println("用法:")
	fmt.Println("  lxc-builder build <发行版> <版本> <架构>    # 构建容器")
	fmt.Println("  lxc-builder publish <容器名>                # 发布容器")
	fmt.Println()
	fmt.Println("支持的镜像:")
	fmt.Println("  - centos 10-Stream amd64")
	fmt.Println("  - centos 10-Stream arm64")
	fmt.Println("  - centos 9-Stream amd64")
	fmt.Println("  - centos 9-Stream arm64")
}

func buildContainer(distro, version, arch string) {
	log.Printf("开始构建 %s %s %s 容器", distro, version, arch)
	
	// 1. 检测并下载镜像
	imageURL := detectImage(distro, version, arch)
	if imageURL == "" {
		log.Fatalf("未找到 %s %s %s 的可用镜像", distro, version, arch)
	}
	
	// 2. 创建容器
	containerName := generateContainerName(distro, version, arch)
	createContainer(containerName, imageURL)
	
	// 3. 配置SSH
	setupSSH(containerName)
	
	// 4. 保存容器信息
	saveContainerInfo(containerName, distro, version, arch)
	
	log.Printf("容器 %s 构建完成", containerName)
}

func detectImage(distro, version, arch string) string {
	// 支持的镜像URL列表
	supportedImages := map[string]string{
		"centos/10-Stream/amd64": "https://images.linuxcontainers.org/images/centos/10-Stream/amd64/default/latest/rootfs.tar.xz",
		"centos/10-Stream/arm64": "https://images.linuxcontainers.org/images/centos/10-Stream/arm64/default/latest/rootfs.tar.xz",
		"centos/9-Stream/amd64":  "https://images.linuxcontainers.org/images/centos/9-Stream/amd64/default/latest/rootfs.tar.xz",
		"centos/9-Stream/arm64":  "https://images.linuxcontainers.org/images/centos/9-Stream/arm64/default/latest/rootfs.tar.xz",
	}
	
	key := fmt.Sprintf("%s/%s/%s", distro, version, arch)
	imageURL, exists := supportedImages[key]
	if !exists {
		log.Printf("不支持的镜像配置: %s", key)
		return ""
	}
	
	// 检查镜像URL是否存在
	if checkURLExists(imageURL) {
		return imageURL
	}
	
	// 如果rootfs.tar.xz不存在，尝试rootfs.squashfs
	altURL := strings.Replace(imageURL, "rootfs.tar.xz", "rootfs.squashfs", 1)
	if checkURLExists(altURL) {
		return altURL
	}
	
	return ""
}

func checkURLExists(url string) bool {
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", url)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.TrimSpace(string(output)) == "200"
}

func createContainer(containerName, imageURL string) {
	log.Printf("创建容器 %s，使用镜像: %s", containerName, imageURL)
	
	// 下载镜像文件
	imageFile := "/tmp/lxc-image.tar.xz"
	downloadImage(imageURL, imageFile)
	
	// 创建容器配置
	createContainerConfig(containerName)
	
	// 导入镜像
	importImage(containerName, imageFile)
	
	// 启动容器
	startContainer(containerName)
}

func downloadImage(url, outputPath string) {
	log.Printf("下载镜像: %s", url)
	cmd := exec.Command("curl", "-L", "-o", outputPath, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("下载镜像失败: %v", err)
	}
}

func createContainerConfig(containerName string) {
	configContent := fmt.Sprintf(`
# 容器配置
lxc.include = /usr/share/lxc/config/common.conf
lxc.arch = x86_64

# 网络配置
lxc.net.0.type = veth
lxc.net.0.link = lxcbr0
lxc.net.0.flags = up
lxc.net.0.hwaddr = 00:16:3e:xx:xx:xx

# 根文件系统
lxc.rootfs.path = dir:/var/lib/lxc/%s/rootfs
lxc.rootfs.mount = /var/lib/lxc/%s/rootfs
`, containerName, containerName)
	
	configPath := fmt.Sprintf("/var/lib/lxc/%s/config", containerName)
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		log.Fatalf("创建容器配置失败: %v", err)
	}
}

func importImage(containerName, imageFile string) {
	// 创建根文件系统目录
	rootfsPath := fmt.Sprintf("/var/lib/lxc/%s/rootfs", containerName)
	os.MkdirAll(rootfsPath, 0755)
	
	// 解压镜像
	cmd := exec.Command("tar", "-xf", imageFile, "-C", rootfsPath)
	if err := cmd.Run(); err != nil {
		log.Fatalf("解压镜像失败: %v", err)
	}
}

func startContainer(containerName string) {
	cmd := exec.Command("lxc-start", "-n", containerName, "-d")
	if err := cmd.Run(); err != nil {
		log.Fatalf("启动容器失败: %v", err)
	}
	
	log.Printf("容器 %s 已启动", containerName)
}

func setupSSH(containerName string) {
	log.Printf("为容器 %s 配置SSH", containerName)
	
	// 设置root密码
	setRootPassword(containerName, "github-actions-123")
	
	// 安装SSH服务器
	installSSHServer(containerName)
	
	// 配置SSH服务
	configureSSH(containerName)
	
	// 重启SSH服务
	restartSSH(containerName)
	
	log.Printf("容器 %s SSH配置完成", containerName)
}

func setRootPassword(containerName, password string) {
	cmd := exec.Command("lxc-attach", "-n", containerName, "--", 
		"bash", "-c", fmt.Sprintf("echo 'root:%s' | chpasswd", password))
	if err := cmd.Run(); err != nil {
		log.Fatalf("设置root密码失败: %v", err)
	}
}

func installSSHServer(containerName string) {
	// 尝试不同的包管理器
	commands := []string{
		"apt-get update && apt-get install -y openssh-server",
		"yum install -y openssh-server",
		"apk add openssh-server",
	}
	
	for _, cmdStr := range commands {
		cmd := exec.Command("lxc-attach", "-n", containerName, "--", "bash", "-c", cmdStr)
		if cmd.Run() == nil {
			return
		}
	}
	
	log.Fatal("无法安装SSH服务器")
}

func configureSSH(containerName string) {
	sshdConfig := `
Port 22
PermitRootLogin yes
PasswordAuthentication yes
PubkeyAuthentication yes
X11Forwarding yes
PrintMotd no
AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/openssh/sftp-server
`
	
	cmd := exec.Command("lxc-attach", "-n", containerName, "--", 
		"bash", "-c", fmt.Sprintf("echo '%s' > /etc/ssh/sshd_config", sshdConfig))
	if err := cmd.Run(); err != nil {
		log.Fatalf("配置SSH失败: %v", err)
	}
}

func restartSSH(containerName string) {
	commands := []string{
		"systemctl restart ssh",
		"systemctl restart sshd", 
		"service ssh restart",
		"service sshd restart",
	}
	
	for _, cmdStr := range commands {
		cmd := exec.Command("lxc-attach", "-n", containerName, "--", "bash", "-c", cmdStr)
		if cmd.Run() == nil {
			return
		}
	}
	
	log.Fatal("无法重启SSH服务")
}

func publishContainer(containerName string) {
	log.Printf("发布容器 %s", containerName)
	
	// 停止容器
	exec.Command("lxc-stop", "-n", containerName).Run()
	
	// 打包容器
	packageContainer(containerName)
	
	// 上传到GitHub Releases或其他存储
	uploadContainer(containerName)
	
	log.Printf("容器 %s 发布完成", containerName)
}

func packageContainer(containerName string) {
	// 打包容器为tar文件
	tarFile := fmt.Sprintf("%s.tar.gz", containerName)
	cmd := exec.Command("tar", "-czf", tarFile, "-C", "/var/lib/lxc", containerName)
	if err := cmd.Run(); err != nil {
		log.Fatalf("打包容器失败: %v", err)
	}
}

func uploadContainer(containerName string) {
	// 这里可以实现上传到GitHub Releases或其他存储的逻辑
	// 在GitHub Actions中可以使用gh命令上传
	log.Printf("容器 %s 已打包，准备上传", containerName)
}

func generateContainerName(distro, version, arch string) string {
	// 生成唯一的容器名称
	return fmt.Sprintf("%s-%s-%s-%d", distro, version, arch, os.Getpid())
}

func saveContainerInfo(containerName, distro, version, arch string) {
	info := fmt.Sprintf("容器名称: %s\n发行版: %s\n版本: %s\n架构: %s\nSSH信息:\n  地址: 容器IP\n  端口: 22\n  用户: root\n  密码: github-actions-123\n", 
		containerName, distro, version, arch)
	
	err := os.WriteFile("container-info.txt", []byte(info), 0644)
	if err != nil {
		log.Printf("保存容器信息失败: %v", err)
	}
}