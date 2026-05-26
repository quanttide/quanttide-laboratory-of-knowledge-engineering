# 实验场景：智能客服知识库构建

> 从产品文档到智能问答，完整走通 Data → Information → Knowledge → Wisdom 四层管线。

## 场景概述

**公司**: TechMart — 在线电子产品零售商

**目标**: 将内部产品文档自动转化为客服知识库（FAQ + API + 数据模型）

**原始材料**: 4 份 Markdown 产品文档（`_examples/scenario/products/`）

## 管线执行

```bash
# 1. 数据提取
go run ./cmd/extractor -input _examples/scenario/products -output docs/scenario/parsed

# 2. 智能标记（需 OPENAI_API_KEY）
go run ./cmd/assessor -input docs/scenario/parsed -output docs/scenario/triples.jsonl

# 3. 入库
go run ./cmd/loader -input docs/scenario/triples.jsonl -threshold 0.7

# 4. 推理
go run ./cmd/reasoner -rule rules/scenario -query "subclass_tc(?X, ?Y)"

# 5. 生成
go run ./cmd/generator -target faq -query "belongs_to(?X, ?Y)" -output docs/scenario/generated
```

一键执行：`bash scripts/run-scenario.sh`

## 预期输出

| 步骤 | 输出 | 说明 |
|------|------|------|
| extractor | `parsed/*.json` | 结构化文档对象 |
| assessor | `triples.jsonl` | 带置信度的三元组 |
| loader | `store.jsonl` | 高置信度三元组图谱 |
| reasoner | stdout | 推理结果（如 subclass_tc） |
| generator | `generated/` | FAQ Markdown / API 设计 / Go 类型 |

## 价值

传统方式：客服手动整理 FAQ（3 天）→ 本管线自动生成（5 分钟）
