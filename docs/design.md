# 设计文档

## 架构

```
原始文档 ──→ cmd/extractor ──→ cmd/assessor ──→ cmd/loader ──→ cmd/reasoner ──→ cmd/generator ──→ 产物
  Data           Information       Knowledge       Knowledge        Wisdom
```

四层各司其职：

| 层 | 入口 | 输入 | 输出 |
|---|------|------|------|
| 数据层 | `cmd/extractor` | `.md` / `.txt` 文件 | 结构化 Document JSON |
| 信息层 | `cmd/assessor` | Document JSON | 三元组 JSONL（置信度 + 溯源） |
| 知识层 | `cmd/loader` + `cmd/reasoner` | 三元组 JSONL + `.mgl` 规则 | 图谱存储 + 推理结果 |
| 智慧层 | `cmd/generator` | 推理结果 | Go 代码 / API 设计 / FAQ |

## 关键决策

| 问题 | 决策 | 理由 |
|------|------|------|
| 存储 | JSON 文件 (`JSONStore`) | 零依赖，go test 可直接运行 |
| 推理引擎 | 内嵌 Datalog (`internal/rule`) | 保证终止，顺序无关 |
| 图查询 | `QueryNeighbors` / `QueryPath` | 邻接点与路径搜索，无需外部图谱库 |
| 三元组格式 | JSONL | 行级追加，流式处理 |
| 不确定性 | 置信度 + Verification 状态机 | 保留溯源路径，支持增量累积 |
| 分块 | 重叠窗口 10% + truncated 标记 | 减少边界截断导致的语义丢失 |
| 外部依赖 | 无 | markdown 解析、YAML 解析、编码识别全部内联 |

## 异常路径

每层都有可独立验证的失败模式：

- **数据层**：YAML 格式错误降级 / 非 UTF-8 编码回退 / 空/超大/二进制文件拒绝 / 失败记录 `errors.jsonl` 不中断
- **信息层**：LLM 超时指数退避 / 非 JSON 响应重试 / 同一来源连续 3 次失败移出队列
- **知识层**：写入冲突 max 置信度合并 / 未定义谓词检测 / 循环依赖检测 / 推理超时 10s
- **智慧层**：不足 3 条拒绝生成 / 无写入权限降级 stdout / `go fmt` 修复

## 阈值策略

```
≥ 0.7  自动入库
0.4–0.7 暂存待确认（写入 pending.jsonl）
< 0.4  保留不入库
```

阈值通过 `-threshold` 可配置。
