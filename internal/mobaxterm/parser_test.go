package mobaxterm

import (
	"crypto/aes"
	"crypto/sha512"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

// TestUnescapeSpecialChars 测试特殊字符反转义
func TestUnescapeSpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"DIEZE 转换为 #", "test__DIEZE__value", "test#value"},
		{"PTVIRG 转换为 ;", "test__PTVIRG__value", "test;value"},
		{"DBLQUO 转换为双引号", "test__DBLQUO__value", "test\"value"},
		{"PIPE 转换为 |", "test__PIPE__value", "test|value"},
		{"PERCENT 转换为 %", "test__PERCENT__value", "test%value"},
		{"多个转义组合", "__DIEZE____PTVIRG____PIPE__", "#;|"},
		{"无转义字符", "hello_world", "hello_world"},
		{"空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unescapeSpecialChars(tt.input)
			if result != tt.expected {
				t.Errorf("期望 %q，得到 %q", tt.expected, result)
			}
		})
	}
}

// TestParseSessionLine 测试从会话行解析 SSH 会话
func TestParseSessionLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		folder   string
		wantHost string
		wantPort int
		wantUser string
		wantName string
		wantErr  bool
	}{
		{
			name:     "标准 SSH 会话",
			line:     "my-server=0%192.168.1.100%22%root%#font%size",
			folder:   "",
			wantHost: "192.168.1.100",
			wantPort: 22,
			wantUser: "root",
			wantName: "my-server",
		},
		{
			name:     "自定义端口",
			line:     "dev-box=0%10.0.0.5%2222%admin%#font",
			folder:   "production",
			wantHost: "10.0.0.5",
			wantPort: 2222,
			wantUser: "admin",
			wantName: "dev-box",
		},
		{
			name:     "会话名包含转义字符",
			line:     "server__DIEZE__1=0%host.example.com%22%user%",
			folder:   "",
			wantHost: "host.example.com",
			wantPort: 22,
			wantUser: "user",
			wantName: "server#1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := parseSessionLine(tt.line, "/tmp/MobaXterm.ini", tt.folder)
			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有")
				}
				return
			}
			if err != nil {
				t.Fatalf("解析失败: %v", err)
			}
			if session.Hostname != tt.wantHost {
				t.Errorf("Hostname 期望 %q，得到 %q", tt.wantHost, session.Hostname)
			}
			if session.Port != tt.wantPort {
				t.Errorf("Port 期望 %d，得到 %d", tt.wantPort, session.Port)
			}
			if session.Username != tt.wantUser {
				t.Errorf("Username 期望 %q，得到 %q", tt.wantUser, session.Username)
			}
			if session.Name != tt.wantName {
				t.Errorf("Name 期望 %q，得到 %q", tt.wantName, session.Name)
			}
			if session.Folder != tt.folder {
				t.Errorf("Folder 期望 %q，得到 %q", tt.folder, session.Folder)
			}
		})
	}
}

// TestParseSessionLineNonSSH 测试非 SSH 类型会话被跳过
func TestParseSessionLineNonSSH(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"RDP 类型 (type=1)", "rdp-server=1%10.0.0.1%3389%admin%"},
		{"VNC 类型 (type=2)", "vnc-server=2%10.0.0.2%5900%user%"},
		{"FTP 类型 (type=4)", "ftp-server=4%10.0.0.3%21%ftpuser%"},
		{"Telnet 类型 (type=109)", "telnet=109%10.0.0.4%23%user%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSessionLine(tt.line, "/tmp/MobaXterm.ini", "")
			if err == nil {
				t.Error("期望返回错误（非 SSH 类型应被跳过），但没有")
			}
		})
	}
}

// TestParseSessionLineDefaultPort 测试端口为空时默认 22
func TestParseSessionLineDefaultPort(t *testing.T) {
	// 端口为空字符串
	session, err := parseSessionLine("server=0%10.0.0.1%%root%", "/tmp/test.ini", "")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if session.Port != 22 {
		t.Errorf("默认端口期望 22，得到 %d", session.Port)
	}

	// 没有端口字段（字段数不足）
	session2, err := parseSessionLine("server2=0%10.0.0.2%", "/tmp/test.ini", "")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if session2.Port != 22 {
		t.Errorf("默认端口期望 22，得到 %d", session2.Port)
	}
}

