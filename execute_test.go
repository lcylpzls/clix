package clix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

func TestExecuteHelpFlags(t *testing.T) {
	app := newTestApp(t)
	app.AddCommand(&Command{Name: "hello", Description: "问候", Action: okAction})
	for _, flag := range []string{"-h", "--help", "help"} {
		t.Run(flag, func(t *testing.T) {
			var out, errBuf bytes.Buffer
			app2, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
			app2.AddCommand(&Command{Name: "hello", Description: "问候", Action: okAction})
			testx.RequireEqual(t, app2.Execute(context.Background(), []string{flag}), ExitOK)
			if !strings.Contains(out.String(), "命令:") {
				t.Fatalf("帮助缺失：\n%s", out.String())
			}
			if errBuf.Len() != 0 {
				t.Fatalf("错误流应为空：%s", errBuf.String())
			}
		})
	}
}

func TestExecuteHelpCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	app.AddCommand(&Command{Name: "hello", Description: "问候", Action: okAction})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"help", "hello"}), ExitOK)
	if !strings.Contains(out.String(), "greet hello [选项...] [参数...]") {
		t.Fatalf("命令帮助缺失：\n%s", out.String())
	}
}

func TestExecuteHelpUnknownCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"help", "nope"}), ExitUsage)
	if out.Len() != 0 {
		t.Fatalf("帮助失败时 Out 应为空：%s", out.String())
	}
	if !strings.Contains(errBuf.String(), "未知命令") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestExecuteVersion(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"--version"}), ExitOK)
	if got := out.String(); got != "greet 0.1.0\n" {
		t.Fatalf("版本行不匹配：%q", got)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("错误流应为空：%s", errBuf.String())
	}
}

func TestExecuteMissingCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	testx.RequireEqual(t, app.Execute(context.Background(), nil), ExitUsage)
	if !strings.Contains(errBuf.String(), "缺少命令") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "运行 --help 查看用法") {
		t.Fatalf("提示信息缺失：%s", errBuf.String())
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"nope"}), ExitUsage)
	if !strings.Contains(errBuf.String(), "未知命令 \"nope\"") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestExecuteRootAction(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf),
			WithRootAction(func(ctx context.Context, c *Context) error {
				fmt.Fprintln(c.Out(), "根命令输出")
				return nil
			}))
		testx.RequireEqual(t, app.Execute(context.Background(), nil), ExitOK)
		if !strings.Contains(out.String(), "根命令输出") {
			t.Fatalf("根命令输出缺失：%s", out.String())
		}
	})
	t.Run("普通错误", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
			WithRootAction(func(ctx context.Context, c *Context) error {
				return errors.New("根命令失败")
			}))
		testx.RequireEqual(t, app.Execute(context.Background(), nil), ExitFailure)
		if !strings.Contains(errBuf.String(), "根命令失败") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
	t.Run("errx 错误", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf),
			WithRootAction(func(ctx context.Context, c *Context) error {
				return errx.NewCode("CLI_TEST_FAIL", "根命令业务失败").
					WithField("order_id", "10086")
			}))
		testx.RequireEqual(t, app.Execute(context.Background(), nil), ExitFailure)
		if !strings.Contains(errBuf.String(), "CLI_TEST_FAIL: 根命令业务失败") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
}

func TestExecuteCommand(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
		app.AddCommand(&Command{
			Name: "hello",
			Action: func(ctx context.Context, c *Context) error {
				fmt.Fprintf(c.Out(), "你好，%s！\n", strings.Join(c.Args, " "))
				return nil
			},
		})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"hello", "小明", "小红"}), ExitOK)
		if !strings.Contains(out.String(), "你好，小明 小红！") {
			t.Fatalf("命令输出缺失：%s", out.String())
		}
	})
	t.Run("命令错误", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
		app.AddCommand(&Command{
			Name: "boom",
			Action: func(ctx context.Context, c *Context) error {
				return errx.NewCode("CLI_TEST_FAIL", "命令执行失败")
			},
		})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"boom"}), ExitFailure)
		if !strings.Contains(errBuf.String(), "命令执行失败") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
	t.Run("命令恐慌", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
		app.AddCommand(&Command{
			Name: "panic",
			Action: func(ctx context.Context, c *Context) error {
				panic("内部崩溃")
			},
		})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"panic"}), ExitFailure)
		if !strings.Contains(errBuf.String(), "CLI_ACTION_PANIC") {
			t.Fatalf("恐慌错误码缺失：%s", errBuf.String())
		}
		if !strings.Contains(errBuf.String(), "内部崩溃") {
			t.Fatalf("恐慌信息缺失：%s", errBuf.String())
		}
	})
}

