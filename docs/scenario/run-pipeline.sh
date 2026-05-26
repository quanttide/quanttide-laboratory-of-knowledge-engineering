#!/bin/bash
# TechMart 智能客服知识库管线
# 完整演示 Data → Information → Knowledge → Wisdom
set -e

SCENARIO_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(dirname "$SCENARIO_DIR")/.."
DATA="$ROOT/testdata/scenario/products"
PARSED="$SCENARIO_DIR/parsed"
TRIPLES="$SCENARIO_DIR/triples.jsonl"
GENERATED="$SCENARIO_DIR/generated"

echo "=== 1. 数据提取 ==="
go run "$ROOT/cmd/extractor" -input "$DATA" -output "$PARSED"
echo ""

echo "=== 2. 智能标记（需设置 OPENAI_API_KEY）==="
if [ -n "$OPENAI_API_KEY" ]; then
  go run "$ROOT/cmd/assessor" -input "$PARSED" -output "$TRIPLES"
else
  echo "跳过：OPENAI_API_KEY 未设置"
fi
echo ""

echo "=== 3. 入库 ==="
go run "$ROOT/cmd/loader" -input "$TRIPLES" -threshold 0.7
echo ""

echo "=== 4. 推理 ==="
go run "$ROOT/cmd/reasoner" -rule "$ROOT/rules/scenario" -query "subclass_tc(笔记本电脑, ?X)"
go run "$ROOT/cmd/reasoner" -rule "$ROOT/rules/scenario" -query "product_category(?X, 电子产品)"
go run "$ROOT/cmd/reasoner" -rule "$ROOT/rules/scenario" -query "budget_friendly(?X)"
echo ""

echo "=== 5. 生成 ==="
mkdir -p "$GENERATED"
go run "$ROOT/cmd/generator" -target faq -query "belongs_to(?X, ?Y)" -output "$GENERATED"
go run "$ROOT/cmd/generator" -target api -query "belongs_to(?X, ?Y)" -output "$GENERATED"
go run "$ROOT/cmd/generator" -target code -query "belongs_to(?X, ?Y)" -output "$GENERATED"
echo ""

echo "=== 完成 ==="
echo "输出文件:"
ls -la "$GENERATED/"
