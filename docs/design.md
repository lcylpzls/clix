# 设计

## 1. 定位与范围

clix 是家族生态的 CLI 框架，定位：

- 轻量、确定性、可测试；
- 与 errx（错误码/退出码）、logx（结构化日志）、confx（配置）、
  validx（参数校验）天然打通；
- 自用协议，不兼容第三方 CLI 框架。

范围：

- App / Command 模型、帮助与版本、命令分发；
- 参数解析（位置参数、变参、flag、必填/默认/枚举）；
- 子命令树、别名、分组；
- 环境变量、全局 flag、生命周期钩子；
- confx + validx 联动示例；
- 退出码映射与结构化错误输出。

非目标：

- 插件体系与动态加载；
- 第三方 CLI 框架兼容层；
- shell 补全脚本生成（按需评估）；
- GUI / TUI。

## 2. 核心模型

```text
App
├── name / version / description / usage
├── out / err 输出流（可注入）
├── logger（可选，logx.Logger）
├── root（可选，根级 Action）
└── commands：Command 有序表
     └── Command
          ├── name / description / usage
          ├── args / flags（v0.2.0 起）
          ├── aliases / group（v0.3.0 起）
          ├── env / hooks（v0.4.0 起）
          └── Action(ctx, *Context) error
```

`Context` 携带 App、当前 Command 与参数；Action 只返回 error，
退出码由 clix 统一映射。

## 3. 执行流程

```text
Execute(ctx, args)
  └─ dispatch
       ├─ ctx 已取消 → CLI_CANCELLED（退出码 130）
       ├─ 空参数且无根 Action → CLI_MISSING_COMMAND（退出码 2）
       ├─ -h/--help/help [命令] → 帮助到 Out，退出码 0
       ├─ --version → 版本到 Out，退出码 0
       ├─ 未知命令 → CLI_UNKNOWN_COMMAND（退出码 2）
       └─ 命中命令/根 Action
            ├─ 正常返回 → 退出码 0
            ├─ 返回 error → 输出到 Err，退出码 1
            └─ panic → 恢复为 CLI_ACTION_PANIC，退出码 1
```

## 4. 错误与退出码

| 错误码 | 分类 | 含义 | 退出码 |
| --- | --- | --- | --- |
| `CLI_INVALID_APP` | invalid | 应用/命令配置非法 | 2 |
| `CLI_MISSING_COMMAND` | invalid | 缺少命令 | 2 |
| `CLI_UNKNOWN_COMMAND` | invalid | 未知命令 | 2 |
| `CLI_CANCELLED` | cancelled | 上下文取消 | 130 |
| `CLI_ACTION_PANIC` | internal | 命令未捕获异常 | 1 |
| 用户 Action 返回的错误 | 原样保留 | 命令执行失败 | 1 |

用户 Action 返回的错误**原样透传**，不包装、不改写，保证 `errors.Is/As`
链路完整；仅当其不是 errx 错误时，日志层按普通错误记录。

## 5. 输出与日志

- 帮助与版本写 `Out`；错误写 `Err`；
- `WithIO` 可注入任意 `io.Writer`，库自身不直接依赖 stdout/stderr 默认值之外的全局状态；
- `Logger` 可选：设置后记录命令开始（Debug）、成功（Info）、
  用法错误（Warn）、执行失败/取消/恐慌（Error），字段含
  command / args / duration_ms 与 errx 结构化字段。

## 6. 版本与兼容

- 语义化版本；pre-1.0 允许按路线图演进 API；
- 每个里程碑版本完成即发布 tag，CI 全绿后由 Release 工作流生成 Release；
- v1.0.0 起 API 冻结，破坏性变更必须提升主版本；
- 依赖约定：`errx` / `logx` / `validx` 跟随 1.x；`confx` 锁定 `v0.3.3`
  （`LoadConfig` 仅使用其稳定子集），confx 发布 1.0 后评审升级；
- v1.0.0 是否发布由维护者决定，clix 只推进到 1.0 候选即停。
