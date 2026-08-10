package clix

import (
	testx "github.com/lcylpzls/testx"
	"strings"
	"testing"
	"time"
)

func TestFlagEnvPrecedence(t *testing.T) {
	t.Setenv("CLIX_TEST_OUTPUT", "env值")
	t.Setenv("CLIX_TEST_RETRY", "9")
	t.Setenv("CLIX_TEST_VERBOSE", "true")
	t.Setenv("CLIX_TEST_MODE", "slow")
	t.Setenv("CLIX_TEST_TAGS", " a , b , c ")
	t.Setenv("CLIX_TEST_TIMEOUT", "3s")

	flags := []FlagSpec{
		StringFlag("output", "").Env("CLIX_TEST_OUTPUT").Default("默认"),
		IntFlag("retry", "").Env("CLIX_TEST_RETRY"),
		BoolFlag("verbose", "").Env("CLIX_TEST_VERBOSE"),
		EnumFlag("mode", "", "fast", "slow").Env("CLIX_TEST_MODE"),
		StringSliceFlag("tags", "").Env("CLIX_TEST_TAGS"),
		DurationFlag("timeout", "").Env("CLIX_TEST_TIMEOUT"),
	}
	_, fv, err := parseCommandArgs(nil, flags, []string{"--output", "命令行"})
	testx.RequireNoError(t, err)

	ctx := &Context{Flags: fv}
	if ctx.String("output") != "命令行" {
		t.Fatalf("命令行应覆盖环境变量：%q", ctx.String("output"))
	}
	if ctx.Int("retry") != 9 {
		t.Fatalf("环境变量应提供 retry：%d", ctx.Int("retry"))
	}
	if !ctx.Bool("verbose") {
		t.Fatal("环境变量应提供 verbose=true")
	}
	if ctx.Enum("mode") != "slow" {
		t.Fatalf("环境变量应提供 mode=slow：%q", ctx.Enum("mode"))
	}
	if got := strings.Join(ctx.Strings("tags"), ","); got != "a,b,c" {
		t.Fatalf("环境变量切片应逗号分隔并去空格：%q", got)
	}
	if ctx.Duration("timeout") != 3*time.Second {
		t.Fatalf("环境变量应提供 timeout=3s：%v", ctx.Duration("timeout"))
	}
}

func TestFlagEnvSatisfiesRequired(t *testing.T) {
	t.Setenv("CLIX_TEST_REQUIRED", "ok")
	_, fv, err := parseCommandArgs(nil, []FlagSpec{
		StringFlag("required", "").Required().Env("CLIX_TEST_REQUIRED"),
	}, nil)
	testx.RequireNoError(t, err)

	ctx := &Context{Flags: fv}
	if !ctx.HasFlag("required") {
		t.Fatal("环境变量提供的值应视为已设置")
	}
}

func TestFlagEnvErrors(t *testing.T) {
	t.Setenv("CLIX_TEST_BAD_BOOL", "maybe")
	_, _, err := parseCommandArgs(nil, []FlagSpec{BoolFlag("b", "").Env("CLIX_TEST_BAD_BOOL")}, nil)
	assertErrCode(t, err, CodeInvalidFlagValue)

	t.Setenv("CLIX_TEST_BAD_INT", "x")
	_, _, err = parseCommandArgs(nil, []FlagSpec{IntFlag("i", "").Env("CLIX_TEST_BAD_INT")}, nil)
	assertErrCode(t, err, CodeInvalidFlagValue)

	t.Setenv("CLIX_TEST_BAD_MODE", "nope")
	_, _, err = parseCommandArgs(nil, []FlagSpec{EnumFlag("m", "", "fast").Env("CLIX_TEST_BAD_MODE")}, nil)
	assertErrCode(t, err, CodeInvalidEnumValue)
}

func TestFlagEnvMissing(t *testing.T) {
	_, fv, err := parseCommandArgs(nil, []FlagSpec{
		StringFlag("output", "").Env("CLIX_TEST_NOT_SET").Default("默认值"),
	}, nil)
	testx.RequireNoError(t, err)

	ctx := &Context{Flags: fv}
	if ctx.String("output") != "默认值" {
		t.Fatalf("环境变量未设置时应使用默认值：%q", ctx.String("output"))
	}
	if ctx.HasFlag("output") {
		t.Fatal("环境变量未设置时不应视为已设置")
	}
}

func TestValidEnvName(t *testing.T) {
	for _, name := range []string{"APP_X", "_X", "a1"} {
		if !validEnvName(name) {
			t.Fatalf("%q 应为合法环境变量名", name)
		}
	}
	for _, name := range []string{"", "1X", "X-Y"} {
		if validEnvName(name) {
			t.Fatalf("%q 应为非法环境变量名", name)
		}
	}
}

func TestValidateFlagSpecsInvalidEnv(t *testing.T) {
	err := validateFlagSpecs([]FlagSpec{StringFlag("a", "").Env("X-Y")})
	assertErrCode(t, err, CodeInvalidFlagDef)
}
