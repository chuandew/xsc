package xftp

import (
	"testing"
)

// TestValidateFileName 测试路径穿越防护
func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"正常文件名", "myfile.txt", false},
		{"正常目录名", "mydir", false},
		{"包含斜杠", "path/file", true},
		{"包含双点", "..hidden", true},
		{"纯双点", "..", true},
		{"路径穿越", "../etc/passwd", true},
		{"绝对路径", "/etc/passwd", true},
		{"中文文件名", "测试文件.txt", false},
		{"带空格文件名", "my file.txt", false},
		{"隐藏文件", ".gitignore", false},
		{"带点的文件名", "file.tar.gz", false},
		{"中间含双点", "a..b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFileName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
