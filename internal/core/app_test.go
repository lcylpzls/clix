package core

import (
	"bytes"
	"context"
	testx "github.com/lcylpzls/testx"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/logx"
)

func TestNewValidation(t *testing.T) {
	t.Run("空应用名", func(t *testing.T) {
		app, err := New("", "0.1.0")
		assertErrCode(t, err, CodeInvalidApp)
		testx.RequireNil(t, app)

	})
	t.Run("空白应用名", func(t *testing.T) {
		_, err := New("   ", "0.1.0")
		assertErrCode(t, err, CodeInvalidApp)
	})
	t.Run("空版本号", func(t *testing.T) {
		_, err := New("greet", "")
		assertErrCode(t, err, CodeInvalidApp)
	})
	t.Run("空输出流", func(t *testing.T) {
		_, err := New("greet", "0.1.0", WithIO(nil, &bytes.Buffer{}))
		assertErrCode(t, err, CodeInvalidApp)
	})
	t.Run("空错误流", func(t *testing.T) {
		_, err := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, nil))
		assertErrCode(t, err, CodeInvalidApp)
	})
	t.Run("nil 选项安全跳过", func(t *testing.T) {
		app, err := New("greet", "0.1.0", nil)
		testx.RequireNoError(t, err)

		if app.Name() != "greet" {
			t.Fatalf("期望名称 greet，得到 %s", app.Name())
		}
	})
}

func TestNewDefaultsAndAccessors(t *testing.T) {
	var out, errBuf bytes.Buffer
	logger := &fakeLogger{}
	app, err := New(" greet ", " 0.1.0 ",
		WithDescription("测试应用"),
		WithUsage("greet 自定义用法"),
		WithIO(&out, &errBuf),
		WithLogger(logger),
		WithRootAction(func(ctx context.Context, c *Context) error { return nil }),
	)
	testx.RequireNoError(t, err)

	if got := app.Name(); got != "greet" {
		t.Fatalf("名称应去除空白，得到 %q", got)
	}
	if got := app.Version(); got != "0.1.0" {
		t.Fatalf("版本应去除空白，得到 %q", got)
	}
	if got := app.Description(); got != "测试应用" {
		t.Fatalf("描述不匹配：%q", got)
	}
	if app.Out() != &out {
		t.Fatal("Out 访问器不匹配")
	}
	if app.Err() != &errBuf {
		t.Fatal("Err 访问器不匹配")
	}
	if app.Logger() != logger {
		t.Fatal("Logger 访问器不匹配")
	}
}

func TestNewDefaultIO(t *testing.T) {
	app, err := New("greet", "0.1.0")
	testx.RequireNoError(t, err)

	if app.Out() != os.Stdout {
		t.Fatal("默认 Out 应为 os.Stdout")
	}
	if app.Err() != os.Stderr {
		t.Fatal("默认 Err 应为 os.Stderr")
	}
	if app.Logger() != nil {
		t.Fatal("默认 Logger 应为 nil")
	}
}

func TestAddCommandValidation(t *testing.T) {
	app := newTestApp(t)
	tests := []struct {
		name string
		cmd  *Command
	}{
		{"空命令", nil},
		{"空命令名", &Command{Name: "", Action: okAction}},
		{"空白命令名", &Command{Name: "  ", Action: okAction}},
		{"短横线命令名", &Command{Name: "-x", Action: okAction}},
		{"空执行函数", &Command{Name: "hello"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.AddCommand(tt.cmd)
			assertErrCode(t, err, CodeInvalidApp)
		})
	}
}

func TestAddCommandDuplicate(t *testing.T) {
	app := newTestApp(t)
	testx.RequireNoError(t, app.AddCommand(&Command{Name: "hello", Action: okAction}))
	err := app.AddCommand(&Command{Name: "hello", Action: okAction})
	assertErrCode(t, err, CodeInvalidApp)
}

func TestAddCommandInvalidDeclarations(t *testing.T) {
	app := newTestApp(t)
	if err := app.AddCommand(&Command{
		Name:   "badflag",
		Flags:  []FlagSpec{{Name: ""}},
		Action: okAction,
	}); err == nil {
		t.Fatal("非法 flag 定义应报错")
	} else {
		assertErrCode(t, err, CodeInvalidFlagDef)
	}
	if err := app.AddCommand(&Command{
		Name:   "badarg",
		Args:   []ArgSpec{{Name: ""}},
		Action: okAction,
	}); err == nil {
		t.Fatal("非法参数定义应报错")
	} else {
		assertErrCode(t, err, CodeInvalidArgDef)
	}
}

func TestAddCommandNormalizesName(t *testing.T) {
	app := newTestApp(t)
	testx.RequireNoError(t, app.AddCommand(&Command{Name: "  greet  ", Description: "问候", Action: okAction}))
	commands := app.commandList()
	if len(commands) != 1 || commands[0].Name != "greet" {
		t.Fatalf("命令名应去除空白，得到 %v", commands)
	}
}