// TestParseBookmarksSections 测试 [Bookmarks] 和 [Bookmarks_N] 解析
func TestParseBookmarksSections(t *testing.T) {
	content := `[Bookmarks]
SubRep=
ImgNum=42
my-server=0%192.168.1.1%22%root%#font%10

[Bookmarks_1]
SubRep=production\web
ImgNum=42
web-server=0%10.0.0.1%22%admin%#font%10
db-server=0%10.0.0.2%3306%dba%#font%10

[SomeOtherSection]
key=value
`

	sections := parseBookmarksSections(content)

	if len(sections) != 2 {
		t.Fatalf("期望 2 个 bookmarks section，得到 %d", len(sections))
	}

	// 第一个 section: [Bookmarks]
	if sections[0].SubRep != "" {
		t.Errorf("第一个 section SubRep 期望为空，得到 %q", sections[0].SubRep)
	}
	if len(sections[0].Sessions) != 1 {
		t.Fatalf("第一个 section 期望 1 个会话，得到 %d", len(sections[0].Sessions))
	}
	if sections[0].Sessions[0] != "my-server=0%192.168.1.1%22%root%#font%10" {
		t.Errorf("会话行不匹配: %q", sections[0].Sessions[0])
	}

	// 第二个 section: [Bookmarks_1]
	if sections[1].SubRep != `production\web` {
		t.Errorf("第二个 section SubRep 期望 'production\\web'，得到 %q", sections[1].SubRep)
	}
	if len(sections[1].Sessions) != 2 {
		t.Fatalf("第二个 section 期望 2 个会话，得到 %d", len(sections[1].Sessions))
	}
}

// TestSubRepFolderHierarchy 测试 SubRep 目录层级在加载时的构建
func TestSubRepFolderHierarchy(t *testing.T) {
	// 创建临时 INI 文件
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "MobaXterm.ini")

	content := `[Bookmarks]
SubRep=
ImgNum=42
root-server=0%10.0.0.1%22%root%#font%10

[Bookmarks_1]
SubRep=datacenter\rack01
ImgNum=42
web01=0%10.0.0.10%22%admin%#font%10
`

	if err := os.WriteFile(iniPath, []byte(content), 0600); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	sessions, err := LoadSessions(Config{SessionPath: iniPath})
	if err != nil {
		t.Fatalf("加载会话失败: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("期望 2 个会话，得到 %d", len(sessions))
	}

	// 检查 folder 赋值
	sessionMap := make(map[string]*Session)
	for _, s := range sessions {
		sessionMap[s.Name] = s
	}

	rootSrv, ok := sessionMap["root-server"]
	if !ok {
		t.Fatal("缺少 root-server 会话")
	}
	if rootSrv.Folder != "" {
		t.Errorf("root-server folder 期望为空，得到 %q", rootSrv.Folder)
	}

	web01, ok := sessionMap["web01"]
	if !ok {
		t.Fatal("缺少 web01 会话")
	}
	if web01.Folder != `datacenter\rack01` {
		t.Errorf("web01 folder 期望 'datacenter\\rack01'，得到 %q", web01.Folder)
	}
}

