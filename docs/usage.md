# 使用文档

## 管线总览

```
输入文件 ─→ cmd/extractor ─→ cmd/assessor ─→ cmd/loader ─→ cmd/reasoner ─→ cmd/generator ─→ 产物
             提取结构化      提取三元组        入库            推理            生成代码/文档
             文档对象        (需 LLM)         建图谱
```

每一步产出文件，下一步消费。可独立运行，也可串联。

## cmd/extractor — 数据提取

从 Markdown 中提取 frontmatter 和正文，输出 JSON。

```bash
# 提取单个文件
go run ./cmd/extractor -input ./doc.md -output ./parsed

# 提取整个目录
go run ./cmd/extractor -input ./docs -output ./parsed
```

支持 `.md` 和 `.txt`，自动递归子目录。

## cmd/assessor — 智能标记

对文档正文做认知密度评估，提取语义三元组。

```bash
export OPENAI_API_KEY="sk-..."
go run ./cmd/assessor -input ./parsed -output ./triples.jsonl
```

输出格式：每行一个 JSON 三元组（subject, predicate, object, confidence, source, context）。

## cmd/loader — 入库

按置信度阈值分类，高置信度写入 JSON 图谱存储。

```bash
go run ./cmd/loader -input ./triples.jsonl -threshold 0.7
```

阈值策略：≥0.7 自动入库，0.4–0.7 暂存待确认（pending.jsonl），<0.4 不入库。

## cmd/reasoner — 推理

加载 Datalog 规则（.mgl 文件），对图谱事实执行推理查询。

```bash
go run ./cmd/reasoner -rule ./rules.mgl -query "contains_tc(?X, ?Y)"
```

规则样例（保存为 rules.mgl）：
```prolog
contains_tc(A, C) :- contains(A, B), contains_tc(B, C)
contains_tc(A, B) :- contains(A, B)
```

## cmd/generator — 生成

基于推理结果生成三种产物：

```bash
# Go 代码
go run ./cmd/generator -target code -query "contains(?X, ?Y)" -output ./types.go

# API 设计
go run ./cmd/generator -target api -query "contains(?X, ?Y)" -output ./api.md

# FAQ
go run ./cmd/generator -target faq -query "contains(?X, ?Y)" -output ./faq.md
```

不足 3 条事实时拒绝生成。
输出路径无写入权限时降级到 stdout。
Go 代码生成后自动 `go fmt`。

## cmd/federate — 多源融合

联合多个数据源，融合为统一三元组。

```bash
go run ./cmd/federate \
  -sources "docs=file:./docs,api=api:https://api.example.com/data" \
  -output ./fused.json
```

## 图查询

JSON 图谱存储支持图遍历查询（Go 代码中直接调用）：

```go
store := storage.NewJSONStore("./data")

// 邻接点查询
neighbors, _ := store.QueryNeighbors("订单")

// 路径搜索（DFS）
paths, _ := store.QueryPath("订单", "库存", 5)
```
