# xftp 设计文档（团队评审最终版）

## 概述

在 xsc 项目中新增 `xftp` 可执行文件，实现 TUI 版本的 SFTP 文件管理器，作为 SecureFX 的终端替代方案。完全复用 xsc 的 session 管理体系，将 SSH 登录行为替换为 SFTP 文件操作。

## 架构决策

| # | 决策 | 选择 | 理由 | 来源 |
|---|------|------|------|------|
| 1 | TUI 布局 | 经典双面板（左本地、右远程） | 类 SecureFX/WinSCP，最直观 | A+B 共识 |
| 2 | 启动流程 | 无参数显示帮助，`tui` 启动选择器 | 与 xsc 保持一致（用户确认） | 用户决策 |
| 3 | 键盘操作 | Vim 风格 | 与 xsc 一致 | A+B 共识 |
| 4 | 代码共享 | 直接 import session/ssh/config，独立 TUI | 不重构现有代码 | A+B 共识 |
| 5 | FileSystem 接口 | 精简版（浏览操作），传输层直接用 sftp.Client | 传输需进度回调，简单 io.Reader 不够 | A 提出，Reviewer 采纳 |
| 6 | FilePanel | 独立 Bubble Tea Model（有 Update/View） | 双面板各自有独立状态 | B 提出，Reviewer 采纳 |
| 7 | TUI 文件组织 | 多文件拆分（9个文件） | xftp 复杂度高于 xsc | A+B 共识 |
| 8 | 传输模式 | channel + tea.Cmd，MVP 串行传输 | Bubble Tea 最佳实践 | B 提出 |
| 9 | 断线处理 | 不自动重连，用户手动 :reconnect | 避免数据不一致 | A+B 共识 |
| 10 | 目录加载 | 懒加载（只加载当前一层） | 大目录性能 | B 提出 |
| 11 | SSH keepalive | 纳入 MVP | SFTP 场景长时间浏览需要 | 用户确认 |
| 12 | FindSession() | 从 xsc 抽取到 session 包 | xsc 和 xftp 共用 | 用户确认 |

## 目录结构

```
cmd/xftp/main.go              ← 新入口，命令调度
internal/xftp/
├── model.go                   ← Bubble Tea 主 Model + Init/Update/View 骨架
├── filepanel.go               ← FilePanel 子组件（独立 Model，本地/远程共用）
├── filesystem.go              ← FileSystem 接口 + LocalFS + RemoteFS 实现
├── transfer.go                ← 传输管理器（单文件上传/下载/进度/取消）
├── operations.go              ← 文件操作（删除/重命名/mkdir 等）
├── selector.go                ← Session 选择器（轻量实现，复用 session 树数据）
├── keymap.go                  ← Vim 风格快捷键映射
├── styles.go                  ← Gruvbox 样式定义
└── messages.go                ← 自定义 tea.Msg 类型定义
```

## TUI 布局

```
┌─ Local: /home/user/ ─────────────┬─ Remote: /var/www/ ────────────────┐
│ Name          Size    Modified   │ Name          Size    Perm  Modified│
│ ▾ documents/          Mar 01     │ ▾ html/              drwxr Mar 01  │
│   readme.md   4.2K    Mar 01     │   index.html  8.1K   -rw-r Mar 01  │
│   notes.txt   1.1K    Feb 28     │   style.css   2.3K   -rw-r Feb 28  │
│ ▸ downloads/          Feb 25     │ ▾ logs/              drwxr Feb 25  │
│ ▸ projects/           Feb 20     │   access.log  156M   -rw-r Mar 01  │
│   .bashrc     512B    Jan 15     │   error.log   23M    -rw-r Mar 01  │
│                                  │ ▸ config/            drwxr Feb 20  │
│                                  │   .htaccess   256B   -rw-r Jan 15  │
├─ [Tab] Switch Panel ─────────────┴────────────────────────────────────┤
│ ↑ readme.md → /var/www/html/     100%  ✓  4.2KB                      │
│ ↑ notes.txt → /var/www/html/      45%  [======>     ]  512B/1.1KB    │
│ Queue: 3 pending | Speed: 2.1 MB/s                                   │
├───────────────────────────────────────────────────────────────────────┤
│ prod-server (10.0.0.1:22) | user@ubuntu | 5/12 | ? Help  :q Quit    │
└───────────────────────────────────────────────────────────────────────┘
```

## 交互设计

### 面板导航（Vim 风格）

- `j/k` — 上下移动光标
- `h/l` — 折叠/展开目录
- `gg/G` — 跳顶/跳底
- `Ctrl+d/u` — 半页滚动
- `Tab` — 切换本地/远程面板焦点
- `/` — 搜索过滤当前面板文件
- `Enter` — 进入目录（或展开）
- `Backspace` — 返回上级目录

### 文件操作（MVP）

- `y` — 标记（yank）选中文件用于传输
- `p` — 粘贴（传输到对面面板当前目录）
- `D` — 删除（带确认）
- `r` — 重命名
- `m` — 创建目录（mkdir）
- `Space` — 多选/取消选择

### 文件操作（后续版本）

