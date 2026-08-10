package clix

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

func TestNestedCommands(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	parent := &Command{Name: "parent", Description: "父命令"}
	zh := &Command{
		Name:        "zh",
		Description: "中文",
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString("中文执行:" + strings.Join(c.Args, ",") + "\n")
			return nil
		},
	}
	en := &Command{
		Name:        "en",
		Description: "英文",
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString("英文执行\n")
			return nil
		},
	}
	testx.RequireNoError(t, parent.AddCommand(zh))
	testx.RequireNoError(t, parent.AddCommand(en))
	testx.RequireNoError(t, app.AddCommand(parent))

	testx.RequireEqual(t, app.Execute(context.Background(), []string{"parent", "zh", "a", "b"}), ExitOK)
	if !strings.Contains(out.String(), "中文执行:a,b") {
		t.Fatalf("子命令输出缺失：%s", out.String())
	}
	if zh.FullName() != "parent zh" {
		t.Fatalf("FullName 不匹配：%q", zh.FullName())
	}
	if zh.Parent() != parent {
		t.Fatal("Parent 不匹配")
	}
	if parent.Parent() != nil {
		t.Fatal("顶层命令 Parent 应为 nil")
	}
}

func TestNestedCommandAliases(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	parent := &Command{Name: "parent"}
	zh := &Command{
		Name:    "zh",
		Aliases: []string{"cn", "zhs"},
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString("别名命中\n")
			return nil
		},
	}
	testx.RequireNoError(t, parent.AddCommand(zh))
	testx.RequireNoError(t, app.AddCommand(parent))
	for _, path := range [][]string{{"parent", "cn"}, {"parent", "zhs"}} {
		out.Reset()
		testx.RequireEqual(t, app.Execute(context.Background(), path), ExitOK)
		testx.RequireTrue(t, strings.Contains(out.String(), "别名命中"))
	}
}

func TestCommandWithActionAndSubcommands(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	parent := &Command{
		Name: "parent",
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString("父命令执行\n")
			return nil
		},
	}
	child := &Command{
		Name: "child",
		Action: func(ctx context.Context, c *Context) error {
			out.WriteString("子命令执行\n")
			return nil
		},
	}
	testx.RequireNoError(t, parent.AddCommand(child))
	testx.RequireNoError(t, app.AddCommand(parent))
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"parent"}), ExitOK)
	if !strings.Contains(out.String(), "父命令执行") {
		t.Fatalf("父命令未执行：%s", out.String())
	}
	out.Reset()
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"parent", "child"}), ExitOK)
	if !strings.Contains(out.String(), "子命令执行") {
		t.Fatalf("子命令未执行：%s", out.String())
	}
}

func TestNestedCommandErrors(t *testing.T) {
	t.Run("缺少子命令", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
		parent := &Command{Name: "parent"}
		_ = parent.AddCommand(&Command{Name: "child", Action: okAction})
		_ = app.AddCommand(parent)
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"parent"}), ExitUsage)
		if !strings.Contains(errBuf.String(), "需要子命令") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
	t.Run("未知子命令", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
		parent := &Command{Name: "parent"}
		_ = parent.AddCommand(&Command{Name: "child", Action: okAction})
		_ = app.AddCommand(parent)
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"parent", "nope"}), ExitUsage)
		if !strings.Contains(errBuf.String(), "未知子命令 \"nope\"") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
	t.Run("深层未知", func(t *testing.T) {
		var errBuf bytes.Buffer
		app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
		parent := &Command{Name: "parent"}
		child := &Command{Name: "child"}
		_ = child.AddCommand(&Command{Name: "leaf", Action: okAction})
		_ = parent.AddCommand(child)
		_ = app.AddCommand(parent)
		testx.RequireEqual(t, app.Execute(context.Background(), []string{"parent", "child", "nope"}), ExitUsage)
		if !strings.Contains(errBuf.String(), "未知子命令 \"nope\"（命令 \"parent child\" 下）") {
			t.Fatalf("错误信息缺失：%s", errBuf.String())
		}
	})
}

func TestNestedHelp(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	parent := &Command{Name: "parent", Description: "父命令", Aliases: []string{"p"}}
	zh := &Command{Name: "zh", Description: "中文", Action: okAction}
	_ = parent.AddCommand(zh)
	_ = app.AddCommand(parent)

	testx.RequireEqual(t, app.Execute(context.Background(), []string{"parent", "--help"}), ExitOK)
	text := out.String()
	for _, want := range []string{
		"用法:\n  greet parent [选项...] [参数...]",
		"别名:\n  p",
		"子命令:",
		"zh",
	} {
		testx.RequireTrue(t, strings.Contains(text, want))
	}

	out.Reset()
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"help", "parent", "zh"}), ExitOK)
	if !strings.Contains(out.String(), "用法:\n  greet parent zh [选项...] [参数...]") {
		t.Fatalf("子命令帮助缺失：\n%s", out.String())
	}
}

