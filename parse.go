package clix

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lcylpzls/errx"
)

// ValueKind 标识 flag 值的类型。
type ValueKind uint8

// flag 值类型枚举。
const (
	KindString ValueKind = iota
	KindBool
	KindInt
	KindInt64
	KindFloat64
	KindDuration
	KindEnum
	KindStringSlice
)

// ArgSpec 描述一个位置参数。
type ArgSpec struct {
	// Name 参数名，用于帮助与错误提示。
	Name string
	// Description 帮助中的参数说明。
	Description string
	// Required 是否必填。
	Required bool
	// Variadic 变参：收集剩余全部位置参数；只能是最后一个参数。
	Variadic bool
}

// FlagSpec 描述一个长 flag（--name）。
type FlagSpec struct {
	// Name flag 名（不含 "--"）。
	Name string
	// Usage 帮助中的说明。
	Usage string
	// Allowed 枚举 flag 的允许值（KindEnum 使用，须非空）。
	Allowed []string

	required   bool
	defaultVal any
	env        string
	kind       ValueKind
}

// StringFlag 构造字符串 flag。
func StringFlag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindString}
}

// BoolFlag 构造布尔 flag。布尔 flag 支持 --name 与 --name=值 两种写法。
func BoolFlag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindBool}
}

// IntFlag 构造整数 flag。
func IntFlag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindInt}
}

// Int64Flag 构造 64 位整数 flag。
func Int64Flag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindInt64}
}

// FloatFlag 构造浮点数 flag。
func FloatFlag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindFloat64}
}

// DurationFlag 构造时长 flag，值使用 time.ParseDuration 语法。
func DurationFlag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindDuration}
}

// EnumFlag 构造枚举 flag；allowed 必须至少提供一个允许值。
func EnumFlag(name, usage string, allowed ...string) FlagSpec {
	return FlagSpec{
		Name:    name,
		Usage:   usage,
		Allowed: append([]string(nil), allowed...),
		kind:    KindEnum,
	}
}

// StringSliceFlag 构造可重复字符串 flag：--name a --name b 收集为切片。
func StringSliceFlag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindStringSlice}
}

// Required 将 flag 标记为必填。
func (f FlagSpec) Required() FlagSpec {
	f.required = true
	return f
}

// Default 设置 flag 默认值；类型必须与 flag 类型匹配。
func (f FlagSpec) Default(v any) FlagSpec {
	f.defaultVal = v
	return f
}

// Env 绑定环境变量：命令行未显式指定时按 环境变量 > 默认值 取值。
// 可重复 flag 的环境变量值使用逗号分隔。
func (f FlagSpec) Env(name string) FlagSpec {
	f.env = name
	return f
}

// FlagValues 保存解析后的 flag 值，通过 Context 的类型化访问器读取。
type FlagValues struct {
	values map[string]flagValue
}

// flagValue 是单一 flag 的内部存储。
type flagValue struct {
	kind    ValueKind
	present bool
	str     string
	b       bool
	i       int
	i64     int64
	f       float64
	dur     time.Duration
	strs    []string
}

// HasFlag 判断 flag 是否在本次调用中被显式指定。
func (c *Context) HasFlag(name string) bool {
	if c.Flags.values == nil {
		return false
	}
	v, ok := c.Flags.values[name]
	return ok && v.present
}

// String 返回字符串 flag 的值；未指定或未声明时返回空字符串。
func (c *Context) String(name string) string {
	v, ok := c.Flags.values[name]
	if !ok {
		return ""
	}
	return v.str
}

// Bool 返回布尔 flag 的值；未指定或未声明时返回 false。
func (c *Context) Bool(name string) bool {
	v, ok := c.Flags.values[name]
	if !ok {
		return false
	}
	return v.b
}

// Int 返回整数 flag 的值；未指定或未声明时返回 0。
func (c *Context) Int(name string) int {
	v, ok := c.Flags.values[name]
	if !ok {
		return 0
	}
	return v.i
}

// Int64 返回 64 位整数 flag 的值；未指定或未声明时返回 0。
func (c *Context) Int64(name string) int64 {
	v, ok := c.Flags.values[name]
	if !ok {
		return 0
	}
	return v.i64
}

