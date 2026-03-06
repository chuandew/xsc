# XSC - XShell CLI

基于 Go 和 Bubble Tea 开发的 SSH 会话管理工具，支持本地 YAML 会话配置，以及导入 SecureCRT 和 Xshell 会话。

## 特性

- 🗂️ 文件即会话：通过 YAML 文件管理 SSH 配置，目录结构即分组层级
- 🖥️ 优雅的 TUI：使用 Bubble Tea 构建的交互式界面，Gruvbox 配色方案
- 🔍 实时搜索：输入即过滤，支持会话名称模糊匹配
- 🌳 树形结构：支持无限层级目录组织，Vim 风格折叠操作
- 🔐 多种认证：支持密码、密钥、SSH Agent、keyboard-interactive
- 📱 原生体验：使用 Go 原生 SSH 客户端连接，无需依赖外部 `ssh` 或 `sshpass`
- 📜 自动滚屏：列表内容超出屏幕时自动滚动保持光标可见
- 🔗 SecureCRT 集成：直接加载 SecureCRT 会话配置，支持加密密码解密和多认证方式
- 🔗 Xshell 集成：直接加载 Xshell `.xsh` 会话文件，支持 RC4 加密密码解密
- ⌨️ 命令自动补全：命令模式下支持 Tab 补全

## 快速开始

### 安装

```bash
# 克隆项目
git clone <repo-url>
cd xsc

# 构建
make build          # 输出到 ./build/xsc

# 安装到系统（需要 root 权限）
make install        # 安装到 /usr/local/bin/xsc
```

### 创建第一个会话

```bash
# 创建会话目录
mkdir -p ~/.xsc/sessions/prod

# 创建会话文件
cat > ~/.xsc/sessions/prod/my-server.yaml << 'EOF'
host: "192.168.1.100"
port: 22
user: "root"
auth_type: "password"
password: "your_password"
description: "生产服务器"
EOF

# 启动 TUI
xsc tui
```

### 基本用法

```bash
xsc                      # 显示帮助
xsc tui                  # 启动 TUI 交互界面
xsc list                 # 列出所有会话
xsc connect prod/my-server   # 直接连接指定会话
xsc connect web          # 模糊匹配连接（匹配名称包含 "web" 的会话）
xsc import-securecrt     # 将 SecureCRT 会话转换为 xsc 本地格式
xsc import-xshell        # 将 Xshell 会话转换为 xsc 本地格式
```

## 会话配置

会话文件存储在 `~/.xsc/sessions/`，使用 YAML 格式。文件系统中的目录结构直接映射为 TUI 中的树形层级。

### 配置字段说明

```yaml
host: "192.168.1.100"       # 必填，SSH 主机地址
port: 22                    # 可选，默认 22
user: "root"                # 可选，默认当前系统用户
auth_type: "password"       # 可选，默认 "agent"，可选值：password | key | agent
password: "my_secret"       # auth_type=password 时必填
key_path: "~/.ssh/id_rsa"   # auth_type=key 时必填，支持 ~ 展开
description: "生产数据库"    # 可选，会话描述信息
```

### 认证方式

| 认证方式 | auth_type | 必填字段 | 说明 |
|---------|-----------|---------|------|
| 密码认证 | `password` | `password` | 密码直接存储在配置文件中，连接时自动发送 |
| 密钥认证 | `key` | `key_path` | 使用 SSH 私钥文件，支持 ed25519/ecdsa/rsa/dsa |
| SSH Agent | `agent` | 无 | 使用系统 `ssh-agent`，需确保 Agent 已加载密钥 |

> 建议优先使用 **SSH Key** 或 **SSH Agent** 方式以获得更好的安全性。会话文件权限自动设为 0600（仅用户可读写）。

### 配置示例

**密码认证：**
```yaml
host: "10.0.1.50"
port: 22
user: "admin"
auth_type: "password"
password: "P@ssw0rd"
description: "内网跳板机"
```

**密钥认证：**
```yaml
host: "github-runner.internal"
port: 2222
user: "deploy"
auth_type: "key"
key_path: "~/.ssh/id_ed25519"
description: "CI/CD 部署服务器"
```

**SSH Agent 认证：**
```yaml
host: "bastion.example.com"
user: "ops"
auth_type: "agent"
description: "堡垒机"
```

## TUI 快捷键

### 导航 (Vim 风格)
| 按键 | 功能 |
|------|------|
| `↑/k` | 向上移动 |
| `↓/j` | 向下移动 |
| `PgUp/Ctrl+b` | 向上整页翻页 |
| `PgDn/Ctrl+f` | 向下整页翻页 |
| `Ctrl+u` | 向上半页翻页 |
| `Ctrl+d` | 向下半页翻页 |
| `gg` / `Home/g` | 跳转到第一行 |
| `G` / `End` | 跳转到最后行 |
| `nG` / `:n` | 跳转到第 n 行（如 `42G` 或 `:42`） |
| `0` | 跳转到第一行 |
| `$` | 跳转到最后行 |
| `^` | 跳转到第一个会话文件 |
| `n` | 查找下一个匹配 |
| `N` | 查找上一个匹配 |
| `:q` / `:quit` | 退出程序 |
| `Ctrl+c` | 退出程序 |

