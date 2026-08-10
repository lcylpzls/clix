# 更新日志

## [v0.6.2] - 2026-08-10

### 文档与约定

- API 快照同步至 v0.6.2；
- 明确 `CommandHelpText(name)` 仅接受顶层命令名，完整路径帮助使用
  `help 路径` 子命令；
- 明确依赖约定：`confx` 锁定 `v0.3.3`，`LoadConfig` 仅依赖其稳定子集，
  confx 发布 1.0 后统一评审升级；
- 设计文档补充 v1.0.0 API 冻结承诺。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v0.6.1] - 2026-08-10

### 修复

- 命令名含空白时，别名与去空格后的命令名相同不再漏检；
- `registered` 标记改为原子操作，并发注册更安全；
- `Strings` / `GlobalStrings` 返回副本，防止调用方修改内部存储。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v0.6.0] - 2026-08-10

### 修复

- 全局 flag 此前未执行 validx 规则校验，现与命令 flag 一致在解析期校验；
- 带校验规则的 flag 默认值现在注册期即校验，配置错误尽早暴露。

### 改进

- 帮助中 flag 的必填/默认/枚举/校验标记统一为括号包裹；
- 新增基准测试（解析 ~1µs/次、分发 ~1µs/次）与 `BENCHMARKS.md`；
- 新增 `docs/errors.md` 错误码手册与 `docs/final-review.md`
  1.0 候选终审清单；
- 新增 Issue 模板与 README CI 徽章。

### 结论

- clix 达到 1.0 候选标准；**v1.0.0 是否发布由维护者决定**。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v0.5.0] - 2026-08-10

### 新增

- validx 联动：`FlagSpec.Validate(rules)` 绑定 validx 规则串，
  注册期预编译规则语法，解析期逐值校验；失败归一为
  `CLI_FLAG_VALIDATION_FAILED`（用法错误，退出码 2）；
- confx 联动：`LoadConfig` 助手在 Before 钩子中加载配置，
  路径优先取 flag（默认 `config`）值，其次使用默认路径；
- 错误提示：`WithErrorHint(code, hint)` 为错误码注册提示，
  Action 失败时输出"提示：..."一行；
- 日志级别联动：可重试错误（errx.Retryable）记录 Warn 级别
  "命令执行失败（可重试）"，其余失败保持 Error；
- 帮助文本展示 flag 校验规则；
- 示例新增 `config` 命令：confx TOML 加载 + validx 规则校验 +
  环境变量路径。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v0.4.0] - 2026-08-10

### 新增

- flag 环境变量绑定：`FlagSpec.Env`，取值优先级
  命令行 > 环境变量 > 默认值；可重复 flag 使用逗号分隔；
- 全局 flag：`WithGlobalFlags`，位于命令名之前，`Context.Global*`
  访问器读取；帮助文本新增"全局选项"区块；
- 生命周期钩子：`Command.Before` / `Command.After`：
  - Before 失败则中止 Action 与 After；
  - After 在 Action 失败后仍执行；
  - 两者同时失败时合并为 errx 聚合错误；
- 全局 flag 支持环境变量与必填校验，未知/非法统一为用法错误码。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v0.3.0] - 2026-08-10

### 新增

- 子命令树：`Command.AddCommand` 任意深度嵌套，`Parent` / `FullName`
  访问器，`help parent child` 完整路径帮助；
- 命令别名：顶层与子命令均支持，帮助列表展示别名，分发按别名命中；
- 帮助分组：`Command.Group` 分组渲染（空分组无标题，组内保持注册顺序）；
- 命令可同时拥有 Action 与子命令：无子命令参数时执行 Action，
  命中子命令时执行子命令；
- 无 Action 命令的缺少/未知子命令归一为 `CLI_MISSING_COMMAND` /
  `CLI_UNKNOWN_COMMAND`（退出码 2）；
- 命令帮助新增别名、子命令区块；命令名校验收紧为合法 ASCII 名称。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v0.2.0] - 2026-08-10

### 新增

- 类型化参数解析：
  - `ArgSpec` 位置参数（必填/变参，变参只能是最后一个）；
  - `FlagSpec` 八种类型：string / bool / int / int64 / float / duration /
    enum / string[]，支持必填、默认值与枚举允许列表；
  - `--name=值` 与 `--name 值` 两种写法，`--` 终止 flag 解析；
  - 未知 flag、重复 flag、缺少值、类型非法、缺少必填、枚举越界、
    位置参数数量错误均归一为 errx 用法错误码（退出码 2）；
- Context 类型化访问器：`String` / `Bool` / `Int` / `Int64` / `Float64` /
  `Duration` / `Enum` / `Strings` / `HasFlag`；
- 命令帮助自动渲染参数与选项（含必填/默认/枚举说明）；
- 命令参数中的 `-h` / `--help` 显示命令帮助且不执行 Action；
- `FuzzParseCommandArgs` fuzz 目标接入 CI；
- 示例新增 `sum` 命令（变参 + int/enum flag）。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v0.1.0] - 2026-08-10

### 新增

- App/Command 骨架：`New` / `AddCommand` / 根 Action / Context；
- 内置命令约定：`-h` / `--help` / `help [命令]` / `--version`；
- errx 结构化错误码（`CLI_INVALID_APP` / `CLI_MISSING_COMMAND` /
  `CLI_UNKNOWN_COMMAND` / `CLI_CANCELLED` / `CLI_ACTION_PANIC`）
  与退出码映射（0/1/2/130）；
- logx 可选注入：命令开始（Debug）、成功（Info）、用法错误（Warn）、
  失败/取消/恐慌（Error）结构化日志；
- Out/Err 可注入；帮助文本返回字符串便于快照测试；
- Action panic 恢复为结构化错误，不向调用方泄漏崩溃栈；
- fuzz 目标（`FuzzExecute` / `FuzzCommandHelpText`）接入 CI；
- 三平台 CI + Linux 多发行版容器矩阵 + 示例模块 + Release 工作流。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。
