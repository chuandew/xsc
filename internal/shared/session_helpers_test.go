package shared

import (
	"testing"

	"github.com/ketor/xsc/internal/session"
)

// TestMatchCommand 测试命令匹配
func TestMatchCommand(t *testing.T) {
	commands := []Command{
		{Name: "q", Aliases: []string{"quit"}, Description: "退出程序"},
		{Name: "pw", Aliases: []string{"password"}, Description: "切换密码显示"},
		{Name: "noh", Aliases: []string{"nohlsearch"}, Description: "清除搜索过滤"},
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"q", "q"},
		{"quit", "q"},
		{"pw", "pw"},
		{"password", "pw"},
		{"noh", "noh"},
		{"nohlsearch", "noh"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MatchCommand(tt.input, commands)
			if result != tt.expected {
				t.Errorf("MatchCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetCommandCompletions 测试命令补全
func TestGetCommandCompletions(t *testing.T) {
	commands := []Command{
		{Name: "q", Aliases: []string{"quit"}, Description: "退出程序"},
		{Name: "pw", Aliases: []string{"password"}, Description: "切换密码显示"},
		{Name: "noh", Aliases: []string{"nohlsearch"}, Description: "清除搜索过滤"},
	}

	// 空前缀返回所有命令
	all := GetCommandCompletions("", commands)
	if len(all) != len(commands) {
		t.Errorf("空前缀应返回 %d 个命令，实际返回 %d", len(commands), len(all))
	}

	// "p" 前缀应匹配 "pw"
	pCmds := GetCommandCompletions("p", commands)
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
	zCmds := GetCommandCompletions("z", commands)
	if len(zCmds) != 0 {
		t.Errorf("前缀 'z' 应无匹配，实际返回 %d", len(zCmds))
	}

	// 别名匹配：前缀 "pass" 应匹配 "pw"
	passCmds := GetCommandCompletions("pass", commands)
	found = false
	for _, cmd := range passCmds {
		if cmd.Name == "pw" {
			found = true
		}
	}
	if !found {
		t.Error("前缀 'pass' 应通过别名匹配 'pw' 命令")
	}
}

// TestGetIndent 测试缩进计算
func TestGetIndent(t *testing.T) {
	root := &session.SessionNode{Name: "root", IsDir: true}
	child := &session.SessionNode{Name: "child", IsDir: true, Parent: root}
	grandchild := &session.SessionNode{Name: "gc", IsDir: false, Parent: child}

	if indent := GetIndent(root); indent != "" {
		t.Errorf("根节点缩进应为空，实际 %q", indent)
	}
	if indent := GetIndent(child); indent != "  " {
		t.Errorf("子节点缩进应为 2 空格，实际 %q", indent)
	}
	if indent := GetIndent(grandchild); indent != "    " {
		t.Errorf("孙节点缩进应为 4 空格，实际 %q", indent)
	}
}

// TestCountSessions 测试会话计数
func TestCountSessions(t *testing.T) {
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

	count := CountSessions(tree)
	if count != 3 {
		t.Errorf("期望 3 个会话，实际 %d", count)
	}

	// 空树
	empty := &session.SessionNode{IsDir: true}
	if c := CountSessions(empty); c != 0 {
		t.Errorf("空树期望 0 个会话，实际 %d", c)
	}
}

// TestExpandAll 测试展开所有目录
func TestExpandAll(t *testing.T) {
	tree := &session.SessionNode{
		IsDir: true,
		Children: []*session.SessionNode{
			{Name: "dir1", IsDir: true, Children: []*session.SessionNode{
				{Name: "dir2", IsDir: true},
			}},
			{Name: "file1", IsDir: false},
		},
	}

	ExpandAll(tree)

	if !tree.Expanded {
		t.Error("根目录应已展开")
	}
	if !tree.Children[0].Expanded {
		t.Error("dir1 应已展开")
	}
	if !tree.Children[0].Children[0].Expanded {
		t.Error("dir2 应已展开")
	}
}

// TestCollapseAll 测试折叠所有目录
func TestCollapseAll(t *testing.T) {
	tree := &session.SessionNode{
		IsDir:    true,
		Expanded: true,
		Children: []*session.SessionNode{
			{Name: "dir1", IsDir: true, Expanded: true, Children: []*session.SessionNode{
				{Name: "dir2", IsDir: true, Expanded: true},
			}},
		},
	}

	CollapseAll(tree)

	if tree.Expanded {
		t.Error("根目录应已折叠")
	}
	if tree.Children[0].Expanded {
		t.Error("dir1 应已折叠")
	}
	if tree.Children[0].Children[0].Expanded {
		t.Error("dir2 应已折叠")
	}
}
