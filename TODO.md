# TODO

> 当前阶段：项目骨架（v0.1.0 尚未开始）

## v0.1.0 — 数据层：文件处理管道

- [ ] 定义 `internal/document/` 包：Document 模型（元信息 + 正文 + 块序列）
- [ ] 定义 `internal/markdown/` 包：Markdown + YAML frontmatter 解析器
  - [ ] frontmatter 解析（`---` 分隔符提取）
  - [ ] YAML 解析（支持错误降级）
  - [ ] 正文提取（frontmatter 后的 Markdown 内容）
- [ ] 创建 `cmd/extractor/` 批处理入口
  - [ ] `-input` 参数（文件或目录）
  - [ ] `-output` 参数（输出目录）
  - [ ] 递归遍历目录
  - [ ] 过滤支持的文件类型（`.md`、`.txt`）
- [ ] 异常路径实现
  - [ ] YAML frontmatter 格式错误降级
  - [ ] 非 UTF-8 编码回退
  - [ ] 空文件/超大文件/二进制文件拒绝处理
  - [ ] 失败记录写入 `errors.jsonl`，不中断批处理
- [ ] 单元测试
  - [ ] `internal/markdown/` 测试（正常 frontmatter、格式错误、无 frontmatter）
  - [ ] `internal/document/` 测试
  - [ ] `cmd/extractor/` 集成测试（含异常路径）

## v0.2.0 — 信息层：智能标记

- [ ] 定义 `internal/triple/` 包：三元组模型
  - [ ] 核心字段：Subject, Predicate, Object, Confidence, Source, Verification 状态
  - [ ] JSON 序列化/反序列化
- [ ] 定义 `internal/chunker/` 包：文本分块
  - [ ] 按字符数分块
  - [ ] 重叠窗口策略（10% overlap）
  - [ ] 尾部 `truncated` 标记
- [ ] 实现 `internal/llm/` 包：LLM 客户端
  - [ ] OpenAI API 兼容接口
  - [ ] 指数退避重试
  - [ ] 超时控制
- [ ] 实现认知密度评估 prompt
  - [ ] 评估标准定义（新颖性、领域特异性、反直觉程度）
  - [ ] JSON 输出格式约束
  - [ ] 必填字段缺失时标记 `malformed`
- [ ] 创建 `cmd/assessor/` 入口
  - [ ] 读取 parsed JSON → 分块 → LLM 评估 → 输出 triples.jsonl
  - [ ] 同一来源连续失败 3 次则移除
- [ ] 单元测试
  - [ ] chunker 测试（边界截断、重叠窗口）
  - [ ] triple 模型测试
  - [ ] llm 客户端 mock 测试

## v0.3.0 — 知识层：入库与推理

- [ ] 定义 `internal/storage/` 接口
  - [ ] Store 接口：InsertTriples, Query, Close
  - [ ] CozoDB 实现
  - [ ] JSON 文件回退实现
- [ ] 定义 `internal/rule/` 包：规则管理
  - [ ] 规则加载（`.mgl` 文件）
  - [ ] 编译错误检测（未定义谓词列表）
  - [ ] 循环依赖检测
- [ ] 实现阈值策略
  - [ ] 高置信度（≥0.7）自动入库
  - [ ] 中置信度（0.4~0.7）暂存待确认
  - [ ] 低置信度（<0.4）保留不入库
  - [ ] 阈值可配置
- [ ] 创建 `cmd/loader/` 入口
  - [ ] `-threshold` 参数
  - [ ] `-db` 参数（支持 cozo / fallback）
  - [ ] 写入冲突合并（max 置信度）
- [ ] 创建 `cmd/reasoner/` 入口
  - [ ] `-rule` 参数（规则文件路径）
  - [ ] `-query` 参数（查询目标）
  - [ ] 单规则超时控制（10s）
- [ ] 异常路径实现
  - [ ] 数据库连接重试 + 回退
  - [ ] 推理超时不中断其他规则
  - [ ] 空结果记录日志
- [ ] 单元测试 + 集成测试

## v0.4.0 — 智慧层：应用输出

- [ ] 定义 `internal/generator/` 包
  - [ ] Generator 接口：Generate(input Facts) → Output
  - [ ] code 生成器（Go struct 模板）
  - [ ] api 生成器（RESTful API 设计建议）
  - [ ] faq 生成器（问答对）
- [ ] 创建 `cmd/generator/` 入口
  - [ ] `-target` 参数（code / api / faq）
  - [ ] `-query` 参数（查询词）
  - [ ] `-output` 参数
- [ ] 异常路径实现
  - [ ] 推理结果 < 3 条时拒绝生成
  - [ ] 输出路径无写入权限降级到 stdout
  - [ ] 生成代码语法检查 + go fmt 修复
- [ ] 示例 rules 文件
  - [ ] 传递闭包规则
  - [ ] 约束检查规则
  - [ ] 分类层次规则
- [ ] 端到端集成测试

## 基础设施

- [ ] 项目结构对齐 ROADMAP（`cmd/`、`internal/`、`rules/`、`docs/`）
- [ ] CI 配置（`go build`、`go test`、`go vet`）
- [ ] 添加 `.gitignore`（Go 标准 + IDE 文件）
- [ ] 补充 Makefile（`build`、`test`、`lint` 目标）

## 已完成

- [x] 初始化 Go module（go 1.22）
- [x] 添加 AGENTS.md（设计方案）
- [x] 添加 docs/design.md（含异常路径）
- [x] 添加 ROADMAP.md（v0.1.0–v0.5.0+）
- [x] 清理旧 Python 项目遗留文件
