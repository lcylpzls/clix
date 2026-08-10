package clix

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lcylpzls/errx"
	errxlogx "github.com/lcylpzls/errx/logx"
	"github.com/lcylpzls/logx"
)

// runResult 记录一次分发的结果，供日志输出使用。
type runResult struct {
	command *Command
	args    []string
}

// Run 执行参数解析与命令分发，返回原始错误（不打印、不映射退出码）。
// 帮助与版本视为成功，写 Out 并返回 nil。
func (a *App) Run(ctx context.Context, args []string) error {
	_, err := a.dispatch(ctx, args)
	return err
}

// Execute 执行命令并返回进程退出码；不调用 os.Exit。
//
// 错误输出到 Err；成功时帮助/版本输出到 Out。
// 日志在注入 Logger 后记录命令开始（Debug）、成功（Info）、
// 用法错误（Warn）、失败/取消/恐慌（Error）。
func (a *App) Execute(ctx context.Context, args []string) int {
	start := time.Now()
	result, err := a.dispatch(ctx, args)
	if err == nil {
		a.logInfo("命令执行成功", start, result, nil)
		return ExitOK
	}
	if isUsageError(err) {
		fmt.Fprintln(a.err, err.Error())
		if errx.Is(err, CodeMissingCommand) {
			fmt.Fprintln(a.err, "运行 --help 查看用法")
		}
		a.logWarn("命令用法错误", start, result, err)
		return ExitUsage
	}
	switch {
	case errx.Is(err, CodeCancelled):
		fmt.Fprintln(a.err, err.Error())
		a.logError("命令执行被取消", start, result, err)
		return ExitCancelled
	case errx.Is(err, CodeActionPanic):
		fmt.Fprintln(a.err, err.Error())
		a.logError("命令执行发生未捕获异常", start, result, err)
		return ExitFailure
	default:
		fmt.Fprintln(a.err, err.Error())
		a.logError("命令执行失败", start, result, err)
		return ExitFailure
	}
}

// dispatch 完成参数解析与命令分发。
func (a *App) dispatch(ctx context.Context, args []string) (runResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return runResult{}, errx.NewCode(CodeCancelled, "执行被上下文取消")
	}
	if len(args) == 0 {
		if a.root != nil {
			return a.invoke(ctx, nil, &Context{App: a, Args: nil})
		}
		return runResult{}, errx.NewCode(CodeMissingCommand, "缺少命令")
	}
	switch args[0] {
	case "-h", "--help":
		return runResult{}, a.printHelp(args[1:])
	case "help":
		return runResult{}, a.printHelp(args[1:])
	case "--version":
		fmt.Fprintf(a.out, "%s %s\n", a.name, a.version)
		return runResult{}, nil
	}
	cmd, rest := a.resolve(args)
	if cmd == nil {
		return runResult{}, errx.NewCodef(CodeUnknownCommand, "未知命令 %q，请运行 --help 查看可用命令", args[0])
	}
	if hasHelpRequest(rest) {
		fmt.Fprintln(a.out, a.renderCommandHelp(cmd))
		return runResult{command: cmd}, nil
	}
	if cmd.Action == nil {
		if len(rest) == 0 {
			return runResult{command: cmd}, errx.NewCodef(CodeMissingCommand,
				"命令 %q 需要子命令，请运行 %s --help 查看可用子命令", cmd.FullName(), cmd.FullName())
		}
		return runResult{command: cmd}, errx.NewCodef(CodeUnknownCommand,
			"未知子命令 %q（命令 %q 下），请运行 %s --help 查看可用子命令",
			rest[0], cmd.FullName(), cmd.FullName())
	}
	positional, flags, err := parseCommandArgs(cmd.Args, cmd.Flags, rest)
	if err != nil {
		return runResult{command: cmd}, err
	}
	return a.invoke(ctx, cmd, &Context{App: a, Command: cmd, Args: positional, Flags: flags})
}

