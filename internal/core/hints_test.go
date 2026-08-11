package core

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

func TestErrorHint(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0",
		WithIO(&bytes.Buffer{}, &errBuf),
		WithErrorHint("CLI_TEST_HINT", "请检查网络后重试"),
	)
	_ = app.AddCommand(&Command{
		Name: "run",
		Action: func(ctx context.Context, c *Context) error {
			return errx.NewCode("CLI_TEST_HINT", "连接失败")
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if !strings.Contains(errBuf.String(), "提示：请检查网络后重试") {
		t.Fatalf("错误提示缺失：%s", errBuf.String())
	}
}

func TestErrorHintNoMatch(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0",
		WithIO(&bytes.Buffer{}, &errBuf),
		WithErrorHint("CLI_OTHER", "其他提示"),
	)
	_ = app.AddCommand(&Command{
		Name: "run",
		Action: func(ctx context.Context, c *Context) error {
			return errx.NewCode("CLI_TEST_PLAIN", "普通失败")
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if strings.Contains(errBuf.String(), "提示：") {
		t.Fatalf("未注册的错误码不应输出提示：%s", errBuf.String())
	}
}

func TestErrorHintPlainError(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0",
		WithIO(&bytes.Buffer{}, &errBuf),
		WithErrorHint("CLI_TEST_HINT", "提示文本"),
	)
	_ = app.AddCommand(&Command{
		Name: "run",
		Action: func(ctx context.Context, c *Context) error {
			return &plainError{}
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	if strings.Contains(errBuf.String(), "提示：") {
		t.Fatalf("普通错误不应输出提示：%s", errBuf.String())
	}
}

type plainError struct{}

func (e *plainError) Error() string {
	return "普通错误"
}

func TestRetryableErrorLogsWarn(t *testing.T) {
	logger := &fakeLogger{}
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}), WithLogger(logger))
	_ = app.AddCommand(&Command{
		Name: "run",
		Action: func(ctx context.Context, c *Context) error {
			return errx.New(errx.KindTimeout, "CLI_TEST_RETRY", "上游超时")
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"run"}), ExitFailure)
	_, _, warns, errors := logger.counts()
	if warns != 1 || errors != 0 {
		t.Fatalf("可重试错误应记 Warn：warns=%d errors=%d", warns, errors)
	}
}

func TestErrorHintFunctionNoHints(t *testing.T) {
	app := newTestApp(t)
	if got := app.errorHint(nil); got != "" {
		t.Fatalf("无提示表时应返回空：%q", got)
	}
}
