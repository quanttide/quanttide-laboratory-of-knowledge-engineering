# 创始人小说示例

DIKW 四层结构的真实数据示例，基于创始人的职场言情小说章节。

```
data/     原始小说 → cmd/extractor  → info/
info/     标注三元组 → cmd/loader    → knowl/
knowl/    图谱+规则  → cmd/reasoner  → wisdom/
wisdom/   生成产物
```

## data

13 篇小说章节（Markdown），含对话、叙事、情感描写。与 founder-journal 的
技术反思不同，fiction 提供创意写作领域的语料，验证知识抽取在不同文体上的表现。

## info

`cmd/extractor` 解析后的结构化 JSON。

## knowl

高置信度三元组入库后的 JSON 图谱。

## wisdom

`cmd/generator` 生成的代码、API 设计、FAQ。
