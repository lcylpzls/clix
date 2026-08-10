# 1.0 候选终审

> 本清单用于确认 clix 达到 1.0 候选标准；**v1.0.0 是否发布由维护者决定**。

## 1. API 冻结

- [x] 公开 API 签名稳定：App/Command/Context/FlagSpec/ArgSpec 语义明确；
- [x] 帮助文本、错误文本与退出码约定形成稳定快照；
- [x] 错误码全集与 `docs/errors.md` 一致；
- [x] pre-1.0 兼容承诺：v0.1.0 起的核心行为（分发/解析/帮助/退出码）
  无意外破坏。

## 2. 质量门禁

- [x] 根包语句覆盖率 100%；
- [x] 测试乱序（`-shuffle=on`）、race 全平台通过；
- [x] vet / staticcheck / govulncheck 通过；
- [x] fuzz 目标短跑 5s 通过；
- [x] 示例模块（含 confx/validx/全局 flag/子命令）全绿。

## 3. 设计确认

- [x] 无全局可变状态；App 实例即全部状态；
- [x] 输出流与日志器可注入，帮助文本确定性生成；
- [x] 命令注册与查找并发安全；
- [x] Action/钩子 panic 恢复为结构化错误；
- [x] 家族一致性：errx 错误码、logx 日志、confx 配置、validx 校验。

## 4. 性能

- [x] `BENCHMARKS.md` 记录解析与分发基准；
- [x] 解析热路径无意外分配（`-benchmem` 复核）。

## 5. 文档与安全

- [x] README / docs/api.md / docs/errors.md / docs/roadmap.md 一致；
- [x] SECURITY.md / CONTRIBUTING.md / CODEOWNERS / Issue 模板齐全；
- [x] 发布流程：tag 触发 Release，CI 全绿后发布。

## 结论

clix 已通过 1.0 候选终审清单，并经维护者确认于 **v1.0.0** 正式发布。
自 v1.0.0 起 API 冻结，破坏性变更必须提升主版本。
