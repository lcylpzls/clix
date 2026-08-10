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
	b.WriteString(renderCommandList(a.commandList()))
	if len(a.globalFlags) > 0 {
		b.WriteString("\n全局选项:\n")
		width := 0
		for _, f := range a.globalFlags {
			if l := len(flagDisplay(f)); l > width {
				width = l
			}
		}
		for _, f := range a.globalFlags {
			fmt.Fprintf(&b, "  %-*s  %s\n", width, flagDisplay(f), flagHelpDesc(f))
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
	return a.renderCommandHelp(cmd), nil
}

// renderCommandHelp 渲染命令帮助文本；调用方必须保证命令存在。
func (a *App) renderCommandHelp(cmd *Command) string {
	var b strings.Builder
	usage := strings.TrimSpace(cmd.Usage)
	if usage == "" {
		usage = fmt.Sprintf("%s %s [选项...] [参数...]", a.name, cmd.FullName())
	}
	fmt.Fprintf(&b, "用法:\n  %s\n", usage)
	if cmd.Description != "" {
		fmt.Fprintf(&b, "\n描述:\n  %s\n", cmd.Description)
	}
	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(&b, "\n别名:\n  %s\n", strings.Join(cmd.Aliases, "、"))
	}
	if children := cmd.children.list(); len(children) > 0 {
		b.WriteString("\n子命令:\n")
		b.WriteString(renderCommandList(children))
	}
	if len(cmd.Args) > 0 {
		b.WriteString("\n参数:\n")
		width := 0
		for _, arg := range cmd.Args {
			if len(arg.Name) > width {
				width = len(arg.Name)
			}
		}
		for _, arg := range cmd.Args {
			display := arg.Name
			if arg.Variadic {
				display += "..."
			}
			fmt.Fprintf(&b, "  %-*s  %s\n", width+3, display, argHelpDesc(arg))
		}
	}
	if len(cmd.Flags) > 0 {
		b.WriteString("\n选项:\n")
		width := 0
		for _, f := range cmd.Flags {
			line := flagDisplay(f)
			if len(line) > width {
				width = len(line)
			}
		}
		for _, f := range cmd.Flags {
			fmt.Fprintf(&b, "  %-*s  %s\n", width, flagDisplay(f), flagHelpDesc(f))
		}
		b.WriteString("  " + padRight("-h, --help", width) + "  显示帮助\n")
	}
	return b.String()
}

// renderCommandList 渲染命令列表：按 Group 分组（空分组无标题），
// 组内保持注册顺序；命令名后展示别名。
func renderCommandList(cmds []*Command) string {
	if len(cmds) == 0 {
		return "  （无）\n"
	}
	type group struct {
		name  string
		items []*Command
	}
	var groups []*group
	index := make(map[string]*group)
	for _, c := range cmds {
		g, ok := index[c.Group]
		if !ok {
			g = &group{name: c.Group}
			index[c.Group] = g
			groups = append(groups, g)
		}
		g.items = append(g.items, c)
	}
	width := 0
	for _, g := range groups {
		for _, c := range g.items {
			if l := len(commandDisplay(c)); l > width {
				width = l
			}
		}
	}
	var b strings.Builder
	for _, g := range groups {
		if g.name != "" {
			fmt.Fprintf(&b, "  %s:\n", g.name)
		}
		for _, c := range g.items {
			fmt.Fprintf(&b, "    %-*s  %s\n", width, commandDisplay(c), c.Description)
		}
	}
	return b.String()
}

// commandDisplay 返回命令显示名（含别名后缀）。
func commandDisplay(c *Command) string {
	if len(c.Aliases) == 0 {
		return c.Name
	}
	return c.Name + "(" + strings.Join(c.Aliases, ",") + ")"
}

// flagDisplay 返回 flag 的显示行（--name 类型）。
func flagDisplay(f FlagSpec) string {
	return fmt.Sprintf("--%s %s", f.Name, flagTypeName(f.kind))
}

// flagTypeName 返回 flag 类型的帮助标签。
func flagTypeName(kind ValueKind) string {
	switch kind {
	case KindString, KindEnum:
		return "string"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindInt64:
		return "int64"
	case KindFloat64:
		return "float"
	case KindDuration:
		return "duration"
	case KindStringSlice:
		return "string[]"
	default:
		return "value"
	}
}

// flagHelpDesc 组合 flag 的说明、必填、默认值与枚举允许值。
func flagHelpDesc(f FlagSpec) string {
	var marks []string
	if f.required {
		marks = append(marks, "必填")
	}
	if f.kind == KindEnum {
		marks = append(marks, "可选值："+strings.Join(f.Allowed, "、"))
	}
	if f.defaultVal != nil {
		marks = append(marks, "默认 "+fmt.Sprint(f.defaultVal))
	}
	if f.env != "" {
		marks = append(marks, "环境变量 "+f.env)
	}
	if f.validate != "" {
		marks = append(marks, "校验 "+f.validate)
	}
	desc := f.Usage
	if len(marks) > 0 {
		desc += "（" + strings.Join(marks, "；") + "）"
	}
	return desc
}

// argHelpDesc 组合位置参数的说明、必填与可重复标记。
func argHelpDesc(arg ArgSpec) string {
	var marks []string
	if arg.Required {
		marks = append(marks, "必填")
	}
	if arg.Variadic {
		marks = append(marks, "可重复")
	}
	desc := arg.Description
	if len(marks) > 0 {
		desc += "（" + strings.Join(marks, "，") + "）"
	}
	return desc
}

// padRight 将文本右补空格到指定宽度。
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
