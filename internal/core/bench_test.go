package core

import (
	"context"
	"testing"
)

// BenchmarkParseCommandArgs 基准：类型化 flag + 位置参数解析。
func BenchmarkParseCommandArgs(b *testing.B) {
	args := []ArgSpec{
		{Name: "name", Required: true},
		{Name: "rest", Variadic: true},
	}
	flags := []FlagSpec{
		StringFlag("output", "输出路径").Default("out.txt"),
		BoolFlag("verbose", "详细输出"),
		IntFlag("retry", "重试次数").Default(3),
		DurationFlag("timeout", "超时"),
		EnumFlag("mode", "模式", "fast", "slow").Default("fast"),
		StringSliceFlag("tag", "标签"),
	}
	raw := []string{
		"alice", "x",
		"--verbose", "--retry=5", "--mode", "slow",
		"--tag", "a", "--tag=b", "--timeout", "2s",
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, _, err := parseCommandArgs(args, flags, raw); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExecute 基准：完整分发（含解析、帮助检测与 Action 调用）。
func BenchmarkExecute(b *testing.B) {
	app, _ := New("bench", "0.1.0",
		WithIO(discardWriter{}, discardWriter{}),
		WithGlobalFlags(BoolFlag("verbose", "详细输出")))
	_ = app.AddCommand(&Command{
		Name:  "hello",
		Flags: []FlagSpec{StringFlag("name", "名称").Validate("required")},
		Action: func(ctx context.Context, c *Context) error {
			return nil
		},
	})
	args := []string{"--verbose", "hello", "--name", "x"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		app.Execute(context.Background(), args)
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
