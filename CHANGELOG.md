# 更新日志

## [v0.1.0] - 2026-08-10

### 新增

- App/Command 骨架：`New` / `AddCommand` / 根 Action / Context；
- 内置命令约定：`-h` / `--help` / `help [命令]` / `--version`；
- errx 结构化错误码（`CLI_INVALID_APP` / `CLI_MISSING_COMMAND` /
  `CLI_UNKNOWN_COMMAND` / `CLI_CANCELLED` / `CLI_ACTION_PANIC`）
  与退出码映射（0/1/2/130）；
- logx 可选注入：命令开始（Debug）、成功（Info）、用法错误（Warn）、
  失败/取消/恐慌（Error）结构化日志；
- Out/Err 可注入；帮助文本返回字符串便于快照测试；
- Action panic 恢复为结构化错误，不向调用方泄漏崩溃栈；
- fuzz 目标（`FuzzExecute` / `FuzzCommandHelpText`）接入 CI；
- 三平台 CI + Linux 多发行版容器矩阵 + 示例模块 + Release 工作流。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck / fuzz / govulncheck 全绿。

