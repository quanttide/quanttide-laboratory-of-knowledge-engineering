# AGENTS

## 项目概览

量潮知识工程通用示例项目。Go 实现的 DIKW 管线：Data → Information → Knowledge → Wisdom。

## 目录结构

```
cmd/extractor  数据层  — 从 Markdown 提取结构化文档
cmd/assessor   信息层  — 认知密度评估 → 三元组
cmd/loader    知识层  — 三元组入库（JSON 图谱存储）
cmd/reasoner  知识层  — Datalog 推理
cmd/generator 智慧层  — 生成 Go 代码 / API / FAQ
cmd/federate  多源融合 — 多数据源联合与融合
internal/      内部包：document, markdown, triple, chunker, llm, storage, rule, generator, source
rules/         Datalog 规则文件 (.mgl)
testdata/      测试夹具 + 实验场景数据
```

## 关键约定

- 零外部依赖：不使用 CozoDB/Neo4j/golang.org/x/text/yaml，全部内联实现
- 每个 cmd 有独立的 `flag.FlagSet` + `run()` 模式，便于测试
- 测试分三层：Example（可执行文档）、单元测试、集成测试（cmd/）
- 模块路径：`github.com/quanttide/quanttide-example-of-knowledge-engineering`

## 子模组

本项目为子模组，路径 `examples/default`。提交后回到主仓库更新引用。
