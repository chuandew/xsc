package session

import (
	"testing"
)

func TestSessionNodeIsSecureCRT(t *testing.T) {
	root := &SessionNode{
		Name:     "sessions",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}

	localSession := &SessionNode{
		Name:    "local-session",
		IsDir:   false,
		Session: &Session{},
	}
	root.Children = append(root.Children, localSession)

	securecrtDir := &SessionNode{
		Name:     "securecrt",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}
	root.Children = append(root.Children, securecrtDir)

	scSession := &SessionNode{
		Name:    "sc-session",
		IsDir:   false,
		Session: &Session{},
	}
	securecrtDir.Children = append(securecrtDir.Children, scSession)

	root.SetParent(nil)

	if localSession.IsSecureCRT() {
		t.Error("Local session should not be detected as SecureCRT")
	}
	if !securecrtDir.IsSecureCRT() {
		t.Error("SecureCRT directory should be detected as SecureCRT")
	}
	if !scSession.IsSecureCRT() {
		t.Error("SecureCRT session should be detected as SecureCRT")
	}
	if root.IsSecureCRT() {
		t.Error("Root node should not be detected as SecureCRT")
	}
}

func TestSessionNodeIsSecureCRTNested(t *testing.T) {
	root := &SessionNode{
		Name:     "sessions",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}

	securecrtDir := &SessionNode{
		Name:     "securecrt",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}
	root.Children = append(root.Children, securecrtDir)

	folder := &SessionNode{
		Name:     "folder",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}
	securecrtDir.Children = append(securecrtDir.Children, folder)

	nestedSession := &SessionNode{
		Name:    "nested-session",
		IsDir:   false,
		Session: &Session{},
	}
	folder.Children = append(folder.Children, nestedSession)

	root.SetParent(nil)

	if !nestedSession.IsSecureCRT() {
		t.Error("Nested session under SecureCRT should be detected as SecureCRT")
	}
	if !folder.IsSecureCRT() {
		t.Error("Folder under SecureCRT should be detected as SecureCRT")
	}
}

// TestSessionNodeIsXShell 测试 XShell 节点检测
func TestSessionNodeIsXShell(t *testing.T) {
	root := &SessionNode{
		Name:     "sessions",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}

	xshellDir := &SessionNode{
		Name:     "xshell",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}
	root.Children = append(root.Children, xshellDir)

	xsSession := &SessionNode{
		Name:    "xs-session",
		IsDir:   false,
		Session: &Session{},
	}
	xshellDir.Children = append(xshellDir.Children, xsSession)

	root.SetParent(nil)

	if !xsSession.IsXShell() {
		t.Error("XShell 子节点应检测为 XShell")
	}
	if !xshellDir.IsXShell() {
		t.Error("XShell 目录应检测为 XShell")
	}
	if root.IsXShell() {
		t.Error("根节点不应检测为 XShell")
	}
}

// TestSessionNodeIsMobaXterm 测试 MobaXterm 节点检测
func TestSessionNodeIsMobaXterm(t *testing.T) {
	root := &SessionNode{
		Name:     "sessions",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}

	mxDir := &SessionNode{
		Name:     "mobaxterm",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}
	root.Children = append(root.Children, mxDir)

	mxSession := &SessionNode{
		Name:    "mx-session",
		IsDir:   false,
		Session: &Session{},
	}
	mxDir.Children = append(mxDir.Children, mxSession)

	root.SetParent(nil)

	if !mxSession.IsMobaXterm() {
		t.Error("MobaXterm 子节点应检测为 MobaXterm")
	}
	if !mxDir.IsMobaXterm() {
		t.Error("MobaXterm 目录应检测为 MobaXterm")
	}
	if root.IsMobaXterm() {
		t.Error("根节点不应检测为 MobaXterm")
	}
}

// TestSessionNodeIsReadOnly 测试只读节点判断
func TestSessionNodeIsReadOnly(t *testing.T) {
	root := &SessionNode{
		Name:     "sessions",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}

	// SecureCRT 子节点应为只读
	scDir := &SessionNode{Name: "securecrt", IsDir: true, Children: make([]*SessionNode, 0)}
	scNode := &SessionNode{Name: "sc1", IsDir: false, Session: &Session{}}
	scDir.Children = append(scDir.Children, scNode)
	root.Children = append(root.Children, scDir)

	// 本地节点应非只读
	localNode := &SessionNode{Name: "local1", IsDir: false, Session: &Session{}}
	root.Children = append(root.Children, localNode)

	root.SetParent(nil)

	if !scNode.IsReadOnly() {
		t.Error("SecureCRT 节点应为只读")
	}
	if localNode.IsReadOnly() {
		t.Error("本地节点不应为只读")
	}
}

