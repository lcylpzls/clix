// Package clix 是轻量、确定性、可测试的 CLI 框架。
//
// clix 提供 App/Command 模型、内置帮助与版本、errx 结构化错误码与
// 退出码映射、logx 可选结构化日志，并支持注入 Out/Err 输出流。
//
// 典型用法：
//
//	app, err := clix.New("greet", "0.1.0")
//	// ...
//	err = app.AddCommand(&clix.Command{
//	    Name:        "hello",
//	    Description: "向指定名称问好",
//	    Action: func(ctx context.Context, c *clix.Context) error {
//	        fmt.Fprintf(c.Out(), "你好，%s！\n", strings.Join(c.Args, " "))
//	        return nil
//	    },
//	})
//	// ...
//	os.Exit(app.Execute(context.Background(), os.Args[1:]))
package clix
