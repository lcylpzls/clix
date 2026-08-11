package clix

import (
	"context"
	"io"

	"github.com/lcylpzls/clix/internal/core"
	"github.com/lcylpzls/confx"
	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/logx"
)

type (
	ValueKind  = core.ValueKind
	ArgSpec    = core.ArgSpec
	FlagSpec   = core.FlagSpec
	FlagValues = core.FlagValues
	Observer   = core.Observer
	App        = core.App
	Option     = core.Option
	ActionFunc = core.ActionFunc
	Command    = core.Command
	Context    = core.Context
)

const (
	ExitOK        = core.ExitOK
	ExitFailure   = core.ExitFailure
	ExitUsage     = core.ExitUsage
	ExitCancelled = core.ExitCancelled
)

const (
	KindString      = core.KindString
	KindBool        = core.KindBool
	KindInt         = core.KindInt
	KindInt64       = core.KindInt64
	KindUint64      = core.KindUint64
	KindFloat64     = core.KindFloat64
	KindDuration    = core.KindDuration
	KindEnum        = core.KindEnum
	KindStringSlice = core.KindStringSlice
)

const (
	CodeInvalidApp           = core.CodeInvalidApp
	CodeMissingCommand       = core.CodeMissingCommand
	CodeUnknownCommand       = core.CodeUnknownCommand
	CodeCancelled            = core.CodeCancelled
	CodeActionPanic          = core.CodeActionPanic
	CodeInvalidFlagDef       = core.CodeInvalidFlagDef
	CodeInvalidArgDef        = core.CodeInvalidArgDef
	CodeUnknownFlag          = core.CodeUnknownFlag
	CodeDuplicateFlag        = core.CodeDuplicateFlag
	CodeMissingFlagValue     = core.CodeMissingFlagValue
	CodeInvalidFlagValue     = core.CodeInvalidFlagValue
	CodeMissingRequiredFlag  = core.CodeMissingRequiredFlag
	CodeInvalidEnumValue     = core.CodeInvalidEnumValue
	CodeMissingArg           = core.CodeMissingArg
	CodeTooManyArgs          = core.CodeTooManyArgs
	CodeFlagValidationFailed = core.CodeFlagValidationFailed
)

func New(name, version string, opts ...Option) (*App, error) { return core.New(name, version, opts...) }
func WithDescription(desc string) Option                     { return core.WithDescription(desc) }
func WithUsage(usage string) Option                          { return core.WithUsage(usage) }
func WithIO(out, err io.Writer) Option                       { return core.WithIO(out, err) }
func WithLogger(logger logx.Logger) Option                   { return core.WithLogger(logger) }
func WithRootAction(action ActionFunc) Option                { return core.WithRootAction(action) }
func WithGlobalFlags(flags ...FlagSpec) Option               { return core.WithGlobalFlags(flags...) }
func WithErrorHint(code errx.Code, hint string) Option       { return core.WithErrorHint(code, hint) }
func WithObserver(obs Observer) Option                       { return core.WithObserver(obs) }
func LoadConfig(ctx context.Context, c *Context, manager *confx.ConfigManager, pathFlag, fallback string, target any) error {
	return core.LoadConfig(ctx, c, manager, pathFlag, fallback, target)
}
func StringFlag(name, usage string) FlagSpec   { return core.StringFlag(name, usage) }
func BoolFlag(name, usage string) FlagSpec     { return core.BoolFlag(name, usage) }
func IntFlag(name, usage string) FlagSpec      { return core.IntFlag(name, usage) }
func Int64Flag(name, usage string) FlagSpec    { return core.Int64Flag(name, usage) }
func FloatFlag(name, usage string) FlagSpec    { return core.FloatFlag(name, usage) }
func DurationFlag(name, usage string) FlagSpec { return core.DurationFlag(name, usage) }
func EnumFlag(name, usage string, allowed ...string) FlagSpec {
	return core.EnumFlag(name, usage, allowed...)
}
func StringSliceFlag(name, usage string) FlagSpec { return core.StringSliceFlag(name, usage) }