// TestSessionNodeIsLeaf 测试叶子节点判断
func TestSessionNodeIsLeaf(t *testing.T) {
	dirNode := &SessionNode{Name: "dir", IsDir: true}
	fileNode := &SessionNode{Name: "file", IsDir: false}

	if dirNode.IsLeaf() {
		t.Error("目录不应为叶子节点")
	}
	if !fileNode.IsLeaf() {
		t.Error("文件应为叶子节点")
	}
}

// TestSessionNodeGetPath 测试路径获取
func TestSessionNodeGetPath(t *testing.T) {
	root := &SessionNode{Name: "sessions", IsDir: true, Children: make([]*SessionNode, 0)}
	child := &SessionNode{Name: "group1", IsDir: true, Children: make([]*SessionNode, 0)}
	leaf := &SessionNode{Name: "server1", IsDir: false}
	root.Children = append(root.Children, child)
	child.Children = append(child.Children, leaf)
	root.SetParent(nil)

	rootPath := root.GetPath()
	if rootPath != "sessions" {
		t.Errorf("根节点路径应为 'sessions'，实际: %s", rootPath)
	}

	childPath := child.GetPath()
	if childPath != "sessions/group1" {
		t.Errorf("子节点路径应为 'sessions/group1'，实际: %s", childPath)
	}

	leafPath := leaf.GetPath()
	if leafPath != "sessions/group1/server1" {
		t.Errorf("叶子路径应为 'sessions/group1/server1'，实际: %s", leafPath)
	}
}

// TestSessionNodeFlattenVisible 测试展平可见节点
func TestSessionNodeFlattenVisible(t *testing.T) {
	root := &SessionNode{
		Name:     "root",
		IsDir:    true,
		Expanded: true,
		Children: []*SessionNode{
			{Name: "file1", IsDir: false},
			{
				Name:     "dir1",
				IsDir:    true,
				Expanded: true,
				Children: []*SessionNode{
					{Name: "file2", IsDir: false},
				},
			},
			{
				Name:     "dir2",
				IsDir:    true,
				Expanded: false,
				Children: []*SessionNode{
					{Name: "hidden", IsDir: false},
				},
			},
		},
	}

	visible := root.FlattenVisible()
	// 应看到: file1, dir1, file2, dir2（dir2 的 hidden 因折叠不可见）
	if len(visible) != 4 {
		t.Errorf("期望 4 个可见节点，实际: %d", len(visible))
	}
}

// TestSessionNodeFlattenVisibleAllCollapsed 测试全部折叠时的可见节点
func TestSessionNodeFlattenVisibleAllCollapsed(t *testing.T) {
	root := &SessionNode{
		Name:     "root",
		IsDir:    true,
		Expanded: false,
		Children: []*SessionNode{
			{Name: "file1", IsDir: false},
			{Name: "dir1", IsDir: true, Expanded: false, Children: []*SessionNode{
				{Name: "file2", IsDir: false},
			}},
		},
	}

	// 根节点折叠时，FlattenVisible 仍然列出根的子节点
	visible := root.FlattenVisible()
	if len(visible) != 2 {
		t.Errorf("期望 2 个可见节点，实际: %d", len(visible))
	}
}

// TestSessionNodeFindNode 测试按路径查找节点
func TestSessionNodeFindNode(t *testing.T) {
	root := &SessionNode{Name: "sessions", IsDir: true, Children: make([]*SessionNode, 0)}
	child := &SessionNode{Name: "group1", IsDir: true, Children: make([]*SessionNode, 0)}
	leaf := &SessionNode{Name: "server1", IsDir: false}
	root.Children = append(root.Children, child)
	child.Children = append(child.Children, leaf)
	root.SetParent(nil)

	found := root.FindNode("sessions/group1/server1")
	if found == nil {
		t.Fatal("应能找到节点")
	}
	if found.Name != "server1" {
		t.Errorf("找到的节点名称应为 server1，实际: %s", found.Name)
	}

	notFound := root.FindNode("sessions/nonexistent")
	if notFound != nil {
		t.Error("查找不存在路径应返回 nil")
	}
}

