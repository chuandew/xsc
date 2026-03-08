package session

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/user/xsc/internal/mobaxterm"
	"github.com/user/xsc/internal/securecrt"
	"github.com/user/xsc/internal/xshell"
	"github.com/user/xsc/pkg/config"
)

// LoadAllSessions 递归加载目录中的所有会话文件
func LoadAllSessions(rootDir string) ([]*Session, error) {
	var sessions []*Session

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}

		if info.IsDir() {
			return nil
		}

		// 只处理 .yaml 和 .yml 文件
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		session, err := LoadSession(path)
		if err != nil {
			return nil // 继续加载其他会话
		}

		sessions = append(sessions, session)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// LoadSessionsTree 以树形结构加载会话
func LoadSessionsTree(rootDir string) (*SessionNode, error) {
	root := &SessionNode{
		Name:     "sessions",
		IsDir:    true,
		Children: make([]*SessionNode, 0),
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil
		}

		parts := strings.Split(relPath, string(filepath.Separator))
		current := root

		// 遍历路径部分，构建树结构
		for i, part := range parts {
			if i == len(parts)-1 && !info.IsDir() {
				// 处理文件
				ext := strings.ToLower(filepath.Ext(part))
				if ext != ".yaml" && ext != ".yml" {
					return nil
				}

				session, err := LoadSession(path)
				if err != nil {
					return nil
				}

				node := &SessionNode{
					Name:    strings.TrimSuffix(part, ext),
					IsDir:   false,
					Session: session,
				}
				current.Children = append(current.Children, node)
			} else {
				// 处理目录
				var found *SessionNode
				for _, child := range current.Children {
					if child.IsDir && child.Name == part {
						found = child
						break
					}
				}

				if found == nil {
					found = &SessionNode{
						Name:     part,
						IsDir:    true,
						Children: make([]*SessionNode, 0),
					}
					current.Children = append(current.Children, found)
				}
				current = found
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return root, nil
}

// SessionNode 表示会话树中的节点
type SessionNode struct {
	Name     string
	IsDir    bool
	Expanded bool
	Session  *Session
	Children []*SessionNode
	Parent   *SessionNode
}

// IsLeaf 检查节点是否为叶子节点（会话文件）
func (n *SessionNode) IsLeaf() bool {
	return !n.IsDir
}

// IsSecureCRT 检查节点或其祖先是否为 SecureCRT 会话
func (n *SessionNode) IsSecureCRT() bool {
	current := n
	for current != nil {
		if current.Name == "securecrt" {
			return true
		}
		current = current.Parent
	}
	return false
}

// IsXShell 检查节点或其祖先是否为 XShell 会话
func (n *SessionNode) IsXShell() bool {
	current := n
	for current != nil {
		if current.Name == "xshell" {
			return true
		}
		current = current.Parent
	}
	return false
}

// IsMobaXterm 检查节点或其祖先是否为 MobaXterm 会话
func (n *SessionNode) IsMobaXterm() bool {
	current := n
	for current != nil {
		if current.Name == "mobaxterm" {
			return true
		}
		current = current.Parent
	}
	return false
}

// IsReadOnly 检查节点是否为只读（导入的外部会话）
func (n *SessionNode) IsReadOnly() bool {
	return n.IsSecureCRT() || n.IsXShell() || n.IsMobaXterm()
}

// GetPath 返回从根节点到当前节点的路径
func (n *SessionNode) GetPath() string {
	if n.Parent == nil {
		return n.Name
	}
	return filepath.Join(n.Parent.GetPath(), n.Name)
}

// FlattenVisible 返回可见的节点列表（考虑展开/折叠状态）
func (n *SessionNode) FlattenVisible() []*SessionNode {
	var result []*SessionNode

	// 遍历所有子节点，无论根节点还是非根节点逻辑相同
	for _, child := range n.Children {
		result = append(result, child)
		if child.IsDir && child.Expanded {
			result = append(result, child.FlattenVisible()...)
		}
	}

	return result
}

// FindNode 根据路径查找节点
func (n *SessionNode) FindNode(path string) *SessionNode {
	if n.GetPath() == path {
		return n
	}

	for _, child := range n.Children {
		if found := child.FindNode(path); found != nil {
			return found
		}
	}

	return nil
}

// SetParent 递归设置父节点引用
func (n *SessionNode) SetParent(parent *SessionNode) {
	n.Parent = parent
	for _, child := range n.Children {
		child.SetParent(n)
	}
}

// GetSessionPath 返回会话的相对路径（用于显示）
func GetSessionPath(rootDir string, session *Session) string {
	relPath, _ := filepath.Rel(rootDir, session.FilePath)
	return strings.TrimSuffix(relPath, ".yaml")
}

// importedSessionEntry 表示从外部来源导入的单个会话条目
type importedSessionEntry struct {
	Name    string
	Folder  string
	Session *Session
}

// buildImportedTree 通用的导入会话树构建函数
func buildImportedTree(rootName string, entries []importedSessionEntry) *SessionNode {
	root := &SessionNode{
		Name:     rootName,
		IsDir:    true,
		Expanded: true,
		Children: make([]*SessionNode, 0),
	}

	for _, entry := range entries {
		node := &SessionNode{
			Name:    entry.Name,
			IsDir:   false,
			Session: entry.Session,
		}

		if entry.Folder != "" {
			folderPath := strings.Split(entry.Folder, string(filepath.Separator))
			current := root

			for _, folderName := range folderPath {
				var found *SessionNode
				for _, child := range current.Children {
					if child.IsDir && child.Name == folderName {
						found = child
						break
					}
				}

				if found == nil {
					found = &SessionNode{
						Name:     folderName,
						IsDir:    true,
						Children: make([]*SessionNode, 0),
					}
					current.Children = append(current.Children, found)
				}
				current = found
			}

			current.Children = append(current.Children, node)
		} else {
			root.Children = append(root.Children, node)
		}
	}

	return root
}

// convertSessionData 从 sessionData map 中提取通用会话字段
func convertSessionData(sessionData map[string]interface{}, passwordSource, masterPassword string) *Session {
	host, _ := sessionData["host"].(string)
	port, _ := sessionData["port"].(int)
	user, _ := sessionData["user"].(string)
	authTypeStr, _ := sessionData["auth_type"].(string)

	s := &Session{
		Host:           host,
		Port:           port,
		User:           user,
		AuthType:       AuthType(authTypeStr),
		Valid:          true,
		PasswordSource: passwordSource,
	}

	// 处理已解密的密码
	if pwd, ok := sessionData["password"].(string); ok && pwd != "" {
		s.Password = pwd
	}

	// 保存加密密码和主密码，用于延迟解密
	if ep, ok := sessionData["encrypted_password"].(string); ok && ep != "" {
		s.EncryptedPassword = ep
		s.MasterPassword = masterPassword
	}

	return s
}

// LoadSecureCRTSessions 加载 SecureCRT 会话
func LoadSecureCRTSessions(cfg config.SecureCRTConfig) (*SessionNode, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	scConfig := securecrt.Config{
		SessionPath: cfg.SessionPath,
		Password:    cfg.Password,
	}

	sessions, err := securecrt.LoadSessions(scConfig)
	if err != nil {
		return nil, err
	}

	var entries []importedSessionEntry
	for _, scSession := range sessions {
		sessionData := scSession.ConvertToXSSHSession()
		s := convertSessionData(sessionData, "securecrt", cfg.Password)

		// SecureCRT 特有：处理认证方法列表
		if authMethods, ok := sessionData["auth_methods"].([]securecrt.AuthMethod); ok {
			for _, am := range authMethods {
				authMethod := AuthMethod{
					Type:     am.Type,
					Priority: am.Priority,
					KeyPath:  am.KeyFile,
				}
				if am.Type == "password" && am.Password != "" {
					authMethod.EncryptedPassword = am.Password
					s.MasterPassword = cfg.Password
				}
				s.AuthMethods = append(s.AuthMethods, authMethod)
			}
		}

		entries = append(entries, importedSessionEntry{
			Name:    scSession.Name,
			Folder:  scSession.Folder,
			Session: s,
		})
	}

	return buildImportedTree("securecrt", entries), nil
}

// LoadMobaXtermSessions 加载 MobaXterm 会话
func LoadMobaXtermSessions(cfg config.MobaXtermConfig) (*SessionNode, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	mxConfig := mobaxterm.Config{
		SessionPath: cfg.SessionPath,
		Password:    cfg.Password,
	}

	sessions, err := mobaxterm.LoadSessions(mxConfig)
	if err != nil {
		return nil, err
	}

	var entries []importedSessionEntry
	for _, mxSession := range sessions {
		sessionData := mxSession.ConvertToXSSHSession()
		entries = append(entries, importedSessionEntry{
			Name:    mxSession.Name,
			Folder:  mxSession.Folder,
			Session: convertSessionData(sessionData, "mobaxterm", cfg.Password),
		})
	}

	return buildImportedTree("mobaxterm", entries), nil
}

// LoadXShellSessions 加载 XShell 会话
func LoadXShellSessions(cfg config.XShellConfig) (*SessionNode, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	xsConfig := xshell.Config{
		SessionPath: cfg.SessionPath,
		Password:    cfg.Password,
	}

	sessions, err := xshell.LoadSessions(xsConfig)
	if err != nil {
		return nil, err
	}

	var entries []importedSessionEntry
	for _, xsSession := range sessions {
		sessionData := xsSession.ConvertToXSSHSession()
		entries = append(entries, importedSessionEntry{
			Name:    xsSession.Name,
			Folder:  xsSession.Folder,
			Session: convertSessionData(sessionData, "xshell", cfg.Password),
		})
	}

	return buildImportedTree("xshell", entries), nil
}
