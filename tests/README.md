# 测试目录

本目录包含基于规范的自动化测试。

## 测试结构

- 测试文件应该与规范文件一一对应
- 使用 BDD 框架将自然语言规范映射到测试代码
- 测试名称应该清晰表达测试的意图

## 建议的测试框架

对于 Go 项目，可以使用以下测试框架：

1. **标准库 testing** - Go 内置的测试框架
2. **testify** - 提供断言和 mock 功能
3. **Ginkgo + Gomega** - BDD 风格的测试框架

## 测试示例

测试文件应该放在与源代码相同的包中，命名为 `*_test.go`。

```go
package session

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestLoadSession(t *testing.T) {
    session, err := LoadSession("/path/to/session.yaml")
    assert.NoError(t, err)
    assert.Equal(t, "192.168.1.100", session.Host)
    assert.Equal(t, 22, session.Port)
}
```

## 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./internal/session/...

# 运行特定测试函数
go test -run TestLoadSession ./internal/session/...

# 带覆盖率报告
go test -v -cover ./...
```

## 测试规范对应

- `../specs/xssh.feature` - XSSH 功能规范
- 每个 Scenario 应该对应一个或多个测试用例
- 保持测试与规范同步更新
