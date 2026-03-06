// Package xshell 提供解析和解密 Xshell 会话文件的功能
package xshell

import (
	"crypto/rc4"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Session 表示一个 Xshell 会话
type Session struct {
	Name              string // 会话名称
	Hostname          string // 主机地址
	Port              int    // 端口号
	Username          string // 用户名
	Password          string // 解密后的密码（延迟解密时为空）
	EncryptedPassword string // Base64 编码的加密密码（用于延迟解密）
	FilePath          string // .xsh 文件路径
	Folder            string // 相对目录路径
	Version           string // Xshell 版本
}

// Config 表示 Xshell 导入配置
type Config struct {
	SessionPath string // 会话文件目录
	Password    string // 主密码（用于解密）
}

// LoadSessions 加载目录下所有 Xshell 会话文件
func LoadSessions(config Config) ([]*Session, error) {
	if _, err := os.Stat(config.SessionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Xshell 会话目录不存在: %s", config.SessionPath)
	}

	var sessions []*Session

	err := filepath.Walk(config.SessionPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Xshell 会话文件以 .xsh 结尾
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".xsh") {
			session, err := parseSessionFile(path, config.SessionPath, config.Password)
			if err != nil {
				// 跳过解析失败的文件
				return nil
			}
			// 只添加有 hostname 的会话
			if session.Hostname != "" {
				sessions = append(sessions, session)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// DecryptPassword 解密密码（公开接口，供外部延迟解密调用）
func DecryptPassword(encryptedBase64, masterPassword string) (string, error) {
	if encryptedBase64 == "" || masterPassword == "" {
		return "", fmt.Errorf("加密密码或主密码为空")
	}

	return decryptRC4Password(encryptedBase64, masterPassword)
}

// ConvertToXSCSession 将 Xshell 会话转换为 xsc 会话格式
func (s *Session) ConvertToXSCSession() map[string]interface{} {
	result := map[string]interface{}{
		"host": s.Hostname,
		"port": s.Port,
		"user": s.Username,
	}

	// 有加密密码 → password 认证
	if s.EncryptedPassword != "" {
		result["auth_type"] = "password"
		result["encrypted_password"] = s.EncryptedPassword
	} else {
		// 无加密密码 → agent 认证（默认）
		result["auth_type"] = "agent"
	}

	// 如果密码已解密，填入
	if s.Password != "" {
		result["password"] = s.Password
	}

	return result
}

// parseSessionFile 解析单个 .xsh 会话文件
func parseSessionFile(filePath, basePath, masterPassword string) (*Session, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 解码 UTF-16LE（Xshell 使用 UTF-16LE 编码）
	var content string
	if isUTF16LE(data) {
		decoded, err := decodeUTF16LE(data)
		if err != nil {
			content = string(data)
		} else {
			content = decoded
		}
	} else {
		content = string(data)
	}

	session := &Session{
		FilePath: filePath,
		Port:     22, // 默认端口
	}

	// 计算相对目录路径
	relPath, _ := filepath.Rel(basePath, filepath.Dir(filePath))
	if relPath != "." {
		session.Folder = relPath
	}

	// 从文件名获取会话名称（去掉 .xsh 后缀）
	baseName := filepath.Base(filePath)
	session.Name = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// 解析 INI 格式
	sections := parseINISections(content)

	// 从 CONNECTION 段读取基本连接信息
	if conn, ok := sections["CONNECTION"]; ok {
		if host, ok := conn["Host"]; ok {
			session.Hostname = host
		}
		if portStr, ok := conn["Port"]; ok {
			if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
				session.Port = port
			}
		}
		if version, ok := conn["Version"]; ok {
			session.Version = version
		}
	}

	// 从 CONNECTION:AUTHENTICATION 段读取认证信息
	if auth, ok := sections["CONNECTION:AUTHENTICATION"]; ok {
		if user, ok := auth["UserName"]; ok {
			session.Username = user
		}
		if encPwd, ok := auth["Password"]; ok && encPwd != "" {
			session.EncryptedPassword = encPwd
		}
	}

	return session, nil
}

// isUTF16LE 检测数据是否为 UTF-16LE 编码
// 通过检查 BOM 或者检查是否存在大量 0x00 字节（ASCII 范围的 UTF-16LE 特征）
func isUTF16LE(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// 检查 UTF-16LE BOM (FF FE)
	if data[0] == 0xFF && data[1] == 0xFE {
		return true
	}

	// 检查 ASCII 范围的 UTF-16LE 特征：奇数位置大部分为 0x00
	if len(data) < 4 {
		return false
	}
	nullCount := 0
	checkLen := len(data)
	if checkLen > 100 {
		checkLen = 100
	}
	for i := 1; i < checkLen; i += 2 {
		if data[i] == 0x00 {
			nullCount++
		}
	}
	// 如果超过 80% 的奇数位置是 0x00，认为是 UTF-16LE
	total := checkLen / 2
	return total > 0 && nullCount*100/total > 80
}

// decodeUTF16LE 将 UTF-16LE 编码（可能含 BOM）的字节转换为 UTF-8 字符串
func decodeUTF16LE(data []byte) (string, error) {
	if len(data) < 2 {
		return "", fmt.Errorf("数据太短，无法解码 UTF-16LE")
	}

	// 使用 golang.org/x/text 的 UTF-16LE 解码器（自动处理 BOM）
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	result, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return "", fmt.Errorf("UTF-16LE 解码失败: %w", err)
	}

	return string(result), nil
}

// parseINISections 解析 INI 格式内容，支持冒号分层 Section（如 CONNECTION:AUTHENTICATION）
func parseINISections(content string) map[string]map[string]string {
	sections := make(map[string]map[string]string)
	currentSection := ""

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 去掉 \r（Windows 换行）
		line = strings.TrimRight(line, "\r")

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// 检查 Section 头（支持 [CONNECTION:AUTHENTICATION] 格式）
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			if _, ok := sections[currentSection]; !ok {
				sections[currentSection] = make(map[string]string)
			}
			continue
		}

		// 解析 key=value
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			if currentSection != "" {
				sections[currentSection][key] = value
			}
		}
	}

	return sections
}

// decryptRC4Password 使用 RC4 解密 Xshell 密码（主密码模式）
// 流程: Base64 解码 → SHA256(masterPassword) 作为密钥 → RC4 解密 → SHA256 校验
func decryptRC4Password(encryptedBase64, masterPassword string) (string, error) {
	// Base64 解码
	cipherData, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %w", err)
	}

	// 密钥 = SHA256(masterPassword)
	key := sha256.Sum256([]byte(masterPassword))

	// 校验和在末尾 32 字节
	if len(cipherData) <= sha256.Size {
		return "", fmt.Errorf("加密数据太短")
	}

	// 分离密文和校验和
	ciphertext := cipherData[:len(cipherData)-sha256.Size]
	checksum := cipherData[len(cipherData)-sha256.Size:]

	// RC4 解密
	cipher, err := rc4.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("创建 RC4 密钥失败: %w", err)
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.XORKeyStream(plaintext, ciphertext)

	// SHA256(plaintext) 与校验和比对
	expected := sha256.Sum256(plaintext)
	for i := 0; i < sha256.Size; i++ {
		if checksum[i] != expected[i] {
			return "", fmt.Errorf("校验失败: 主密码可能不正确")
		}
	}

	return string(plaintext), nil
}
