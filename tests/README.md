# 测试目录

本目录包含基于规范的自动化测试。

## 测试结构

- 测试文件应该与规范文件一一对应
- 使用 BDD 框架将自然语言规范映射到测试代码
- 测试名称应该清晰表达测试的意图

## 示例 (Python + pytest-bdd)

```python
from pytest_bdd import given, when, then, scenario

@scenario('../specs/login.feature', '使用有效凭据登录成功')
def test_login_success():
    pass

@given('用户已注册')
def user_registered():
    # 设置测试数据
    pass

@when('用户尝试登录')
def user_attempts_login():
    # 执行登录操作
    pass

@then('登录应该成功')
def login_succeeds():
    # 验证结果
    assert response.status_code == 200
```
