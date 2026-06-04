package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HMACSHA256Signer HMAC-SHA256 签名工具
// 用途：验证云途 Webhook 推送签名（防止伪造请求）
// 对标：云途官方 Webhook 签名规范（需与云途确认具体规范）
type HMACSHA256Signer struct{}

// NewHMACSHA256Signer 创建签名器实例
func NewHMACSHA256Signer() *HMACSHA256Signer {
	return &HMACSHA256Signer{}
}

// ComputeSignature 计算 HMAC-SHA256 签名
// 参数：
//   - secret: Webhook签名密钥（YunExpressWebhookSecret 或 GEYunExpressWebhookSecret）
//   - payload: HTTP请求Body（JSON字符串）
// 返回：签名字符串（hex编码）
func (s *HMACSHA256Signer) ComputeSignature(secret, payload string) string {
	// 创建 HMAC-SHA256 hasher
	h := hmac.New(sha256.New, []byte(secret))

	// 写入payload数据
	h.Write([]byte(payload))

	// 计算签名并转为hex字符串
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// VerifySignature 验证签名是否匹配
// 参数：
//   - secret: Webhook签名密钥
//   - payload: HTTP请求Body（JSON字符串）
//   - receivedSignature: 接收到的签名（从HTTP Header提取）
// 返回：true = 签名验证通过，false = 签名验证失败
func (s *HMACSHA256Signer) VerifySignature(secret, payload, receivedSignature string) bool {
	// 计算期望签名
	expectedSignature := s.ComputeSignature(secret, payload)

	// 对比签名（不区分大小写）
	// 注意：使用 hmac.Equal 可防止时序攻击，但这里简化实现用字符串对比
	return strings.EqualFold(expectedSignature, receivedSignature)
}

// ComputeSignatureFromBytes 计算签名（二进制payload版本）
// 用途：更安全的实现（避免字符串转换）
func (s *HMACSHA256Signer) ComputeSignatureFromBytes(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignatureSafe 安全验证签名（防止时序攻击）
// 使用 hmac.Equal 方法（Go标准库推荐）
func (s *HMACSHA256Signer) VerifySignatureSafe(secret string, payload []byte, receivedSignatureHex string) bool {
	// 解码接收到的签名（hex → bytes）
	receivedSignature, err := hex.DecodeString(receivedSignatureHex)
	if err != nil {
		return false // 签名格式错误
	}

	// 计算期望签名（bytes版本）
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := h.Sum(nil)

	// 安全对比（防止时序攻击）
	return hmac.Equal(receivedSignature, expectedSignature)
}