package clix

import (
	"fmt"
	"strings"

	"github.com/lcylpzls/errx"
)

// HelpText 渲染应用级帮助文本。输出确定性文本，便于快照测试。
func (a *App) HelpText() string {
	var b strings.Builder
	if a.description != "" {
		fmt.Fprintf(&b, "%s - %s\n\n", a.name, a.description)
	} else {
		fmt.Fprintf(&b, "%s\n\n", a.name)
	}
	b.WriteString("用法:\n")
	fmt.Fprintf(&b, "  %s\n\n", a.usageLine())
	b.WriteString("命令:\n")
	commands := a.commandList()
	if len(commands) == 0 {
		b.WriteString("  （无）\n")
	} else {
		width := 0
		for _, c := range commands {
			if len(c.Name) > width {
				width = len(c.Name)
			}
		}
		for _, c := range commands {
			fmt.Fprintf(&b, "  %-*s  %s\n", width, c.Name, c.Description)
		}
	}
	b.WriteString("\n选项:\n")
	b.WriteString("  -h, --help     显示帮助\n")
	b.WriteString("      --version  显示版本\n")
	return b.String()
}

// CommandHelpText 渲染指定命令的帮助文本；命令不存在时返回 errx 错误。
func (a *App) CommandHelpText(name string) (string, error) {
	cmd := a.lookup(name)
	if cmd == nil {
		return "", errx.NewCodef(CodeUnknownCommand, "未知命令 %q，请运行 --help 查看可用命令", name)
	}
	var b strings.Builder
	usage := strings.TrimSpace(cmd.Usage)
	if usage == "" {
		usage = fmt.Sprintf("%s %s [参数...]", a.name, cmd.Name)
	}
	fmt.Fprintf(&b, "用法:\n  %s\n", usage)
	if cmd.Description != "" {
		fmt.Fprintf(&b, "\n描述:\n  %s\n", cmd.Description)
	}
	return b.String(), nil
}
