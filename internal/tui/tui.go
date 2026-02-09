package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/xsc/internal/session"
	internalssh "github.com/user/xsc/internal/ssh"
	"github.com/user/xsc/pkg/config"
)

// 样式定义
var (
	// 树形结构样式
	treeStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3c3836")).
			Foreground(lipgloss.Color("#fabd2f")).
			Bold(true)

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#83a598")).
			Bold(true)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#b8bb26"))

	invalidStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fb4934"))

	lineNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#665c54")).
			Width(4).
			Align(lipgloss.Right)

	// 详情面板样式
	detailTitleStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#504945")).
				Foreground(lipgloss.Color("#ebdbb2")).
				Bold(true).
				Padding(0, 1)

	detailContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ebdbb2")).
				Padding(1)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#665c54")).
			Padding(1)

	detailKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#83a598"))

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fabd2f"))

	// 状态栏样式
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3c3836")).
			Foreground(lipgloss.Color("#a89984")).
			Padding(0, 1)

	// 搜索框样式
	searchStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#282828")).
			Foreground(lipgloss.Color("#ebdbb2")).
			Padding(0, 1)

	// 命令补全提示样式
	cmdHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#665c54"))

	cmdHintActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fabd2f"))

	// 帮助视图样式
	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fabd2f")).
				Bold(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#83a598")).
			Width(16)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ebdbb2"))

	helpContainerStyle = lipgloss.NewStyle().
				Padding(1, 2)
)

// Command 定义一个 : 模式下的命令
type Command struct {
	Name        string   // 主命令名, e.g. "q"
	Aliases     []string // 别名, e.g. ["quit"]
	Description string   // 中文描述
}

// commands 是所有 : 命令的注册表（单一数据源）
var commands = []Command{
	{Name: "q", Aliases: []string{"quit"}, Description: "退出程序"},
	{Name: "noh", Aliases: []string{"nohlsearch"}, Description: "清除搜索高亮/过滤"},
	{Name: "pw", Aliases: []string{"password"}, Description: "切换密码明文显示"},
}

// matchCommand 根据输入返回匹配的命令规范名，无匹配返回空字符串
func matchCommand(input string) string {
	for _, cmd := range commands {
		if input == cmd.Name {
			return cmd.Name
		}
		for _, alias := range cmd.Aliases {
			if input == alias {
				return cmd.Name
			}
		}
	}
	return ""
}

// getCommandCompletions 根据前缀返回匹配的命令列表
func getCommandCompletions(prefix string) []Command {
	if prefix == "" {
		return commands
	}
	var result []Command
	for _, cmd := range commands {
		if strings.HasPrefix(cmd.Name, prefix) {
			result = append(result, cmd)
			continue
		}
		for _, alias := range cmd.Aliases {
			if strings.HasPrefix(alias, prefix) {
				result = append(result, cmd)
				break
			}
		}
	}
	return result
}

// KeyMap 定义快捷键
type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GoToTop      key.Binding
	GoToBottom   key.Binding
	Enter        key.Binding
	Space        key.Binding
	Search       key.Binding
	Edit         key.Binding
	New          key.Binding
	Delete       key.Binding
	Quit         key.Binding
	Help         key.Binding
	// 折叠相关
	ToggleFold    key.Binding
	OpenFold      key.Binding
	CloseFold     key.Binding
	OpenAllFolds  key.Binding
	CloseAllFolds key.Binding
}

// DefaultKeyMap 返回默认快捷键配置
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+b"),
			key.WithHelp("PgUp/C-b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+f"),
			key.WithHelp("PgDn/C-f", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("C-u", "half page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("C-d", "half page down"),
		),
		GoToTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("Home/g", "top"),
		),
		GoToBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("End/G", "bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "connect"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c/:q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		// 折叠快捷键
		ToggleFold: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "toggle fold"),
		),
		OpenFold: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "open fold"),
		),
		CloseFold: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "close fold"),
		),
		OpenAllFolds: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "expand all"),
		),
		CloseAllFolds: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "collapse all"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the help.KeyMap interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// help.KeyMap interface.
// 注意：实际帮助渲染使用 renderHelp() 方法
func (k KeyMap) FullHelp() [][]key.Binding {
	return nil
}

// AgentKeyCache SSH Agent keys 缓存
type AgentKeyCache struct {
	keys      []internalssh.AgentKeyInfo
	err       error
	timestamp int64
}

