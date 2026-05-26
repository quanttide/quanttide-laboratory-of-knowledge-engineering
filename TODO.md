# TODO

> 当前阶段：全部版本已完成，等待 v0.5.0+ 规划

## v0.5.0 — 多源联合

- [x] 数据源抽象接口（Source）+ FileSource / DBSource / APISource
- [x] 联合查询与跨源关联（Federator）
- [x] 统一三元组融合策略（MaxConfidenceFusion / WeightedAverageFusion）
- [x] CLI 入口 `cmd/federate`
- [x] 单元测试

## v0.5.0+ — 增量抽取

- [ ] 文件哈希缓存（SHA256 → 跳过未变更文件）
- [ ] Git-aware 差异检测（仅处理 diff 文件）
- [ ] 增量合并到已有三元组
- [ ] 单元测试

## v0.5.0+ — 自动规则挖掘

- [ ] 子图模式挖掘（Apriori / gSpan）
- [ ] 规则置信度评估
- [ ] 人工确认工作流
- [ ] 单元测试

## v0.5.0+ — 可视化看板

- [ ] Web UI（React / Vue）
- [ ] 三元组浏览器（按置信度 / 来源 / 谓词过滤）
- [ ] 批量确认 / 拒绝 / 修改

## v0.5.0+ — CI 流水线

- [ ] GitHub Actions：push 时运行完整 pipeline
- [ ] 推理结果快照对比（diff 检测）
- [ ] 质量门禁（置信度分布、规则覆盖率）
