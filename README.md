# quanttide-example-of-knowledge-engineering

量潮知识工程通用示例项目 — 从模糊的原始文档到可执行的知识产物。

## 项目结构

```
cmd/
  extractor/        # 数据层：Markdown 文件解析与结构化提取
  assessor/         # 信息层：认知密度评估与三元组提取
  loader/           # 知识层：三元组入库（JSON 图谱存储）
  reasoner/         # 知识层：Datalog 声明式推理
  generator/        # 智慧层：生成 Go 代码 / API 设计 / FAQ
  federate/         # 多源联合：多数据源融合
internal/
  document/         # 文档模型
  markdown/         # Markdown + YAML 解析器
  triple/           # 三元组模型
  chunker/          # 文本分块
  llm/              # LLM 客户端封装
  storage/          # 图谱存储引擎 + 图查询
  rule/             # Datalog 规则引擎
  generator/        # 多目标生成器
  source/           # 多数据源接口
rules/              # Datalog 规则文件
docs/               # 设计文档 + 示例 + 实验场景
```

## 管线

```text
Data             →  Information        →  Knowledge         →  Wisdom
cmd/extractor       cmd/assessor          cmd/loader           cmd/generator
  + docs/*.md         + internal/llm        cmd/reasoner         + internal/generator
                      + internal/triple     + internal/storage
                      + internal/chunker    + internal/rule
                                            + rules/*.mgl
```

## 快速开始

```bash
# 1. 数据提取
go run ./cmd/extractor -input ./docs/scenario/products -output ./parsed

# 2. 入库
go run ./cmd/loader -input ./triples.jsonl -threshold 0.7

# 3. 推理
go run ./cmd/reasoner -rule ./rules/scenario -query "subclass_tc(?X, ?Y)"

# 4. 生成
go run ./cmd/generator -target faq -query "belongs_to(?X, ?Y)" -output ./generated
```

## 许可

Apache License 2.0
