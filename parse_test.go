package clix

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lcylpzls/errx"
)

func TestValidateFlagSpecs(t *testing.T) {
	tests := []struct {
		name  string
		flags []FlagSpec
	}{
		{"空名", []FlagSpec{{Name: ""}}},
		{"数字开头", []FlagSpec{{Name: "1x"}}},
		{"非法字符", []FlagSpec{{Name: "a b"}}},
		{"保留名", []FlagSpec{{Name: "help"}}},
		{"重复定义", []FlagSpec{{Name: "a"}, {Name: "a"}}},
		{"枚举无允许值", []FlagSpec{EnumFlag("mode", "模式")}},
		{"字符串默认值类型错误", []FlagSpec{StringFlag("s", "").Default(1)}},
		{"布尔默认值类型错误", []FlagSpec{BoolFlag("b", "").Default("x")}},
		{"整数默认值类型错误", []FlagSpec{IntFlag("i", "").Default("x")}},
		{"64 位整数默认值类型错误", []FlagSpec{Int64Flag("i64", "").Default("x")}},
		{"浮点默认值类型错误", []FlagSpec{FloatFlag("f", "").Default("x")}},
		{"时长默认值类型错误", []FlagSpec{DurationFlag("d", "").Default("x")}},
		{"切片默认值类型错误", []FlagSpec{StringSliceFlag("sl", "").Default("x")}},
		{"枚举默认值不在列表", []FlagSpec{EnumFlag("mode", "", "fast").Default("slow")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlagSpecs(tt.flags)
			assertErrCode(t, err, CodeInvalidFlagDef)
		})
	}
}

func TestValidateFlagSpecsValid(t *testing.T) {
	flags := []FlagSpec{
		StringFlag("s", "").Default("x"),
		BoolFlag("b", "").Default(true),
		IntFlag("i", "").Default(1),
		Int64Flag("i64", "").Default(int64(1)),
		FloatFlag("f", "").Default(1.5),
		DurationFlag("d", "").Default(time.Second),
		EnumFlag("e", "", "fast").Default("fast"),
		StringSliceFlag("sl", "").Default([]string{"a", "b"}),
	}
	if err := validateFlagSpecs(flags); err != nil {
		t.Fatalf("期望全部合法，得到 %v", err)
	}
}

