// Package utils 工具函数
// 包含：加解密、日期计算、财务精度处理
package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ===================== 加解密工具 =====================

// HMACSHA256Signer HMAC-SHA256签名器
type HMACSHA256Signer struct{}

// NewHMACSHA256Signer 创建签名器
func NewHMACSHA256Signer() *HMACSHA256Signer {
	return &HMACSHA256Signer{}
}

// Sign 生成签名
func (s *HMACSHA256Signer) Sign(secret string, data []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Verify 验证签名
func (s *HMACSHA256Signer) Verify(secret string, data []byte, signature string) bool {
	expected := s.Sign(secret, data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ===================== 日期计算工具 =====================

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, timeStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time: %s", timeStr)
}

// FormatTimestamp 格式化Unix时间戳
func FormatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// GetCurrentTimestamp 获取当前Unix时间戳
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// ===================== 财务精度处理 =====================

// Decimal 精确小数计算（避免浮点误差）
// ERP费用计算必须使用，避免0.1+0.2=0.30000000000000004
type Decimal struct {
	value int64 // 使用整数存储，乘以精度因子
	scale int   // 精度因子（10^scale）
}

// NewDecimal 创建Decimal
func NewDecimal(value float64, scale int) Decimal {
	factor := math.Pow10(scale)
	return Decimal{
		value: int64(value * factor),
		scale: scale,
	}
}

// Add 加法
func (d Decimal) Add(other Decimal) Decimal {
	if d.scale != other.scale {
		// 转换为相同精度
		maxScale := max(d.scale, other.scale)
		d = d.rescale(maxScale)
		other = other.rescale(maxScale)
	}
	return Decimal{
		value: d.value + other.value,
		scale: d.scale,
	}
}

// Sub 减法
func (d Decimal) Sub(other Decimal) Decimal {
	if d.scale != other.scale {
		maxScale := max(d.scale, other.scale)
		d = d.rescale(maxScale)
		other = other.rescale(maxScale)
	}
	return Decimal{
		value: d.value - other.value,
		scale: d.scale,
	}
}

// Mul 乘法
func (d Decimal) Mul(factor float64) Decimal {
	result := float64(d.value) * factor
	return Decimal{
		value: int64(result),
		scale: d.scale,
	}
}

// ToFloat 转换为float64
func (d Decimal) ToFloat() float64 {
	factor := math.Pow10(d.scale)
	return float64(d.value) / factor
}

// String 转换为字符串
func (d Decimal) String() string {
	factor := math.Pow10(d.scale)
	value := float64(d.value) / factor
	return strconv.FormatFloat(value, 'f', d.scale, 64)
}

// rescale 重设精度
func (d Decimal) rescale(newScale int) Decimal {
	if newScale == d.scale {
		return d
	}
	factor := math.Pow10(newScale - d.scale)
	return Decimal{
		value: int64(float64(d.value) * factor),
		scale: newScale,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ===================== 字符串工具 =====================

// IsEmpty 判断字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty 判断字符串是否非空
func IsNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// Truncate 截断字符串
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}