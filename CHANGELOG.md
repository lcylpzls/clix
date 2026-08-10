# 更新日志

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
