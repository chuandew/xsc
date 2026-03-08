package xftp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/xsc/internal/session"
)

// TestNewSelector 测试创建选择器
func TestNewSelector(t *testing.T) {
	s := NewSelector()
	if !s.loading {
		t.Error("新选择器应处于 loading 状态")
	}
	if s.showPassword {
		t.Error("showPassword 默认应为 false")
	}
	if s.searching {
		t.Error("新选择器不应处于搜索状态")
	}
	if s.commanding {
		t.Error("新选择器不应处于命令状态")
	}
}

// TestSelectorCommandRegistration 测试命令注册
func TestSelectorCommandRegistration(t *testing.T) {
	// 验证 :pw 命令已注册
	found := false
	for _, cmd := range selectorCommands {
		if cmd.Name == "pw" {
			found = true
			break
		}
	}
	if !found {
		t.Error("selectorCommands 应包含 'pw' 命令")
	}
}

// TestMatchSelectorCommand 测试命令匹配
func TestMatchSelectorCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"q", "q"},
		{"quit", "q"},
		{"noh", "noh"},
		{"nohlsearch", "noh"},
		{"pw", "pw"},
		{"password", "pw"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := matchSelectorCommand(tt.input)
			if result != tt.expected {
				t.Errorf("matchSelectorCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetSelectorCommandCompletions 测试命令补全
func TestGetSelectorCommandCompletions(t *testing.T) {
	// 空前缀返回所有命令
	all := getSelectorCommandCompletions("")
	if len(all) != len(selectorCommands) {
		t.Errorf("空前缀应返回 %d 个命令，实际返回 %d", len(selectorCommands), len(all))
	}

	// "p" 前缀应匹配 "pw"
	pCmds := getSelectorCommandCompletions("p")
	found := false
	for _, cmd := range pCmds {
		if cmd.Name == "pw" {
			found = true
		}
	}
	if !found {
		t.Error("前缀 'p' 应匹配 'pw' 命令")
	}

	// "z" 前缀应无匹配
	zCmds := getSelectorCommandCompletions("z")
	if len(zCmds) != 0 {
		t.Errorf("前缀 'z' 应无匹配，实际返回 %d", len(zCmds))
	}
}

// TestSelectorTogglePassword 测试 :pw 命令切换密码显示
func TestSelectorTogglePassword(t *testing.T) {
	s := NewSelector()

	if s.showPassword {
		t.Error("初始 showPassword 应为 false")
	}

	// 模拟输入 :pw
	s.commanding = true
	s.commandInput.SetValue("pw")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	s, _ = s.handleCommandKey(msg)

	if !s.showPassword {
		t.Error(":pw 后 showPassword 应为 true")
	}

	// 再次切换
	s.commanding = true
	s.commandInput.SetValue("pw")
	s, _ = s.handleCommandKey(msg)

	if s.showPassword {
		t.Error("再次 :pw 后 showPassword 应为 false")
	}
}

// TestSelectorTogglePasswordAlias 测试 :password 别名
func TestSelectorTogglePasswordAlias(t *testing.T) {
	s := NewSelector()

	s.commanding = true
	s.commandInput.SetValue("password")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	s, _ = s.handleCommandKey(msg)

	if !s.showPassword {
		t.Error(":password 别名应切换 showPassword 为 true")
	}
}

// TestSelectorRenderNodeHighlight 测试选中节点的完整高亮
func TestSelectorRenderNodeHighlight(t *testing.T) {
	s := NewSelector()

	// 创建一个有效会话节点
	node := &session.SessionNode{
		Name:  "test-server",
		IsDir: false,
		Session: &session.Session{
			Host:  "192.168.1.1",
			Port:  22,
			User:  "root",
			Valid: true,
		},
	}

	// 渲染选中状态
	selectedLine := s.renderNode(node, true, 80)
	// 渲染非选中状态
	unselectedLine := s.renderNode(node, false, 80)

	// 选中行和非选中行应该不同
	if selectedLine == unselectedLine {
		t.Error("选中行和非选中行应该不同")
	}

	// 选中行不应包含 dimInfo 的单独样式（因为选中时使用纯文本）
	// 连接信息应该包含在输出中
	if !strings.Contains(selectedLine, "192.168.1.1") {
		t.Error("选中行应包含主机地址")
	}
}

// TestSelectorRenderNodeDirHighlight 测试选中目录节点的高亮
func TestSelectorRenderNodeDirHighlight(t *testing.T) {
	s := NewSelector()

	// IsSecureCRT() 通过检查祖先名称来判断
	// 创建一个 "securecrt" 父节点
	parent := &session.SessionNode{
		Name:  "securecrt",
		IsDir: true,
	}
	node := &session.SessionNode{
		Name:     "01.WRT",
		IsDir:    true,
		Expanded: true,
		Parent:   parent,
	}

	selectedLine := s.renderNode(node, true, 60)
	if !strings.Contains(selectedLine, "[CRT]") {
		t.Error("选中的 SecureCRT 目录应包含 [CRT] 标记")
	}
}

// TestSelectorCountSessions 测试会话计数
func TestSelectorCountSessions(t *testing.T) {
	s := NewSelector()
	tree := &session.SessionNode{
		IsDir: true,
		Children: []*session.SessionNode{
			{Name: "srv1", IsDir: false},
			{Name: "srv2", IsDir: false},
			{
				Name:  "group",
				IsDir: true,
				Children: []*session.SessionNode{
					{Name: "srv3", IsDir: false},
				},
			},
		},
	}

	count := s.countSessions(tree)
	if count != 3 {
		t.Errorf("期望 3 个会话，实际 %d", count)
	}
}

// TestSelectorViewHeight 测试可视高度计算
func TestSelectorViewHeight(t *testing.T) {
	s := NewSelector()

	s.height = 30
	h := s.viewHeight()
	if h != 26 { // 30 - 4
		t.Errorf("期望高度 26，实际 %d", h)
	}

	// 最小高度
	s.height = 2
	h = s.viewHeight()
	if h != 1 {
		t.Errorf("最小高度应为 1，实际 %d", h)
	}
}

// TestSelectorMoveCursor 测试光标移动
func TestSelectorMoveCursor(t *testing.T) {
	s := NewSelector()
	s.height = 30
	s.flatNodes = make([]*session.SessionNode, 10)

	s.moveCursor(3)
	if s.cursor != 3 {
		t.Errorf("期望光标位置 3，实际 %d", s.cursor)
	}

	// 不超过上界
	s.moveCursor(-10)
	if s.cursor != 0 {
		t.Errorf("光标不应小于 0，实际 %d", s.cursor)
	}

	// 不超过下界
	s.moveCursor(100)
	if s.cursor != 9 {
		t.Errorf("光标不应超过 %d，实际 %d", 9, s.cursor)
	}
}

// TestSelectorSearchNext 测试搜索下一个
func TestSelectorSearchNext(t *testing.T) {
	s := NewSelector()
	s.height = 30
	s.filter = "test"
	s.flatNodes = []*session.SessionNode{
		{Name: "other1", IsDir: false, Session: &session.Session{}},
		{Name: "test-1", IsDir: false, Session: &session.Session{}},
		{Name: "other2", IsDir: false, Session: &session.Session{}},
		{Name: "test-2", IsDir: false, Session: &session.Session{}},
	}
	s.cursor = 0

	// 搜索下一个匹配
	s.searchNext(1)
	if s.cursor != 1 {
		t.Errorf("期望光标跳到 1，实际 %d", s.cursor)
	}

	s.searchNext(1)
	if s.cursor != 3 {
		t.Errorf("期望光标跳到 3，实际 %d", s.cursor)
	}
}

// TestSelectorGetIndent 测试缩进计算
func TestSelectorGetIndent(t *testing.T) {
	s := NewSelector()

	root := &session.SessionNode{Name: "root", IsDir: true}
	child := &session.SessionNode{Name: "child", IsDir: true, Parent: root}
	grandchild := &session.SessionNode{Name: "gc", IsDir: false, Parent: child}

	if indent := s.getIndent(root); indent != "" {
		t.Errorf("根节点缩进应为空，实际 %q", indent)
	}
	if indent := s.getIndent(child); indent != "  " {
		t.Errorf("子节点缩进应为 2 空格，实际 %q", indent)
	}
	if indent := s.getIndent(grandchild); indent != "    " {
		t.Errorf("孙节点缩进应为 4 空格，实际 %q", indent)
	}
}
