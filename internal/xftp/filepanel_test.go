package xftp

import (
	"os"
	"testing"
)

// TestNewFilePanel 测试创建文件面板
func TestNewFilePanel(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/tmp")
	if p.side != PanelLeft {
		t.Error("面板方向应为 PanelLeft")
	}
	if p.cwd != "/tmp" {
		t.Errorf("cwd 应为 /tmp，实际 %s", p.cwd)
	}
}

// TestFilePanelCursorMovement 测试光标移动
func TestFilePanelCursorMovement(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")
	p.height = 30
	p.entries = make([]FileEntry, 20)

	p.CursorDown()
	if p.cursor != 1 {
		t.Errorf("CursorDown 后期望光标 1，实际 %d", p.cursor)
	}

	p.CursorUp()
	if p.cursor != 0 {
		t.Errorf("CursorUp 后期望光标 0，实际 %d", p.cursor)
	}

	// 不允许光标超出上界
	p.CursorUp()
	if p.cursor != 0 {
		t.Errorf("光标不应小于 0，实际 %d", p.cursor)
	}

	// 跳到底部
	p.GoBottom()
	if p.cursor != 19 {
		t.Errorf("GoBottom 后期望光标 19，实际 %d", p.cursor)
	}

	// 不允许光标超出下界
	p.CursorDown()
	if p.cursor != 19 {
		t.Errorf("光标不应超过 19，实际 %d", p.cursor)
	}

	// 跳到顶部
	p.GoTop()
	if p.cursor != 0 {
		t.Errorf("GoTop 后期望光标 0，实际 %d", p.cursor)
	}
}

// TestFilePanelPageUpDown 测试半页滚动
func TestFilePanelPageUpDown(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")
	p.height = 20 // viewHeight = 20 - 3 = 17
	p.entries = make([]FileEntry, 50)

	// 半页下滚
	p.PageDown()
	expectedCursor := 17 / 2 // 8
	if p.cursor != expectedCursor {
		t.Errorf("PageDown 后期望光标 %d，实际 %d", expectedCursor, p.cursor)
	}

	// 半页上滚
	p.PageUp()
	if p.cursor != 0 {
		t.Errorf("PageUp 后期望光标 0，实际 %d", p.cursor)
	}
}

// TestFilePanelFullPageUpDown 测试全页滚动（Ctrl+F/B）
func TestFilePanelFullPageUpDown(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")
	p.height = 20 // viewHeight = 20 - 3 = 17
	p.entries = make([]FileEntry, 50)

	// 全页下滚
	p.FullPageDown()
	if p.cursor != 17 {
		t.Errorf("FullPageDown 后期望光标 17，实际 %d", p.cursor)
	}

	// 再次全页下滚
	p.FullPageDown()
	if p.cursor != 34 {
		t.Errorf("第二次 FullPageDown 后期望光标 34，实际 %d", p.cursor)
	}

	// 全页上滚
	p.FullPageUp()
	if p.cursor != 17 {
		t.Errorf("FullPageUp 后期望光标 17，实际 %d", p.cursor)
	}

	// 全页上滚到顶
	p.FullPageUp()
	if p.cursor != 0 {
		t.Errorf("第二次 FullPageUp 后期望光标 0，实际 %d", p.cursor)
	}
}

// TestFilePanelFullPageDownBoundary 测试全页滚动边界
func TestFilePanelFullPageDownBoundary(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")
	p.height = 20
	p.entries = make([]FileEntry, 10)

	// 全页下滚不应超过最后一项
	p.FullPageDown()
	if p.cursor != 9 {
		t.Errorf("FullPageDown 不应超过最后一项，期望 9，实际 %d", p.cursor)
	}
}

// TestFilePanelFullPageUpBoundary 测试全页上滚边界
func TestFilePanelFullPageUpBoundary(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")
	p.height = 20
	p.entries = make([]FileEntry, 10)
	p.cursor = 3

	// 全页上滚不应小于 0
	p.FullPageUp()
	if p.cursor != 0 {
		t.Errorf("FullPageUp 不应小于 0，期望 0，实际 %d", p.cursor)
	}
}