> 当列表内容超出屏幕时，会自动滚屏保持光标可见

### 操作
| 按键 | 功能 |
|------|------|
| `Enter` | 连接选中会话 |
| `Space` | 展开/折叠目录 |
| `/` | 进入搜索/过滤模式 |
| `:` | 进入命令模式 |
| `n` | 新建会话（在当前目录下） |
| `e` | 编辑选中会话（打开 `$EDITOR`） |
| `c` | 重命名会话 |
| `D` | 删除选中会话（需输入 YES 确认） |
| `?` | 显示快捷键帮助（任意键退出） |

> 导入的外部会话（SecureCRT / Xshell）为只读，不支持编辑、删除和重命名操作。

### 目录折叠 (Vim 风格)
| 按键 | 功能 |
|------|------|
| `o` | 展开/折叠当前目录 |
| `h/←` | 折叠当前目录或跳转到父目录 |
| `l/→` | 展开当前目录 |
| `E` | 展开所有目录 |
| `C` | 折叠所有目录 |

### 搜索模式
按 `/` 进入搜索模式，输入关键词可**实时过滤**会话列表：

**搜索时按键：**
| 按键 | 功能 |
|------|------|
| `Enter` | 确认搜索并退出搜索模式 |
| `Esc` | 取消搜索并**清空**过滤条件 |
| `Ctrl+c` | 退出搜索模式但**保留**过滤结果 |
| `Ctrl+u` | 清空当前输入内容（Vim 风格） |

**已确认搜索后（普通模式）：**
| 按键 | 功能 |
|------|------|
| `Esc` | 直接清空当前搜索过滤，显示全部会话 |

**搜索结果管理：**
- 底部状态栏显示 `Filter: '关键词' (匹配数)` 表示当前有过滤
- 使用 `:noh` 命令可清除过滤恢复显示全部会话
- 搜索是实时过滤，输入时即时显示匹配结果

### 命令模式 (:)
按 `:` 进入命令模式（类似 Vim 的命令行），支持 Tab 自动补全：

| 命令 | 功能 |
|------|------|
| `:q` / `:quit` | 退出程序 |
| `:noh` / `:nohlsearch` | 清除搜索过滤 |
| `:pw` | 切换密码明文/隐藏显示（状态栏显示 `[PW]` 标记） |
| `:<number>` | 跳转到第 n 行（如 `:42` 跳转到第 42 行） |

### 状态栏说明
TUI 底部状态栏显示当前状态信息：

- **`Session: xxx`** — 当前选中的会话名称（仅当选中会话时显示）
- **`Total: N`** — 当前可见节点总数（会话+目录）
- **`Filter: 'xxx' (N)`** — 搜索过滤状态（仅当有过滤时显示）
- **`[PW]`** — 密码明文显示已开启（通过 `:pw` 切换）
- **`Press ? for help, :q or Ctrl+c to quit`** — 操作提示

## SSH 连接行为

选中会话按 `Enter` 连接时：
- TUI 进入**暂停模式**，终端恢复到正常状态
- 使用 Go 原生 SSH 客户端建立连接，**连接超时 10 秒**
- SSH 会话中的所有快捷键都能正常使用
- 自动处理终端窗口大小调整（SIGWINCH）
- 按 `Ctrl+d` 退出 SSH 或连接断开后，TUI 自动恢复并可以继续操作
- 连接失败时显示红色边框错误弹窗，按任意键关闭

### 多认证方式回退

当会话配置了多种认证方式（常见于 SecureCRT 导入的会话）时，xsc 会按优先级顺序逐一尝试，直到成功或全部失败：

```
publickey → password → keyboard-interactive → agent
```

如果使用 publickey 认证且未指定密钥文件，会自动在 `~/.ssh/` 下查找默认密钥，按以下顺序尝试：
`id_ed25519` → `id_ecdsa` → `id_ecdsa_sk` → `id_ed25519_sk` → `id_rsa` → `id_dsa`

## SecureCRT 集成

xsc 支持直接加载 SecureCRT 的会话配置文件（`.ini` 格式），无需手动转换。启用后，SecureCRT 会话在 TUI 中以紫色样式显示在独立的 `[CRT] securecrt/` 目录下，与本地会话并列展示。

### 第一步：找到 SecureCRT 会话目录

SecureCRT 在不同平台的默认会话存储路径：

