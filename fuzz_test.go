package clix

import (
	"bytes"
	"context"
	testx "github.com/lcylpzls/testx"
	"strings"
	"testing"
)

// FuzzExecute 验证任意参数输入下分发逻辑不 panic、不产生异常输出。
func FuzzExecute(f *testing.F) {
	f.Add("")
	f.Add("--help")
	f.Add("--version")
	f.Add("help")
	f.Add("help hello")
	f.Add("hello 小明 小红")
	f.Add("sum 1 2")
	f.Add("sum 1 2 --base 10")
	f.Add("parent child")
	f.Add("parent nope")
	f.Add("help parent child")
	f.Add("nope")
	f.Add("help nope")
	f.Add("-h")
	f.Add("hello --verbose --mode fast")
	f.Fuzz(func(t *testing.T, input string) {
		app, err := New("fuzz", "0.1.0",
			WithIO(&bytes.Buffer{}, &bytes.Buffer{}),
			WithRootAction(func(ctx context.Context, c *Context) error { return nil }),
			WithGlobalFlags(BoolFlag("verbose", "详细输出")),
		)
		testx.RequireNoError(t, err)

		_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
		_ = app.AddCommand(&Command{
			Name: "sum",
			Action: func(ctx context.Context, c *Context) error {
				return nil
			},
		})
		parent := &Command{Name: "parent"}
		_ = parent.AddCommand(&Command{Name: "child", Action: okAction})
		_ = app.AddCommand(parent)
		app.AddCommand(&Command{
			Name:   "hook",
			Before: func(ctx context.Context, c *Context) error { return nil },
			Action: okAction,
			After:  func(ctx context.Context, c *Context) error { return nil },
		})
		app.Execute(context.Background(), strings.Fields(input))
	})
}

// FuzzParseCommandArgs 验证任意参数输入下 flag/位置参数解析不 panic、不越界。
func FuzzParseCommandArgs(f *testing.F) {
	args := []ArgSpec{
		{Name: "name", Required: true},
		{Name: "rest", Variadic: true},
	}
	flags := []FlagSpec{
		StringFlag("output", "输出路径").Default("out.txt"),
		BoolFlag("verbose", "详细输出"),
		IntFlag("retry", "重试次数").Default(3),
		EnumFlag("mode", "模式", "fast", "slow"),
		StringSliceFlag("tag", "标签"),
		StringFlag("name", "名称").Validate("required,min=2"),
	}
	f.Add("")
	f.Add("hello --verbose")
	f.Add("--mode fast --retry 3")
	f.Add("--tag a --tag b")
	f.Add("-- --mode")
	f.Fuzz(func(t *testing.T, input string) {
		_, _, _ = parseCommandArgs(args, flags, strings.Fields(input))
	})
}

// FuzzCommandHelpText 验证任意命令名输入下帮助渲染不 panic。
func FuzzCommandHelpText(f *testing.F) {
	f.Add("hello")
	f.Add("nope")
	f.Add("")
	f.Add("--help")
	f.Fuzz(func(t *testing.T, name string) {
		app := newTestApp(t)
		_ = app.AddCommand(&Command{Name: "hello", Description: "问候", Action: okAction})
		_, _ = app.CommandHelpText(name)
	})
}
