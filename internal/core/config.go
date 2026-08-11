package core

import (
	"context"

	"github.com/lcylpzls/confx"
	"github.com/lcylpzls/errx"
)

// LoadConfig 是 confx 联动助手，适合在 Before 钩子中调用：
// 路径优先取 flag pathFlag（默认 "config"）的值，否则使用 fallback；
// 随后通过 manager.Load 加载到 target。错误透传 confx 结构化错误。
func LoadConfig(ctx context.Context, c *Context, manager *confx.ConfigManager, pathFlag, fallback string, target any) error {
	if c == nil {
		return errx.NewCode(CodeInvalidApp, "Context 不能为空")
	}
	if manager == nil {
		return errx.NewCode(CodeInvalidApp, "配置管理器不能为空")
	}
	if pathFlag == "" {
		pathFlag = "config"
	}
	path := fallback
	if v := c.String(pathFlag); v != "" {
		path = v
	}
	if path == "" {
		return errx.NewCodef(CodeMissingRequiredFlag, "缺少配置文件路径：flag %q 未设置且未提供默认路径", pathFlag)
	}
	return manager.Load(path, target)
}