| 平台 | 默认路径 |
|------|---------|
| **macOS** | `~/Library/Application Support/VanDyke/SecureCRT/Config/Sessions` |
| **Linux** | `~/.vandyke/SecureCRT/Config/Sessions` |
| **Windows** | `%APPDATA%\VanDyke\Config\Sessions` |

你也可以在 SecureCRT 中通过菜单 **Options → Global Options → General → Configuration Paths** 查看实际路径。

### 第二步：配置 xsc

编辑 `~/.xsc/config.yaml`（首次运行 xsc 会自动创建 `~/.xsc/` 目录）：

```yaml
securecrt:
  enabled: true                                    # 启用 SecureCRT 集成
  session_path: "/home/user/.vandyke/SecureCRT/Config/Sessions"  # SecureCRT 会话目录路径
  password: "your_master_password"                 # SecureCRT 主密码（用于解密会话中的加密密码）
```

> **关于主密码**：如果你在 SecureCRT 中设置了配置加密密码（Options → Global Options → General → Use config encryption passphrase），需要在这里填入相同的密码。如果没有设置过主密码或不需要解密密码，可以留空，此时会话会以 SSH Agent 方式尝试连接。

### 第三步：启动 TUI

```bash
xsc tui
```

启用后，TUI 界面中会多出一个紫色的 `securecrt/` 目录，其中包含你所有的 SecureCRT 会话，保持原有的目录层级结构。

### SecureCRT 会话在 TUI 中的表现

| 特性 | 说明 |
|------|------|
| 显示样式 | 紫色文字，目录带 `[CRT]` 前缀，会话带 🔒 图标 |
| 只读模式 | 不支持编辑(e)、删除(D)、重命名(c) 操作 |
| 密码解密 | 延迟解密 — 仅在光标选中时实时解密当前会话密码，启动不卡顿 |
| 密码显示 | 默认隐藏，使用 `:pw` 命令切换明文/密文显示 |
| 多认证方式 | 详情面板显示所有认证方式及优先级顺序 |
| SSH 连接 | 按 Enter 直接连接，支持多认证方式自动回退 |

### 支持的认证方式

xsc 能解析 SecureCRT 会话中配置的以下认证方式，并在连接时按配置的优先级顺序尝试：

| 认证方式 | 图标 | 说明 |
|---------|------|------|
| Public Key | 🔐 | 使用公钥认证，支持指定密钥文件或自动发现 `~/.ssh/` 下的默认密钥 |
| Password | 🔑 | 使用加密密码，连接时自动解密 |
| Keyboard Interactive | ⌨️ | 键盘交互式认证 |
| GSSAPI | 🎫 | Kerberos/GSSAPI 认证 |
| SSH Agent | 🔐 | 使用本地 SSH Agent |

### 密码加密支持

xsc 支持解密 SecureCRT V2 格式的加密密码：

- **Prefix 02**：使用 SHA256(passphrase) 作为 AES-256-CBC 密钥
- **Prefix 03**：使用 bcrypt_pbkdf 派生 AES-256-CBC 密钥和 IV

解密后通过 SHA256 校验和验证密码完整性。V1 格式（SecureCRT 7.3.3 之前版本）暂不支持。

### 永久转换为 xsc 本地格式

如果你希望将 SecureCRT 会话永久转换为 xsc 本地 YAML 格式（不再依赖 SecureCRT 配置文件）：

```bash
xsc import-securecrt
```

转换后的会话保存在 `~/.xsc/sessions/securecrt-converted/YYYYMMDD-HHMMSS/` 目录下，保持原有目录结构。完成后会显示转换统计信息。

## Xshell 集成

xsc 支持直接加载 Xshell 的会话文件（`.xsh` 格式），无需手动转换。启用后，Xshell 会话在 TUI 中以青色样式显示在独立的 `[XSH] xshell/` 目录下，与本地会话和 SecureCRT 会话并列展示。

### 第一步：找到 Xshell 会话目录

Xshell 的会话文件默认存储路径：

| 平台 | 默认路径 |
|------|---------|
| **Windows** | `%APPDATA%\NetSarang Computer\6\Xshell\Sessions` |
| **Windows (旧版)** | `我的文档\NetSarang Computer\6\Xshell\Sessions` |

你也可以在 Xshell 中通过菜单 **工具 → 选项 → 常规** 查看会话文件的实际存储路径。

> 如果你的 Xshell 会话目录在 Windows 上，可以将整个 Sessions 目录拷贝到 Linux/macOS 上，xsc 能正确处理 UTF-16LE 编码的 `.xsh` 文件。

### 第二步：配置 xsc

编辑 `~/.xsc/config.yaml`：

```yaml
xshell:
  enabled: true                                      # 启用 Xshell 集成
  session_path: "/home/user/xshell-sessions"         # Xshell 会话目录路径（包含 .xsh 文件）
  password: "your_master_password"                   # Xshell 主密码（用于解密会话中的加密密码）
```