- `V` — 进入批量选择模式
- `i` — 查看文件属性
- `P` — chmod/chown（远程文件）
- `e` — 用外部编辑器编辑文件
- `b` — 书签当前目录

### 传输队列

- `t` — 聚焦传输面板
- 显示进度条、速度、状态
- 支持取消传输

## 启动流程

```
xftp                        → 显示帮助信息（与 xsc 一致）
xftp tui                    → 启动 TUI（含 session 选择器）
xftp <session-path>         → 直连指定 session
xftp connect <session-path> → 同上（CLI 风格）
xftp help                   → 帮助信息
```

## 技术设计

### FileSystem 接口（精简版，仅用于浏览操作）

```go
type FileInfo struct {
    Name    string
    Size    int64
    Mode    os.FileMode
    ModTime time.Time
    IsDir   bool
    Owner   string      // 远程文件有效
    Group   string      // 远程文件有效
    LinkTarget string   // 符号链接目标
}

type FileSystem interface {
    ReadDir(path string) ([]FileInfo, error)
    Stat(path string) (*FileInfo, error)
    Mkdir(path string) error
    Remove(path string) error
    Rename(old, new string) error
    Chmod(path string, mode os.FileMode) error
    Getwd() (string, error)
}
```

传输操作不经过 FileSystem 接口，Transfer 层直接使用 `*sftp.Client` 和 `os` 包。

### SFTP 连接层

```go
type RemoteFS struct {
    client  *sftp.Client
    sshConn *ssh.Client
    cleanup func()
}

func NewRemoteFS(s *session.Session) (*RemoteFS, error) {
    sshClient, cleanup, err := sshclient.Dial(s)
    sftpClient, err := sftp.NewClient(sshClient)
    return &RemoteFS{client: sftpClient, sshConn: sshClient, cleanup: cleanup}, nil
}
```

### 传输管理器（channel + tea.Cmd 模式）

```go
type TransferTask struct {
    ID        int
    Source    string
    Dest      string
    Direction Direction  // Upload / Download
    Size      int64
    Status    TaskStatus // Pending / Active / Completed / Failed / Cancelled
    Progress  float64
    Speed     float64
    Error     error
}

// 进度监听（Bubble Tea 标准异步模式）
func listenProgress(ch <-chan ProgressMsg) tea.Cmd {
    return func() tea.Msg {
        return <-ch
    }
}
```

### 主 Model 结构

```go
type Model struct {
    localPanel   FilePanel
    remotePanel  FilePanel
    activePanel  PanelSide    // PanelLeft / PanelRight
    session      *session.Session
    remoteFS     *RemoteFS
    connected    bool
    transfer     *TransferManager
    mode         Mode  // Normal / Search / Command / Help / Error / Confirm / Selector
    width, height int
    keys         KeyMap
    statusMsg    string
}
```

### FilePanel 子组件

```go
type FilePanel struct {
    fs        FileSystem
    cwd       string
    entries   []FileEntry
    cursor    int
    offset    int
    selected  map[int]bool
    width, height int
    loading   bool
}

func (p FilePanel) Update(msg tea.Msg) (FilePanel, tea.Cmd)
func (p FilePanel) View() string
```

### 复用层

- `internal/session` — Session 结构、树、导入加载器、FindSession()（新增）
- `internal/ssh` — Dial() 建立 SSH 连接 + keepalive（新增）
- `internal/securecrt` / `internal/xshell` / `internal/mobaxterm` — 密码解密（间接复用）
- `pkg/config` — 全局配置

### 新依赖

- `github.com/pkg/sftp` — Go 社区标准 SFTP 客户端库

### Makefile 扩展

```makefile
BINARY_XFTP=xftp

build-xftp:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_XFTP) ./cmd/xftp

build-all: build build-xftp

install: build-all
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	cp $(BUILD_DIR)/$(BINARY_XFTP) $(INSTALL_DIR)/

run-xftp:
	go run ./cmd/xftp
```

## MVP 实施阶段

### Phase 1: 基础框架
- cmd/xftp/main.go 入口和命令调度
- FileSystem 接口 + LocalFS + RemoteFS
- SFTP 连接建立（复用 ssh.Dial）
- 基本 TUI 框架（双面板骨架）
- styles.go + keymap.go + messages.go
- Makefile 扩展

### Phase 2: 文件浏览
- 本地/远程目录列表和导航
- 目录进入/返回（Enter/Backspace）
- 文件属性列显示（大小、权限、日期）
- Vim 风格导航（j/k/gg/G/Ctrl+d/u）
- Tab 面板焦点切换
- 异步目录加载（Loading 指示）

### Phase 3: Session 选择器
- 内嵌 session 选择器（复用 session 树数据）
- CLI 参数直连（session.FindSession）
- 搜索过滤

### Phase 4: 文件传输
- 单文件上传/下载
- 传输进度显示（进度条 + 速度）
- 传输取消（context.WithCancel）

### Phase 5: 文件操作
- mkdir / delete（带确认）/ rename
- 确认对话框模式

### Phase 6: 帮助、状态栏和 keepalive
- 帮助视图
- 状态栏（连接信息、当前路径）
- 错误提示
- SSH keepalive
- :reconnect 命令
