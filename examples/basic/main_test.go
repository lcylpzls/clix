package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lcylpzls/clix"
)

func TestGreetCLI(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, err := newApp(&out, &errBuf)
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"hello", "小明"}); code != clix.ExitOK {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitOK, code)
	}
	if !strings.Contains(out.String(), "你好，小明！") {
		t.Fatalf("命令输出缺失：%s", out.String())
	}
	if errBuf.Len() != 0 {
		t.Fatalf("错误流应为空：%s", errBuf.String())
	}
}

func TestGreetCLIHelp(t *testing.T) {
	var out bytes.Buffer
	app, err := newApp(&out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"--help"}); code != clix.ExitOK {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitOK, code)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("帮助应包含命令列表：%s", out.String())
	}
}

func TestGreetCLIUsageError(t *testing.T) {
	var errBuf bytes.Buffer
	app, err := newApp(&bytes.Buffer{}, &errBuf)
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"nope"}); code != clix.ExitUsage {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitUsage, code)
	}
	if !strings.Contains(errBuf.String(), "未知命令") {
		t.Fatalf("错误输出缺失：%s", errBuf.String())
	}
}
