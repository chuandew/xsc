package xftp

// PanelSide 标识面板方向
type PanelSide int

const (
	PanelLeft  PanelSide = iota // 本地面板
	PanelRight                  // 远程面板
)

// 连接相关消息

// ConnectedMsg 连接成功
type ConnectedMsg struct {
	RemoteFS *RemoteFS
}

// ConnectErrMsg 连接失败
type ConnectErrMsg struct {
	Err error
}

// DisconnectedMsg 连接断开
type DisconnectedMsg struct{}

// 目录加载相关消息

// DirLoadedMsg 目录加载完成
type DirLoadedMsg struct {
	Panel   PanelSide
	Entries []FileInfo
	Path    string
}

// DirLoadErrMsg 目录加载失败
type DirLoadErrMsg struct {
	Panel PanelSide
	Err   error
}

// 传输相关消息

// TransferProgressMsg 传输进度更新
type TransferProgressMsg struct {
	TaskID      int
	Progress    float64 // 0.0 ~ 1.0
	Speed       float64 // 字节/秒
	Transferred int64   // 已传输字节
}

// TransferCompleteMsg 传输完成
type TransferCompleteMsg struct {
	TaskID int
}

// TransferErrorMsg 传输失败
type TransferErrorMsg struct {
	TaskID int
	Err    error
}

// 文件操作相关消息

// FileOpCompleteMsg 文件操作完成
type FileOpCompleteMsg struct {
	Op string // 操作名称：mkdir, delete, rename, chmod
}

// FileOpErrorMsg 文件操作失败
type FileOpErrorMsg struct {
	Op  string
	Err error
}

// errorDismissMsg 错误提示自动消失
type errorDismissMsg struct{}

// reconnectedMsg 重连成功
type reconnectedMsg struct {
	RemoteFS *RemoteFS
}