func TestValidateArgSpecs(t *testing.T) {
	tests := []struct {
		name string
		args []ArgSpec
	}{
		{"空名", []ArgSpec{{Name: ""}}},
		{"重复定义", []ArgSpec{{Name: "a"}, {Name: "a"}}},
		{"变参不在最后", []ArgSpec{{Name: "a", Variadic: true}, {Name: "b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgSpecs(tt.args)
			assertErrCode(t, err, CodeInvalidArgDef)
		})
	}
}

func TestValidateArgSpecsValid(t *testing.T) {
	args := []ArgSpec{
		{Name: "a", Required: true},
		{Name: "rest", Variadic: true},
	}
	if err := validateArgSpecs(args); err != nil {
		t.Fatalf("期望全部合法，得到 %v", err)
	}
}

func TestParseCommandArgsSuccess(t *testing.T) {
	args := []ArgSpec{
		{Name: "name", Required: true},
		{Name: "rest", Variadic: true},
	}
	flags := []FlagSpec{
		StringFlag("output", "输出路径").Default("out.txt"),
		BoolFlag("verbose", "详细输出"),
		IntFlag("retry", "重试次数").Default(3),
		Int64Flag("size", "大小"),
		FloatFlag("ratio", "比率"),
		DurationFlag("timeout", "超时"),
		EnumFlag("mode", "模式", "fast", "slow").Default("fast"),
		StringSliceFlag("tag", "标签"),
	}
	positional, fv, err := parseCommandArgs(args, flags, []string{
		"alice", "x", "y",
		"--verbose", "--retry=5", "--mode", "slow",
		"--tag", "a", "--tag=b", "--size=8", "--ratio", "0.5",
		"--timeout", "2s", "--output=build.log",
	})
	if err != nil {
		t.Fatalf("期望解析成功，得到 %v", err)
	}
	if len(positional) != 3 || positional[0] != "alice" || positional[1] != "x" || positional[2] != "y" {
		t.Fatalf("位置参数不匹配：%v", positional)
	}
	if fv.values["output"].str != "build.log" {
		t.Fatalf("output 应覆盖默认值：%v", fv.values["output"])
	}
	if !fv.values["verbose"].present || !fv.values["verbose"].b {
		t.Fatal("verbose 应被置为 true")
	}
	if fv.values["retry"].i != 5 {
		t.Fatalf("retry 应为 5：%v", fv.values["retry"])
	}
	if fv.values["size"].i64 != 8 {
		t.Fatalf("size 应为 8：%v", fv.values["size"])
	}
	if fv.values["ratio"].f != 0.5 {
		t.Fatalf("ratio 应为 0.5：%v", fv.values["ratio"])
	}
	if fv.values["timeout"].dur != 2*time.Second {
		t.Fatalf("timeout 应为 2s：%v", fv.values["timeout"])
	}
	if fv.values["mode"].str != "slow" {
		t.Fatalf("mode 应为 slow：%v", fv.values["mode"])
	}
	if got := strings.Join(fv.values["tag"].strs, ","); got != "a,b" {
		t.Fatalf("tag 应为 [a b]：%q", got)
	}
	if fv.values["mode"].present != true || fv.values["retry"].present != true {
		t.Fatal("显式指定的 flag 应标记 present")
	}
}

func TestParseCommandArgsBoolFalse(t *testing.T) {
	_, fv, err := parseCommandArgs(nil, []FlagSpec{BoolFlag("v", "")}, []string{"--v=false"})
	if err != nil {
		t.Fatalf("期望解析成功，得到 %v", err)
	}
	if !fv.values["v"].present || fv.values["v"].b {
		t.Fatalf("v 应为 false 且 present：%v", fv.values["v"])
	}
}

func TestParseCommandArgsTerminator(t *testing.T) {
	positional, fv, err := parseCommandArgs(nil, []FlagSpec{StringFlag("a", "")}, []string{"--", "--a", "x"})
	if err != nil {
		t.Fatalf("期望解析成功，得到 %v", err)
	}
	if len(positional) != 2 || positional[0] != "--a" || positional[1] != "x" {
		t.Fatalf("终止符后的内容应作为位置参数：%v", positional)
	}
	if fv.values["a"].present {
		t.Fatal("终止符后的 --a 不应被解析为 flag")
	}
}

func TestParseCommandArgsErrors(t *testing.T) {
	tests := []struct {
		name  string
		args  []ArgSpec
		flags []FlagSpec
		raw   []string
		code  errx.Code
	}{
		{"未知 flag", nil, []FlagSpec{StringFlag("a", "")}, []string{"--b", "1"}, CodeUnknownFlag},
		{"重复 flag", nil, []FlagSpec{StringFlag("a", "")}, []string{"--a", "1", "--a", "2"}, CodeDuplicateFlag},
		{"缺少值", nil, []FlagSpec{StringFlag("a", "")}, []string{"--a"}, CodeMissingFlagValue},
		{"切片缺少值", nil, []FlagSpec{StringSliceFlag("a", "")}, []string{"--a"}, CodeMissingFlagValue},
		{"布尔值非法", nil, []FlagSpec{BoolFlag("v", "")}, []string{"--v=maybe"}, CodeInvalidFlagValue},
		{"整数非法", nil, []FlagSpec{IntFlag("i", "")}, []string{"--i", "x"}, CodeInvalidFlagValue},
		{"64 位整数非法", nil, []FlagSpec{Int64Flag("i", "")}, []string{"--i", "x"}, CodeInvalidFlagValue},
		{"浮点非法", nil, []FlagSpec{FloatFlag("f", "")}, []string{"--f", "x"}, CodeInvalidFlagValue},
		{"时长非法", nil, []FlagSpec{DurationFlag("d", "")}, []string{"--d", "x"}, CodeInvalidFlagValue},
		{"枚举非法", nil, []FlagSpec{EnumFlag("m", "", "fast")}, []string{"--m", "slow"}, CodeInvalidEnumValue},
		{"缺少必填 flag", nil, []FlagSpec{StringFlag("a", "").Required()}, nil, CodeMissingRequiredFlag},
		{"缺少必填位置参数", []ArgSpec{{Name: "a", Required: true}}, nil, nil, CodeMissingArg},
		{"位置参数过多", []ArgSpec{{Name: "a"}}, nil, []string{"x", "y"}, CodeTooManyArgs},
		{"必填变参缺少", []ArgSpec{{Name: "a", Required: true, Variadic: true}}, nil, nil, CodeMissingArg},
		{"空声明仍严格", []ArgSpec{}, nil, []string{"x"}, CodeTooManyArgs},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseCommandArgs(tt.args, tt.flags, tt.raw)
			assertErrCode(t, err, tt.code)
		})
	}
}