// Model 是 TUI 的模型
type Model struct {
	keys          KeyMap
	help          help.Model
	tree          *session.SessionNode
	cursor        int
	width         int
	height        int
	sessionsDir   string
	searchInput   textinput.Model
	searchMode    bool
	searchQuery   string
	lineNumInput  textinput.Model
	lineNumMode   bool
	lineNumBuffer string
	detailView    viewport.Model
	showHelp      bool
	showError     bool
	errorMessage  string
	agentKeyCache *AgentKeyCache
	lastKeyG      bool // 用于检测 'gg' 快捷键
	showPassword  bool // 是否显示密码明文，默认隐藏
}

// 初始化 Model
func initialModel() Model {
	keys := DefaultKeyMap()

	// 初始化搜索输入框
	searchInput := textinput.New()
	searchInput.Placeholder = "Search..."
	searchInput.Prompt = "/"
	searchInput.CharLimit = 50
	searchInput.Width = 30

	// 初始化行号输入框
	lineNumInput := textinput.New()
	lineNumInput.Placeholder = ""
	lineNumInput.Prompt = ":"
	lineNumInput.CharLimit = 20
	lineNumInput.Width = 20

	return Model{
		keys:         keys,
		help:         help.New(),
		searchInput:  searchInput,
		lineNumInput: lineNumInput,
	}
}

// Init 初始化 Bubble Tea 程序
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadSessions(),
		tea.EnterAltScreen,
	)
}