> **关于主密码**：如果你在 Xshell 中设置了主密码（工具 → 选项 → 安全 → 设置主密码），需要在这里填入相同的密码。如果没有设置过主密码，可以留空，此时有密码的会话将无法解密，会以 SSH Agent 方式尝试连接。

### 第三步：启动 TUI

```bash
xsc tui
```

启用后，TUI 界面中会多出一个青色的 `xshell/` 目录，其中包含你所有的 Xshell 会话，保持原有的目录层级结构。

### Xshell 会话在 TUI 中的表现

| 特性 | 说明 |
|------|------|
| 显示样式 | 青色文字，目录带 `[XSH]` 前缀，会话带 🔒 图标 |
| 只读模式 | 不支持编辑(e)、删除(D)、重命名(c) 操作 |
| 密码解密 | 延迟解密 — 仅在光标选中时实时解密当前会话密码 |
| 密码显示 | 默认隐藏，使用 `:pw` 命令切换明文/密文显示 |
| SSH 连接 | 按 Enter 直接连接 |

### Xshell 密码加密

xsc 支持解密 Xshell 使用主密码加密的会话密码：

- 加密格式：Base64 编码的密文 + SHA256 校验和
- 解密流程：Base64 解码 → SHA256(masterPassword) 生成密钥 → RC4 解密 → SHA256 校验验证

### Xshell 文件格式说明

Xshell 会话文件（`.xsh`）是 INI 格式，通常使用 UTF-16LE 编码（xsc 自动处理编码检测和转换）。主要解析以下段和字段：

```ini
[CONNECTION]
Host=192.168.1.100
Port=22

[CONNECTION:AUTHENTICATION]
UserName=root
Password=<base64 encoded encrypted password>
```

### 永久转换为 xsc 本地格式

如果你希望将 Xshell 会话永久转换为 xsc 本地 YAML 格式：

```bash
xsc import-xshell
```

转换后的会话保存在 `~/.xsc/sessions/xshell-converted/YYYYMMDD-HHMMSS/` 目录下，保持原有目录结构。完成后会显示转换统计信息。

## 全局配置

xsc 的全局配置文件为 `~/.xsc/config.yaml`，完整配置项：

```yaml
# SecureCRT 集成配置
securecrt:
  enabled: false                    # 是否启用 SecureCRT 集成
  session_path: ""                  # SecureCRT 会话目录路径
  password: ""                      # SecureCRT 主密码（用于解密加密密码）

# Xshell 集成配置
xshell:
  enabled: false                    # 是否启用 Xshell 集成
  session_path: ""                  # Xshell 会话目录路径（包含 .xsh 文件）
  password: ""                      # Xshell 主密码（用于解密加密密码）

# SSH 连接配置
ssh:
  strict_host_key: false            # 是否启用严格主机密钥验证（默认关闭）
  known_hosts_file: ""              # 自定义 known_hosts 文件路径
```

### SSH 主机密钥验证

- `strict_host_key: false`（默认）：自动接受所有主机密钥，适合内网环境
- `strict_host_key: true`：使用 known_hosts 验证主机密钥，首次连接未知主机将拒绝连接

known_hosts 文件查找顺序：
1. 配置中指定的 `known_hosts_file` 路径
2. `~/.ssh/known_hosts`（标准 SSH 路径）
3. `~/.xsc/known_hosts`（xsc 自有路径）

> 配置文件权限建议设置为 `0600`（仅用户可读写），避免密码泄露。

## 目录结构

```
~/.xsc/
├── config.yaml                    # 全局配置文件
├── known_hosts                    # 主机密钥文件（启用 strict_host_key 时使用）
└── sessions/                      # 本地会话目录
    ├── prod/                      # 按环境分组
    │   └── db/
    │       ├── master.yaml        # 各个会话配置文件
    │       └── slave-01.yaml
    ├── staging/
    │   └── web-server.yaml
    ├── securecrt-converted/       # import-securecrt 转换后的会话
    │   └── 20240101-120000/
    │       └── ...                # 保持 SecureCRT 原有目录结构
    └── xshell-converted/          # import-xshell 转换后的会话
        └── 20240101-120000/
            └── ...                # 保持 Xshell 原有目录结构
```

## 开发

```bash
make build          # 构建到 ./build/xsc
make run            # 直接运行 TUI 模式
make test           # 运行所有测试：go test -v ./...
make fmt            # 格式化代码：go fmt ./...
make vet            # 静态分析：go vet ./...
make deps           # 下载并整理依赖
make clean          # 清理构建产物
make uninstall      # 从系统卸载
```

运行指定测试：
```bash
go test -v ./internal/securecrt/... -run TestDecryptPasswordV2Real
```

完整质量检查：
```bash
make fmt && make vet && make test
```
