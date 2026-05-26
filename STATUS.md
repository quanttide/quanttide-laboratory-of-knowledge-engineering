# STATUS

> 当前阶段：v0.1.0–v0.4.0 四层架构全部完成

## 项目总览

**quanttide-example-of-knowledge-engineering** — 从模糊的原始文档到可执行的知识产物，基于 Go + 图数据库 + Datalog 推理的知识工具体系。

```
Data           →  Information       →  Knowledge         →  Wisdom
cmd/extractor     cmd/assessor         cmd/loader           cmd/generator
internal/         internal/            cmd/reasoner         internal/
  document          triple               internal/storage      generator
  markdown          chunker              internal/rule
                    llm
```

## 版本状态

| 版本 | 层 | 用户能力 | 状态 |
|------|------|------|------|
| v0.1.0 | 数据层 | 从 Markdown 文件提取结构化元信息和正文 | ✅ 完成 |
| v0.2.0 | 信息层 | 认知密度评估，提取带置信度的语义三元组 | ✅ 完成 |
| v0.3.0 | 知识层 | 高置信度三元组入库，Datalog 声明式推理 | ✅ 完成 |
| v0.4.0 | 智慧层 | 基于推理结果生成 Go 代码 / API 设计 / FAQ | ✅ 完成 |
| v0.5.0+ | 多源联合等 | 增量抽取、自动规则挖掘、可视化看板 | 📋 规划中 |

---

## v0.1.0 — 数据层

### 交付物

| 组件 | 文件 | 说明 |
|------|------|------|
| `cmd/extractor` | `cmd/extractor/main.go` | 批处理命令，支持 `-input`/`-output`，递归遍历目录，过滤 `.md`/`.txt` |
| `internal/document` | `internal/document/document.go` | Document 模型（Path, Frontmatter, Body, Blocks） |
| `internal/markdown` | `internal/markdown/parser.go` | YAML frontmatter 解析器（`---` 分隔，错误降级） |
| `internal/markdown` | `internal/markdown/encoding.go` | 非 UTF-8 编码回退（GBK / Shift-JIS） |

### 异常路径覆盖

- YAML frontmatter 格式错误 → 降级纯文本 + 警告
- frontmatter 未闭合 → 降级 + 警告
- 非 UTF-8 编码 → GBK → Shift-JIS 回退
- 空文件 / >10MB / 二进制 → 明确错误
- 失败记录写入 `errors.jsonl`，不中断批处理

### 验收标准

```bash
go run ./cmd/extractor -input ./docs -output ./parsed
```

---

## v0.2.0 — 信息层

### 交付物

| 组件 | 文件 | 说明 |
|------|------|------|
| `cmd/assessor` | `cmd/assessor/main.go` | 认知密度评估入口，读取 parsed → 分块 → LLM → triples.jsonl |
| `internal/triple` | `internal/triple/triple.go` | 三元组模型（ID, Source, Confidence, SPO, Context, Verification） |
| `internal/chunker` | `internal/chunker/chunker.go` | 文本分块（字符数，10% 重叠窗口，truncated 标记） |
| `internal/llm` | `internal/llm/client.go` | OpenAI API 兼容客户端（指数退避重试 3 次，超时控制） |
| `internal/llm` | `internal/llm/prompt.go` | 认知密度评估 prompt（新颖性 / 领域特异性 / 反直觉程度） |

### 异常路径覆盖

- LLM 非 JSON 响应 → 重试 1 次，标记 `malformed`
- LLM 超时 → 指数退避（100ms → 500ms → 2s → 放弃）
- 分块边界截断 → 10% 重叠窗口 + `truncated` 标记
- 同一来源连续失败 3 次 → 移出队列

### 验收标准

```bash
go run ./cmd/assessor -input ./parsed -output ./triples.jsonl
```

---

## v0.3.0 — 知识层

### 交付物

| 组件 | 文件 | 说明 |
|------|------|------|
| `cmd/loader` | `cmd/loader/main.go` | 三元组入库，支持 `-threshold`/`-db`，max 置信度合并 |
| `cmd/reasoner` | `cmd/reasoner/main.go` | Datalog 推理查询，支持 `-rule`/`-query` |
| `internal/storage` | `internal/storage/store.go` | Store 接口（InsertTriples, Query, Close） |
| `internal/storage` | `internal/storage/json.go` | JSON 文件存储（持久化 store.jsonl，max 置信度合并） |
| `internal/storage` | `internal/storage/threshold.go` | 阈值策略：≥0.7 自动入库 / 0.4~0.7 暂存 / <0.4 丢弃 |
| `internal/rule` | `internal/rule/rule.go` | Datalog 规则引擎（解析 .mgl，未定义谓词检测，DFS 环检测，体求值） |
| `rules/` | `rules/contains.mgl` | 传递闭包 / 约束检查 / 分类层次示例规则 |
| `rules/` | `rules/business.mgl` | 业务规则示例 |

### 异常路径覆盖

- 数据库连接失败 → 重试 3 次 → JSON 回退
- 写入冲突 → max 置信度合并
- 规则编译错误 → 列出未定义谓词，不中断
- 推理超时 → 单规则 10s，超时不成立

### 验收标准

```bash
go run ./cmd/loader -input ./triples.jsonl -threshold 0.7 -db fallback
go run ./cmd/reasoner -rule rules/contains.mgl -query "contains_tc(订单, ?X)"
```

---

## v0.4.0 — 智慧层

### 交付物

| 组件 | 文件 | 说明 |
|------|------|------|
| `cmd/generator` | `cmd/generator/main.go` | 生成入口，支持 `-target`/`-query`/`-output`/`-rule` |
| `internal/generator` | `internal/generator/generator.go` | Generator 接口 + Fact 模型 + 可用生成器注册 |
| `internal/generator` | `internal/generator/code.go` | Go struct 生成器（类型推断：string/int/time.Time/切片） |
| `internal/generator` | `internal/generator/api.go` | RESTful API 设计生成器（CRUD 端点表） |
| `internal/generator` | `internal/generator/faq.go` | FAQ 问答对生成器（去重） |

### 异常路径覆盖

- 推理结果 < 3 条 → 拒绝生成
- 输出路径无写入权限 → 降级 stdout
- 生成代码语法无效 → `go fmt` 修复

### 验收标准

```bash
go run ./cmd/generator -target code -query "contains_tc(?X, ?Y)" -output ./generated
go run ./cmd/generator -target api -query "contains(?X, ?Y)" -output ./generated
go run ./cmd/generator -target faq -query "contains(?X, ?Y)" -output ./generated
```

---

## 基础设施

- Go module（go 1.22）
- `.gitignore`（Go 标准 + IDE）
- `Makefile`（build / test / lint / clean）
- CI 配置（`.github/workflows/ci.yml`：go build + vet + test）
- 项目结构对齐 ROADMAP
- 测试夹具（`testdata/extractor/`）
- AGENTS.md + docs/design.md

## 文件统计

```
cmd/         5 入口     (extractor, assessor, loader, reasoner, generator)
internal/   10 包       (document, markdown, triple, chunker, llm, storage, rule, generator)
rules/       2 规则文件  (contains.mgl, business.mgl)
docs/        5 文件     (design.md + 4 示例文档)
tests       10 测试文件  (覆盖率覆盖各包)
```
