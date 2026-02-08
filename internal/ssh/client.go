package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/xsc/internal/session"
	"github.com/user/xsc/pkg/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// Connect 连接到 SSH 会话
func Connect(s *session.Session) error {
	if !s.Valid {
		return fmt.Errorf("invalid session: %v", s.Error)
	}

	// 延迟解密密码（SecureCRT 会话）
	if s.AuthType == session.AuthTypePassword && s.Password == "" && s.EncryptedPassword != "" {
		if err := s.ResolvePassword(); err != nil {
			return fmt.Errorf("failed to resolve password: %w", err)
		}
	}

	switch s.AuthType {
	case session.AuthTypePassword:
		return connectWithPassword(s)
	case session.AuthTypeKey:
		return connectWithKey(s)
	case session.AuthTypeAgent:
		return connectWithAgent(s)
	default:
		return fmt.Errorf("unsupported auth type: %s", s.AuthType)
	}
}

// getSSHConfig 根据认证类型获取 SSH 客户端配置
// 返回的 cleanup 函数用于关闭 SSH Agent 连接（非 agent 模式时为 nil）
func getSSHConfig(s *session.Session) (*ssh.ClientConfig, func(), error) {
	// 默认忽略主机密钥验证
	hostKeyCallback := ssh.InsecureIgnoreHostKey()

	// 如果配置中启用了严格主机密钥验证，则使用 known_hosts
	cfg, err := config.LoadGlobalConfig()
	if err == nil && cfg.SSH.StrictHostKey {
		knownHostsPath, err := config.GetKnownHostsPath()
		if err == nil && knownHostsPath != "" {
			if _, statErr := os.Stat(knownHostsPath); statErr == nil {
				// 文件存在，使用 known_hosts 验证
				hostKeyCallback, err = knownhosts.New(knownHostsPath)
				if err != nil {
					// 如果创建 known_hosts 回调失败，回退到忽略
					hostKeyCallback = ssh.InsecureIgnoreHostKey()
				}
			}
		}
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.User,
		HostKeyCallback: hostKeyCallback,
	}

	var cleanup func()

	switch s.AuthType {
	case session.AuthTypePassword:
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(s.Password),
		}
	case session.AuthTypeKey:
		key, err := os.ReadFile(s.KeyPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read key file: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	case session.AuthTypeAgent:
		authMethod, agentConn, err := getSSHAgentAuth()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get SSH agent auth: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{authMethod}
		cleanup = func() { agentConn.Close() }
	default:
		return nil, nil, fmt.Errorf("unsupported auth type: %s", s.AuthType)
	}

	return sshConfig, cleanup, nil
}

// AgentKeyInfo 描述 SSH Agent 中的一个密钥
type AgentKeyInfo struct {
	Type    string
	Bits    int
	Comment string
}

// ListAgentKeys 列出 SSH Agent 中的所有密钥
func ListAgentKeys() ([]AgentKeyInfo, error) {
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	conn, err := net.Dial("unix", authSock)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ssh-agent: %w", err)
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)
	keys, err := agentClient.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	var result []AgentKeyInfo
	for _, k := range keys {
		info := AgentKeyInfo{
			Type:    k.Type(),
			Comment: k.Comment,
		}
		result = append(result, info)
	}
	return result, nil
}

// getSSHAgentAuth 获取 SSH Agent 认证方法
// 返回的 net.Conn 需要调用方在 SSH 连接结束后关闭
func getSSHAgentAuth() (ssh.AuthMethod, net.Conn, error) {
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		return nil, nil, fmt.Errorf("SSH_AUTH_SOCK not set, is ssh-agent running?")
	}

	conn, err := net.Dial("unix", authSock)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to ssh-agent: %w", err)
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers), conn, nil
}

// connectInteractive 建立交互式 SSH 连接
func connectInteractive(s *session.Session, config *ssh.ClientConfig) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer sess.Close()

	// 获取终端尺寸
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	// 配置终端模式
	// ONLCR: 将输出中的 \n 转换为 \r\n，解决换行问题
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ONLCR:         1,
		ssh.OPOST:         1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	// 请求伪终端
	if err := sess.RequestPty(termType, height, width, modes); err != nil {
		return fmt.Errorf("failed to request pty: %w", err)
	}

	// 将本地终端设为 raw 模式（必须在启动 shell 之前）
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// 获取 stdin/stdout/stderr pipes
	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	defer stdinPipe.Close()

	stdoutPipe, err := sess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderrPipe, err := sess.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// 启动 shell
	if err := sess.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// 设置窗口大小调整处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleWindowResize(ctx, sess)

	// 使用 goroutines 在本地终端和 SSH session 之间传输数据
	errChan := make(chan error, 3)

	// 本地 stdin -> 远程 stdin
	go func() {
		_, err := io.Copy(stdinPipe, os.Stdin)
		if err != nil {
			errChan <- fmt.Errorf("stdin copy error: %w", err)
		}
	}()

	// 远程 stdout -> 本地 stdout
	go func() {
		_, err := io.Copy(os.Stdout, stdoutPipe)
		if err != nil {
			errChan <- fmt.Errorf("stdout copy error: %w", err)
		}
	}()

	// 远程 stderr -> 本地 stderr
	go func() {
		_, err := io.Copy(os.Stderr, stderrPipe)
		if err != nil {
			errChan <- fmt.Errorf("stderr copy error: %w", err)
		}
	}()

	// 等待会话结束
	err = sess.Wait()
	if err != nil {
		// ExitError 表示远程命令以非零状态退出，属于正常退出
		if _, ok := err.(*ssh.ExitError); ok {
			return nil
		}
		return err
	}

	return nil
}

