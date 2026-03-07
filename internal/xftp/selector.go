package xftp

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/user/xsc/internal/session"
	"github.com/user/xsc/pkg/config"
)

// SessionSelectedMsg 用户选择了一个 session
type SessionSelectedMsg struct {
	Session *session.Session
}

// sessionsLoadedMsg session 树加载完成（内部消息）
type sessionsLoadedMsg struct {
	tree        *session.SessionNode
	sessionsDir string
}

// Selector session 选择器
type Selector struct {
	tree      *session.SessionNode   // session 树根节点
	flatNodes []*session.SessionNode // 展平后的可见节点
	cursor    int
	offset    int
	width     int
	height    int

	// 搜索
	searchInput textinput.Model
	searching   bool
	filter      string

	// 状态
	loading    bool
	lastKeyG   bool // 检测 gg 组合
	statusMsg  string
}

// NewSelector 创建 session 选择器
func NewSelector() Selector {
	searchInput := textinput.New()
	searchInput.Placeholder = "搜索会话..."
	searchInput.Prompt = "/"
	searchInput.CharLimit = 50
	searchInput.Width = 30

	return Selector{
		loading:     true,
		searchInput: searchInput,
		statusMsg:   "加载会话中...",
	}
}

// Init 初始化选择器（加载 session 树）
func (s Selector) Init() tea.Cmd {
	return s.loadSessions()
}

// loadSessions 异步加载所有 session
func (s Selector) loadSessions() tea.Cmd {
	return func() tea.Msg {
		sessionsDir, err := config.GetSessionsDir()
		if err != nil {
			return sessionsLoadedMsg{tree: nil}
		}

		tree, err := session.LoadSessionsTree(sessionsDir)
		if err != nil {
			return sessionsLoadedMsg{tree: nil}
		}

		// 加载全局配置，添加外部 session 源
		globalConfig, err := config.LoadGlobalConfig()
		if err == nil {
			if globalConfig.SecureCRT.Enabled {
				scTree, err := session.LoadSecureCRTSessions(globalConfig.SecureCRT)
				if err == nil && scTree != nil {
					tree.Children = append(tree.Children, scTree)
				}
			}
			if globalConfig.XShell.Enabled {
				xsTree, err := session.LoadXShellSessions(globalConfig.XShell)
				if err == nil && xsTree != nil {
					tree.Children = append(tree.Children, xsTree)
				}
			}
			if globalConfig.MobaXterm.Enabled {
				mxTree, err := session.LoadMobaXtermSessions(globalConfig.MobaXterm)
				if err == nil && mxTree != nil {
					tree.Children = append(tree.Children, mxTree)
				}
			}
		}

		return sessionsLoadedMsg{tree: tree, sessionsDir: sessionsDir}
	}
}

// Update 处理选择器消息
func (s Selector) Update(msg tea.Msg) (Selector, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsLoadedMsg:
		s.loading = false
		s.tree = msg.tree
		if s.tree != nil {
			s.tree.SetParent(nil)
			s.expandAll(s.tree)
			s.updateFlatNodes()
			s.statusMsg = fmt.Sprintf("%d 个会话", s.countSessions(s.tree))
		} else {
			s.statusMsg = "未找到会话"
		}
		return s, nil

	case tea.KeyMsg:
		if s.searching {
			return s.handleSearchKey(msg)
		}
		return s.handleNormalKey(msg)
	}
	return s, nil
}

// handleNormalKey 处理普通模式的键盘输入
func (s Selector) handleNormalKey(msg tea.KeyMsg) (Selector, tea.Cmd) {
	keys := DefaultKeyMap()

	switch {
	case key.Matches(msg, keys.Quit):
		return s, tea.Quit

	case key.Matches(msg, keys.Up):
		s.moveCursor(-1)
		s.lastKeyG = false

	case key.Matches(msg, keys.Down):
		s.moveCursor(1)
		s.lastKeyG = false

	case key.Matches(msg, keys.HalfPageUp):
		h := s.viewHeight()
		s.moveCursor(-(h / 2))
		s.lastKeyG = false

	case key.Matches(msg, keys.HalfPageDown):
		h := s.viewHeight()
		s.moveCursor(h / 2)
		s.lastKeyG = false

	// gg 跳顶
	case msg.String() == "g":
		if s.lastKeyG {
			s.cursor = 0
			s.ensureVisible()
			s.lastKeyG = false
		} else {
			s.lastKeyG = true
		}
		return s, nil

	// G 跳底
	case msg.String() == "G":
		if len(s.flatNodes) > 0 {
			s.cursor = len(s.flatNodes) - 1
			s.ensureVisible()
		}
		s.lastKeyG = false

	case key.Matches(msg, keys.Enter):
		s.lastKeyG = false
		return s.selectCurrent()

	case key.Matches(msg, keys.OpenFold):
		// l — 展开目录
		s.lastKeyG = false
		node := s.currentNode()
		if node != nil && node.IsDir && !node.Expanded {
			node.Expanded = true
			s.updateFlatNodes()
		}

	case key.Matches(msg, keys.CloseFold):
		// h — 折叠目录或跳到父目录
		s.lastKeyG = false
		node := s.currentNode()
		if node != nil {
			if node.IsDir && node.Expanded {
				node.Expanded = false
				s.updateFlatNodes()
			} else if node.Parent != nil {
				// 跳到父目录
				for i, n := range s.flatNodes {
					if n == node.Parent {
						s.cursor = i
						s.ensureVisible()
						break
					}
				}
			}
		}

	case key.Matches(msg, keys.Select):
		// Space — 切换展开/折叠
		s.lastKeyG = false
		node := s.currentNode()
		if node != nil && node.IsDir {
			node.Expanded = !node.Expanded
			s.updateFlatNodes()
		}

	case key.Matches(msg, keys.Search):
		// / — 进入搜索模式
		s.lastKeyG = false
		s.searching = true
		s.searchInput.Focus()
		return s, textinput.Blink

	default:
		s.lastKeyG = false
	}

	return s, nil
}

