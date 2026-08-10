package clix

import (
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/logx"
)

// ActionFunc 是命令或根 Action 的执行函数。
// 返回 nil 表示成功；返回错误时由 Execute 统一输出并按退出码约定映射。
type ActionFunc func(ctx context.Context, c *Context) error

// Command 描述一个命令，可嵌套子命令。
//
// Name 必须非空、不得以 "-" 开头；Action 与子命令至少提供一个。
// Aliases 是可选的别名（须合法且不与同级命令名/别名冲突）；
// Group 用于帮助列表分组；Args/Flags 声明参数（nil 表示原样透传）。
//
// 命令注册后请勿修改其导出字段；子命令通过 AddCommand 添加。
type Command struct {
	Name        string
	Description string
	Usage       string
	// Args 声明位置参数；为 nil 时原样透传原始参数。
	Args []ArgSpec
	// Flags 声明 flag；为 nil 时不做 flag 解析。
	Flags []FlagSpec
	// Aliases 命令别名，帮助与分发时均可使用。
	Aliases []string
	// Group 帮助列表分组名；为空时归入默认（无标题）分组。
	Group string
	// Action 执行函数；与子命令至少提供一个。
	Action ActionFunc
	// Before 执行函数之前运行；返回错误时中止 Action 与 After。
	Before ActionFunc
	// After 执行函数之后运行（Action 失败也会运行）；
	// 与 Action 同时失败时合并为 errx 聚合错误。
	After ActionFunc

	parent     *Command
	children   registry
	registered atomic.Bool
}

// Context 是 Action 的执行上下文，携带 App、当前命令与参数。
type Context struct {
	// App 是所属应用实例。
	App *App
	// Command 是当前执行的命令；根 Action 时为 nil。
	Command *Command
	// Args 是解析后的位置参数（未声明时原样透传）。
	Args []string
	// Flags 是解析后的 flag 值；未声明 flag 或未解析时为空。
	Flags FlagValues
	// Global 是解析后的全局 flag 值。
	Global FlagValues
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

// AddCommand 在当前命令下注册子命令。
// 子命令名/别名不得与同级命令冲突；失败返回 errx 错误。
func (c *Command) AddCommand(sub *Command) error {
	if c == nil {
		return errx.NewCode(CodeInvalidApp, "父命令不能为空")
	}
	if err := c.children.add(sub); err != nil {
		return err
	}
	sub.parent = c
	return nil
}

// Parent 返回父命令；顶层命令与根 Action 的 Command 返回 nil。
func (c *Command) Parent() *Command {
	if c == nil {
		return nil
	}
	return c.parent
}

// FullName 返回从顶层到当前命令的完整路径，如 "parent child"。
func (c *Command) FullName() string {
	if c == nil {
		return ""
	}
	var parts []string
	for cur := c; cur != nil; cur = cur.parent {
		parts = append(parts, cur.Name)
	}
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, " ")
}

// registry 是命令集合：按名称与别名索引，保持注册顺序。
// App 根与每个命令各自持有一个 registry。
type registry struct {
	mu      sync.RWMutex
	byName  map[string]*Command
	aliases map[string]*Command
	order   []*Command
}

// add 校验并注册命令；命令名/别名规范化后写入 cmd。
func (r *registry) add(cmd *Command) error {
	if cmd == nil {
		return errx.NewCode(CodeInvalidApp, "命令不能为空")
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return errx.NewCode(CodeInvalidApp, "命令名不能为空")
	}
	if !validFlagName(name) {
		return errx.NewCodef(CodeInvalidApp, "非法命令名 %q：需以字母开头，仅含字母、数字、下划线与短横线", name)
	}
	if err := validateFlagSpecs(cmd.Flags); err != nil {
		return err
	}
	if err := validateArgSpecs(cmd.Args); err != nil {
		return err
	}
	if cmd.Action == nil && cmd.children.count() == 0 {
		return errx.NewCodef(CodeInvalidApp, "命令 %q 未定义执行函数且无子命令", name)
	}
	aliases, err := normalizeAliases(cmd, name)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if cmd.registered.Load() {
		return errx.NewCodef(CodeInvalidApp, "命令 %q 已注册", name)
	}
	if r.byName == nil {
		r.byName = make(map[string]*Command)
		r.aliases = make(map[string]*Command)
	}
	if _, dup := r.byName[name]; dup {
		return errx.NewCodef(CodeInvalidApp, "命令 %q 已存在", name)
	}
	if _, dup := r.aliases[name]; dup {
		return errx.NewCodef(CodeInvalidApp, "命令名 %q 与同级命令别名冲突", name)
	}
	for _, alias := range aliases {
		if _, dup := r.byName[alias]; dup {
			return errx.NewCodef(CodeInvalidApp, "别名 %q 与同级命令名冲突", alias)
		}
		if _, dup := r.aliases[alias]; dup {
			return errx.NewCodef(CodeInvalidApp, "别名 %q 已存在", alias)
		}
	}
	cmd.Name = name
	cmd.Aliases = aliases
	cmd.registered.Store(true)
	r.byName[name] = cmd
	for _, alias := range aliases {
		r.aliases[alias] = cmd
	}
	r.order = append(r.order, cmd)
	return nil
}

// normalizeAliases 校验并规范化命令别名，返回去重后的列表。
func normalizeAliases(cmd *Command, name string) ([]string, error) {
	seen := make(map[string]struct{}, len(cmd.Aliases))
	var out []string
	for _, alias := range cmd.Aliases {
		a := strings.TrimSpace(alias)
		if a == "" {
			return nil, errx.NewCode(CodeInvalidApp, "命令别名不能为空")
		}
		if !validFlagName(a) {
			return nil, errx.NewCodef(CodeInvalidApp, "非法命令别名 %q", a)
		}
		if a == name {
			return nil, errx.NewCodef(CodeInvalidApp, "命令别名 %q 不能与命令名相同", a)
		}
		if _, dup := seen[a]; dup {
			return nil, errx.NewCodef(CodeInvalidApp, "命令别名 %q 重复", a)
		}
		seen[a] = struct{}{}
		out = append(out, a)
	}
	return out, nil
}

// lookup 按命令名或别名查找；未找到返回 nil。并发安全。
func (r *registry) lookup(name string) *Command {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cmd, ok := r.byName[name]; ok {
		return cmd
	}
	return r.aliases[name]
}

// list 返回按注册顺序排列的命令副本。并发安全。
func (r *registry) list() []*Command {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Command, len(r.order))
	copy(out, r.order)
	return out
}

// count 返回已注册命令数量。并发安全。
func (r *registry) count() int {
	if r == nil {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.order)
}
