package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Environment string `yaml:"environment"`
		Port        int    `yaml:"port"`
	} `yaml:"app"`

	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`

	JWT struct {
		Secret      string `yaml:"secret"`
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`

	Logging struct {
		Level    string `yaml:"level"`
		ToFile   bool   `yaml:"to_file"`
		FilePath string `yaml:"file_path"`
	} `yaml:"logging"`

	Scheduler struct {
		Enabled           bool   `yaml:"enabled"`
		HostCheckInterval string `yaml:"host_check_interval"`
	} `yaml:"scheduler"`
}

func Load() (*Config, error) {
	// 获取配置文件路径
	configPath := getEnv("CONFIG_PATH", "config/config.yaml")
	
	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("获取工作目录失败: %v", err)
		}
		configPath = filepath.Join(wd, configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 环境变量覆盖（如果存在）
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.App.Environment = env
	}
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.App.Port = p
		}
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}
	if logToFile := os.Getenv("LOG_TO_FILE"); logToFile != "" {
		if ltf, err := strconv.ParseBool(logToFile); err == nil {
			config.Logging.ToFile = ltf
		}
	}

	return &config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
