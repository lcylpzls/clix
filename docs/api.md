# API 快照

> 随版本更新。v0.1.0 快照如下；新版本发布后同步替换。

## v0.1.0

### 类型

```go
type App struct{ /* 不可直接构造 */ }
type Command struct {
    Name        string     // 命令名（必填，不可为空或以 - 开头）
    Description string     // 帮助中的命令描述
    Usage       string     // 可选；为空时生成 "app 命令 [参数...]"
    Action      ActionFunc // 执行函数（必填）
}
type Context struct {
    App     *App
    Command *Command // 根 Action 时为 nil
    Args    []string // 命令名之后的原始参数
}
type ActionFunc func(ctx context.Context, c *Context) error
type Option func(*appConfig)
```

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
- `--version`：版本行 `名称 版本`，退出码 0；
- 未知命令、缺少命令：错误写 Err，退出码 2。