// Float64 返回浮点 flag 的值；未指定或未声明时返回 0。
func (c *Context) Float64(name string) float64 {
	v, ok := c.Flags.values[name]
	if !ok {
		return 0
	}
	return v.f
}

// Duration 返回时长 flag 的值；未指定或未声明时返回 0。
func (c *Context) Duration(name string) time.Duration {
	v, ok := c.Flags.values[name]
	if !ok {
		return 0
	}
	return v.dur
}

// Enum 返回枚举 flag 的值；未指定或未声明时返回空字符串。
func (c *Context) Enum(name string) string {
	return c.String(name)
}

// Strings 返回可重复 flag 的切片；未指定或未声明时返回 nil。
func (c *Context) Strings(name string) []string {
	v, ok := c.Flags.values[name]
	if !ok {
		return nil
	}
	return v.strs
}

// validateFlagSpecs 校验 flag 定义集合。
func validateFlagSpecs(flags []FlagSpec) error {
	seen := make(map[string]struct{}, len(flags))
	for _, f := range flags {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			return errx.NewCode(CodeInvalidFlagDef, "flag 名不能为空")
		}
		if !validFlagName(name) {
			return errx.NewCodef(CodeInvalidFlagDef, "非法 flag 名 %q：需以字母开头，仅含字母、数字、下划线与短横线", name)
		}
		if name == "help" {
			return errx.NewCodef(CodeInvalidFlagDef, "flag 名 %q 与内置 --help 冲突", name)
		}
		if _, dup := seen[name]; dup {
			return errx.NewCodef(CodeInvalidFlagDef, "flag %q 重复定义", name)
		}
		seen[name] = struct{}{}
		if f.kind == KindEnum && len(f.Allowed) == 0 {
			return errx.NewCodef(CodeInvalidFlagDef, "枚举 flag %q 必须提供允许值", name)
		}
		if f.env != "" && !validEnvName(f.env) {
			return errx.NewCodef(CodeInvalidFlagDef, "flag %q 的环境变量名 %q 非法", name, f.env)
		}
		if f.defaultVal != nil {
			if err := checkDefaultValue(name, f.kind, f.defaultVal); err != nil {
				return err
			}
			if f.kind == KindEnum && !enumAllowed(f.Allowed, fmt.Sprint(f.defaultVal)) {
				return errx.NewCodef(CodeInvalidFlagDef, "枚举 flag %q 的默认值 %q 不在允许列表", name, f.defaultVal)
			}
		}
	}
	return nil
}

// validEnvName 校验环境变量名：以字母或下划线开头，仅含字母、数字、下划线。
func validEnvName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
		if !ok && i > 0 {
			ok = r >= '0' && r <= '9'
		}
		if !ok {
			return false
		}
	}
	return true
}

// validateArgSpecs 校验位置参数定义集合。
func validateArgSpecs(args []ArgSpec) error {
	seen := make(map[string]struct{}, len(args))
	for i := range args {
		arg := &args[i]
		name := strings.TrimSpace(arg.Name)
		if name == "" {
			return errx.NewCode(CodeInvalidArgDef, "位置参数名不能为空")
		}
		if _, dup := seen[name]; dup {
			return errx.NewCodef(CodeInvalidArgDef, "位置参数 %q 重复定义", name)
		}
		seen[name] = struct{}{}
		if arg.Variadic {
			if i != len(args)-1 {
				return errx.NewCodef(CodeInvalidArgDef, "变参 %q 必须是最后一个位置参数", name)
			}
		}
	}
	return nil
}

// validFlagName 校验 flag 名：以字母开头，仅含字母、数字、下划线与短横线。
func validFlagName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		if !ok && i > 0 {
			ok = r == '_' || r == '-' || (r >= '0' && r <= '9')
		}
		if !ok {
			return false
		}
	}
	return true
}

// checkDefaultValue 校验默认值类型与 flag 类型一致。
func checkDefaultValue(name string, kind ValueKind, v any) error {
	ok := false
	switch kind {
	case KindString, KindEnum:
		_, ok = v.(string)
	case KindBool:
		_, ok = v.(bool)
	case KindInt, KindInt64:
		_, ok = v.(int)
		if !ok {
			_, ok = v.(int64)
		}
	case KindFloat64:
		_, ok = v.(float64)
		if !ok {
			_, ok = v.(int)
			if !ok {
				_, ok = v.(int64)
			}
		}
	case KindDuration:
		_, ok = v.(time.Duration)
	case KindStringSlice:
		_, ok = v.([]string)
	}
	if !ok {
		return errx.NewCodef(CodeInvalidFlagDef, "flag %q 的默认值类型与声明类型不匹配", name)
	}
	return nil
}

