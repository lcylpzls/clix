package clix_test

import (
	"context"
	"io"
	"testing"

	"github.com/lcylpzls/clix"
	"github.com/lcylpzls/confx"
)

// TestPublicAPI 黑盒冒烟测试：覆盖根包全部转发函数、类型别名与常量。
func TestPublicAPI(t *testing.T) {
	app, err := clix.New("smoke", "v1.0.0",
		clix.WithDescription("冒烟"),
		clix.WithUsage("smoke [command]"),
		clix.WithIO(io.Discard, io.Discard),
		clix.WithLogger(nil),
		clix.WithRootAction(func(context.Context, *clix.Context) error { return nil }),
		clix.WithGlobalFlags(clix.StringFlag("verbose", "详细输出")),
		clix.WithErrorHint(clix.CodeInvalidApp, "配置错误"),
		clix.WithObserver(nil),
	)
	if err != nil || app == nil {
		t.Fatalf("New 失败：%v", err)
	}

	cmd := &clix.Command{
		Name: "hello",
		Flags: []clix.FlagSpec{
			clix.BoolFlag("upper", "大写"),
			clix.IntFlag("n", "次数"),
			clix.Int64Flag("i64", "64 位"),
			clix.FloatFlag("f", "浮点"),
			clix.DurationFlag("d", "时长"),
			clix.EnumFlag("mode", "模式", "a", "b"),
			clix.StringSliceFlag("tags", "标签"),
		},
		Action: func(context.Context, *clix.Context) error { return nil },
	}
	if err := app.AddCommand(cmd); err != nil {
		t.Fatalf("AddCommand 失败：%v", err)
	}

	_ = app.Name()
	_ = app.Version()
	_ = app.Description()
	_ = app.Out()
	_ = app.Err()
	_ = app.Logger()

	mgr, err := confx.NewConfigManager(confx.Toml)
	if err != nil {
		t.Fatalf("confx.NewConfigManager 失败：%v", err)
	}
	ctx := &clix.Context{Args: []string{"--config", "/nonexistent.toml"}}
	_ = clix.LoadConfig(context.Background(), ctx, mgr, "config", "", &struct{}{})

	_ = clix.ExitOK
	_ = clix.ExitFailure
	_ = clix.ExitUsage
	_ = clix.ExitCancelled
	_ = clix.KindString
	_ = clix.KindBool
	_ = clix.KindInt
	_ = clix.KindInt64
	_ = clix.KindFloat64
	_ = clix.KindDuration
	_ = clix.KindEnum
	_ = clix.KindStringSlice
	_ = clix.CodeInvalidApp
	_ = clix.CodeMissingCommand
	_ = clix.CodeUnknownCommand
	_ = clix.CodeCancelled
	_ = clix.CodeActionPanic
	_ = clix.CodeInvalidFlagDef
	_ = clix.CodeInvalidArgDef
	_ = clix.CodeUnknownFlag
	_ = clix.CodeDuplicateFlag
	_ = clix.CodeMissingFlagValue
	_ = clix.CodeInvalidFlagValue
	_ = clix.CodeMissingRequiredFlag
	_ = clix.CodeInvalidEnumValue
	_ = clix.CodeMissingArg
	_ = clix.CodeTooManyArgs
	_ = clix.CodeFlagValidationFailed

	var _ clix.ValueKind
	var _ clix.ArgSpec
	var _ clix.FlagSpec
	var _ clix.FlagValues
	var _ clix.Observer
	var _ clix.Option
	var _ clix.ActionFunc
	var _ clix.App
	var _ clix.Command
	var _ clix.Context
}
