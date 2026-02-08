# 源代码目录

本目录包含 XSC 功能的实际实现。

## 开发原则

1. **只实现规范要求的功能**
   - 不要添加规范之外的特性
   - 保持代码简单直接

2. **代码应该能通过所有测试**
   - 运行测试确保实现正确
   - 如果测试失败，修复代码而非测试

3. **持续重构**
   - 在测试通过后优化代码结构
   - 保持代码整洁和可维护性

## 项目结构

```
├── cmd/xsc/           # 应用程序入口
│   └── main.go        # 主程序
├── internal/          # 内部包
│   ├── session/       # 会话管理
│   ├── ssh/          # SSH 连接
│   ├── tui/          # TUI 界面
│   ├── securecrt/    # SecureCRT 集成
│   └── tree/         # 树形数据结构
└── pkg/config/       # 公共配置包
    └── config.go     # 配置管理
```

## 包说明

- **cmd/xsc** - 命令行入口，处理参数解析和命令分发
- **internal/session** - 会话管理（加载、保存、验证）
- **internal/ssh** - SSH 客户端实现（纯 Go）
- **internal/tui** - Bubble Tea TUI 实现
- **internal/securecrt** - SecureCRT 会话解析和解密
- **pkg/config** - 全局配置管理

## 编码规范

- 使用标准 Go 项目布局
- 包名使用小写，简短且有意义
- 导出函数使用 PascalCase
- 私有函数使用 camelCase
- 添加适当的注释说明功能
