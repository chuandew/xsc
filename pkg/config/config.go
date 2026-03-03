package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GlobalConfig 全局配置
type GlobalConfig struct {
	SecureCRT SecureCRTConfig `yaml:"securecrt,omitempty"`
	XShell    XShellConfig    `yaml:"xshell,omitempty"`
	SSH       SSHConfig       `yaml:"ssh,omitempty"`
}

// SSHConfig SSH配置
type SSHConfig struct {
	KnownHostsFile string `yaml:"known_hosts_file,omitempty"`
	StrictHostKey  bool   `yaml:"strict_host_key,omitempty"`
}

// SecureCRTConfig SecureCRT配置
type SecureCRTConfig struct {
	Enabled     bool   `yaml:"enabled"`
	SessionPath string `yaml:"session_path"`
	Password    string `yaml:"password"`
}

// XShellConfig XShell配置
type XShellConfig struct {
	Enabled     bool   `yaml:"enabled"`
	SessionPath string `yaml:"session_path"`
	Password    string `yaml:"password"`
}

var globalConfig *GlobalConfig

// LoadGlobalConfig 加载全局配置
func LoadGlobalConfig() (*GlobalConfig, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// 默认配置
	globalConfig = &GlobalConfig{
		SecureCRT: SecureCRTConfig{
			Enabled:     false,
			SessionPath: filepath.Join(configDir, "securecrt_sessions"),
			Password:    "",
		},
		XShell: XShellConfig{
			Enabled:     false,
			SessionPath: filepath.Join(configDir, "xshell_sessions"),
			Password:    "",
		},
		SSH: SSHConfig{
			KnownHostsFile: "",
			StrictHostKey:  false,
		},
	}

	// 如果配置文件存在，加载它
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, globalConfig); err != nil {
			return nil, err
		}
	}

	return globalConfig, nil
}

// SaveGlobalConfig 保存全局配置
func SaveGlobalConfig(config *GlobalConfig) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return err
	}

	globalConfig = config
	return nil
}

// GetSessionsDir 返回会话目录路径
func GetSessionsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	sessionsDir := filepath.Join(homeDir, ".xsc", "sessions")

	// 确保目录存在
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return "", err
	}

	return sessionsDir, nil
}

// GetConfigDir 返回配置目录路径
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".xsc")

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// GetKnownHostsPath 返回 known_hosts 文件路径
// 优先级：配置中的路径 > ~/.ssh/known_hosts > ~/.xsc/known_hosts
func GetKnownHostsPath() (string, error) {
	// 首先检查配置
	cfg, err := LoadGlobalConfig()
	if err == nil && cfg.SSH.KnownHostsFile != "" {
		return cfg.SSH.KnownHostsFile, nil
	}

	// 检查默认的 ~/.ssh/known_hosts
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	sshKnownHosts := filepath.Join(homeDir, ".ssh", "known_hosts")
	if _, err := os.Stat(sshKnownHosts); err == nil {
		return sshKnownHosts, nil
	}

	// 使用 xsc 的 known_hosts
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "known_hosts"), nil
}
