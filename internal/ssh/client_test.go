package ssh

import (
	"fmt"
	"strings"
	"testing"

	"github.com/user/xsc/internal/session"
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