// handleSearchKey 处理搜索模式的键盘输入
func (s Selector) handleSearchKey(msg tea.KeyMsg) (Selector, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		s.searching = false
		s.filter = ""
		s.searchInput.SetValue("")
		s.updateFlatNodes()
		s.cursor = 0
		s.offset = 0
		return s, nil

	case tea.KeyEnter:
		s.searching = false
		s.filter = s.searchInput.Value()
		s.updateFlatNodes()
		s.cursor = 0
		s.offset = 0
		return s, nil

	default:
		var cmd tea.Cmd
		s.searchInput, cmd = s.searchInput.Update(msg)
		// 实时过滤
		s.filter = s.searchInput.Value()
		s.updateFlatNodes()
		s.cursor = 0
		s.offset = 0
		return s, cmd
	}
}

// selectCurrent 选择当前光标指向的 session
func (s Selector) selectCurrent() (Selector, tea.Cmd) {
	node := s.currentNode()
	if node == nil {
		return s, nil
	}

	if node.IsDir {
		// 目录：切换展开/折叠
		node.Expanded = !node.Expanded
		s.updateFlatNodes()
		return s, nil
	}

	if node.Session != nil && node.Session.Valid {
		return s, func() tea.Msg {
			return SessionSelectedMsg{Session: node.Session}
		}
	}

	s.statusMsg = "无效的会话"
	return s, nil
}

// View 渲染选择器
func (s Selector) View(width, height int) string {
	// 标题
	title := PanelTitleActiveStyle.Width(width - 4).Render("xftp — 选择会话")

	if s.loading {
		content := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgDim)).
			Render("\n  加载中...")
		return ConfirmBoxStyle.Width(width - 2).Height(height - 2).Render(
			lipgloss.JoinVertical(lipgloss.Left, title, content),
		)
	}

	if len(s.flatNodes) == 0 {
		var msg string
		if s.filter != "" {
			msg = fmt.Sprintf("\n  未找到匹配 \"%s\" 的会话", s.filter)
		} else {
			msg = "\n  未找到会话\n\n  请先使用 xsc 导入或创建 SSH 会话"
		}
		content := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgDim)).
			Render(msg)
		return ConfirmBoxStyle.Width(width - 2).Height(height - 2).Render(
			lipgloss.JoinVertical(lipgloss.Left, title, content),
		)
	}

	// 树形列表
	viewH := s.viewHeight()
	endIdx := s.offset + viewH
	if endIdx > len(s.flatNodes) {
		endIdx = len(s.flatNodes)
	}

	var lines []string
	for i := s.offset; i < endIdx; i++ {
		lines = append(lines, s.renderNode(s.flatNodes[i], i == s.cursor, width-6))
	}
	for len(lines) < viewH {
		lines = append(lines, "")
	}
	treeContent := strings.Join(lines, "\n")

	// 状态行
	statusLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorFgDim)).
		Render(fmt.Sprintf(" %d/%d  %s", s.cursor+1, len(s.flatNodes), s.statusMsg))

	// 搜索栏
	var bottom string
	if s.searching {
		bottom = SearchStyle.Render(s.searchInput.View())
	} else if s.filter != "" {
		bottom = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorYellow)).
			Render(fmt.Sprintf(" 过滤: %s  (Esc 清除)", s.filter))
	} else {
		bottom = statusLine
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, treeContent, bottom)
	return ConfirmBoxStyle.Width(width - 2).Render(content)
}

