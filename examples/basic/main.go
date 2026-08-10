package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/lcylpzls/clix"
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
	)
	if err != nil {
		return nil, err
	}
	if err := app.AddCommand(&clix.Command{
		Name:        "hello",
		Description: "向指定名称问好",
		Action: func(ctx context.Context, c *clix.Context) error {
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
	return app, nil
}
