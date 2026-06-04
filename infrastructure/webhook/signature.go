package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// VerifySignature HMAC-SHA256签名验证
func VerifySignature(secret string, body []byte, signature string) bool {
	// 解码接收到的签名
	receivedSig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	// 计算期望签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	expectedSig := h.Sum(nil)

	// 安全对比（防止时序攻击）
	return hmac.Equal(receivedSig, expectedSig)
}