func TestHelpPathUnknown(t *testing.T) {
	var errBuf bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
	parent := &Command{Name: "parent"}
	_ = parent.AddCommand(&Command{Name: "child", Action: okAction})
	_ = app.AddCommand(parent)
	testx.RequireEqual(t, app.Execute(context.Background(), []string{"help", "parent", "nope"}), ExitUsage)
	if !strings.Contains(errBuf.String(), "未知命令 \"nope\"") {
		t.Fatalf("错误信息缺失：%s", errBuf.String())
	}
}

func TestHelpGroupsAndAliases(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	_ = app.AddCommand(&Command{Name: "hello", Description: "问候", Group: "基础", Aliases: []string{"hi"}, Action: okAction})
	_ = app.AddCommand(&Command{Name: "sum", Description: "求和", Group: "工具", Action: okAction})
	_ = app.AddCommand(&Command{Name: "raw", Description: "未分组", Action: okAction})
	text := app.HelpText()
	for _, want := range []string{
		"基础:",
		"hello(hi)  问候",
		"工具:",
		"sum        求和",
		"raw        未分组",
	} {
		testx.RequireTrue(t, strings.Contains(text, want))
	}
}

func TestNilRegistry(t *testing.T) {
	var r *registry
	if r.lookup("x") != nil {
		t.Fatal("nil registry lookup 应返回 nil")
	}
	if r.list() != nil {
		t.Fatal("nil registry list 应返回 nil")
	}
	if r.count() != 0 {
		t.Fatal("nil registry count 应为 0")
	}
}

func TestRenderCommandListEmpty(t *testing.T) {
	if got := renderCommandList(nil); got != "  （无）\n" {
		t.Fatalf("空列表渲染不匹配：%q", got)
	}
}

func TestCommandRegistrationValidation(t *testing.T) {
	tests := []struct {
		name string
		run  func(app *App) error
		code errx.Code
	}{
		{"空别名", func(app *App) error {
			return app.AddCommand(&Command{Name: "a", Aliases: []string{" "}, Action: okAction})
		}, CodeInvalidApp},
		{"非法别名", func(app *App) error {
			return app.AddCommand(&Command{Name: "a", Aliases: []string{"-x"}, Action: okAction})
		}, CodeInvalidApp},
		{"别名与命令名相同", func(app *App) error {
			return app.AddCommand(&Command{Name: "a", Aliases: []string{"a"}, Action: okAction})
		}, CodeInvalidApp},
		{"别名与去空格命令名相同", func(app *App) error {
			return app.AddCommand(&Command{Name: " a ", Aliases: []string{"a"}, Action: okAction})
		}, CodeInvalidApp},
		{"别名重复", func(app *App) error {
			return app.AddCommand(&Command{Name: "a", Aliases: []string{"x", "x"}, Action: okAction})
		}, CodeInvalidApp},
		{"命令名与别名冲突", func(app *App) error {
			if err := app.AddCommand(&Command{Name: "a", Aliases: []string{"b"}, Action: okAction}); err != nil {
				return err
			}
			return app.AddCommand(&Command{Name: "b", Action: okAction})
		}, CodeInvalidApp},
		{"别名与命令名冲突", func(app *App) error {
			if err := app.AddCommand(&Command{Name: "a", Action: okAction}); err != nil {
				return err
			}
			return app.AddCommand(&Command{Name: "b", Aliases: []string{"a"}, Action: okAction})
		}, CodeInvalidApp},
		{"别名重复注册", func(app *App) error {
			if err := app.AddCommand(&Command{Name: "a", Aliases: []string{"x"}, Action: okAction}); err != nil {
				return err
			}
			return app.AddCommand(&Command{Name: "b", Aliases: []string{"x"}, Action: okAction})
		}, CodeInvalidApp},
		{"重复注册同一命令", func(app *App) error {
			cmd := &Command{Name: "a", Action: okAction}
			if err := app.AddCommand(cmd); err != nil {
				return err
			}
			return app.AddCommand(cmd)
		}, CodeInvalidApp},
		{"子命令名冲突", func(app *App) error {
			parent := &Command{Name: "parent"}
			if err := parent.AddCommand(&Command{Name: "x", Action: okAction}); err != nil {
				return err
			}
			return parent.AddCommand(&Command{Name: "x", Action: okAction})
		}, CodeInvalidApp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(t)
			err := tt.run(app)
			assertErrCode(t, err, tt.code)
		})
	}
}

func TestNilCommandMethods(t *testing.T) {
	var c *Command
	if c.Parent() != nil {
		t.Fatal("nil 命令 Parent 应为 nil")
	}
	if c.FullName() != "" {
		t.Fatalf("nil 命令 FullName 应为空，得到 %q", c.FullName())
	}
	err := c.AddCommand(&Command{Name: "x", Action: okAction})
	assertErrCode(t, err, CodeInvalidApp)
}
