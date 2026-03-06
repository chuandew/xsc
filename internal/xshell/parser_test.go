package xshell

import (
	"crypto/rc4"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDecodeUTF16LE_WithBOM 测试含 BOM 的 UTF-16LE 解码
func TestDecodeUTF16LE_WithBOM(t *testing.T) {
	// UTF-16LE BOM (FF FE) + "Hello" 的 UTF-16LE 编码
	data := []byte{
		0xFF, 0xFE, // BOM
		0x48, 0x00, // H
		0x65, 0x00, // e
		0x6C, 0x00, // l
		0x6C, 0x00, // l
		0x6F, 0x00, // o
	}

	result, err := decodeUTF16LE(data)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	if result != "Hello" {
		t.Errorf("期望 'Hello'，得到 '%s'", result)
	}
}

// TestDecodeUTF16LE_WithoutBOM 测试不含 BOM 的 UTF-16LE 解码
func TestDecodeUTF16LE_WithoutBOM(t *testing.T) {
	// "Test" 的 UTF-16LE 编码（不含 BOM）
	data := []byte{
		0x54, 0x00, // T
		0x65, 0x00, // e
		0x73, 0x00, // s
		0x74, 0x00, // t
	}

	result, err := decodeUTF16LE(data)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	if result != "Test" {
		t.Errorf("期望 'Test'，得到 '%s'", result)
	}
}

// TestDecodeUTF16LE_Chinese 测试中文字符的 UTF-16LE 解码
func TestDecodeUTF16LE_Chinese(t *testing.T) {
	// UTF-16LE BOM + "你好" 的 UTF-16LE 编码
	data := []byte{
		0xFF, 0xFE, // BOM
		0x60, 0x4F, // 你
		0x7D, 0x59, // 好
	}

	result, err := decodeUTF16LE(data)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	if result != "你好" {
		t.Errorf("期望 '你好'，得到 '%s'", result)
	}
}

// TestDecodeUTF16LE_TooShort 测试过短的数据
func TestDecodeUTF16LE_TooShort(t *testing.T) {
	data := []byte{0xFF}

	_, err := decodeUTF16LE(data)
	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestParseINISections_Basic 测试基本 INI 解析
func TestParseINISections_Basic(t *testing.T) {
	content := `[CONNECTION]
Host=192.168.1.1
Port=22
Version=5.0

[CONNECTION:AUTHENTICATION]
UserName=root
Password=abc123
`

	sections := parseINISections(content)

	// 检查 CONNECTION 段
	conn, ok := sections["CONNECTION"]
	if !ok {
		t.Fatal("缺少 CONNECTION 段")
	}
	if conn["Host"] != "192.168.1.1" {
		t.Errorf("Host 期望 '192.168.1.1'，得到 '%s'", conn["Host"])
	}
	if conn["Port"] != "22" {
		t.Errorf("Port 期望 '22'，得到 '%s'", conn["Port"])
	}
	if conn["Version"] != "5.0" {
		t.Errorf("Version 期望 '5.0'，得到 '%s'", conn["Version"])
	}

	// 检查 CONNECTION:AUTHENTICATION 段
	auth, ok := sections["CONNECTION:AUTHENTICATION"]
	if !ok {
		t.Fatal("缺少 CONNECTION:AUTHENTICATION 段")
	}
	if auth["UserName"] != "root" {
		t.Errorf("UserName 期望 'root'，得到 '%s'", auth["UserName"])
	}
	if auth["Password"] != "abc123" {
		t.Errorf("Password 期望 'abc123'，得到 '%s'", auth["Password"])
	}
}

// TestParseINISections_MultipleColonSections 测试多层冒号分隔的 Section
func TestParseINISections_MultipleColonSections(t *testing.T) {
	content := `[CONNECTION]
Host=10.0.0.1

[CONNECTION:AUTHENTICATION]
UserName=admin

[CONNECTION:PROXY]
ProxyHost=proxy.example.com
ProxyPort=8080
`

	sections := parseINISections(content)

	if len(sections) != 3 {
		t.Errorf("期望 3 个 Section，得到 %d 个", len(sections))
	}

	if sections["CONNECTION:PROXY"]["ProxyHost"] != "proxy.example.com" {
		t.Errorf("ProxyHost 期望 'proxy.example.com'，得到 '%s'", sections["CONNECTION:PROXY"]["ProxyHost"])
	}
}

// TestParseINISections_CommentsAndEmptyLines 测试注释和空行
func TestParseINISections_CommentsAndEmptyLines(t *testing.T) {
	content := `; 这是注释
# 这也是注释

[CONNECTION]
Host=10.0.0.1
; Port=2222
Port=22
`

	sections := parseINISections(content)

	conn := sections["CONNECTION"]
	if conn["Host"] != "10.0.0.1" {
		t.Errorf("Host 期望 '10.0.0.1'，得到 '%s'", conn["Host"])
	}
	if conn["Port"] != "22" {
		t.Errorf("Port 期望 '22'，得到 '%s'", conn["Port"])
	}
}

// TestDecryptRC4Password 测试 RC4 解密
func TestDecryptRC4Password(t *testing.T) {
	// 构造已知的加密/解密对
	masterPassword := "test_master_password"
	plaintext := "my_secret_password"

	// 使用相同的加密流程构造测试数据
	key := sha256.Sum256([]byte(masterPassword))
	checksum := sha256.Sum256([]byte(plaintext))

	// RC4 加密明文部分
	cipher, err := rc4.NewCipher(key[:])
	if err != nil {
		t.Fatalf("创建 RC4 密钥失败: %v", err)
	}

	encryptedPlaintext := make([]byte, len(plaintext))
	cipher.XORKeyStream(encryptedPlaintext, []byte(plaintext))

	// 组合：RC4 加密后的密文 + 明文的 SHA256 校验和（不加密）
	cipherData := append(encryptedPlaintext, checksum[:]...)

	// Base64 编码
	encryptedBase64 := base64.StdEncoding.EncodeToString(cipherData)

	// 测试解密
	result, err := decryptRC4Password(encryptedBase64, masterPassword)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if result != plaintext {
		t.Errorf("期望 '%s'，得到 '%s'", plaintext, result)
	}
}

// TestDecryptRC4Password_WrongPassword 测试错误主密码
func TestDecryptRC4Password_WrongPassword(t *testing.T) {
	masterPassword := "correct_password"
	plaintext := "secret"

	key := sha256.Sum256([]byte(masterPassword))
	checksum := sha256.Sum256([]byte(plaintext))

	// RC4 加密明文部分
	cipher, _ := rc4.NewCipher(key[:])
	encryptedPlaintext := make([]byte, len(plaintext))
	cipher.XORKeyStream(encryptedPlaintext, []byte(plaintext))

	// 组合：加密密文 + 明文校验和
	cipherData := append(encryptedPlaintext, checksum[:]...)
	encryptedBase64 := base64.StdEncoding.EncodeToString(cipherData)

	// 使用错误密码解密
	_, err := decryptRC4Password(encryptedBase64, "wrong_password")
	if err == nil {
		t.Error("期望返回错误，但没有")
	}
	if err != nil && !strings.Contains(err.Error(), "校验失败") {
		t.Errorf("期望校验失败错误，得到: %v", err)
	}
}

// TestDecryptPassword_Empty 测试空密码
func TestDecryptPassword_Empty(t *testing.T) {
	_, err := DecryptPassword("", "master")
	if err == nil {
		t.Error("期望返回错误，但没有")
	}

	_, err = DecryptPassword("encrypted", "")
	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestConvertToXSCSession_WithPassword 测试有密码的会话转换
func TestConvertToXSCSession_WithPassword(t *testing.T) {
	session := &Session{
		Name:              "test-server",
		Hostname:          "192.168.1.1",
		Port:              22,
		Username:          "root",
		EncryptedPassword: "encrypted_data",
	}

	result := session.ConvertToXSCSession()

	if result["host"] != "192.168.1.1" {
		t.Errorf("host 期望 '192.168.1.1'，得到 '%s'", result["host"])
	}
	if result["port"] != 22 {
		t.Errorf("port 期望 22，得到 %v", result["port"])
	}
	if result["user"] != "root" {
		t.Errorf("user 期望 'root'，得到 '%s'", result["user"])
	}
	if result["auth_type"] != "password" {
		t.Errorf("auth_type 期望 'password'，得到 '%s'", result["auth_type"])
	}
	if result["encrypted_password"] != "encrypted_data" {
		t.Errorf("encrypted_password 期望 'encrypted_data'，得到 '%s'", result["encrypted_password"])
	}
}

// TestConvertToXSCSession_WithoutPassword 测试无密码的会话转换
func TestConvertToXSCSession_WithoutPassword(t *testing.T) {
	session := &Session{
		Name:     "test-server",
		Hostname: "10.0.0.1",
		Port:     2222,
		Username: "admin",
	}

	result := session.ConvertToXSCSession()

	if result["auth_type"] != "agent" {
		t.Errorf("auth_type 期望 'agent'，得到 '%s'", result["auth_type"])
	}
}

// TestConvertToXSCSession_WithDecryptedPassword 测试已解密密码的会话转换
func TestConvertToXSCSession_WithDecryptedPassword(t *testing.T) {
	session := &Session{
		Hostname:          "10.0.0.1",
		Port:              22,
		Username:          "root",
		Password:          "decrypted_pwd",
		EncryptedPassword: "encrypted_data",
	}

	result := session.ConvertToXSCSession()

	if result["password"] != "decrypted_pwd" {
		t.Errorf("password 期望 'decrypted_pwd'，得到 '%s'", result["password"])
	}
}

// TestLoadSessions_NotExistPath 测试不存在的路径
func TestLoadSessions_NotExistPath(t *testing.T) {
	config := Config{
		SessionPath: "/nonexistent/path",
		Password:    "test",
	}

	_, err := LoadSessions(config)
	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestLoadSessions_WithXshFiles 测试加载 .xsh 文件
func TestLoadSessions_WithXshFiles(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建子目录
	subDir := filepath.Join(tmpDir, "production")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}

	// 创建 UTF-8 格式的 .xsh 文件（Xshell 也支持 UTF-8）
	xshContent := `[CONNECTION]
Host=192.168.1.100
Port=22
Version=5.0

[CONNECTION:AUTHENTICATION]
UserName=admin
Password=
`
	if err := os.WriteFile(filepath.Join(tmpDir, "server1.xsh"), []byte(xshContent), 0600); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 创建子目录中的 .xsh 文件
	xshContent2 := `[CONNECTION]
Host=10.0.0.50
Port=2222

[CONNECTION:AUTHENTICATION]
UserName=root
Password=encrypted_pwd_data
`
	if err := os.WriteFile(filepath.Join(subDir, "db-server.xsh"), []byte(xshContent2), 0600); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 创建无 hostname 的文件（应跳过）
	xshContent3 := `[CONNECTION]
Port=22

[CONNECTION:AUTHENTICATION]
UserName=test
`
	if err := os.WriteFile(filepath.Join(tmpDir, "empty.xsh"), []byte(xshContent3), 0600); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 加载会话
	config := Config{
		SessionPath: tmpDir,
		Password:    "master",
	}

	sessions, err := LoadSessions(config)
	if err != nil {
		t.Fatalf("加载会话失败: %v", err)
	}

	// 应该加载 2 个有 hostname 的会话（跳过 empty.xsh）
	if len(sessions) != 2 {
		t.Fatalf("期望 2 个会话，得到 %d 个", len(sessions))
	}

	// 验证会话信息
	sessionMap := make(map[string]*Session)
	for _, s := range sessions {
		sessionMap[s.Name] = s
	}

	s1, ok := sessionMap["server1"]
	if !ok {
		t.Fatal("缺少 server1 会话")
	}
	if s1.Hostname != "192.168.1.100" {
		t.Errorf("server1 hostname 期望 '192.168.1.100'，得到 '%s'", s1.Hostname)
	}
	if s1.Port != 22 {
		t.Errorf("server1 port 期望 22，得到 %d", s1.Port)
	}
	if s1.Folder != "" {
		t.Errorf("server1 folder 期望为空，得到 '%s'", s1.Folder)
	}

	s2, ok := sessionMap["db-server"]
	if !ok {
		t.Fatal("缺少 db-server 会话")
	}
	if s2.Hostname != "10.0.0.50" {
		t.Errorf("db-server hostname 期望 '10.0.0.50'，得到 '%s'", s2.Hostname)
	}
	if s2.Port != 2222 {
		t.Errorf("db-server port 期望 2222，得到 %d", s2.Port)
	}
	if s2.Folder != "production" {
		t.Errorf("db-server folder 期望 'production'，得到 '%s'", s2.Folder)
	}
	if s2.EncryptedPassword != "encrypted_pwd_data" {
		t.Errorf("db-server encrypted password 期望 'encrypted_pwd_data'，得到 '%s'", s2.EncryptedPassword)
	}
}

// TestParseSessionFile_DefaultPort 测试默认端口
func TestParseSessionFile_DefaultPort(t *testing.T) {
	tmpDir := t.TempDir()

	xshContent := `[CONNECTION]
Host=10.0.0.1

[CONNECTION:AUTHENTICATION]
UserName=user1
`
	filePath := filepath.Join(tmpDir, "test.xsh")
	if err := os.WriteFile(filePath, []byte(xshContent), 0600); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	session, err := parseSessionFile(filePath, tmpDir, "")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if session.Port != 22 {
		t.Errorf("默认端口期望 22，得到 %d", session.Port)
	}
	if session.Name != "test" {
		t.Errorf("名称期望 'test'，得到 '%s'", session.Name)
	}
}
