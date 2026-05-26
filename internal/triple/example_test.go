package triple_test

import (
	"fmt"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"
)

func ExampleNew() {
	t := triple.New("extract-001", "doc.md", 0.85, "用户", "下单", "订单")
	fmt.Println(t.Subject)
	fmt.Println(t.Predicate)
	fmt.Println(t.Object)
	fmt.Println(t.Confidence)
	fmt.Println(t.Verification)
	// Output:
	// 用户
	// 下单
	// 订单
	// 0.85
	// unverified
}

func ExampleTriple_Context() {
	t := triple.New("t1", "src.md", 0.9, "订单", "包含", "商品")
	t.SetContext("sentence", "订单包含商品")
	fmt.Println(t.Context["sentence"])
	// Output: 订单包含商品
}

func ExampleTriple_Verification() {
	t := triple.New("t1", "doc.md", 0.95, "A", "B", "C")
	fmt.Println(string(t.Verification))
	t.MarkVerified()
	fmt.Println(string(t.Verification))
	// Output:
	// unverified
	// verified
}


func ExampleTriple_MarkVerified() {
	t := triple.New("t1", "doc.md", 0.95, "A", "B", "C")
	t.MarkVerified()
	fmt.Println(t.Verification == triple.VerificationVerified)
	// Output: true
}
