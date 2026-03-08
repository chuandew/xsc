package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

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
	Rename       key.Binding
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
		Rename: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "rename"),
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
