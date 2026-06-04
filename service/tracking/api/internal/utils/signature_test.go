package utils

import (
	"strings"
	"testing"
)

func TestComputeSignature(t *testing.T) {
	signer := NewHMACSHA256Signer()

	// 测试用例：模拟云途Webhook payload
	secret := "your-webhook-secret-key"
	payload := `{"trackingNumber":"YT2606500704802225","wayBillNumber":"YT2606500704802225","trackingStatus":"20","packageState":"2","providerName":"云途物流","orderTrackingDetails":[{"processDate":"2025-05-20T08:00:00Z","processLocation":"深圳","processContent":"快件电子信息已收到","trackingStatus":"10"}]}`

	// 计算签名
	signature := signer.ComputeSignature(secret, payload)

	// 验证签名不为空
	if signature == "" {
		t.Error("Signature should not be empty")
	}

	// 验证签名长度（SHA256 hex编码 = 64字符）
	if len(signature) != 64 {
		t.Errorf("Signature length should be 64, got %d", len(signature))
	}

	t.Logf("Computed signature: %s", signature)
}

func TestVerifySignature(t *testing.T) {
	signer := NewHMACSHA256Signer()

	secret := "test-secret"
	payload := `{"test":"data"}`

	// 计算正确签名
	correctSignature := signer.ComputeSignature(secret, payload)

	// 测试1：正确签名应该验证通过
	if !signer.VerifySignature(secret, payload, correctSignature) {
		t.Error("Correct signature should pass verification")
	}

	// 测试2：错误签名应该验证失败
	wrongSignature := "wrong-signature-abc123"
	if signer.VerifySignature(secret, payload, wrongSignature) {
		t.Error("Wrong signature should fail verification")
	}

	// 测试3：大小写不敏感（hex编码可能大写）
	uppercaseSignature := strings.ToUpper(correctSignature)
	if !signer.VerifySignature(secret, payload, uppercaseSignature) {
		t.Error("Uppercase signature should pass verification")
	}

	// 测试4：不同secret应该验证失败
	differentSecret := "different-secret"
	if signer.VerifySignature(differentSecret, payload, correctSignature) {
		t.Error("Different secret should fail verification")
	}

	t.Logf("All signature verification tests passed")
}

func TestVerifySignatureSafe(t *testing.T) {
	signer := NewHMACSHA256Signer()

	secret := "safe-secret"
	payload := []byte(`{"safe":"test"}`)

	// 计算签名
	signatureHex := signer.ComputeSignatureFromBytes(secret, payload)

	// 测试安全验证（防止时序攻击）
	if !signer.VerifySignatureSafe(secret, payload, signatureHex) {
		t.Error("Safe verification should pass for correct signature")
	}

	// 测试错误签名
	wrongSignatureHex := "aabbccdd12345678"
	if signer.VerifySignatureSafe(secret, payload, wrongSignatureHex) {
		t.Error("Safe verification should fail for wrong signature")
	}

	t.Logf("Safe signature verification test passed")
}