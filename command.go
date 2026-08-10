package clix

import (
	"context"
	"io"

	"github.com/lcylpzls/logx"
)

// ActionFunc 是命令或根 Action 的执行函数。
// 返回 nil 表示成功；返回错误时由 Execute 统一输出并按退出码约定映射。
type ActionFunc func(ctx context.Context, c *Context) error

// Command 描述一个子命令。
//
// Name 必须非空、不得以 "-" 开头；Action 必须非空。
// Description 用于帮助列表；Usage 为空时自动生成
// "app 命令 [参数...]" 形式的用法行。
type Command struct {
	Name        string
	Description string
	Usage       string
	// Args 声明位置参数；为 nil 时原样透传原始参数。
	Args []ArgSpec
	// Flags 声明 flag；为 nil 时不做 flag 解析。
	Flags  []FlagSpec
	Action ActionFunc
}

// Context 是 Action 的执行上下文，携带 App、当前命令与原始参数。
type Context struct {
	// App 是所属应用实例。
	App *App
	// Command 是当前执行的命令；根 Action 时为 nil。
	Command *Command
	// Args 是命令名之后的原始参数（v0.2.0 起由参数解析层消费）。
	Args []string
	// Flags 是解析后的 flag 值；未声明 flag 或未解析时为空。
	Flags FlagValues
}

// Out 返回应用注入的标准输出流。
func (c *Context) Out() io.Writer {
	return c.App.out
}

// Err 返回应用注入的错误输出流。
func (c *Context) Err() io.Writer {
	return c.App.err
}

// Logger 返回应用注入的日志器；未注入时返回 nil。
func (c *Context) Logger() logx.Logger {
	return c.App.logger
}
