package clix

import (
	"bytes"
	"context"
	testx "github.com/lcylpzls/testx"
	"strings"
	"testing"
	"time"
)

func TestGlobalFlagsBeforeCommand(t *testing.T) {
	var out bytes.Buffer
	var got struct {
		level    string
		verbose  bool
		retry    int
		size     int64
		ratio    float64
		timeout  time.Duration
		mode     string
		tags     []string
		args     []string
		hasLevel bool
	}
	app, err := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}),
		WithGlobalFlags(
			StringFlag("log-level", "日志级别").Default("info").Env("CLIX_G_LOG"),
			BoolFlag("verbose", "详细"),
			IntFlag("retry", "重试").Default(3),
			Int64Flag("size", "大小"),
			FloatFlag("ratio", "比率"),
			DurationFlag("timeout", "超时"),
			EnumFlag("mode", "模式", "fast", "slow").Default("fast"),
			StringSliceFlag("tag", "标签"),
		),
	)
	testx.RequireNoError(t, err)

	_ = app.AddCommand(&Command{
		Name: "hello",
		Action: func(ctx context.Context, c *Context) error {
			got.level = c.GlobalString("log-level")
			got.verbose = c.GlobalBool("verbose")
			got.retry = c.GlobalInt("retry")
			got.size = c.GlobalInt64("size")
			got.ratio = c.GlobalFloat64("ratio")
			got.timeout = c.GlobalDuration("timeout")
			got.mode = c.GlobalEnum("mode")
			got.tags = c.GlobalStrings("tag")
			got.args = c.Args
			got.hasLevel = c.HasGlobalFlag("log-level")
			return nil
		},
	})
	code := app.Execute(context.Background(), []string{
		"--log-level=debug", "--verbose", "--retry", "7", "--size=8",
		"--ratio", "0.5", "--timeout", "2s", "--mode", "slow",
		"--tag", "a", "--tag=b",
		"hello", "x", "y",
	})
	testx.RequireEqual(t, code, ExitOK)

	if got.level != "debug" || !got.verbose || got.retry != 7 || got.size != 8 ||
		got.ratio != 0.5 || got.timeout != 2*time.Second || got.mode != "slow" ||
		!got.hasLevel {
		t.Fatalf("全局 flag 值不匹配：%+v", got)
	}
	if strings.Join(got.tags, ",") != "a,b" {
		t.Fatalf("全局可重复 flag 不匹配：%v", got.tags)
	}
	if strings.Join(got.args, ",") != "x,y" {
		t.Fatalf("命令参数不匹配：%v", got.args)
	}
}

func TestGlobalFlagsRootAction(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}),
		WithGlobalFlags(BoolFlag("verbose", "详细")),
		WithRootAction(func(ctx context.Context, c *Context) error {
			if c.GlobalBool("verbose") {
				out.WriteString("详细模式\n")
			}
			return nil
		}),
	)
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"--verbose"}), ExitOK)
	if !strings.Contains(out.String(), "详细模式") {
		t.Fatalf("根 Action 未读取全局 flag：%s", out.String())
	}
}

func TestGlobalBoolEquals(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}),
		WithGlobalFlags(BoolFlag("verbose", "详细")))
	_ = app.AddCommand(&Command{
		Name: "hello",
		Action: func(ctx context.Context, c *Context) error {
			if c.GlobalBool("verbose") {
				out.WriteString("详细\n")
			}
			return nil
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"--verbose=false", "hello"}), ExitOK)
	if out.Len() != 0 {
		t.Fatalf("--verbose=false 应关闭详细模式：%s", out.String())
	}
}

func TestGlobalFlagErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"缺少值", []string{"--level"}, "缺少值"},
		{"重复", []string{"--level", "a", "--level", "b", "hello"}, "重复"},
		{"布尔非法", []string{"--verbose=maybe", "hello"}, "需要布尔值"},
		{"未知全局 flag", []string{"--nope", "hello"}, "未知全局 flag"},
		{"整数非法", []string{"--retry", "x", "hello"}, "需要整数"},
		{"切片缺少值", []string{"--tag"}, "缺少值"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errBuf bytes.Buffer
			app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
				WithGlobalFlags(
					StringFlag("level", ""),
					BoolFlag("verbose", ""),
					IntFlag("retry", ""),
					StringSliceFlag("tag", ""),
				))
			_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
			testx.RequireEqual(t, app.Execute(context.Background(), tt.args), ExitUsage)
			if !strings.Contains(errBuf.String(), tt.want) {
				t.Fatalf("错误信息缺少 %q：%s", tt.want, errBuf.String())
			}
		})
	}
}

func TestGlobalFlagEnvInvalid(t *testing.T) {
	t.Setenv("CLIX_G_BAD", "x")
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
		WithGlobalFlags(IntFlag("n", "").Env("CLIX_G_BAD")))
	_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"hello"}), ExitUsage)
	if !strings.Contains(errBuf.String(), "需要整数") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestGlobalFlagValidation(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
		WithGlobalFlags(StringFlag("mode", "模式").Validate("oneof=dev prod")))
	_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"--mode", "staging", "hello"}), ExitUsage)
	if !strings.Contains(errBuf.String(), "CLI_FLAG_VALIDATION_FAILED") {
		t.Fatalf("校验错误缺失：%s", errBuf.String())
	}
}

func TestGlobalFlagValidationPass(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}),
		WithGlobalFlags(StringFlag("mode", "模式").Validate("oneof=dev prod")))
	_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"--mode", "dev", "hello"}), ExitOK)
}

func TestRequiredGlobalFlag(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
		WithGlobalFlags(StringFlag("token", "").Required()))
	_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"hello"}), ExitUsage)
	if !strings.Contains(errBuf.String(), "缺少必填全局 flag") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestGlobalFlagAfterCommandIsLocal(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
		WithGlobalFlags(BoolFlag("verbose", "详细")))
	_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"hello", "--verbose"}), ExitUsage)
	if !strings.Contains(errBuf.String(), "未知 flag") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestGlobalFlagTerminator(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
		WithGlobalFlags(BoolFlag("verbose", "详细")))
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"--", "--verbose"}), ExitUsage)
}

func TestGlobalFlagHelp(t *testing.T) {
	app, _ := New("greet", "0.1.0",
		WithGlobalFlags(
			StringFlag("log-level", "日志级别").Default("info").Env("CLIX_G_LOG"),
			BoolFlag("verbose", "详细"),
		))
	text := app.HelpText()
	for _, want := range []string{
		"全局选项:",
		"--log-level string",
		"环境变量 CLIX_G_LOG",
		"--verbose bool",
	} {
		testx.RequireTrue(t, strings.Contains(text, want))
	}
}

func TestWithGlobalFlagsInvalid(t *testing.T) {
	_, err := New("greet", "0.1.0", WithGlobalFlags(StringFlag("a", "").Env("X-Y")))
	assertErrCode(t, err, CodeInvalidFlagDef)
}

func TestGlobalAccessorsZero(t *testing.T) {
	ctx := &Context{}
	if ctx.HasGlobalFlag("x") || ctx.GlobalString("x") != "" || ctx.GlobalBool("x") ||
		ctx.GlobalInt("x") != 0 || ctx.GlobalInt64("x") != 0 || ctx.GlobalFloat64("x") != 0 ||
		ctx.GlobalDuration("x") != 0 || ctx.GlobalEnum("x") != "" || ctx.GlobalStrings("x") != nil {
		t.Fatal("未声明全局 flag 应返回零值")
	}
}

func TestGlobalStringsReturnsCopy(t *testing.T) {
	app, _ := New("greet", "0.1.0",
		WithGlobalFlags(StringSliceFlag("tag", "")))
	fv, _, err := app.stripGlobalFlags([]string{"--tag", "a"})
	testx.RequireNoError(t, err)

	ctx := &Context{Global: fv}
	got := ctx.GlobalStrings("tag")
	got[0] = "改坏"
	if ctx.GlobalStrings("tag")[0] != "a" {
		t.Fatal("修改返回切片不应影响内部存储")
	}
}
