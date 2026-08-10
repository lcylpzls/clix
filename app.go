package clix

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/logx"
)

// App 是 CLI 应用入口，持有名称、版本、输出流、日志器与命令表。
//
// 通过 New 构造；构造后命令表可变（AddCommand），其余字段不可变。
type App struct {
	name        string
	version     string
	description string
	usage       string
	out         io.Writer
	err         io.Writer
	logger      logx.Logger
	root        ActionFunc

	rootRegistry *registry
	globalFlags  []FlagSpec
	errorHints   map[errx.Code]string
}

// Option 修改 App 构造配置。
type Option func(*appConfig)

// appConfig 是 App 的构造配置。
type appConfig struct {
	description string
	usage       string
	out         io.Writer
	err         io.Writer
	logger      logx.Logger
	root        ActionFunc
	globalFlags []FlagSpec
	errorHints  map[errx.Code]string
}

// WithDescription 设置应用描述，显示在帮助首行与命令列表中。
func WithDescription(desc string) Option {
	return func(c *appConfig) {
		c.description = desc
	}
}

// WithUsage 覆盖应用级用法行；为空时使用默认 "app [命令] [参数...]"。
func WithUsage(usage string) Option {
	return func(c *appConfig) {
		c.usage = usage
	}
}

// WithIO 注入标准输出与错误输出流；默认分别为 os.Stdout 与 os.Stderr。
// 两者均不可为空。
func WithIO(out, err io.Writer) Option {
	return func(c *appConfig) {
		c.out = out
		c.err = err
	}
}

// WithLogger 注入 logx.Logger；未注入时日志路径静默。
func WithLogger(logger logx.Logger) Option {
	return func(c *appConfig) {
		c.logger = logger
	}
}

// WithRootAction 设置根级 Action：无参数调用时执行，此时不需要子命令。
func WithRootAction(action ActionFunc) Option {
	return func(c *appConfig) {
		c.root = action
	}
}

// WithGlobalFlags 声明应用级全局 flag。全局 flag 必须位于命令名之前，
// 值通过 Context 的 Global* 访问器读取。
func WithGlobalFlags(flags ...FlagSpec) Option {
	return func(c *appConfig) {
		c.globalFlags = append([]FlagSpec(nil), flags...)
	}
}

// WithErrorHint 为指定错误码注册错误提示；Action 返回该错误码时，
// 错误消息后追加一行"提示：..."。
func WithErrorHint(code errx.Code, hint string) Option {
	return func(c *appConfig) {
		if c.errorHints == nil {
			c.errorHints = make(map[errx.Code]string)
		}
		c.errorHints[code] = hint
	}
}

// New 构造 App。name 与 version 必须非空；配置非法时返回 errx 错误。
func New(name, version string, opts ...Option) (*App, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errx.NewCode(CodeInvalidApp, "应用名不能为空")
	}
	if strings.TrimSpace(version) == "" {
		return nil, errx.NewCode(CodeInvalidApp, "版本号不能为空")
	}
	cfg := appConfig{out: os.Stdout, err: os.Stderr}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.out == nil {
		return nil, errx.NewCode(CodeInvalidApp, "标准输出流不能为空")
	}
	if cfg.err == nil {
		return nil, errx.NewCode(CodeInvalidApp, "错误输出流不能为空")
	}
	if err := validateFlagSpecs(cfg.globalFlags); err != nil {
		return nil, err
	}
	return &App{
		name:         strings.TrimSpace(name),
		version:      strings.TrimSpace(version),
		description:  cfg.description,
		usage:        strings.TrimSpace(cfg.usage),
		out:          cfg.out,
		err:          cfg.err,
		logger:       cfg.logger,
		root:         cfg.root,
		rootRegistry: &registry{},
		globalFlags:  cfg.globalFlags,
		errorHints:   cfg.errorHints,
	}, nil
}

// AddCommand 注册子命令。命令名必须非空、不得以 "-" 开头且不能重复；
// Action 必须非空。失败返回 errx 错误，注册顺序即帮助列表顺序。
func (a *App) AddCommand(cmd *Command) error {
	return a.rootRegistry.add(cmd)
}

// Name 返回应用名。
func (a *App) Name() string {
	return a.name
}

// Version 返回版本号。
func (a *App) Version() string {
	return a.version
}

// Description 返回应用描述。
func (a *App) Description() string {
	return a.description
}

// Out 返回标准输出流。
func (a *App) Out() io.Writer {
	return a.out
}

// Err 返回错误输出流。
func (a *App) Err() io.Writer {
	return a.err
}

// Logger 返回日志器；未注入时返回 nil。
func (a *App) Logger() logx.Logger {
	return a.logger
}

// lookup 按名称查找命令；未找到返回 nil。并发安全。
func (a *App) lookup(name string) *Command {
	return a.rootRegistry.lookup(name)
}

// commandList 返回按注册顺序排列的命令副本。并发安全。
func (a *App) commandList() []*Command {
	return a.rootRegistry.list()
}

// usageLine 返回应用级用法行。
func (a *App) usageLine() string {
	if a.usage != "" {
		return a.usage
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s", a.name)
	if a.root != nil {
		b.WriteString(" [参数...]")
	}
	if len(a.commandList()) > 0 {
		b.WriteString(" [命令] [参数...]")
	}
	return b.String()
}
