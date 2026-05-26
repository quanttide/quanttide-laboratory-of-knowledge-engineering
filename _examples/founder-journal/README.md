# 创始人日志示例

DIKW 四层结构的真实数据示例，基于创始人日常工作日志。

```
data/     原始日志 → cmd/extractor  → info/
info/     标注三元组 → cmd/loader    → knowl/
knowl/    图谱+规则  → cmd/reasoner  → wisdom/
wisdom/   生成产物
```

## data — 原始文档

6 篇 Markdown 日志（带 YAML frontmatter），每日一篇。

## info — 信息层

`cmd/extractor` 解析后的结构化 JSON，以及 `cmd/assessor` 提取的三元组。

## knowl — 知识层

高置信度三元组入库后的 JSON 图谱，以及 Datalog 规则和推理结果。

## wisdom — 智慧层

`cmd/generator` 生成的 Go 代码、API 设计、FAQ。