func TestExecuteCancelled(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	testx.RequireEqual(t, app.Execute(ctx, []string{"hello"}), ExitCancelled)
	if !strings.Contains(errBuf.String(), "CLI_CANCELLED") {
		t.Fatalf("取消错误码缺失：%s", errBuf.String())
	}
}

func TestExecuteNilContext(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	var nilCtx context.Context
	testx.RequireEqual(t, app.Execute(nilCtx, []string{"--version"}), ExitOK)
	if out.String() != "greet 0.1.0\n" {
		t.Fatalf("版本行不匹配：%q", out.String())
	}
}

func TestRunReturnsErrorWithoutPrinting(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	err := app.Run(context.Background(), []string{"nope"})
	assertErrCode(t, err, CodeUnknownCommand)
	if out.Len() != 0 {
		t.Fatalf("Run 不应输出帮助：%s", out.String())
	}
	if errBuf.Len() != 0 {
		t.Fatalf("Run 不应打印错误：%s", errBuf.String())
	}
}

func TestRunHelpSuccess(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	testx.RequireNoError(t, app.Run(context.Background(), []string{"--help"}))
	if !strings.Contains(out.String(), "命令:") {
		t.Fatalf("帮助缺失：%s", out.String())
	}
}

func TestExecuteLogging(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		logger := &fakeLogger{}
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}), WithLogger(logger))
		app.AddCommand(&Command{Name: "hello", Action: okAction})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"hello", "a"}), ExitOK)
		_, infos, _, _ := logger.counts()
		if infos != 1 {
			t.Fatalf("期望 1 条 Info 日志，得到 %d", infos)
		}
	})
	t.Run("用法错误", func(t *testing.T) {
		logger := &fakeLogger{}
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}), WithLogger(logger))
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"nope"}), ExitUsage)
		_, _, warns, _ := logger.counts()
		if warns != 1 {
			t.Fatalf("期望 1 条 Warn 日志，得到 %d", warns)
		}
	})
	t.Run("执行失败", func(t *testing.T) {
		logger := &fakeLogger{}
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}), WithLogger(logger))
		app.AddCommand(&Command{
			Name: "boom",
			Action: func(ctx context.Context, c *Context) error {
				return errors.New("失败")
			},
		})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"boom"}), ExitFailure)
		_, _, _, errorsCount := logger.counts()
		if errorsCount != 1 {
			t.Fatalf("期望 1 条 Error 日志，得到 %d", errorsCount)
		}
	})
	t.Run("取消", func(t *testing.T) {
		logger := &fakeLogger{}
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}), WithLogger(logger))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		testx.RequireEqual(t, app.Execute(ctx, nil), ExitCancelled)
		_, _, _, errorsCount := logger.counts()
		if errorsCount != 1 {
			t.Fatalf("期望 1 条 Error 日志，得到 %d", errorsCount)
		}
	})
	t.Run("恐慌", func(t *testing.T) {
		logger := &fakeLogger{}
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}), WithLogger(logger))
		app.AddCommand(&Command{
			Name: "panic",
			Action: func(ctx context.Context, c *Context) error {
				panic("崩溃")
			},
		})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"panic"}), ExitFailure)
		_, _, _, errorsCount := logger.counts()
		if errorsCount != 1 {
			t.Fatalf("期望 1 条 Error 日志，得到 %d", errorsCount)
		}
	})
	t.Run("根命令成功", func(t *testing.T) {
		logger := &fakeLogger{}
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}), WithLogger(logger),
			WithRootAction(func(ctx context.Context, c *Context) error { return nil }))
		testx.RequireEqual(t, app.Execute(context.Background(), nil), ExitOK)
		_, infos, _, _ := logger.counts()
		if infos != 1 {
			t.Fatalf("期望 1 条 Info 日志，得到 %d", infos)
		}
	})
	t.Run("无日志器", func(t *testing.T) {
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}))
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"--version"}), ExitOK)
	})
}

func TestExecuteInvalidAppCodeMapsToUsage(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	// 直接构造分发无法触发 CodeInvalidApp，这里通过命令返回同名错误码验证映射分支。
	app.AddCommand(&Command{
		Name: "bad",
		Action: func(ctx context.Context, c *Context) error {
			return errx.NewCode(CodeInvalidApp, "配置非法")
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"bad"}), ExitUsage)
}
