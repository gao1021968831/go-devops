package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"github.com/pkg/sftp"
	"go-devops/internal/logger"
	"go-devops/internal/models"
)

// SSHClient SSH客户端结构
type SSHClient struct {
	client *ssh.Client
	host   *models.Host
}

// NewSSHClient 创建SSH客户端
func NewSSHClient(host *models.Host) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User:            host.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
		Timeout:         30 * time.Second,
	}

	// 根据认证类型配置认证方法
	switch host.AuthType {
	case "password":
		if host.Password == "" {
			return nil, fmt.Errorf("密码认证需要提供密码")
		}
		config.Auth = []ssh.AuthMethod{
			ssh.Password(host.Password),
		}
		logger.Infof("使用密码认证连接主机: %s@%s", host.Username, host.IP)

	case "key":
		if host.PrivateKey == "" {
			return nil, fmt.Errorf("密钥认证需要提供私钥")
		}
		
		var signer ssh.Signer
		var err error
		
		if host.Passphrase != "" {
			// 带密码短语的私钥
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(host.PrivateKey), []byte(host.Passphrase))
		} else {
			// 无密码短语的私钥
			signer, err = ssh.ParsePrivateKey([]byte(host.PrivateKey))
		}
		
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %v", err)
		}
		
		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
		logger.Infof("使用密钥认证连接主机: %s@%s", host.Username, host.IP)

	default:
		return nil, fmt.Errorf("不支持的认证类型: %s，支持的类型: password, key", host.AuthType)
	}

	// 建立SSH连接
	addr := fmt.Sprintf("%s:%d", host.IP, host.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		logger.Errorf("SSH连接失败: %s@%s:%d, 错误: %v", host.Username, host.IP, host.Port, err)
		return nil, fmt.Errorf("SSH连接失败: %v", err)
	}

	logger.Infof("SSH连接成功: %s@%s:%d", host.Username, host.IP, host.Port)
	return &SSHClient{
		client: client,
		host:   host,
	}, nil
}

// Close 关闭SSH连接
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ExecuteCommand 执行命令
func (c *SSHClient) ExecuteCommand(command string) (string, string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	logger.Infof("在主机 %s 上执行命令: %s", c.host.IP, command)

	// 执行命令并获取输出
	output, err := session.CombinedOutput(command)
	if err != nil {
		logger.Errorf("命令执行失败: %v", err)
		return "", string(output), err
	}

	logger.Infof("命令执行成功，输出长度: %d 字节", len(output))
	return string(output), "", nil
}

// TestConnection 测试SSH连接
func (c *SSHClient) TestConnection() error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建测试会话失败: %v", err)
	}
	defer session.Close()

	// 执行简单的测试命令
	_, err = session.CombinedOutput("echo 'connection test'")
	if err != nil {
		return fmt.Errorf("连接测试失败: %v", err)
	}

	return nil
}

// TestSSHConnection 测试SSH连接（静态方法）
func TestSSHConnection(host *models.Host) (*models.SSHTestResponse, error) {
	start := time.Now()
	
	client, err := NewSSHClient(host)
	if err != nil {
		return &models.SSHTestResponse{
			Success: false,
			Message: fmt.Sprintf("连接失败: %v", err),
		}, nil
	}
	defer client.Close()

	err = client.TestConnection()
	latency := time.Since(start)

	if err != nil {
		return &models.SSHTestResponse{
			Success: false,
			Message: fmt.Sprintf("测试失败: %v", err),
			Latency: latency.String(),
		}, nil
	}

	return &models.SSHTestResponse{
		Success: true,
		Message: "连接测试成功",
		Latency: latency.String(),
	}, nil
}

