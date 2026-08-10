package main

import (
	"bytes"
	"context"
	testx "github.com/lcylpzls/testx"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lcylpzls/clix"
)

func TestGreetCLI(t *testing.T) {
	var out, errBuf bytes.Buffer
	app, err := newApp(&out, &errBuf)
	testx.RequireNoError(t, err)

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
	testx.RequireNoError(t, err)

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
	testx.RequireNoError(t, err)

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
	testx.RequireNoError(t, err)

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
	testx.RequireNoError(t, err)

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
	testx.RequireNoError(t, err)

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
	testx.RequireNoError(t, err)

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
	testx.RequireNoError(t, err)

	if code := app.Execute(context.Background(), []string{"remote"}); code != clix.ExitUsage {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitUsage, code)
	}
	if !strings.Contains(errBuf.String(), "需要子命令") {
		t.Fatalf("错误输出缺失：%s", errBuf.String())
	}
}

func TestGreetCLIGlobalFlag(t *testing.T) {
	var out bytes.Buffer
	app, err := newApp(&out, &bytes.Buffer{})
	testx.RequireNoError(t, err)

	if code := app.Execute(context.Background(), []string{"--verbose", "hello", "小明"}); code != clix.ExitOK {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitOK, code)
	}
	if !strings.Contains(out.String(), "详细模式") || !strings.Contains(out.String(), "你好，小明！") {
		t.Fatalf("全局 flag 输出缺失：%s", out.String())
	}
}

func TestGreetCLIConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	if err := os.WriteFile(path, []byte("greeting = \"你好，配置文件！\"\nretries = 5\n"), 0o600); err != nil {
		t.Fatalf("写配置失败：%v", err)
	}
	var out bytes.Buffer
	app, err := newApp(&out, &bytes.Buffer{})
	testx.RequireNoError(t, err)

	if code := app.Execute(context.Background(), []string{"config", "--path", path, "--mode", "prod"}); code != clix.ExitOK {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitOK, code)
	}
	if !strings.Contains(out.String(), "你好，配置文件！") || !strings.Contains(out.String(), "模式 prod") {
		t.Fatalf("配置输出缺失：%s", out.String())
	}
}

func TestGreetCLIConfigValidation(t *testing.T) {
	var errBuf bytes.Buffer
	app, err := newApp(&bytes.Buffer{}, &errBuf)
	testx.RequireNoError(t, err)

	if code := app.Execute(context.Background(), []string{"config", "--mode", "staging"}); code != clix.ExitUsage {
		t.Fatalf("期望退出码 %d，得到 %d", clix.ExitUsage, code)
	}
	if !strings.Contains(errBuf.String(), "CLI_FLAG_VALIDATION_FAILED") {
		t.Fatalf("校验错误缺失：%s", errBuf.String())
	}
}
