package clix

import (
	"bytes"
	"context"
	testx "github.com/lcylpzls/testx"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lcylpzls/confx"
)

func TestLoadConfigSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	testx.RequireNoError(t, os.WriteFile(path, []byte("greeting = \"你好\"\nretries = 3\n"), 0o600))
	var cfg struct {
		Greeting string `toml:"greeting"`
		Retries  int    `toml:"retries"`
	}
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	_ = app.AddCommand(&Command{
		Name:  "config",
		Flags: []FlagSpec{StringFlag("path", "配置文件路径")},
		Before: func(ctx context.Context, c *Context) error {
			manager, err := confx.NewConfigManager(confx.Toml)
			if err != nil {
				return err
			}
			return LoadConfig(ctx, c, manager, "path", "", &cfg)
		},
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString(cfg.Greeting + "\n")
			return nil
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"config", "--path", path}), ExitOK)
	if !strings.Contains(out.String(), "你好") {
		t.Fatalf("配置未加载：%s", out.String())
	}
	if cfg.Retries != 3 {
		t.Fatalf("retries 未加载：%d", cfg.Retries)
	}
}

func TestLoadConfigFlagWinsAndFallback(t *testing.T) {
	dir := t.TempDir()
	flagPath := filepath.Join(dir, "flag.toml")
	fallbackPath := filepath.Join(dir, "fallback.toml")
	testx.RequireNoError(t, os.WriteFile(flagPath, []byte("value = \"flag\"\n"), 0o600))
	testx.RequireNoError(t, os.WriteFile(fallbackPath, []byte("value = \"fallback\"\n"), 0o600))
	var got string
	run := func(args []string) int {
		var cfg struct {
			Value string `toml:"value"`
		}
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}))
		_ = app.AddCommand(&Command{
			Name:  "config",
			Flags: []FlagSpec{StringFlag("path", "配置文件路径")},
			Before: func(ctx context.Context, c *Context) error {
				manager, _ := confx.NewConfigManager(confx.Toml)
				return LoadConfig(ctx, c, manager, "path", fallbackPath, &cfg)
			},
			Action: func(ctx context.Context, c *Context) error {
				got = cfg.Value
				return nil
			},
		})
		return app.Execute(context.Background(), args)
	}
	testx.RequireEqual(t, run([]string{"config", "--path", flagPath}), ExitOK)
	testx.RequireEqual(t, got, "flag")

	testx.RequireEqual(t, run([]string{"config"}), ExitOK)
	testx.RequireEqual(t, got, "fallback")

}

func TestLoadConfigDefaultPathFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	testx.RequireNoError(t, os.WriteFile(path, []byte("value = \"ok\"\n"), 0o600))
	var cfg struct {
		Value string `toml:"value"`
	}
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &bytes.Buffer{}))
	_ = app.AddCommand(&Command{
		Name:  "config",
		Flags: []FlagSpec{StringFlag("config", "配置文件路径")},
		Before: func(ctx context.Context, c *Context) error {
			manager, _ := confx.NewConfigManager(confx.Toml)
			return LoadConfig(ctx, c, manager, "", "", &cfg)
		},
		Action: func(ctx context.Context, c *Context) error {
			return nil
		},
	})
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"config", "--config", path}), ExitOK)
	testx.RequireEqual(t, cfg.Value, "ok")

}

func TestLoadConfigErrors(t *testing.T) {
	t.Run("缺少路径", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
		_ = app.AddCommand(&Command{
			Name:  "config",
			Flags: []FlagSpec{StringFlag("path", "配置文件路径")},
			Before: func(ctx context.Context, c *Context) error {
				manager, _ := confx.NewConfigManager(confx.Toml)
				return LoadConfig(ctx, c, manager, "path", "", &struct{}{})
			},
			Action: okAction,
		})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"config"}), ExitUsage)
		if !strings.Contains(errBuf.String(), "缺少配置文件路径") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
	t.Run("读取失败", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
		_ = app.AddCommand(&Command{
			Name:  "config",
			Flags: []FlagSpec{StringFlag("path", "配置文件路径")},
			Before: func(ctx context.Context, c *Context) error {
				manager, _ := confx.NewConfigManager(confx.Toml)
				return LoadConfig(ctx, c, manager, "path", "", &struct{}{})
			},
			Action: okAction,
		})
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"config", "--path", filepath.Join(t.TempDir(), "nope.toml")}), ExitFailure)
		if !strings.Contains(errBuf.String(), "读取配置文件失败") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
	t.Run("nil Context", func(t *testing.T) {
		manager, _ := confx.NewConfigManager(confx.Toml)
		err := LoadConfig(context.Background(), nil, manager, "path", "x.toml", &struct{}{})
		assertErrCode(t, err, CodeInvalidApp)
	})
	t.Run("nil manager", func(t *testing.T) {
		err := LoadConfig(context.Background(), &Context{}, nil, "path", "x.toml", &struct{}{})
		assertErrCode(t, err, CodeInvalidApp)
	})
}
