# clix

轻量、确定性、可测试的 Go CLI 框架：自用协议，与 errx / logx / confx /
validx 家族天然打通。

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.26-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "os"
    "strings"

    "github.com/lcylpzls/clix"
)

func main() {
    app, err := clix.New("greet", "0.1.0",
        clix.WithDescription("问候示例"),
        clix.WithIO(os.Stdout, os.Stderr),
    )
    if err != nil {
        panic(err)
    }
    if err := app.AddCommand(&clix.Command{
        Name:        "hello",
        Description: "向指定名称问好",
        Action: func(ctx context.Context, c *clix.Context) error {
            fmt.Fprintf(c.Out(), "你好，%s！\n", strings.Join(c.Args, " "))
            return nil
        },
    }); err != nil {
        panic(err)
    }
    os.Exit(app.Execute(context.Background(), os.Args[1:]))
}
```

## 核心特性

- **命令分发**：根 Action + 子命令，`-h/--help/help [命令]/--version`
  内置约定；
- **退出码约定**：0 成功、1 执行失败、2 用法错误、130 取消；
- **errx 家族错误**：用法/取消/恐慌均结构化错误码，Action 错误原样透传；
- **logx 结构化日志**：可选注入，记录命令开始/成功/失败与耗时；
- **确定性输出**：Out/Err 可注入，帮助文本返回字符串，便于快照测试；
- **无全局状态**：不隐式调用 `os.Exit`，`Execute` 只返回退出码。

## 文档

- [docs/research.md](docs/research.md) — 竞品调研与取舍
- [docs/design.md](docs/design.md) — 设计
- [docs/architecture.md](docs/architecture.md) — 架构
- [docs/api.md](docs/api.md) — API 快照
- [docs/roadmap.md](docs/roadmap.md) — 路线图

## License

MIT © [lcylpzls](https://github.com/lcylpzls)