// renderNode 渲染单个树节点
func (s Selector) renderNode(node *session.SessionNode, selected bool, width int) string {
	indent := s.getIndent(node)

	var icon string
	var name string

	if node.IsDir {
		if node.Expanded {
			icon = "▾ "
		} else {
			icon = "▸ "
		}
		// 外部来源使用前缀标记
		if node.IsSecureCRT() && node.Name == "securecrt" {
			name = "[CRT] " + node.Name + "/"
		} else if node.IsXShell() && node.Name == "xshell" {
			name = "[XSH] " + node.Name + "/"
		} else if node.IsMobaXterm() && node.Name == "mobaxterm" {
			name = "[MXT] " + node.Name + "/"
		} else {
			name = node.Name + "/"
		}
	} else {
		icon = "  "
		name = node.Name
		// 显示连接信息
		if node.Session != nil && node.Session.Valid {
			info := fmt.Sprintf(" (%s@%s:%d)", node.Session.User, node.Session.Host, node.Session.Port)
			maxName := width - len(indent) - len(icon) - len(info) - 2
			if maxName > 0 && len(name) > maxName {
				name = name[:maxName-3] + "..."
			}
			name = name + lipgloss.NewStyle().Foreground(lipgloss.Color(colorFgDim)).Render(info)
		} else if node.Session != nil && !node.Session.Valid {
			name = name + lipgloss.NewStyle().Foreground(lipgloss.Color(colorRed)).Render(" [invalid]")
		}
	}

	line := indent + icon + name

	if selected {
		return CursorStyle.Width(width).Render(line)
	}

	// 目录和文件使用不同样式
	if node.IsDir {
		if node.IsSecureCRT() {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#b16286")).Bold(true).Render(line)
		} else if node.IsXShell() {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#458588")).Bold(true).Render(line)
		} else if node.IsMobaXterm() {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#d65d0e")).Bold(true).Render(line)
		}
		return DirStyle.Render(line)
	}

	return FileStyle.Render(line)
}

// ============================================================
// 内部辅助方法
// ============================================================

// updateFlatNodes 根据展开状态和过滤条件重新生成可见节点列表
func (s *Selector) updateFlatNodes() {
	if s.tree == nil {
		s.flatNodes = nil
		return
	}

	if s.filter == "" {
		s.flatNodes = s.tree.FlattenVisible()
	} else {
		// 带过滤的展平：只显示匹配的叶子节点及其祖先路径
		s.flatNodes = s.filterNodes(s.tree)
	}
}

// filterNodes 递归过滤节点
func (s *Selector) filterNodes(node *session.SessionNode) []*session.SessionNode {
	var result []*session.SessionNode
	filterLower := strings.ToLower(s.filter)

	for _, child := range node.Children {
		if child.IsDir {
			// 目录：递归检查是否有匹配的子节点
			childMatches := s.filterNodes(child)
			if len(childMatches) > 0 {
				result = append(result, child)
				child.Expanded = true
				result = append(result, childMatches...)
			}
		} else {
			// 叶子节点：匹配名称、host、user
			if s.matchesFilter(child, filterLower) {
				result = append(result, child)
			}
		}
	}
	return result
}

// matchesFilter 检查节点是否匹配过滤条件
func (s *Selector) matchesFilter(node *session.SessionNode, filterLower string) bool {
	if strings.Contains(strings.ToLower(node.Name), filterLower) {
		return true
	}
	if node.Session != nil {
		if strings.Contains(strings.ToLower(node.Session.Host), filterLower) {
			return true
		}
		if strings.Contains(strings.ToLower(node.Session.User), filterLower) {
			return true
		}
	}
	return false
}

// currentNode 返回当前光标指向的节点
func (s *Selector) currentNode() *session.SessionNode {
	if s.cursor < 0 || s.cursor >= len(s.flatNodes) {
		return nil
	}
	return s.flatNodes[s.cursor]
}

// moveCursor 移动光标
func (s *Selector) moveCursor(delta int) {
	s.cursor += delta
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(s.flatNodes) {
		s.cursor = len(s.flatNodes) - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	s.ensureVisible()
}

// ensureVisible 确保光标在可视区域内
func (s *Selector) ensureVisible() {
	viewH := s.viewHeight()
	if s.cursor < s.offset {
		s.offset = s.cursor
	}
	if s.cursor >= s.offset+viewH {
		s.offset = s.cursor - viewH + 1
	}
}

// viewHeight 可用于树列表的高度
// 总高度减去：标题(1) + 状态行(1) + 边框(2)
func (s *Selector) viewHeight() int {
	h := s.height - 4
	if h < 1 {
		h = 1
	}
	return h
}

// getIndent 获取节点缩进
func (s Selector) getIndent(node *session.SessionNode) string {
	depth := 0
	current := node
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return strings.Repeat("  ", depth)
}

// expandAll 展开所有目录
func (s *Selector) expandAll(node *session.SessionNode) {
	if node.IsDir {
		node.Expanded = true
		for _, child := range node.Children {
			s.expandAll(child)
		}
	}
}

// countSessions 统计叶子节点数量
func (s *Selector) countSessions(node *session.SessionNode) int {
	count := 0
	for _, child := range node.Children {
		if child.IsDir {
			count += s.countSessions(child)
		} else {
			count++
		}
	}
	return count
}

// SetSize 设置尺寸（View 方法用参数传递，这里保留供外部调用）
func (s *Selector) SetSize(w, h int) {
	s.width = w
	s.height = h
}
