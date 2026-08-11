package core

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/validx"
)

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
	if err := validateFlagValues(flags, values); err != nil {
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

// validateFlagValues 对声明了 validx 规则的 flag 值执行校验。
func validateFlagValues(flags []FlagSpec, values map[string]flagValue) error {
	validator, _ := validx.New()
	for i := range flags {
		f := &flags[i]
		if f.validate == "" {
			continue
		}
		for _, val := range valuesForValidation(values[f.Name]) {
			if err := validator.ValidateField(val, f.validate); err != nil {
				return errx.WrapCodef(err, CodeFlagValidationFailed, "flag %q 校验失败", f.Name)
			}
		}
	}
	return nil
}

// valuesForValidation 将 flag 存储转换为可校验的值列表（切片逐元素校验）。
func valuesForValidation(v flagValue) []any {
	if v.kind == KindStringSlice {
		out := make([]any, 0, len(v.strs))
		for _, s := range v.strs {
			out = append(out, s)
		}
		return out
	}
	switch v.kind {
	case KindBool:
		return []any{v.b}
	case KindInt:
		return []any{v.i}
	case KindInt64:
		return []any{v.i64}
	case KindFloat64:
		return []any{v.f}
	case KindDuration:
		return []any{v.dur}
	default:
		return []any{v.str}
	}
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
