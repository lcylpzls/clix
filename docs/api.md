# API 快照

> v1.0.0 起为 API 冻结快照；破坏性变更需提升主版本。

## v1.0.0

### 类型

```go
type App struct{ /* 不可直接构造 */ }
type Command struct {
    Name        string     // 命令名（必填，须合法且不与同级冲突）
    Description string     // 帮助中的命令描述
    Usage       string     // 可选；为空时生成 "app 路径 [选项...] [参数...]"
    Args        []ArgSpec  // 位置参数声明；nil 表示原样透传
    Flags       []FlagSpec // flag 声明；nil 表示不做解析
    Aliases     []string   // 命令别名（须合法且不与同级冲突）
    Group       string     // 帮助列表分组名；空为默认组
    Action      ActionFunc // 执行函数；与子命令至少提供一个
    Before      ActionFunc // Action 前钩子；失败则中止 Action 与 After
    After       ActionFunc // Action 后钩子；Action 失败也执行
}
type Context struct {
    App     *App
    Command *Command // 根 Action 时为 nil
    Args    []string // 解析后的位置参数（未声明时原样透传）
    Flags   FlagValues
    Global  FlagValues // 全局 flag 值
}
type ActionFunc func(ctx context.Context, c *Context) error
type Option func(*appConfig)

type ArgSpec struct {
    Name        string // 参数名
    Description string // 帮助说明
    Required    bool   // 必填
    Variadic    bool   // 变参（只能是最后一个）
}

type FlagSpec struct{ /* 通过构造函数创建，配合 Required()/Default() 链式配置 */ }
type FlagValues struct{ /* 通过 Context 类型化访问器读取 */ }
```

### 命令树

```go
func (c *Command) AddCommand(sub *Command) error // 注册子命令（任意深度）
func (c *Command) Parent() *Command              // 父命令；顶层为 nil
func (c *Command) FullName() string              // 完整路径，如 "parent child"
```

- 命令与子命令均支持别名；别名须为合法 ASCII 名称，不得与同级命令名/
  别名冲突；
- 命令可以同时拥有 Action 与子命令：无子命令参数时执行 Action，
  命中子命令时执行子命令；
- 无 Action 的命令必须至少有一个子命令；执行时缺少/未知子命令
  分别返回 `CLI_MISSING_COMMAND` / `CLI_UNKNOWN_COMMAND`（退出码 2）。

### flag 构造函数

```go
func StringFlag(name, usage string) FlagSpec
func BoolFlag(name, usage string) FlagSpec
func IntFlag(name, usage string) FlagSpec
func Int64Flag(name, usage string) FlagSpec
func FloatFlag(name, usage string) FlagSpec
func DurationFlag(name, usage string) FlagSpec
func EnumFlag(name, usage string, allowed ...string) FlagSpec
func StringSliceFlag(name, usage string) FlagSpec
func (f FlagSpec) Required() FlagSpec
func (f FlagSpec) Default(v any) FlagSpec
func (f FlagSpec) Env(name string) FlagSpec
func (f FlagSpec) Validate(rules string) FlagSpec
```

默认值规则：字符串/枚举为 string，布尔为 bool，整数为 int/int64，
浮点为 float64/int/int64，时长为 time.Duration，可重复 flag 为 []string；
类型不匹配在注册期报错。枚举默认值必须在允许列表内。

取值优先级：命令行 > 环境变量 > 默认值；可重复 flag 的环境变量值
使用逗号分隔并去除空白。环境变量同样满足必填校验。

`Validate` 绑定 validx 规则串（如 `required,min=3`、`oneof=dev prod`）；
规则语法与默认值在注册期预编译校验，解析期逐值校验；失败返回
`CLI_FLAG_VALIDATION_FAILED`（退出码 2）。全局 flag 同样参与校验。

### Context 访问器

```go
func (c *Context) HasFlag(name string) bool
func (c *Context) String(name string) string
func (c *Context) Bool(name string) bool
func (c *Context) Int(name string) int
func (c *Context) Int64(name string) int64
func (c *Context) Float64(name string) float64
func (c *Context) Duration(name string) time.Duration
func (c *Context) Enum(name string) string
func (c *Context) Strings(name string) []string
```

未声明或未指定的 flag 返回类型零值；`HasFlag` 仅对显式指定的 flag
返回 true。`Strings` / `GlobalStrings` 返回副本，修改不影响内部存储。

### 全局 flag

```go
func WithGlobalFlags(flags ...FlagSpec) Option
func (c *Context) HasGlobalFlag(name string) bool
func (c *Context) GlobalString(name string) string
func (c *Context) GlobalBool(name string) bool
func (c *Context) GlobalInt(name string) int
func (c *Context) GlobalInt64(name string) int64
func (c *Context) GlobalFloat64(name string) float64
func (c *Context) GlobalDuration(name string) time.Duration
func (c *Context) GlobalEnum(name string) string
func (c *Context) GlobalStrings(name string) []string
```

