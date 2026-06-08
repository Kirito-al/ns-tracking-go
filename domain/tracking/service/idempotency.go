package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateIdempotencyKey 生成幂等key
// 用途：去重同一次推送的重复投递（网络重试等）
// 技术方案：provider_code + waybill_number + track_node_code + process_time
// 算法：SHA256(raw) → 取前16字节 → hex编码
//
// 参数：
//   - providerCode: 服务商代码（如：yunexpress）
//   - waybillNumber: 主单号（如：YT2615000705163221）
//   - trackNodeCode: 节点代码（如：ORDER_CREATION）
//   - processTime: 事件时间（如：2025-06-10T16:30:16）
//
// 返回：
//   - 幂等key（64字符hex字符串，取SHA256前16字节）
//
// 注意：
//   - 幂等key用于去重同一事件的重复推送，不去重相同track_node_code的不同轨迹事件
//   - 多个中转站的IN_TRANSIT事件，每个都该存（process_time不同）
func GenerateIdempotencyKey(providerCode, waybillNumber, trackNodeCode, processTime string) string {
	// 拼接原始字符串（冒号分隔）
	raw := fmt.Sprintf("%s:%s:%s:%s", providerCode, waybillNumber, trackNodeCode, processTime)

	// SHA256哈希
	hash := sha256.Sum256([]byte(raw))

	// 取前16字节（32字符hex）
	// 注意：SHA256完整输出是32字节（64字符hex），取前16字节足够去重
	return hex.EncodeToString(hash[:16])
}