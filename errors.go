package clix

import "github.com/lcylpzls/errx"

// 退出码约定：0 成功、1 执行失败、2 用法错误、130 上下文取消。
const (
	// ExitOK 成功退出码。
	ExitOK int = 0
	// ExitFailure 命令执行失败退出码。
	ExitFailure int = 1
	// ExitUsage 用法错误退出码（缺少/未知命令、非法参数）。
	ExitUsage int = 2
	// ExitCancelled 上下文取消退出码（与 shell 130 约定一致）。
	ExitCancelled int = 130
)

// clix 错误码全集：所有失败场景统一使用 errx 结构化错误。
const (
	// CodeInvalidApp 应用或命令配置非法。
	CodeInvalidApp errx.Code = "CLI_INVALID_APP"
	// CodeMissingCommand 缺少命令。
	CodeMissingCommand errx.Code = "CLI_MISSING_COMMAND"
	// CodeUnknownCommand 未知命令。
	CodeUnknownCommand errx.Code = "CLI_UNKNOWN_COMMAND"
	// CodeCancelled 执行被上下文取消。
	CodeCancelled errx.Code = "CLI_CANCELLED"
	// CodeActionPanic 命令执行发生未捕获异常。
	CodeActionPanic errx.Code = "CLI_ACTION_PANIC"
)

func init() {
	errx.RegisterCode(CodeInvalidApp, "应用或命令配置非法")
	errx.RegisterCodeKind(CodeInvalidApp, errx.KindInvalid)
	errx.RegisterCode(CodeMissingCommand, "缺少命令")
	errx.RegisterCodeKind(CodeMissingCommand, errx.KindInvalid)
	errx.RegisterCode(CodeUnknownCommand, "未知命令")
	errx.RegisterCodeKind(CodeUnknownCommand, errx.KindInvalid)
	errx.RegisterCode(CodeCancelled, "执行被上下文取消")
	errx.RegisterCodeKind(CodeCancelled, errx.KindCancelled)
	errx.RegisterCode(CodeActionPanic, "命令执行发生未捕获异常")
	errx.RegisterCodeKind(CodeActionPanic, errx.KindInternal)
}