- 全局 flag 必须位于命令名之前连续出现；命令名之后出现同名 flag
  视为该命令的本地 flag；
- 全局 flag 支持全部类型、默认值、环境变量与必填校验；
- 帮助文本展示"全局选项"区块。

### confx 联动

```go
func LoadConfig(ctx context.Context, c *Context, manager *confx.ConfigManager,
    pathFlag, fallback string, target any) error
```

- 适合在 `Before` 钩子中调用；路径优先取 flag `pathFlag`
  （默认 `config`）的值，其次使用 `fallback`；
- 加载错误透传 confx 结构化错误。

### 依赖约定（v1.0.0 起冻结）

- `errx` / `logx` / `validx` 跟随家族 1.x 版本；
- `confx` 当前锁定 `v0.3.3`：`LoadConfig` 仅依赖其稳定子集
  （`NewConfigManager` / `Load`），confx 发布 1.0 后统一评审升级；
- 冻结后破坏性 API 变更需提升主版本。

### 错误提示与日志级别

```go
func WithErrorHint(code errx.Code, hint string) Option
```

- Action 返回注册过提示的错误码时，错误消息后输出"提示：..."；
- 可重试错误（errx.Retryable）日志级别为 Warn，其余失败为 Error。

### 生命周期钩子

- `Before` 在参数解析后、Action 前运行；返回错误则中止 Action 与 After；
- `After` 在 Action 后运行，Action 失败时仍执行；
- Action 与 After 同时失败时返回 errx 聚合错误（`errx.Join`）；
- 钩子与 Action 的 panic 统一恢复为 `CLI_ACTION_PANIC`。

### 构造与注册

```go
func New(name, version string, opts ...Option) (*App, error)
func (a *App) AddCommand(cmd *Command) error
```

### 选项

```go
func WithDescription(desc string) Option
func WithUsage(usage string) Option
func WithIO(out, err io.Writer) Option
func WithLogger(logger logx.Logger) Option
func WithRootAction(action ActionFunc) Option
```

### 执行

```go
func (a *App) Run(ctx context.Context, args []string) error // 不打印错误、不映射退出码
func (a *App) Execute(ctx context.Context, args []string) int
```

### 帮助

```go
func (a *App) HelpText() string
func (a *App) CommandHelpText(name string) (string, error)
```

- 命令列表按 `Group` 分组（空分组无标题），组内保持注册顺序；
- 命令帮助展示别名、子命令、参数与选项；
- `help parent child` 支持完整路径；
- `CommandHelpText(name)` 仅接受**顶层命令名**；完整路径帮助请使用
  `help 路径` 子命令或 `Run` / `Execute`。

### 访问器

```go
func (a *App) Name() string
func (a *App) Version() string
func (a *App) Description() string
func (a *App) Out() io.Writer
func (a *App) Err() io.Writer
func (a *App) Logger() logx.Logger
func (c *Context) Out() io.Writer
func (c *Context) Err() io.Writer
func (c *Context) Logger() logx.Logger
```

### 退出码

```go
const (
    ExitOK        = 0
    ExitFailure   = 1
    ExitUsage     = 2
    ExitCancelled = 130
)
```

### 内置命令约定

- `-h` / `--help` / `help [路径...]`：帮助，退出码 0；
- 命令参数中出现的 `-h` / `--help` 显示该命令帮助且不执行 Action；
- `--` 终止 flag 解析，其后内容全部视为位置参数；
- `--version`：版本行 `名称 版本`，退出码 0；
- 未知命令、缺少命令：错误写 Err，退出码 2。

### 错误码

全部错误码（退出码 2 的用法错误）：

| 错误码 | 含义 |
| --- | --- |
| `CLI_INVALID_APP` | 应用/命令配置非法 |
| `CLI_MISSING_COMMAND` | 缺少命令或子命令 |
| `CLI_UNKNOWN_COMMAND` | 未知命令或子命令 |
| `CLI_INVALID_FLAG_DEF` / `CLI_INVALID_ARG_DEF` | flag/参数定义非法 |
| `CLI_UNKNOWN_FLAG` | 未知 flag |
| `CLI_DUPLICATE_FLAG` | 非重复 flag 重复指定 |
| `CLI_MISSING_FLAG_VALUE` | flag 缺少值 |
| `CLI_INVALID_FLAG_VALUE` | flag 值类型非法 |
| `CLI_MISSING_REQUIRED_FLAG` | 缺少必填 flag |
| `CLI_INVALID_ENUM_VALUE` | 枚举值不在允许列表 |
| `CLI_MISSING_ARG` / `CLI_TOO_MANY_ARGS` | 位置参数数量错误 |

非用法错误码：`CLI_CANCELLED`（退出码 130）、`CLI_ACTION_PANIC`（退出码 1）。