// TestLoadSessions 创建临时 INI 文件，测试完整加载流程
func TestLoadSessions(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "MobaXterm.ini")

	content := `[Bookmarks]
SubRep=
ImgNum=42
server1=0%192.168.1.100%22%root%#font%10
server2=0%10.0.0.5%2222%admin%#font%10

[Bookmarks_1]
SubRep=internal
ImgNum=42
db=0%172.16.0.1%5432%postgres%#font%10
`

	if err := os.WriteFile(iniPath, []byte(content), 0600); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	sessions, err := LoadSessions(Config{SessionPath: iniPath})
	if err != nil {
		t.Fatalf("加载会话失败: %v", err)
	}

	if len(sessions) != 3 {
		t.Fatalf("期望 3 个会话，得到 %d", len(sessions))
	}

	// 验证各会话
	sessionMap := make(map[string]*Session)
	for _, s := range sessions {
		sessionMap[s.Name] = s
	}

	s1, ok := sessionMap["server1"]
	if !ok {
		t.Fatal("缺少 server1 会话")
	}
	if s1.Hostname != "192.168.1.100" {
		t.Errorf("server1 hostname 期望 '192.168.1.100'，得到 %q", s1.Hostname)
	}
	if s1.Port != 22 {
		t.Errorf("server1 port 期望 22，得到 %d", s1.Port)
	}
	if s1.Username != "root" {
		t.Errorf("server1 user 期望 'root'，得到 %q", s1.Username)
	}

	s2, ok := sessionMap["server2"]
	if !ok {
		t.Fatal("缺少 server2 会话")
	}
	if s2.Port != 2222 {
		t.Errorf("server2 port 期望 2222，得到 %d", s2.Port)
	}

	db, ok := sessionMap["db"]
	if !ok {
		t.Fatal("缺少 db 会话")
	}
	if db.Folder != "internal" {
		t.Errorf("db folder 期望 'internal'，得到 %q", db.Folder)
	}
	if db.Hostname != "172.16.0.1" {
		t.Errorf("db hostname 期望 '172.16.0.1'，得到 %q", db.Hostname)
	}
}

// TestLoadSessionsFileNotExist 测试文件不存在时返回错误
func TestLoadSessionsFileNotExist(t *testing.T) {
	_, err := LoadSessions(Config{SessionPath: "/nonexistent/MobaXterm.ini"})
	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestLoadSessionsEmptyBookmarks 测试空书签处理
func TestLoadSessionsEmptyBookmarks(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "MobaXterm.ini")

	content := `[Bookmarks]
SubRep=
ImgNum=42

[SomeOtherSection]
key=value
`

	if err := os.WriteFile(iniPath, []byte(content), 0600); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	sessions, err := LoadSessions(Config{SessionPath: iniPath})
	if err != nil {
		t.Fatalf("加载会话失败: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("期望 0 个会话，得到 %d", len(sessions))
	}
}

// TestLoadSessionsMultipleSections 测试多个 [Bookmarks_N] section
func TestLoadSessionsMultipleSections(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "MobaXterm.ini")

	content := `[Bookmarks]
SubRep=
ImgNum=42
s1=0%10.0.0.1%22%user1%#font

[Bookmarks_1]
SubRep=group1
ImgNum=42
s2=0%10.0.0.2%22%user2%#font

[Bookmarks_2]
SubRep=group2
ImgNum=42
s3=0%10.0.0.3%22%user3%#font

[Bookmarks_3]
SubRep=group3
ImgNum=42
s4=0%10.0.0.4%22%user4%#font
`

	if err := os.WriteFile(iniPath, []byte(content), 0600); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	sessions, err := LoadSessions(Config{SessionPath: iniPath})
	if err != nil {
		t.Fatalf("加载会话失败: %v", err)
	}

	if len(sessions) != 4 {
		t.Fatalf("期望 4 个会话，得到 %d", len(sessions))
	}

	// 验证各会话的 folder
	folderMap := make(map[string]string)
	for _, s := range sessions {
		folderMap[s.Name] = s.Folder
	}

	if folderMap["s1"] != "" {
		t.Errorf("s1 folder 期望为空，得到 %q", folderMap["s1"])
	}
	if folderMap["s2"] != "group1" {
		t.Errorf("s2 folder 期望 'group1'，得到 %q", folderMap["s2"])
	}
	if folderMap["s3"] != "group2" {
		t.Errorf("s3 folder 期望 'group2'，得到 %q", folderMap["s3"])
	}
	if folderMap["s4"] != "group3" {
		t.Errorf("s4 folder 期望 'group3'，得到 %q", folderMap["s4"])
	}
}

