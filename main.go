package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("参数不足，需要: <distro> <version> <arch>")
	}

	distro := os.Args[1]
	version := os.Args[2]
	arch := os.Args[3]

	log.Printf("开始获取 %s %s %s 的镜像链接", distro, version, arch)

	baseURL := fmt.Sprintf("https://images.linuxcontainers.org/images/%s/%s/%s/default/", distro, version, arch)
	log.Printf("1. 获取到基础链接: %s", baseURL)

	// 获取最新目录
	latestDir, err := getLatestDirectory(baseURL)
	if err != nil {
		log.Fatalf("获取最新目录失败: %v", err)
	}

	log.Printf("2. 获取到最新目录: %s", latestDir)

	// 获取两个文件的链接
	rootfsURL := baseURL + latestDir + "rootfs.tar.xz"
	metaURL := baseURL + latestDir + "meta.tar.xz"

	log.Printf("3. 获取到rootfs文件: %s", rootfsURL)
	log.Printf("4. 获取到meta文件: %s", metaURL)

	// 检查文件是否存在
	if checkURLExists(rootfsURL) {
		log.Printf("5. rootfs文件验证成功")
		fmt.Printf("rootfs URL: %s", rootfsURL)
	} else {
		log.Fatalf("rootfs文件不存在: %s", rootfsURL)
	}

	if checkURLExists(metaURL) {
		log.Printf("6. meta文件验证成功")
		fmt.Printf("meta URL: %s", metaURL)
	} else {
		log.Fatalf("meta文件不存在: %s", metaURL)
	}

	// 下载文件
	log.Printf("7. 开始下载文件...")
	if err := downloadFile(rootfsURL, "rootfs.tar.xz"); err != nil {
		log.Fatalf("下载rootfs文件失败: %v", err)
	}
	log.Printf("8. rootfs文件下载完成")

	if err := downloadFile(metaURL, "meta.tar.xz"); err != nil {
		log.Fatalf("下载meta文件失败: %v", err)
	}
	log.Printf("9. meta文件下载完成")

	log.Printf("10. 所有文件下载完成")

	// 导入LXC镜像
	imageName := fmt.Sprintf("%s-%s-%s", distro, version, arch)
	log.Printf("11. 开始导入LXC镜像: %s", imageName)

	if err := importLXCImage("meta.tar.xz", "rootfs.tar.xz", imageName); err != nil {
		log.Fatalf("导入LXC镜像失败: %v", err)
	}
	log.Printf("12. LXC镜像导入完成")

	// 启动LXC容器
	containerName := fmt.Sprintf("%s-container", imageName)
	log.Printf("13. 启动LXC容器: %s", containerName)

	if err := launchLXCContainer(imageName, containerName); err != nil {
		log.Fatalf("启动LXC容器失败: %v", err)
	}
	log.Printf("14. LXC容器启动完成")

	log.Printf("15. 所有操作完成")
}

func checkURLExists(url string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getLatestDirectory(baseURL string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(baseURL)
	if err != nil {
		return "", fmt.Errorf("获取目录页面失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("目录页面不可用，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	content := string(body)

	// 匹配目录链接模式：20250924_08:35/
	pattern := `href="([0-9]{8}_[0-9]{2}%3A[0-9]{2}/)"`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("未找到任何构建目录")
	}

	var directories []string
	for _, match := range matches {
		if len(match) > 1 {
			// 解码URL编码的冒号
			dir := strings.Replace(match[1], "%3A", ":", -1)
			directories = append(directories, dir)
		}
	}

	// 按日期时间排序，选择最新的目录
	sort.Sort(sort.Reverse(sort.StringSlice(directories)))

	if len(directories) == 0 {
		return "", fmt.Errorf("未找到有效的构建目录")
	}

	return directories[0], nil
}

func downloadFile(url, filename string) error {
	log.Printf("下载 %s 到 %s", url, filename)

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 显示下载进度
	contentLength := resp.ContentLength
	var downloaded int64
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			written, err := file.Write(buffer[:n])
			if err != nil {
				return fmt.Errorf("写入文件失败: %v", err)
			}
			downloaded += int64(written)

			// 显示下载进度
			if contentLength > 0 {
				percent := float64(downloaded) / float64(contentLength) * 100
				log.Printf("下载进度: %.2f%% (%d/%d bytes)", percent, downloaded, contentLength)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取响应失败: %v", err)
		}
	}

	log.Printf("文件下载完成: %s (%d bytes)", filename, downloaded)
	return nil
}

func importLXCImage(metaFile, rootfsFile, imageName string) error {
	log.Printf("执行: lxc image import %s %s --alias %s", metaFile, rootfsFile, imageName)

	cmd := exec.Command("lxc", "image", "import", metaFile, rootfsFile, "--alias", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("镜像导入失败: %v", err)
	}

	log.Printf("镜像导入成功: %s", imageName)
	return nil
}

func launchLXCContainer(imageName, containerName string) error {
	log.Printf("执行: lxc launch %s %s", imageName, containerName)

	cmd := exec.Command("lxc", "launch", imageName, containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("容器启动失败: %v", err)
	}

	log.Printf("容器启动成功: %s", containerName)
	return nil
}
