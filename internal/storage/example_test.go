package storage_test

import (
	"fmt"
	"os"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/storage"
	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"
)

func ExampleThresholdPolicy_Classify() {
	policy := storage.ThresholdPolicy{High: 0.7, Medium: 0.4}
	triples := []*triple.Triple{
		triple.New("t1", "doc.md", 0.9, "用户", "下单", "订单"),
		triple.New("t2", "doc.md", 0.5, "订单", "包含", "商品"),
		triple.New("t3", "doc.md", 0.2, "商品", "有", "库存"),
	}
	result := policy.Classify(triples)
	fmt.Printf("高置信度: %d\n", len(result.AutoInsert))
	fmt.Printf("中置信度: %d\n", len(result.Pending))
	fmt.Printf("低置信度: %d\n", len(result.Discarded))
	// Output:
	// 高置信度: 1
	// 中置信度: 1
	// 低置信度: 1
}

func ExampleJSONStore_InsertTriples() {
	dir, _ := os.MkdirTemp("", "store-example")
	defer os.RemoveAll(dir)

	store := storage.NewJSONStore(dir)
	defer store.Close()

	t := triple.New("ex", "doc.md", 0.8, "A", "knows", "B")
	inserted, err := store.InsertTriples([]*triple.Triple{t})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("插入 %d 条", inserted)
	// Output: 插入 1 条
}
