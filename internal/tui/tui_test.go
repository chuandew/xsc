package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/xsc/internal/session"
)

// TestShowPasswordDefaultFalse 测试 showPassword 默认值为 false
func TestShowPasswordDefaultFalse(t *testing.T) {
	m := initialModel()
	if m.showPassword {
		t.Error("showPassword should default to false")
	}
}

// TestTogglePasswordCommand 测试 :pw 命令切换密码显示
func TestTogglePasswordCommand(t *testing.T) {
	m := initialModel()

	// 模拟输入 :pw
	m.lineNumMode = true
	m.lineNumInput.SetValue("pw")

	result, _ := m.handleLineNumInput(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	if !model.showPassword {
		t.Error("showPassword should be true after :pw command")
	}
	if model.lineNumMode {
		t.Error("lineNumMode should be false after command execution")
	}

	// 再次切换
	model.lineNumMode = true
	model.lineNumInput.SetValue("pw")

	result, _ = model.handleLineNumInput(tea.KeyMsg{Type: tea.KeyEnter})
	model = result.(Model)

	if model.showPassword {
		t.Error("showPassword should be false after second :pw command")
	}
}

// TestTogglePasswordCommandAlias 测试 :password 别名
func TestTogglePasswordCommandAlias(t *testing.T) {
	m := initialModel()

	m.lineNumMode = true
	m.lineNumInput.SetValue("password")

	result, _ := m.handleLineNumInput(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	if !model.showPassword {
		t.Error("showPassword should be true after :password command")
	}
}

// newTestModel 创建一个带有测试会话的 Model
func newTestModel(s *session.Session) Model {
	m := initialModel()
	m.width = 120
	m.height = 40
	m.tree = &session.SessionNode{
		Name:     "root",
		IsDir:    true,
		Expanded: true,
		Children: []*session.SessionNode{
			{
				Name:    "test-session",
				IsDir:   false,
				Session: s,
			},
		},
	}
	m.tree.SetParent(nil)
	m.cursor = 0
	return m
}

// TestRenderDetailMasksPassword 测试密码隐藏时显示 ********
func TestRenderDetailMasksPassword(t *testing.T) {
	s := &session.Session{
		Host:     "example.com",
		Port:     22,
		User:     "root",
		AuthType: session.AuthTypePassword,
		Password: "mysecret",
		Valid:    true,
	}
	m := newTestModel(s)

	// showPassword=false 时应该显示 ********
	detail := m.renderDetail(40, 20)
	if !strings.Contains(detail, "********") {
		t.Error("password should be masked when showPassword is false")
	}
	if strings.Contains(detail, "mysecret") {
		t.Error("actual password should not appear when showPassword is false")
	}
}

// TestRenderDetailShowsPassword 测试密码显示时显示明文
func TestRenderDetailShowsPassword(t *testing.T) {
	s := &session.Session{
		Host:     "example.com",
		Port:     22,
		User:     "root",
		AuthType: session.AuthTypePassword,
		Password: "mysecret",
		Valid:    true,
	}
	m := newTestModel(s)
	m.showPassword = true

	detail := m.renderDetail(40, 20)
	if strings.Contains(detail, "********") {
		t.Error("password should not be masked when showPassword is true")
	}
	if !strings.Contains(detail, "mysecret") {
		t.Error("actual password should appear when showPassword is true")
	}
}

// TestRenderDetailEncryptedPasswordSkipsDecrypt 测试隐藏时跳过解密
func TestRenderDetailEncryptedPasswordSkipsDecrypt(t *testing.T) {
	s := &session.Session{
		Host:              "example.com",
		Port:              22,
		User:              "root",
		AuthType:          session.AuthTypePassword,
		EncryptedPassword: "02:somefakeencrypteddata",
		Valid:             true,
	}
	m := newTestModel(s)

	// showPassword=false 时，不应调用 ResolvePassword，密码字段应该保持为空
	detail := m.renderDetail(40, 20)
	if !strings.Contains(detail, "********") {
		t.Error("encrypted password should show ******** when showPassword is false")
	}
	// 验证密码没有被解密（Password 字段仍为空）
	if s.Password != "" {
		t.Error("ResolvePassword should not have been called when showPassword is false")
	}
}

// TestStatusBarShowsPWIndicator 测试状态栏显示 [PW] 指示符
func TestStatusBarShowsPWIndicator(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 40

	// showPassword=false 时不应显示 [PW]
	bar := m.renderStatusBar()
	if strings.Contains(bar, "[PW]") {
		t.Error("[PW] should not appear when showPassword is false")
	}

	// showPassword=true 时应显示 [PW]
	m.showPassword = true
	bar = m.renderStatusBar()
	if !strings.Contains(bar, "[PW]") {
		t.Error("[PW] should appear when showPassword is true")
	}
}