func TestHelpTextVariants(t *testing.T) {
	t.Run("无命令无根 Action", func(t *testing.T) {
		app, _ := New("greet", "0.1.0")
		text := app.HelpText()
		testx.RequireTrue(t, strings.Contains(text, "用法:\n  greet\n"))
		testx.RequireTrue(t, strings.Contains(text, "（无）"))
	})
	t.Run("有命令", func(t *testing.T) {
		app := newTestApp(t)
		testx.RequireNoError(t, app.AddCommand(&Command{Name: "hello", Description: "问候", Action: okAction}))
		testx.RequireNoError(t, app.AddCommand(&Command{Name: "sum", Description: "求和", Action: okAction}))
		text := app.HelpText()
		testx.RequireTrue(t, strings.Contains(text, "用法:\n  greet [命令] [参数...]\n"))
		testx.RequireTrue(t, strings.Contains(text, "hello  问候"))
		testx.RequireTrue(t, strings.Contains(text, "sum    求和"))
	})
	t.Run("根 Action", func(t *testing.T) {
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}),
			WithRootAction(func(ctx context.Context, c *Context) error { return nil }))
		_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
		if !strings.Contains(app.HelpText(), "greet [参数...] [命令] [参数...]") {
			t.Fatalf("根 Action 用法行缺失：\n%s", app.HelpText())
		}
	})
	t.Run("自定义用法与描述", func(t *testing.T) {
		app, _ := New("greet", "0.1.0",
			WithDescription("测试描述"),
			WithUsage("greet 专用用法"),
		)
		text := app.HelpText()
		testx.RequireTrue(t, strings.Contains(text, "greet - 测试描述"))
		testx.RequireTrue(t, strings.Contains(text, "greet 专用用法"))
	})
	t.Run("无描述", func(t *testing.T) {
		app, _ := New("greet", "0.1.0")
		if strings.Contains(app.HelpText(), " - ") {
			t.Fatalf("无描述时不应出现分隔符：\n%s", app.HelpText())
		}
	})
}

func TestCommandHelpText(t *testing.T) {
	app := newTestApp(t)
	app.AddCommand(&Command{
		Name:        "hello",
		Description: "问候",
		Action:      okAction,
	})
	text, err := app.CommandHelpText("hello")
	testx.RequireNoError(t, err)

	testx.RequireTrue(t, strings.Contains(text, "用法:\n  greet hello [选项...] [参数...]\n"))
	testx.RequireTrue(t, strings.Contains(text, "描述:\n  问候"))

	app.AddCommand(&Command{Name: "quiet", Usage: "quiet 特殊用法", Action: okAction})
	text, err = app.CommandHelpText("quiet")
	testx.RequireNoError(t, err)

	testx.RequireTrue(t, strings.Contains(text, "quiet 特殊用法"))
	if strings.Contains(text, "描述:") {
		t.Fatalf("无描述时不应输出描述段：\n%s", text)
	}
}

func TestCommandHelpTextUnknown(t *testing.T) {
	app := newTestApp(t)
	_, err := app.CommandHelpText("nope")
	assertErrCode(t, err, CodeUnknownCommand)
}

func TestContextAccessors(t *testing.T) {
	var out, errBuf bytes.Buffer
	logger := &fakeLogger{}
	app, err := New("greet", "0.1.0", WithIO(&out, &errBuf), WithLogger(logger))
	testx.RequireNoError(t, err)

	var gotOut, gotErr any
	var gotLogger logx.Logger
	app.AddCommand(&Command{
		Name: "check",
		Action: func(ctx context.Context, c *Context) error {
			gotOut = c.Out()
			gotErr = c.Err()
			gotLogger = c.Logger()
			return nil
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"check"}), ExitOK)
	if gotOut != &out {
		t.Fatal("Context.Out 不匹配")
	}
	if gotErr != &errBuf {
		t.Fatal("Context.Err 不匹配")
	}
	testx.RequireEqual(t, gotLogger, logger)

}

// 测试辅助

func okAction(ctx context.Context, c *Context) error {
	return nil
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	app, err := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}))
	testx.RequireNoError(t, err)

	return app
}

func assertErrCode(t *testing.T, err error, want errx.Code) {
	t.Helper()
	testx.RequireError(t, err)

	if !errx.Is(err, want) {
		t.Fatalf("期望错误码 %q，得到 %v", want, err)
	}
}

// fakeLogger 是 logx.Logger 的最小实现，用于断言 clix 日志调用。
type fakeLogger struct {
	mu     sync.Mutex
	debugs []string
	infos  []string
	warns  []string
	errors []string
}

func (l *fakeLogger) IsDebugEnabled() bool { return true }

func (l *fakeLogger) Debug(msg string, _ logx.FieldGroup) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugs = append(l.debugs, msg)
}

func (l *fakeLogger) Info(msg string, _ logx.FieldGroup) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infos = append(l.infos, msg)
}

func (l *fakeLogger) Warn(msg string, _ logx.FieldGroup) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warns = append(l.warns, msg)
}

func (l *fakeLogger) Error(msg string, _ logx.FieldGroup) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, msg)
}

func (l *fakeLogger) Panic(msg string, _ logx.FieldGroup) {}
func (l *fakeLogger) Fatal(msg string, _ logx.FieldGroup) {}

func (l *fakeLogger) Debugf(format string, args ...any) {}
func (l *fakeLogger) Infof(format string, args ...any)  {}
func (l *fakeLogger) Warnf(format string, args ...any)  {}
func (l *fakeLogger) Errorf(format string, args ...any) {}
func (l *fakeLogger) Panicf(format string, args ...any) {}
func (l *fakeLogger) Fatalf(format string, args ...any) {}

func (l *fakeLogger) WithContext(ctx context.Context) logx.Logger { return l }
func (l *fakeLogger) WithField(key string, val any) logx.Logger   { return l }
func (l *fakeLogger) Sync() error                                 { return nil }
func (l *fakeLogger) Close() error                                { return nil }
func (l *fakeLogger) SafeExit(exitFunc func()) {
	if exitFunc != nil {
		exitFunc()
	}
}

func (l *fakeLogger) counts() (debugs, infos, warns, errors int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.debugs), len(l.infos), len(l.warns), len(l.errors)
}