// Update 处理消息
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.detailView.Width = m.width * 30 / 100
		m.detailView.Height = m.height - 3
		return m, nil

	case tea.KeyMsg:
		// 如果显示错误信息，按任意键关闭
		if m.showError {
			m.showError = false
			m.errorMessage = ""
			return m, nil
		}

		// 如果显示帮助，按任意键关闭帮助（除了 q/Ctrl+c 仍然退出）
		if m.showHelp {
			if key.Matches(msg, m.keys.Quit) {
				return m, tea.Quit
			}
			m.showHelp = false
			return m, nil
		}

		if m.searchMode {
			// 直接处理 ESC 键，避免被 textinput 拦截
			if msg.Type == tea.KeyEsc {
				m.searchMode = false
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.cursor = 0
				return m, nil
			}
			return m.handleSearchInput(msg)
		}

		if m.lineNumMode {
			return m.handleLineNumInput(msg)
		}

		// 普通模式下，Esc 清空搜索过滤（如果有过滤条件）
		if msg.Type == tea.KeyEsc && m.searchQuery != "" {
			m.searchQuery = ""
			m.searchInput.SetValue("")
			m.cursor = 0
			return m, nil
		}

		// 处理键盘输入
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.Up):
			m.moveCursor(-1)
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.Down):
			m.moveCursor(1)
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.PageUp):
			m.moveCursor(-(m.height - 3))
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.PageDown):
			m.moveCursor(m.height - 3)
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.HalfPageUp):
			m.moveCursor(-((m.height - 3) / 2))
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.HalfPageDown):
			m.moveCursor((m.height - 3) / 2)
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		// Vim: gg - 跳转到顶部（或者 g 后面跟 g）
		case msg.String() == "g":
			// 检测是否是 'gg' 组合
			if m.lastKeyG {
				m.cursor = 0
				m.lastKeyG = false
				return m, nil
			}
			m.lastKeyG = true
			return m, nil

		// Vim: G - 跳转到底部，或者数字+G跳转到指定行
		case msg.String() == "G":
			visibleNodes := m.getVisibleNodes()
			if m.lineNumBuffer != "" {
				// 如果有累积的数字，跳转到指定行
				var lineNum int
				fmt.Sscanf(m.lineNumBuffer, "%d", &lineNum)
				if lineNum > 0 && len(visibleNodes) > 0 {
					m.cursor = lineNum - 1
					if m.cursor >= len(visibleNodes) {
						m.cursor = len(visibleNodes) - 1
					}
					if m.cursor < 0 {
						m.cursor = 0
					}
				}
				m.lineNumBuffer = ""
			} else {
				// 没有数字，跳转到底部
				if len(visibleNodes) > 0 {
					m.cursor = len(visibleNodes) - 1
				}
			}
			m.lastKeyG = false
			return m, nil

		// Vim: 0 - 跳转到行首（对于列表，跳到顶部）
		case msg.String() == "0":
			m.cursor = 0
			m.lastKeyG = false
			return m, nil

		// Vim: $ - 跳转到行尾（对于列表，跳到底部）
		case msg.String() == "$":
			visibleNodes := m.getVisibleNodes()
			if len(visibleNodes) > 0 {
				m.cursor = len(visibleNodes) - 1
			}
			m.lastKeyG = false
			return m, nil

		// Vim: ^ - 跳转到第一个非空字符（对于树形列表，跳到第一个文件）
		case msg.String() == "^":
			visibleNodes := m.getVisibleNodes()
			for i, node := range visibleNodes {
				if !node.IsDir {
					m.cursor = i
					break
				}
			}
			m.lastKeyG = false
			return m, nil

		// Vim: n - 有搜索时查找下一个，无搜索时新建会话
		case msg.String() == "n":
			if m.searchQuery != "" {
				m.searchNext(1)
				m.lastKeyG = false
				return m, nil
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, m.newSession()

		// Vim: N - 查找上一个
		case msg.String() == "N":
			if m.searchQuery != "" {
				m.searchNext(-1)
			}
			m.lastKeyG = false
			return m, nil

		// Vim: : - 进入行号跳转模式
		case msg.String() == ":":
			m.lineNumMode = true
			m.lineNumBuffer = ""
			m.lineNumInput.SetValue("")
			m.lineNumInput.Focus()
			m.lastKeyG = false
			return m, textinput.Blink

		// 数字键 - 可能是在输入行号
		case len(msg.String()) == 1 && msg.String()[0] >= '1' && msg.String()[0] <= '9':
			// 开始累积数字
			m.lineNumBuffer += msg.String()
			// 如果输入了数字后按 G，会在下一个 case 处理
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.GoToTop):
			m.cursor = 0
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.GoToBottom):
			visibleNodes := m.getVisibleNodes()
			if len(visibleNodes) > 0 {
				m.cursor = len(visibleNodes) - 1
			}
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.Space):
			selected := m.getSelectedNode()
			if selected != nil && selected.IsDir {
				selected.Expanded = !selected.Expanded
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			selected := m.getSelectedNode()
			if selected != nil && !selected.IsDir && selected.Session != nil && selected.Session.Valid {
				// 使用 execCommand 执行外部命令，确保完全退出 TUI 后再连接
				return m, m.execSSHCommand(selected.Session)
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.Search):
			m.searchMode = true
			m.searchInput.Focus()
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, textinput.Blink

		case key.Matches(msg, m.keys.Edit):
			selected := m.getSelectedNode()
			if selected != nil && !selected.IsDir && selected.Session != nil {
				return m, m.execEditCommand(selected.Session)
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		case key.Matches(msg, m.keys.Delete):
			selected := m.getSelectedNode()
			if selected != nil && !selected.IsDir {
				return m, m.deleteSession(selected)
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		// Vim: o - Toggle fold (展开/折叠当前目录)
		case msg.String() == "o":
			selected := m.getSelectedNode()
			if selected != nil && selected.IsDir {
				selected.Expanded = !selected.Expanded
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		// Vim: h/← - 折叠当前目录（如果已展开）或跳到父目录
		case key.Matches(msg, m.keys.CloseFold):
			selected := m.getSelectedNode()
			if selected != nil {
				if selected.IsDir && selected.Expanded {
					selected.Expanded = false
				} else if selected.Parent != nil {
					// 查找父目录在可见列表中的位置
					visibleNodes := m.getVisibleNodes()
					for i, node := range visibleNodes {
						if node == selected.Parent {
							m.cursor = i
							break
						}
					}
				}
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		// Vim: l/→ - 展开当前目录（如果已折叠）
		case key.Matches(msg, m.keys.OpenFold):
			selected := m.getSelectedNode()
			if selected != nil && selected.IsDir && !selected.Expanded {
				selected.Expanded = true
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		// Vim: E - 展开所有目录
		case key.Matches(msg, m.keys.OpenAllFolds):
			if m.tree != nil {
				m.expandAll(m.tree)
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil

		// Vim: C - 折叠所有目录
		case key.Matches(msg, m.keys.CloseAllFolds):
			if m.tree != nil {
				m.collapseAll(m.tree)
			}
			m.lineNumBuffer = ""
			m.lastKeyG = false
			return m, nil
		}

	case sessionsLoadedMsg:
		m.tree = msg.tree
		m.sessionsDir = msg.sessionsDir
		if m.tree != nil {
			m.tree.SetParent(nil)
			// 默认展开所有目录
			m.expandAll(m.tree)
		}
		return m, func() tea.Msg {
			// 触发一次刷新以确保界面正确渲染
			if m.width > 0 && m.height > 0 {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
			return nil
		}

	case connectCompleteMsg:
		// SSH 连接完成，自动返回 TUI
		cmds := []tea.Cmd{
			tea.EnterAltScreen,
			func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			},
		}
		if msg.err != nil {
			cmds = append(cmds, func() tea.Msg {
				return showErrorMsg{err: msg.err}
			})
		}
		return m, tea.Batch(cmds...)

	case showErrorMsg:
		// 在TUI界面中显示错误信息
		m.errorMessage = fmt.Sprintf("Connection failed: %v", msg.err)
		m.showError = true
		return m, nil

	case editorCompleteMsg:
		// 编辑器关闭，重新加载会话
		return m, tea.Batch(
			tea.EnterAltScreen,
			m.loadSessions(),
			func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			},
		)
	}

	var cmd tea.Cmd
	m.detailView, cmd = m.detailView.Update(msg)
	return m, cmd
}

// View 渲染界面
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	if m.showError {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fb4934")).
			Background(lipgloss.Color("#282828")).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#fb4934"))
		return errorStyle.Render(m.errorMessage + "\n\nPress any key to continue...")
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// 计算布局
	treeWidth := m.width * 70 / 100
	detailWidth := m.width * 30 / 100
	contentHeight := m.height - 2 // 留出状态栏空间

	// 构建树形视图
	treeView := m.renderTree(treeWidth, contentHeight)

	// 构建详情视图
	detailView := m.renderDetail(detailWidth, contentHeight)

	// 合并主内容区
	content := lipgloss.JoinHorizontal(lipgloss.Top, treeView, detailView)

	// 构建状态栏
	statusBar := m.renderStatusBar()

	// 合并所有内容
	if m.searchMode {
		searchBar := m.renderSearchBar()
		return lipgloss.JoinVertical(lipgloss.Left, content, statusBar, searchBar)
	}

	if m.lineNumMode {
		lineNumBar := m.renderLineNumBar()
		return lipgloss.JoinVertical(lipgloss.Left, content, statusBar, lineNumBar)
	}

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

// renderTree 渲染树形视图
func (m Model) renderTree(width, height int) string {
	if m.tree == nil {
		return treeStyle.Width(width).Height(height).Render("Loading sessions...")
	}

	visibleNodes := m.getVisibleNodes()
	totalNodes := len(visibleNodes)

	if totalNodes == 0 {
		return treeStyle.Width(width).Height(height).Render("No sessions found")
	}

	// 计算滚动的起始位置，确保光标在可视区域内
	startIdx := 0
	if m.cursor >= height {
		startIdx = m.cursor - height + 1
	}
	// 如果光标靠近底部，调整起始位置
	if totalNodes > height && m.cursor > height/2 {
		startIdx = min(m.cursor-height/2, totalNodes-height)
	}

	endIdx := min(startIdx+height, totalNodes)

	// 计算行号宽度（根据总节点数的位数）
	lineNumWidth := len(fmt.Sprintf("%d", totalNodes))
	if lineNumWidth < 3 {
		lineNumWidth = 3
	}

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		nodeLine := m.renderNode(visibleNodes[i], i == m.cursor)
		// 添加行号前缀
		lineNum := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#665c54")).
			Width(lineNumWidth).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%d", i+1))
		line := lineNum + " " + nodeLine
		lines = append(lines, line)
	}

	// 填充空行保持高度
	for len(lines) < height {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return treeStyle.Width(width).Height(height).Render(content)
}

// renderNode 渲染单个节点
func (m Model) renderNode(node *session.SessionNode, selected bool) string {
	indent := m.getIndent(node)

	var icon string
	var name string

	if node.IsDir {
		if node.Expanded {
			icon = "▾ "
		} else {
			icon = "▸ "
		}
		name = folderStyle.Render(node.Name + "/")
	} else {
		icon = "  "
		if node.Session != nil && !node.Session.Valid {
			name = invalidStyle.Render(node.Name + " [invalid]")
		} else {
			name = fileStyle.Render(node.Name)
		}
	}

	line := indent + icon + name

	if selected {
		return selectedStyle.Render(line)
	}
	return line
}

// getIndent 获取节点的缩进
func (m Model) getIndent(node *session.SessionNode) string {
	depth := 0
	current := node
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return strings.Repeat("  ", depth)
}

// renderDetail 渲染详情视图
func (m Model) renderDetail(width, height int) string {
	selected := m.getSelectedNode()
	if selected == nil {
		return detailBoxStyle.
			Width(width - 4).
			Height(height - 2).
			Render("No session selected")
	}

	if selected.IsDir {
		content := fmt.Sprintf("Folder: %s\n\nContains %d items",
			selected.Name, len(selected.Children))
		return detailBoxStyle.
			Width(width - 4).
			Height(height - 2).
			Render(content)
	}

	if selected.Session == nil {
		return detailBoxStyle.
			Width(width - 4).
			Height(height - 2).
			Render("No session data")
	}

	s := selected.Session
	var content strings.Builder

	// 标题 - 显示节点文件名（不含后缀）
	content.WriteString(detailTitleStyle.Render(selected.Name))
	content.WriteString("\n\n")

	// 配置详情
	content.WriteString(detailKeyStyle.Render("Host: "))
	content.WriteString(detailValueStyle.Render(s.Host) + "\n\n")

	content.WriteString(detailKeyStyle.Render("Port: "))
	content.WriteString(detailValueStyle.Render(fmt.Sprintf("%d", s.Port)) + "\n\n")

	content.WriteString(detailKeyStyle.Render("User: "))
	content.WriteString(detailValueStyle.Render(s.User) + "\n\n")

	content.WriteString(detailKeyStyle.Render("Auth: "))
	content.WriteString(detailValueStyle.Render(string(s.AuthType)) + "\n\n")

	// 根据认证类型显示详细信息
	switch s.AuthType {
	case session.AuthTypePassword:
		if s.Password != "" {
			content.WriteString(detailKeyStyle.Render("Password: "))
			if m.showPassword {
				content.WriteString(detailValueStyle.Render(s.Password) + "\n\n")
			} else {
				content.WriteString(detailValueStyle.Render("********") + "\n\n")
			}
		} else if s.EncryptedPassword != "" {
			content.WriteString(detailKeyStyle.Render("Password: "))
			if m.showPassword {
				// 仅在显示密码时才解密
				if err := s.ResolvePassword(); err == nil {
					content.WriteString(detailValueStyle.Render(s.Password) + "\n\n")
				} else {
					content.WriteString(invalidStyle.Render(fmt.Sprintf("(decrypt failed: %v)", err)) + "\n\n")
				}
			} else {
				content.WriteString(detailValueStyle.Render("********") + "\n\n")
			}
		}
	case session.AuthTypeKey:
		if s.KeyPath != "" {
			content.WriteString(detailKeyStyle.Render("Key: "))
			content.WriteString(detailValueStyle.Render(s.KeyPath) + "\n\n")
		}
	case session.AuthTypeAgent:
		content.WriteString(detailKeyStyle.Render("SSH Agent Keys:\n"))
		// 使用缓存的 SSH Agent keys
		var keys []internalssh.AgentKeyInfo
		var err error
		if m.agentKeyCache != nil {
			keys = m.agentKeyCache.keys
			err = m.agentKeyCache.err
		} else {
			keys, err = internalssh.ListAgentKeys()
			m.agentKeyCache = &AgentKeyCache{
				keys: keys,
				err:  err,
			}
		}
		if err != nil {
			content.WriteString(invalidStyle.Render("  "+err.Error()) + "\n\n")
		} else if len(keys) == 0 {
			content.WriteString(detailValueStyle.Render("  (no keys loaded)") + "\n\n")
		} else {
			for _, k := range keys {
				comment := k.Comment
				if comment == "" {
					comment = "(no comment)"
				}
				content.WriteString(detailValueStyle.Render(
					fmt.Sprintf("  %s %s", k.Type, comment)) + "\n")
			}
			content.WriteString("\n")
		}
	}

	if s.Description != "" {
		content.WriteString(detailKeyStyle.Render("Description:\n"))
		content.WriteString(s.Description + "\n\n")
	}

	if !s.Valid {
		content.WriteString(invalidStyle.Render("Error: " + s.Error.Error()))
	}

	// 应用边框样式
	return detailBoxStyle.
		Width(width - 4).   // 减去边框和padding的宽度
		Height(height - 2). // 减去边框的高度
		Render(content.String())
}

// renderStatusBar 渲染状态栏
func (m Model) renderStatusBar() string {
	var status strings.Builder

	if m.searchMode {
		status.WriteString("Search mode | ")
	}

	selected := m.getSelectedNode()
	if selected != nil && !selected.IsDir {
		status.WriteString(fmt.Sprintf("Session: %s | ", selected.Name))
	}

	// 显示搜索状态
	if m.searchQuery != "" {
		status.WriteString(fmt.Sprintf("Filter: '%s' (%d) | ", m.searchQuery, len(m.getVisibleNodes())))
		status.WriteString("Esc:clear Enter:confirm | ")
	} else {
		status.WriteString(fmt.Sprintf("Total: %d | ", len(m.getVisibleNodes())))
	}
	if m.showPassword {
		status.WriteString("[PW] ")
	}
	status.WriteString("Press ? for help, :q or Ctrl+c to quit")

	return statusBarStyle.Width(m.width).Render(status.String())
}

// renderSearchBar 渲染搜索栏
func (m Model) renderSearchBar() string {
	// 添加退出提示到搜索栏
	searchWithHint := m.searchInput.View() + "  (Esc:clear Enter:confirm)"
	return searchStyle.Width(m.width).Render(searchWithHint)
}

// renderLineNumBar 渲染行号跳转栏（带命令补全提示）
func (m Model) renderLineNumBar() string {
	input := m.lineNumInput.Value()
	completions := getCommandCompletions(input)

	var hints []string
	for i, cmd := range completions {
		hint := fmt.Sprintf(":%s - %s", cmd.Name, cmd.Description)
		if i == 0 {
			hints = append(hints, cmdHintActiveStyle.Render(hint))
		} else {
			hints = append(hints, cmdHintStyle.Render(hint))
		}
	}

	bar := m.lineNumInput.View()
	if len(hints) > 0 {
		bar += "  " + strings.Join(hints, "  ")
	}
	bar += "  " + cmdHintStyle.Render("(Tab:补全 Enter:执行 Esc:取消)")

	return searchStyle.Width(m.width).Render(bar)
}

// renderHelp 渲染自定义帮助视图
func (m Model) renderHelp() string {
	var b strings.Builder

	renderSection := func(title string, items [][2]string) {
		b.WriteString(helpSectionStyle.Render(title))
		b.WriteString("\n")
		for _, item := range items {
			b.WriteString("  ")
			b.WriteString(helpKeyStyle.Render(item[0]))
			b.WriteString(helpDescStyle.Render(item[1]))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	renderSection("移动", [][2]string{
		{"↑/k, ↓/j", "上下移动"},
		{"PgUp/C-b", "向上翻页"},
		{"PgDn/C-f", "向下翻页"},
		{"C-u, C-d", "向上/下半页"},
		{"gg", "跳转到顶部"},
		{"G", "跳转到底部"},
		{"<n>G, :<n>", "跳转到第 n 行"},
		{"0", "跳转到第一行"},
		{"$", "跳转到最后一行"},
		{"^", "跳转到第一个会话"},
	})

	renderSection("折叠", [][2]string{
		{"Space/o", "展开/折叠目录"},
		{"h/←", "折叠目录或跳到父目录"},
		{"l/→", "展开目录"},
		{"E", "展开所有目录"},
		{"C", "折叠所有目录"},
	})

	renderSection("搜索", [][2]string{
		{"/", "进入搜索模式"},
		{"Enter", "确认搜索"},
		{"Esc", "取消搜索并清除过滤"},
		{"Ctrl+c", "退出搜索并保留过滤"},
		{"n/N", "下一个/上一个匹配"},
	})

	renderSection("会话操作", [][2]string{
		{"Enter", "连接到选中会话"},
		{"e", "编辑会话配置"},
		{"n", "新建会话"},
		{"D", "删除会话"},
	})

	// 从命令注册表自动生成命令部分
	cmdItems := make([][2]string, len(commands))
	for i, cmd := range commands {
		aliases := strings.Join(cmd.Aliases, "/")
		cmdItems[i] = [2]string{
			fmt.Sprintf(":%s/:%s", cmd.Name, aliases),
			cmd.Description,
		}
	}
	renderSection("命令 (: 模式)", cmdItems)

	renderSection("其他", [][2]string{
		{"?", "显示/关闭帮助"},
		{"Ctrl+c/:q", "退出程序"},
	})

	return helpContainerStyle.Render(b.String())
}

// handleSearchInput 处理搜索输入
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// ESC: 取消搜索，清空过滤条件
		m.searchMode = false
		m.searchQuery = ""
		m.searchInput.SetValue("")
		// 重置光标到顶部，避免光标位置超出新的可见节点范围
		m.cursor = 0
		return m, nil

	case tea.KeyCtrlC:
		// Ctrl+c: 取消搜索但保留过滤结果（仅退出输入模式）
		m.searchMode = false
		m.searchQuery = m.searchInput.Value()
		return m, nil

	case tea.KeyEnter:
		// Enter: 确认搜索
		m.searchMode = false
		m.searchQuery = m.searchInput.Value()
		return m, nil

	case tea.KeyCtrlU:
		// Ctrl+u: 清空当前输入（Vim 风格）
		m.searchInput.SetValue("")
		m.searchQuery = ""
		return m, nil

	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchQuery = m.searchInput.Value()
		return m, cmd
	}
}

// moveCursor 移动光标
func (m *Model) moveCursor(delta int) {
	visibleNodes := m.getVisibleNodes()
	if len(visibleNodes) == 0 {
		return
	}

	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(visibleNodes) {
		m.cursor = len(visibleNodes) - 1
	}
}

// getSelectedNode 获取当前选中的节点
func (m Model) getSelectedNode() *session.SessionNode {
	visibleNodes := m.getVisibleNodes()
	if m.cursor >= 0 && m.cursor < len(visibleNodes) {
		return visibleNodes[m.cursor]
	}
	return nil
}

// getVisibleNodes 获取可见节点列表（根据搜索查询过滤）
func (m Model) getVisibleNodes() []*session.SessionNode {
	if m.tree == nil {
		return nil
	}

	allNodes := m.tree.FlattenVisible()

	// 如果有搜索查询，过滤节点
	if m.searchQuery != "" {
		query := strings.ToLower(m.searchQuery)
		var filtered []*session.SessionNode
		for _, node := range allNodes {
			if strings.Contains(strings.ToLower(node.Name), query) {
				filtered = append(filtered, node)
			}
		}
		return filtered
	}

	return allNodes
}

// expandAll 展开所有目录
func (m Model) expandAll(node *session.SessionNode) {
	if node.IsDir {
		node.Expanded = true
		for _, child := range node.Children {
			m.expandAll(child)
		}
	}
}

// collapseAll 折叠所有目录
func (m Model) collapseAll(node *session.SessionNode) {
	if node.IsDir {
		node.Expanded = false
		for _, child := range node.Children {
			m.collapseAll(child)
		}
	}
}

// 消息类型
type connectCompleteMsg struct {
	err error
}

// showErrorMsg 显示错误消息
type showErrorMsg struct {
	err error
}

// sessionsLoadedMsg 会话加载完成消息
type sessionsLoadedMsg struct {
	tree        *session.SessionNode
	sessionsDir string
}

// loadSessions 加载会话
func (m *Model) loadSessions() tea.Cmd {
	return func() tea.Msg {
		sessionsDir, err := config.GetSessionsDir()
		if err != nil {
			return sessionsLoadedMsg{tree: nil}
		}

		tree, err := session.LoadSessionsTree(sessionsDir)
		if err != nil {
			return sessionsLoadedMsg{tree: nil}
		}

		// 加载全局配置
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			return sessionsLoadedMsg{tree: tree, sessionsDir: sessionsDir}
		}

		// 如果启用了 SecureCRT，加载 SecureCRT 会话
		if globalConfig.SecureCRT.Enabled {
			scTree, err := session.LoadSecureCRTSessions(globalConfig.SecureCRT)
			if err == nil && scTree != nil {
				// 将 SecureCRT 会话作为子树添加到本地会话树
				tree.Children = append(tree.Children, scTree)
			}
		}

		return sessionsLoadedMsg{tree: tree, sessionsDir: sessionsDir}
	}
}

// newSession 创建新会话
func (m Model) newSession() tea.Cmd {
	return func() tea.Msg {
		selected := m.getSelectedNode()
		var dir string

		if selected != nil {
			if selected.IsDir {
				dir = filepath.Join(m.sessionsDir, selected.GetPath())
			} else if selected.Parent != nil {
				dir = filepath.Join(m.sessionsDir, selected.Parent.GetPath())
			}
		}

		if dir == "" {
			dir = m.sessionsDir
		}

		// 创建模板文件
		templatePath := filepath.Join(dir, "new-session.yaml")
		template := &session.Session{
			Host:     "example.com",
			Port:     22,
			User:     "root",
			AuthType: "agent",
		}

		if err := session.SaveSession(template, templatePath); err != nil {
			return nil
		}

		// 加载刚创建的会话并打开编辑器
		newSession, err := session.LoadSession(templatePath)
		if err != nil {
			return nil
		}

		// 使用 execEditCommand 打开编辑器
		return m.execEditCommand(newSession)()
	}
}

// deleteSession 删除会话
func (m Model) deleteSession(node *session.SessionNode) tea.Cmd {
	return func() tea.Msg {
		if node.Session == nil {
			return nil
		}

		err := os.Remove(node.Session.FilePath)
		if err != nil {
			return nil
		}

		return m.loadSessions()()
	}
}

// sshProcess 实现 tea.ExecCommand 接口，使用纯 Go 建立 SSH 连接
type sshProcess struct {
	session *session.Session
}

func (p sshProcess) Run() error {
	return internalssh.Connect(p.session)
}

func (p sshProcess) SetStdin(r io.Reader)  {}
func (p sshProcess) SetStdout(w io.Writer) {}
func (p sshProcess) SetStderr(w io.Writer) {}

// execSSHCommand 通过 tea.Exec 执行 SSH 连接
// tea.Exec 会让 Bubble Tea 正确暂停 TUI 并恢复终端到正常状态，
// 然后执行 SSH 连接，结束后重新进入 TUI
func (m Model) execSSHCommand(s *session.Session) tea.Cmd {
	return tea.Exec(sshProcess{session: s}, func(err error) tea.Msg {
		return connectCompleteMsg{err: err}
	})
}

// execEditCommand 执行编辑命令，确保 TUI 完全退出
func (m Model) execEditCommand(s *session.Session) tea.Cmd {
	return tea.Exec(editorProcess{filepath: s.FilePath}, func(err error) tea.Msg {
		return editorCompleteMsg{err: err}
	})
}

// editorProcess 实现 tea.Exec 接口
type editorProcess struct {
	filepath string
}

func (p editorProcess) Run() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, p.filepath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p editorProcess) SetStdin(r io.Reader)  {}
func (p editorProcess) SetStdout(w io.Writer) {}
func (p editorProcess) SetStderr(w io.Writer) {}

// editorCompleteMsg 编辑器完成消息
type editorCompleteMsg struct {
	err error
}

// Run 启动 TUI
func Run() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleLineNumInput 处理行号输入
func (m Model) handleLineNumInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.lineNumMode = false
		m.lineNumBuffer = ""
		m.lineNumInput.SetValue("")
		return m, nil

	case tea.KeyTab:
		// Tab 自动补全：匹配第一个命令
		input := m.lineNumInput.Value()
		completions := getCommandCompletions(input)
		if len(completions) > 0 {
			m.lineNumInput.SetValue(completions[0].Name)
			m.lineNumInput.CursorEnd()
		}
		return m, nil

	case tea.KeyEnter:
		m.lineNumMode = false
		cmdStr := m.lineNumInput.Value()
		if cmdStr == "" {
			cmdStr = m.lineNumBuffer
		}

		// 通过命令注册表匹配命令
		switch matchCommand(cmdStr) {
		case "q":
			return m, tea.Quit
		case "noh":
			m.searchQuery = ""
			m.searchInput.SetValue("")
			m.lineNumBuffer = ""
			m.lineNumInput.SetValue("")
			return m, nil
		case "pw":
			m.showPassword = !m.showPassword
			m.lineNumBuffer = ""
			m.lineNumInput.SetValue("")
			return m, nil
		}

		// 未匹配命令，尝试解析行号并跳转
		if cmdStr != "" {
			var lineNum int
			fmt.Sscanf(cmdStr, "%d", &lineNum)
			if lineNum > 0 {
				m.cursor = lineNum - 1
				visibleNodes := m.getVisibleNodes()
				if m.cursor >= len(visibleNodes) {
					m.cursor = len(visibleNodes) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
			}
		}
		m.lineNumBuffer = ""
		m.lineNumInput.SetValue("")
		return m, nil

	default:
		var cmd tea.Cmd
		m.lineNumInput, cmd = m.lineNumInput.Update(msg)
		return m, cmd
	}
}

// searchNext 查找下一个/上一个匹配项
func (m *Model) searchNext(direction int) {
	if m.searchQuery == "" {
		return
	}

	visibleNodes := m.getVisibleNodes()
	if len(visibleNodes) == 0 {
		return
	}

	query := strings.ToLower(m.searchQuery)
	startIdx := m.cursor

	// 从当前位置开始搜索
	for i := 1; i <= len(visibleNodes); i++ {
		idx := startIdx + (i * direction)

		// 循环搜索
		if idx >= len(visibleNodes) {
			idx = idx - len(visibleNodes)
		} else if idx < 0 {
			idx = idx + len(visibleNodes)
		}

		if strings.Contains(strings.ToLower(visibleNodes[idx].Name), query) {
			m.cursor = idx
			return
		}
	}
}
