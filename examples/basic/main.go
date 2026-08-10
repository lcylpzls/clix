package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/lcylpzls/clix"
	"github.com/lcylpzls/confx"
)

func main() {
	app, err := newApp(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	os.Exit(app.Execute(context.Background(), os.Args[1:]))
}

// newApp 构建问候示例应用，测试与 main 共用同一构造。
func newApp(out, errOut io.Writer) (*clix.App, error) {
	app, err := clix.New("greet", "0.1.0",
		clix.WithDescription("问候示例 CLI"),
		clix.WithIO(out, errOut),
		clix.WithGlobalFlags(clix.BoolFlag("verbose", "详细输出").Env("GREET_VERBOSE")),
	)
	if err != nil {
		return nil, err
	}
	if err := app.AddCommand(&clix.Command{
		Name:        "hello",
		Description: "向指定名称问好",
		Action: func(ctx context.Context, c *clix.Context) error {
			if c.GlobalBool("verbose") {
				fmt.Fprintln(c.Out(), "详细模式")
			}
			name := strings.Join(c.Args, " ")
			if name == "" {
				name = "世界"
			}
			fmt.Fprintf(c.Out(), "你好，%s！\n", name)
			return nil
		},
	}); err != nil {
		return nil, err
	}
	if err := app.AddCommand(&clix.Command{
		Name:        "sum",
		Description: "对若干数字求和或求平均值",
		Args: []clix.ArgSpec{
			{Name: "numbers", Description: "参与计算的数字", Required: true, Variadic: true},
		},
		Flags: []clix.FlagSpec{
			clix.IntFlag("base", "基数").Default(0),
			clix.EnumFlag("mode", "计算模式", "sum", "average").Default("sum"),
		},
		Action: func(ctx context.Context, c *clix.Context) error {
			total := c.Int("base")
			for _, raw := range c.Args {
				n, err := strconv.Atoi(raw)
				if err != nil {
					return fmt.Errorf("无效数字 %q", raw)
				}
				total += n
			}
			if c.Enum("mode") == "average" && len(c.Args) > 0 {
				total /= len(c.Args)
			}
			fmt.Fprintf(c.Out(), "%d\n", total)
			return nil
		},
	}); err != nil {
		return nil, err
	}
	remote := &clix.Command{
		Name:        "remote",
		Description: "远端仓库管理",
		Group:       "工具",
	}
	if err := remote.AddCommand(&clix.Command{
		Name:        "list",
		Description: "列出远端仓库",
		Aliases:     []string{"ls"},
		Action: func(ctx context.Context, c *clix.Context) error {
			fmt.Fprintln(c.Out(), "origin")
			return nil
		},
	}); err != nil {
		return nil, err
	}
	if err := app.AddCommand(remote); err != nil {
		return nil, err
	}
	var cfg struct {
		Greeting string `toml:"greeting"`
		Retries  int    `toml:"retries"`
	}
	if err := app.AddCommand(&clix.Command{
		Name:        "config",
		Description: "加载 TOML 配置并打印问候语",
		Flags: []clix.FlagSpec{
			clix.StringFlag("path", "配置文件路径").Env("GREET_CONFIG").Default("config.toml").Validate("required"),
			clix.StringFlag("mode", "运行模式").Default("dev").Validate("oneof=dev prod"),
		},
		Before: func(ctx context.Context, c *clix.Context) error {
			manager, err := confx.NewConfigManager(confx.Toml)
			if err != nil {
				return err
			}
			return clix.LoadConfig(ctx, c, manager, "path", "", &cfg)
		},
		Action: func(ctx context.Context, c *clix.Context) error {
			fmt.Fprintf(c.Out(), "问候语：%s（重试 %d 次，模式 %s）\n",
				cfg.Greeting, cfg.Retries, c.String("mode"))
			return nil
		},
	}); err != nil {
		return nil, err
	}
	return app, nil
}