// TestSessionNodeSetParent 测试递归设置父节点
func TestSessionNodeSetParent(t *testing.T) {
	root := &SessionNode{Name: "root", IsDir: true, Children: make([]*SessionNode, 0)}
	child := &SessionNode{Name: "child", IsDir: true, Children: make([]*SessionNode, 0)}
	grandchild := &SessionNode{Name: "gc", IsDir: false}
	root.Children = append(root.Children, child)
	child.Children = append(child.Children, grandchild)

	root.SetParent(nil)

	if root.Parent != nil {
		t.Error("根节点的 Parent 应为 nil")
	}
	if child.Parent != root {
		t.Error("child 的 Parent 应为 root")
	}
	if grandchild.Parent != child {
		t.Error("grandchild 的 Parent 应为 child")
	}
}

// TestBuildImportedTree 测试导入会话树构建
func TestBuildImportedTree(t *testing.T) {
	entries := []importedSessionEntry{
		{Name: "srv1", Folder: "", Session: &Session{Host: "10.0.0.1"}},
		{Name: "srv2", Folder: "group1", Session: &Session{Host: "10.0.0.2"}},
		{Name: "srv3", Folder: "group1/sub", Session: &Session{Host: "10.0.0.3"}},
	}

	tree := buildImportedTree("testroot", entries)

	if tree.Name != "testroot" {
		t.Errorf("根节点名称应为 'testroot'，实际: %s", tree.Name)
	}
	if !tree.IsDir {
		t.Error("根节点应为目录")
	}
	if !tree.Expanded {
		t.Error("根节点应展开")
	}

	// srv1 应在根级
	// group1/ 应在根级
	if len(tree.Children) != 2 {
		t.Errorf("期望 2 个子节点（1 文件 + 1 目录），实际: %d", len(tree.Children))
	}
}

// TestBuildImportedTreeNestedFolders 测试导入嵌套目录
func TestBuildImportedTreeNestedFolders(t *testing.T) {
	entries := []importedSessionEntry{
		{Name: "srv1", Folder: "a/b/c", Session: &Session{Host: "10.0.0.1"}},
	}

	tree := buildImportedTree("root", entries)

	// root -> a -> b -> c -> srv1
	if len(tree.Children) != 1 {
		t.Fatalf("期望 1 个子节点，实际: %d", len(tree.Children))
	}
	a := tree.Children[0]
	if a.Name != "a" || !a.IsDir {
		t.Errorf("第一层应为目录 'a'，实际: %s isDir=%v", a.Name, a.IsDir)
	}
	if len(a.Children) != 1 {
		t.Fatalf("'a' 应有 1 个子节点，实际: %d", len(a.Children))
	}
}

// TestConvertSessionData 测试会话数据转换
func TestConvertSessionData(t *testing.T) {
	data := map[string]interface{}{
		"host":               "10.0.0.1",
		"port":               22,
		"user":               "admin",
		"auth_type":          "password",
		"password":           "secret",
		"encrypted_password": "enc123",
	}

	s := convertSessionData(data, "securecrt", "masterpass")

	if s.Host != "10.0.0.1" {
		t.Errorf("Host = %s, want 10.0.0.1", s.Host)
	}
	if s.Port != 22 {
		t.Errorf("Port = %d, want 22", s.Port)
	}
	if s.User != "admin" {
		t.Errorf("User = %s, want admin", s.User)
	}
	if s.Password != "secret" {
		t.Errorf("Password = %s, want secret", s.Password)
	}
	if s.EncryptedPassword != "enc123" {
		t.Errorf("EncryptedPassword = %s, want enc123", s.EncryptedPassword)
	}
	if s.PasswordSource != "securecrt" {
		t.Errorf("PasswordSource = %s, want securecrt", s.PasswordSource)
	}
	if s.MasterPassword != "masterpass" {
		t.Errorf("MasterPassword = %s, want masterpass", s.MasterPassword)
	}
}

// TestConvertSessionDataPartialFields 测试部分字段缺失时的转换
func TestConvertSessionDataPartialFields(t *testing.T) {
	data := map[string]interface{}{
		"host": "10.0.0.1",
	}

	s := convertSessionData(data, "xshell", "")

	if s.Host != "10.0.0.1" {
		t.Errorf("Host = %s, want 10.0.0.1", s.Host)
	}
	if s.Port != 0 {
		t.Errorf("Port = %d, want 0（未设置）", s.Port)
	}
	if s.User != "" {
		t.Errorf("User = %s, want empty", s.User)
	}
}

// TestGetSessionPath 测试获取会话的相对路径
func TestGetSessionPath(t *testing.T) {
	s := &Session{
		FilePath: "/home/user/.xsc/sessions/group1/server.yaml",
	}

	result := GetSessionPath("/home/user/.xsc/sessions", s)
	if result != "group1/server" {
		t.Errorf("GetSessionPath = %s, want group1/server", result)
	}
}