func TestContextFlagAccessors(t *testing.T) {
	flags := []FlagSpec{
		StringFlag("s", "").Default("默认"),
		BoolFlag("b", "").Default(true),
		IntFlag("i", "").Default(int64(7)),
		IntFlag("i2", "").Default(5),
		Int64Flag("i64", "").Default(8),
		Int64Flag("i64b", "").Default(int64(9)),
		FloatFlag("f", "").Default(int64(3)),
		FloatFlag("f2", "").Default(3),
		FloatFlag("f3", "").Default(1.5),
		DurationFlag("d", "").Default(time.Minute),
		EnumFlag("e", "", "fast", "slow").Default("fast"),
		StringSliceFlag("sl", "").Default([]string{"a"}),
	}
	_, fv, err := parseCommandArgs(nil, flags, []string{"--sl", "b", "--s=显式"})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	ctx := &Context{Flags: fv}
	if !ctx.HasFlag("s") {
		t.Fatal("s 应标记为显式指定")
	}
	if ctx.HasFlag("absent") {
		t.Fatal("absent 不应标记为显式指定")
	}
	if ctx.String("s") != "显式" {
		t.Fatalf("String 不匹配：%q", ctx.String("s"))
	}
	if !ctx.Bool("b") {
		t.Fatal("Bool 不匹配")
	}
	if ctx.Int("i") != 7 {
		t.Fatalf("Int 不匹配：%d", ctx.Int("i"))
	}
	if ctx.Int("i2") != 5 {
		t.Fatalf("Int(int) 不匹配：%d", ctx.Int("i2"))
	}
	if ctx.Int64("i64") != 8 {
		t.Fatalf("Int64 不匹配：%d", ctx.Int64("i64"))
	}
	if ctx.Int64("i64b") != 9 {
		t.Fatalf("Int64(int64) 不匹配：%d", ctx.Int64("i64b"))
	}
	if ctx.Float64("f") != 3 {
		t.Fatalf("Float64 不匹配：%v", ctx.Float64("f"))
	}
	if ctx.Float64("f2") != 3 {
		t.Fatalf("Float64(int) 不匹配：%v", ctx.Float64("f2"))
	}
	if ctx.Float64("f3") != 1.5 {
		t.Fatalf("Float64(float64) 不匹配：%v", ctx.Float64("f3"))
	}
	if ctx.Duration("d") != time.Minute {
		t.Fatalf("Duration 不匹配：%v", ctx.Duration("d"))
	}
	if ctx.Enum("e") != "fast" {
		t.Fatalf("Enum 不匹配：%q", ctx.Enum("e"))
	}
	if got := strings.Join(ctx.Strings("sl"), ","); got != "b" {
		t.Fatalf("显式指定的切片应覆盖默认值：%q", got)
	}
	if ctx.HasFlag("undeclared") ||
		ctx.String("undeclared") != "" ||
		ctx.Bool("undeclared") ||
		ctx.Int("undeclared") != 0 ||
		ctx.Int64("undeclared") != 0 ||
		ctx.Float64("undeclared") != 0 ||
		ctx.Duration("undeclared") != 0 ||
		ctx.Strings("undeclared") != nil {
		t.Fatal("未声明 flag 应返回零值")
	}
	if (&Context{}).HasFlag("x") {
		t.Fatal("空 Context 不应命中任何 flag")
	}
}

func TestStringsReturnsCopy(t *testing.T) {
	_, fv, err := parseCommandArgs(nil, []FlagSpec{
		StringSliceFlag("sl", "").Default([]string{"a"}),
	}, []string{"--sl", "b"})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	ctx := &Context{Flags: fv}
	got := ctx.Strings("sl")
	got[0] = "改坏"
	if ctx.Strings("sl")[0] != "b" {
		t.Fatal("修改返回切片不应影响内部存储")
	}
}

func TestValidFlagName(t *testing.T) {
	if validFlagName("") {
		t.Fatal("空名应非法")
	}
	if validFlagName("-x") {
		t.Fatal("短横线开头应非法")
	}
	if !validFlagName("a-1_b") {
		t.Fatal("合法名应通过")
	}
}

func TestParseScalarUnknownKind(t *testing.T) {
	_, err := parseScalar("x", ValueKind(99), "v", nil)
	assertErrCode(t, err, CodeInvalidFlagValue)
}

func TestHelpHelpers(t *testing.T) {
	if got := flagTypeName(ValueKind(99)); got != "value" {
		t.Fatalf("未知类型标签应回退为 value，得到 %q", got)
	}
	if got := padRight("abc", 2); got != "abc" {
		t.Fatalf("超长文本应原样返回，得到 %q", got)
	}
}

