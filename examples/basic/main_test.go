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

func TestGreetCLISum(t *testing.T) {
	var out bytes.Buffer
	app, err := newApp(&out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"sum", "1", "2", "3", "--base", "10"}); code != clix.ExitOK {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitOK, code)
	}
	if !strings.Contains(out.String(), "16") {
		t.Fatalf("求和输出缺失：%s", out.String())
	}
}

func TestGreetCLISumAverage(t *testing.T) {
	var out bytes.Buffer
	app, err := newApp(&out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"sum", "2", "4", "6", "--mode", "average"}); code != clix.ExitOK {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitOK, code)
	}
	if !strings.Contains(out.String(), "4") {
		t.Fatalf("平均值输出缺失：%s", out.String())
	}
}

func TestGreetCLISumUsageError(t *testing.T) {
	var errBuf bytes.Buffer
	app, err := newApp(&bytes.Buffer{}, &errBuf)
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"sum", "1", "--mode", "slow"}); code != clix.ExitUsage {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitUsage, code)
	}
	if !strings.Contains(errBuf.String(), "不在允许列表") {
		t.Fatalf("错误输出缺失：%s", errBuf.String())
	}
}

func TestGreetCLINestedRemote(t *testing.T) {
	var out bytes.Buffer
	app, err := newApp(&out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"remote", "ls"}); code != clix.ExitOK {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitOK, code)
	}
	if !strings.Contains(out.String(), "origin") {
		t.Fatalf("子命令输出缺失：%s", out.String())
	}
}

func TestGreetCLINestedRemoteMissingSubcommand(t *testing.T) {
	var errBuf bytes.Buffer
	app, err := newApp(&bytes.Buffer{}, &errBuf)
	if err != nil {
		t.Fatalf("构造失败：%v", err)
	}
	if code := app.Execute(context.Background(), []string{"remote"}); code != clix.ExitUsage {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitUsage, code)
	}
	if !strings.Contains(errBuf.String(), "需要子命令") {
		t.Fatalf("错误输出缺失：%s", errBuf.String())
	}
}
