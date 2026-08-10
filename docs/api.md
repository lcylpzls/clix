# API 快照

> 随版本更新。v0.2.0 快照如下；新版本发布后同步替换。

## v0.2.0

### 类型

```go
type App struct{ /* 不可直接构造 */ }
type Command struct {
    Name        string     // 命令名（必填，不可为空或以 - 开头）
    Description string     // 帮助中的命令描述
    Usage       string     // 可选；为空时生成 "app 命令 [选项...] [参数...]"
    Args        []ArgSpec  // 位置参数声明；nil 表示原样透传
    Flags       []FlagSpec // flag 声明；nil 表示不做解析
    Action      ActionFunc // 执行函数（必填）
}
type Context struct {
    App     *App
    Command *Command // 根 Action 时为 nil
    Args    []string // 解析后的位置参数（未声明时原样透传）
    Flags   FlagValues
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
```

默认值规则：字符串/枚举为 string，布尔为 bool，整数为 int/int64，
浮点为 float64/int/int64，时长为 time.Duration，可重复 flag 为 []string；
类型不匹配在注册期报错。枚举默认值必须在允许列表内。

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
返回 true。

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

- `-h` / `--help` / `help [命令]`：帮助，退出码 0；
- 命令参数中出现的 `-h` / `--help` 显示命令帮助且不执行 Action；
- `--` 终止 flag 解析，其后内容全部视为位置参数；
- `--version`：版本行 `名称 版本`，退出码 0；
- 未知命令、缺少命令：错误写 Err，退出码 2。

### v0.2.0 错误码

在 v0.1.0 错误码基础上新增：

| 错误码 | 含义 |
| --- | --- |
| `CLI_INVALID_FLAG_DEF` / `CLI_INVALID_ARG_DEF` | flag/参数定义非法 |
| `CLI_UNKNOWN_FLAG` | 未知 flag |
| `CLI_DUPLICATE_FLAG` | 非重复 flag 重复指定 |
| `CLI_MISSING_FLAG_VALUE` | flag 缺少值 |
| `CLI_INVALID_FLAG_VALUE` | flag 值类型非法 |
| `CLI_MISSING_REQUIRED_FLAG` | 缺少必填 flag |
| `CLI_INVALID_ENUM_VALUE` | 枚举值不在允许列表 |
| `CLI_MISSING_ARG` / `CLI_TOO_MANY_ARGS` | 位置参数数量错误 |

以上均为用法错误，退出码 2。
