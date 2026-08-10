# 架构

## 1. 包内模块

```text
clix（根包）
├── app.go       App 模型、构造、选项、命令注册
├── command.go   Command / Context / ActionFunc
├── errors.go    错误码定义与注册
├── help.go      帮助文本渲染
└── execute.go   分发、执行、退出码与日志
```

依赖方向：

```text
execute.go ──→ help.go ──→ app.go/command.go
      │
      └─────→ errors.go（被全部模块引用）
```

## 2. 关键设计

- **状态集中**：所有可变状态在 App 内；命令表由 `RWMutex` 保护，
  帮助/查找与注册并发安全；
- **动作隔离**：Action 的 panic 在 `invoke` 内恢复为结构化错误，
  不向调用方泄漏崩溃栈；
- **输出抽象**：所有面向用户的文本经 Out/Err 写出，字符串渲染与
  写出分离（`HelpText` / `CommandHelpText` 返回字符串，便于快照测试）；
- **日志可选**：logger 为 nil 时全部日志路径为零开销短路。

## 3. 后续演进扩展点

- `parse` 层（v0.2.0）：flag 声明与解析，独立文件，不侵入 App 分发；
- 命令树（v0.3.0）：parent/children 指针与分组索引；
- 钩子（v0.4.0）：在 `invoke` 前后插入生命周期点；
- validx 联动（v0.5.0）：Action 入口统一校验，错误归一为 errx 聚合。