// connectWithPassword 使用密码认证建立 SSH 连接
func connectWithPassword(s *session.Session) error {
	config, cleanup, err := getSSHConfig(s)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	return connectInteractive(s, config)
}

// connectWithKey 使用密钥认证建立 SSH 连接
func connectWithKey(s *session.Session) error {
	config, cleanup, err := getSSHConfig(s)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	return connectInteractive(s, config)
}

// connectWithAgent 使用 SSH Agent 建立 SSH 连接
func connectWithAgent(s *session.Session) error {
	config, cleanup, err := getSSHConfig(s)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	return connectInteractive(s, config)
}

// handleWindowResize 处理终端窗口大小调整
func handleWindowResize(ctx context.Context, sess *ssh.Session) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	defer signal.Stop(sigChan)

	for {
		select {
		case <-sigChan:
			width, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				continue
			}
			sess.WindowChange(height, width)
		case <-ctx.Done():
			return
		}
	}
}

// ConnectWithIO 使用自定义输入输出流连接
func ConnectWithIO(s *session.Session, stdin io.Reader, stdout, stderr io.Writer) error {
	if !s.Valid {
		return fmt.Errorf("invalid session: %v", s.Error)
	}

	switch s.AuthType {
	case session.AuthTypePassword:
		return connectWithPasswordIO(s, stdin, stdout, stderr)
	case session.AuthTypeKey:
		return connectWithKeyIO(s, stdin, stdout, stderr)
	case session.AuthTypeAgent:
		return connectWithAgentIO(s, stdin, stdout, stderr)
	default:
		return fmt.Errorf("unsupported auth type: %s", s.AuthType)
	}
}

// connectWithIO 建立非交互式 SSH 连接（支持自定义 IO）
func connectWithIO(s *session.Session, stdin io.Reader, stdout, stderr io.Writer, config *ssh.ClientConfig) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer sess.Close()

	// 请求伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ONLCR:         1,
		ssh.OPOST:         1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	if err := sess.RequestPty(termType, 24, 80, modes); err != nil {
		return fmt.Errorf("failed to request pty: %w", err)
	}

	// 获取 pipes
	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	defer stdinPipe.Close()

	stdoutPipe, err := sess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderrPipe, err := sess.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := sess.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// 使用 goroutines 传输数据
	errChan := make(chan error, 3)

	go func() {
		_, err := io.Copy(stdinPipe, stdin)
		if err != nil {
			errChan <- fmt.Errorf("stdin copy error: %w", err)
		}
	}()

	go func() {
		_, err := io.Copy(stdout, stdoutPipe)
		if err != nil {
			errChan <- fmt.Errorf("stdout copy error: %w", err)
		}
	}()

	go func() {
		_, err := io.Copy(stderr, stderrPipe)
		if err != nil {
			errChan <- fmt.Errorf("stderr copy error: %w", err)
		}
	}()

	if err := sess.Wait(); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			return fmt.Errorf("ssh session exited with code %d", exitErr.ExitStatus())
		}
		return err
	}

	// 检查传输错误
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	default:
	}

	return nil
}

// connectWithPasswordIO 使用密码认证建立 SSH 连接（支持自定义 IO）
func connectWithPasswordIO(s *session.Session, stdin io.Reader, stdout, stderr io.Writer) error {
	config, cleanup, err := getSSHConfig(s)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	return connectWithIO(s, stdin, stdout, stderr, config)
}

// connectWithKeyIO 使用密钥认证建立 SSH 连接（支持自定义 IO）
func connectWithKeyIO(s *session.Session, stdin io.Reader, stdout, stderr io.Writer) error {
	config, cleanup, err := getSSHConfig(s)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	return connectWithIO(s, stdin, stdout, stderr, config)
}

// connectWithAgentIO 使用 SSH Agent 建立 SSH 连接（支持自定义 IO）
func connectWithAgentIO(s *session.Session, stdin io.Reader, stdout, stderr io.Writer) error {
	config, cleanup, err := getSSHConfig(s)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	return connectWithIO(s, stdin, stdout, stderr, config)
}
