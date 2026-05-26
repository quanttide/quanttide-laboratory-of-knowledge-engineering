# ROADMAP

> 从模糊的原始文档到可执行的知识产物。
>
> 每个版本交付的不是功能，是可独立验证的用户能力。

## 当前状态：项目骨架

```
cmd/app/          # 单文件入口
pkg/knowledge/    # 一个 Document 类型
```

已建立项目结构和构建体系（Go 1.22），无任何业务能力。

## v0.1.0 — 数据层：文件处理管道

**用户能力**：从一批 Markdown 文件中提取结构化元信息和正文内容。

### 交付物

- `cmd/extractor/` — 批处理命令：遍历目录，读取文件，解析 frontmatter
- `internal/markdown/` — Markdown + YAML frontmatter 解析器
- `internal/document/` — Document 模型（元信息 + 正文 + 块序列）
- 支持格式：`.md`（含 YAML frontmatter）、纯文本

### 异常路径覆盖

- YAML frontmatter 格式错误 → 降级为纯文本 + 警告
- 非 UTF-8 编码 → 尝试 GBK/Shift-JIS 回退
- 空文件 / 超大文件 / 二进制文件 → 明确错误类型
- 处理失败的来源记录到 `errors.jsonl`，不中断批处理

### 验收标准

```
go run ./cmd/extractor -input ./docs -output ./parsed
# ✓ 处理 15 个文件，成功 14 个，失败 1 个（详见 errors.jsonl）
# ✓ 输出目录包含每文件对应的 JSON 文档对象
```

---

## v0.2.0 — 信息层：智能标记

**用户能力**：对文档内容进行认知密度评估，提取带置信度的语义三元组。

### 交付物

- `cmd/assessor/` — 认知密度评估命令
- `internal/triple/` — 三元组模型（Subject, Predicate, Object + 置信度 + 溯源）
- `internal/chunker/` — 文本分块（重叠窗口策略）
- `internal/llm/` — LLM 客户端封装（OpenAI API 兼容）
- 输出格式：`triples.jsonl`（每行一个候选事实）

### 异常路径覆盖

- LLM 非 JSON 响应 → 重试 1 次，失败标记 `malformed`
- LLM 超时 → 指数退避（100ms → 500ms → 2s → 放弃）
- 分块边界截断 → 10% 重叠窗口 + 尾部 `truncated` 标记
- 同一来源连续失败 3 次 → 移出队列，通知人工审查

### 验收标准

```
go run ./cmd/assessor -input ./parsed -output ./triples.jsonl
# ✓ 评估 14 个文档，产生 42 条三元组
# ✓ 置信度分布：≥0.7 有 18 条，0.4~0.7 有 15 条，<0.4 有 9 条
```

---

## v0.3.0 — 知识层：入库与推理

**用户能力**：将高置信度三元组加载到图数据库，用 Datalog 规则进行声明式推理。

### 交付物

- `cmd/loader/` — 三元组入库命令（支持 CozoDB / Neo4j）
- `cmd/reasoner/` — Datalog 推理查询命令
- `internal/storage/` — 图数据库驱动接口 + CozoDB 实现
- `internal/rule/` — 规则管理（加载、编译、循环检测）
- `rules/` — 示例 Datalog 规则文件（传递闭包、约束检查）
- 本地回退：当数据库不可用时，降级为 JSON 文件存储

### 异常路径覆盖

- 数据库连接失败 → 重试 3 次，全部失败 → JSON 文件回退
- 写入冲突 → max 置信度合并 + 合并历史记录
- 规则编译错误 → 列出未定义谓词，不中断其他规则
- 推理超时 → 单规则 10s 上限，超时视为不成立

### 验收标准

```
go run ./cmd/loader -input ./triples.jsonl -threshold 0.7 -db cozo
# ✓ 加载 18 条高置信度三元组

go run ./cmd/reasoner -rule rules/contains.mgl -query "contains_tc(订单, ?X)"
# ✓ 返回 ["商品", "用户"]
```

---

## v0.4.0 — 智慧层：应用输出

**用户能力**：基于推理结果生成 Go 代码结构、API 设计建议、FAQ 问答对。

### 交付物

- `cmd/generator/` — 代码/文档生成命令
- `internal/generator/` — 多目标生成器（code / api / faq）
- 输出格式：Go 文件 / Markdown / JSON

### 异常路径覆盖

- 推理结果不足 3 条 → 返回"数据不足"，不生成空壳代码
- 生成器模板缺失 → 列出可用生成器
- 输出路径无写入权限 → 降级到 stdout
- 生成的代码语法无效 → 尝试 `go fmt` 修复，失败则附注

### 验收标准

```
go run ./cmd/generator -target code -query "订单" -output ./generated
# ✓ 生成 type Order struct { ... }
```

---

## v0.5.0+（规划中）

- **多源联合**：支持文档 + 数据库表 + API 响应多源交叉建模
- **增量抽取**：源文档变更后仅重抽取差异内容
- **自动规则挖掘**：从频繁子图生成 Datalog 规则候选
- **可视化看板**：候选事实人工确认界面
- **CI 流水线**：知识库变更触发自动回归推理
