# 错误码手册

> clix 全部错误使用 errx 结构化错误码；错误码说明经 `errx.Describe`
> 可查，日志经 `errx/logx.Fields` 输出。

## 退出码约定

| 退出码 | 含义 |
| --- | --- |
| 0 | 成功（含帮助/版本） |
| 1 | 命令执行失败（Action/钩子错误、未捕获异常） |
| 2 | 用法错误（命令/参数/flag/校验） |
| 130 | 上下文取消 |

## 用法错误码（退出码 2）

| 错误码 | 含义 |
| --- | --- |
| `CLI_INVALID_APP` | 应用/命令配置非法 |
| `CLI_MISSING_COMMAND` | 缺少命令或子命令 |
| `CLI_UNKNOWN_COMMAND` | 未知命令或子命令 |
| `CLI_INVALID_FLAG_DEF` | flag 定义非法（名/默认值/环境变量/规则） |
| `CLI_INVALID_ARG_DEF` | 位置参数定义非法 |
| `CLI_UNKNOWN_FLAG` | 未知 flag（含命令前未知全局 flag） |
| `CLI_DUPLICATE_FLAG` | 非重复 flag 重复指定 |
| `CLI_MISSING_FLAG_VALUE` | flag 缺少值 |
| `CLI_INVALID_FLAG_VALUE` | flag 值类型非法 |
| `CLI_MISSING_REQUIRED_FLAG` | 缺少必填 flag（命令行/环境变量均未提供） |
| `CLI_INVALID_ENUM_VALUE` | 枚举值不在允许列表 |
| `CLI_MISSING_ARG` | 缺少必填位置参数 |
| `CLI_TOO_MANY_ARGS` | 位置参数过多 |
| `CLI_FLAG_VALIDATION_FAILED` | flag 值未通过 validx 规则校验 |

## 非用法错误码

| 错误码 | 分类 | 退出码 |
| --- | --- | --- |
| `CLI_CANCELLED` | cancelled | 130 |
| `CLI_ACTION_PANIC` | internal | 1 |

## 设计约定

- Action/钩子返回的错误**原样透传**，不包装、不改写；
- 可重试错误（`errx.Retryable`）日志级别为 Warn，其余失败为 Error；
- 错误码提示可通过 `WithErrorHint` 注册。
