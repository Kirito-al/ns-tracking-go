package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

// DecryptPayload AES-CBC解密云途Webhook加密payload
// 对标：tms-talk/app/services/logistics/adapters/yuntu.py:_decrypt_if_needed
// 算法：AES-256-CBC + PKCS7Padding
//
// 参数：
//   - encryptedBase64: Base64编码的加密数据（IV + ciphertext）
//   - encryptKey: AES密钥（必须为16/24/32字节）
//
// 返回：
//   - 解密后的JSON字符串
//   - 错误信息（解密失败时）
//
// 数据格式：
//   encryptedBase64 = Base64Encode(IV[16字节] + AES_CBC_Encrypt(plaintext))
//
// 使用场景：
//   云途Webhook推送加密payload时，payload格式为：
//   { "encrypt": "Base64编码的密文" }
//   需先解密，再解析JSON
func DecryptPayload(encryptedBase64 string, encryptKey string) (string, error) {
	// 验证密钥长度（AES-128=16字节, AES-192=24字节, AES-256=32字节）
	keyLen := len(encryptKey)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return "", fmt.Errorf("invalid key length: %d, must be 16/24/32", keyLen)
	}

	// 1. Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	// 2. 验证ciphertext长度
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short (must >= 16 bytes)")
	}
	if (len(ciphertext)-aes.BlockSize)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of block size")
	}

	// 3. 提取IV（前16字节）
	iv := ciphertext[:aes.BlockSize]
	actualCiphertext := ciphertext[aes.BlockSize:]

	// 4. AES-CBC解密
	block, err := aes.NewCipher([]byte(encryptKey))
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher failed: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(actualCiphertext))
	mode.CryptBlocks(plaintext, actualCiphertext)

	// 5. PKCS7去填充
	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("pkcs7 unpad failed: %w", err)
	}

	return string(plaintext), nil
}

// pkcs7Unpad PKCS7去填充
// 对标：Python padding.PKCS7.unpad()
//
// 参数：
//   - data: 解密后的数据（带填充）
//   - blockSize: 填充块大小（AES固定为16字节）
//
// 返回：
//   - 去填充后的原始数据
//   - 错误信息（填充无效时）
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	// 提取填充长度（最后一个字节）
	padLen := int(data[len(data)-1])

	// 验证填充长度合法性
	if padLen <= 0 || padLen > blockSize {
		return nil, fmt.Errorf("invalid padding length: %d (must be 1~%d)", padLen, blockSize)
	}

	// 验证填充内容（所有填充字节必须等于padLen）
	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return nil, errors.New("invalid PKCS7 padding (padding bytes mismatch)")
		}
	}

	// 返回去填充后的数据
	return data[:len(data)-padLen], nil
}

// pkcs7Pad PKCS7填充（用于测试加密）
// 对标：Python padding.PKCS7.pad()
//
// 参数：
//   - data: 原始数据
//   - blockSize: 填充块大小（AES固定为16字节）
//
// 返回：
//   - 填充后的数据
func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padLen)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}
	return padded
}