// ExecuteScript 执行脚本
func ExecuteScript(host *models.Host, script *models.Script) (string, string, error) {
	client, err := NewSSHClient(host)
	if err != nil {
		return "", "", fmt.Errorf("建立SSH连接失败: %v", err)
	}
	defer client.Close()

	logger.Infof("开始在主机 %s 上执行脚本: %s", host.IP, script.Name)

	var command string
	switch script.Type {
	case "shell", "bash":
		// 对于shell脚本，直接执行
		command = script.Content
	case "python2":
		// 对于Python2脚本，尝试多种执行方式绕过安全软件
		command = fmt.Sprintf("cat > /tmp/temp_py_script.py << 'EOF'\n%s\nEOF\n# 尝试多种Python执行方式\nif command -v python2 >/dev/null 2>&1; then\n    python2 /tmp/temp_py_script.py 2>/dev/null || /usr/bin/python2 /tmp/temp_py_script.py 2>/dev/null || python /tmp/temp_py_script.py\nelse\n    python /tmp/temp_py_script.py\nfi\nrm -f /tmp/temp_py_script.py", script.Content)
	case "python3":
		// 对于Python3脚本，尝试多种执行方式绕过安全软件
		command = fmt.Sprintf("cat > /tmp/temp_py_script.py << 'EOF'\n%s\nEOF\n# 尝试多种Python执行方式\nif command -v python3 >/dev/null 2>&1; then\n    python3 /tmp/temp_py_script.py 2>/dev/null || /usr/bin/python3 /tmp/temp_py_script.py 2>/dev/null || python /tmp/temp_py_script.py\nelse\n    python /tmp/temp_py_script.py\nfi\nrm -f /tmp/temp_py_script.py", script.Content)
	default:
		// 默认作为shell脚本处理
		command = script.Content
	}

	output, stderr, err := client.ExecuteCommand(command)
	
	if err != nil {
		logger.Errorf("脚本执行失败: %v, 输出: %s", err, stderr)
		return output, stderr, err
	}

	logger.Infof("脚本执行成功: %s", script.Name)
	return output, stderr, nil
}

// ExecuteScriptWithFiles 执行脚本并传递输入文件
func ExecuteScriptWithFiles(host *models.Host, script *models.Script, inputFiles []models.File) (string, string, error) {
	client, err := NewSSHClient(host)
	if err != nil {
		return "", "", fmt.Errorf("建立SSH连接失败: %v", err)
	}
	defer client.Close()

	logger.Infof("开始在主机 %s 上执行脚本: %s，输入文件数量: %d", host.IP, script.Name, len(inputFiles))

	// 上传输入文件到远程主机
	for i, file := range inputFiles {
		logger.Infof("准备上传第 %d/%d 个文件: %s (ID: %d, 路径: %s, 大小: %d)", 
			i+1, len(inputFiles), file.OriginalName, file.ID, file.Path, file.Size)
		
		// 检查本地文件是否存在
		if _, err := os.Stat(file.Path); os.IsNotExist(err) {
			logger.Errorf("本地文件不存在: %s", file.Path)
			return "", "", fmt.Errorf("本地文件不存在: %s", file.Path)
		}
		
		err := client.UploadFile(file.Path, file.OriginalName)
		if err != nil {
			logger.Errorf("上传文件失败: %s (路径: %s), 错误: %v", file.OriginalName, file.Path, err)
			return "", "", fmt.Errorf("上传文件失败: %s", file.OriginalName)
		}
		logger.Infof("文件上传成功: %s -> %s@%s:%s", file.Path, host.Username, host.IP, file.OriginalName)
	}

	var command string
	switch script.Type {
	case "shell", "bash":
		command = script.Content
	case "python2":
		command = fmt.Sprintf("cat > temp_py_script.py << 'EOF'\n%s\nEOF\nif command -v python2 >/dev/null 2>&1; then\n    python2 temp_py_script.py 2>/dev/null || /usr/bin/python2 temp_py_script.py 2>/dev/null || python temp_py_script.py\nelse\n    python temp_py_script.py\nfi\nrm -f temp_py_script.py", script.Content)
	case "python3":
		command = fmt.Sprintf("cat > temp_py_script.py << 'EOF'\n%s\nEOF\nif command -v python3 >/dev/null 2>&1; then\n    python3 temp_py_script.py 2>/dev/null || /usr/bin/python3 temp_py_script.py 2>/dev/null || python temp_py_script.py\nelse\n    python temp_py_script.py\nfi\nrm -f temp_py_script.py", script.Content)
	default:
		command = script.Content
	}

	output, stderr, err := client.ExecuteCommand(command)
	
	// 清理上传的文件和临时脚本文件
	for _, file := range inputFiles {
		client.ExecuteCommand(fmt.Sprintf("rm -f %s", file.OriginalName))
	}
	// 清理Python临时脚本文件（防止执行失败时未清理）
	if script.Type == "python2" || script.Type == "python3" {
		client.ExecuteCommand("rm -f temp_py_script.py")
	}
	
	if err != nil {
		logger.Errorf("脚本执行失败: %v, 输出: %s", err, stderr)
		return output, stderr, err
	}

	logger.Infof("脚本执行成功: %s", script.Name)
	return output, stderr, nil
}

