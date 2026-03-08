package ssh

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ketor/xsc/internal/session"
)

// TestDialInvalidSession 测试无效 session 返回错误
func TestDialInvalidSession(t *testing.T) {
	s := &session.Session{
		Valid: false,
		Error: fmt.Errorf("test error"),
	}

	client, cleanup, err := Dial(s)
	if err == nil {
		t.Fatal("期望返回错误，但得到 nil")
	}
	if client != nil {
		t.Fatal("期望 client 为 nil")
	}
	if cleanup != nil {
		t.Fatal("期望 cleanup 为 nil")
	}
	if !strings.Contains(err.Error(), "invalid session") {
		t.Errorf("错误消息应包含 'invalid session'，实际: %s", err.Error())
	}
}

// TestDialDefaultPort 测试默认端口为 22
func TestDialDefaultPort(t *testing.T) {
	s := &session.Session{
		Host:     "192.168.1.1",
		Port:     0,
		User:     "testuser",
		AuthType: session.AuthTypePassword,
		Password: "testpass",
		Valid:    true,
	}

	// 先验证 Validate 会将 Port 设为 22
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate 失败: %v", err)
	}
	if s.Port != 22 {
		t.Errorf("期望端口为 22，实际: %d", s.Port)
	}
}

// TestDialPasswordAuthConfig 测试密码认证的 SSH 配置构建正确
func TestDialPasswordAuthConfig(t *testing.T) {
	s := &session.Session{
		Host:     "192.168.1.1",
		Port:     22,
		User:     "testuser",
		AuthType: session.AuthTypePassword,
		Password: "testpass",
		Valid:    true,
	}

	// 使用 getSSHConfig 来验证配置正确构建
	config, cleanup, err := getSSHConfig(s)
	if err != nil {
		t.Fatalf("getSSHConfig 失败: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	if config.User != "testuser" {
		t.Errorf("期望用户名为 'testuser'，实际: %s", config.User)
	}
	if len(config.Auth) != 1 {
		t.Errorf("期望 1 个认证方法，实际: %d", len(config.Auth))
	}
}

// TestDialKeyAuthConfig 测试密钥认证的 SSH 配置构建（无效密钥路径返回错误）
func TestDialKeyAuthConfig(t *testing.T) {
	s := &session.Session{
		Host:     "192.168.1.1",
		Port:     22,
		User:     "testuser",
		AuthType: session.AuthTypeKey,
		KeyPath:  "/nonexistent/key",
		Valid:    true,
	}

	_, _, err := getSSHConfig(s)
	if err == nil {
		t.Fatal("期望因无效密钥路径返回错误")
	}
	if !strings.Contains(err.Error(), "failed to read key file") {
		t.Errorf("错误消息应包含 'failed to read key file'，实际: %s", err.Error())
	}
}

// TestDialUnsupportedAuthType 测试不支持的认证类型返回错误
func TestDialUnsupportedAuthType(t *testing.T) {
	s := &session.Session{
		Host:     "192.168.1.1",
		Port:     22,
		User:     "testuser",
		AuthType: session.AuthType("unknown"),
		Valid:    true,
	}

	client, cleanup, err := Dial(s)
	if err == nil {
		t.Fatal("期望返回错误")
	}
	if client != nil {
		t.Fatal("期望 client 为 nil")
	}
	if cleanup != nil {
		t.Fatal("期望 cleanup 为 nil")
	}
}

// TestDialMultipleAuthConfig 测试多认证方式的配置构建
func TestDialMultipleAuthConfig(t *testing.T) {
	s := &session.Session{
		Host:     "192.168.1.1",
		Port:     22,
		User:     "testuser",
		AuthType: session.AuthTypePassword,
		Valid:    true,
		AuthMethods: []session.AuthMethod{
			{
				Type:     "password",
				Password: "testpass",
			},
		},
	}

	// 验证多认证路径的配置构建
	config, cleanup, err := getSSHConfigForAuthMethod(s, s.AuthMethods[0])
	if err != nil {
		t.Fatalf("getSSHConfigForAuthMethod 失败: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	if config.User != "testuser" {
		t.Errorf("期望用户名为 'testuser'，实际: %s", config.User)
	}
	if len(config.Auth) != 1 {
		t.Errorf("期望 1 个认证方法，实际: %d", len(config.Auth))
	}
}

// TestGetHostKeyCallback 测试主机密钥回调创建
func TestGetHostKeyCallback(t *testing.T) {
	// getHostKeyCallback 应该不 panic
	callback := getHostKeyCallback()
	if callback == nil {
		t.Fatal("getHostKeyCallback 不应返回 nil")
	}
}

// TestAppendHostKey 测试主机密钥追加（最佳努力）
func TestAppendHostKey(t *testing.T) {
	// 使用临时文件测试
	tmpFile, err := os.CreateTemp("", "known_hosts_test")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// appendHostKey 不应 panic（即使参数无效）
	appendHostKey("/nonexistent/path", nil, nil)

	// 使用有效路径但 nil key 也不应 panic
	// 注意：实际写入会失败因为 key 为 nil，但不应 panic
	// 这只是验证函数的鲁棒性
}

// TestFindDefaultSSHKeys 测试默认 SSH 密钥查找
func TestFindDefaultSSHKeys(t *testing.T) {
	// 函数不应 panic
	keys := findDefaultSSHKeys()
	// 结果取决于环境，但不应 panic
	_ = keys
}

// TestDialKeyboardInteractiveConfig 测试键盘交互认证配置
func TestDialKeyboardInteractiveConfig(t *testing.T) {
	s := &session.Session{
		Host:  "192.168.1.1",
		Port:  22,
		User:  "testuser",
		Valid: true,
	}

	authMethod := session.AuthMethod{
		Type:     "keyboard-interactive",
		Password: "testpass",
	}

	config, cleanup, err := getSSHConfigForAuthMethod(s, authMethod)
	if err != nil {
		t.Fatalf("getSSHConfigForAuthMethod 失败: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	if len(config.Auth) != 1 {
		t.Errorf("期望 1 个认证方法，实际: %d", len(config.Auth))
	}
}

// TestConnectInvalidSession 测试无效会话连接
func TestConnectInvalidSession(t *testing.T) {
	s := &session.Session{
		Valid: false,
		Error: fmt.Errorf("test error"),
	}

	err := Connect(s)
	if err == nil {
		t.Fatal("期望返回错误")
	}
	if !strings.Contains(err.Error(), "invalid session") {
		t.Errorf("错误消息应包含 'invalid session'，实际: %s", err.Error())
	}
}

// TestConnectUnsupportedAuthType 测试不支持的认证类型
func TestConnectUnsupportedAuthType(t *testing.T) {
	s := &session.Session{
		Host:     "192.168.1.1",
		Port:     22,
		User:     "testuser",
		AuthType: session.AuthType("unknown"),
		Valid:    true,
	}

	err := Connect(s)
	if err == nil {
		t.Fatal("期望返回错误")
	}
	if !strings.Contains(err.Error(), "unsupported auth type") {
		t.Errorf("错误消息应包含 'unsupported auth type'，实际: %s", err.Error())
	}
}

// TestListAgentKeysNoSocket 测试无 SSH Agent 时的行为
func TestListAgentKeysNoSocket(t *testing.T) {
	// 保存并清空 SSH_AUTH_SOCK
	origSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", "")
	defer os.Setenv("SSH_AUTH_SOCK", origSock)

	_, err := ListAgentKeys()
	if err == nil {
		t.Fatal("无 SSH_AUTH_SOCK 时应返回错误")
	}
	if !strings.Contains(err.Error(), "SSH_AUTH_SOCK not set") {
		t.Errorf("错误消息应包含 'SSH_AUTH_SOCK not set'，实际: %s", err.Error())
	}
}
