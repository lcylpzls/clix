package core

import (
	"context"
	"time"
)

// Observer 是命令生命周期观察者（零依赖可选接口，默认无）。
// 由 eventx / tracex / metricsx 等外部适配器接入。
type Observer interface {
	// OnCommandStart 在命令 Action 执行前调用。
	OnCommandStart(ctx context.Context, command string, args []string)
	// OnCommandFinish 在命令 Action 执行后调用（含 panic 恢复后的错误）。
	OnCommandFinish(ctx context.Context, command string, args []string, err error, duration time.Duration)
}
