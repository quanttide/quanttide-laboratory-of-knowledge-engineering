package triple

import (
	"encoding/json"
	"testing"
)

func TestNewTriple(t *testing.T) {
	tri := New("test-1", "doc.md", 0.85, "用户", "下单", "订单")
	if tri.Subject != "用户" {
		t.Errorf("expected Subject=用户, got %s", tri.Subject)
	}
	if tri.Predicate != "下单" {
		t.Errorf("expected Predicate=下单, got %s", tri.Predicate)
	}
	if tri.Object != "订单" {
		t.Errorf("expected Object=订单, got %s", tri.Object)
	}
	if tri.Confidence != 0.85 {
		t.Errorf("expected Confidence=0.85, got %f", tri.Confidence)
	}
	if tri.Verification != VerificationUnverified {
		t.Errorf("expected Verification=unverified, got %s", tri.Verification)
	}
	if tri.ID != "test-1" {
		t.Errorf("expected ID=test-1, got %s", tri.ID)
	}
}

func TestMarkVerified(t *testing.T) {
	tri := New("t1", "src", 0.9, "A", "B", "C")
	tri.MarkVerified()
	if tri.Verification != VerificationVerified {
		t.Errorf("expected verified, got %s", tri.Verification)
	}
}

func TestMarkMalformed(t *testing.T) {
	tri := New("t1", "src", 0.9, "A", "B", "C")
	tri.MarkMalformed()
	if tri.Verification != VerificationMalformed {
		t.Errorf("expected malformed, got %s", tri.Verification)
	}
}

func TestSetContext(t *testing.T) {
	tri := New("t1", "src", 0.9, "A", "B", "C")
	tri.SetContext("sentence", "A B C")
	if tri.Context["sentence"] != "A B C" {
		t.Errorf("unexpected context: %v", tri.Context)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	tri := New("t1", "src.md", 0.75, "用户", "包含", "商品")
	tri.SetContext("sentence", "订单包含商品")

	data, err := json.Marshal(tri)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Triple
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != tri.ID {
		t.Errorf("ID mismatch: %s vs %s", decoded.ID, tri.ID)
	}
	if decoded.Confidence != tri.Confidence {
		t.Errorf("Confidence mismatch: %f vs %f", decoded.Confidence, tri.Confidence)
	}
	if decoded.Context["sentence"] != "订单包含商品" {
		t.Errorf("Context mismatch: %v", decoded.Context)
	}
}

func TestEmptyContext(t *testing.T) {
	tri := New("t1", "src", 0.5, "A", "B", "C")
	if tri.Context == nil {
		t.Error("expected non-nil context map even when empty")
	}
}