// resolve 沿命令树解析参数：逐个匹配子命令名/别名，直到首个非命令参数。
// 返回最终命令与剩余参数；顶层未命中时返回 nil。
func (a *App) resolve(args []string) (*Command, []string) {
	cmd := a.lookup(args[0])
	if cmd == nil {
		return nil, args
	}
	rest := args[1:]
	for len(rest) > 0 {
		next := cmd.children.lookup(rest[0])
		if next == nil {
			break
		}
		cmd = next
		rest = rest[1:]
	}
	return cmd, rest
}

// resolvePath 按完整路径解析命令（供 help 子命令使用）；
// 任一段未知时返回 CodeUnknownCommand。
func (a *App) resolvePath(parts []string) (*Command, error) {
	cur := a.rootRegistry
	var cmd *Command
	for i, part := range parts {
		next := cur.lookup(part)
		if next == nil {
			return nil, errx.NewCodef(CodeUnknownCommand, "未知命令 %q，请运行 --help 查看可用命令", parts[i])
		}
		cmd = next
		cur = &next.children
	}
	return cmd, nil
}

// invoke 执行根 Action 或子命令 Action，并恢复未捕获异常。
func (a *App) invoke(ctx context.Context, cmd *Command, c *Context) (res runResult, err error) {
	action := a.root
	if cmd != nil {
		action = cmd.Action
	}
	res = runResult{command: cmd, args: c.Args}
	defer func() {
		if r := recover(); r != nil {
			res = runResult{command: cmd, args: c.Args}
			err = errx.NewCodef(CodeActionPanic, "命令执行发生未捕获异常：%v", r)
		}
	}()
	err = action(ctx, c)
	return res, err
}

// hasHelpRequest 判断参数中是否请求命令帮助（遇到 "--" 后不再识别）。
func hasHelpRequest(raw []string) bool {
	for _, tok := range raw {
		if tok == "--" {
			return false
		}
		if tok == "-h" || tok == "--help" {
			return true
		}
	}
	return false
}

// printHelp 输出应用或命令帮助；help 后跟未知命令时返回 errx 错误。
func (a *App) printHelp(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(a.out, a.HelpText())
		return nil
	}
	cmd, err := a.resolvePath(args)
	if err != nil {
		return err
	}
	fmt.Fprintln(a.out, a.renderCommandHelp(cmd))
	return nil
}

// logInfo 记录成功日志。
func (a *App) logInfo(msg string, start time.Time, res runResult, err error) {
	if a.logger == nil {
		return
	}
	a.logger.Info(msg, a.outcomeFields(start, res, err))
}

// logWarn 记录用法错误日志。
func (a *App) logWarn(msg string, start time.Time, res runResult, err error) {
	if a.logger == nil {
		return
	}
	a.logger.Warn(msg, a.outcomeFields(start, res, err))
}

// logError 记录失败/取消/恐慌日志。
func (a *App) logError(msg string, start time.Time, res runResult, err error) {
	if a.logger == nil {
		return
	}
	a.logger.Error(msg, a.outcomeFields(start, res, err))
}

// outcomeFields 汇总结果字段：命令、参数、耗时与错误结构化字段。
func (a *App) outcomeFields(start time.Time, res runResult, err error) logx.FieldGroup {
	name := "（根命令）"
	if res.command != nil {
		name = res.command.FullName()
	}
	groups := []logx.FieldGroup{
		logx.Fields(
			logx.String("command", name),
			logx.String("args", strings.Join(res.args, " ")),
			logx.Int64("duration_ms", time.Since(start).Milliseconds()),
		),
	}
	if err != nil {
		groups = append(groups, errxlogx.Fields(err))
	}
	var fs []logx.Field
	for _, g := range groups {
		for i := 0; i < g.Len(); i++ {
			fs = append(fs, g.At(i))
		}
	}
	return logx.Fields(fs...)
}