// GetSystemInfo 获取系统信息
func GetSystemInfo(host *models.Host) (map[string]interface{}, error) {
	client, err := NewSSHClient(host)
	if err != nil {
		return nil, fmt.Errorf("建立SSH连接失败: %v", err)
	}
	defer client.Close()

	info := make(map[string]interface{})

	// 获取系统信息
	commands := map[string]string{
		"hostname": "hostname",
		"uptime":   "uptime",
		"uname":    "uname -a",
		"memory":   "free -h",
		"disk":     "df -h",
		"cpu":      "cat /proc/cpuinfo | grep 'model name' | head -1",
	}

	for key, cmd := range commands {
		output, _, err := client.ExecuteCommand(cmd)
		if err != nil {
			logger.Warnf("获取 %s 信息失败: %v", key, err)
			info[key] = fmt.Sprintf("获取失败: %v", err)
		} else {
			info[key] = output
		}
	}

	return info, nil
}

// UploadFile 上传文件到远程主机
func (c *SSHClient) UploadFile(localPath, remotePath string) error {
	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	// 打开本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer localFile.Close()

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	err = sftpClient.MkdirAll(remoteDir)
	if err != nil {
		logger.Warnf("创建远程目录失败: %v", err)
	}

	// 创建远程文件
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	// 复制文件内容
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("文件传输失败: %v", err)
	}

	logger.Infof("文件上传成功: %s -> %s@%s:%s", localPath, c.host.Username, c.host.IP, remotePath)
	return nil
}

// DownloadFile 从远程主机下载文件
func (c *SSHClient) DownloadFile(remotePath, localPath string) error {
	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	// 打开远程文件
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("打开远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	// 确保本地目录存在
	localDir := filepath.Dir(localPath)
	err = os.MkdirAll(localDir, 0755)
	if err != nil {
		return fmt.Errorf("创建本地目录失败: %v", err)
	}

	// 创建本地文件
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %v", err)
	}
	defer localFile.Close()

	// 复制文件内容
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("文件传输失败: %v", err)
	}

	logger.Infof("文件下载成功: %s@%s:%s -> %s", c.host.Username, c.host.IP, remotePath, localPath)
	return nil
}

// FileExists 检查远程文件是否存在
func (c *SSHClient) FileExists(remotePath string) (bool, error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return false, fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	_, err = sftpClient.Stat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetFileInfo 获取远程文件信息
func (c *SSHClient) GetFileInfo(remotePath string) (os.FileInfo, error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return nil, fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	return sftpClient.Stat(remotePath)
}

// RemoveFile 删除远程文件
func (c *SSHClient) RemoveFile(remotePath string) error {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	err = sftpClient.Remove(remotePath)
	if err != nil {
		return fmt.Errorf("删除远程文件失败: %v", err)
	}

	logger.Infof("删除远程文件成功: %s@%s:%s", c.host.Username, c.host.IP, remotePath)
	return nil
}
