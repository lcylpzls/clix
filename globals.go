package clix

import (
	"strconv"
	"strings"
	"time"

	"github.com/lcylpzls/errx"
)

// stripGlobalFlags 从参数头部剥离已声明的全局 flag，返回解析值与剩余参数。
// 全局 flag 只识别命令名之前连续出现的已知 flag。
func (a *App) stripGlobalFlags(args []string) (FlagValues, []string, error) {
	if len(a.globalFlags) == 0 {
		return FlagValues{}, args, nil
	}
	byName := make(map[string]*FlagSpec, len(a.globalFlags))
	values := make(map[string]flagValue, len(a.globalFlags))
	for i := range a.globalFlags {
		f := &a.globalFlags[i]
		byName[f.Name] = f
		values[f.Name] = flagValue{kind: f.kind}
		if f.defaultVal != nil {
			v := values[f.Name]
			applyDefault(&v, f)
			values[f.Name] = v
		}
	}
	i := 0
	for i < len(args) {
		tok := args[i]
		if tok == "--" || !strings.HasPrefix(tok, "--") || len(tok) <= 2 {
			break
		}
		name, val, hasVal := strings.Cut(tok[2:], "=")
		spec, ok := byName[name]
		if !ok {
			break
		}
		v := values[name]
		switch spec.kind {
		case KindBool:
			b := true
			if hasVal {
				parsed, err := strconv.ParseBool(val)
				if err != nil {
					return FlagValues{}, nil, errx.NewCodef(CodeInvalidFlagValue,
						"全局 flag %q 需要布尔值，得到 %q", name, val)
				}
				b = parsed
			}
			v = flagValue{kind: KindBool, present: true, b: b}
		case KindStringSlice:
			if !v.present {
				v = flagValue{kind: KindStringSlice, present: true}
			}
			if !hasVal {
				if i+1 >= len(args) {
					return FlagValues{}, nil, errx.NewCodef(CodeMissingFlagValue, "全局 flag %q 缺少值", name)
				}
				i++
				val = args[i]
			}
			v.strs = append(v.strs, val)
		default:
			if v.present {
				return FlagValues{}, nil, errx.NewCodef(CodeDuplicateFlag, "全局 flag %q 重复指定", name)
			}
			if !hasVal {
				if i+1 >= len(args) {
					return FlagValues{}, nil, errx.NewCodef(CodeMissingFlagValue, "全局 flag %q 缺少值", name)
				}
				i++
				val = args[i]
			}
			parsed, err := parseScalar(name, spec.kind, val, spec.Allowed)
			if err != nil {
				return FlagValues{}, nil, err
			}
			parsed.present = true
			v = parsed
		}
		values[name] = v
		i++
	}
	if err := applyEnvValues(a.globalFlags, values); err != nil {
		return FlagValues{}, nil, err
	}
	if err := validateFlagValues(a.globalFlags, values); err != nil {
		return FlagValues{}, nil, err
	}
	for _, f := range a.globalFlags {
		if f.required && !values[f.Name].present {
			return FlagValues{}, nil, errx.NewCodef(CodeMissingRequiredFlag, "缺少必填全局 flag %q", f.Name)
		}
	}
	return FlagValues{values: values}, args[i:], nil
}

// globalValue 返回全局 flag 的内部存储；未声明时返回 ok=false。
func (c *Context) globalValue(name string) (flagValue, bool) {
	if c.Global.values == nil {
		return flagValue{}, false
	}
	v, ok := c.Global.values[name]
	return v, ok
}

// HasGlobalFlag 判断全局 flag 是否在本次调用中被显式指定。
func (c *Context) HasGlobalFlag(name string) bool {
	v, ok := c.globalValue(name)
	return ok && v.present
}

// GlobalString 返回全局字符串 flag 的值；未指定或未声明时返回空字符串。
func (c *Context) GlobalString(name string) string {
	v, ok := c.globalValue(name)
	if !ok {
		return ""
	}
	return v.str
}

// GlobalBool 返回全局布尔 flag 的值；未指定或未声明时返回 false。
func (c *Context) GlobalBool(name string) bool {
	v, ok := c.globalValue(name)
	if !ok {
		return false
	}
	return v.b
}

// GlobalInt 返回全局整数 flag 的值；未指定或未声明时返回 0。
func (c *Context) GlobalInt(name string) int {
	v, ok := c.globalValue(name)
	if !ok {
		return 0
	}
	return v.i
}

// GlobalInt64 返回全局 64 位整数 flag 的值；未指定或未声明时返回 0。
func (c *Context) GlobalInt64(name string) int64 {
	v, ok := c.globalValue(name)
	if !ok {
		return 0
	}
	return v.i64
}

// GlobalFloat64 返回全局浮点 flag 的值；未指定或未声明时返回 0。
func (c *Context) GlobalFloat64(name string) float64 {
	v, ok := c.globalValue(name)
	if !ok {
		return 0
	}
	return v.f
}

// GlobalDuration 返回全局时长 flag 的值；未指定或未声明时返回 0。
func (c *Context) GlobalDuration(name string) time.Duration {
	v, ok := c.globalValue(name)
	if !ok {
		return 0
	}
	return v.dur
}

// GlobalEnum 返回全局枚举 flag 的值；未指定或未声明时返回空字符串。
func (c *Context) GlobalEnum(name string) string {
	return c.GlobalString(name)
}

// GlobalStrings 返回全局可重复 flag 的切片；未指定或未声明时返回 nil。
func (c *Context) GlobalStrings(name string) []string {
	v, ok := c.globalValue(name)
	if !ok {
		return nil
	}
	return v.strs
}
