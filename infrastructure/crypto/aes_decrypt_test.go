package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

// TestDecryptPayload_ValidPayload 测试正常解密流程
func TestDecryptPayload_ValidPayload(t *testing.T) {
	// 测试密钥（32字节，AES-256）
	key := "12345678901234567890123456789012"

	// 测试数据（模拟云途推送JSON）
	testCases := []string{
		`{"trackingNumber":"YT2606500704802225","trackingStatus":"20"}`,
		`{"data_code":"tisPushData","body":{"waybill_number":"YT123"}}`,
		`{"package_status":"T","track_events":[{"process_time":"2026-06-05T10:00:00Z"}]}`,
	}

	for i, plaintext := range testCases {
		t.Run("TestCase"+string(rune('A'+i)), func(t *testing.T) {
			// 加密（模拟云途）
			encrypted := encryptForTest(plaintext, key)

			// 解密
			result, err := DecryptPayload(encrypted, key)
			if err != nil {
				t.Fatalf("DecryptPayload failed: %v", err)
			}

			// 验证结果
			if result != plaintext {
				t.Fatalf("expected %s, got %s", plaintext, result)
			}
		})
	}
}

// TestDecryptPayload_InvalidKeyLength 测试密钥长度错误
func TestDecryptPayload_InvalidKeyLength(t *testing.T) {
	testCases := []struct {
		key        string
		expected   string
	}{
		{"12345678", "invalid key length: 8, must be 16/24/32"},
		{"12345678901234567890", "invalid key length: 20, must be 16/24/32"},
	}

	for _, tc := range testCases {
		t.Run("Key_"+tc.key, func(t *testing.T) {
			_, err := DecryptPayload("test", tc.key)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.expected {
				t.Fatalf("expected error %s, got %s", tc.expected, err.Error())
			}
		})
	}
}

// TestDecryptPayload_InvalidBase64 测试Base64解码失败
func TestDecryptPayload_InvalidBase64(t *testing.T) {
	key := "12345678901234567890123456789012"

	_, err := DecryptPayload("not-valid-base64!!!", key)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// 只验证错误包含 "base64 decode failed"
	if !contains(err.Error(), "base64 decode failed") {
		t.Fatalf("expected error contains 'base64 decode failed', got: %v", err)
	}
}

// TestDecryptPayload_CiphertextTooShort 测试密文长度不足
func TestDecryptPayload_CiphertextTooShort(t *testing.T) {
	key := "12345678901234567890123456789012"

	// 构造一个短密文（< 16字节）
	shortCiphertext := base64.StdEncoding.EncodeToString([]byte("short"))

	_, err := DecryptPayload(shortCiphertext, key)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "ciphertext too short (must >= 16 bytes)" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDecryptPayload_InvalidPadding 测试填充无效
func TestDecryptPayload_InvalidPadding(t *testing.T) {
	key := "12345678901234567890123456789012"

	// 构造一个无效填充的密文（手动填充错误）
	invalidData := make([]byte, 32)
	rand.Read(invalidData)
	invalidData[31] = 17 // 填充字节>16（无效）

	encrypted := base64.StdEncoding.EncodeToString(invalidData)

	_, err := DecryptPayload(encrypted, key)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestPkcs7Unpad 测试PKCS7去填充
func TestPkcs7Unpad(t *testing.T) {
	testCases := []struct {
		data       []byte
		blockSize  int
		expected   []byte
	}{
		// 16字节完整填充
		{
			data:      []byte("hello world\x05\x05\x05\x05\x05"),
			blockSize: 16,
			expected:  []byte("hello world"),
		},
		// 15字节填充1字节
		{
			data:      []byte("hello world\x01"),
			blockSize: 16,
			expected:  []byte("hello world"),
		},
	}

	for i, tc := range testCases {
		t.Run("Case"+string(rune('A'+i)), func(t *testing.T) {
			result, err := pkcs7Unpad(tc.data, tc.blockSize)
			if err != nil {
				t.Fatalf("pkcs7Unpad failed: %v", err)
			}
			if string(result) != string(tc.expected) {
				t.Fatalf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestPkcs7Pad 测试PKCS7填充
func TestPkcs7Pad(t *testing.T) {
	testCases := []struct {
		data       []byte
		blockSize  int
		expectedLen int
	}{
		{[]byte("hello"), 16, 16},         // 5字节 → 填充11字节 → 总16字节
		{[]byte("hello world"), 16, 16},   // 11字节 → 填充5字节 → 总16字节
		{[]byte("hello world!"), 16, 16},  // 12字节 → 填充4字节 → 总16字节（修正）
	}

	for i, tc := range testCases {
		t.Run("Case"+string(rune('A'+i)), func(t *testing.T) {
			result := pkcs7Pad(tc.data, tc.blockSize)
			if len(result) != tc.expectedLen {
				t.Fatalf("expected length %d, got %d", tc.expectedLen, len(result))
			}

			// 验证填充字节
			padLen := tc.blockSize - len(tc.data)%tc.blockSize
			for i := len(tc.data); i < len(result); i++ {
				if result[i] != byte(padLen) {
					t.Fatalf("invalid padding byte at %d", i)
				}
			}
		})
	}
}

// contains 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

// encryptForTest 加密函数（用于测试）
func encryptForTest(plaintext string, key string) string {
	block, _ := aes.NewCipher([]byte(key))

	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)

	// PKCS7填充
	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	// AES-CBC加密
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	// IV + ciphertext
	encrypted := append(iv, ciphertext...)

	// Base64编码
	return base64.StdEncoding.EncodeToString(encrypted)
}