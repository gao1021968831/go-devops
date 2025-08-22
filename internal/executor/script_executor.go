package executor

import (
	"fmt"

	"go-devops/internal/models"
	"go-devops/internal/ssh"
)

// ScriptExecutor 脚本执行器
type ScriptExecutor struct{}

// NewScriptExecutor 创建脚本执行器实例
func NewScriptExecutor() *ScriptExecutor {
	return &ScriptExecutor{}
}

// ExecuteScript 执行脚本的统一入口
func (e *ScriptExecutor) ExecuteScript(host *models.Host, script *models.Script) (string, string, error) {
	switch script.Type {
	case "shell":
		return e.executeShellScript(host, script)
	case "python2":
		return e.executePython2Script(host, script)
	case "python3":
		return e.executePython3Script(host, script)
	default:
		return "", "", fmt.Errorf("不支持的脚本类型: %s", script.Type)
	}
}

// Shell脚本执行 - 优化后的实现
func (e *ScriptExecutor) executeShellScript(host *models.Host, script *models.Script) (string, string, error) {
	// 直接执行原始脚本内容，不需要额外包装
	return ssh.ExecuteScript(host, script)
}

// Python2脚本执行
func (e *ScriptExecutor) executePython2Script(host *models.Host, script *models.Script) (string, string, error) {
	// 直接传递给SSH模块处理，避免双重包装
	return ssh.ExecuteScript(host, script)
}

// Python3脚本执行
func (e *ScriptExecutor) executePython3Script(host *models.Host, script *models.Script) (string, string, error) {
	// 直接传递给SSH模块处理，避免双重包装
	return ssh.ExecuteScript(host, script)
}