// enumAllowed 判断值是否在允许列表中。
func enumAllowed(allowed []string, val string) bool {
	for _, a := range allowed {
		if a == val {
			return true
		}
	}
	return false
}

// parseCommandArgs 解析命令的 flag 与位置参数。
// 返回值：位置参数列表与 flag 值表。
func parseCommandArgs(args []ArgSpec, flags []FlagSpec, raw []string) ([]string, FlagValues, error) {
	flagsByName := make(map[string]*FlagSpec, len(flags))
	for i := range flags {
		flagsByName[flags[i].Name] = &flags[i]
	}
	values := make(map[string]flagValue, len(flags))
	for i := range flags {
		f := &flags[i]
		values[f.Name] = flagValue{kind: f.kind}
		if f.defaultVal != nil {
			v := values[f.Name]
			applyDefault(&v, f)
			values[f.Name] = v
		}
	}

	var positional []string
	parseFlags := true
	for i := 0; i < len(raw); i++ {
		tok := raw[i]
		if parseFlags && tok == "--" {
			parseFlags = false
			continue
		}
		if parseFlags && strings.HasPrefix(tok, "--") && len(tok) > 2 {
			name, val, hasVal := strings.Cut(tok[2:], "=")
			spec, ok := flagsByName[name]
			if !ok {
				return nil, FlagValues{}, errx.NewCodef(CodeUnknownFlag, "未知 flag %q", "--"+name)
			}
			v := values[name]
			switch spec.kind {
			case KindBool:
				b := true
				if hasVal {
					parsed, err := strconv.ParseBool(val)
					if err != nil {
						return nil, FlagValues{}, errx.NewCodef(CodeInvalidFlagValue, "flag %q 需要布尔值，得到 %q", name, val)
					}
					b = parsed
				}
				v = flagValue{kind: KindBool, present: true, b: b}
			case KindStringSlice:
				if !v.present {
					v = flagValue{kind: KindStringSlice, present: true}
				}
				if !hasVal {
					if i+1 >= len(raw) {
						return nil, FlagValues{}, errx.NewCodef(CodeMissingFlagValue, "flag %q 缺少值", name)
					}
					i++
					val = raw[i]
				}
				v.strs = append(v.strs, val)
			default:
				if v.present {
					return nil, FlagValues{}, errx.NewCodef(CodeDuplicateFlag, "flag %q 重复指定", name)
				}
				if !hasVal {
					if i+1 >= len(raw) {
						return nil, FlagValues{}, errx.NewCodef(CodeMissingFlagValue, "flag %q 缺少值", name)
					}
					i++
					val = raw[i]
				}
				parsed, err := parseScalar(name, spec.kind, val, spec.Allowed)
				if err != nil {
					return nil, FlagValues{}, err
				}
				parsed.present = true
				v = parsed
			}
			values[name] = v
			continue
		}
		positional = append(positional, tok)
	}

	if err := applyEnvValues(flags, values); err != nil {
		return nil, FlagValues{}, err
	}
	for _, f := range flags {
		if f.required && !values[f.Name].present {
			return nil, FlagValues{}, errx.NewCodef(CodeMissingRequiredFlag, "缺少必填 flag %q", f.Name)
		}
	}
	// Args 为 nil 表示位置参数透传（不限数量）；空切片表示严格零参数。
	if args != nil {
		if err := checkPositionalCount(args, positional); err != nil {
			return nil, FlagValues{}, err
		}
	}
	return positional, FlagValues{values: values}, nil
}

