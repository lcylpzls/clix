package clix

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestFlagValidateSuccess(t *testing.T) {
	flags := []FlagSpec{
		StringFlag("name", "").Validate("required,min=2"),
		IntFlag("retry", "").Validate("gte=0,lte=10").Default(3),
	}
	_, fv, err := parseCommandArgs(nil, flags, []string{"--name", "小明"})
	if err != nil {
		t.Fatalf("期望校验通过，得到 %v", err)
	}
	ctx := &Context{Flags: fv}
	if ctx.String("name") != "小明" || ctx.Int("retry") != 3 {
		t.Fatalf("flag 值不匹配：%q/%d", ctx.String("name"), ctx.Int("retry"))
	}
}

func TestFlagValidateFailure(t *testing.T) {
	_, _, err := parseCommandArgs(nil, []FlagSpec{
		StringFlag("name", "").Validate("required,min=2"),
	}, []string{"--name", "a"})
	assertErrCode(t, err, CodeFlagValidationFailed)
}

func TestFlagValidateSlice(t *testing.T) {
	flags := []FlagSpec{
		StringSliceFlag("tag", "").Validate("oneof=a b c"),
	}
	if _, _, err := parseCommandArgs(nil, flags, []string{"--tag", "a", "--tag", "b"}); err != nil {
		t.Fatalf("期望合法值通过，得到 %v", err)
	}
	_, _, err := parseCommandArgs(nil, flags, []string{"--tag", "a", "--tag", "x"})
	assertErrCode(t, err, CodeFlagValidationFailed)
}

func TestFlagValidateExecuteExitUsage(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	_ = app.AddCommand(&Command{
		Name:  "hello",
		Flags: []FlagSpec{StringFlag("name", "").Validate("required,min=2")},
		Action: func(ctx context.Context, c *Context) error {
			return nil
		},
	})
	if code := app.Execute(context.Background(), []string{"hello", "--name", "a"}); code != ExitUsage {
		t.Fatalf("期望退出码 %d，得到 %d", ExitUsage, code)
	}
	if !strings.Contains(errBuf.String(), "CLI_FLAG_VALIDATION_FAILED") {
		t.Fatalf("错误码缺失：%s", errBuf.String())
	}
}

func TestValidateRulesSyntaxInvalid(t *testing.T) {
	app := newTestApp(t)
	err := app.AddCommand(&Command{
		Name:  "bad",
		Flags: []FlagSpec{StringFlag("name", "").Validate("不存在的规则")},
		Action: func(ctx context.Context, c *Context) error {
			return nil
		},
	})
	assertErrCode(t, err, CodeInvalidFlagDef)
}

func TestValidateRulesZeroValueAllowed(t *testing.T) {
	// 规则合法但零值不通过（required）：注册期应通过，解析期按实际值校验。
	flags := []FlagSpec{StringFlag("name", "").Validate("required")}
	if err := validateFlagSpecs(flags); err != nil {
		t.Fatalf("注册期不应报错：%v", err)
	}
}

func TestZeroValueForKindAll(t *testing.T) {
	flags := []FlagSpec{
		BoolFlag("b", "").Validate("required"),
		IntFlag("i", "").Validate("required"),
		Int64Flag("i64", "").Validate("required"),
		FloatFlag("f", "").Validate("required"),
		DurationFlag("d", "").Validate("required"),
		StringSliceFlag("sl", "").Validate("required"),
		StringFlag("s", "").Validate("max=5"),
	}
	if err := validateFlagSpecs(flags); err != nil {
		t.Fatalf("全部类型的零值预编译应通过：%v", err)
	}
}

func TestFlagValidateAllKinds(t *testing.T) {
	flags := []FlagSpec{
		BoolFlag("b", "").Validate("eq=true"),
		IntFlag("i", "").Validate("gte=3"),
		Int64Flag("i64", "").Validate("gte=3"),
		FloatFlag("f", "").Validate("gte=1.5"),
		DurationFlag("d", "").Validate("gte=0"),
	}
	_, fv, err := parseCommandArgs(nil, flags, []string{
		"--b=true", "--i", "3", "--i64", "4", "--f", "2.0", "--d", "3s",
	})
	if err != nil {
		t.Fatalf("期望全部类型校验通过，得到 %v", err)
	}
	ctx := &Context{Flags: fv}
	if !ctx.Bool("b") || ctx.Int("i") != 3 || ctx.Int64("i64") != 4 ||
		ctx.Float64("f") != 2.0 || ctx.Duration("d").String() != "3s" {
		t.Fatalf("校验后的值不匹配：%+v", ctx)
	}
}

func TestFlagValidateHelp(t *testing.T) {
	app := newTestApp(t)
	_ = app.AddCommand(&Command{
		Name:  "hello",
		Flags: []FlagSpec{StringFlag("name", "名称").Validate("required")},
		Action: func(ctx context.Context, c *Context) error {
			return nil
		},
	})
	text, err := app.CommandHelpText("hello")
	if err != nil {
		t.Fatalf("期望成功，得到 %v", err)
	}
	if !strings.Contains(text, "校验 required") {
		t.Fatalf("帮助缺少校验规则：\n%s", text)
	}
}
