package clix

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestObserver(t *testing.T) {
	obs := &fakeObserver{}
	app, _ := New("greet", "0.1.0",
		WithIO(&bytes.Buffer{}, &bytes.Buffer{}),
		WithObserver(obs),
	)
	_ = app.AddCommand(&Command{Name: "hello", Action: okAction})
	if code := app.Execute(context.Background(), []string{"hello", "a"}); code != ExitOK {
		t.Fatalf("期望退出码 0，得到 %d", code)
	}
	starts, finishes := obs.counts()
	if starts != 1 || finishes != 1 {
		t.Fatalf("观察者调用次数不匹配：start=%d finish=%d", starts, finishes)
	}
	finish := obs.lastFinish()
	if finish.command != "hello" || strings.Join(finish.args, ",") != "a" ||
		finish.err != nil || finish.duration < 0 {
		t.Fatalf("finish 事件不匹配：%+v", finish)
	}
}

func TestObserverRootAction(t *testing.T) {
	obs := &fakeObserver{}
	app, _ := New("greet", "0.1.0",
		WithIO(&bytes.Buffer{}, &bytes.Buffer{}),
		WithObserver(obs),
		WithRootAction(func(ctx context.Context, c *Context) error { return nil }),
	)
	if code := app.Execute(context.Background(), nil); code != ExitOK {
		t.Fatalf("期望退出码 0，得到 %d", code)
	}
	if f := obs.lastFinish(); f.command != "（根命令）" {
		t.Fatalf("根命令观察事件不匹配：%+v", f)
	}
}

func TestObserverHelpNotTriggered(t *testing.T) {
	obs := &fakeObserver{}
	app, _ := New("greet", "0.1.0",
		WithIO(&bytes.Buffer{}, &bytes.Buffer{}),
		WithObserver(obs),
	)
	if code := app.Execute(context.Background(), []string{"--help"}); code != ExitOK {
		t.Fatalf("期望退出码 0，得到 %d", code)
	}
	starts, finishes := obs.counts()
	if starts != 0 || finishes != 0 {
		t.Fatalf("帮助不应触发观察者：start=%d finish=%d", starts, finishes)
	}
}

func TestObserverErrorAndPanic(t *testing.T) {
	t.Run("错误", func(t *testing.T) {
		obs := &fakeObserver{}
		app, _ := New("greet", "0.1.0",
			WithIO(&bytes.Buffer{}, &bytes.Buffer{}),
			WithObserver(obs),
		)
		_ = app.AddCommand(&Command{
			Name: "boom",
			Action: func(ctx context.Context, c *Context) error {
				return errors.New("失败")
			},
		})
		app.Execute(context.Background(), []string{"boom"})
		if f := obs.lastFinish(); f.err == nil || f.err.Error() != "失败" {
			t.Fatalf("finish 应携带错误：%+v", f)
		}
	})
	t.Run("panic", func(t *testing.T) {
		obs := &fakeObserver{}
		app, _ := New("greet", "0.1.0",
			WithIO(&bytes.Buffer{}, &bytes.Buffer{}),
			WithObserver(obs),
		)
		_ = app.AddCommand(&Command{
			Name: "panic",
			Action: func(ctx context.Context, c *Context) error {
				panic("崩溃")
			},
		})
		app.Execute(context.Background(), []string{"panic"})
		if f := obs.lastFinish(); f.err == nil {
			t.Fatalf("panic 后 finish 应携带恢复错误：%+v", f)
		}
	})
}

type observerFinish struct {
	command  string
	args     []string
	err      error
	duration time.Duration
}

type fakeObserver struct {
	mu       sync.Mutex
	starts   int
	finishes []observerFinish
}

func (o *fakeObserver) OnCommandStart(_ context.Context, command string, args []string) {
	o.mu.Lock()
	o.starts++
	o.mu.Unlock()
}

func (o *fakeObserver) OnCommandFinish(_ context.Context, command string, args []string, err error, duration time.Duration) {
	o.mu.Lock()
	o.finishes = append(o.finishes, observerFinish{
		command: command, args: append([]string(nil), args...),
		err: err, duration: duration,
	})
	o.mu.Unlock()
}

func (o *fakeObserver) counts() (int, int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.starts, len(o.finishes)
}

func (o *fakeObserver) lastFinish() observerFinish {
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.finishes) == 0 {
		return observerFinish{}
	}
	return o.finishes[len(o.finishes)-1]
}
