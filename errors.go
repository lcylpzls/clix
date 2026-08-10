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
	// CodeInvalidFlagDef flag 定义非法（配置错误）。
	CodeInvalidFlagDef errx.Code = "CLI_INVALID_FLAG_DEF"
	// CodeInvalidArgDef 位置参数定义非法（配置错误）。
	CodeInvalidArgDef errx.Code = "CLI_INVALID_ARG_DEF"
	// CodeUnknownFlag 未知 flag。
	CodeUnknownFlag errx.Code = "CLI_UNKNOWN_FLAG"
	// CodeDuplicateFlag 非重复 flag 被重复指定。
	CodeDuplicateFlag errx.Code = "CLI_DUPLICATE_FLAG"
	// CodeMissingFlagValue flag 缺少值。
	CodeMissingFlagValue errx.Code = "CLI_MISSING_FLAG_VALUE"
	// CodeInvalidFlagValue flag 值类型非法。
	CodeInvalidFlagValue errx.Code = "CLI_INVALID_FLAG_VALUE"
	// CodeMissingRequiredFlag 缺少必填 flag。
	CodeMissingRequiredFlag errx.Code = "CLI_MISSING_REQUIRED_FLAG"
	// CodeInvalidEnumValue 枚举 flag 值不在允许列表。
	CodeInvalidEnumValue errx.Code = "CLI_INVALID_ENUM_VALUE"
	// CodeMissingArg 缺少必填位置参数。
	CodeMissingArg errx.Code = "CLI_MISSING_ARG"
	// CodeTooManyArgs 位置参数过多。
	CodeTooManyArgs errx.Code = "CLI_TOO_MANY_ARGS"
	// CodeFlagValidationFailed flag 值未通过 validx 规则校验。
	CodeFlagValidationFailed errx.Code = "CLI_FLAG_VALIDATION_FAILED"
)

// usageErrorCodes 是全部用法错误码；命中任一码时 Execute 返回 ExitUsage。
var usageErrorCodes = []errx.Code{
	CodeInvalidApp,
	CodeMissingCommand,
	CodeUnknownCommand,
	CodeInvalidFlagDef,
	CodeInvalidArgDef,
	CodeUnknownFlag,
	CodeDuplicateFlag,
	CodeMissingFlagValue,
	CodeInvalidFlagValue,
	CodeMissingRequiredFlag,
	CodeInvalidEnumValue,
	CodeMissingArg,
	CodeTooManyArgs,
	CodeFlagValidationFailed,
}

// isUsageError 判断错误是否为用法错误（沿错误链匹配错误码）。
func isUsageError(err error) bool {
	for _, code := range usageErrorCodes {
		if errx.Is(err, code) {
			return true
		}
	}
	return false
}

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
	errx.RegisterCode(CodeInvalidFlagDef, "flag 定义非法")
	errx.RegisterCodeKind(CodeInvalidFlagDef, errx.KindInvalid)
	errx.RegisterCode(CodeInvalidArgDef, "位置参数定义非法")
	errx.RegisterCodeKind(CodeInvalidArgDef, errx.KindInvalid)
	errx.RegisterCode(CodeUnknownFlag, "未知 flag")
	errx.RegisterCodeKind(CodeUnknownFlag, errx.KindInvalid)
	errx.RegisterCode(CodeDuplicateFlag, "非重复 flag 被重复指定")
	errx.RegisterCodeKind(CodeDuplicateFlag, errx.KindInvalid)
	errx.RegisterCode(CodeMissingFlagValue, "flag 缺少值")
	errx.RegisterCodeKind(CodeMissingFlagValue, errx.KindInvalid)
	errx.RegisterCode(CodeInvalidFlagValue, "flag 值类型非法")
	errx.RegisterCodeKind(CodeInvalidFlagValue, errx.KindInvalid)
	errx.RegisterCode(CodeMissingRequiredFlag, "缺少必填 flag")
	errx.RegisterCodeKind(CodeMissingRequiredFlag, errx.KindInvalid)
	errx.RegisterCode(CodeInvalidEnumValue, "枚举 flag 值不在允许列表")
	errx.RegisterCodeKind(CodeInvalidEnumValue, errx.KindInvalid)
	errx.RegisterCode(CodeMissingArg, "缺少必填位置参数")
	errx.RegisterCodeKind(CodeMissingArg, errx.KindInvalid)
	errx.RegisterCode(CodeTooManyArgs, "位置参数过多")
	errx.RegisterCodeKind(CodeTooManyArgs, errx.KindInvalid)
	errx.RegisterCode(CodeFlagValidationFailed, "flag 值未通过规则校验")
	errx.RegisterCodeKind(CodeFlagValidationFailed, errx.KindInvalid)
}
