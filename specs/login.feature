# Feature: 用户登录
# 规范文档 - 使用 Gherkin 语法描述功能

Feature: 用户登录功能
  作为用户
  我希望能够使用用户名和密码登录
  以便访问受保护的功能

  Scenario: 使用有效凭据登录成功
    Given 用户已注册
      And 用户名是 "alice"
      And 密码是 "secret123"
    When 用户尝试登录
    Then 登录应该成功
      And 返回用户令牌

  Scenario: 使用无效密码登录失败
    Given 用户已注册
      And 用户名是 "alice"
      And 密码是 "wrongpassword"
    When 用户尝试登录
    Then 登录应该失败
      And 返回错误信息 "密码错误"

  Scenario: 使用不存在的用户名登录失败
    Given 用户名是 "nonexistent"
      And 密码是 "anypassword"
    When 用户尝试登录
    Then 登录应该失败
      And 返回错误信息 "用户不存在"
