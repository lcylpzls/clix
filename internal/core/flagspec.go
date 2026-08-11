package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/validx"
)

// ValueKind 标识 flag 值的类型。
type ValueKind uint8

// flag 值类型枚举。
const (
	KindString ValueKind = iota
	KindBool
	KindInt
	KindInt64
	KindUint64
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
	validate   string
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

// Uint64Flag 构造 64 位无符号整数 flag。
func Uint64Flag(name, usage string) FlagSpec {
	return FlagSpec{Name: name, Usage: usage, kind: KindUint64}
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

// Validate 绑定 validx 规则串（如 "required,min=3" / "oneof=dev prod"）。
// 值未通过校验时返回 CLI_FLAG_VALIDATION_FAILED（用法错误，退出码 2）。
func (f FlagSpec) Validate(rules string) FlagSpec {
	f.validate = strings.TrimSpace(rules)
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
	u64     uint64
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

// Uint64 返回 64 位无符号整数 flag 的值；未指定或未声明时返回 0。
func (c *Context) Uint64(name string) uint64 {
	v, ok := c.Flags.values[name]
	if !ok {
		return 0
	}
	return v.u64
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
	return append([]string(nil), v.strs...)
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
		if f.validate != "" {
			if err := checkValidationRules(f.kind, f.validate); err != nil {
				return err
			}
		}
		if f.defaultVal != nil {
			if err := checkDefaultValue(name, f.kind, f.defaultVal); err != nil {
				return err
			}
			if f.kind == KindEnum && !enumAllowed(f.Allowed, fmt.Sprint(f.defaultVal)) {
				return errx.NewCodef(CodeInvalidFlagDef, "枚举 flag %q 的默认值 %q 不在允许列表", name, f.defaultVal)
			}
			if f.validate != "" {
				if err := validateDefaultAgainstRules(f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validateDefaultAgainstRules 在注册期校验默认值是否满足 validx 规则。
func validateDefaultAgainstRules(f FlagSpec) error {
	if err := validx.ValidateField(f.defaultVal, f.validate); err != nil {
		return errx.WrapCodef(err, CodeInvalidFlagDef,
			"flag %q 的默认值未通过校验规则", f.Name)
	}
	return nil
}

// checkValidationRules 在注册期预编译 validx 规则串，语法非法时返回错误。
func checkValidationRules(kind ValueKind, rules string) error {
	// 注册期预编译仅校验规则语法。
	if err := validx.ValidateField(zeroValueForKind(kind), rules); err != nil {
		if errx.Is(err, validx.CodeInvalidRule) {
			return errx.WrapCode(err, CodeInvalidFlagDef, "flag 校验规则非法")
		}
		// 规则合法但零值未通过：注册期忽略，解析期按实际值校验。
		return nil
	}
	return nil
}

// zeroValueForKind 返回各 flag 类型的零值（用于规则预编译）。
func zeroValueForKind(kind ValueKind) any {
	switch kind {
	case KindBool:
		return false
	case KindInt:
		return 0
	case KindInt64:
		return int64(0)
	case KindUint64:
		return uint64(0)
	case KindFloat64:
		return float64(0)
	case KindDuration:
		return time.Duration(0)
	case KindStringSlice:
		return []string{}
	default:
		return ""
	}
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
	case KindUint64:
		switch dv := v.(type) {
		case uint64, uint:
			ok = true
		case int:
			ok = dv >= 0
		case int64:
			ok = dv >= 0
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
