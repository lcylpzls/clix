package clix

import (
	"bytes"
	"context"
	"errors"
	"github.com/lcylpzls/testx"
	"strings"
	"testing"
)

func TestBeforeAfterHooksOrder(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	_ = app.AddCommand(&Command{
		Name: "run",
		Before: func(ctx context.Context, c *Context) error {
			out.WriteString("before\n")
			return nil
		},
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString("action\n")
			return nil
		},
		After: func(ctx context.Context, c *Context) error {
			out.WriteString("after\n")
			return nil
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitOK)
	if got := out.String(); got != "before\naction\nafter\n" {
		t.Fatalf("钩子顺序不匹配：%q", got)
	}
}

func TestBeforeHookErrorAborts(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	_ = app.AddCommand(&Command{
		Name: "run",
		Before: func(ctx context.Context, c *Context) error {
			return errors.New("前置校验失败")
		},
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString("不应执行\n")
			return nil
		},
		After: func(ctx context.Context, c *Context) error {
			out.WriteString("不应执行 After\n")
			return nil
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if out.Len() != 0 {
		t.Fatalf("Before 失败后不应执行 Action/After：%s", out.String())
	}
	if !strings.Contains(errBuf.String(), "前置校验失败") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestAfterHookOnActionError(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &errBuf))
	_ = app.AddCommand(&Command{
		Name: "run",
		Action: func(ctx context.Context, c *Context) error {
			return errors.New("动作失败")
		},
		After: func(ctx context.Context, c *Context) error {
			out.WriteString("清理执行\n")
			return nil
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if !strings.Contains(out.String(), "清理执行") {
		t.Fatalf("Action 失败后 After 应执行：%s", out.String())
	}
	if !strings.Contains(errBuf.String(), "动作失败") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestAfterHookError(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	_ = app.AddCommand(&Command{
		Name:   "run",
		Action: okAction,
		After: func(ctx context.Context, c *Context) error {
			return errors.New("清理失败")
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if !strings.Contains(errBuf.String(), "清理失败") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestBothHooksFailAggregate(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	_ = app.AddCommand(&Command{
		Name: "run",
		Action: func(ctx context.Context, c *Context) error {
			return errors.New("动作失败")
		},
		After: func(ctx context.Context, c *Context) error {
			return errors.New("清理失败")
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if !strings.Contains(errBuf.String(), "动作失败") || !strings.Contains(errBuf.String(), "清理失败") {
		t.Fatalf("两个错误都应输出：%s", errBuf.String())
	}
}

func TestBeforeHookPanic(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	_ = app.AddCommand(&Command{
		Name: "run",
		Before: func(ctx context.Context, c *Context) error {
			panic("前置崩溃")
		},
		Action: okAction,
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if !strings.Contains(errBuf.String(), "CLI_ACTION_PANIC") {
		t.Fatalf("恐慌错误码缺失：%s", errBuf.String())
	}
}

func TestAfterHookPanic(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	_ = app.AddCommand(&Command{
		Name:   "run",
		Action: okAction,
		After: func(ctx context.Context, c *Context) error {
			panic("清理崩溃")
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if !strings.Contains(errBuf.String(), "CLI_ACTION_PANIC") {
		t.Fatalf("恐慌错误码缺失：%s", errBuf.String())
	}
}