func TestCommandHelpTextWithArgsAndFlags(t *testing.T) {
	app := newTestApp(t)
	err := app.AddCommand(&Command{
		Name:        "build",
		Description: "构建目标",
		Args: []ArgSpec{
			{Name: "target", Description: "构建目标", Required: true},
			{Name: "extra", Description: "额外参数", Variadic: true},
		},
		Flags: []FlagSpec{
			StringFlag("out", "输出路径").Default("a.out"),
			BoolFlag("verbose", "详细输出"),
			IntFlag("jobs", "并发数").Required(),
			Int64Flag("size", "大小"),
			FloatFlag("ratio", "比率"),
			DurationFlag("timeout", "超时"),
			EnumFlag("mode", "模式", "fast", "slow").Default("fast"),
			StringSliceFlag("tag", "标签"),
		},
		Action: okAction,
	})
	if err != nil {
		t.Fatalf("注册失败：%v", err)
	}
	text, err := app.CommandHelpText("build")
	if err != nil {
		t.Fatalf("期望成功，得到 %v", err)
	}
	for _, want := range []string{
		"greet build [选项...] [参数...]",
		"参数:",
		"target",
		"（必填）",
		"extra...",
		"（可重复）",
		"选项:",
		"--out string",
		"默认 a.out",
		"--jobs int",
		"--size int64",
		"--ratio float",
		"--timeout duration",
		"--tag string[]",
		"可选值：fast、slow",
		"-h, --help",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("命令帮助缺少 %q：\n%s", want, text)
		}
	}
}

func TestExecuteCommandHelpFlag(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			var out bytes.Buffer
			app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
			app.AddCommand(&Command{
				Name: "build",
				Action: func(ctx context.Context, c *Context) error {
					t.Fatal("请求帮助时不应执行 Action")
					return nil
				},
			})
			if code := app.Execute(context.Background(), []string{"build", flag}); code != ExitOK {
				t.Fatalf("期望退出码 0，得到 %d", code)
			}
			if !strings.Contains(out.String(), "用法:") {
				t.Fatalf("命令帮助缺失：%s", out.String())
			}
		})
	}
}

func TestExecuteCommandHelpTerminator(t *testing.T) {
	var out bytes.Buffer
	app, _ := New("greet", "0.1.0", WithIO(&out, &bytes.Buffer{}))
	app.AddCommand(&Command{
		Name: "echo",
		Action: func(ctx context.Context, c *Context) error {
			for _, a := range c.Args {
				out.WriteString(a + "\n")
			}
			return nil
		},
	})
	if code := app.Execute(context.Background(), []string{"echo", "--", "--help"}); code != ExitOK {
		t.Fatalf("期望退出码 0，得到 %d", code)
	}
	if got := out.String(); got != "--help\n" {
		t.Fatalf("终止符后的 --help 应作为参数：%q", got)
	}
}

func TestExecuteUsageErrorsForParse(t *testing.T) {
	tests := []struct {
		name string
		cmd  *Command
		args []string
		want string
	}{
		{"未知 flag", &Command{
			Name:   "build",
			Flags:  []FlagSpec{StringFlag("out", "")},
			Action: okAction,
		}, []string{"build", "--bad"}, "未知 flag"},
		{"缺少必填 flag", &Command{
			Name:   "build",
			Flags:  []FlagSpec{StringFlag("out", "").Required()},
			Action: okAction,
		}, []string{"build"}, "缺少必填 flag"},
		{"枚举非法", &Command{
			Name:   "build",
			Flags:  []FlagSpec{EnumFlag("mode", "", "fast")},
			Action: okAction,
		}, []string{"build", "--mode", "slow"}, "不在允许列表"},
		{"位置参数过多", &Command{
			Name:   "build",
			Args:   []ArgSpec{{Name: "target"}},
			Action: okAction,
		}, []string{"build", "a", "b"}, "位置参数过多"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errBuf bytes.Buffer
			app, _ := New("greet", "0.1.0", WithIO(&bytes.Buffer{}, &errBuf))
			if err := app.AddCommand(tt.cmd); err != nil {
				t.Fatalf("注册失败：%v", err)
			}
			if code := app.Execute(context.Background(), tt.args); code != ExitUsage {
				t.Fatalf("期望退出码 %d，得到 %d", ExitUsage, code)
			}
			if !strings.Contains(errBuf.String(), tt.want) {
				t.Fatalf("错误信息缺少 %q：%s", tt.want, errBuf.String())
			}
		})
	}
}
