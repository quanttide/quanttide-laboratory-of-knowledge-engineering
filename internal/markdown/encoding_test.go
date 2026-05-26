package markdown

import (
	"testing"
)

func TestDetectUTF8(t *testing.T) {
	raw := []byte("hello world")
	_, enc, err := DetectAndDecode(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != EncodingUTF8 {
		t.Errorf("expected UTF-8, got %s", enc)
	}
}

func TestDetectUTF8Chinese(t *testing.T) {
	raw := []byte("你好世界")
	_, enc, err := DetectAndDecode(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != EncodingUTF8 {
		t.Errorf("expected UTF-8, got %s", enc)
	}
}

func TestDetectEmpty(t *testing.T) {
	_, _, err := DetectAndDecode([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestDetectGBK(t *testing.T) {
	raw := []byte{0xD6, 0xD0, 0xB9, 0xFA} // 中国 in GBK
	decoded, enc, err := DetectAndDecode(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != EncodingGBK {
		t.Errorf("expected GBK, got %s", enc)
	}
	if string(decoded) != "中国" {
		t.Errorf("expected 中国, got %s", string(decoded))
	}
}

func TestDetectGBKNotFound(t *testing.T) {
	raw := []byte{0xFF, 0xFE, 0xFD} // invalid GBK
	_, _, err := DetectAndDecode(raw)
	if err == nil {
		t.Fatal("expected error for invalid encoding")
	}
}

func TestIsLikelyUTF8(t *testing.T) {
	if !isLikelyUTF8([]byte("hello")) {
		t.Error("plain text should be likely UTF-8")
	}
	if isLikelyUTF8([]byte{0xFF, 0xFE}) {
		t.Error("0xFF 0xFE should not be likely UTF-8")
	}
	if !isLikelyUTF8([]byte("你好")) {
		t.Error("Chinese UTF-8 should pass")
	}
}

func TestTryDecodeGBK(t *testing.T) {
	decoded, ok := tryDecodeGBK([]byte{0xD6, 0xD0, 0xB9, 0xFA})
	if !ok {
		t.Fatal("expected GBK decode to succeed")
	}
	if string(decoded) != "中国" {
		t.Errorf("expected 中国, got %s", string(decoded))
	}
}

func TestTryDecodeGBKInvalid(t *testing.T) {
	_, ok := tryDecodeGBK([]byte{0x81, 0x20}) // 0x20 is not in GBK second byte range
	if ok {
		t.Error("expected invalid GBK to fail")
	}

	_, ok = tryDecodeGBK([]byte{0x81}) // truncated GBK sequence
	if ok {
		t.Error("expected truncated GBK to fail")
	}
}

func TestTryDecodeShiftJIS(t *testing.T) {
	raw := []byte{0x94, 0x4C, 0x8D, 0x91} // 中国 in Shift-JIS
	decoded, ok := tryDecodeShiftJIS(raw)
	if !ok {
		t.Fatal("expected Shift-JIS decode to succeed")
	}
	if string(decoded) != "中国" {
		t.Errorf("expected 中国, got %s", string(decoded))
	}
}

func TestTryDecodeShiftJISInvalid(t *testing.T) {
	_, ok := tryDecodeShiftJIS([]byte{0x81, 0x20}) // invalid second byte
	if ok {
		t.Error("expected invalid Shift-JIS to fail")
	}
}

func TestIsLikelyText(t *testing.T) {
	if !isLikelyText("hello") {
		t.Error("plain text should be likely text")
	}
	if isLikelyText("\x00\x00\x00") {
		t.Error("null bytes should not be likely text")
	}
}

func TestDetectAndDecodeUTF8MultiByte(t *testing.T) {
	raw := []byte{0xE4, 0xBD, 0xA0, 0xE5, 0xA5, 0xBD} // 你好
	decoded, enc, err := DetectAndDecode(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != EncodingUTF8 {
		t.Errorf("expected UTF-8, got %s", enc)
	}
	if string(decoded) != "你好" {
		t.Errorf("expected 你好, got %s", string(decoded))
	}
}

func TestGBKCodepoint(t *testing.T) {
	r := gbkCodepoint(0xB0, 0xA1) // 啊
	if r != '啊' {
		t.Errorf("expected 啊 (U+554A), got %U", r)
	}
	r = gbkCodepoint(0xFF, 0xFF) // unknown
	if r != 0 {
		t.Errorf("expected 0 for unknown codepoint, got %U", r)
	}
}

func TestShiftJISCodepoint(t *testing.T) {
	r := shiftJISCodepoint(0x82, 0xA0) // あ
	if r != 'あ' {
		t.Errorf("expected あ (U+3042), got %U", r)
	}
	r = shiftJISCodepoint(0xFF, 0xFF) // unknown
	if r != 0 {
		t.Errorf("expected 0 for unknown codepoint, got %U", r)
	}
}

func TestDetectAlreadyUTF8(t *testing.T) {
	raw := []byte("ASCII only")
	decoded, enc, err := DetectAndDecode(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != EncodingUTF8 {
		t.Errorf("expected UTF-8 for ASCII, got %s", enc)
	}
	if string(decoded) != "ASCII only" {
		t.Errorf("content mismatch: %s", string(decoded))
	}
}
