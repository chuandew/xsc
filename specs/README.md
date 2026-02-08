# Spec Driven Development 工作流

## 工作步骤

1. **编写规范** (specs/)
   - 使用自然语言描述功能
   - 定义验收标准
   - 示例: login.feature

2. **编写测试** (tests/)
   - 基于规范编写测试代码
   - 使用 BDD 框架（如 pytest-bdd, behave）
   - 确保测试能表达规范的意图

3. **实现功能** (src/)
   - 编写最少的代码让测试通过
   - 重构代码保持整洁
   - 不要超越规范实现额外功能

## 目录结构

```
├── specs/          # 规范文档 (.feature 文件)
├── tests/          # 测试代码
└── src/            # 实现代码
```

## 建议工具

- **Python**: pytest-bdd, behave
- **JavaScript**: Cucumber.js, Jest
- **Java**: Cucumber JVM

## 工作流命令

```bash
# 1. 阅读规范
cat specs/*.feature

# 2. 运行测试
pytest tests/  # Python
npm test       # JavaScript

# 3. 实现功能
# 编辑 src/ 下的文件

# 4. 重构代码
# 在测试通过后优化实现
```

## 基本原则

- ✅ 规范先于代码
- ✅ 测试基于规范
- ✅ 实现满足测试即可
- ❌ 不要过度设计
- ❌ 不要实现规范外的功能