// TestLoadSessionsSkipsNonSSH 测试非 SSH 会话被跳过
func TestLoadSessionsSkipsNonSSH(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "MobaXterm.ini")

	content := `[Bookmarks]
SubRep=
ImgNum=42
ssh-server=0%10.0.0.1%22%root%#font
rdp-server=1%10.0.0.2%3389%admin%#font
vnc-server=2%10.0.0.3%5900%user%#font
`

	if err := os.WriteFile(iniPath, []byte(content), 0600); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	sessions, err := LoadSessions(Config{SessionPath: iniPath})
	if err != nil {
		t.Fatalf("加载会话失败: %v", err)
	}

	// 只有 SSH 类型 (type=0) 被加载
	if len(sessions) != 1 {
		t.Fatalf("期望 1 个 SSH 会话，得到 %d", len(sessions))
	}
	if sessions[0].Name != "ssh-server" {
		t.Errorf("期望会话名 'ssh-server'，得到 %q", sessions[0].Name)
	}
}

// TestLoadSessionsSkipsEmptyHostname 测试空 hostname 会话被跳过
func TestLoadSessionsSkipsEmptyHostname(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "MobaXterm.ini")

	content := `[Bookmarks]
SubRep=
ImgNum=42
no-host=0%%22%root%#font
has-host=0%10.0.0.1%22%root%#font
`

	if err := os.WriteFile(iniPath, []byte(content), 0600); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	sessions, err := LoadSessions(Config{SessionPath: iniPath})
	if err != nil {
		t.Fatalf("加载会话失败: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("期望 1 个会话（跳过空 hostname），得到 %d", len(sessions))
	}
	if sessions[0].Name != "has-host" {
		t.Errorf("期望会话名 'has-host'，得到 %q", sessions[0].Name)
	}
}

// TestConvertToXSSHSession 测试会话格式转换（有密码）
func TestConvertToXSSHSession(t *testing.T) {
	session := &Session{
		Name:              "test-server",
		Hostname:          "192.168.1.1",
		Port:              22,
		Username:          "root",
		Password:          "my_password",
		EncryptedPassword: "encrypted_data",
	}

	result := session.ConvertToXSSHSession()

	if result["host"] != "192.168.1.1" {
		t.Errorf("host 期望 '192.168.1.1'，得到 %q", result["host"])
	}
	if result["port"] != 22 {
		t.Errorf("port 期望 22，得到 %v", result["port"])
	}
	if result["user"] != "root" {
		t.Errorf("user 期望 'root'，得到 %q", result["user"])
	}
	if result["auth_type"] != "password" {
		t.Errorf("auth_type 期望 'password'，得到 %q", result["auth_type"])
	}
	if result["encrypted_password"] != "encrypted_data" {
		t.Errorf("encrypted_password 期望 'encrypted_data'，得到 %q", result["encrypted_password"])
	}
	if result["password"] != "my_password" {
		t.Errorf("password 期望 'my_password'，得到 %q", result["password"])
	}
}

// TestConvertToXSSHSessionNoPassword 测试无密码时 auth_type 为 agent
func TestConvertToXSSHSessionNoPassword(t *testing.T) {
	session := &Session{
		Name:     "agent-server",
		Hostname: "10.0.0.5",
		Port:     2222,
		Username: "admin",
	}

	result := session.ConvertToXSSHSession()

	if result["auth_type"] != "agent" {
		t.Errorf("auth_type 期望 'agent'，得到 %q", result["auth_type"])
	}
	if _, hasPassword := result["password"]; hasPassword {
		t.Error("无密码会话不应包含 password 字段")
	}
	if _, hasEncrypted := result["encrypted_password"]; hasEncrypted {
		t.Error("无密码会话不应包含 encrypted_password 字段")
	}
}

// TestDecryptPassword 测试公开解密接口
func TestDecryptPassword(t *testing.T) {
	// 空参数测试
	_, err := DecryptPassword("", "master")
	if err == nil {
		t.Error("空加密密码应返回错误")
	}

	_, err = DecryptPassword("encrypted", "")
	if err == nil {
		t.Error("空主密码应返回错误")
	}
}

