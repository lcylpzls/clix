# 基准测试

> 采集环境：Windows / AMD Ryzen 5 7600 / Go 1.26.5
> 采集日期：2026-08-10
> 命令：`go test -bench=. -benchmem -run '^$' .`

## BenchmarkParseCommandArgs

六种 flag 类型（string/bool/int/duration/enum/string[]）+ 必填与变参
位置参数：

| 指标 | 数值 |
| --- | --- |
| 耗时 | 1027 ns/op |
| 内存 | 1184 B/op |
| 分配 | 7 allocs/op |

## BenchmarkExecute

全局 flag + 命令解析 + validx 校验 + Action 调用的完整分发：

| 指标 | 数值 |
| --- | --- |
| 耗时 | 1012 ns/op |
| 内存 | 2632 B/op |
| 分配 | 16 allocs/op |

## 说明

- 基准仅反映本机相对量级；CI 不设硬性性能门槛，防止环境抖动误报；
- 解析热路径核心为 map 查表与 `strconv`，常规 CLI 启动开销可忽略。
