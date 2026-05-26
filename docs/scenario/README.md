# 实验场景：智能客服知识库构建

> 从产品文档到智能问答，完整走通 Data → Information → Knowledge → Wisdom 四层管线。

## 场景概述

**公司**: TechMart — 在线电子产品零售商

**目标**: 将内部产品文档自动转化为客服知识库（FAQ + API + 数据模型）

**原始材料**: 4 份 Markdown 产品文档（含 YAML frontmatter）

```
docs/scenario/
  products/
    laptop.md       # 笔记本电脑
    tablet.md       # 平板电脑
    return-policy.md  # 退换货政策
    shipping.md       # 配送说明
```

---

## 管线执行

### 第 1 步：数据提取

```bash
go run ./cmd/extractor -input docs/scenario/products -output docs/scenario/parsed
```

从 Markdown 中提取 frontmatter 元信息 + 正文 + 块序列，输出 JSON。

### 第 2 步：智能标记

```bash
export OPENAI_API_KEY="sk-..."
go run ./cmd/assessor -input docs/scenario/parsed -output docs/scenario/triples.jsonl
```

对正文进行认知密度评估，提取带置信度的三元组，如：

```json
{"subject":"笔记本电脑","predicate":"属于","object":"电子产品","confidence":0.95}
{"subject":"退换货","predicate":"期限","object":"30天","confidence":0.90}
```

### 第 3 步：入库

```bash
go run ./cmd/loader -input docs/scenario/triples.jsonl -threshold 0.7
```

高置信度（≥0.7）自动入库，中置信度暂存待确认。

### 第 4 步：推理

```bash
go run ./cmd/reasoner -rule rules/scenario.mgl -query "subclass_tc(?X, ?Y)"
```

推理出隐含关系：
```
subclass(笔记本电脑, 电子产品) ∧ subclass(电子产品, 商品) → subclass_tc(笔记本电脑, 商品)
```

### 第 5 步：生成

```bash
go run ./cmd/generator -target faq -query "contains(?X, ?Y)" -output docs/scenario/faq.md
go run ./cmd/generator -target api -query "contains(?X, ?Y)" -output docs/scenario/api.md
go run ./cmd/generator -target code -query "contains(?X, ?Y)" -output docs/scenario/types.go
```

---

## 预期输出

### FAQ 示例

```markdown
## Q1: 笔记本电脑 的 属于 是什么？
笔记本电脑 的 属于 是 电子产品。

## Q2: 退换货 的 期限 是什么？
退换货 的 期限 是 30天。
```

### API 设计示例

```markdown
## 笔记本电脑
资源路径: `/api/v1/laptop`
| GET | `/api/v1/laptop` | 列表查询 |
| POST | `/api/v1/laptop` | 创建 |
```

### Go 类型示例

```go
type 笔记本电脑 struct {
    电子产品 string `json:"电子产品"`
}
```

---

## 价值

| 传统方式 | 本管线 |
|---------|--------|
| 客服手动整理 FAQ（3 天） | 自动生成（5 分钟） |
| API 设计依赖人工评审 | 从数据关系自动推导 |
| 文档变更后需全量重做 | 增量抽取 + 推理 |