// TestFilePanelToggleSelect 测试多选切换
func TestFilePanelToggleSelect(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")
	p.entries = []FileEntry{
		{Info: FileInfo{Name: "file1"}},
		{Info: FileInfo{Name: "file2"}},
		{Info: FileInfo{Name: "file3"}},
	}

	// 选中第一个
	p.ToggleSelect()
	if !p.entries[0].Selected {
		t.Error("第一个文件应被选中")
	}
	// 光标应自动下移
	if p.cursor != 1 {
		t.Errorf("选中后光标应下移到 1，实际 %d", p.cursor)
	}

	// 选中第二个
	p.ToggleSelect()
	selected := p.SelectedFiles()
	if len(selected) != 2 {
		t.Errorf("期望选中 2 个，实际 %d", len(selected))
	}

	// 清除选择
	p.ClearSelection()
	selected = p.SelectedFiles()
	if len(selected) != 0 {
		t.Errorf("清除后应无选中，实际 %d", len(selected))
	}
}

// TestFilePanelApplyFilter 测试搜索过滤
func TestFilePanelApplyFilter(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")
	p.allEntries = []FileEntry{
		{Info: FileInfo{Name: "README.md"}},
		{Info: FileInfo{Name: "main.go"}},
		{Info: FileInfo{Name: "readme.txt"}},
	}
	p.entries = make([]FileEntry, len(p.allEntries))
	copy(p.entries, p.allEntries)

	// 过滤 "read"
	p.ApplyFilter("read")
	if len(p.entries) != 2 {
		t.Errorf("过滤 'read' 应匹配 2 个，实际 %d", len(p.entries))
	}

	// 清除过滤
	p.ClearFilter()
	if len(p.entries) != 3 {
		t.Errorf("清除过滤后应有 3 个，实际 %d", len(p.entries))
	}
}

// TestFilePanelViewHeight 测试可视高度计算
func TestFilePanelViewHeight(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")

	p.height = 20
	h := p.viewHeight()
	if h != 17 { // 20 - 3 (title + header + scroll indicator)
		t.Errorf("期望高度 17，实际 %d", h)
	}

	// 最小高度
	p.height = 1
	h = p.viewHeight()
	if h != 1 {
		t.Errorf("最小高度应为 1，实际 %d", h)
	}
}

// TestFilePanelCurrentEntry 测试获取当前条目
func TestFilePanelCurrentEntry(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")

	// 空面板
	if entry := p.CurrentEntry(); entry != nil {
		t.Error("空面板 CurrentEntry 应返回 nil")
	}

	p.entries = []FileEntry{
		{Info: FileInfo{Name: "file1"}},
	}
	p.cursor = 0

	entry := p.CurrentEntry()
	if entry == nil {
		t.Fatal("CurrentEntry 不应返回 nil")
	}
	if entry.Info.Name != "file1" {
		t.Errorf("期望 'file1'，实际 '%s'", entry.Info.Name)
	}
}

// TestFormatSize 测试文件大小格式化
func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
		{-1, "-"},
	}

	for _, tt := range tests {
		result := formatSize(tt.size)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %s, want %s", tt.size, result, tt.expected)
		}
	}
}

// TestFormatPerm 测试权限格式化
func TestFormatPerm(t *testing.T) {
	// 普通文件 -rwxr-xr-x
	result := formatPerm(os.FileMode(0755))
	if result != "-rwxr-xr-x" {
		t.Errorf("formatPerm(0755) = %s, want -rwxr-xr-x", result)
	}

	// 目录 drwxr-xr-x
	result = formatPerm(os.ModeDir | os.FileMode(0755))
	if result != "drwxr-xr-x" {
		t.Errorf("formatPerm(dir|0755) = %s, want drwxr-xr-x", result)
	}
}

// TestTruncate 测试字符串截断
func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"longfilename.txt", 10, "longfil..."},
		{"ab", 2, "ab"},
		{"abc", 2, "ab"},
		{"abcdef", 5, "ab..."},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// TestShortenPath 测试路径缩短
func TestShortenPath(t *testing.T) {
	p := NewFilePanel(PanelLeft, nil, "/")

	// 短路径不变
	result := p.shortenPath("/tmp", 20)
	if result != "/tmp" {
		t.Errorf("短路径不应被缩短: %s", result)
	}

	// 长路径被缩短
	longPath := "/very/long/path/to/some/directory"
	result = p.shortenPath(longPath, 15)
	if len(result) > 15 {
		t.Errorf("缩短后长度应 <= 15，实际 %d: %s", len(result), result)
	}
}