// applyEnvValues 为未显式指定的 flag 应用环境变量值（环境变量 > 默认值）。
func applyEnvValues(flags []FlagSpec, values map[string]flagValue) error {
	for i := range flags {
		f := &flags[i]
		if f.env == "" || values[f.Name].present {
			continue
		}
		raw, ok := os.LookupEnv(f.env)
		if !ok {
			continue
		}
		v := values[f.Name]
		switch f.kind {
		case KindBool:
			b, err := strconv.ParseBool(raw)
			if err != nil {
				return errx.NewCodef(CodeInvalidFlagValue,
					"环境变量 %s 需要布尔值，得到 %q", f.env, raw)
			}
			v = flagValue{kind: KindBool, present: true, b: b}
		case KindStringSlice:
			v = flagValue{kind: KindStringSlice, present: true}
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					v.strs = append(v.strs, part)
				}
			}
		default:
			parsed, err := parseScalar(f.Name, f.kind, raw, f.Allowed)
			if err != nil {
				return err
			}
			parsed.present = true
			v = parsed
		}
		values[f.Name] = v
	}
	return nil
}

// applyDefault 将默认值写入 flag 存储。
func applyDefault(v *flagValue, f *FlagSpec) {
	switch f.kind {
	case KindString, KindEnum:
		v.str = f.defaultVal.(string)
	case KindBool:
		v.b = f.defaultVal.(bool)
	case KindInt:
		switch dv := f.defaultVal.(type) {
		case int:
			v.i = dv
		case int64:
			v.i = int(dv)
		}
	case KindInt64:
		switch dv := f.defaultVal.(type) {
		case int:
			v.i64 = int64(dv)
		case int64:
			v.i64 = dv
		}
	case KindFloat64:
		switch dv := f.defaultVal.(type) {
		case float64:
			v.f = dv
		case int:
			v.f = float64(dv)
		case int64:
			v.f = float64(dv)
		}
	case KindDuration:
		v.dur = f.defaultVal.(time.Duration)
	case KindStringSlice:
		v.strs = append([]string(nil), f.defaultVal.([]string)...)
	}
}

// parseScalar 解析非布尔、非切片的标量 flag 值。
func parseScalar(name string, kind ValueKind, val string, allowed []string) (flagValue, error) {
	var out flagValue
	switch kind {
	case KindString:
		out.kind, out.str = KindString, val
	case KindInt:
		n, err := strconv.Atoi(val)
		if err != nil {
			return out, errx.NewCodef(CodeInvalidFlagValue, "flag %q 需要整数，得到 %q", name, val)
		}
		out.kind, out.i = KindInt, n
	case KindInt64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return out, errx.NewCodef(CodeInvalidFlagValue, "flag %q 需要 64 位整数，得到 %q", name, val)
		}
		out.kind, out.i64 = KindInt64, n
	case KindFloat64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return out, errx.NewCodef(CodeInvalidFlagValue, "flag %q 需要浮点数，得到 %q", name, val)
		}
		out.kind, out.f = KindFloat64, n
	case KindDuration:
		d, err := time.ParseDuration(val)
		if err != nil {
			return out, errx.NewCodef(CodeInvalidFlagValue, "flag %q 需要时长，得到 %q", name, val)
		}
		out.kind, out.dur = KindDuration, d
	case KindEnum:
		if !enumAllowed(allowed, val) {
			return out, errx.NewCodef(CodeInvalidEnumValue,
				"flag %q 的值 %q 不在允许列表 %v 中", name, val, allowed)
		}
		out.kind, out.str = KindEnum, val
	default:
		return out, errx.NewCodef(CodeInvalidFlagValue, "flag %q 类型不支持", name)
	}
	return out, nil
}

// checkPositionalCount 校验位置参数数量。
func checkPositionalCount(args []ArgSpec, positional []string) error {
	requiredFixed := 0
	variadicIndex := -1
	for i := range args {
		arg := &args[i]
		if arg.Variadic {
			variadicIndex = i
			break
		}
		if arg.Required {
			requiredFixed++
		}
	}
	minRequired := requiredFixed
	if variadicIndex >= 0 && args[variadicIndex].Required {
		minRequired++
	}
	if len(positional) < minRequired {
		return errx.NewCodef(CodeMissingArg,
			"缺少必填位置参数：当前 %d 个，至少需要 %d 个", len(positional), minRequired)
	}
	if variadicIndex < 0 && len(positional) > len(args) {
		return errx.NewCodef(CodeTooManyArgs,
			"位置参数过多：当前 %d 个，最多 %d 个", len(positional), len(args))
	}
	return nil
}
