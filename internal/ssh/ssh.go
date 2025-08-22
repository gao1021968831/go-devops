package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
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