// TestDecryptModernWrongPassword 测试错误密码解密
// 使用正确密码加密数据，然后用错误密码解密，验证结果不同
func TestDecryptModernWrongPassword(t *testing.T) {
	correctPassword := "correct_master_password"
	plaintext := "secret_password_123"

	// 使用正确密码加密
	encrypted := cfb8EncryptForTest(t, plaintext, correctPassword)
	encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)

	// 用正确密码解密应该成功
	result, err := decryptModern(encryptedBase64, correctPassword)
	if err != nil {
		t.Fatalf("正确密码解密失败: %v", err)
	}
	if result != plaintext {
		t.Errorf("正确密码解密结果期望 %q，得到 %q", plaintext, result)
	}

	// 用错误密码解密应该得到不同的结果
	wrongResult, err := decryptModern(encryptedBase64, "wrong_password")
	if err != nil {
		// 如果返回错误也是可接受的
		return
	}
	if wrongResult == plaintext {
		t.Error("错误密码解密结果不应与正确结果相同")
	}
}

// TestCFB8Decrypt 使用已知测试向量验证 CFB-8 解密
func TestCFB8Decrypt(t *testing.T) {
	masterPassword := "test_password"
	plaintext := "hello world 123"

	// 先加密
	ciphertext := cfb8EncryptForTest(t, plaintext, masterPassword)

	// 再解密
	hash := sha512.Sum512([]byte(masterPassword))
	key := hash[:32]
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("创建 AES cipher 失败: %v", err)
	}
	iv := make([]byte, aes.BlockSize)
	block.Encrypt(iv, iv)

	decrypted := cfb8Decrypt(block, iv, ciphertext)

	if string(decrypted) != plaintext {
		t.Errorf("CFB-8 解密结果期望 %q，得到 %q", plaintext, string(decrypted))
	}
}

// TestCFB8DecryptEmptyInput 测试空输入
func TestCFB8DecryptEmptyInput(t *testing.T) {
	hash := sha512.Sum512([]byte("password"))
	key := hash[:32]
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("创建 AES cipher 失败: %v", err)
	}
	iv := make([]byte, aes.BlockSize)
	block.Encrypt(iv, iv)

	result := cfb8Decrypt(block, iv, []byte{})
	if len(result) != 0 {
		t.Errorf("空输入解密结果期望长度 0，得到 %d", len(result))
	}
}

// TestDecryptModernInvalidBase64 测试无效 Base64 输入
func TestDecryptModernInvalidBase64(t *testing.T) {
	_, err := decryptModern("not-valid-base64!!!", "password")
	if err == nil {
		t.Error("无效 Base64 应返回错误")
	}
}

// TestDecryptModernEmptyData 测试空加密数据
func TestDecryptModernEmptyData(t *testing.T) {
	// Base64 编码的空数据
	emptyBase64 := base64.StdEncoding.EncodeToString([]byte{})
	_, err := decryptModern(emptyBase64, "password")
	if err == nil {
		t.Error("空加密数据应返回错误")
	}
}

// TestParseSessionLineInvalidFormat 测试无效格式
func TestParseSessionLineInvalidFormat(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"无等号", "no-equals-sign"},
		{"空值少字段", "name="},
		{"仅一个字段", "name=onlytype"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSessionLine(tt.line, "/tmp/test.ini", "")
			if err == nil {
				t.Error("期望返回错误，但没有")
			}
		})
	}
}

// cfb8EncryptForTest 是用于测试的 CFB-8 加密函数
// 与 cfb8Decrypt 配对使用
func cfb8EncryptForTest(t *testing.T, plaintext, masterPassword string) []byte {
	t.Helper()

	hash := sha512.Sum512([]byte(masterPassword))
	key := hash[:32]

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("创建 AES cipher 失败: %v", err)
	}

	// IV = AES-ECB 加密 16 字节全零（与 decryptModern 一致）
	iv := make([]byte, aes.BlockSize)
	block.Encrypt(iv, iv)

	// CFB-8 加密
	plaintextBytes := []byte(plaintext)
	ciphertext := make([]byte, len(plaintextBytes))
	shiftReg := make([]byte, aes.BlockSize)
	copy(shiftReg, iv)

	output := make([]byte, aes.BlockSize)

	for i := 0; i < len(plaintextBytes); i++ {
		block.Encrypt(output, shiftReg)
		ciphertext[i] = plaintextBytes[i] ^ output[0]
		// 左移 shift register，填入密文字节（加密时也填密文字节）
		copy(shiftReg, shiftReg[1:])
		shiftReg[aes.BlockSize-1] = ciphertext[i]
	}

	return ciphertext
}
